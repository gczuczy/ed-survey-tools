package auth

import (
	"fmt"
	"errors"
	"net/url"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/capi"
)

func callbackHandler(r *wrappers.Request) wrappers.IResponse {
	var (
		userinfo *Userinfo
		err error
	)
	ctx := ctxWithClient(r.R.Context())

	code := r.R.URL.Query().Get("code")
	state := r.R.URL.Query().Get("state")

	if len(code) == 0 {
		return wrappers.NewError(
			fmt.Errorf("Missing authorization code"),
			http.StatusBadRequest)
	}

	cfg := *oauth2config
	cfg.RedirectURL = fmt.Sprintf("%s/api/auth/callback", hostURL(r.R))

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		r.L.Debug().Interface("oauth2config", cfg).Msg("Oauth2 config")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Code exchange failed")),
			http.StatusInternalServerError)
	}

	client := cfg.Client(ctx, token)
	if userinfo, err = getUserinfo(client); err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed to fetch userinfo")),
			http.StatusInternalServerError)
	}

	s, _ := sessions.Get(r.R)
	if userinfo.User == nil || userinfo.User.CustomerID == 0 {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Userinfo missing customer_id")),
			http.StatusInternalServerError)
	}
	// now callout to CAPI for cmdr name
	capicl, err := capi.New(token.AccessToken)
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Unable to get CAPI client")),
			http.StatusInternalServerError)
	}

	p, err := capicl.GetProfile()
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Unable to get Profile from CAPI")),
			http.StatusInternalServerError)
	}

	// final sanity check
	if len(p.Commander.Name)==0 || userinfo.User.CustomerID == 0 {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Out Of Cheese error")),
			http.StatusInternalServerError)
	}

	user, err := db.Pool.LoginCMDR(p.Commander.Name, userinfo.User.CustomerID)
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed login")),
			http.StatusInternalServerError)
	}
	s.Values["user"] = user

	v := url.Values{}
	v.Add("code", code)
	v.Add("state", state)
	rurl := url.URL{
		Path: "/",
		RawQuery: v.Encode(),
	}
	ret := wrappers.NewRedirect(&rurl, http.StatusFound)
	ret.Session(s)

	return ret
}
