package auth

import (
	"io"
	"fmt"
	"errors"
	"net/url"
	"net/http"
	"encoding/json"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
)

func callbackHandler(r *http.Request) wrappers.IResponse {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if len(code) == 0 {
		return wrappers.NewError(
			fmt.Errorf("Missing authorization ode"),
			http.StatusBadRequest)
	}

	token, err := oauth2config.Exchange(ctx, code)
	if err != nil {
		return wrappers.NewError(
			fmt.Errorf("Code exchange failed"),
			http.StatusInternalServerError)
	}

	client := oauth2config.Client(ctx, token)
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
	s, _ := sessions.Get(r)
	if fdevcid, ok := userinfo["customer_id"]; ok {
		s.Values["fdev_customerid"] = userinfo["customer_id"]
	} else {
		// testing IdP
		s.Values["fdev_custemerid"] = 42069
	}

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
