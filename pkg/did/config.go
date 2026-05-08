package did

import (
	"os"
	"time"
)

// Config holds the configuration for the DID service.
//
// Phase 1 (per ADR-0001): the DID method is did:key (P-256), which is
// self-certifying — there is no external DID resolver RPC or proof-of-stake
// chain client in the registration path. Anchoring of issuance records / trust commitments / VC
// status lives on the Constellation Identity Metagraph and is configured
// via pkg/credentials.MetagraphConfig.
type Config struct {
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

// DIDConfig holds DID specific settings
type DIDConfig struct {
	Method             string        // "key"
	Network            string        // unused for did:key (kept for forward-compatibility with did:web etc.)
	GenerationTimeout  time.Duration // 30 seconds
	ResolutionTimeout  time.Duration // 2 seconds
	AnchoringTimeout   time.Duration // unused for did:key
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
		DID: DIDConfig{
			Method:             "key",
			Network:            "",
			GenerationTimeout:  30 * time.Second,
			ResolutionTimeout:  2 * time.Second,
			AnchoringTimeout:   60 * time.Second,
			SupportedKeyTypes:  []string{"JsonWebKey2020"},
			DIDDocumentVersion: "2024-01",
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
	cfg := DefaultConfig()

	if portStr := os.Getenv("DID_SERVER_PORT"); portStr != "" {
		_ = portStr // parsed elsewhere
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

	if c.DID.Method != "key" {
		errors.Add("did.method", "method must be 'key' (Phase 1 — see ADR-0001)")
	}

	if c.DID.GenerationTimeout < 10*time.Second {
		errors.Add("did.generation_timeout", "generation timeout must be at least 10 seconds")
	}

	if c.Cache.TTL < 1*time.Hour {
		errors.Add("cache.ttl", "cache TTL must be at least 1 hour")
	}

	return errors
}
