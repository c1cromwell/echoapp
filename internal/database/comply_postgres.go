package database

import (
	"context"
	"fmt"
	"time"
)

// --- PostgresDB ComplyStore (WO-250) ---

func (p *PostgresDB) CreateRetentionPolicy(ctx context.Context, pol *RetentionPolicy) error {
	if pol == nil || pol.ID == "" {
		return fmt.Errorf("invalid retention policy")
	}
	createdAt := pol.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	effectiveAt := pol.EffectiveAt
	if effectiveAt.IsZero() {
		effectiveAt = createdAt
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO comply_retention_policies
		 (id, org_did, policy_type, conversation_id, scope_label, effective_at, expires_at, data_l1_ref, active, created_by_did, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		pol.ID, pol.OrgDID, string(pol.PolicyType), nullIfEmpty(pol.ConversationID), nullIfEmpty(pol.ScopeLabel),
		effectiveAt, pol.ExpiresAt, nullIfEmpty(pol.DataL1Ref), pol.Active, pol.CreatedByDID, createdAt)
	if err != nil {
		return fmt.Errorf("create retention policy: %w", err)
	}
	return nil
}

func (p *PostgresDB) ListRetentionPolicies(ctx context.Context, orgDID string, activeOnly bool) ([]*RetentionPolicy, error) {
	q := `SELECT id, org_did, policy_type, COALESCE(conversation_id,''), COALESCE(scope_label,''),
		effective_at, expires_at, COALESCE(data_l1_ref,''), active, created_by_did, created_at
		FROM comply_retention_policies WHERE org_did = $1`
	if activeOnly {
		q += ` AND active = TRUE`
	}
	rows, err := p.pool.Query(ctx, q, orgDID)
	if err != nil {
		return nil, fmt.Errorf("list retention policies: %w", err)
	}
	defer rows.Close()
	var out []*RetentionPolicy
	for rows.Next() {
		pol, err := scanRetentionPolicy(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, pol)
	}
	return out, rows.Err()
}

func (p *PostgresDB) GetRetentionPolicy(ctx context.Context, policyID string) (*RetentionPolicy, error) {
	row := p.pool.QueryRow(ctx,
		`SELECT id, org_did, policy_type, COALESCE(conversation_id,''), COALESCE(scope_label,''),
			effective_at, expires_at, COALESCE(data_l1_ref,''), active, created_by_did, created_at
		 FROM comply_retention_policies WHERE id = $1`, policyID)
	return scanRetentionPolicy(row.Scan)
}

func (p *PostgresDB) BindConversationPolicy(ctx context.Context, b *ConversationPolicyBinding) error {
	if b == nil {
		return fmt.Errorf("invalid binding")
	}
	boundAt := b.BoundAt
	if boundAt.IsZero() {
		boundAt = time.Now()
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO comply_conversation_bindings (conversation_id, policy_id, org_did, bound_at)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (conversation_id) DO UPDATE SET policy_id = EXCLUDED.policy_id, org_did = EXCLUDED.org_did, bound_at = EXCLUDED.bound_at`,
		b.ConversationID, b.PolicyID, b.OrgDID, boundAt)
	if err != nil {
		return fmt.Errorf("bind conversation policy: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetConversationBinding(ctx context.Context, conversationID string) (*ConversationPolicyBinding, error) {
	var b ConversationPolicyBinding
	err := p.pool.QueryRow(ctx,
		`SELECT conversation_id, policy_id, org_did, bound_at FROM comply_conversation_bindings WHERE conversation_id = $1`,
		conversationID).Scan(&b.ConversationID, &b.PolicyID, &b.OrgDID, &b.BoundAt)
	if err != nil {
		return nil, ErrComplyPolicyNotFound
	}
	return &b, nil
}

func (p *PostgresDB) UnbindConversation(ctx context.Context, conversationID string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM comply_conversation_bindings WHERE conversation_id = $1`, conversationID)
	if err != nil {
		return fmt.Errorf("unbind conversation: %w", err)
	}
	return nil
}

func (p *PostgresDB) UpsertOrgProfile(ctx context.Context, profile *ComplyOrgProfile) error {
	if profile == nil || profile.OrgDID == "" {
		return fmt.Errorf("invalid org profile")
	}
	tier := profile.Tier
	if tier == "" {
		tier = "starter"
	}
	seats := profile.Seats
	if seats <= 0 {
		seats = 10
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO comply_org_profiles (org_did, tier, seats, updated_at) VALUES ($1,$2,$3,NOW())
		 ON CONFLICT (org_did) DO UPDATE SET tier = EXCLUDED.tier, seats = EXCLUDED.seats, updated_at = NOW()`,
		profile.OrgDID, tier, seats)
	if err != nil {
		return fmt.Errorf("upsert org profile: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetOrgProfile(ctx context.Context, orgDID string) (*ComplyOrgProfile, error) {
	var profile ComplyOrgProfile
	err := p.pool.QueryRow(ctx,
		`SELECT org_did, tier, seats, updated_at FROM comply_org_profiles WHERE org_did = $1`, orgDID).
		Scan(&profile.OrgDID, &profile.Tier, &profile.Seats, &profile.UpdatedAt)
	if err != nil {
		return nil, ErrComplyPolicyNotFound
	}
	return &profile, nil
}

func (p *PostgresDB) CountActivePolicies(ctx context.Context, orgDID string, policyType RetentionPolicyType) (int, error) {
	var n int
	var err error
	if policyType == "" {
		err = p.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM comply_retention_policies WHERE org_did = $1 AND active = TRUE`, orgDID).Scan(&n)
	} else {
		err = p.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM comply_retention_policies WHERE org_did = $1 AND active = TRUE AND policy_type = $2`,
			orgDID, string(policyType)).Scan(&n)
	}
	if err != nil {
		return 0, fmt.Errorf("count policies: %w", err)
	}
	return n, nil
}

func scanRetentionPolicy(scan func(dest ...any) error) (*RetentionPolicy, error) {
	var pol RetentionPolicy
	var policyType string
	var expiresAt *time.Time
	err := scan(
		&pol.ID, &pol.OrgDID, &policyType, &pol.ConversationID, &pol.ScopeLabel,
		&pol.EffectiveAt, &expiresAt, &pol.DataL1Ref, &pol.Active, &pol.CreatedByDID, &pol.CreatedAt,
	)
	if err != nil {
		return nil, ErrComplyPolicyNotFound
	}
	pol.PolicyType = RetentionPolicyType(policyType)
	pol.ExpiresAt = expiresAt
	return &pol, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
