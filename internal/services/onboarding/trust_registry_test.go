package onboarding

import (
	"testing"
	"time"
)

func TestTrustRegistry(t *testing.T) {
	t.Run("register_issuer", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "test_issuer",
			Name:            "Test Issuer",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4test",
			Type:            IssuerTypeGovernment,
			Jurisdiction:    JurisdictionUS,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		err := registry.RegisterIssuer(issuer)

		if err != nil {
			t.Errorf("Unexpected error registering issuer: %v", err)
		}

		retrieved, err := registry.GetIssuer(issuer.ID)
		if err != nil || retrieved == nil {
			t.Error("Failed to retrieve registered issuer")
		}
	})

	t.Run("register_duplicate_issuer_fails", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "dup_issuer",
			Name:            "Duplicate",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4dup",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeNationalID},
		}

		registry.RegisterIssuer(issuer)
		err := registry.RegisterIssuer(issuer)

		if err == nil {
			t.Error("Expected error registering duplicate issuer")
		}
	})

	t.Run("get_issuer_by_id", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "issuer_123",
			Name:            "Issuer 123",
			DID:             "did:key:z6Mk123456",
			Type:            IssuerTypeFinancial,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeBankAccount},
		}

		registry.RegisterIssuer(issuer)
		retrieved, err := registry.GetIssuer("issuer_123")

		if err != nil || retrieved == nil {
			t.Error("Failed to get issuer by ID")
		}
		if retrieved.Name != "Issuer 123" {
			t.Errorf("Expected name 'Issuer 123', got %s", retrieved.Name)
		}
	})

	t.Run("get_issuer_by_did", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "issuer_abc",
			Name:            "Issuer ABC",
			DID:             "did:key:z6Mkabcdef",
			Type:            IssuerTypeEducational,
			TrustLevel:      TrustLevelMedium,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeEducationVerification},
		}

		registry.RegisterIssuer(issuer)
		retrieved, err := registry.GetIssuerByDID("did:key:z6Mkabcdef")

		if err != nil || retrieved == nil {
			t.Error("Failed to get issuer by DID")
		}
		if retrieved.ID != "issuer_abc" {
			t.Errorf("Expected ID 'issuer_abc', got %s", retrieved.ID)
		}
	})

	t.Run("verify_credential_type", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "multi_cred_issuer",
			Name:            "Multi Cred",
			DID:             "did:key:z6MkMulti",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport, CredTypeDriversLicense},
		}

		registry.RegisterIssuer(issuer)

		// Should succeed for supported type
		ok, err := registry.VerifyCredentialType("multi_cred_issuer", CredTypePassport)
		if !ok || err != nil {
			t.Error("Expected issuer to support passport credential type")
		}

		// Should fail for unsupported type
		ok, err = registry.VerifyCredentialType("multi_cred_issuer", CredTypePhoneVerification)
		if ok {
			t.Error("Expected issuer to not support phone verification credential type")
		}
	})

	t.Run("suspend_and_resume_issuer", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "suspend_issuer",
			Name:            "Suspendable",
			DID:             "did:key:z6MkSuspend",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeNationalID},
		}

		registry.RegisterIssuer(issuer)

		// Suspend
		err := registry.SuspendIssuer("suspend_issuer")
		if err != nil {
			t.Errorf("Unexpected error suspending issuer: %v", err)
		}

		// Should not be able to get suspended issuer
		_, err = registry.GetIssuer("suspend_issuer")
		if err == nil {
			t.Error("Expected error getting suspended issuer")
		}

		// Resume
		err = registry.ResumeIssuer("suspend_issuer")
		if err != nil {
			t.Errorf("Unexpected error resuming issuer: %v", err)
		}

		// Should be able to get again
		retrieved, err := registry.GetIssuer("suspend_issuer")
		if err != nil || retrieved == nil {
			t.Error("Failed to get resumed issuer")
		}
	})

	t.Run("revoke_issuer", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "revoke_issuer",
			Name:            "Revokable",
			DID:             "did:key:z6MkRevoke",
			Type:            IssuerTypeFinancial,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeBankAccount},
		}

		registry.RegisterIssuer(issuer)

		// Revoke
		err := registry.RevokeIssuer("revoke_issuer")
		if err != nil {
			t.Errorf("Unexpected error revoking issuer: %v", err)
		}

		// Should not be able to get revoked issuer
		_, err = registry.GetIssuer("revoke_issuer")
		if err == nil {
			t.Error("Expected error getting revoked issuer")
		}
	})

	t.Run("get_issuers_by_jurisdiction", func(t *testing.T) {
		registry := NewTrustRegistryService()

		usIssuer := &TrustedIssuer{
			ID:              "us_issuer",
			Name:            "US",
			DID:             "did:key:z6MkUS",
			Type:            IssuerTypeGovernment,
			Jurisdiction:    JurisdictionUS,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		euIssuer := &TrustedIssuer{
			ID:              "eu_issuer",
			Name:            "EU",
			DID:             "did:key:z6MkEU",
			Type:            IssuerTypeGovernment,
			Jurisdiction:    JurisdictionEU,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		registry.RegisterIssuer(usIssuer)
		registry.RegisterIssuer(euIssuer)

		usIssuers := registry.GetIssuersByJurisdiction(JurisdictionUS)
		if len(usIssuers) == 0 {
			t.Error("Expected to find US issuer")
		}

		euIssuers := registry.GetIssuersByJurisdiction(JurisdictionEU)
		if len(euIssuers) == 0 {
			t.Error("Expected to find EU issuer")
		}
	})

	t.Run("get_issuers_by_type", func(t *testing.T) {
		registry := NewTrustRegistryService()

		govIssuer := &TrustedIssuer{
			ID:              "gov_type",
			Name:            "Government",
			DID:             "did:key:z6MkGov",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		bankIssuer := &TrustedIssuer{
			ID:              "bank_type",
			Name:            "Bank",
			DID:             "did:key:z6MkBank",
			Type:            IssuerTypeFinancial,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeBankAccount},
		}

		registry.RegisterIssuer(govIssuer)
		registry.RegisterIssuer(bankIssuer)

		governments := registry.GetIssuersByType(IssuerTypeGovernment)
		if len(governments) == 0 {
			t.Error("Expected to find government issuer")
		}

		banks := registry.GetIssuersByType(IssuerTypeFinancial)
		if len(banks) == 0 {
			t.Error("Expected to find financial issuer")
		}
	})

	t.Run("verify_issuer_qualifications", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "qual_issuer",
			Name:            "Qualified",
			DID:             "did:key:z6MkQual",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
			Qualifications: []Qualification{
				{
					ID:           "soc2_cert",
					Type:         "soc2",
					IssuanceDate: time.Now().AddDate(-1, 0, 0),
					ExpiryDate:   time.Now().AddDate(1, 0, 0),
					Verified:     true,
				},
				{
					ID:           "iso_cert",
					Type:         "iso",
					IssuanceDate: time.Now().AddDate(-2, 0, 0),
					ExpiryDate:   time.Now().AddDate(-1, 0, 0), // Expired
					Verified:     true,
				},
			},
		}

		registry.RegisterIssuer(issuer)

		valid, qualifications := registry.VerifyIssuerQualifications("qual_issuer")

		if !valid {
			t.Error("Expected issuer to have valid qualifications")
		}
		if len(qualifications) == 0 {
			t.Error("Expected to find SOC2 qualification")
		}

		foundSOC2 := false
		for _, qual := range qualifications {
			if qual == "soc2" {
				foundSOC2 = true
			}
		}
		if !foundSOC2 {
			t.Error("Expected SOC2 qualification to be found")
		}
	})

	t.Run("update_issuer_status", func(t *testing.T) {
		registry := NewTrustRegistryService()

		issuer := &TrustedIssuer{
			ID:              "status_issuer",
			Name:            "Status",
			DID:             "did:key:z6MkStatus",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		registry.RegisterIssuer(issuer)

		err := registry.UpdateIssuerStatus("status_issuer", "inactive")
		if err != nil {
			t.Errorf("Unexpected error updating status: %v", err)
		}

		// Should not be able to get inactive issuer
		_, err = registry.GetIssuer("status_issuer")
		if err == nil {
			t.Error("Expected error getting inactive issuer")
		}
	})

	t.Run("well_known_issuers_initialized", func(t *testing.T) {
		registry := NewTrustRegistryService()

		// Should have 3 well-known issuers initialized
		active := registry.GetActiveTrustedIssuers()
		if len(active) < 3 {
			t.Errorf("Expected at least 3 well-known issuers, got %d", len(active))
		}

		// Check DMV exists
		dmv, err := registry.GetIssuer("gvt_us_dmv")
		if err != nil || dmv == nil {
			t.Error("Expected DMV issuer to be initialized")
		}

		// Check Wells Fargo exists
		wf, err := registry.GetIssuer("bank_wellsfargo")
		if err != nil || wf == nil {
			t.Error("Expected Wells Fargo issuer to be initialized")
		}
	})
}
