package auth

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
)

func deleteMeHandler(r *wrappers.Request) wrappers.IResponse {
	if r.U == nil {
		return wrappers.NewError(
			fmt.Errorf("Not authenticated"),
			http.StatusUnauthorized)
	}

	if err := db.Pool.NullifyCMDRCustomerID(r.U.ID); err != nil {
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Failed to unlink account")),
			http.StatusInternalServerError)
	}

	r.S.Options.MaxAge = -1
	resp := wrappers.Success(nil)
	resp.Session(r.S)
	return resp
}
