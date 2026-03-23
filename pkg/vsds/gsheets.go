package vsds

import (
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
)

func openGoogleSheets(
	gss *gcp.GSpreadsheetsService,
	item gcp.FolderItem,
) ([]gcp.Sheet, error) {
	ss, err := gss.Sheet(item.ID)
	if err != nil {
		return nil, err
	}
	return ss.GetSheets()
}
