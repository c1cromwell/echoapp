package onboarding

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// CredentialType represents the type of verifiable credential
type CredentialType string

const (
	CredTypeGovernmentID           CredentialType = "government_id"
	CredTypePassport               CredentialType = "passport"
	CredTypeDriversLicense         CredentialType = "drivers_license"
	CredTypeNationalID             CredentialType = "national_id"
	CredTypeBankAccount            CredentialType = "bank_account"
	CredTypeProofOfAddress         CredentialType = "proof_of_address"
	CredTypeEmploymentVerification CredentialType = "employment"
	CredTypeEducationVerification  CredentialType = "education"
	CredTypeAgeVerification        CredentialType = "age_verification"
	CredTypeEmailVerification      CredentialType = "email_verification"
	CredTypePhoneVerification      CredentialType = "phone_verification"
)

// VerifiableCredential represents a W3C Verifiable Credential
type VerifiableCredential struct {
	ID                 string
	Type               []string
	Issuer             string
	IssuanceDate       string
	ExpirationDate     string
	CredentialSubject  map[string]interface{}
	CredentialStatus   *CredentialStatus
	ProofType          string
	ProofValue         string
	VerificationMethod string
}

// CredentialStatus tracks revocation information
type CredentialStatus struct {
	ID                   string
	Type                 string
	StatusPurpose        string
	StatusListIndex      string
	StatusListCredential string
}

// VerifiablePresentation represents a W3C Verifiable Presentation
type VerifiablePresentation struct {
	Type                  []string
	Holder                string
	VerifiableCredentials []VerifiableCredential
	ProofType             string
	Challenge             string
	Domain                string
	ProofValue            string
}

// CredentialVerificationResult stores credential verification outcome
type CredentialVerificationResult struct {
	Valid              bool
	CredentialType     CredentialType
	Issuer             *TrustedIssuer
	IssuanceTrustLevel TrustLevel
	Claims             map[string]interface{}
	Error              string
	VerifiedAt         time.Time
}

// VerificationResult stores the complete verification of a presentation
type VerificationResult struct {
	Valid               bool
	Credentials         []CredentialVerificationResult
	CredentialsVerified int
	IssuersVerified     bool
	NotRevoked          bool
	OverallValid        bool
	ExtractedClaims     map[string]interface{}
	VerifiedAt          time.Time
	Error               string
}

// CredentialVerificationService handles the multi-stage credential verification
type CredentialVerificationService struct {
	trustRegistry   *TrustRegistryService
	revokedCreds    map[string]bool // credentialID -> isRevoked
	usedCredentials map[string]bool // credentialHash -> used
}

// NewCredentialVerificationService creates the verification service
func NewCredentialVerificationService(registry *TrustRegistryService) *CredentialVerificationService {
	return &CredentialVerificationService{
		trustRegistry:   registry,
		revokedCreds:    make(map[string]bool),
		usedCredentials: make(map[string]bool),
	}
}

// VerifyPresentation performs multi-stage verification of a VP
func (cvs *CredentialVerificationService) VerifyPresentation(
	vp *VerifiablePresentation,
	sessionNonce string,
) *VerificationResult {
	result := &VerificationResult{
		Valid:           false,
		Credentials:     make([]CredentialVerificationResult, 0),
		ExtractedClaims: make(map[string]interface{}),
	}

	// Stage 1: Validate presentation structure
	if len(vp.VerifiableCredentials) == 0 {
		result.Error = "presentation contains no credentials"
		return result
	}

	// Validate challenge/nonce
	if vp.Challenge != sessionNonce {
		result.Error = "challenge/nonce mismatch"
		return result
	}

	// Stage 2-5: Verify each credential
	validCount := 0
	for _, vc := range vp.VerifiableCredentials {
		credResult := cvs.verifyCredential(&vc)
		result.Credentials = append(result.Credentials, credResult)

		if credResult.Valid {
			validCount++
			// Merge claims (only verified ones)
			for key, value := range credResult.Claims {
				result.ExtractedClaims[key] = value
			}
		}
	}

	result.CredentialsVerified = validCount
	result.IssuersVerified = validCount > 0
	result.NotRevoked = validCount > 0

	// Overall result requires at least one valid credential
	result.Valid = validCount > 0
	result.OverallValid = result.Valid
	result.VerifiedAt = time.Now()

	return result
}

