package credentials

import (
	"os"
	"strconv"
	"time"
)

// Config represents credentials system configuration
type Config struct {
	// Credential settings
	CredentialConfig CredentialConfig

	// Identity Metagraph anchoring + StatusList2021 publishing settings (WO-272 / WO-274).
	MetagraphConfig MetagraphConfig

	// Issuer settings
	IssuerConfig IssuerConfig

	// Verifier settings
	VerifierConfig VerifierConfig

	// OIDC4VC protocol settings
	OIDC4VCConfig OIDC4VCConfig

	// Revocation settings
	RevocationConfig RevocationConfig

	// Server settings
	ServerConfig ServerConfig

	// Logging settings
	LoggingConfig LoggingConfig
}

// CredentialConfig contains credential-specific configuration
type CredentialConfig struct {
	// Expiration periods for each credential type
	ProofOfHumanityExpiration time.Duration
	KYCLiteExpiration         time.Duration
	HighAssuranceExpiration   time.Duration
	ProfessionalExpiration    time.Duration

	// Issuance timeout
	IssuanceTimeout time.Duration

	// Verification timeout
	VerificationTimeout time.Duration

	// Supported formats
	SupportedFormats []CredentialFormat

	// Default format
	DefaultFormat CredentialFormat

	// Enable blockchain storage
	EnableBlockchainStorage bool

	// Storage path for local credentials
	StoragePath string

	// UseW3CVC2 emits VC 2.0 JSON-LD (validFrom/validUntil, DataIntegrityProof
	// ecdsa-2019, StatusList2021Entry) per WO-274.
	UseW3CVC2 bool

	// StatusListCredentialBaseURL is the base URI for StatusList2021 credentials
	// (index + fragment are derived per issued credential until a publisher assigns slots).
	StatusListCredentialBaseURL string
}

// MetagraphConfig contains Constellation Identity Metagraph settings used by
// the credential issuer + StatusList2021 publisher (Phase 1; ADR-0001).
type MetagraphConfig struct {
	// L0 + L1 endpoints for the Identity Metagraph (env-overridable;
	// see docs/E2E_QUICK_START.md and Makefile dev target).
	IdentityL0URL string
	IdentityL1URL string

	// Authorized issuer DID — Phase 1 always equals the Identity Service
	// did:key. Phase 4+ this expands to a per-org issuer registry proven
	// by EchoOrgRoleCredentials anchored on the same metagraph.
	IssuerDID string

	// HTTP retry knobs for L1 submissions.
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration

	// Anchor settings (W3C VC 2.0 issuance records on Identity L1).
	EnableAnchor bool

	// StatusList2021 publishing.
	StatusListPublishInterval time.Duration // batch cadence; W3C recommends 5 min
	RevocationIndexFile       string        // local cache of bit-vector positions
}

// IssuerConfig contains issuer settings
type IssuerConfig struct {
	IssuerDID           string
	PublicKeyID         string
	PrivateKeyPath      string
	SigningAlgorithm    string // "Ed25519", "ECDSA"
	ProofType           string // "Ed25519Signature2018", "JsonWebSignature2020"
	EnableAutoAnchor    bool
	AnchorDelaySeconds  int
	BatchIssuanceSize   int
	MaxConcurrentIssues int
}

// VerifierConfig contains verifier settings
type VerifierConfig struct {
	VerifierDID                string
	EnableRevocation           bool
	RevocationCacheTTL         time.Duration
	CheckExpiration            bool
	StrictSignature            bool
	TrustRegistry              []string // List of trusted issuer DIDs
	MaxConcurrentVerifications int
}

// OIDC4VCConfig contains OIDC4VC protocol settings
type OIDC4VCConfig struct {
	Enabled                    bool
	IssuerBaseURL              string
	VerifierBaseURL            string
	SupportedProofTypes        []string // ["jwt", "ldp_vc", "ldp_vp"]
	SupportedCredentialFormats []string // ["json-ld", "jwt", "sd-jwt"]
	TokenEndpointTimeout       time.Duration
	AuthorizationCodeTTL       time.Duration
	PreAuthorizedCodeTTL       time.Duration
	AccessTokenTTL             time.Duration
	RefreshTokenTTL            time.Duration
	EnablePKCE                 bool
	RequireProofOfPossession   bool
}

