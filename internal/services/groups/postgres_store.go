package groups

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists groups in PostgreSQL (migration 022_groups.sql).
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a Postgres-backed group store.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (p *PostgresStore) SaveGroup(ctx context.Context, g *Group) error {
	if g == nil || g.GroupID == "" {
		return errors.New("group id required")
	}
	tags, _ := json.Marshal(g.Tags)
	req, _ := json.Marshal(g.Requirements)
	settings, _ := json.Marshal(g.Settings)
	perms, _ := json.Marshal(g.Permissions)
	gov, _ := json.Marshal(g.Governance)
	stats, _ := json.Marshal(g.Stats)
	createdAt := g.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO echo_groups (
			group_id, owner_id, group_type, name, description, avatar, category,
			tags, rules, created_at, requirements, max_members, current_members,
			max_admins, max_moderators, settings, permissions, governance, stats,
			creation_tx_hash, last_update_tx_hash, snapshot_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22
		)
		ON CONFLICT (group_id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			group_type = EXCLUDED.group_type,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			avatar = EXCLUDED.avatar,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			rules = EXCLUDED.rules,
			requirements = EXCLUDED.requirements,
			max_members = EXCLUDED.max_members,
			current_members = EXCLUDED.current_members,
			max_admins = EXCLUDED.max_admins,
			max_moderators = EXCLUDED.max_moderators,
			settings = EXCLUDED.settings,
			permissions = EXCLUDED.permissions,
			governance = EXCLUDED.governance,
			stats = EXCLUDED.stats,
			creation_tx_hash = EXCLUDED.creation_tx_hash,
			last_update_tx_hash = EXCLUDED.last_update_tx_hash,
			snapshot_id = EXCLUDED.snapshot_id`,
		g.GroupID, g.OwnerID, string(g.Type), g.Name, g.Description, g.Avatar, string(g.Category),
		tags, g.Rules, createdAt, req, g.MaxMembers, g.CurrentMembers,
		g.MaxAdmins, g.MaxModerators, settings, perms, gov, stats,
		g.CreationTxHash, g.LastUpdateTxHash, g.SnapshotID,
	)
	if err != nil {
		return fmt.Errorf("save group: %w", err)
	}
	return nil
}

func (p *PostgresStore) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT group_id, owner_id, group_type, name, description, avatar, category,
		       tags, rules, created_at, requirements, max_members, current_members,
		       max_admins, max_moderators, settings, permissions, governance, stats,
		       creation_tx_hash, last_update_tx_hash, snapshot_id
		FROM echo_groups WHERE group_id = $1`, groupID)

	g, err := scanGroup(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGroupNotFound
	}
	return g, err
}

func (p *PostgresStore) GroupExists(ctx context.Context, groupID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM echo_groups WHERE group_id = $1)`, groupID).Scan(&exists)
	return exists, err
}

func (p *PostgresStore) SaveMember(ctx context.Context, m *GroupMember) error {
	if m == nil || m.GroupID == "" || m.MemberID == "" {
		return errors.New("group and member id required")
	}
	perms, _ := json.Marshal(m.Permissions)
	badges, _ := json.Marshal(m.Badges)
	joinedAt := m.JoinedAt
	if joinedAt.IsZero() {
		joinedAt = time.Now()
	}
	lastActive := m.LastActiveAt
	if lastActive.IsZero() {
		lastActive = joinedAt
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO echo_group_members (
			group_id, member_id, display_name, avatar, persona_id, role, permissions,
			trust_score, trust_level, badges, verified_at, joined_at, last_active_at,
			message_count, warning_count, is_muted, muted_until, is_banned,
			notification_level, nickname, show_trust_score
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21
		)`,
		m.GroupID, m.MemberID, m.DisplayName, m.Avatar, m.PersonaID, string(m.Role), perms,
		m.TrustScore, string(m.TrustLevel), badges, nullTime(m.VerifiedAt), joinedAt, lastActive,
		m.MessageCount, m.WarningCount, m.IsMuted, m.MutedUntil, m.IsBanned,
		string(m.NotificationLevel), m.Nickname, m.ShowTrustScore,
	)
	if err != nil {
		return fmt.Errorf("save member: %w", err)
	}
	return nil
}