// verifyCredential performs all verification stages for a single credential
// Stage 1: Structure validation
// Stage 2: Issuer trust verification
// Stage 3: Cryptographic signature verification
// Stage 4: Revocation check
// Stage 5: Claims validation
func (cvs *CredentialVerificationService) verifyCredential(vc *VerifiableCredential) CredentialVerificationResult {
	result := CredentialVerificationResult{
		Valid:  false,
		Claims: make(map[string]interface{}),
	}

	// Stage 1: Basic structure validation
	if vc.IssuanceDate == "" || vc.Issuer == "" {
		result.Error = "missing required credential fields"
		return result
	}

	// Stage 2: Issuer trust verification
	issuer, err := cvs.trustRegistry.GetIssuer(vc.Issuer)
	if err != nil || issuer == nil {
		result.Error = "issuer not found in trust registry"
		return result
	}

	if issuer.Status != "active" {
		result.Error = "issuer is not in active status"
		return result
	}

	result.Issuer = issuer
	result.IssuanceTrustLevel = issuer.TrustLevel

	// Stage 3: Cryptographic verification (simulated)
	// In production, would verify signature using issuer's public keys
	if vc.ProofValue == "" {
		result.Error = "credential has no valid signature"
		return result
	}

	// Stage 4: Check revocation status (simulated)
	if vc.CredentialStatus != nil {
		credHash := fmt.Sprintf("%s:%s", vc.Issuer, vc.ID)
		if cvs.revokedCreds[credHash] {
			result.Error = "credential has been revoked"
			return result
		}
	}

	// Check expiration
	if vc.ExpirationDate != "" {
		if expiryTime, err := time.Parse(time.RFC3339, vc.ExpirationDate); err == nil {
			if time.Now().After(expiryTime) {
				result.Error = "credential has expired"
				return result
			}
		}
	}

	// Stage 5: Extract and validate claims
	credType := cvs.mapClaimsToCrentialType(vc.CredentialSubject)
	result.CredentialType = credType
	result.Claims = vc.CredentialSubject
	result.Valid = true
	result.VerifiedAt = time.Now()

	return result
}

// CheckCredentialUniqueness prevents Sybil attacks by checking if credential was used
func (cvs *CredentialVerificationService) CheckCredentialUniqueness(vc *VerifiableCredential) (bool, string) {
	credHash := cvs.hashCredential(vc)

	if cvs.usedCredentials[credHash] {
		return false, "This credential has already been used to create an account"
	}

	return true, ""
}

// MarkCredentialAsUsed registers a credential to prevent reuse
func (cvs *CredentialVerificationService) MarkCredentialAsUsed(vc *VerifiableCredential) {
	credHash := cvs.hashCredential(vc)
	cvs.usedCredentials[credHash] = true
}

// RevokeCredential marks a credential as revoked
func (cvs *CredentialVerificationService) RevokeCredential(issuerID, credentialID string) {
	credHash := fmt.Sprintf("%s:%s", issuerID, credentialID)
	cvs.revokedCreds[credHash] = true
}

// Helper functions

