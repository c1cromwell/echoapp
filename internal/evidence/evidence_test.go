package evidence

import (
	"strings"
	"testing"
)

func validConfig() *ClientConfig {
	return &ClientConfig{
		APIKey:         "test-api-key-123",
		OrganizationID: "org-echo-001",
		TenantID:       "tenant-001",
		BaseURL:        "https://evidence.constellationnetwork.io",
	}
}

// --- ClientConfig Tests ---

func TestClientConfig_Validate_Valid(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestClientConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := validConfig()
	cfg.APIKey = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestClientConfig_Validate_MissingOrgID(t *testing.T) {
	cfg := validConfig()
	cfg.OrganizationID = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing org ID")
	}
}

func TestClientConfig_Validate_MissingTenantID(t *testing.T) {
	cfg := validConfig()
	cfg.TenantID = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing tenant ID")
	}
}

func TestClientConfig_Validate_MissingBaseURL(t *testing.T) {
	cfg := validConfig()
	cfg.BaseURL = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

// --- ComputeFingerprint Tests ---

func TestComputeFingerprint_Deterministic(t *testing.T) {
	data := []byte("test data for fingerprinting")
	hash1 := ComputeFingerprint(data)
	hash2 := ComputeFingerprint(data)

	if hash1 != hash2 {
		t.Error("fingerprint should be deterministic")
	}
	if len(hash1) != 64 { // SHA-256 hex = 64 chars
		t.Errorf("expected 64-char hex hash, got %d", len(hash1))
	}
}

func TestComputeFingerprint_DifferentInputs(t *testing.T) {
	hash1 := ComputeFingerprint([]byte("data A"))
	hash2 := ComputeFingerprint([]byte("data B"))

	if hash1 == hash2 {
		t.Error("different inputs should produce different fingerprints")
	}
}

// --- CanAccessEvidence Tests ---

func TestCanAccessEvidence_FreeTier(t *testing.T) {
	types := []EvidenceType{EvidenceAuditTrail, EvidenceMediaAuth, EvidenceSmartCheck, EvidenceRetentionProof}
	for _, et := range types {
		if CanAccessEvidence(TierFree, et) {
			t.Errorf("free tier should not access %s", et)
		}
	}
}

func TestCanAccessEvidence_VIPTier(t *testing.T) {
	if !CanAccessEvidence(TierVIP, EvidenceMediaAuth) {
		t.Error("VIP should access media authenticity")
	}
	if CanAccessEvidence(TierVIP, EvidenceAuditTrail) {
		t.Error("VIP should not access audit trail")
	}
	if CanAccessEvidence(TierVIP, EvidenceSmartCheck) {
		t.Error("VIP should not access smart checkmark")
	}
	if CanAccessEvidence(TierVIP, EvidenceRetentionProof) {
		t.Error("VIP should not access retention proof")
	}
}

func TestCanAccessEvidence_OrgTier(t *testing.T) {
	types := []EvidenceType{EvidenceAuditTrail, EvidenceMediaAuth, EvidenceSmartCheck, EvidenceRetentionProof}
	for _, et := range types {
		if !CanAccessEvidence(TierOrganization, et) {
			t.Errorf("org tier should access %s", et)
		}
	}
}

func TestCanAccessEvidence_UnknownTier(t *testing.T) {
	if CanAccessEvidence(UserTier("unknown"), EvidenceAuditTrail) {
		t.Error("unknown tier should not access anything")
	}
}

// --- AuditBatchFingerprint Tests ---

func TestAuditBatchFingerprint_Valid(t *testing.T) {
	cfg := validConfig()
	req, err := AuditBatchFingerprint(cfg, "QmIPFSCID123", "batchmetahash456", "did:dag:relay1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.EvidenceType != EvidenceAuditTrail {
		t.Errorf("expected audit_trail, got %s", req.EvidenceType)
	}
	if req.Hash == "" {
		t.Error("hash should not be empty")
	}
	if req.Metadata.IPFSCid != "QmIPFSCID123" {
		t.Errorf("expected IPFS CID in metadata, got %s", req.Metadata.IPFSCid)
	}
	if req.OrganizationID != "org-echo-001" {
		t.Errorf("expected org ID from config")
	}
}

func TestAuditBatchFingerprint_InvalidConfig(t *testing.T) {
	cfg := &ClientConfig{} // empty config
	_, err := AuditBatchFingerprint(cfg, "cid", "hash", "did")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

// --- MediaFingerprint Tests ---

func TestMediaFingerprint_Valid(t *testing.T) {
	cfg := validConfig()
	mediaHash := ComputeFingerprint([]byte("raw image bytes"))
	req, err := MediaFingerprint(cfg, mediaHash, "did:dag:user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.EvidenceType != EvidenceMediaAuth {
		t.Errorf("expected media_authenticity, got %s", req.EvidenceType)
	}
	if req.Hash != mediaHash {
		t.Error("hash should match the provided media hash")
	}
}

// --- SmartCheckmarkFingerprint Tests ---

func TestSmartCheckmarkFingerprint_Valid(t *testing.T) {
	cfg := validConfig()
	req, err := SmartCheckmarkFingerprint(cfg, "msgcontenthash", "2026-03-08T12:00:00Z", "did:dag:sender1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.EvidenceType != EvidenceSmartCheck {
		t.Errorf("expected smart_checkmark, got %s", req.EvidenceType)
	}
	if req.Hash == "" {
		t.Error("hash should not be empty")
	}
	if req.Metadata.SourceDID != "did:dag:sender1" {
		t.Errorf("expected sender DID in metadata")
	}
}

func TestSmartCheckmarkFingerprint_DifferentInputs(t *testing.T) {
	cfg := validConfig()
	req1, _ := SmartCheckmarkFingerprint(cfg, "hash1", "2026-03-08T12:00:00Z", "did:dag:sender1")
	req2, _ := SmartCheckmarkFingerprint(cfg, "hash2", "2026-03-08T12:00:00Z", "did:dag:sender1")

	if req1.Hash == req2.Hash {
		t.Error("different message hashes should produce different fingerprints")
	}
}

// --- RetentionProofFingerprint Tests ---

func TestRetentionProofFingerprint_Valid(t *testing.T) {
	cfg := validConfig()
	req, err := RetentionProofFingerprint(cfg, "auditbatchhash", "deletion-confirmed", "did:dag:admin1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.EvidenceType != EvidenceRetentionProof {
		t.Errorf("expected retention_proof, got %s", req.EvidenceType)
	}
	if !strings.Contains(req.Metadata.Description, "retention") {
		t.Error("description should mention retention")
	}
}

func TestRetentionProofFingerprint_InvalidConfig(t *testing.T) {
	cfg := &ClientConfig{APIKey: "key"} // missing other fields
	_, err := RetentionProofFingerprint(cfg, "hash", "del", "did")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
