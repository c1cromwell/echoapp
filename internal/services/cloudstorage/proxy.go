package cloudstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// RemoteFile is a cloud file listing entry.
type RemoteFile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time,omitempty"`
}

// ListFiles returns files from the connected provider.
func ListFiles(ctx context.Context, provider Provider, accessToken string) ([]RemoteFile, error) {
	if os.Getenv("CLOUD_OAUTH_STUB") == "true" {
		return []RemoteFile{
			{ID: "stub-1", Name: "report.pdf", MimeType: "application/pdf", Size: 4096},
			{ID: "stub-2", Name: "photo.jpg", MimeType: "image/jpeg", Size: 8192},
		}, nil
	}
	switch provider {
	case GoogleDrive:
		return listGoogleDrive(ctx, accessToken)
	case Dropbox:
		return listDropbox(ctx, accessToken)
	case OneDrive:
		return listOneDrive(ctx, accessToken)
	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}

// StreamFile downloads file bytes from the provider (streamed through backend proxy).
func StreamFile(ctx context.Context, provider Provider, accessToken, fileID string) ([]byte, string, error) {
	if os.Getenv("CLOUD_OAUTH_STUB") == "true" {
		return []byte("echo-cloud-stub-content"), "application/octet-stream", nil
	}
	switch provider {
	case GoogleDrive:
		return downloadGoogleDrive(ctx, accessToken, fileID)
	case Dropbox:
		return downloadDropbox(ctx, accessToken, fileID)
	case OneDrive:
		return downloadOneDrive(ctx, accessToken, fileID)
	default:
		return nil, "", fmt.Errorf("unsupported provider")
	}
}

func listGoogleDrive(ctx context.Context, token string) ([]RemoteFile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/drive/v3/files?pageSize=25&fields=files(id,name,mimeType,size,modifiedTime)&q=trashed=false", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("drive list: %s", string(b))
	}
	var out struct {
		Files []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			MimeType     string `json:"mimeType"`
			Size         string `json:"size"`
			ModifiedTime string `json:"modifiedTime"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	files := make([]RemoteFile, 0, len(out.Files))
	for _, f := range out.Files {
		if strings.Contains(f.MimeType, "folder") {
			continue
		}
		var size int64
		fmt.Sscanf(f.Size, "%d", &size)
		files = append(files, RemoteFile{
			ID: f.ID, Name: f.Name, MimeType: f.MimeType, Size: size, ModifiedTime: f.ModifiedTime,
		})
	}
	return files, nil
}

func downloadGoogleDrive(ctx context.Context, token, fileID string) ([]byte, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/drive/v3/files/"+fileID+"?alt=media", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("drive download: %s", string(data))
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	return data, mime, nil
}

func listDropbox(ctx context.Context, token string) ([]RemoteFile, error) {
	body := `{"path":"","recursive":false,"include_media_info":false}`
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Entries []struct {
			Tag         string `json:".tag"`
			ID          string `json:"id"`
			Name        string `json:"name"`
			Size        int64  `json:"size"`
			ClientMtime string `json:"client_modified"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	files := make([]RemoteFile, 0)
	for _, e := range out.Entries {
		if e.Tag != "file" {
			continue
		}
		files = append(files, RemoteFile{ID: e.ID, Name: e.Name, Size: e.Size, ModifiedTime: e.ClientMtime})
	}
	return files, nil
}

func downloadDropbox(ctx context.Context, token, fileID string) ([]byte, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://content.dropboxapi.com/2/files/download", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Dropbox-API-Arg", fmt.Sprintf(`{"path":%q}`, fileID))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	return data, mime, nil
}

func listOneDrive(ctx context.Context, token string) ([]RemoteFile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.microsoft.com/v1.0/me/drive/root/children?$top=25", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Value []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Size int64  `json:"size"`
			File *struct {
				MimeType string `json:"mimeType"`
			} `json:"file"`
			LastModifiedDate string `json:"lastModifiedDateTime"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	files := make([]RemoteFile, 0)
	for _, v := range out.Value {
		if v.File == nil {
			continue
		}
		files = append(files, RemoteFile{
			ID: v.ID, Name: v.Name, MimeType: v.File.MimeType, Size: v.Size, ModifiedTime: v.LastModifiedDate,
		})
	}
	return files, nil
}

func downloadOneDrive(ctx context.Context, token, fileID string) ([]byte, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.microsoft.com/v1.0/me/drive/items/"+fileID+"/content", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	return data, mime, nil
}
