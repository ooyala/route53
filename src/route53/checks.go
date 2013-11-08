package route53

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// XML RPC types.

type CreateHealthCheckRequest struct {
	XMLName           xml.Name `xml:"CreateHealthCheckRequest"`
	CallerReference   string
	HealthCheckConfig HealthCheckConfig
}

type HealthCheckConfig struct {
	IPAddress                string
	Port                     uint16
	Type                     string
	ResourcePath             string
	FullyQualifiedDomainName string
}

type CreateHealthCheckResponse struct {
	XMLName     xml.Name `xml:"CreateHealthCheckResponse"`
	HealthCheck HealthCheck
}

type HealthCheck struct {
	Id                string
	CallerReference   string
	HealthCheckConfig HealthCheckConfig
}

type GetHealthCheckResponse struct {
	XMLName     xml.Name `xml:"GetHealthCheckResponse"`
	HealthCheck HealthCheck
}

type ListHealthChecksResponse struct {
	XMLName      xml.Name `xml:"ListHealthChecksResponse"`
	HealthChecks []HealthCheck
	IsTruncated  bool
	Marker       string
	NextMarker   string
	MaxItems     uint
}

type DeleteHealthCheckResponse struct {
	XMLName xml.Name `xml:"DeleteHealthCheckResponse"`
}

// Route53 API requests.

func (r53 *Route53) CreateHealthCheck(config HealthCheckConfig, reference string) (string, error) {
	xmlReq := &CreateHealthCheckRequest{
		CallerReference:   reference,
		HealthCheckConfig: config,
	}

	req := request{
		method: "POST",
		path:   "/2012-12-12/healthcheck",
		body:   xmlReq,
	}

	xmlRes := &CreateHealthCheckResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return "invalid", err
	}

	return xmlRes.HealthCheck.Id, nil
}

func (r53 *Route53) GetHealthCheck(id string) (HealthCheck, error) {
	req := request{
		method: "GET",
		path:   fmt.Sprintf("/2012-12-12/healthcheck/%s", id),
	}

	xmlRes := &GetHealthCheckResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return HealthCheck{}, err
	}

	return xmlRes.HealthCheck, nil
}

func (r53 *Route53) ListHealthChecks() ([]HealthCheck, error) {
	req := request{
		method: "GET",
		path:   "/2012-12-12/healthcheck",
	}

	xmlRes := &ListHealthChecksResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return []HealthCheck{}, err
	}
	if xmlRes.IsTruncated {
		return []HealthCheck{}, errors.New("cannot handle truncated response")
	}

	return xmlRes.HealthChecks, nil
}

func (r53 *Route53) DeleteHealthCheck(id string) error {
	req := request{
		method: "DELETE",
		path:   fmt.Sprintf("/2012-12-12/healthcheck/%s", id),
	}

	xmlRes := &DeleteHealthCheckResponse{}

	if err := r53.run(req, xmlRes); err != nil {
		return err
	}

	return nil
}
