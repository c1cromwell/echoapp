package onboarding

import (
	"fmt"
	"sync"
	"time"
)

// TrustLevel represents the trust level of an issuer
type TrustLevel string

const (
	TrustLevelHigh   TrustLevel = "high"
	TrustLevelMedium TrustLevel = "medium"
	TrustLevelBasic  TrustLevel = "basic"
)

// IssuerType categorizes issuer classification
type IssuerType string

const (
	IssuerTypeGovernment       IssuerType = "government"
	IssuerTypeFinancial        IssuerType = "financial"
	IssuerTypeEducational      IssuerType = "educational"
	IssuerTypeEmployment       IssuerType = "employment"
	IssuerTypeTelecom          IssuerType = "telecom"
	IssuerTypeIdentityProvider IssuerType = "identity_provider"
)

// Jurisdiction represents the geographic jurisdiction
type Jurisdiction string

const (
	JurisdictionUS     Jurisdiction = "us"
	JurisdictionEU     Jurisdiction = "eu"
	JurisdictionUK     Jurisdiction = "uk"
	JurisdictionCA     Jurisdiction = "ca"
	JurisdictionAU     Jurisdiction = "au"
	JurisdictionGlobal Jurisdiction = "global"
)

// Qualification represents issuer qualifications and certifications
type Qualification struct {
	ID           string
	Type         string // audit, soc2, iso, compliance, etc.
	IssuanceDate time.Time
	ExpiryDate   time.Time
	Verified     bool
}

// TrustedIssuer represents a verified issuer in the registry
type TrustedIssuer struct {
	ID                   string
	Name                 string
	DID                  string
	Type                 IssuerType
	Jurisdiction         Jurisdiction
	TrustLevel           TrustLevel
	Status               string // active, suspended, revoked
	CredentialTypes      []CredentialType
	PublicKeyURL         string
	VerifiedTimestamp    time.Time
	LastVerificationDate time.Time
	Qualifications       []Qualification
	RiskScore            int     // 0-100, lower is better
	OnboardingWeight     int     // influence on trust score
	ActivationThreshold  float64 // minimum score needed for activation
	ConstellationAnchor  string  // distributed ledger anchor
	EstablishedDate      time.Time
	ContactEmail         string
	DocumentationURL     string
}

// TrustRegistryService manages the list of trusted issuers
type TrustRegistryService struct {
	mu               sync.RWMutex
	issuers          map[string]*TrustedIssuer // issuerID -> TrustedIssuer
	dids             map[string]*TrustedIssuer // DID -> TrustedIssuer (reverse lookup)
	suspendedIssuers map[string]time.Time      // issuerID -> suspension time
	revokedIssuers   map[string]bool           // issuerID -> revoked
}

// NewTrustRegistryService creates a new trust registry
func NewTrustRegistryService() *TrustRegistryService {
	registry := &TrustRegistryService{
		issuers:          make(map[string]*TrustedIssuer),
		dids:             make(map[string]*TrustedIssuer),
		suspendedIssuers: make(map[string]time.Time),
		revokedIssuers:   make(map[string]bool),
	}

	// Initialize with well-known issuers
	registry.initializeWellKnownIssuers()

	return registry
}

// RegisterIssuer adds a new issuer to the trust registry
func (tr *TrustRegistryService) RegisterIssuer(issuer *TrustedIssuer) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if issuer.ID == "" || issuer.DID == "" {
		return fmt.Errorf("issuer must have ID and DID")
	}

	// Check if already exists
	if _, exists := tr.issuers[issuer.ID]; exists {
		return fmt.Errorf("issuer %s already registered", issuer.ID)
	}

	issuer.VerifiedTimestamp = time.Now()
	issuer.LastVerificationDate = time.Now()

	tr.issuers[issuer.ID] = issuer
	tr.dids[issuer.DID] = issuer

	return nil
}

