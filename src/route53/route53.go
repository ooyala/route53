package route53

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"log"
	"strings"
	"sync"
	"time"
)

var debug bool

func DebugOn() {
	debug = true
}

func DebugOff() {
	debug = false
}

type Route53 struct {
	auth          aws.Auth
	authLock      sync.RWMutex
	IncludeWeight bool
}

func (r53 *Route53) updateAuth() {
	r53.authLock.Lock()
	// update auth
	auth, err := aws.GetAuth("", "", "", time.Time{})
	for ; err != nil; auth, err = aws.GetAuth("", "", "", time.Time{}) {
		if debug {
			log.Printf("[Route53] Error getting auth (sleeping 5s before retry): %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	r53.auth = auth
	if debug {
		log.Printf("[Route53] auth updated. expires at %v.", auth.Expiration())
	}
	r53.authLock.Unlock()
}

func (r53 *Route53) updateAuthLoop() {
	if r53.auth.Expiration().IsZero() {
		// no exp, don't update
		log.Printf("[Route53] No need to update auth, exiting token update loop.")
		return
	}
	for {
		if diff := r53.auth.Expiration().Sub(time.Now()); diff <= 0 {
			r53.updateAuth()
		} else {
			// sleep
			if debug {
				log.Printf("[Route53] auth not expired. sleeping %v until expiry.", diff)
			}
			time.Sleep(diff)
		}
	}
}

func New() (*Route53, error) {
	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		return nil, err
	}
	r53 := &Route53{
		auth:     auth,
		authLock: sync.RWMutex{},
	}
	go r53.updateAuthLoop()
	return r53, nil
}

func NewWithAuth(auth aws.Auth) *Route53 {
	r53 := &Route53{
		auth:     auth,
		authLock: sync.RWMutex{},
	}
	return r53
}

type ChangeInfo struct {
	r53         *Route53 `xml:"-"`
	ID          string `xml:"Id"`
	Status      string
	SubmittedAt string
}

type GetChangeResponse struct {
	XMLName    xml.Name `xml:"GetChangeResponse"`
	ChangeInfo ChangeInfo
}

func (r53 *Route53) GetChange(id string) (ChangeInfo, error) {
	req := request{
		method: "GET",
		path:   fmt.Sprintf("/2012-12-12/change/%s", strings.Replace(id, "/change/", "", -1)),
	}

	xmlRes := &GetChangeResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return ChangeInfo{}, err
	}

	return xmlRes.ChangeInfo, nil
}

func (c *ChangeInfo) PollForSync(every, tout time.Duration) chan error {
	result := make(chan error)
	go func() {
		toutC := time.After(tout)
		pollC := time.Tick(every)
		for {
			select {
			case <-pollC:
				change, err := c.r53.GetChange(c.ID)
				if err != nil {
					result <- err
					return
				}
				if change.Status == "INSYNC" {
					result <- nil
					return
				}
			case <-toutC:
				result <- errors.New("timed out")
				return
			}
		}
	}()

	return result
}
