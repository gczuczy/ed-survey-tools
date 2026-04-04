package capi

import (
	"fmt"
	"net/http"
)

type userAgentTransport struct {
	base http.RoundTripper
	userAgent string
	additionalHeaders map[string]string
}

func (uat *userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.Header.Set("User-Agent", uat.userAgent)
	for k,v := range uat.additionalHeaders {
		r2.Header.Set(k, v)
	}
	return uat.base.RoundTrip(r2)
}

func NewHTTPClient(headers ...string) (*http.Client, error) {
	// parse headers vararg into map
	aheaders := make(map[string]string, 0)

	if len(headers) & 1 != 0 {
		return nil, fmt.Errorf("headers arg takes even args for key,value pairs")
	}
	for i:=0; i<len(headers)/2; i += 1 {
		aheaders[headers[i]] = headers[i+1]
	}

	// return the crafted client
	return &http.Client{
		Transport: &userAgentTransport{
			base: http.DefaultTransport,
			userAgent: "EDCD-EDSurveyTools-0.1.0",
			additionalHeaders: aheaders,
		},
	}, nil
}
