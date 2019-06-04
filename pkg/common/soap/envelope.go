package soap

import (
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
)

// Envelope is used to unmarshal service info out of X-road request
type Envelope struct {
	SubsystemCode  string `xml:"Header>service>subsystemCode"`
	ServiceCode    string `xml:"Header>service>serviceCode"`
	ServiceVersion string `xml:"Header>service>serviceVersion"`
	Service        string
}

// FromRequestBody unmarshals request body bytes to envelope
func FromRequestBody(requestBody []byte) (Envelope, error) {
	s := Envelope{}

	err := xml.Unmarshal(requestBody, &s)
	if err != nil {
		return Envelope{}, errors.Wrap(err, "failed to unmarshal SOAP envelope data")
	}

	s.Service = fmt.Sprintf("%v.%v.%v", s.SubsystemCode, s.ServiceCode, s.ServiceVersion)

	return s, nil
}
