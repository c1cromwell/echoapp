package metagraph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeviceKeyRegistrationUpdateJSONShape(t *testing.T) {
	pub := "04" + strings.Repeat("a", 128)
	u := DeviceKeyRegistrationUpdate{
		SubjectDID:   "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		PublicKeyHex: pub,
		DeviceLabel:  "ipad",
		AddedAt:      1700000000000,
	}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["subjectDID"] != u.SubjectDID {
		t.Fatalf("subjectDID: %v", m["subjectDID"])
	}
	if m["publicKeyHex"] != pub {
		t.Fatal("publicKeyHex mismatch")
	}
	if m["deviceLabel"] != "ipad" {
		t.Fatal("deviceLabel mismatch")
	}
	if _, ok := m["addedAt"]; !ok {
		t.Fatal("missing addedAt")
	}
}

func TestSubmitIdentityL1RequiresURL(t *testing.T) {
	c := NewMetagraphClient(MetagraphConfig{})
	_, err := c.SubmitIdentityL1(t.Context(), DeviceKeyRegistrationUpdate{})
	if err == nil || !strings.Contains(err.Error(), "IdentityL1URL") {
		t.Fatalf("expected IdentityL1URL error, got %v", err)
	}
}