func (cvs *CredentialVerificationService) hashCredential(vc *VerifiableCredential) string {
	input := fmt.Sprintf("%s:%s:%v", vc.Issuer, vc.ID, vc.Type)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func (cvs *CredentialVerificationService) mapClaimsToCrentialType(
	claims map[string]interface{},
) CredentialType {
	// Map based on claim presence
	if _, ok := claims["passportNumber"]; ok {
		return CredTypePassport
	}
	if _, ok := claims["licenseNumber"]; ok {
		return CredTypeDriversLicense
	}
	if _, ok := claims["nationalIdNumber"]; ok {
		return CredTypeNationalID
	}
	if _, ok := claims["accountNumber"]; ok {
		return CredTypeBankAccount
	}
	if _, ok := claims["employmentStatus"]; ok {
		return CredTypeEmploymentVerification
	}
	if _, ok := claims["educationLevel"]; ok {
		return CredTypeEducationVerification
	}
	return CredTypeGovernmentID
}

// TrustScoreCalculator computes initial trust based on credentials
type TrustScoreCalculator struct {
	scoreMatrix map[CredentialType]map[TrustLevel]int
}

// NewTrustScoreCalculator creates score calculator
func NewTrustScoreCalculator() *TrustScoreCalculator {
	return &TrustScoreCalculator{
		scoreMatrix: map[CredentialType]map[TrustLevel]int{
			CredTypePassport: {
				TrustLevelHigh:   90,
				TrustLevelMedium: 70,
			},
			CredTypeNationalID: {
				TrustLevelHigh:   85,
				TrustLevelMedium: 65,
			},
			CredTypeDriversLicense: {
				TrustLevelHigh:   80,
				TrustLevelMedium: 60,
			},
			CredTypeBankAccount: {
				TrustLevelHigh:   75,
				TrustLevelMedium: 55,
				TrustLevelBasic:  40,
			},
			CredTypeEmploymentVerification: {
				TrustLevelHigh:   60,
				TrustLevelMedium: 45,
				TrustLevelBasic:  30,
			},
			CredTypeEducationVerification: {
				TrustLevelHigh:   55,
				TrustLevelMedium: 40,
				TrustLevelBasic:  25,
			},
			CredTypePhoneVerification: {
				TrustLevelHigh:   35,
				TrustLevelMedium: 30,
				TrustLevelBasic:  25,
			},
			CredTypeEmailVerification: {
				TrustLevelHigh:   30,
				TrustLevelMedium: 25,
				TrustLevelBasic:  20,
			},
		},
	}
}

// CalculateScore computes trust score from verified credentials
func (tsc *TrustScoreCalculator) CalculateScore(
	credentials []CredentialVerificationResult,
) (int, string) {
	if len(credentials) == 0 {
		return 0, ""
	}

	// Get base score from highest-trust credential
	var maxScore int
	var primaryBadge string

	for _, cred := range credentials {
		if cred.Valid {
			score := tsc.getCredentialScore(cred.CredentialType, cred.IssuanceTrustLevel)
			if score > maxScore {
				maxScore = score
				primaryBadge = getBadgeForCredential(cred.CredentialType)
			}
		}
	}

	// Apply bonuses for multiple credentials
	validCount := 0
	for _, cred := range credentials {
		if cred.Valid {
			validCount++
		}
	}

	bonus := 0
	if validCount >= 4 {
		bonus = 15
	} else if validCount >= 3 {
		bonus = 10
	} else if validCount >= 2 {
		bonus = 5
	}

	score := maxScore + bonus
	if score > 100 {
		score = 100
	}

	return score, primaryBadge
}

func (tsc *TrustScoreCalculator) getCredentialScore(
	credType CredentialType,
	trustLevel TrustLevel,
) int {
	if matrix, ok := tsc.scoreMatrix[credType]; ok {
		if score, ok := matrix[trustLevel]; ok {
			return score
		}
	}
	return 20 // Default fallback
}

// Badges for credentials
var credentialBadges = map[CredentialType]string{
	CredTypePassport:               "🛂 Passport Verified",
	CredTypeNationalID:             "🪪 ID Verified",
	CredTypeDriversLicense:         "🚗 License Verified",
	CredTypeBankAccount:            "🏦 Bank Verified",
	CredTypeEducationVerification:  "🎓 Education Verified",
	CredTypeEmploymentVerification: "💼 Employment Verified",
	CredTypePhoneVerification:      "📱 Phone Verified",
	CredTypeEmailVerification:      "✉️ Email Verified",
}

func getBadgeForCredential(credType CredentialType) string {
	if badge, ok := credentialBadges[credType]; ok {
		return badge
	}
	return "✓ Verified"
}
