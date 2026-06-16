package database

import (
	"context"
	"errors"
	"time"
)

// OrgMemberRole matches EchoOrgRoleCredential roles (WO-310).
type OrgMemberRole string

const (
	OrgRoleOwner     OrgMemberRole = "owner"
	OrgRoleAdmin     OrgMemberRole = "admin"
	OrgRoleModerator OrgMemberRole = "moderator"
	OrgRoleMember    OrgMemberRole = "member"
)

func (r OrgMemberRole) CanReadComply() bool {
	switch r {
	case OrgRoleOwner, OrgRoleAdmin, OrgRoleModerator, OrgRoleMember:
		return true
	default:
		return false
	}
}

func (r OrgMemberRole) CanWriteComply() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// OrgMember links a DID to an org with a portal/comply role.
type OrgMember struct {
	OrgDID    string        `json:"orgDid"`
	MemberDID string        `json:"memberDid"`
	Role      OrgMemberRole `json:"role"`
	CreatedAt time.Time     `json:"createdAt"`
}

// DEFingerprintRecord tracks a message-level DE hash ref (no content).
type DEFingerprintRecord struct {
	OrgDID         string    `json:"orgDid"`
	MessageID      string    `json:"messageId"`
	FingerprintRef string    `json:"fingerprintRef"`
	RecordedAt     time.Time `json:"recordedAt"`
}

var ErrComplyOrgForbidden = errors.New("org membership required")

// ComplyRBACStore adds org membership and DE coverage tracking.
type ComplyRBACStore interface {
	UpsertOrgMember(ctx context.Context, m *OrgMember) error
	GetOrgMember(ctx context.Context, orgDID, memberDID string) (*OrgMember, error)
	RecordDEFingerprint(ctx context.Context, rec *DEFingerprintRecord) error
	CountDEFingerprints(ctx context.Context, orgDID string) (int, error)
	CountOrgScopedMessages(ctx context.Context, orgDID string) (int, error)
}
