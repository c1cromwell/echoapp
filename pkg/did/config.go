package did

import (
	"os"
	"time"
)

// Config holds the configuration for the DID service
type Config struct {
	// Atala PRISM settings
	AtalaPRISM AtalaPRISMConfig

	// Cardano settings
	Cardano CardanoConfig

	// DID settings
	DID DIDConfig

	// Cache settings
	Cache CacheConfig

	// Database settings
	Database DatabaseConfig

	// Server settings
	Server ServerConfig

	// Logging settings
	Logging LoggingConfig
}

// AtalaPRISMConfig holds Atala PRISM specific configuration
type AtalaPRISMConfig struct {
	Endpoint       string
	APIKey         string
	APISecret      string
	Timeout        time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
	ConnectionPool int
	VerifySSL      bool
	ProxyURL       string
}

// CardanoConfig holds Cardano blockchain settings
type CardanoConfig struct {
	NetworkID             string // mainnet, testnet, preprod
	NodeURL               string
	NodeTimeout           time.Duration
	ProtocolParameters    string
	MaxTxSize             int
	MinFee                int64
	ConfirmationThreshold int
	BlockPollingInterval  time.Duration
	MaxBlockWaitTime      time.Duration
}

// DIDConfig holds DID specific settings
type DIDConfig struct {
	Method             string        // "prism"
	Network            string        // "cardano"
	GenerationTimeout  time.Duration // 30 seconds
	ResolutionTimeout  time.Duration // 2 seconds
	AnchoringTimeout   time.Duration // 60 seconds
	SupportedKeyTypes  []string
	DIDDocumentVersion string
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	Enabled         bool
	TTL             time.Duration // 24 hours
	MaxSize         int
	CleanupInterval time.Duration
	SyncInterval    time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver         string // "postgres", "mysql", "sqlite"
	Host           string
	Port           int
	Database       string
	User           string
	Password       string
	MaxConnections int
	MaxIdleTime    time.Duration
	ConnectTimeout time.Duration
	SSLMode        string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TLSEnabled      bool
	CertFile        string
	KeyFile         string
	CORS            CORSConfig
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string // "debug", "info", "warn", "error"
	Format     string // "json", "text"
	OutputPath string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AtalaPRISM: AtalaPRISMConfig{
			Endpoint:       "https://prism.atalaprism.io",
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			RetryBackoff:   1 * time.Second,
			ConnectionPool: 10,
			VerifySSL:      true,
		},
		Cardano: CardanoConfig{
			NetworkID:             "testnet",
			NodeURL:               "https://cardano-testnet-node.example.com",
			NodeTimeout:           30 * time.Second,
			ConfirmationThreshold: 6,
			BlockPollingInterval:  5 * time.Second,
			MaxBlockWaitTime:      120 * time.Second,
		},
		DID: DIDConfig{
			Method:             "prism",
			Network:            "cardano",
			GenerationTimeout:  30 * time.Second,
			ResolutionTimeout:  2 * time.Second,
			AnchoringTimeout:   60 * time.Second,
			SupportedKeyTypes:  []string{"Ed25519VerificationKey2018"},
			DIDDocumentVersion: "2022-09",
		},
		Cache: CacheConfig{
			Enabled:         true,
			TTL:             24 * time.Hour,
			MaxSize:         10000,
			CleanupInterval: 1 * time.Hour,
			SyncInterval:    5 * time.Minute,
		},
		Database: DatabaseConfig{
			Driver:         "postgres",
			Host:           "localhost",
			Port:           5432,
			MaxConnections: 25,
			MaxIdleTime:    5 * time.Minute,
			ConnectTimeout: 30 * time.Second,
			SSLMode:        "disable",
		},
		Server: ServerConfig{
			Port:            8080,
			Host:            "0.0.0.0",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			TLSEnabled:      false,
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				MaxAge:         3600,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig(configPath string) (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// Load from environment variables
	if apiKey := os.Getenv("DID_ATALA_PRISM_API_KEY"); apiKey != "" {
		cfg.AtalaPRISM.APIKey = apiKey
	}

	if apiSecret := os.Getenv("DID_ATALA_PRISM_API_SECRET"); apiSecret != "" {
		cfg.AtalaPRISM.APISecret = apiSecret
	}

	if endpoint := os.Getenv("DID_ATALA_PRISM_ENDPOINT"); endpoint != "" {
		cfg.AtalaPRISM.Endpoint = endpoint
	}

	if nodeURL := os.Getenv("DID_CARDANO_NODE_URL"); nodeURL != "" {
		cfg.Cardano.NodeURL = nodeURL
	}

	if networkID := os.Getenv("DID_CARDANO_NETWORK_ID"); networkID != "" {
		cfg.Cardano.NetworkID = networkID
	}

	if portStr := os.Getenv("DID_SERVER_PORT"); portStr != "" {
		// In production, parse the port number
		// For now, keep default
	}

	if logLevel := os.Getenv("DID_LOGGING_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() *ValidationErrors {
	errors := &ValidationErrors{}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors.Add("server.port", "port must be between 1 and 65535")
	}

	if c.AtalaPRISM.Endpoint == "" {
		errors.Add("atala_prism.endpoint", "endpoint is required")
	}

	if c.DID.Method != "prism" {
		errors.Add("did.method", "method must be 'prism'")
	}

	if c.DID.GenerationTimeout < 10*time.Second {
		errors.Add("did.generation_timeout", "generation timeout must be at least 10 seconds")
	}

	if c.Cache.TTL < 1*time.Hour {
		errors.Add("cache.ttl", "cache TTL must be at least 1 hour")
	}

	return errors
}
