package bots

import (
	"os"
	"strings"
	"sync"
)

// TokenValidator checks bot API tokens (WO-11 dev slice).
type TokenValidator struct {
	mu     sync.RWMutex
	tokens map[string]string // botDID -> token
}

// NewTokenValidator creates an empty validator.
func NewTokenValidator() *TokenValidator {
	return &TokenValidator{tokens: make(map[string]string)}
}

// NewTokenValidatorFromEnv loads `ECHO_BOT_TOKENS` as `bot_did:token` comma-separated pairs.
func NewTokenValidatorFromEnv() *TokenValidator {
	v := NewTokenValidator()
	raw := os.Getenv("ECHO_BOT_TOKENS")
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		v.SetToken(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	// Dev tokens for catalog bots when env unset.
	if len(v.tokens) == 0 {
		for _, m := range DefaultCatalog() {
			v.SetToken(m.BotDID, "dev-token-"+m.BotDID[len(m.BotDID)-8:])
		}
	}
	return v
}

// SetToken registers a bot API token.
func (v *TokenValidator) SetToken(botDID, token string) {
	if v == nil || botDID == "" || token == "" {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.tokens == nil {
		v.tokens = make(map[string]string)
	}
	v.tokens[botDID] = token
}

// Validate reports whether token is valid for botDID.
func (v *TokenValidator) Validate(botDID, token string) bool {
	if v == nil || botDID == "" || token == "" {
		return false
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.tokens[botDID] == token
}
