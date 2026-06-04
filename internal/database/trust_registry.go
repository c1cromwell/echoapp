package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// TrustedIssuerRow is the Postgres representation of a trusted VC issuer (WO-118).
type TrustedIssuerRow struct {
	IssuerID                 string
	Name                     string
	DID                      string
	IssuerType               string
	Jurisdiction             string
	TrustLevel               string
	Status                   string
	CredentialTypes          []string
	VerificationPublicKeyB64 string
	PublicKeyURL             string
	RiskScore                int
	OnboardingWeight         int
	ActivationThreshold      float64
	ContactEmail             string
	DocumentationURL         string
	EstablishedAt            *time.Time
	VerifiedAt               time.Time
	LastVerifiedAt           time.Time
	SuspendedAt              *time.Time
	RevokedAt                *time.Time
}

func (p *PostgresDB) ListTrustedIssuers(ctx context.Context) ([]TrustedIssuerRow, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT issuer_id, name, did, issuer_type, jurisdiction, trust_level, status,
		       credential_types, COALESCE(verification_public_key_b64, ''),
		       COALESCE(public_key_url, ''), risk_score, onboarding_weight,
		       activation_threshold, COALESCE(contact_email, ''), COALESCE(documentation_url, ''),
		       established_at, verified_at, last_verified_at, suspended_at, revoked_at
		FROM trusted_issuers
		ORDER BY issuer_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TrustedIssuerRow
	for rows.Next() {
		var row TrustedIssuerRow
		if err := rows.Scan(
			&row.IssuerID, &row.Name, &row.DID, &row.IssuerType, &row.Jurisdiction,
			&row.TrustLevel, &row.Status, &row.CredentialTypes, &row.VerificationPublicKeyB64,
			&row.PublicKeyURL, &row.RiskScore, &row.OnboardingWeight, &row.ActivationThreshold,
			&row.ContactEmail, &row.DocumentationURL, &row.EstablishedAt, &row.VerifiedAt,
			&row.LastVerifiedAt, &row.SuspendedAt, &row.RevokedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (p *PostgresDB) UpsertTrustedIssuer(ctx context.Context, row TrustedIssuerRow) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO trusted_issuers (
			issuer_id, name, did, issuer_type, jurisdiction, trust_level, status,
			credential_types, verification_public_key_b64, public_key_url,
			risk_score, onboarding_weight, activation_threshold,
			contact_email, documentation_url, established_at,
			verified_at, last_verified_at, suspended_at, revoked_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, NULLIF($9, ''), NULLIF($10, ''),
			$11, $12, $13,
			NULLIF($14, ''), NULLIF($15, ''), $16,
			$17, $18, $19, $20, NOW()
		)
		ON CONFLICT (issuer_id) DO UPDATE SET
			name = EXCLUDED.name,
			did = EXCLUDED.did,
			issuer_type = EXCLUDED.issuer_type,
			jurisdiction = EXCLUDED.jurisdiction,
			trust_level = EXCLUDED.trust_level,
			status = EXCLUDED.status,
			credential_types = EXCLUDED.credential_types,
			verification_public_key_b64 = EXCLUDED.verification_public_key_b64,
			public_key_url = EXCLUDED.public_key_url,
			risk_score = EXCLUDED.risk_score,
			onboarding_weight = EXCLUDED.onboarding_weight,
			activation_threshold = EXCLUDED.activation_threshold,
			contact_email = EXCLUDED.contact_email,
			documentation_url = EXCLUDED.documentation_url,
			established_at = EXCLUDED.established_at,
			verified_at = EXCLUDED.verified_at,
			last_verified_at = EXCLUDED.last_verified_at,
			suspended_at = EXCLUDED.suspended_at,
			revoked_at = EXCLUDED.revoked_at,
			updated_at = NOW()`,
		row.IssuerID, row.Name, row.DID, row.IssuerType, row.Jurisdiction, row.TrustLevel, row.Status,
		row.CredentialTypes, row.VerificationPublicKeyB64, row.PublicKeyURL,
		row.RiskScore, row.OnboardingWeight, row.ActivationThreshold,
		row.ContactEmail, row.DocumentationURL, row.EstablishedAt,
		row.VerifiedAt, row.LastVerifiedAt, row.SuspendedAt, row.RevokedAt,
	)
	return err
}

func (p *PostgresDB) GetTrustedIssuerByDID(ctx context.Context, did string) (*TrustedIssuerRow, error) {
	row := &TrustedIssuerRow{}
	err := p.pool.QueryRow(ctx, `
		SELECT issuer_id, name, did, issuer_type, jurisdiction, trust_level, status,
		       credential_types, COALESCE(verification_public_key_b64, ''),
		       COALESCE(public_key_url, ''), risk_score, onboarding_weight,
		       activation_threshold, COALESCE(contact_email, ''), COALESCE(documentation_url, ''),
		       established_at, verified_at, last_verified_at, suspended_at, revoked_at
		FROM trusted_issuers WHERE did = $1`, did).Scan(
		&row.IssuerID, &row.Name, &row.DID, &row.IssuerType, &row.Jurisdiction,
		&row.TrustLevel, &row.Status, &row.CredentialTypes, &row.VerificationPublicKeyB64,
		&row.PublicKeyURL, &row.RiskScore, &row.OnboardingWeight, &row.ActivationThreshold,
		&row.ContactEmail, &row.DocumentationURL, &row.EstablishedAt, &row.VerifiedAt,
		&row.LastVerifiedAt, &row.SuspendedAt, &row.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row, nil
}
