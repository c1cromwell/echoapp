// Command migrate applies SQL migrations from ./migrations to PostgreSQL.
// Usage: DATABASE_HOST=localhost DATABASE_NAME=echoapp DATABASE_USER=echoapp DATABASE_PASSWORD=echoapp_dev go run ./cmd/migrate
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func main() {
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		log.Fatal("DATABASE_HOST is required")
	}
	port := os.Getenv("DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	cfg := database.PostgresConfig{
		Host:     host,
		Port:     port,
		Database: os.Getenv("DATABASE_NAME"),
		User:     os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		SSLMode:  os.Getenv("DATABASE_SSLMODE"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgDB, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pgDB.Close()

	migrationsDir := filepath.Join(".", "migrations")
	if err := database.Migrate(ctx, pgDB.Pool(), migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("Migrations applied successfully")
}
