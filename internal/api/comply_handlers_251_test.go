package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

func TestLitigationHold_ActivateAndQuery(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	_ = db.Enqueue(context.Background(), &database.QueuedMessage{
		MessageID: "m1", ConversationID: "c1",
		SenderDID: "did:custodian", RecipientDID: "did:peer",
	})

	body := map[string]any{
		"matterId":      "matter-1",
		"custodianDids": []string{"did:custodian"},
		"scope":         "Q2 discovery",
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodPost, "/comply/litigation/hold", "did:org:acme", "tok", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("hold want 201, got %d: %s", rec.Code, rec.Body.String())
	}

	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, complyRequest(http.MethodGet, "/comply/litigation/hold/matter-1", "did:org:acme", "tok", nil))
	if grec.Code != http.StatusOK {
		t.Fatalf("get hold want 200, got %d", grec.Code)
	}
	retained, _ := db.IsConversationRetained(context.Background(), "c1")
	if !retained {
		t.Fatal("hold should retain custodian conversation")
	}
}

func TestEDiscoveryExport_CreateAndPoll(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	_, _ = svc.ActivateLitigationHold(context.Background(), comply.ActivateLitigationHoldInput{
		OrgDID: "did:org:acme", MatterID: "matter-2",
		CustodianDIDs: []string{"did:a"}, ActivatedByDID: "did:admin",
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodPost, "/comply/ediscovery/export", "did:org:acme", "tok",
		map[string]any{"matterId": "matter-2"}))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("export want 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ExportID string `json:"exportId"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	// Poll until ready (async goroutine).
	deadline := httptest.NewRecorder()
	for i := 0; i < 20; i++ {
		deadline = httptest.NewRecorder()
		mux.ServeHTTP(deadline, complyRequest(http.MethodGet, "/comply/ediscovery/export/"+created.ExportID, "did:org:acme", "tok", nil))
		var st struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(deadline.Body.Bytes(), &st)
		if st.Status == "ready" {
			break
		}
	}
	if deadline.Code != http.StatusOK {
		t.Fatalf("poll export want 200, got %d", deadline.Code)
	}
}

func TestAuditReport_JSON(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodGet, "/comply/audit/report?format=json", "did:org:acme", "tok", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("audit want 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("verificationNotice")) {
		t.Fatalf("expected audit report json, got %s", rec.Body.String())
	}
}
