package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/thechadcromwell/echoapp/internal/complyserver"
	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
)

func main() {
	cfg := complyserver.DefaultConfig()
	db, _ := initDatabase()
	notif := notification.NewService(db)
	handler := complyserver.NewHandler(db, notif, cfg.ServiceToken)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("comply listen: %v", err)
	}
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("ECHO Comply service listening on :%s (WO-250/251/252)", cfg.Port)

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("comply serve: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func initDatabase() (database.DB, *database.PostgresDB) {
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		log.Println("DATABASE_HOST not set — Comply using in-memory database")
		return database.NewMemoryDB(), nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pg, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		log.Printf("Postgres unavailable (%v) — in-memory fallback", err)
		return database.NewMemoryDB(), nil
	}
	migrationsDir := filepath.Join(".", "migrations")
	if _, err := os.Stat(migrationsDir); err == nil {
		if err := database.Migrate(ctx, pg.Pool(), migrationsDir); err != nil {
			log.Printf("migration warning: %v", err)
		}
	}
	log.Printf("Comply connected to Postgres at %s:%s", cfg.Host, cfg.Port)
	return pg, pg
}

func init() {
	if os.Getenv("COMPLY_PORT") == "" {
		_ = os.Setenv("COMPLY_PORT", "8011")
	}
}
