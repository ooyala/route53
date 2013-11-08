package route53

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// XML RPC types.

type ChangeRRSetRequest struct {
	XMLName xml.Name      `xml:"ChangeResourceRecordSetsRequest"`
	Comment string        `xml:"ChangeBatch>Comment"`
	Changes []RRSetChange `xml:"ChangeBatch>Changes"`
}

type RRSetChange struct {
	Action string
	RRSet  RRSet
}

type RRSet struct {
	zone *HostedZone `xml:"-"`
	// Basic Resource Record
	Name          string
	Type          string
	TTL           uint
	Values        []string `xml:"ResourceRecords>ResourceRecord>Value"`
	HealthCheckId string   `xml:"omitempty"`
	// Weight Syntax
	Weight        uint8  `xml:"omitempty"`
	SetIdentifier string `xml:"omitempty"`
	// Alias Syntax
	AliasTarget AliasTarget `xml:"omitempty"`
	// Fail Syntax
	FailOver string `xml:"omitempty"`
	// Latency Syntax
	Region string `xml:"omitempty"`
}

type AliasTarget struct {
	HostedZoneId         string
	DNSName              string
	EvaluateTargetHealth bool
}

type ChangeRRSetResponse struct {
	XMLName    xml.Name `xml:"ChangeResourceRecordSetResponse"`
	ChangeInfo ChangeInfo
}

type ListRRSetResponse struct {
	XMLName              xml.Name `xml:"ListResourceRecordSetsResponse"`
	RRSets               []RRSet  `xml:"ResourceRecordSets>ResourceRecordSet"`
	IsTruncated          bool
	NextRecordName       string
	NextRecordIdentifier string
	MaxItems             uint
}

// Route53 API requests.

func (z *HostedZone) ChangeRRSetRequest(changes []RRSetChange, comment string) (ChangeInfo, error) {
	xmlReq := &ChangeRRSetRequest{
		Comment: comment,
		Changes: changes,
	}

	req := request{
		method: "POST",
		path:   fmt.Sprintf("/2012-12-12/hostedzone/%s/rrset", z.Id),
		body:   xmlReq,
	}

	xmlRes := &ChangeRRSetResponse{}

	if err := z.r53.run(req, xmlRes); err != nil {
		return ChangeInfo{}, err
	}

	return xmlRes.ChangeInfo, nil
}

func (z *HostedZone) ListRRSet() ([]RRSet, error) {
	req := request{
		method: "GET",
		path:   fmt.Sprintf("/2012-12-12/hostedzone/%s/rrset", z.Id),
	}

	xmlRes := &ListRRSetResponse{}

	if err := z.r53.run(req, xmlRes); err != nil {
		return []RRSet{}, err
	}
	if xmlRes.IsTruncated {
		return []RRSet{}, errors.New("cannot handle truncated responses")
	}

	for _, rrset := range xmlRes.RRSets {
		rrset.zone = z
	}

	return xmlRes.RRSets, nil
}

// Convenience functions on AWS APIs.

func (z HostedZone) CreateRRSet(rrset RRSet, comment string) (ChangeInfo, error) {
	change := RRSetChange{
		Action: "CREATE",
		RRSet:  rrset,
	}

	return z.ChangeRRSetRequest([]RRSetChange{change}, comment)
}

func (rrset RRSet) Delete(comment string) (ChangeInfo, error) {
	change := RRSetChange{
		Action: "DELETE",
		RRSet:  rrset,
	}

	return rrset.zone.ChangeRRSetRequest([]RRSetChange{change}, comment)
}
