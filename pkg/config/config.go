package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	API      APIConfig
	TLS      TLSConfig
	CORS     CORSConfig
	Auth     AuthConfig
	Cardano  CardanoConfig
	Identity IdentityConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	Storage  StorageConfig
	APNs     APNsConfig
}

// ServerConfig holds server-level configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
}

// APIConfig holds API-level configuration
type APIConfig struct {
	Version               string
	Environment           string
	LogLevel              string
	RequestIDHeader       string
	IncludeRequestIDInLog bool
}

// TLSConfig holds TLS/HTTPS configuration
type TLSConfig struct {
	Enabled    bool
	CertFile   string
	KeyFile    string
	MinVersion string
	MaxVersion string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string
	MaxAge         int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled             bool
	RequireAuth         bool
	TokenType           string // "Bearer", "Basic", "Custom"
	SkipAuthPaths       []string
	PasskeyVerification bool
}

// CardanoConfig holds Cardano blockchain configuration
type CardanoConfig struct {
	// Network configuration
	Network   string // testnet, mainnet, custom
	NodeURL   string // Cardano node RPC endpoint
	WalletURL string // Cardano wallet API endpoint
	NetworkID int    // Network ID (0 for testnet, 1 for mainnet)

	// Wallet configuration
	WalletName     string // Name of the wallet to use
	WalletPassword string // Wallet password (from env)
	WalletMnemonic string // Wallet seed phrase (from env, never logged)
	PaymentAddress string // Payment address for transactions
	StakingAddress string // Staking address

	// Transaction configuration
	TxTimeoutSeconds   int
	ConfirmationBlocks int64
	GasMultiplier      float64
	MinFeeLovelace     int64 // Minimum fee in Lovelace

	// Protocol configuration
	ProtocolVersion string
	EraString       string // Current Cardano era

	// Connection settings
	MaxConnections int
	ConnectionPool int
	RetryAttempts  int
	RetryDelayMs   int
	RequestTimeout time.Duration

	// Feature flags
	EnableSmartContracts bool
	EnableMetadata       bool
	EnableNFT            bool
}

