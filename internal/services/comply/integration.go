package comply

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/evidence"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// IntegrationConfig holds optional Data L1 and Digital Evidence wiring (Priority 2).
type IntegrationConfig struct {
	DataL1         DataL1Submitter
	Evidence       evidence.EvidenceSubmitter
	EvidenceConfig *evidence.ClientConfig
}

// LoadIntegrationFromEnv wires metagraph Data L1 and Digital Evidence clients from env.
func LoadIntegrationFromEnv() IntegrationConfig {
	var out IntegrationConfig

	if d1 := strings.TrimSpace(os.Getenv("DATA_L1_URL")); d1 != "" {
		cfg := metagraph.MetagraphConfig{
			DataL1URL: d1,
			Timeout:   30 * time.Second,
		}
		if pemPath := strings.TrimSpace(os.Getenv("IDENTITY_SERVICE_KEY_PEM")); pemPath != "" {
			if pem, err := os.ReadFile(pemPath); err != nil {
				log.Printf("comply: Data L1 signing disabled: read key: %v", err)
			} else if signer, err := metagraph.LoadIdentitySigningFromPEM(pem); err != nil {
				log.Printf("comply: Data L1 signing disabled: %v", err)
			} else {
				cfg.IdentitySigner = &signer
			}
		}
		out.DataL1 = NewMetagraphAnchor(metagraph.NewMetagraphClient(cfg))
		log.Printf("comply: Data L1 anchoring enabled (%s)", d1)
	}

	if deCfg, err := evidence.LoadClientConfigFromEnv(); err == nil {
		if client, err := evidence.NewHTTPClient(deCfg); err == nil {
			out.Evidence = client
			out.EvidenceConfig = deCfg
			log.Println("comply: Digital Evidence API client enabled")
		} else {
			log.Printf("comply: Digital Evidence client init failed: %v", err)
		}
	}

	return out
}
