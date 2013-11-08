package route53

import (
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
	auth  aws.Auth
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
