package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitize_PrivateKeyRedacted(t *testing.T) {
	key := strings.Repeat("a", 64) // 32-byte lowercase hex private key
	got := Sanitize("key=" + key)
	if strings.Contains(got, key) {
		t.Error("64-char hex key should be redacted")
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Error("expected [REDACTED] marker")
	}
}

func TestSanitize_EmailRedacted(t *testing.T) {
	got := Sanitize("user chad.cromwell@gmail.com registered")
	if strings.Contains(got, "@") {
		t.Error("email should be redacted")
	}
}

func TestSanitize_PhoneRedacted(t *testing.T) {
	got := Sanitize("phone +14155552671 verified")
	if strings.Contains(got, "+14155552671") {
		t.Error("phone number should be redacted")
	}
}

func TestSanitize_DIDNotRedacted(t *testing.T) {
	did := "did:key:zDnaeSmR7pPxd2kTbq5ir8q9sZcn3CWM2E7cUrUJhUMNwQKx1"
	got := Sanitize("registered " + did)
	if !strings.Contains(got, did) {
		t.Errorf("DID should NOT be redacted (T7 public chain data), got: %s", got)
	}
}

func TestSanitize_NormalTextUnchanged(t *testing.T) {
	msg := "request processed in 12ms"
	if Sanitize(msg) != msg {
		t.Error("normal text should not be modified")
	}
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	log := NewLoggerWriter(&buf, "test-service", LevelDebug)
	log.Info("hello world", F("count", 42), F("action", "login"))

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output must be valid JSON: %v\noutput: %s", err, buf.String())
	}
	if entry["msg"] != "hello world" {
		t.Errorf("msg field wrong: %v", entry["msg"])
	}
	if entry["level"] != "info" {
		t.Errorf("level field wrong: %v", entry["level"])
	}
	if entry["service"] != "test-service" {
		t.Errorf("service field wrong: %v", entry["service"])
	}
}

func TestLogger_PiiInFieldRedacted(t *testing.T) {
	var buf bytes.Buffer
	log := NewLoggerWriter(&buf, "svc", LevelInfo)
	email := "test@example.com"
	log.Info("user action", F("contact", email))

	if strings.Contains(buf.String(), email) {
		t.Error("email in field value should be redacted")
	}
}

func TestLogger_BelowLevelSuppressed(t *testing.T) {
	var buf bytes.Buffer
	log := NewLoggerWriter(&buf, "svc", LevelWarn)
	log.Info("this should not appear")
	log.Debug("this should not appear either")
	if buf.Len() > 0 {
		t.Errorf("below-level messages should be suppressed, got: %s", buf.String())
	}
	log.Warn("this should appear")
	if buf.Len() == 0 {
		t.Error("warn message should appear")
	}
}

func TestLogger_ErrorLevelWithErrorField(t *testing.T) {
	var buf bytes.Buffer
	log := NewLoggerWriter(&buf, "svc", LevelError)
	log.Error("operation failed", F("err", errTest("some error")))
	if !strings.Contains(buf.String(), "operation failed") {
		t.Error("error message should appear")
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }
