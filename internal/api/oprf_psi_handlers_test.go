package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/mobile/echooprf"
)

// TestPSI_EndToEnd_MobileClientMatchesServer verifies WO-221 mobile/echooprf
// interop with POST /v3/contacts/psi (WO-220).
func TestPSI_EndToEnd_MobileClientMatchesServer(t *testing.T) {
	keyHex, err := contacts.GenerateOPRFKeyHex()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTACT_OPRF_KEY", keyHex)

	oprfSvc, err := contacts.NewOPRFService()
	if err != nil {
		t.Fatal(err)
	}
	db := database.NewMemoryDB()
	contactsSvc := contacts.NewService(db)
	contactsSvc.SetOPRF(oprfSvc)

	phone := "+15559876543"
	friendDID := "did:key:zFriendPSI"
	optIn := true
	if err := db.CreateUser(t.Context(), &database.User{
		DID:                 friendDID,
		TrustTier:           1,
		PhoneDiscoveryOptIn: &optIn,
	}); err != nil {
		t.Fatal(err)
	}
	if err := contactsSvc.RegisterPhoneForDiscovery(t.Context(), friendDID, phone); err != nil {
		t.Fatal(err)
	}

	h := &V3Handlers{DB: db, Contacts: contactsSvc}

	client := echooprf.NewClient()
	blind, err := client.BlindPhones([]string{phone})
	if err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]interface{}{"blinded": blind.Blinded})
	req := httptest.NewRequest(http.MethodPost, "/v3/contacts/psi", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.handleContactsPSI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("psi status = %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Evaluated []string          `json:"evaluated"`
		Index     map[string]string `json:"index"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	keys, err := client.FinalizePhones(blind.SessionID, resp.Evaluated)
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.Index[keys[0]]; got != friendDID {
		t.Fatalf("match = %q, want %q (key %s)", got, friendDID, keys[0])
	}
}
