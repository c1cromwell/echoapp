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

	// Cardano blockchain settings
	CardanoConfig CardanoConfig

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
}

// CardanoConfig contains Cardano blockchain settings
type CardanoConfig struct {
	NetworkID    string // "testnet", "mainnet", "sidechain"
	NodeURL      string
	APIKey       string
	APISecret    string
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration

	// DID anchor settings
	EnableAnchor        bool
	AnchorConfirmations uint64

	// Revocation registry settings
	RevocationRegistryAddress string
	RevocationIndexFile       string
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

// RevocationConfig contains revocation management settings
type RevocationConfig struct {
	Enabled           bool
	RegistryType      string // "cardano", "postgres", "in-memory"
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
			ProofOfHumanityExpiration: 365 * 24 * time.Hour,
			KYCLiteExpiration:         7 * 24 * time.Hour,
			HighAssuranceExpiration:   5 * 365 * 24 * time.Hour,
			ProfessionalExpiration:    2 * 365 * 24 * time.Hour,
			IssuanceTimeout:           60 * time.Second,
			VerificationTimeout:       5 * time.Second,
			SupportedFormats:          []CredentialFormat{JSONLDFormat, JWTFormat, SDJWTFormat},
			DefaultFormat:             JSONLDFormat,
			EnableBlockchainStorage:   true,
			StoragePath:               "/tmp/credentials",
		},
		CardanoConfig: CardanoConfig{
			NetworkID:           "testnet",
			Timeout:             30 * time.Second,
			MaxRetries:          3,
			RetryBackoff:        1 * time.Second,
			EnableAnchor:        true,
			AnchorConfirmations: 5,
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
			RegistryType:      "cardano",
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

	// Cardano settings
	if val := os.Getenv("CARDANO_NETWORK_ID"); val != "" {
		config.CardanoConfig.NetworkID = val
	}
	if val := os.Getenv("CARDANO_NODE_URL"); val != "" {
		config.CardanoConfig.NodeURL = val
	}
	if val := os.Getenv("CARDANO_API_KEY"); val != "" {
		config.CardanoConfig.APIKey = val
	}
	if val := os.Getenv("CARDANO_API_SECRET"); val != "" {
		config.CardanoConfig.APISecret = val
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

	// Validate Cardano settings
	if c.CardanoConfig.NodeURL == "" {
		errors.Add("cardano_node_url", "Cardano node URL is required", "MISSING_CARDANO_NODE_URL")
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
