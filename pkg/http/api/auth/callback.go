package auth

import (
	"io"
	"fmt"
	"errors"
	"strconv"
	"net/url"
	"net/http"
	"encoding/json"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
)

func callbackHandler(r *wrappers.Request) wrappers.IResponse {
	ctx := r.R.Context()

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
		fmt.Printf("Oauth2config: %s\n", cfg)
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Code exchange failed")),
			http.StatusInternalServerError)
	}

	client := cfg.Client(ctx, token)
	resp, err := client.Get(config.UserInfoURL)
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed to fetch userinfo")),
			http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed to read userinfo")),
			http.StatusInternalServerError)
	}

	var userinfo map[string]any
	if err := json.Unmarshal(body, &userinfo); err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed to parse userinfo")),
			http.StatusInternalServerError)
	}

	// TODO: session
	fmt.Printf("userinfo: %v\n", userinfo)
	s, _ := sessions.Get(r.R)
	var (
		customerid int64
		cmdrname string
	)
	if fdevcid, ok := userinfo["customer_id"]; ok {
		// TODO: acquire CMDR name from CAPI
		cmdrname = "CAPI-once-we-have-the-oauth2-apps-approved"
		customerid, err = strconv.ParseInt(fdevcid.(string), 10, 64)
		if err != nil {
			return wrappers.NewError(
				errors.Join(err, fmt.Errorf("Failed to parse FDev custoemr_id")),
				http.StatusInternalServerError)
		}
	} else {
		// testing IdP
		cmdrname = userinfo["sub"].(string)
		customerid = 42069
	}

	user, err := db.Pool.LoginCMDR(cmdrname, customerid)
	if err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed login")),
			http.StatusInternalServerError)
	}
	fmt.Printf("User to %+v\n", user)
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
