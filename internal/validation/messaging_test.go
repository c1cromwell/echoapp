package validation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func testDIDKey(t *testing.T) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	did, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return did
}

func TestValidateDIDKey(t *testing.T) {
	valid := testDIDKey(t)
	if err := ValidateDIDKey(valid); err != nil {
		t.Fatalf("expected valid did:key: %v", err)
	}
	if err := ValidateDIDKey(""); err == nil {
		t.Fatal("empty did should fail")
	}
	if err := ValidateDIDKey("did:prism:foo"); err == nil {
		t.Fatal("did:prism should fail")
	}
}

func TestValidateGroupID(t *testing.T) {
	if err := ValidateGroupID("grp-abc_123"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGroupID(""); err == nil {
		t.Fatal("empty group id")
	}
	if err := ValidateGroupID(strings.Repeat("a", 200)); err == nil {
		t.Fatal("long group id")
	}
}

func TestValidateGroupCreate(t *testing.T) {
	owner := testDIDKey(t)
	if err := ValidateGroupCreate("grp1", owner, groups.GroupTypePrivate, "Team", ""); err != nil {
		t.Fatal(err)
	}
	if err := ValidateGroupCreate("grp1", owner, groups.GroupType("invalid"), "Team", ""); err == nil {
		t.Fatal("bad group type")
	}
}

func TestValidateSyncPush(t *testing.T) {
	if err := ValidateSyncPush("dev-1", []byte("opaque")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSyncPush("", []byte("x")); err == nil {
		t.Fatal("missing device")
	}
}