// RevocationConfig contains revocation management settings.
// Phase 1: registry type is "metagraph" (StatusList2021 anchored on the
// Constellation Identity Metagraph). "postgres" / "in-memory" remain
// supported for unit-test backends.
type RevocationConfig struct {
	Enabled           bool
	RegistryType      string // "metagraph", "postgres", "in-memory"
	CacheTTL          time.Duration
	SyncInterval      time.Duration
	MaxCacheSize      int
	CheckFrequency    time.Duration
	LocalIndexPath    string
	EnableBatchChecks bool
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port            int
	Host            string
	TLSEnabled      bool
	TLSCertPath     string
	TLSKeyPath      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxConnections  int
	CORSEnabled     bool
	CORSOrigins     []string
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string // "debug", "info", "warn", "error"
	Format     string // "json", "text"
	OutputPath string // stdout, file path
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		CredentialConfig: CredentialConfig{
			ProofOfHumanityExpiration:   365 * 24 * time.Hour,
			KYCLiteExpiration:           2 * 365 * 24 * time.Hour,
			HighAssuranceExpiration:     5 * 365 * 24 * time.Hour,
			ProfessionalExpiration:      2 * 365 * 24 * time.Hour,
			IssuanceTimeout:             60 * time.Second,
			VerificationTimeout:         5 * time.Second,
			SupportedFormats:            []CredentialFormat{JSONLDFormat, JWTFormat, SDJWTFormat},
			DefaultFormat:               JSONLDFormat,
			EnableBlockchainStorage:     true,
			StoragePath:                 "/tmp/credentials",
			UseW3CVC2:                   true,
			StatusListCredentialBaseURL: "https://identity-metagraph.echo.app/status",
		},
		MetagraphConfig: MetagraphConfig{
			IdentityL0URL:             "http://localhost:9600",
			IdentityL1URL:             "http://localhost:9500",
			Timeout:                   30 * time.Second,
			MaxRetries:                3,
			RetryBackoff:              1 * time.Second,
			EnableAnchor:              true,
			StatusListPublishInterval: 5 * time.Minute,
		},
		IssuerConfig: IssuerConfig{
			SigningAlgorithm:    "Ed25519",
			ProofType:           "Ed25519Signature2018",
			EnableAutoAnchor:    true,
			AnchorDelaySeconds:  10,
			BatchIssuanceSize:   100,
			MaxConcurrentIssues: 10,
		},
		VerifierConfig: VerifierConfig{
			EnableRevocation:           true,
			RevocationCacheTTL:         1 * time.Hour,
			CheckExpiration:            true,
			StrictSignature:            true,
			MaxConcurrentVerifications: 20,
		},
		OIDC4VCConfig: OIDC4VCConfig{
			Enabled:                    true,
			SupportedProofTypes:        []string{"jwt", "ldp_vc", "ldp_vp"},
			SupportedCredentialFormats: []string{"json-ld", "jwt", "sd-jwt"},
			TokenEndpointTimeout:       30 * time.Second,
			AuthorizationCodeTTL:       10 * time.Minute,
			PreAuthorizedCodeTTL:       15 * time.Minute,
			AccessTokenTTL:             1 * time.Hour,
			RefreshTokenTTL:            7 * 24 * time.Hour,
			EnablePKCE:                 true,
			RequireProofOfPossession:   true,
		},
		RevocationConfig: RevocationConfig{
			Enabled:           true,
			RegistryType:      "metagraph",
			CacheTTL:          24 * time.Hour,
			SyncInterval:      1 * time.Hour,
			MaxCacheSize:      10000,
			CheckFrequency:    5 * time.Second,
			EnableBatchChecks: true,
		},
		ServerConfig: ServerConfig{
			Port:            8080,
			Host:            "0.0.0.0",
			TLSEnabled:      false,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
			MaxConnections:  1000,
			CORSEnabled:     true,
			CORSOrigins:     []string{"*"},
		},
		LoggingConfig: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	// Credential settings
	if val := os.Getenv("CRED_POH_EXPIRATION_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			config.CredentialConfig.ProofOfHumanityExpiration = time.Duration(days) * 24 * time.Hour
		}
	}
	if val := os.Getenv("CRED_HA_EXPIRATION_YEARS"); val != "" {
		if years, err := strconv.Atoi(val); err == nil {
			config.CredentialConfig.HighAssuranceExpiration = time.Duration(years) * 365 * 24 * time.Hour
		}
	}
	if val := os.Getenv("CRED_ISSUANCE_TIMEOUT_SECONDS"); val != "" {
		if sec, err := strconv.Atoi(val); err == nil {
			config.CredentialConfig.IssuanceTimeout = time.Duration(sec) * time.Second
		}
	}
	if val := os.Getenv("CRED_VERIFICATION_TIMEOUT_SECONDS"); val != "" {
		if sec, err := strconv.Atoi(val); err == nil {
			config.CredentialConfig.VerificationTimeout = time.Duration(sec) * time.Second
		}
	}
	if val := os.Getenv("CRED_STORAGE_PATH"); val != "" {
		config.CredentialConfig.StoragePath = val
	}
	if val := os.Getenv("CRED_USE_W3C_VC2"); val == "false" || val == "0" {
		config.CredentialConfig.UseW3CVC2 = false
	}
	if val := os.Getenv("CRED_STATUS_LIST_BASE_URL"); val != "" {
		config.CredentialConfig.StatusListCredentialBaseURL = val
	}

	// Identity Metagraph settings (override defaults via IDENTITY_* env vars).
	if val := os.Getenv("IDENTITY_L0_URL"); val != "" {
		config.MetagraphConfig.IdentityL0URL = val
	}
	if val := os.Getenv("IDENTITY_L1_URL"); val != "" {
		config.MetagraphConfig.IdentityL1URL = val
	}
	if val := os.Getenv("IDENTITY_SERVICE_DID"); val != "" {
		config.MetagraphConfig.IssuerDID = val
		if config.IssuerConfig.IssuerDID == "" {
			config.IssuerConfig.IssuerDID = val
		}
	}

	// Issuer settings
	if val := os.Getenv("ISSUER_DID"); val != "" {
		config.IssuerConfig.IssuerDID = val
	}
	if val := os.Getenv("ISSUER_PRIVATE_KEY_PATH"); val != "" {
		config.IssuerConfig.PrivateKeyPath = val
	}
	if val := os.Getenv("ISSUER_PROOF_TYPE"); val != "" {
		config.IssuerConfig.ProofType = val
	}

	// Verifier settings
	if val := os.Getenv("VERIFIER_DID"); val != "" {
		config.VerifierConfig.VerifierDID = val
	}

	// OIDC4VC settings
	if val := os.Getenv("OIDC4VC_ENABLED"); val == "false" || val == "0" {
		config.OIDC4VCConfig.Enabled = false
	}
	if val := os.Getenv("OIDC4VC_ISSUER_BASE_URL"); val != "" {
		config.OIDC4VCConfig.IssuerBaseURL = val
	}
	if val := os.Getenv("OIDC4VC_VERIFIER_BASE_URL"); val != "" {
		config.OIDC4VCConfig.VerifierBaseURL = val
	}
	if val := os.Getenv("OIDC4VC_ENABLE_PKCE"); val != "" {
		config.OIDC4VCConfig.EnablePKCE = val == "true"
	}

	// Server settings
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.ServerConfig.Port = port
		}
	}
	if val := os.Getenv("SERVER_HOST"); val != "" {
		config.ServerConfig.Host = val
	}
	if val := os.Getenv("SERVER_TLS_ENABLED"); val != "" {
		config.ServerConfig.TLSEnabled = val == "true"
	}

	// Logging settings
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.LoggingConfig.Level = val
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	errors := ValidationErrors{}

	// Validate issuer settings
	if c.IssuerConfig.IssuerDID == "" {
		errors.Add("issuer_did", "issuer DID is required", "MISSING_ISSUER_DID")
	}
	if c.IssuerConfig.PrivateKeyPath == "" {
		errors.Add("issuer_private_key_path", "issuer private key path is required", "MISSING_PRIVATE_KEY_PATH")
	}

	// Validate verifier settings
	if c.VerifierConfig.VerifierDID == "" {
		errors.Add("verifier_did", "verifier DID is required", "MISSING_VERIFIER_DID")
	}

	// Validate Identity Metagraph settings
	if c.MetagraphConfig.IdentityL1URL == "" {
		errors.Add("identity_l1_url", "Identity Metagraph L1 URL is required", "MISSING_IDENTITY_L1_URL")
	}

	// Validate OIDC4VC settings if enabled
	if c.OIDC4VCConfig.Enabled {
		if c.OIDC4VCConfig.IssuerBaseURL == "" {
			errors.Add("oidc4vc_issuer_base_url", "OIDC4VC issuer base URL is required", "MISSING_OIDC4VC_ISSUER_BASE_URL")
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}
