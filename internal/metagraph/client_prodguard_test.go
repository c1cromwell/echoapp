package metagraph

import (
	"context"
	"strings"
	"testing"
)

// TestSubmitDataL1_RefusesUnsignedInProduction verifies the fail-closed guard:
// without an IdentitySigner, a Data L1 submission is rejected in production but
// allowed (attempted) outside it.
func TestSubmitDataL1_RefusesUnsignedInProduction(t *testing.T) {
	client := NewMetagraphClient(MetagraphConfig{DataL1URL: "http://metagraph.invalid"})

	t.Run("production rejects unsigned", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		_, err := client.SubmitDataL1(context.Background(), map[string]string{"k": "v"})
		if err == nil {
			t.Fatal("expected error for unsigned submission in production")
		}
		if !strings.Contains(err.Error(), "refusing unsigned") {
			t.Fatalf("expected fail-closed error, got: %v", err)
		}
	})

	t.Run("non-production does not gate on signing", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "development")
		// No signer + unreachable URL: the call proceeds past the guard and fails
		// at transport, NOT with the fail-closed message.
		_, err := client.SubmitDataL1(context.Background(), map[string]string{"k": "v"})
		if err != nil && strings.Contains(err.Error(), "refusing unsigned") {
			t.Fatalf("unexpected fail-closed error outside production: %v", err)
		}
	})
}
