package edsm

import (
	"net/url"
	"net/http"
	"encoding/json"
)

const urlBase = "https://www.edsm.net"

type EDSM struct {
	client *http.Client
}

func New() *EDSM {
	return &EDSM{
		client: &http.Client{},
	}
}

func (e *EDSM) newRequest(method string, endpoint string) (req *http.Request, err error) {
	req, err = http.NewRequest(method, urlBase, nil)
	var (
		base *url.URL
	)
	if base, err = url.Parse(urlBase); err != nil {
		return
	}

	rel := &url.URL{Path: endpoint}
	req.URL = base.ResolveReference(rel)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Usage-Agent", "ed-survey-tools")

	return
}

func (e *EDSM) call(req *http.Request, v any) (resp *http.Response, err error) {

	c := http.Client{}
	if resp, err = c.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return
}

