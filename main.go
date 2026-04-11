package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/rewards"
	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/services/media"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
	rewardsSvc "github.com/thechadcromwell/echoapp/internal/services/rewards"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	Port            string
	TLSCertFile     string
	TLSKeyFile      string
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
}

// Server manages the HTTP server lifecycle.
type Server struct {
	config ServerConfig
	server *http.Server
}

// NewServer creates a new production server.
func NewServer(config ServerConfig) *Server {
	return &Server{config: config}
}

// setupTLS configures TLS 1.3+ settings.
func (s *Server) setupTLS() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}
}

// Start starts the API server.
func (s *Server) Start() error {
	router := api.NewRouter(s.config.AllowedOrigins)

	// Initialize database
	db, pgDB := s.initDatabase()
	_ = pgDB // stored for graceful shutdown if needed

	// Initialize Redis (optional)
	s.initRedis()

	// Initialize NATS (optional)
	s.initNATS()

	// Initialize storage backend
	storage := s.initStorage()

	// Initialize services and wire V3 handlers
	emission := rewards.NewEmissionSchedule(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	router.V3 = &api.V3Handlers{
		DB:           db,
		Contacts:     contacts.NewService(db),
		Notification: notification.NewService(db),
		Media:        media.NewService(db, storage),
		Rewards:      rewardsSvc.NewService(db, emission),
		Groups:       groups.NewGroupService(),
		Broadcasts:   broadcast_channels.NewChannelService(),
	}

	listener, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.config.Port, err)
	}

	s.server = &http.Server{
		Addr:      ":" + s.config.Port,
		Handler:   router.Handler(),
		TLSConfig: s.setupTLS(),
	}

	log.Printf("API Server starting on port %s (TLS 1.3+)", s.config.Port)

	go func() {
		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			if err := s.server.ServeTLS(listener, s.config.TLSCertFile, s.config.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v", err)
			}
		} else {
			if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v", err)
			}
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// initDatabase connects to PostgreSQL if DATABASE_HOST is set, otherwise falls back to in-memory.
// Also runs migrations when using PostgreSQL.
func (s *Server) initDatabase() (database.DB, *database.PostgresDB) {
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		log.Println("DATABASE_HOST not set, using in-memory database")
		return database.NewMemoryDB(), nil
	}

	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	cfg := database.PostgresConfig{
		Host:     dbHost,
		Port:     dbPort,
		Database: os.Getenv("DATABASE_NAME"),
		User:     os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		SSLMode:  os.Getenv("DATABASE_SSLMODE"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgDB, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v — falling back to in-memory", err)
		return database.NewMemoryDB(), nil
	}
	log.Printf("Connected to PostgreSQL at %s:%s/%s", cfg.Host, cfg.Port, cfg.Database)

	// Run migrations
	migrationsDir := filepath.Join(".", "migrations")
	if _, err := os.Stat(migrationsDir); err == nil {
		if err := database.Migrate(ctx, pgDB.Pool(), migrationsDir); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}

	return pgDB, pgDB
}

// initRedis connects to Redis if REDIS_HOST is set.
func (s *Server) initRedis() *infra.RedisClient {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		log.Println("REDIS_HOST not set, Redis features disabled")
		return nil
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := infra.NewRedisClient(ctx, infra.RedisConfig{
		Host:     host,
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	if err != nil {
		log.Printf("Failed to connect to Redis: %v — Redis features disabled", err)
		return nil
	}
	log.Printf("Connected to Redis at %s:%s", host, port)
	return client
}

// initNATS connects to NATS if NATS_URL is set.
func (s *Server) initNATS() *infra.NATSClient {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Println("NATS_URL not set, NATS event bus disabled")
		return nil
	}

	client, err := infra.NewNATSClient(infra.NATSConfig{
		URL:       natsURL,
		ClusterID: os.Getenv("NATS_CLUSTER_ID"),
	})
	if err != nil {
		log.Printf("Failed to connect to NATS: %v — event bus disabled", err)
		return nil
	}
	log.Printf("Connected to NATS at %s", natsURL)
	return client
}

// initStorage creates the media storage backend based on STORAGE_BACKEND env var.
func (s *Server) initStorage() media.StorageBackend {
	backend := os.Getenv("STORAGE_BACKEND")
	switch backend {
	case "s3", "storj":
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		storage, err := media.NewS3Storage(ctx, media.S3Config{
			Endpoint:        os.Getenv("STORAGE_ENDPOINT"),
			Region:          os.Getenv("STORAGE_REGION"),
			Bucket:          os.Getenv("STORAGE_BUCKET"),
			AccessKeyID:     os.Getenv("STORAGE_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("STORAGE_SECRET_ACCESS_KEY"),
			ForcePathStyle:  os.Getenv("STORAGE_FORCE_PATH_STYLE") != "false",
		})
		if err != nil {
			log.Printf("Failed to initialize S3 storage: %v — falling back to memory", err)
			return media.NewMemoryStorage()
		}
		log.Printf("Using S3-compatible storage: %s/%s", os.Getenv("STORAGE_ENDPOINT"), os.Getenv("STORAGE_BUCKET"))
		return storage
	default:
		log.Println("STORAGE_BACKEND not set, using in-memory media storage")
		return media.NewMemoryStorage()
	}
}

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}

	config := ServerConfig{
		Port:        port,
		TLSCertFile: os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:  os.Getenv("TLS_KEY_FILE"),
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8000",
			"https://app.example.com",
		},
		ShutdownTimeout: 10 * time.Second,
	}

	server := NewServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
