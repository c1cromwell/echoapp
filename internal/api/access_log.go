package api

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Access log: one line per request, for diagnosing a deployment. Off unless ECHO_ACCESS_LOG=true.
//
// It records the method, the path with identifiers removed, the status and the time taken. It
// never records who made the request: no address, no DID, no token, no query string, no body.
// That matters most for sealed delivery, where the point is that the server keeps nothing about
// the caller.

func accessLogEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ECHO_ACCESS_LOG")), "true")
}

// redactPath removes anything in a path that could identify a person or a message: DIDs, long
// opaque ids, and everything after the third segment's worth of fixed route words.
func redactPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "did:") || strings.Contains(p, "%3A") || strings.Contains(p, ":") || len(p) >= 20 || looksLikeID(p) {
			parts[i] = redactedID(p)
		}
	}
	return strings.Join(parts, "/")
}

// redactedID is what replaces an identifier in the log. With ECHO_ACCESS_LOG_ID_TAILS=true (local
// debugging only; never in a deployment that serves real people) it keeps the last six characters
// so two requests about the same identifier can be told apart from two about different ones.
func redactedID(id string) string {
	if strings.EqualFold(os.Getenv("ECHO_ACCESS_LOG_ID_TAILS"), "true") && os.Getenv("ENVIRONMENT") != "production" && len(id) > 6 {
		return "{id…" + id[len(id)-6:] + "}"
	}
	return "{id}"
}

// looksLikeID reports a segment that is mostly digits or hex, such as a UUID or a hash.
func looksLikeID(s string) bool {
	if len(s) < 8 {
		return false
	}
	hexish := 0
	for _, c := range s {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '-' {
			hexish++
		}
	}
	return hexish*10 >= len(s)*9
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	// code is the machine-readable error code of a refused request ("AUTH_STALE_TIMESTAMP"),
	// taken from the response. It names the reason, never the caller.
	code string
}

var errorCodePattern = regexp.MustCompile(`"code":"([A-Z0-9_]{1,48})"`)

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status >= 400 && r.code == "" {
		head := b
		if len(head) > 256 {
			head = head[:256]
		}
		if m := errorCodePattern.FindSubmatch(head); m != nil {
			r.code = string(m[1])
		}
	}
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Hijack lets the WebSocket upgrade through the recorder.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer cannot be hijacked")
	}
	r.status = http.StatusSwitchingProtocols
	return h.Hijack()
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		if rec.code != "" {
			log.Printf("access %s %s %d %s %dms", r.Method, redactPath(r.URL.Path), rec.status, rec.code, time.Since(start).Milliseconds())
			return
		}
		log.Printf("access %s %s %d %dms", r.Method, redactPath(r.URL.Path), rec.status, time.Since(start).Milliseconds())
	})
}

// debugTrace writes a line only when identifier tails are switched on (local debugging). It is for
// following one test message through the relay; it must never be on where real people are served.
func debugTrace(format string, ids ...string) {
	if !strings.EqualFold(os.Getenv("ECHO_ACCESS_LOG_ID_TAILS"), "true") || os.Getenv("ENVIRONMENT") == "production" {
		return
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if len(id) > 6 {
			id = "…" + id[len(id)-6:]
		}
		args[i] = id
	}
	log.Printf("trace "+format, args...)
}
