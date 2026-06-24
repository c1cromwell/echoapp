package bots

import "time"

// Permission names bots declare in manifests and users grant at install (WO-11 / WO-63).
type Permission string

const (
	PermSendMessage    Permission = "send_message"
	PermReadMessages   Permission = "read_messages"
	PermRequestPayment Permission = "request_payment"
	PermUploadFile     Permission = "upload_file"
	PermReadChainState Permission = "read_chain_state"
)

// Manifest describes a marketplace bot (content-blind metadata only).
type Manifest struct {
	BotDID              string       `json:"bot_did"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	Version             string       `json:"version"`
	RequiredPermissions []Permission `json:"required_permissions"`
	TrustScore          int          `json:"trust_score"`
}

// Installation records a user's granted permissions for an installed bot.
type Installation struct {
	BotDID             string       `json:"bot_did"`
	UserDID            string       `json:"user_did"`
	GrantedPermissions []Permission `json:"granted_permissions"`
	InstalledAt        time.Time    `json:"installed_at"`
	Active             bool         `json:"active"`
	LastActivityAt     *time.Time   `json:"last_activity_at,omitempty"`
}