// IdentityConfig holds identity service configuration
type IdentityConfig struct {
	// Cache settings
	CacheTTL        time.Duration
	CacheMaxEntries int
	FallbackTimeout time.Duration // 2-second fallback for queries

	// Trust level settings
	DefaultTrustLevel string
	TrustLevelExpiry  time.Duration
	AllowPromotion    bool
	AllowDemotion     bool
	RequireMultiSig   bool
	MinVerifiers      int

	// Credential settings
	DefaultCredentialExpiry time.Duration
	MaxCredentialsPerUser   int
	AllowCredentialSharing  bool
	VerificationRequired    bool

	// Audit settings
	EnableAuditTrail bool
	AuditRetention   time.Duration
	LogAllOperations bool

	// Verification settings
	DeviceVerificationEnabled       bool
	KYCVerificationEnabled          bool
	OrganizationVerificationEnabled bool
	BiometricVerificationEnabled    bool
	AllowedKYCProviders             []string

	// Rate limiting
	TrustLevelUpdatesPerHour int
	CredentialStoragePerDay  int
	QueriesPerSecond         int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string // postgres, mysql, sqlite
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSL      bool
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NATSConfig holds NATS message bus configuration
type NATSConfig struct {
	URL       string
	ClusterID string
}

// StorageConfig holds S3/Storj media storage configuration
type StorageConfig struct {
	Backend         string // "memory", "s3", "storj"
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
}

// APNsConfig holds Apple Push Notification service configuration
type APNsConfig struct {
	TeamID     string
	KeyID      string
	KeyFile    string // path to .p8 file
	BundleID   string
	Production bool
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("API_PORT", "8000"),
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1 MB
		},
		API: APIConfig{
			Version:               "1.0.0",
			Environment:           getEnv("ENVIRONMENT", "development"),
			LogLevel:              getEnv("LOG_LEVEL", "info"),
			RequestIDHeader:       "X-Request-ID",
			IncludeRequestIDInLog: true,
		},
		TLS: TLSConfig{
			Enabled:    getEnv("TLS_ENABLED", "false") == "true",
			CertFile:   getEnv("TLS_CERT_FILE", ""),
			KeyFile:    getEnv("TLS_KEY_FILE", ""),
			MinVersion: "1.3",
			MaxVersion: "1.3",
		},
		CORS: CORSConfig{
			Enabled: getEnv("CORS_ENABLED", "true") == "true",
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:8000",
				"https://app.example.com",
			},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
			ExposedHeaders: []string{"X-Request-ID", "X-Rate-Limit-Remaining"},
			MaxAge:         3600,
		},
		Auth: AuthConfig{
			Enabled:             true,
			RequireAuth:         true,
			TokenType:           "Bearer",
			PasskeyVerification: true, // ES256 JWT validation via internal/auth.TokenService
			SkipAuthPaths: []string{
				"/health",
				"/health/ready",
				"/health/live",
				"/v1/identity/health",
			},
		},
		Cardano: CardanoConfig{
			Network:              getEnv("CARDANO_NETWORK", "testnet"),
			NodeURL:              getEnv("CARDANO_NODE_URL", "http://localhost:8090"),
			WalletURL:            getEnv("CARDANO_WALLET_URL", "http://localhost:8091"),
			NetworkID:            getIntOrDefault("CARDANO_NETWORK_ID", 0), // 0 for testnet, 1 for mainnet
			WalletName:           getEnv("CARDANO_WALLET_NAME", "identity-wallet"),
			WalletPassword:       getEnv("CARDANO_WALLET_PASSWORD", ""), // From env for security
			PaymentAddress:       getEnv("CARDANO_PAYMENT_ADDRESS", ""),
			StakingAddress:       getEnv("CARDANO_STAKING_ADDRESS", ""),
			TxTimeoutSeconds:     getIntOrDefault("CARDANO_TX_TIMEOUT_SECONDS", 300),
			ConfirmationBlocks:   getInt64OrDefault("CARDANO_CONFIRMATION_BLOCKS", 12),
			GasMultiplier:        getFloatOrDefault("CARDANO_GAS_MULTIPLIER", 1.2),
			MinFeeLovelace:       getInt64OrDefault("CARDANO_MIN_FEE_LOVELACE", 200000),
			ProtocolVersion:      getEnv("CARDANO_PROTOCOL_VERSION", "8.0"),
			EraString:            getEnv("CARDANO_ERA", "Babbage"),
			MaxConnections:       getIntOrDefault("CARDANO_MAX_CONNECTIONS", 10),
			ConnectionPool:       getIntOrDefault("CARDANO_CONNECTION_POOL", 5),
			RetryAttempts:        getIntOrDefault("CARDANO_RETRY_ATTEMPTS", 3),
			RetryDelayMs:         getIntOrDefault("CARDANO_RETRY_DELAY_MS", 1000),
			RequestTimeout:       parseDurationOrDefault("CARDANO_REQUEST_TIMEOUT", 30*time.Second),
			EnableSmartContracts: getEnv("CARDANO_ENABLE_SMART_CONTRACTS", "false") == "true",
			EnableMetadata:       getEnv("CARDANO_ENABLE_METADATA", "true") == "true",
			EnableNFT:            getEnv("CARDANO_ENABLE_NFT", "false") == "true",
		},
		Identity: IdentityConfig{
			CacheTTL:                        parseDurationOrDefault("IDENTITY_CACHE_TTL", 5*time.Minute),
			CacheMaxEntries:                 getIntOrDefault("IDENTITY_CACHE_MAX_ENTRIES", 10000),
			FallbackTimeout:                 parseDurationOrDefault("IDENTITY_FALLBACK_TIMEOUT", 2*time.Second),
			DefaultTrustLevel:               getEnv("IDENTITY_DEFAULT_TRUST_LEVEL", "unverified"),
			TrustLevelExpiry:                parseDurationOrDefault("IDENTITY_TRUST_LEVEL_EXPIRY", 90*24*time.Hour),
			AllowPromotion:                  getEnv("IDENTITY_ALLOW_PROMOTION", "true") == "true",
			AllowDemotion:                   getEnv("IDENTITY_ALLOW_DEMOTION", "true") == "true",
			RequireMultiSig:                 getEnv("IDENTITY_REQUIRE_MULTI_SIG", "false") == "true",
			MinVerifiers:                    getIntOrDefault("IDENTITY_MIN_VERIFIERS", 1),
			DefaultCredentialExpiry:         parseDurationOrDefault("IDENTITY_DEFAULT_CREDENTIAL_EXPIRY", 365*24*time.Hour),
			MaxCredentialsPerUser:           getIntOrDefault("IDENTITY_MAX_CREDENTIALS_PER_USER", 1000),
			AllowCredentialSharing:          getEnv("IDENTITY_ALLOW_CREDENTIAL_SHARING", "true") == "true",
			VerificationRequired:            getEnv("IDENTITY_VERIFICATION_REQUIRED", "true") == "true",
			EnableAuditTrail:                getEnv("IDENTITY_ENABLE_AUDIT_TRAIL", "true") == "true",
			AuditRetention:                  parseDurationOrDefault("IDENTITY_AUDIT_RETENTION", 7*365*24*time.Hour),
			LogAllOperations:                getEnv("IDENTITY_LOG_ALL_OPERATIONS", "false") == "true",
			DeviceVerificationEnabled:       getEnv("IDENTITY_DEVICE_VERIFICATION_ENABLED", "true") == "true",
			KYCVerificationEnabled:          getEnv("IDENTITY_KYC_VERIFICATION_ENABLED", "true") == "true",
			OrganizationVerificationEnabled: getEnv("IDENTITY_ORG_VERIFICATION_ENABLED", "true") == "true",
			BiometricVerificationEnabled:    getEnv("IDENTITY_BIOMETRIC_VERIFICATION_ENABLED", "false") == "true",
			AllowedKYCProviders:             strings.Split(getEnv("IDENTITY_ALLOWED_KYC_PROVIDERS", "jumio,veriff,onfido"), ","),
			TrustLevelUpdatesPerHour:        getIntOrDefault("IDENTITY_TRUST_UPDATES_PER_HOUR", 100),
			CredentialStoragePerDay:         getIntOrDefault("IDENTITY_CREDENTIAL_STORAGE_PER_DAY", 1000),
			QueriesPerSecond:                getIntOrDefault("IDENTITY_QUERIES_PER_SECOND", 100),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DATABASE_TYPE", "postgres"),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getIntOrDefault("DATABASE_PORT", 5432),
			Name:     getEnv("DATABASE_NAME", "echoapp"),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", ""),
			SSL:      getEnv("DATABASE_SSL", "false") == "true",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", ""),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntOrDefault("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			URL:       getEnv("NATS_URL", ""),
			ClusterID: getEnv("NATS_CLUSTER_ID", "echo-cluster"),
		},
		Storage: StorageConfig{
			Backend:         getEnv("STORAGE_BACKEND", "memory"),
			Endpoint:        getEnv("STORAGE_ENDPOINT", ""),
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
			Bucket:          getEnv("STORAGE_BUCKET", "echo-media"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
			ForcePathStyle:  getEnv("STORAGE_FORCE_PATH_STYLE", "true") == "true",
		},
		APNs: APNsConfig{
			TeamID:     getEnv("APNS_TEAM_ID", ""),
			KeyID:      getEnv("APNS_KEY_ID", ""),
			KeyFile:    getEnv("APNS_KEY_FILE", ""),
			BundleID:   getEnv("APNS_BUNDLE_ID", "com.echo.app"),
			Production: getEnv("APNS_PRODUCTION", "false") == "true",
		},
	}
}

