package auth

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"github.com/gorilla/mux"

	c "github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

var (
	config *c.OAuth2Config
	oauth2config *oauth2.Config
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

	r.Handle("/config", wrappers.NewAPIHandler().Get(configHandler))
	r.Handle("/callback", wrappers.NewAPIHandler().Get(callbackHandler))
	r.Handle("/logout", wrappers.NewAPIHandler().AuthGet(logoutHandler))
	r.Handle("/user", wrappers.NewAPIHandler().AuthGet(userinfoHandler))
	r.Handle("/me", wrappers.NewAPIHandler().AuthDelete(deleteMeHandler))

	return nil
}
