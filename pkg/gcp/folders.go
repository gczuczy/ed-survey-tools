package gcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type FolderItem struct {
	ID           string
	Name         string
	MimeType     string
	CreatedTime  string
	ModifiedTime string
}

const driveFolderMimeType = "application/vnd.google-apps.folder"

func DownloadFile(fileID string) ([]byte, error) {
	ctx := context.Background()
	scopes := option.WithScopes(drive.DriveReadonlyScope)
	svc, err := drive.NewService(ctx, authOption, scopes)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create drive service: %w", err)
	}

	resp, err := RateLimit(func() (*http.Response, error) {
		return svc.Files.Get(fileID).Download()
	}, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download file %s: %w", fileID, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read file %s: %w", fileID, err)
	}
	return data, nil
}

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

func ListFolder(folderID string) ([]FolderItem, error) {
	ctx := context.Background()
	scopes := option.WithScopes(drive.DriveMetadataReadonlyScope)
	svc, err := drive.NewService(ctx, authOption, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	query := fmt.Sprintf(
		"'%s' in parents and trashed = false and mimeType != '%s'",
		folderID, driveFolderMimeType)

	const fields = "nextPageToken, " +
		"files(id, name, mimeType, createdTime, modifiedTime)"

	var (
		items     []FolderItem
		pageToken string
	)
	for {
		call := svc.Files.List().
			Q(query).
			Fields(fields).
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		page, pErr := RateLimit(func() (*drive.FileList, error) {
			return call.Do()
		}, 30*time.Second)
		if pErr != nil {
			return nil, fmt.Errorf(
				"failed to list folder %s: %w", folderID, pErr)
		}
		for _, f := range page.Files {
			items = append(items, FolderItem{
				ID:           f.Id,
				Name:         f.Name,
				MimeType:     f.MimeType,
				CreatedTime:  f.CreatedTime,
				ModifiedTime: f.ModifiedTime,
			})
		}
		pageToken = page.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return items, nil
}
