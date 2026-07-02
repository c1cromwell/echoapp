package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func TestMemberCountHashDeterministic(t *testing.T) {
	a := MemberCountHash(5, "group-1")
	b := MemberCountHash(5, "group-1")
	if a != b {
		t.Fatalf("hash not deterministic: %s vs %s", a, b)
	}
	if a == MemberCountHash(6, "group-1") {
		t.Fatal("different counts should differ")
	}
}

func TestGroupAnchoringService_CachesProof(t *testing.T) {
	var submitted bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		submitted = true
		_ = json.NewEncoder(w).Encode(map[string]string{"hash": "tx-group-1"})
	}))
	t.Cleanup(srv.Close)

	client := metagraph.NewMetagraphClient(metagraph.MetagraphConfig{
		DataL1URL: srv.URL,
	})
	svc := NewGroupAnchoringService(client)

	tx, err := svc.AnchorGroupCreated(context.Background(), "g1", "did:key:admin", 3)
	if err != nil {
		t.Fatal(err)
	}
	if tx == "" || !submitted {
		t.Fatal("expected submission")
	}
	proof, ok := svc.GetBlockchainProof("g1")
	if !ok || proof.MemberCountHash == "" {
		t.Fatalf("proof missing: %+v", proof)
	}
}
