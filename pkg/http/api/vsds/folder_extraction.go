package vsds

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// FolderExtractionSheetResp is one failing tab in the extraction summary.
type FolderExtractionSheetResp struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// FolderExtractionDocumentResp is one document with errors
// in the extraction summary.
type FolderExtractionDocumentResp struct {
	ID          int                          `json:"id"`
	GCPID       string                       `json:"gcpid"`
	Name        string                       `json:"name"`
	ContentType string                       `json:"content_type"`
	ErrorCount  int                          `json:"error_count"`
	Sheets      []FolderExtractionSheetResp  `json:"sheets"`
}

// FolderExtractionSummaryResp is the response type for
// GET /api/vsds/folders/{id}/extraction
type FolderExtractionSummaryResp struct {
	FolderName string                         `json:"folder_name"`
	Stats      db.FolderProcessingSummary     `json:"stats"`
	Documents  []FolderExtractionDocumentResp `json:"documents"`
}

func getFolderExtractionSummary(
	r *wrappers.Request,
) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing folder ID"),
			http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid folder ID"),
			http.StatusBadRequest)
	}

	folderName, summary, sheetRows, err :=
		db.Pool.GetFolderProcessingDetails(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("No processing run found for folder"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).
			Msg("Error while querying folder extraction")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while querying folder extraction"),
			),
			http.StatusInternalServerError)
	}

	// Group failing sheet rows by document, preserving order.
	docMap := make(map[int]*FolderExtractionDocumentResp)
	docOrder := make([]int, 0)
	for _, row := range sheetRows {
		doc, exists := docMap[row.DocID]
		if !exists {
			doc = &FolderExtractionDocumentResp{
				ID:          row.DocID,
				GCPID:       row.DocGCPID,
				Name:        row.DocName,
				ContentType: row.ContentType,
				Sheets:      []FolderExtractionSheetResp{},
			}
			docMap[row.DocID] = doc
			docOrder = append(docOrder, row.DocID)
		}
		doc.ErrorCount++
		doc.Sheets = append(doc.Sheets, FolderExtractionSheetResp{
			ID:      row.SheetID,
			Name:    row.SheetName,
			Message: row.Message,
		})
	}

	docs := make([]FolderExtractionDocumentResp, 0, len(docOrder))
	for _, docID := range docOrder {
		docs = append(docs, *docMap[docID])
	}

	return wrappers.Success(FolderExtractionSummaryResp{
		FolderName: folderName,
		Stats:      summary,
		Documents:  docs,
	})
}
