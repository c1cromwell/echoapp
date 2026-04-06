package auth

import (
	"context"
	"testing"
	"time"
)

func TestCreateAttestation(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	tests := []struct {
		name        string
		attesterDID string
		subjectDID  string
		type_       AttestationType
		confidence  int
		shouldErr   bool
	}{
		{
			name:        "valid vouch",
			attesterDID: "did:echo:alice",
			subjectDID:  "did:echo:bob",
			type_:       AttestationVouch,
			confidence:  8,
			shouldErr:   false,
		},
		{
			name:        "self vouch",
			attesterDID: "did:echo:alice",
			subjectDID:  "did:echo:alice",
			type_:       AttestationVouch,
			confidence:  5,
			shouldErr:   true,
		},
		{
			name:        "invalid confidence",
			attesterDID: "did:echo:alice",
			subjectDID:  "did:echo:bob",
			type_:       AttestationVouch,
			confidence:  11,
			shouldErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attestation, reward, err := service.CreateAttestation(
				ctx,
				tt.attesterDID,
				tt.subjectDID,
				tt.type_,
				tt.confidence,
				ContextPersonal,
				"Test statement",
			)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if attestation == nil {
				t.Error("attestation is nil")
			}

			if attestation.Type != tt.type_ {
				t.Errorf("type = %v, want %v", attestation.Type, tt.type_)
			}

			if reward == nil {
				t.Error("reward is nil")
			}
		})
	}
}

func TestRevokeAttestation(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	attestation, _, _ := service.CreateAttestation(
		ctx,
		"did:echo:alice",
		"did:echo:bob",
		AttestationVouch,
		8,
		ContextPersonal,
		"Test",
	)

	err := service.RevokeAttestation(ctx, attestation.AttestationID, "did:echo:alice")
	if err != nil {
		t.Errorf("RevokeAttestation failed: %v", err)
	}
}

func TestGetAttestationStats(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	attestations := []*WebOfTrustAttestation{
		{
			Type:       AttestationVouch,
			Confidence: 8,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 6, 0),
			Revoked:    false,
		},
		{
			Type:       AttestationEndorse,
			Confidence: 9,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 6, 0),
			Revoked:    false,
		},
		{
			Type:       AttestationVerify,
			Confidence: 10,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 6, 0),
			Revoked:    false,
		},
		{
			Type:       AttestationVouch,
			Confidence: 5,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 6, 0),
			Revoked:    true,
		},
	}

	stats := service.GetAttestationStats(ctx, attestations)

	if stats.TotalAttestation != 4 {
		t.Errorf("total = %d, want 4", stats.TotalAttestation)
	}

	if stats.ActiveAttestations != 3 {
		t.Errorf("active = %d, want 3", stats.ActiveAttestations)
	}

	if stats.RevokedCount != 1 {
		t.Errorf("revoked = %d, want 1", stats.RevokedCount)
	}
}

func TestCalculateWebOfTrustBoost(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	tests := []struct {
		name             string
		attestations     []*WebOfTrustAttestation
		expectedMinBoost int
	}{
		{
			name:             "no attestations",
			attestations:     []*WebOfTrustAttestation{},
			expectedMinBoost: 0,
		},
		{
			name: "mixed attestations",
			attestations: []*WebOfTrustAttestation{
				{Type: AttestationVouch, Confidence: 10, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0), Revoked: false},
				{Type: AttestationEndorse, Confidence: 10, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0), Revoked: false},
				{Type: AttestationVerify, Confidence: 10, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0), Revoked: false},
			},
			expectedMinBoost: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boost, err := service.CalculateWebOfTrustBoost(ctx, tt.attestations)
			if err != nil {
				t.Errorf("CalculateWebOfTrustBoost failed: %v", err)
			}

			if boost < tt.expectedMinBoost {
				t.Errorf("boost = %d, want >= %d", boost, tt.expectedMinBoost)
			}

			if boost > 15 {
				t.Errorf("boost = %d, exceeds maximum of 15", boost)
			}
		})
	}
}

func TestDetectCircularVouching(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	hasCircle, err := service.DetectCircularVouching(ctx, "did:echo:alice", "did:echo:bob", 5)

	if err != nil {
		t.Errorf("DetectCircularVouching failed: %v", err)
	}

	if hasCircle {
		t.Errorf("unexpected circle detected")
	}
}

func TestTransitiveTrustCalculator(t *testing.T) {
	calculator := NewTransitiveTrustCalculator()
	ctx := context.Background()

	trust, err := calculator.CalculateTransitiveTrust(ctx, "did:echo:alice", "did:echo:alice", 3)

	if err != nil {
		t.Errorf("CalculateTransitiveTrust failed: %v", err)
	}

	if trust != 1.0 {
		t.Errorf("trust in self = %v, want 1.0", trust)
	}
}

func TestAttestationExpiry(t *testing.T) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	attestation, _, _ := service.CreateAttestation(
		ctx,
		"did:echo:alice",
		"did:echo:bob",
		AttestationVouch,
		8,
		ContextPersonal,
		"Test",
	)

	if attestation.ExpiresAt.Before(time.Now()) {
		t.Error("attestation already expired")
	}
}

// Benchmark attestation creation
func BenchmarkCreateAttestation(b *testing.B) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateAttestation(
			ctx,
			"did:echo:alice",
			"did:echo:bob",
			AttestationVouch,
			8,
			ContextPersonal,
			"Benchmark",
		)
	}
}

// Benchmark boost calculation
func BenchmarkCalculateWebOfTrustBoost(b *testing.B) {
	service := NewWebOfTrustService(&WebOfTrustConfig{})
	ctx := context.Background()

	attestations := []*WebOfTrustAttestation{
		{Type: AttestationVouch, Confidence: 8, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0)},
		{Type: AttestationEndorse, Confidence: 9, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0)},
		{Type: AttestationVerify, Confidence: 10, CreatedAt: time.Now(), ExpiresAt: time.Now().AddDate(0, 1, 0)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateWebOfTrustBoost(ctx, attestations)
	}
}
