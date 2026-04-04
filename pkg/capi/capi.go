package capi

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type Client struct {
	accessToken string
	host string
	httpClient *http.Client
}

type Profile struct {
	Commander ProfileCommander `json:"commander"`
}

type ProfileCommander struct {
	Name string `json:"name"`
	ID int64 `json:"id"`
}

func New(accessToken string) (*Client, error) {
	cl, err :=NewHTTPClient("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	if err != nil {
		return nil, err
	}
	return &Client{
		accessToken: accessToken,
		host: "companion.orerve.net",
		httpClient: cl,
	}, nil
}

// returns a page, data must be passed by pointer
func (c *Client) request(uri string, data any) error {
	url := fmt.Sprintf("https://%s%s", c.host, uri)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while fetching profile: %s", resp.Status)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(data)
}

func (c *Client) GetProfile() (*Profile, error) {
	p := Profile{}
	if err := c.request("/profile", &p); err != nil {
		return nil, err
	}
	return &p, nil
}
