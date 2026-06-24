package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

// enforceDIDRateLimit returns false after writing a 429 when the per-DID action
// budget is exhausted (WO-44). Unknown actions are allowed by the limiter.
func (h *V3Handlers) enforceDIDRateLimit(w http.ResponseWriter, r *http.Request, did, action string) bool {
	if h.RateLimiter == nil || did == "" {
		return true
	}
	if err := h.RateLimiter.Check(did, action); err != nil {
		var rl *infra.RateLimitExceededError
		if errors.As(err, &rl) && rl.RetryAfter > 0 {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.RetryAfter.Seconds()))
		}
		WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", err.Error(), r.Header.Get("X-Request-ID"))
		return false
	}
	return true
}
