package cloudstorage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OAuthConfig holds provider client credentials from environment.
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	DropboxAppKey      string
	DropboxAppSecret   string
	MicrosoftClientID  string
	MicrosoftSecret    string
}

// LoadOAuthConfig reads cloud OAuth env vars (WO-46).
func LoadOAuthConfig() OAuthConfig {
	return OAuthConfig{
		GoogleClientID:     os.Getenv("GOOGLE_DRIVE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"),
		DropboxAppKey:      os.Getenv("DROPBOX_APP_KEY"),
		DropboxAppSecret:   os.Getenv("DROPBOX_APP_SECRET"),
		MicrosoftClientID:  os.Getenv("ONEDRIVE_CLIENT_ID"),
		MicrosoftSecret:    os.Getenv("ONEDRIVE_CLIENT_SECRET"),
	}
}

// AuthURL builds the provider authorization URL with client_id when configured.
func AuthURLWithConfig(provider Provider, redirectURI string, cfg OAuthConfig) string {
	encRedirect := url.QueryEscape(redirectURI)
	switch provider {
	case GoogleDrive:
		clientID := cfg.GoogleClientID
		if clientID == "" {
			clientID = "echo-google-stub"
		}
		return fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
			url.QueryEscape(clientID), encRedirect, url.QueryEscape("https://www.googleapis.com/auth/drive.readonly"),
		)
	case Dropbox:
		key := cfg.DropboxAppKey
		if key == "" {
			key = "echo-dropbox-stub"
		}
		return fmt.Sprintf(
			"https://www.dropbox.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&token_access_type=offline",
			url.QueryEscape(key), encRedirect,
		)
	case OneDrive:
		clientID := cfg.MicrosoftClientID
		if clientID == "" {
			clientID = "echo-onedrive-stub"
		}
		return fmt.Sprintf(
			"https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			url.QueryEscape(clientID), encRedirect, url.QueryEscape("Files.Read offline_access"),
		)
	default:
		return ""
	}
}

// TokenResponse is normalized OAuth token output.
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// ExchangeCode trades an authorization code for tokens.
func ExchangeCode(provider Provider, code, redirectURI string, cfg OAuthConfig) (*TokenResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("empty authorization code")
	}
	switch provider {
	case GoogleDrive:
		return postFormToken("https://oauth2.googleapis.com/token", url.Values{
			"code":          {code},
			"client_id":     {cfg.GoogleClientID},
			"client_secret": {cfg.GoogleClientSecret},
			"redirect_uri":  {redirectURI},
			"grant_type":    {"authorization_code"},
		})
	case Dropbox:
		return postFormToken("https://api.dropboxapi.com/oauth2/token", url.Values{
			"code":          {code},
			"grant_type":    {"authorization_code"},
			"redirect_uri":  {redirectURI},
			"client_id":     {cfg.DropboxAppKey},
			"client_secret": {cfg.DropboxAppSecret},
		})
	case OneDrive:
		return postFormToken("https://login.microsoftonline.com/common/oauth2/v2.0/token", url.Values{
			"code":          {code},
			"client_id":     {cfg.MicrosoftClientID},
			"client_secret": {cfg.MicrosoftSecret},
			"redirect_uri":  {redirectURI},
			"grant_type":    {"authorization_code"},
		})
	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}

func postFormToken(endpoint string, form url.Values) (*TokenResponse, error) {
	if os.Getenv("CLOUD_OAUTH_STUB") == "true" {
		return &TokenResponse{
			AccessToken:  "stub-access-token",
			RefreshToken: "stub-refresh-token",
			ExpiresIn:    3600,
		}, nil
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}
	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.AccessToken == "" {
		return nil, fmt.Errorf("missing access_token")
	}
	return &TokenResponse{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		ExpiresIn:    raw.ExpiresIn,
	}, nil
}

// AccessToken returns a valid token, refreshing when expired.
func (s *Service) AccessToken(did string, provider Provider, cfg OAuthConfig) (string, error) {
	s.mu.RLock()
	tok, ok := s.tokens[did][provider]
	s.mu.RUnlock()
	if !ok || tok.AccessToken == "" {
		return "", fmt.Errorf("provider not connected")
	}
	if tok.ExpiresAt.IsZero() || time.Now().Before(tok.ExpiresAt.Add(-time.Minute)) {
		return tok.AccessToken, nil
	}
	if tok.Refresh == "" {
		return tok.AccessToken, nil
	}
	refreshed, err := refreshToken(provider, tok.Refresh, cfg)
	if err != nil {
		return tok.AccessToken, nil
	}
	tok.AccessToken = refreshed.AccessToken
	if refreshed.RefreshToken != "" {
		tok.Refresh = refreshed.RefreshToken
	}
	if refreshed.ExpiresIn > 0 {
		tok.ExpiresAt = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second)
	}
	s.SaveToken(did, tok)
	return tok.AccessToken, nil
}

func refreshToken(provider Provider, refresh string, cfg OAuthConfig) (*TokenResponse, error) {
	switch provider {
	case GoogleDrive:
		return postFormToken("https://oauth2.googleapis.com/token", url.Values{
			"refresh_token": {refresh},
			"client_id":     {cfg.GoogleClientID},
			"client_secret": {cfg.GoogleClientSecret},
			"grant_type":    {"refresh_token"},
		})
	case Dropbox:
		return postFormToken("https://api.dropboxapi.com/oauth2/token", url.Values{
			"refresh_token": {refresh},
			"grant_type":    {"refresh_token"},
			"client_id":     {cfg.DropboxAppKey},
			"client_secret": {cfg.DropboxAppSecret},
		})
	case OneDrive:
		return postFormToken("https://login.microsoftonline.com/common/oauth2/v2.0/token", url.Values{
			"refresh_token": {refresh},
			"client_id":     {cfg.MicrosoftClientID},
			"client_secret": {cfg.MicrosoftSecret},
			"grant_type":    {"refresh_token"},
		})
	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}
