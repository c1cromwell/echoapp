package cloudstorage

import "testing"

func TestSaveListRevoke(t *testing.T) {
	s := NewService()
	did := "did:key:alice"
	s.SaveToken(did, Token{Provider: GoogleDrive, AccessToken: "tok"})
	providers := s.ListProviders(did)
	if len(providers) != 1 || providers[0] != GoogleDrive {
		t.Fatalf("providers: %+v", providers)
	}
	if err := s.Revoke(did, GoogleDrive); err != nil {
		t.Fatal(err)
	}
	if len(s.ListProviders(did)) != 0 {
		t.Fatal("expected revoke to clear provider")
	}
}

func TestAuthURL(t *testing.T) {
	if AuthURL(GoogleDrive, "echo://callback") == "" {
		t.Fatal("expected google auth url")
	}
}
