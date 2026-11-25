package comply

import (
	"log"
	"os"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/evidence"
	"github.com/thechadcromwell/echoapp/internal/ledger"
)

// IntegrationConfig holds optional Data L1 and Digital Evidence wiring (Priority 2).
type IntegrationConfig struct {
	DataL1         DataL1Submitter
	Evidence       evidence.EvidenceSubmitter
	EvidenceConfig *evidence.ClientConfig
}

// LoadIntegrationFromEnv wires the chain anchor and Digital Evidence clients from env.
func LoadIntegrationFromEnv() IntegrationConfig {
	var out IntegrationConfig

	if anchor, err := ledger.BesuAnchorFromEnv(); err != nil {
		log.Printf("comply: chain anchoring disabled: %v", err)
	} else if anchor != nil {
		out.DataL1 = NewMetagraphAnchor(anchor)
		log.Printf("comply: chain anchoring enabled (contract %s)", anchor.Client.AnchorContract().Hex())
	}

	// Digital Evidence is a Constellation-hosted service: constructed only when
	// DIGITAL_EVIDENCE_ENABLED=true (and the API key is set).
	if !evidence.Enabled() {
		if strings.TrimSpace(os.Getenv("DIGITAL_EVIDENCE_API_KEY")) != "" {
			log.Println("comply: DIGITAL_EVIDENCE_API_KEY is set but DIGITAL_EVIDENCE_ENABLED is not true; Digital Evidence stays off")
		}
	} else if deCfg, err := evidence.LoadClientConfigFromEnv(); err == nil {
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
