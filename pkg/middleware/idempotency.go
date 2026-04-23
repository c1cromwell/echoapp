package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

var hex16Pattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// IdempotencyMiddleware caches successful responses keyed by Idempotency-Key header.
// TTL: 7 days. Semantics: same key + same body → cached response; same key + different body → 409.
type IdempotencyMiddleware struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewIdempotencyMiddleware(redisClient *redis.Client) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		redis: redisClient,
		ttl:   7 * 24 * time.Hour,
	}
}

type cachedResponse struct {
	Status      int    `json:"status"`
	Body        []byte `json:"body"`
	ContentType string `json:"content_type"`
	RequestHash string `json:"request_hash"`
}

// Handle wraps a handler with idempotency enforcement.
// Requests without an Idempotency-Key header pass through unchanged.
func (m *IdempotencyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !hex16Pattern.MatchString(key) {
			writeIdempotencyErr(w, http.StatusBadRequest,
				"INVALID_IDEMPOTENCY_KEY",
				"Idempotency-Key must be 32 hex characters (128-bit hex)")
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeIdempotencyErr(w, http.StatusBadRequest, "READ_BODY_FAILED", err.Error())
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		reqHash := sha256Hex(bodyBytes)
		cacheKey := "idem:" + r.URL.Path + ":" + key
		ctx := r.Context()

		if raw, err := m.redis.Get(ctx, cacheKey).Bytes(); err == nil {
			var entry cachedResponse
			if json.Unmarshal(raw, &entry) == nil {
				if entry.RequestHash != reqHash {
					writeIdempotencyErr(w, http.StatusConflict,
						"IDEMPOTENCY_KEY_REUSED",
						"Idempotency-Key was previously used with a different request body")
					return
				}
				w.Header().Set("Content-Type", entry.ContentType)
				w.Header().Set("Idempotent-Replayed", "true")
				w.WriteHeader(entry.Status)
				_, _ = w.Write(entry.Body)
				return
			}
		}

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status >= 200 && rec.status < 300 {
			entry := cachedResponse{
				Status:      rec.status,
				Body:        rec.body.Bytes(),
				ContentType: rec.Header().Get("Content-Type"),
				RequestHash: reqHash,
			}
			if marshaled, err := json.Marshal(entry); err == nil {
				m.redis.Set(ctx, cacheKey, marshaled, m.ttl)
			}
		}
	})
}

// responseRecorder captures the status and body so they can be cached.
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func writeIdempotencyErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"code": code, "message": msg})
}
