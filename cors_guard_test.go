package main

import "testing"

func TestValidateCORSOrigins(t *testing.T) {
	if err := validateCORSOrigins([]string{"https://app.example.com", "http://localhost:3000"}); err != nil {
		t.Fatalf("explicit allowlist should be accepted, got: %v", err)
	}
	if err := validateCORSOrigins([]string{"https://app.example.com", "*"}); err == nil {
		t.Fatal("wildcard origin must be rejected when credentials are enabled")
	}
}
