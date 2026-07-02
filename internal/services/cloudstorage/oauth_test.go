package cloudstorage

import (
	"testing"
	"time"
)

func TestExchangeCodeStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	tok, err := ExchangeCode(GoogleDrive, "auth-code", "echo://oauth/cloud", OAuthConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestListFilesStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	files, err := ListFiles(t.Context(), GoogleDrive, "stub-token")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 1 {
		t.Fatal("expected stub files")
	}
}

func TestStreamFileStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	data, mime, err := StreamFile(t.Context(), GoogleDrive, "stub-token", "stub-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || mime == "" {
		t.Fatalf("unexpected stream: len=%d mime=%q", len(data), mime)
	}
}

func TestAccessTokenReturnsStored(t *testing.T) {
	s := NewService()
	did := "did:key:alice"
	s.SaveToken(did, Token{
		Provider:    GoogleDrive,
		AccessToken: "live-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	tok, err := s.AccessToken(did, GoogleDrive, LoadOAuthConfig())
	if err != nil || tok != "live-token" {
		t.Fatalf("token: %q err=%v", tok, err)
	}
}
