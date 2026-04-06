package onboarding

import (
	"testing"
	"time"
)

func TestCredentialVerificationService(t *testing.T) {
	registry := NewTrustRegistryService()
	cvs := NewCredentialVerificationService(registry)

	t.Run("verify_valid_credential_with_passport", func(t *testing.T) {
		// Arrange
		issuer := &TrustedIssuer{
			ID:              "passport_issuer",
			Name:            "EU Passport Authority",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}
		registry.RegisterIssuer(issuer)

		credential := &VerifiableCredential{
			ID:             "passport_123",
			Type:           []string{"VerifiableCredential", "PassportCredential"},
			Issuer:         issuer.ID,
			IssuanceDate:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "signature_abc123",
			CredentialSubject: map[string]interface{}{
				"passportNumber": "AA123456",
				"name":           "John Doe",
				"nationality":    "EU",
			},
		}

		// Act
		result := cvs.verifyCredential(credential)

		// Assert
		if !result.Valid {
			t.Errorf("Expected credential to be valid, got error: %s", result.Error)
		}
		if result.CredentialType != CredTypePassport {
			t.Errorf("Expected credential type %s, got %s", CredTypePassport, result.CredentialType)
		}
		if result.IssuanceTrustLevel != TrustLevelHigh {
			t.Errorf("Expected trust level %s, got %s", TrustLevelHigh, result.IssuanceTrustLevel)
		}
	})

	t.Run("reject_invalid_credential_missing_issuer", func(t *testing.T) {
		credential := &VerifiableCredential{
			ID:           "invalid_123",
			Type:         []string{"VerifiableCredential"},
			Issuer:       "unknown_issuer",
			IssuanceDate: time.Now().Format(time.RFC3339),
			ProofValue:   "sig_abc",
			CredentialSubject: map[string]interface{}{
				"name": "Unknown",
			},
		}

		result := cvs.verifyCredential(credential)

		if result.Valid {
			t.Error("Expected credential with unknown issuer to be invalid")
		}
		if result.Error == "" {
			t.Error("Expected error message for unknown issuer")
		}
	})

	t.Run("reject_expired_credential", func(t *testing.T) {
		issuer := &TrustedIssuer{
			ID:              "expired_issuer",
			Name:            "Test Issuer",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeNationalID},
		}
		registry.RegisterIssuer(issuer)

		credential := &VerifiableCredential{
			ID:             "expired_123",
			Type:           []string{"VerifiableCredential"},
			Issuer:         issuer.ID,
			IssuanceDate:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			ExpirationDate: time.Now().Add(-24 * time.Hour).Format(time.RFC3339), // Already expired
			ProofValue:     "sig_123",
			CredentialSubject: map[string]interface{}{
				"nationalIdNumber": "12345",
			},
		}

		result := cvs.verifyCredential(credential)

		if result.Valid {
			t.Error("Expected expired credential to be invalid")
		}
	})

	t.Run("reject_revoked_credential", func(t *testing.T) {
		issuer := &TrustedIssuer{
			ID:              "bank_issuer",
			Name:            "Bank",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeFinancial,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeBankAccount},
		}
		registry.RegisterIssuer(issuer)

		credential := &VerifiableCredential{
			ID:             "bank_123",
			Type:           []string{"VerifiableCredential", "BankAccountCredential"},
			Issuer:         issuer.ID,
			IssuanceDate:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_123",
			CredentialStatus: &CredentialStatus{
				ID:   "http://example.com/status",
				Type: "StatusList2021",
			},
			CredentialSubject: map[string]interface{}{
				"accountNumber": "123456789",
			},
		}

		// Mark as revoked
		cvs.RevokeCredential(issuer.ID, credential.ID)

		result := cvs.verifyCredential(credential)

		if result.Valid {
			t.Error("Expected revoked credential to be invalid")
		}
	})

	t.Run("check_credential_uniqueness", func(t *testing.T) {
		issuer := &TrustedIssuer{
			ID:              "edu_issuer",
			Name:            "University",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeEducational,
			TrustLevel:      TrustLevelMedium,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeEducationVerification},
		}
		registry.RegisterIssuer(issuer)

		credential := &VerifiableCredential{
			ID:             "degree_123",
			Type:           []string{"VerifiableCredential", "DegreeCredential"},
			Issuer:         issuer.ID,
			IssuanceDate:   time.Now().Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_123",
			CredentialSubject: map[string]interface{}{
				"degreeType": "Bachelor",
				"major":      "Computer Science",
			},
		}

		// First use should succeed
		unique, msg := cvs.CheckCredentialUniqueness(credential)
		if !unique {
			t.Errorf("Expected first credential use to be unique: %s", msg)
		}

		// Mark as used
		cvs.MarkCredentialAsUsed(credential)

		// Second use should fail
		unique, msg = cvs.CheckCredentialUniqueness(credential)
		if unique {
			t.Error("Expected duplicate credential use to be rejected")
		}
		if msg == "" {
			t.Error("Expected error message for duplicate credential")
		}
	})
}

