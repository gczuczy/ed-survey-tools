package vsds

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// ContributionResponse is the response type for
// GET /api/vsds/contribution.
type ContributionResponse = db.CMDRContribution

// ContribErrorSheetResp is one errored tab within a document.
type ContribErrorSheetResp struct {
	SheetName string `json:"sheet_name"`
	Message   string `json:"message"`
}

// ContribErrorDocResp is one document with its errored tabs, keyed
// by (doc_id, received_at) so re-runs appear as separate entries.
type ContribErrorDocResp struct {
	DocID      int                     `json:"doc_id"`
	DocName    string                  `json:"doc_name"`
	ReceivedAt time.Time               `json:"received_at"`
	ErrorCount int                     `json:"error_count"`
	Sheets     []ContribErrorSheetResp `json:"sheets"`
}

// ContributionErrorsResponse is the response type for
// GET /api/vsds/contribution/errors.
type ContributionErrorsResponse = []ContribErrorDocResp

func getContribution(r *w.Request) w.IResponse {
	contrib, err := db.Pool.GetCMDRContribution(r.U.ID)
	if err != nil {
		r.L.Error().Err(err).
			Msg("Error fetching CMDR contribution")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error fetching contribution")),
			http.StatusInternalServerError)
	}
	return w.Success(contrib)
}

func getContributionErrors(r *w.Request) w.IResponse {
	rows, err := db.Pool.GetCMDRSheetErrors(r.U.ID)
	if err != nil {
		r.L.Error().Err(err).
			Msg("Error fetching CMDR sheet errors")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error fetching contribution errors")),
			http.StatusInternalServerError)
	}

	// Group flat rows by (doc_id, received_at), preserving order.
	type docKey struct {
		docID      int
		receivedAt time.Time
	}
	docMap   := make(map[docKey]*ContribErrorDocResp)
	docOrder := make([]docKey, 0)
	for _, row := range rows {
		key := docKey{row.DocID, row.ReceivedAt}
		doc, exists := docMap[key]
		if !exists {
			doc = &ContribErrorDocResp{
				DocID:      row.DocID,
				DocName:    row.DocName,
				ReceivedAt: row.ReceivedAt,
				Sheets:     []ContribErrorSheetResp{},
			}
			docMap[key]  = doc
			docOrder = append(docOrder, key)
		}
		doc.ErrorCount++
		doc.Sheets = append(doc.Sheets,
			ContribErrorSheetResp{
				SheetName: row.SheetName,
				Message:   row.Message,
			})
	}

	docs := make(ContributionErrorsResponse, 0, len(docOrder))
	for _, key := range docOrder {
		docs = append(docs, *docMap[key])
	}
	return w.Success(docs)
}
