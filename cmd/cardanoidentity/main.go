package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/identity"
)

const (
	appName = "Cardano Identity Service"
	version = "1.0.0"
)

func main() {
	// Load configuration from environment
	host := getEnv("API_HOST", "localhost")
	port := getEnvInt("API_PORT", 8003)
	tlsEnabled := getEnvBool("TLS_ENABLED", false)
	tlsCert := getEnv("TLS_CERT_PATH", "")
	tlsKey := getEnv("TLS_KEY_PATH", "")
	cardanoURL := getEnv("CARDANO_URL", "http://localhost:8090")
	logLevel := getEnv("LOG_LEVEL", "info")

	log.Printf("%s v%s starting", appName, version)

	// Create Gin router
	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(timeoutMiddleware(30 * time.Second))

	// Initialize Cardano client
	cardanoClient := cardano.NewClient(cardanoURL)

	// Initialize identity services
	identityService := identity.NewService(cardanoClient)

	// Register handlers
	handlers := identity.NewHandlers(identityService)
	handlers.RegisterRoutes(router)

	// Health and readiness endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if err := cardanoClient.Health(context.Background()); err != nil {
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
	addr := fmt.Sprintf("%s:%d", host, port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting HTTP server on %s", addr)

	// Start server in goroutine
	go func() {
		var err error
		if tlsEnabled {
			err = httpServer.ListenAndServeTLS(tlsCert, tlsKey)
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// parseDurationOrDefault parses a duration or returns the default
func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
