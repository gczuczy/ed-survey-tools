package auth

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

type Config struct {
	Issuer string `json:"issuer"`
	ClientID string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	AuthURL string `json:"authUrl"`
	TokenURL string `json:"tokenUrl"`
	Scope string `json:"scope"`
}

func configHandler(r *http.Request) wrappers.IResponse {

	var host string
	if r.Host == "localhost" || r.Host == "127.0.0.1" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1:") {
		host = fmt.Sprintf("http://%s", r.Host)
	} else {
		host = fmt.Sprintf("https://%s", r.Host)
	}

	return wrappers.Success(Config{
		Issuer: config.Issuer,
		ClientID: config.ClientID,
		RedirectURI: fmt.Sprintf("%s/api/auth/callback", host),
		Scope: strings.Join(oauth2config.Scopes, " "),
		AuthURL: config.AuthorizeURL,
		TokenURL: config.TokenURL,
	})
}