func TestVerifiablePresentation(t *testing.T) {
	registry := NewTrustRegistryService()
	cvs := NewCredentialVerificationService(registry)

	t.Run("verify_valid_presentation", func(t *testing.T) {
		// Setup issuers
		govIssuer := &TrustedIssuer{
			ID:              "gov_issuer",
			Name:            "Government",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeNationalID},
		}
		registry.RegisterIssuer(govIssuer)

		credential := &VerifiableCredential{
			ID:             "id_123",
			Type:           []string{"VerifiableCredential", "IDCredential"},
			Issuer:         govIssuer.ID,
			IssuanceDate:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_123",
			CredentialSubject: map[string]interface{}{
				"nationalIdNumber": "123456789",
				"name":             "Alice",
			},
		}

		vp := &VerifiablePresentation{
			Type:                  []string{"VerifiablePresentation"},
			Holder:                "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			VerifiableCredentials: []VerifiableCredential{*credential},
			ProofType:             "Ed25519Signature2020",
			Challenge:             "nonce_12345",
			ProofValue:            "sig_vp_123",
		}

		result := cvs.VerifyPresentation(vp, "nonce_12345")

		if !result.Valid {
			t.Error("Expected presentation to be valid")
		}
		if result.CredentialsVerified != 1 {
			t.Errorf("Expected 1 credential verified, got %d", result.CredentialsVerified)
		}
	})

	t.Run("reject_presentation_with_wrong_nonce", func(t *testing.T) {
		issuer := &TrustedIssuer{
			ID:              "test_issuer",
			Name:            "Test",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeNationalID},
		}
		registry.RegisterIssuer(issuer)

		credential := &VerifiableCredential{
			ID:             "id_456",
			Type:           []string{"VerifiableCredential"},
			Issuer:         issuer.ID,
			IssuanceDate:   time.Now().Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_456",
			CredentialSubject: map[string]interface{}{
				"nationalIdNumber": "987654321",
			},
		}

		vp := &VerifiablePresentation{
			Type:                  []string{"VerifiablePresentation"},
			Holder:                "did:key:xyz",
			VerifiableCredentials: []VerifiableCredential{*credential},
			Challenge:             "wrong_nonce",
			ProofValue:            "sig_vp_456",
		}

		result := cvs.VerifyPresentation(vp, "correct_nonce")

		if result.Valid {
			t.Error("Expected presentation with wrong nonce to be invalid")
		}
	})

	t.Run("verify_multiple_credentials_presentation", func(t *testing.T) {
		// Setup multiple issuers
		govIssuer := &TrustedIssuer{
			ID:              "gov",
			Name:            "Government",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeGovernment,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypePassport},
		}

		bankIssuer := &TrustedIssuer{
			ID:              "bank",
			Name:            "Bank",
			DID:             "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
			Type:            IssuerTypeFinancial,
			TrustLevel:      TrustLevelHigh,
			Status:          "active",
			CredentialTypes: []CredentialType{CredTypeBankAccount},
		}

		registry.RegisterIssuer(govIssuer)
		registry.RegisterIssuer(bankIssuer)

		passportCred := &VerifiableCredential{
			ID:             "pp_123",
			Type:           []string{"VerifiableCredential", "PassportCredential"},
			Issuer:         govIssuer.ID,
			IssuanceDate:   time.Now().Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_pp",
			CredentialSubject: map[string]interface{}{
				"passportNumber": "AA123456",
			},
		}

		bankCred := &VerifiableCredential{
			ID:             "ba_123",
			Type:           []string{"VerifiableCredential", "BankCredential"},
			Issuer:         bankIssuer.ID,
			IssuanceDate:   time.Now().Format(time.RFC3339),
			ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			ProofValue:     "sig_ba",
			CredentialSubject: map[string]interface{}{
				"accountNumber": "123456789",
			},
		}

		vp := &VerifiablePresentation{
			Type:                  []string{"VerifiablePresentation"},
			Holder:                "did:key:holder",
			VerifiableCredentials: []VerifiableCredential{*passportCred, *bankCred},
			Challenge:             "nonce_456",
			ProofValue:            "sig_msvp",
		}

		result := cvs.VerifyPresentation(vp, "nonce_456")

		if !result.Valid {
			t.Error("Expected presentation with multiple credentials to be valid")
		}
		if result.CredentialsVerified != 2 {
			t.Errorf("Expected 2 credentials verified, got %d", result.CredentialsVerified)
		}
	})
}