// GetIssuer retrieves an issuer by ID
func (tr *TrustRegistryService) GetIssuer(issuerID string) (*TrustedIssuer, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Check if revoked
	if tr.revokedIssuers[issuerID] {
		return nil, fmt.Errorf("issuer %s has been revoked", issuerID)
	}

	// Check if suspended
	if suspensionTime, suspended := tr.suspendedIssuers[issuerID]; suspended {
		return nil, fmt.Errorf("issuer %s is suspended since %v", issuerID, suspensionTime)
	}

	issuer, ok := tr.issuers[issuerID]
	if !ok {
		return nil, fmt.Errorf("issuer %s not found", issuerID)
	}

	if issuer.Status != "active" {
		return nil, fmt.Errorf("issuer %s is not active", issuerID)
	}

	return issuer, nil
}

// GetIssuerByDID retrieves an issuer by DID
func (tr *TrustRegistryService) GetIssuerByDID(did string) (*TrustedIssuer, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	issuer, ok := tr.dids[did]
	if !ok {
		return nil, fmt.Errorf("issuer with DID %s not found", did)
	}

	// Check if active/not revoked
	if !tr.revokedIssuers[issuer.ID] && issuer.Status == "active" {
		return issuer, nil
	}

	return nil, fmt.Errorf("issuer with DID %s is not active", did)
}

// VerifyCredentialType checks if an issuer can issue a specific credential type
func (tr *TrustRegistryService) VerifyCredentialType(issuerID string, credType CredentialType) (bool, error) {
	issuer, err := tr.GetIssuer(issuerID)
	if err != nil {
		return false, err
	}

	for _, ct := range issuer.CredentialTypes {
		if ct == credType {
			return true, nil
		}
	}

	return false, fmt.Errorf("issuer does not support credential type %s", credType)
}

// SuspendIssuer temporarily suspends an issuer
func (tr *TrustRegistryService) SuspendIssuer(issuerID string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, ok := tr.issuers[issuerID]; !ok {
		return fmt.Errorf("issuer %s not found", issuerID)
	}

	tr.suspendedIssuers[issuerID] = time.Now()
	return nil
}

// ResumeIssuer resumes a suspended issuer
func (tr *TrustRegistryService) ResumeIssuer(issuerID string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, ok := tr.suspendedIssuers[issuerID]; !ok {
		return fmt.Errorf("issuer %s is not suspended", issuerID)
	}

	delete(tr.suspendedIssuers, issuerID)
	return nil
}

// RevokeIssuer permanently revokes an issuer
func (tr *TrustRegistryService) RevokeIssuer(issuerID string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, ok := tr.issuers[issuerID]; !ok {
		return fmt.Errorf("issuer %s not found", issuerID)
	}

	tr.revokedIssuers[issuerID] = true
	return nil
}

// GetIssuersByJurisdiction retrieves all issuers for a jurisdiction
func (tr *TrustRegistryService) GetIssuersByJurisdiction(jurisdiction Jurisdiction) []*TrustedIssuer {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var result []*TrustedIssuer
	for _, issuer := range tr.issuers {
		if (issuer.Jurisdiction == jurisdiction || issuer.Jurisdiction == JurisdictionGlobal) &&
			issuer.Status == "active" &&
			!tr.revokedIssuers[issuer.ID] {
			result = append(result, issuer)
		}
	}
	return result
}

// GetIssuersByType retrieves all issuers of a specific type
func (tr *TrustRegistryService) GetIssuersByType(issuerType IssuerType) []*TrustedIssuer {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var result []*TrustedIssuer
	for _, issuer := range tr.issuers {
		if issuer.Type == issuerType && issuer.Status == "active" && !tr.revokedIssuers[issuer.ID] {
			result = append(result, issuer)
		}
	}
	return result
}

