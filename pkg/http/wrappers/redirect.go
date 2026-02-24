package wrappers

import (
	"fmt"
	"net/url"
	"net/http"
)

type Redirect struct {
	Sessioner
	Loc *url.URL
	Code int
}

func NewRedirect(l *url.URL, c int) *Redirect {
	return &Redirect{
		Loc: l,
		Code: c,
	}
}

func (rd *Redirect) HTTPWrite(w http.ResponseWriter, r *http.Request) error {
	var err error
	code := rd.Code
	if code != http.StatusMovedPermanently || code != http.StatusFound ||
		code != http.StatusSeeOther {
		code = http.StatusFound
	}
	if err = rd.SaveSessions(r, w); err != nil {
		fmt.Printf("Error saving headers: %+v\n", err)
	}
	http.Redirect(w, r, rd.Loc.String(), code)
	return err
}