func TestTrustScoreCalculator(t *testing.T) {
	calculator := NewTrustScoreCalculator()

	t.Run("calculate_score_single_high_trust_credential", func(t *testing.T) {
		credentials := []CredentialVerificationResult{
			{
				Valid:              true,
				CredentialType:     CredTypePassport,
				IssuanceTrustLevel: TrustLevelHigh,
			},
		}

		score, badge := calculator.CalculateScore(credentials)

		if score != 90 {
			t.Errorf("Expected score 90 for high-trust passport, got %d", score)
		}
		if badge != "🛂 Passport Verified" {
			t.Errorf("Expected passport badge, got %s", badge)
		}
	})

	t.Run("calculate_score_with_multiple_credentials_bonus", func(t *testing.T) {
		credentials := []CredentialVerificationResult{
			{
				Valid:              true,
				CredentialType:     CredTypePassport,
				IssuanceTrustLevel: TrustLevelHigh,
			},
			{
				Valid:              true,
				CredentialType:     CredTypeBankAccount,
				IssuanceTrustLevel: TrustLevelHigh,
			},
			{
				Valid:              true,
				CredentialType:     CredTypeEducationVerification,
				IssuanceTrustLevel: TrustLevelHigh,
			},
		}

		score, _ := calculator.CalculateScore(credentials)

		// Base score is 90 (passport), + 10 bonus for 3+ credentials
		if score != 100 {
			t.Errorf("Expected score 100 (with bonus cap), got %d", score)
		}
	})

	t.Run("calculate_score_empty_credentials", func(t *testing.T) {
		credentials := []CredentialVerificationResult{}

		score, badge := calculator.CalculateScore(credentials)

		if score != 0 {
			t.Errorf("Expected score 0 for empty credentials, got %d", score)
		}
		if badge != "" {
			t.Errorf("Expected empty badge, got %s", badge)
		}
	})

	t.Run("calculate_score_ignores_invalid_credentials", func(t *testing.T) {
		credentials := []CredentialVerificationResult{
			{
				Valid:              false,
				CredentialType:     CredTypePassport,
				IssuanceTrustLevel: TrustLevelHigh,
			},
			{
				Valid:              true,
				CredentialType:     CredTypeBankAccount,
				IssuanceTrustLevel: TrustLevelHigh,
			},
		}

		score, _ := calculator.CalculateScore(credentials)

		// Only bank account (75 for high trust)
		if score != 75 {
			t.Errorf("Expected score 75 (bank only), got %d", score)
		}
	})
}
