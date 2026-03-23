package cmdrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// CmdrListResponse is the response type for GET /api/cmdrs
type CmdrListResponse = []*db.User

// CmdrResponse is the response type for PATCH /api/cmdrs/{id}
type CmdrResponse = db.User

func listCmdrs(r *wrappers.Request) wrappers.IResponse {
	cmdrs, err := db.Pool.ListCMDRs()
	if err != nil {
		r.L.Error().Err(err).Msg("Error while listing cmdrs")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while listing cmdrs")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(cmdrs)
}

func setCmdrAdmin(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("missing id"),
			http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(
			fmt.Errorf("invalid id"),
			http.StatusBadRequest)
	}

	var body struct {
		IsAdmin bool `json:"isadmin"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("invalid request body: %v", err),
			http.StatusBadRequest)
	}

	cmdr, err := db.Pool.SetCMDRAdmin(id, body.IsAdmin)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("cmdr not found"),
				http.StatusNotFound)
		}
		if errors.Is(err, db.ErrCustomerIDRequired) {
			return wrappers.NewError(err, http.StatusUnprocessableEntity)
		}
		r.L.Error().Err(err).Msg("Error while updating cmdr admin flag")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while updating cmdr")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(cmdr)
}
