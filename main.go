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
	"syscall"
	"time"

	"github.com/thechadcromwell/echoapp/internal/api"
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
