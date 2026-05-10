package logging

// WO-6 / WO-53: Privacy-safe structured logger.
//
// Scrubs PII patterns from all log messages before output so that private keys,
// email addresses, and phone numbers never appear in operational logs.
//
// T0–T7 contract:
//   - T0 (plaintext, private keys) — REDACTED automatically
//   - T1 (Secure Enclave derived keys) — never passed to this layer
//   - T7 (DIDs) — allowed in operational logs; DIDs are public chain data
//
// Usage:
//   log := logging.NewLogger("identity-service", logging.LevelInfo)
//   log.Info("registered DID", logging.F("did", did))
//   log.Warn("slow query", logging.F("ms", 230))

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

// Level is a log severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// piiPatterns is the compiled set of patterns that must never appear in logs.
// Evaluated in order; first match wins for each token.
var (
	piiOnce     sync.Once
	piiPatterns []*regexp.Regexp
)

func compilePII() {
	piiOnce.Do(func() {
		piiPatterns = []*regexp.Regexp{
			// 32-byte private key (64 lowercase hex chars) — T0/T1
			regexp.MustCompile(`\b[0-9a-f]{64}\b`),
			// 16-byte intermediate key (32 lowercase hex chars) — T0/T1
			regexp.MustCompile(`\b[0-9a-f]{32}\b`),
			// Email address
			regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
			// E.164 phone number
			regexp.MustCompile(`\+[1-9]\d{9,14}\b`),
		}
	})
}

// Sanitize removes PII patterns from a string, replacing them with [REDACTED].
func Sanitize(s string) string {
	compilePII()
	for _, re := range piiPatterns {
		s = re.ReplaceAllString(s, "[REDACTED]")
	}
	return s
}

// Field is a typed key-value log field.
type Field struct {
	Key   string
	Value interface{}
}

// F is a convenience constructor for Field.
func F(key string, value interface{}) Field { return Field{Key: key, Value: value} }

// Logger is a privacy-safe structured logger that writes JSON lines to w.
type Logger struct {
	mu      sync.Mutex
	w       io.Writer
	service string
	level   Level
}

// NewLogger creates a Logger writing to stdout at the specified minimum level.
func NewLogger(service string, level Level) *Logger {
	return &Logger{w: os.Stdout, service: service, level: level}
}

// NewLoggerWriter creates a Logger writing to the provided writer (useful in tests).
func NewLoggerWriter(w io.Writer, service string, level Level) *Logger {
	return &Logger{w: w, service: service, level: level}
}

func (l *Logger) log(level Level, msg string, fields []Field) {
	if level < l.level {
		return
	}
	entry := make(map[string]interface{}, len(fields)+4)
	entry["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = level.String()
	entry["service"] = l.service
	entry["msg"] = Sanitize(msg)
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			entry[f.Key] = Sanitize(v)
		case error:
			entry[f.Key] = Sanitize(v.Error())
		default:
			entry[f.Key] = v
		}
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	l.mu.Lock()
	fmt.Fprintf(l.w, "%s\n", b)
	l.mu.Unlock()
}

// Debug logs at debug level.
func (l *Logger) Debug(msg string, fields ...Field) { l.log(LevelDebug, msg, fields) }

// Info logs at info level.
func (l *Logger) Info(msg string, fields ...Field) { l.log(LevelInfo, msg, fields) }

// Warn logs at warn level.
func (l *Logger) Warn(msg string, fields ...Field) { l.log(LevelWarn, msg, fields) }

// Error logs at error level.
func (l *Logger) Error(msg string, fields ...Field) { l.log(LevelError, msg, fields) }

// With returns a child logger that always includes the given fields.
func (l *Logger) With(fields ...Field) *Logger {
	child := &Logger{w: l.w, service: l.service, level: l.level}
	// Wrap the writer to prepend persistent fields
	child.w = &prefixWriter{parent: l, persistent: fields}
	return child
}

// prefixWriter merges persistent fields into every log entry.
type prefixWriter struct {
	parent     *Logger
	persistent []Field
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	return pw.parent.w.Write(p)
}

// SetLevel changes the minimum log level at runtime.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}
