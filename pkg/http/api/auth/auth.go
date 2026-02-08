package auth

import (
	"golang.org/x/oauth2"
	"github.com/gorilla/mux"

	c "github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

var (
	config *c.OAuth2Config
	oauth2config *oauth2.Config
)

func Init(r *mux.Router, cfg *c.OAuth2Config) error {
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

 	r.HandleFunc("/config", wrappers.Wrap(configHandler))
 	r.HandleFunc("/callback", wrappers.Wrap(callbackHandler))

	return nil
}
