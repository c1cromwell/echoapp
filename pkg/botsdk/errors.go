package botsdk

import (
	"fmt"
	"time"
)

// BotError is the standard SDK error shape (WO-11).
type BotError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after,omitempty"`
}

func (e *BotError) Error() string {
	if e == nil {
		return ""
	}
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s: %s (retry after %ds)", e.Code, e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

const (
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	CodePermissionDenied  = "PERMISSION_DENIED"
	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeRelayFailed       = "RELAY_FAILED"
)

func newBotError(code, message string, retryAfter int) *BotError {
	return &BotError{Code: code, Message: message, RetryAfter: retryAfter}
}

// IsRateLimit reports whether err is a rate-limit BotError.
func IsRateLimit(err error) bool {
	if be, ok := err.(*BotError); ok {
		return be.Code == CodeRateLimitExceeded
	}
	return false
}

// RetryAfter extracts retry delay from a rate-limit error.
func RetryAfter(err error) time.Duration {
	if be, ok := err.(*BotError); ok && be.RetryAfter > 0 {
		return time.Duration(be.RetryAfter) * time.Second
	}
	return 0
}
