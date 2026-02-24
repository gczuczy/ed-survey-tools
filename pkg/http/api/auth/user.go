package auth

import (
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func userinfoHandler(r *wrappers.Request) wrappers.IResponse {
	resp := wrappers.Success(r.U)
	return resp
}
