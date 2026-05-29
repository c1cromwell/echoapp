package onboarding

import (
	"context"
	"fmt"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// TrustRegistryStore persists trusted issuers (WO-118).
type TrustRegistryStore interface {
	ListIssuers(ctx context.Context) ([]*TrustedIssuer, error)
	SaveIssuer(ctx context.Context, issuer *TrustedIssuer, suspendedAt *time.Time, revokedAt *time.Time) error
}

// PostgresTrustRegistryStore implements TrustRegistryStore via Postgres.
type PostgresTrustRegistryStore struct {
	db *database.PostgresDB
}

func NewPostgresTrustRegistryStore(db *database.PostgresDB) *PostgresTrustRegistryStore {
	return &PostgresTrustRegistryStore{db: db}
}

func (s *PostgresTrustRegistryStore) ListIssuers(ctx context.Context) ([]*TrustedIssuer, error) {
	rows, err := s.db.ListTrustedIssuers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*TrustedIssuer, 0, len(rows))
	for i := range rows {
		out = append(out, rowToTrustedIssuer(&rows[i]))
	}
	return out, nil
}

func (s *PostgresTrustRegistryStore) SaveIssuer(ctx context.Context, issuer *TrustedIssuer, suspendedAt, revokedAt *time.Time) error {
	if issuer == nil {
		return fmt.Errorf("issuer required")
	}
	row := trustedIssuerToRow(issuer, suspendedAt, revokedAt)
	return s.db.UpsertTrustedIssuer(ctx, row)
}

func rowToTrustedIssuer(row *database.TrustedIssuerRow) *TrustedIssuer {
	credTypes := make([]CredentialType, len(row.CredentialTypes))
	for i, ct := range row.CredentialTypes {
		credTypes[i] = CredentialType(ct)
	}
	issuer := &TrustedIssuer{
		ID:                          row.IssuerID,
		Name:                        row.Name,
		DID:                         row.DID,
		Type:                        IssuerType(row.IssuerType),
		Jurisdiction:                Jurisdiction(row.Jurisdiction),
		TrustLevel:                  TrustLevel(row.TrustLevel),
		Status:                      row.Status,
		CredentialTypes:             credTypes,
		VerificationPublicKeyBase64: row.VerificationPublicKeyB64,
		PublicKeyURL:                row.PublicKeyURL,
		RiskScore:                   row.RiskScore,
		OnboardingWeight:            row.OnboardingWeight,
		ActivationThreshold:         row.ActivationThreshold,
		ContactEmail:                row.ContactEmail,
		DocumentationURL:            row.DocumentationURL,
		VerifiedTimestamp:           row.VerifiedAt,
		LastVerificationDate:        row.LastVerifiedAt,
	}
	if row.EstablishedAt != nil {
		issuer.EstablishedDate = *row.EstablishedAt
	}
	return issuer
}

func trustedIssuerToRow(issuer *TrustedIssuer, suspendedAt, revokedAt *time.Time) database.TrustedIssuerRow {
	credTypes := make([]string, len(issuer.CredentialTypes))
	for i, ct := range issuer.CredentialTypes {
		credTypes[i] = string(ct)
	}
	var established *time.Time
	if !issuer.EstablishedDate.IsZero() {
		t := issuer.EstablishedDate
		established = &t
	}
	status := issuer.Status
	if revokedAt != nil {
		status = "revoked"
	} else if suspendedAt != nil {
		status = "suspended"
	}
	return database.TrustedIssuerRow{
		IssuerID:                 issuer.ID,
		Name:                     issuer.Name,
		DID:                      issuer.DID,
		IssuerType:               string(issuer.Type),
		Jurisdiction:             string(issuer.Jurisdiction),
		TrustLevel:               string(issuer.TrustLevel),
		Status:                   status,
		CredentialTypes:          credTypes,
		VerificationPublicKeyB64: issuer.VerificationPublicKeyBase64,
		PublicKeyURL:             issuer.PublicKeyURL,
		RiskScore:                issuer.RiskScore,
		OnboardingWeight:         issuer.OnboardingWeight,
		ActivationThreshold:      issuer.ActivationThreshold,
		ContactEmail:             issuer.ContactEmail,
		DocumentationURL:         issuer.DocumentationURL,
		EstablishedAt:            established,
		VerifiedAt:               issuer.VerifiedTimestamp,
		LastVerifiedAt:           issuer.LastVerificationDate,
		SuspendedAt:              suspendedAt,
		RevokedAt:                revokedAt,
	}
}
