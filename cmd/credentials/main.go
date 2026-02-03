package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
)

const (
	appName = "Credentials Service"
	version = "1.0.0"
)

func main() {
	// Load configuration
	config := credentials.LoadConfig()

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("%s v%s starting", appName, version)

	// Initialize credentials service
	service, err := credentials.NewService(config)
	if err != nil {
		log.Fatalf("Failed to initialize credentials service: %v", err)
	}
	defer service.Close()

	log.Println("Credentials service initialized successfully")

	// Create Gin router
	if config.LoggingConfig.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(timeoutMiddleware(config.CredentialConfig.VerificationTimeout))

	// Register credential handlers
	handlers := credentials.NewHandlers(service)
	handlers.RegisterRoutes(router)

	// Register OIDC4VC handlers if enabled
	if config.OIDC4VCConfig.Enabled {
		log.Println("Initializing OIDC4VC endpoints")

		// Create OIDC4VC issuer
		issuer := oidc4vc.NewIssuer(
			config.IssuerConfig.IssuerDID,
			config.VerifierConfig.VerifierDID,
			config.OIDC4VCConfig.IssuerBaseURL,
			config.OIDC4VCConfig.VerifierBaseURL,
		)
		issuer.RegisterRoutes(router)

		// Create OIDC4VC verifier
		verifier := oidc4vc.NewVerifier(
			config.VerifierConfig.VerifierDID,
			config.IssuerConfig.IssuerDID,
			config.OIDC4VCConfig.VerifierBaseURL,
			config.OIDC4VCConfig.IssuerBaseURL,
		)
		verifier.RegisterRoutes(router)

		log.Println("OIDC4VC endpoints registered successfully")
	}

	// Health and readiness endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if err := service.GetStorageHealth(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready":     false,
				"reason":    err.Error(),
				"timestamp": time.Now().Unix(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"ready":     true,
			"timestamp": time.Now().Unix(),
		})
	})

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     appName,
			"version": version,
		})
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  config.ServerConfig.ReadTimeout,
		WriteTimeout: config.ServerConfig.WriteTimeout,
	}

	log.Printf("Starting HTTP server on %s", addr)

	// Start server in goroutine
	go func() {
		var err error
		if config.ServerConfig.TLSEnabled {
			err = httpServer.ListenAndServeTLS(
				config.ServerConfig.TLSCertPath,
				config.ServerConfig.TLSKeyPath,
			)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v", sig)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.ServerConfig.ShutdownTimeout)
	defer cancel()

	log.Println("Shutting down server...")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// corsMiddleware provides CORS support
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// timeoutMiddleware adds request timeout
func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
