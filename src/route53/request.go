package route53

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type request struct {
	method string
	path   string
	params *url.Values
	body   interface{}
}

func (r *request) url() *url.URL {
	url, _ := url.Parse("https://route53.amazonaws.com")
	url.Path = r.path

	// Most requests don't have params.
	if r.params != nil {
		url.RawQuery = r.params.Encode()
	}

	return url
}

type errorResponse struct {
	Type      string `xml:"Error>Type"`
	Code      string `xml:"Error>Code"`
	Message   string `xml:"Error>Message"`
	RequestId string
}

func (r53 *Route53) run(req request, res interface{}) error {
	hreq := &http.Request{
		Method:     req.method,
		URL:        req.url(),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	sign(r53.auth, hreq)

	if debug {
		fmt.Fprintf(os.Stderr, "-- request\n%+v\n\n", hreq)
	}

	if req.body != nil {
		data, err := xml.Marshal(req.body)
		if err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "-- error marshalling\n%s\n%+v\n", err, req.body)
			}
			return err
		}

		if debug {
			ppBody, _ := xml.MarshalIndent(req.body, " ", "    ")
			fmt.Fprintf(os.Stderr, "-- body\n%s\n\n", xml.Header+string(ppBody))
		}

		hreq.Body = ioutil.NopCloser(bytes.NewBufferString(xml.Header + string(data)))
	}

	hres, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return err
	}
	defer hres.Body.Close()

	if debug {
		fmt.Fprintf(os.Stderr, "-- response\n%+v\n\n", hres)
	}

	body, err := ioutil.ReadAll(hres.Body)
	if err != nil {
		return err
	}

	bodyReadCloser := ioutil.NopCloser(bytes.NewReader(body))

	if hres.StatusCode != 200 {
		eres := errorResponse{}

		err := xml.NewDecoder(bodyReadCloser).Decode(&eres)
		if err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "-- error unmarshalling\n%s\n%s\n\n", err, string(body))
			}
			return fmt.Errorf("could not parse: %s", string(body))
		} else {
			if debug {
				ppBody, _ := xml.MarshalIndent(eres, " ", "    ")
				fmt.Fprintf(os.Stderr, "-- body\n%s\n\n", string(ppBody))
			}
			return fmt.Errorf("%s: %s", eres.Code, eres.Message)
		}
	}

	err = xml.NewDecoder(bodyReadCloser).Decode(res)

	if debug {
		if err != nil {
			// Decode error, cannot pretty print this response.
			fmt.Fprintf(os.Stderr, "-- error unmarshalling\n%s\n%s\n\n", err, string(body))
		} else {
			ppBody, _ := xml.MarshalIndent(res, " ", "    ")
			fmt.Fprintf(os.Stderr, "-- body\n%s\n\n", string(ppBody))
		}
	}

	return err
}