func (p *PostgresStore) UpdateMember(ctx context.Context, m *GroupMember) error {
	if m == nil || m.GroupID == "" || m.MemberID == "" {
		return errors.New("group and member id required")
	}
	perms, _ := json.Marshal(m.Permissions)
	badges, _ := json.Marshal(m.Badges)
	tag, err := p.pool.Exec(ctx, `
		UPDATE echo_group_members SET
			display_name = $3, avatar = $4, persona_id = $5, role = $6, permissions = $7,
			trust_score = $8, trust_level = $9, badges = $10, verified_at = $11,
			last_active_at = $12, message_count = $13, warning_count = $14,
			is_muted = $15, muted_until = $16, is_banned = $17,
			notification_level = $18, nickname = $19, show_trust_score = $20
		WHERE group_id = $1 AND member_id = $2`,
		m.GroupID, m.MemberID, m.DisplayName, m.Avatar, m.PersonaID, string(m.Role), perms,
		m.TrustScore, string(m.TrustLevel), badges, nullTime(m.VerifiedAt), m.LastActiveAt,
		m.MessageCount, m.WarningCount, m.IsMuted, m.MutedUntil, m.IsBanned,
		string(m.NotificationLevel), m.Nickname, m.ShowTrustScore,
	)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (p *PostgresStore) GetMember(ctx context.Context, groupID, memberID string) (*GroupMember, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT group_id, member_id, display_name, avatar, persona_id, role, permissions,
		       trust_score, trust_level, badges, verified_at, joined_at, last_active_at,
		       message_count, warning_count, is_muted, muted_until, is_banned,
		       notification_level, nickname, show_trust_score
		FROM echo_group_members WHERE group_id = $1 AND member_id = $2`, groupID, memberID)
	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMemberNotFound
	}
	return m, err
}

func (p *PostgresStore) DeleteMember(ctx context.Context, groupID, memberID string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM echo_group_members WHERE group_id = $1 AND member_id = $2`, groupID, memberID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (p *PostgresStore) ListMembers(ctx context.Context, groupID string) ([]*GroupMember, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT group_id, member_id, display_name, avatar, persona_id, role, permissions,
		       trust_score, trust_level, badges, verified_at, joined_at, last_active_at,
		       message_count, warning_count, is_muted, muted_until, is_banned,
		       notification_level, nickname, show_trust_score
		FROM echo_group_members WHERE group_id = $1 ORDER BY joined_at`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*GroupMember
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		exists, err := p.GroupExists(ctx, groupID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrGroupNotFound
		}
	}
	return out, rows.Err()
}

func (p *PostgresStore) ListGroupIDsForMember(ctx context.Context, memberID string) ([]string, error) {
	rows, err := p.pool.Query(ctx, `SELECT group_id FROM echo_group_members WHERE member_id = $1`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanGroup(row scannable) (*Group, error) {
	var g Group
	var groupType, category string
	var tags, req, settings, perms, gov, stats []byte
	err := row.Scan(
		&g.GroupID, &g.OwnerID, &groupType, &g.Name, &g.Description, &g.Avatar, &category,
		&tags, &g.Rules, &g.CreatedAt, &req, &g.MaxMembers, &g.CurrentMembers,
		&g.MaxAdmins, &g.MaxModerators, &settings, &perms, &gov, &stats,
		&g.CreationTxHash, &g.LastUpdateTxHash, &g.SnapshotID,
	)
	if err != nil {
		return nil, err
	}
	g.Type = GroupType(groupType)
	g.Category = GroupCategory(category)
	_ = json.Unmarshal(tags, &g.Tags)
	_ = json.Unmarshal(req, &g.Requirements)
	_ = json.Unmarshal(settings, &g.Settings)
	_ = json.Unmarshal(perms, &g.Permissions)
	_ = json.Unmarshal(gov, &g.Governance)
	_ = json.Unmarshal(stats, &g.Stats)
	return &g, nil
}

func scanMember(row scannable) (*GroupMember, error) {
	var m GroupMember
	var role, trustLevel, notif string
	var perms, badges []byte
	var verifiedAt *time.Time
	err := row.Scan(
		&m.GroupID, &m.MemberID, &m.DisplayName, &m.Avatar, &m.PersonaID, &role, &perms,
		&m.TrustScore, &trustLevel, &badges, &verifiedAt, &m.JoinedAt, &m.LastActiveAt,
		&m.MessageCount, &m.WarningCount, &m.IsMuted, &m.MutedUntil, &m.IsBanned,
		&notif, &m.Nickname, &m.ShowTrustScore,
	)
	if err != nil {
		return nil, err
	}
	m.Role = GroupRole(role)
	m.TrustLevel = TrustLevel(trustLevel)
	m.NotificationLevel = NotificationLevel(notif)
	_ = json.Unmarshal(perms, &m.Permissions)
	_ = json.Unmarshal(badges, &m.Badges)
	if verifiedAt != nil {
		m.VerifiedAt = *verifiedAt
	}
	return &m, nil
}

func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
