package auth

import (
	"io"
	"fmt"
	"strings"
	"context"
	"net/http"
	"encoding/json"

	"golang.org/x/oauth2"
	"github.com/gorilla/mux"

	c "github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/capi"
)

var (
	config *c.OAuth2Config
	oauth2config *oauth2.Config
	httpClient  *http.Client
)

func hostURL(r *http.Request) string {
	if r.Host == "localhost" || r.Host == "127.0.0.1" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1:") {
		return fmt.Sprintf("http://%s", r.Host)
	}
	return fmt.Sprintf("https://%s", r.Host)
}

func Init(r *mux.Router, cfg *c.OAuth2Config) error {
	var err error

	config = cfg

	oauth2config = &oauth2.Config{
		ClientID: config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL: config.AuthorizeURL,
			TokenURL: config.TokenURL,
		},
		Scopes: append([]string{"auth", "capi"}, config.ExtraScopes...),
	}

	// inject the User-Agent equipped http client
	if httpClient, err = capi.NewHTTPClient(); err != nil {
		return err
	}

	r.Handle("/config", wrappers.NewAPIHandler().Get(configHandler))
	r.Handle("/callback", wrappers.NewAPIHandler().Get(callbackHandler))
	r.Handle("/logout", wrappers.NewAPIHandler().AuthGet(logoutHandler))
	r.Handle("/user", wrappers.NewAPIHandler().AuthGet(userinfoHandler))
	r.Handle("/me", wrappers.NewAPIHandler().AuthDelete(deleteMeHandler))

	return nil
}

func ctxWithClient(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, httpClient)
}

type Userinfo struct {
	Issuer string `json:"iss"`
	IssuedAt uint64 `json:"iat"`
	Expiry uint64 `json:"exp"`
	Sub string `json:"sub"`
	Scope string `json:"scope"`
	User *UserinfoUser `json:"usr"`
}

type UserinfoUser struct {
	CustomerID int64 `json:"customer_id,string"`
	Email string `json:"email"`
	Developer bool `json:"developer"`
	Platform string `json:"platform"`
	Roles []string `json:"roles"`
}

func getUserinfo(client *http.Client) (*Userinfo, error){
	resp, err := client.Get(config.UserInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userinfo Userinfo
	if err := json.Unmarshal(body, &userinfo); err != nil {
		return nil, err
	}
	return &userinfo, nil
}
