package cloudstorage

import (
	"fmt"
	"sync"
	"time"
)

// Provider identifies a cloud file source (WO-46).
type Provider string

const (
	GoogleDrive Provider = "google_drive"
	Dropbox     Provider = "dropbox"
	OneDrive    Provider = "onedrive"
)

// Token is an OAuth access token stored server-side (encrypted at rest in prod).
type Token struct {
	Provider    Provider  `json:"provider"`
	AccessToken string    `json:"access_token"`
	Refresh     string    `json:"refresh_token,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Service stores per-DID cloud integration tokens (MVP in-memory).
type Service struct {
	mu     sync.RWMutex
	tokens map[string]map[Provider]Token // did -> provider -> token
}

// NewService creates a cloud storage integration registry.
func NewService() *Service {
	return &Service{tokens: make(map[string]map[Provider]Token)}
}

// SaveToken stores or replaces a provider token for a DID.
func (s *Service) SaveToken(did string, tok Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens[did] == nil {
		s.tokens[did] = make(map[Provider]Token)
	}
	tok.UpdatedAt = time.Now()
	s.tokens[did][tok.Provider] = tok
}

// ListProviders returns connected providers for a DID.
func (s *Service) ListProviders(did string) []Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.tokens[did]
	out := make([]Provider, 0, len(m))
	for p := range m {
		out = append(out, p)
	}
	return out
}

// Revoke removes a provider token.
func (s *Service) Revoke(did string, provider Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.tokens[did]
	if !ok || m[provider].AccessToken == "" {
		return fmt.Errorf("provider not connected")
	}
	delete(m, provider)
	return nil
}

// AuthURL returns the OAuth authorize URL for a provider (stub).
func AuthURL(provider Provider, redirectURI string) string {
	switch provider {
	case GoogleDrive:
		return "https://accounts.google.com/o/oauth2/v2/auth?scope=https://www.googleapis.com/auth/drive.readonly&redirect_uri=" + redirectURI
	case Dropbox:
		return "https://www.dropbox.com/oauth2/authorize?token_access_type=offline&redirect_uri=" + redirectURI
	case OneDrive:
		return "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?scope=files.read&redirect_uri=" + redirectURI
	default:
		return ""
	}
}
