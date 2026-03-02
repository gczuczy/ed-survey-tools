package vsds

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// FolderListResponse is the response type for GET /api/vsds/folders
type FolderListResponse = []db.VSDSFolder

// FolderResponse is the response type for POST /api/vsds/folders
type FolderResponse = db.VSDSFolder

func listFolders(r *wrappers.Request) wrappers.IResponse {
	folders, err := db.Pool.ListFolders()
	if err != nil {
		r.L.Error().Err(err).Msg("Error while querying folders")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while querying folders")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(folders)
}

func addFolder(r *wrappers.Request) wrappers.IResponse {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	if body.URL == "" {
		return wrappers.NewError(
			fmt.Errorf("url is required"),
			http.StatusBadRequest)
	}

	gcpid, name, err := gcp.ValidateFolder(body.URL)
	if err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid folder URL: %v", err),
			http.StatusBadRequest)
	}

	folder, err := db.Pool.AddFolder(gcpid, name)
	if err != nil {
		if errors.Is(err, db.ErrDuplicate) {
			return wrappers.NewError(fmt.Errorf("Folder already exists"), http.StatusConflict)
		}
		r.L.Error().Err(err).Msg("Error while adding folder")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while adding folder")),
			http.StatusInternalServerError)
	}

	return wrappers.Success(folder)
}

func deleteFolder(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing folder ID"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(fmt.Errorf("Invalid folder ID"), http.StatusBadRequest)
	}

	if err := db.Pool.DeleteFolder(id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(fmt.Errorf("Folder not found"), http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while deleting folder")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while deleting folder")),
			http.StatusInternalServerError)
	}

	return wrappers.Success(nil)
}