// getEnv retrieves an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntOrDefault retrieves an environment variable as int or returns a default value
func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getInt64OrDefault retrieves an environment variable as int64 or returns a default value
func getInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Val, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Val
		}
	}
	return defaultValue
}

// getFloatOrDefault retrieves an environment variable as float64 or returns a default value
func getFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// parseDurationOrDefault parses a duration from environment or returns a default value
func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Cardano configuration
	if c.Cardano.Network == "" {
		return fmt.Errorf("CARDANO_NETWORK is required")
	}

	validNetworks := map[string]bool{"testnet": true, "mainnet": true, "custom": true}
	if !validNetworks[c.Cardano.Network] {
		return fmt.Errorf("invalid CARDANO_NETWORK: %s", c.Cardano.Network)
	}

	if c.Cardano.NodeURL == "" {
		return fmt.Errorf("CARDANO_NODE_URL is required")
	}

	if c.Cardano.WalletURL == "" {
		return fmt.Errorf("CARDANO_WALLET_URL is required")
	}

	// Validate identity configuration
	validTrustLevels := map[string]bool{
		"unverified":            true,
		"device-verified":       true,
		"kyc-verified":          true,
		"organization-verified": true,
	}
	if !validTrustLevels[c.Identity.DefaultTrustLevel] {
		return fmt.Errorf("invalid IDENTITY_DEFAULT_TRUST_LEVEL: %s", c.Identity.DefaultTrustLevel)
	}

	if c.Identity.MinVerifiers < 1 {
		return fmt.Errorf("IDENTITY_MIN_VERIFIERS must be at least 1")
	}

	return nil
}

// GetCardanoNetworkID returns the Cardano network ID as an integer
func (c *CardanoConfig) GetNetworkID() int {
	if c.Network == "mainnet" {
		return 1
	}
	return 0 // testnet
}

// GetCardanoURL returns the appropriate Cardano URL based on network
func (c *CardanoConfig) GetCardanoURL() string {
	if c.NodeURL != "" {
		return c.NodeURL
	}

	switch c.Network {
	case "mainnet":
		return "https://cardano-mainnet.blockfrost.io"
	case "testnet":
		return "https://cardano-testnet.blockfrost.io"
	default:
		return "http://localhost:8090"
	}
}

// IsTLSEnabled returns whether TLS is enabled
func (c *Config) IsTLSEnabled() bool {
	return c.TLS.Enabled && c.TLS.CertFile != "" && c.TLS.KeyFile != ""
}

// GetLogLevel returns the log level as a string
func (c *Config) GetLogLevel() string {
	level := c.API.LogLevel
	if level == "" {
		level = "info"
	}
	return strings.ToLower(level)
}

// GetFullServerAddr returns the full server address with protocol
func (c *Config) GetFullServerAddr() string {
	protocol := "http"
	if c.IsTLSEnabled() {
		protocol = "https"
	}
	return fmt.Sprintf("%s://localhost:%s", protocol, c.Server.Port)
}
