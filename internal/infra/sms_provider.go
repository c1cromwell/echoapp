package infra

// WO-wave12: SMS OTP provider for recovery phone-backup registration.
//
// Privacy model: the backend never stores the raw phone number. It stores only
// H(E.164 phone) in the sms_recovery Postgres table. The raw number is
// passed transiently to the SMS provider (Twilio) to dispatch the OTP, then
// discarded. Only H(phone) persists.
//
// Providers:
//   StubSMSProvider  — dev/test: logs OTP, echoes it in X-Dev-OTP response header
//   TwilioSMSProvider — prod scaffold: reads TWILIO_* env vars; returns
//                       ErrSMSNotConfigured when absent (same pattern as IPFS clients)

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrSMSNotConfigured is returned when no SMS provider credentials are present.
var ErrSMSNotConfigured = errors.New("SMS provider not configured")

// SMSProvider dispatches OTP SMS messages.
type SMSProvider interface {
	// Send sends a one-time passcode to the given E.164 phone number.
	// The callers discard the raw phone after this call returns — it is
	// never stored by any Echo service.
	Send(toE164, body string) error
}

// --- Stub (dev / test) ---

// StubSMSProvider logs OTP delivery and records the last sent code for test
// assertion.  Never use in production.
type StubSMSProvider struct {
	LastTo   string
	LastBody string
	LastOTP  string // extracted from body (last 6 digits)
}

func (s *StubSMSProvider) Send(toE164, body string) error {
	s.LastTo = toE164
	s.LastBody = body
	// Extract the OTP (last occurrence of 6 consecutive digits in the body).
	s.LastOTP = extractOTP(body)
	return nil
}

// extractOTP returns the first 6-digit sequence found in s (for test assertions).
func extractOTP(s string) string {
	for i := 0; i+6 <= len(s); i++ {
		candidate := s[i : i+6]
		allDigits := true
		for _, c := range candidate {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return candidate
		}
	}
	return ""
}

// --- Twilio (production scaffold) ---

// TwilioSMSProvider sends OTPs via the Twilio Programmable SMS API.
// Configure via environment variables:
//
//	TWILIO_ACCOUNT_SID   — Twilio account SID (starts with AC)
//	TWILIO_AUTH_TOKEN    — Twilio auth token
//	TWILIO_FROM          — E.164 sender number (e.g. +15005550006 for test)
type TwilioSMSProvider struct {
	accountSID string
	authToken  string
	from       string
	client     *http.Client
}

// NewTwilioSMSProvider reads credentials from env vars.
// Returns ErrSMSNotConfigured when any required variable is absent.
func NewTwilioSMSProvider() (*TwilioSMSProvider, error) {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM")
	if sid == "" || token == "" || from == "" {
		return nil, ErrSMSNotConfigured
	}
	return &TwilioSMSProvider{
		accountSID: sid,
		authToken:  token,
		from:       from,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send dispatches the OTP via Twilio's Messages API.
// The toE164 number is used only for this HTTP call and not stored.
func (t *TwilioSMSProvider) Send(toE164, body string) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	form := url.Values{}
	form.Set("To", toE164)
	form.Set("From", t.from)
	form.Set("Body", body)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("twilio build request: %w", err)
	}
	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("twilio send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio returned %d", resp.StatusCode)
	}
	return nil
}

// NewSMSProvider returns a production-ready provider if Twilio is configured,
// or a StubSMSProvider in development/test when credentials are absent.
// Callers should log which provider is active at startup.
func NewSMSProvider() (SMSProvider, bool) {
	twilio, err := NewTwilioSMSProvider()
	if err == nil {
		return twilio, true // production
	}
	return &StubSMSProvider{}, false // dev/test stub
}
