package complyserver

import (
	"net/http"
	"os"
	"time"

	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
)

// Config holds Comply microservice settings.
type Config struct {
	Port           string
	ServiceToken   string
	AllowedOrigins []string
}

// DefaultConfig reads COMPLY_PORT and COMPLY_SERVICE_TOKEN from the environment.
func DefaultConfig() Config {
	port := os.Getenv("COMPLY_PORT")
	if port == "" {
		port = "8011"
	}
	return Config{
		Port:           port,
		ServiceToken:   os.Getenv("COMPLY_SERVICE_TOKEN"),
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8011"},
	}
}

// NewHandler builds the Comply HTTP handler (WO-250 / WO-251 / WO-252 on :8011).
func NewHandler(db database.DB, notif *notification.Service, token string) http.Handler {
	integration := comply.LoadIntegrationFromEnv()
	svc := comply.NewService(db, comply.Deps{
		MessageOps:        db,
		ConversationIndex: db,
		ServiceToken:      token,
		Notifier:          comply.NewPushNotifier(notif),
		DataL1:            integration.DataL1,
		Evidence:          integration.Evidence,
		EvidenceConfig:    integration.EvidenceConfig,
	})
	h := &api.ComplyHandlers{Comply: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"service":   "comply-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	h.RegisterComplyRoutes(mux)
	return mux
}
