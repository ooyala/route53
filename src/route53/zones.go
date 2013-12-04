package route53

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

// XML RPC types.

type HostedZone struct {
	r53                    *Route53 `xml:"-"`
	Id                     string
	Name                   string
	CallerReference        string
	Comment                string `xml:"Config>Comment"`
	ResourceRecordSetCount int
}

type CreateHostedZoneRequest struct {
	XMLName         xml.Name `xml:"CreateHostedZoneRequest"`
	XMLNS           string   `xml:"xmlns,attr"`
	Name            string
	CallerReference string
	Comment         string `xml:"HostedZoneConfig>Comment"`
}

type CreateHostedZoneResponse struct {
	XMLName     xml.Name `xml:"CreateHostedZoneResponse"`
	HostedZone  HostedZone
	ChangeInfo  ChangeInfo
	NameServers []string `xml:"DelegationSet>NameServers>NameServer"`
}

type GetHostedZoneResponse struct {
	XMLName     xml.Name `xml:"GetHostedZoneResponse"`
	HostedZone  HostedZone
	NameServers []string `xml:"DelegationSet>NameServers>NameServer"`
}

type ListHostedZonesResponse struct {
	XMLName     xml.Name `xml:"ListHostedZonesResponse"`
	HostedZones []HostedZone
	IsTruncated bool
	Marker      string
	NextMarker  string
	MaxItems    uint
}

type DeleteHostedZoneResponse struct {
	ChangeInfo ChangeInfo
}

// Route53 API requests.

func (r53 *Route53) CreateHostedZone(name, reference, comment string) (ChangeInfo, error) {
	xmlReq := &CreateHostedZoneRequest{
		XMLNS:           "https://route53.amazonaws.com/doc/2012-12-12/",
		Name:            name,
		CallerReference: reference,
		Comment:         comment,
	}

	req := request{
		method: "POST",
		path:   "/2012-12-12/hostedzone",
		body:   xmlReq,
	}

	xmlRes := &CreateHostedZoneResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return ChangeInfo{}, err
	}
	xmlRes.ChangeInfo.r53 = r53

	return xmlRes.ChangeInfo, nil
}

func (r53 *Route53) GetHostedZone(id string) (HostedZone, error) {
	req := request{
		method: "GET",
		path:   fmt.Sprintf("/2012-12-12/hostedzone/%s", strings.Replace(id, "/hostedzone/", "", -1)),
	}

	xmlRes := &GetHostedZoneResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return HostedZone{}, err
	}
	xmlRes.HostedZone.r53 = r53

	return xmlRes.HostedZone, nil
}

func (r53 *Route53) ListHostedZones() ([]HostedZone, error) {
	req := request{
		method: "GET",
		path:   "/2012-12-12/hostedzone",
	}

	xmlRes := &ListHostedZonesResponse{}

	zones := []HostedZone{}

	if err := r53.run(req, xmlRes); err != nil {
		return []HostedZone{}, err
	}
	zones = append(zones, xmlRes.HostedZones...)

	for xmlRes.IsTruncated {
		req.params = &url.Values{
			"marker": []string{xmlRes.NextMarker},
		}

		if err := r53.run(req, xmlRes); err != nil {
			return []HostedZone{}, err
		}
		zones = append(zones, xmlRes.HostedZones...)
	}

	for _, zone := range zones {
		zone.r53 = r53
	}

	return zones, nil
}

func (r53 *Route53) DeleteHostedZone(id string) (ChangeInfo, error) {
	req := request{
		method: "DELETE",
		path:   fmt.Sprintf("/2012-12-12/hostedzone/%s", strings.Replace(id, "/hostedzone/", "", -1)),
	}

	xmlRes := &DeleteHostedZoneResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return ChangeInfo{}, err
	}
	xmlRes.ChangeInfo.r53 = r53

	return xmlRes.ChangeInfo, nil
}
