package edsm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
)

const urlBase = "https://www.edsm.net"

type EDSM struct {
	client  *http.Client
	retries int
}

func New(cfg *config.EDSMConfig) *EDSM {
	return &EDSM{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		retries: cfg.Retries,
	}
}

func (e *EDSM) newRequest(
	method string,
	endpoint string,
) (req *http.Request, err error) {
	var base *url.URL
	if base, err = url.Parse(urlBase); err != nil {
		return
	}

	rel := &url.URL{Path: endpoint}
	if req, err = http.NewRequest(method, base.ResolveReference(rel).String(), nil); err != nil {
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Usage-Agent", "ed-survey-tools")

	return
}

func (e *EDSM) call(
	req *http.Request,
	v any,
) (resp *http.Response, err error) {
	for attempt := 0; attempt <= e.retries; attempt++ {
		resp, err = e.client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			err = fmt.Errorf("unexpected status %d", resp.StatusCode)
			continue
		}
		break
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return
}