// VerifyIssuerQualifications checks if issuer has valid qualifications
func (tr *TrustRegistryService) VerifyIssuerQualifications(issuerID string) (bool, []string) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	issuer, ok := tr.issuers[issuerID]
	if !ok {
		return false, nil
	}

	var validQualifications []string
	now := time.Now()

	for _, qual := range issuer.Qualifications {
		if qual.Verified && now.Before(qual.ExpiryDate) {
			validQualifications = append(validQualifications, qual.Type)
		}
	}

	return len(validQualifications) > 0, validQualifications
}

// UpdateIssuerStatus updates an issuer's status
func (tr *TrustRegistryService) UpdateIssuerStatus(issuerID, status string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	issuer, ok := tr.issuers[issuerID]
	if !ok {
		return fmt.Errorf("issuer %s not found", issuerID)
	}

	issuer.Status = status
	issuer.LastVerificationDate = time.Now()
	return nil
}

// GetActiveTrustedIssuers returns all active and trusted issuers
func (tr *TrustRegistryService) GetActiveTrustedIssuers() []*TrustedIssuer {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	var result []*TrustedIssuer
	for _, issuer := range tr.issuers {
		if issuer.Status == "active" && !tr.revokedIssuers[issuer.ID] {
			result = append(result, issuer)
		}
	}
	return result
}

// Helper to initialize well-known issuers
func (tr *TrustRegistryService) initializeWellKnownIssuers() {
	// Government ID issuer (highest trust)
	govIssuer := &TrustedIssuer{
		ID:                  "gvt_us_dmv",
		Name:                "US DMV",
		DID:                 "did:key:z6MkhaXgBZDvotDkL5257faWxcqACaZiarbKhaWc6nMWaveJ",
		Type:                IssuerTypeGovernment,
		Jurisdiction:        JurisdictionUS,
		TrustLevel:          TrustLevelHigh,
		Status:              "active",
		CredentialTypes:     []CredentialType{CredTypePassport, CredTypeDriversLicense, CredTypeProofOfAddress},
		RiskScore:           5,
		OnboardingWeight:    25,
		ActivationThreshold: 0.8,
		EstablishedDate:     time.Now().AddDate(-10, 0, 0),
	}
	tr.issuers[govIssuer.ID] = govIssuer
	tr.dids[govIssuer.DID] = govIssuer

	// Bank issuer (high trust)
	bankIssuer := &TrustedIssuer{
		ID:                  "bank_wellsfargo",
		Name:                "Wells Fargo",
		DID:                 "did:key:z6MkiY6ZCUS5Z3VVzJ4v9V4v3L3L3L3L3L3L3L3L3L3L3L3L",
		Type:                IssuerTypeFinancial,
		Jurisdiction:        JurisdictionUS,
		TrustLevel:          TrustLevelHigh,
		Status:              "active",
		CredentialTypes:     []CredentialType{CredTypeBankAccount},
		RiskScore:           10,
		OnboardingWeight:    20,
		ActivationThreshold: 0.75,
		EstablishedDate:     time.Now().AddDate(-15, 0, 0),
	}
	tr.issuers[bankIssuer.ID] = bankIssuer
	tr.dids[bankIssuer.DID] = bankIssuer

	// University issuer (medium-high trust)
	eduIssuer := &TrustedIssuer{
		ID:                  "edu_stanford",
		Name:                "Stanford University",
		DID:                 "did:key:z6MkpTHR8VNsBxYAAoHx7qSQ7CnjQSXMu7v8ZzW4FYW1iP2L",
		Type:                IssuerTypeEducational,
		Jurisdiction:        JurisdictionUS,
		TrustLevel:          TrustLevelMedium,
		Status:              "active",
		CredentialTypes:     []CredentialType{CredTypeEducationVerification},
		RiskScore:           15,
		OnboardingWeight:    15,
		ActivationThreshold: 0.65,
		EstablishedDate:     time.Now().AddDate(-100, 0, 0),
	}
	tr.issuers[eduIssuer.ID] = eduIssuer
	tr.dids[eduIssuer.DID] = eduIssuer
}
