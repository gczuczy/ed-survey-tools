package auth

import (
	"fmt"
	"net/http"
	"strings"

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
	host := hostURL(r)

	return wrappers.Success(Config{
		Issuer: config.Issuer,
		ClientID: config.ClientID,
		RedirectURI: fmt.Sprintf("%s/api/auth/callback", host),
		Scope: strings.Join(oauth2config.Scopes, " "),
		AuthURL: config.AuthorizeURL,
		TokenURL: config.TokenURL,
	})
}
