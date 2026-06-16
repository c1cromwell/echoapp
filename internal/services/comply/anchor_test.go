package comply

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func TestMetagraphAnchor_SubmitAndHealth(t *testing.T) {
	var posted bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions" {
			http.NotFound(w, r)
			return
		}
		posted = true
		_ = json.NewEncoder(w).Encode(map[string]string{"txHash": "tx-abc"})
	}))
	defer srv.Close()

	client := metagraph.NewMetagraphClient(metagraph.MetagraphConfig{
		DataL1URL: srv.URL,
		Timeout:   2 * time.Second,
	})
	anchor := NewMetagraphAnchor(client)

	commitment := policyAnchorRef("did:org:1", database.PolicyPermanent, "c1", "scope")
	tx, err := anchor.SubmitComplianceAnchor(context.Background(), commitment, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	if !posted {
		t.Fatal("expected Data L1 POST")
	}
	if tx != "tx-abc" {
		t.Fatalf("tx = %q", tx)
	}
	if anchor.Status() != "healthy" {
		t.Fatalf("status = %q", anchor.Status())
	}
}

func TestComplianceAnchor_FallsBackWithoutDataL1(t *testing.T) {
	svc := NewService(database.NewMemoryDB(), Deps{})
	ref := svc.complianceAnchor(context.Background(), "did:org:1", "retention", "p1", "scope")
	if len(ref) != 64 {
		t.Fatalf("expected 64-char commitment, got %q", ref)
	}
}

func TestAnchorHealthStatus_NoDataL1(t *testing.T) {
	svc := NewService(database.NewMemoryDB(), Deps{})
	if svc.anchorHealthStatus() != "healthy" {
		t.Fatalf("expected healthy local mode")
	}
}
