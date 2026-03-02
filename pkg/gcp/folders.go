package gcp

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const driveFolderMimeType = "application/vnd.google-apps.folder"

func ValidateFolder(folderURL string) (string, string, error) {
	u, err := url.Parse(folderURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid folder URL: %w", err)
	}

	if !strings.Contains(u.Path, "/drive/folders/") {
		return "", "", fmt.Errorf("URL is not a Google Drive folder URL")
	}

	folderID := path.Base(u.Path)
	if folderID == "" || folderID == "." || folderID == "/" {
		return "", "", fmt.Errorf("folder ID not found in URL")
	}

	ctx := context.Background()
	scopes := option.WithScopes(drive.DriveMetadataReadonlyScope)
	svc, err := drive.NewService(ctx, authOption, scopes)
	if err != nil {
		return "", "", fmt.Errorf("failed to create drive service: %w", err)
	}

	file, err := svc.Files.Get(folderID).
		Fields("name,mimeType").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to get folder %s: %w", folderID, err)
	}

	if file.MimeType != driveFolderMimeType {
		return "", "", fmt.Errorf("%s is not a folder", folderID)
	}

	return folderID, file.Name, nil
}
