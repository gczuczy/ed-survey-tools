package auth

import (
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func logoutHandler(r *wrappers.Request) wrappers.IResponse {
	r.S.Options.MaxAge = -1

	resp := wrappers.Success(nil)
	resp.Session(r.S)
	return resp
}
