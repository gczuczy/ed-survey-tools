package wrappers

import (
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
	code := rd.Code
	if code != http.StatusMovedPermanently || code != http.StatusFound ||
		code != http.StatusSeeOther {
		code = http.StatusFound
	}
	http.Redirect(w, r, rd.Loc.String(), code)
	return nil
}
