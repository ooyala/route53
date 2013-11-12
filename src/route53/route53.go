package route53

import (
	"encoding/xml"
	"fmt"
	"github.com/crowdmob/goamz/aws"
)

var debug bool

func DebugOn() {
	debug = true
}

func DebugOff() {
	debug = false
}

type Route53 struct {
	auth aws.Auth
}

func New(auth aws.Auth) *Route53 {
	return &Route53{
		auth: auth,
	}
}

type ChangeInfo struct {
	Id          string
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
		path:   fmt.Sprintf("/2012-12-12/change/%s", id),
	}

	xmlRes := &GetChangeResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return ChangeInfo{}, err
	}

	return xmlRes.ChangeInfo, nil
}
