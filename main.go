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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
	applog "github.com/thechadcromwell/echoapp/internal/logging"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/rewards"
	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/services/media"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
	"github.com/thechadcromwell/echoapp/internal/services/onboarding"
	rewardsSvc "github.com/thechadcromwell/echoapp/internal/services/rewards"
	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
	"github.com/thechadcromwell/echoapp/pkg/passport"
	"github.com/thechadcromwell/echoapp/pkg/passport/recovery"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

// ServerConfig holds the server configuration.
type ServerConfig struct {
	Port            string
	TLSCertFile     string
	TLSKeyFile      string
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
}

// slog is the process-level structured logger (WO-6 / WO-53).
// Replaces ad-hoc log.Printf calls for structured, PII-safe output.
var slog = applog.NewLogger("server", applog.LevelInfo)

// Server manages the HTTP server lifecycle.
type Server struct {
	config         ServerConfig
	server         *http.Server
	stopLogPublish func() // WO-53: stops background log flush goroutine
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

// validateCORSOrigins rejects a wildcard origin. The CORS layer always sends
// Access-Control-Allow-Credentials: true, so reflecting any origin ("*") would
// let any website make credentialed cross-site requests to the API. Operators
// must configure an explicit allowlist.
func validateCORSOrigins(origins []string) error {
	for _, o := range origins {
		if o == "*" {
			return fmt.Errorf("CORS misconfiguration: wildcard origin \"*\" is not allowed with credentialed requests — configure an explicit allowlist")
		}
	}
	return nil
}

// Start starts the API server.
func (s *Server) Start() error {
	if err := validateCORSOrigins(s.config.AllowedOrigins); err != nil {
		return err
	}

	db, pgDB := s.initDatabase()
	redisClient := s.initRedis()

	router := api.NewRouter(s.config.AllowedOrigins)
	if pgDB != nil {
		router.DIDRegistry = api.NewPostgresDIDRegistry(pgDB.Pool())
		router.CredentialStatusPool = pgDB.Pool()
		store := onboarding.NewPostgresTrustRegistryStore(pgDB)
		if err := router.TrustRegistry.AttachStore(context.Background(), store); err != nil {
			log.Printf("Trust registry PG attach failed, using in-memory: %v", err)
		} else {
			log.Println("Trust registry backed by PostgreSQL (WO-118)")
		}
		router.Passport = passport.NewService(pgDB, pgDB)
		blobStore := encblob.Storage(encblob.NewStubStorage())
		if fallback, err := encblob.NewFallbackStorage(); err == nil {
			blobStore = fallback
		}
		router.PassportSync = passport.NewSyncService(pgDB, blobStore)
		router.PassportRecovery = recovery.NewService(pgDB)
		log.Println("Echo Passport holder refs backed by PostgreSQL (WO-293)")
		log.Println("Echo Passport credential sync enabled (WO-294)")
		log.Println("Echo Passport social recovery enabled (WO-296)")
	}
	router.Redis = redisClient
	if redisClient != nil {
		// Durable revocation blocklist + single-use nonces survive restarts (S3).
		router.TokenService().SetRedisBackend(redisClient)
	}
	if l1 := os.Getenv("IDENTITY_L1_URL"); l1 != "" {
		cfg := metagraph.MetagraphConfig{
			IdentityL1URL: l1,
			Timeout:       30 * time.Second,
		}
		if signer, err := loadIdentitySigningConfig(); err != nil {
			log.Printf("Identity L1 signing disabled: %v", err)
		} else if signer != nil {
			cfg.IdentitySigner = signer
		}
		router.IdentityL1 = metagraph.NewMetagraphClient(cfg)
	}
	if d1 := os.Getenv("DATA_L1_URL"); d1 != "" {
		cfg := metagraph.MetagraphConfig{
			DataL1URL: d1,
			Timeout:   30 * time.Second,
		}
		if signer, err := loadIdentitySigningConfig(); err != nil {
			log.Printf("Data L1 signing disabled: %v", err)
		} else if signer != nil {
			cfg.IdentitySigner = signer
		}
		router.DataL1 = metagraph.NewMetagraphClient(cfg)
	}

	credCfg := credentials.LoadConfig()
	if err := credCfg.Validate(); err != nil {
		log.Printf("Credential issuance disabled: %v", err)
	} else {
		var credPool *pgxpool.Pool
		if pgDB != nil {
			credPool = pgDB.Pool()
		}
		credSvc, err := credentials.NewService(credCfg, credPool)
		if err != nil {
			log.Printf("Credential service failed to start: %v", err)
		} else {
			router.CredentialService = credSvc
			log.Println("Credential issuance enabled (POST /identity/credentials)")
			if credCfg.OIDC4VCConfig.Enabled {
				oidcIss := oidc4vc.NewIssuer(
					credCfg.IssuerConfig.IssuerDID,
					credCfg.VerifierConfig.VerifierDID,
					credCfg.OIDC4VCConfig.IssuerBaseURL,
					credCfg.OIDC4VCConfig.VerifierBaseURL,
				)
				oidcIss.SetCredentialService(credSvc)
				oidcVer := oidc4vc.NewVerifier(
					credCfg.VerifierConfig.VerifierDID,
					credCfg.IssuerConfig.IssuerDID,
					credCfg.OIDC4VCConfig.VerifierBaseURL,
					credCfg.OIDC4VCConfig.IssuerBaseURL,
				)
				oidcVer.SetCredentialService(credSvc)
				g := gin.New()
				g.Use(gin.Recovery())
				oidcIss.RegisterRoutes(g)
				oidcVer.RegisterRoutes(g)
				router.OIDC = g
				router.OIDCVerifier = oidcVer
				router.OIDCVerifierBaseURL = credCfg.OIDC4VCConfig.VerifierBaseURL
				log.Println("OpenID4VCI issuer + verifier mounted (/.well-known/*, /oauth/*, /credential, /verification/*)")
			}
		}
	}

	// Initialize NATS (optional)
	s.initNATS()

	// Initialize storage backend
	storage := s.initStorage()

	// Initialize services and wire V3 handlers
	// WO-44: Per-DID tiered rate limiter (base 100/min; VIP 200/min).
	rateLimiter := infra.NewRateLimiter(infra.DefaultRateLimits())
	router.RateLimiter = rateLimiter

	// S5/S9: per-IP throttle for unauthenticated endpoints + per-phone OTP send cap.
	router.PublicRateLimiter = infra.NewRateLimiter(map[string]infra.RateLimitConfig{
		"public_pre_auth": {MaxRequests: 30, Window: time.Minute},
		"otp_send":        {MaxRequests: 3, Window: 15 * time.Minute},
	})

	// Wave 12: SMS provider — Twilio in prod, stub in dev/test.
	smsProvider, isProd := infra.NewSMSProvider()
	router.SMSProvider = smsProvider
	if isProd {
		slog.Info("SMS provider: Twilio")
	} else {
		slog.Info("SMS provider: stub (set TWILIO_* env vars for production)")
	}

	emission := rewards.NewEmissionSchedule(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	mediaSvc := media.NewService(db, storage)
	mediaSvc.DataL1 = router.DataL1 // D3: anchor media content roots on Data L1

	// D2: private (OPRF-PSI) contact discovery. Degrades gracefully — if the key
	// is unset in production, discovery is disabled rather than failing startup.
	contactsSvc := contacts.NewService(db)
	if oprfSvc, oerr := contacts.NewOPRFService(); oerr != nil {
		log.Printf("Contact discovery disabled: %v", oerr)
	} else {
		contactsSvc.SetOPRF(oprfSvc)
		log.Println("Private contact discovery enabled (OPRF-PSI)")
	}

	router.V3 = &api.V3Handlers{
		DB:           db,
		Contacts:     contactsSvc,
		Notification: notification.NewService(db),
		Media:        mediaSvc,
		Rewards:      rewardsSvc.NewService(db, emission),
		Groups:       groups.NewGroupService(),
		Broadcasts:   broadcast_channels.NewChannelService(),
		RateLimiter:  rateLimiter,
		IdentityL1:   router.IdentityL1, // D1: anchor @username -> DID on registration
	}

	// WO-53: Start audit log publisher background goroutine.
	// Uses FallbackIPFSStorage (Pinata→Storj) when env vars are set;
	// falls back to StubIPFSStorage silently when not configured.
	s.startLogPublisher()

	listener, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.config.Port, err)
	}

	s.server = &http.Server{
		Addr:      ":" + s.config.Port,
		Handler:   router.Handler(),
		TLSConfig: s.setupTLS(),
	}

	slog.Info("API server starting", applog.F("port", s.config.Port), applog.F("tls", "1.3+"))

	go func() {
		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			if err := s.server.ServeTLS(listener, s.config.TLSCertFile, s.config.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", applog.F("err", err))
			}
		} else {
			if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", applog.F("err", err))
			}
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.stopLogPublish != nil {
		s.stopLogPublish()
	}
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// startLogPublisher initialises the WO-53 audit log publisher and starts the
// background flush goroutine.  Requires LOG_MASTER_KEY (64 hex chars = 32 bytes).
// Falls back to a no-op stub when not configured.
func (s *Server) startLogPublisher() {
	masterKeyHex := os.Getenv("LOG_MASTER_KEY")
	if masterKeyHex == "" {
		slog.Warn("LOG_MASTER_KEY not set — audit log publisher disabled (WO-53)")
		s.stopLogPublish = func() {}
		return
	}

	masterKey, epoch, err := applog.DeriveMonthlyKey([]byte(masterKeyHex), time.Now())
	if err != nil {
		log.Printf("audit log key derivation failed: %v — publisher disabled", err)
		s.stopLogPublish = func() {}
		return
	}

	pub, err := applog.NewLogPublisher(masterKey, epoch)
	if err != nil {
		log.Printf("audit log publisher init failed: %v — publisher disabled", err)
		s.stopLogPublish = func() {}
		return
	}

	var storage applog.IPFSStorage
	fallback, ferr := applog.NewFallbackIPFSStorage()
	if ferr != nil {
		storage = &applog.StubIPFSStorage{}
		slog.Info("IPFS storage not configured — using in-memory stub (WO-33)")
	} else {
		storage = fallback
		slog.Info("Audit log publisher started with IPFS storage (WO-53)")
	}

	s.stopLogPublish = pub.StartPeriodicFlush(storage)
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
	case "ipfs":
		storage, err := media.NewIPFSStorage(media.IPFSConfig{
			APIURL: os.Getenv("IPFS_API_URL"),
			Root:   os.Getenv("IPFS_MFS_ROOT"),
		})
		if err != nil {
			log.Printf("Failed to initialize IPFS storage: %v — falling back to memory", err)
			return media.NewMemoryStorage()
		}
		log.Printf("Using IPFS storage at %s", os.Getenv("IPFS_API_URL"))
		return storage
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

func loadIdentitySigningConfig() (*metagraph.IdentitySigningConfig, error) {
	pemPath := os.Getenv("IDENTITY_SERVICE_KEY_PEM")
	if pemPath == "" {
		return nil, nil
	}
	pem, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("read IDENTITY_SERVICE_KEY_PEM: %w", err)
	}
	cfg, err := metagraph.LoadIdentitySigningFromPEM(pem)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
