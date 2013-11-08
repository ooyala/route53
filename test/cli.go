package main

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"route53"
	"time"
)

func main() {
	route53.DebugOn()

	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		panic(err)
	}

	r53 := route53.New(auth)

	zone, _ := r53.GetHostedZone("ZEMXKDVPI3AMD")
	fmt.Printf("-- zone\n%#v\n", zone)
}

