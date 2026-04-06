package hidden_folders

import (
	"math/rand"
	"time"
)

// NotificationMode defines how notifications are handled for hidden folders
type NotificationMode string

const (
	NotificationSuppressed   NotificationMode = "suppressed"    // No notifications (default)
	NotificationRedacted     NotificationMode = "redacted"      // "New message" with no details
	NotificationUnlockedOnly NotificationMode = "unlocked_only" // Show only when folder unlocked
)

// SecurityTier defines the security level for a hidden folder
type SecurityTier string

const (
	SecurityTierStandard SecurityTier = "standard" // 5 min auto-lock, no wipe
	SecurityTierElevated SecurityTier = "elevated" // 30 sec auto-lock, wipe at 15
	SecurityTierMaximum  SecurityTier = "maximum"  // Immediate auto-lock, wipe at 10, auto-lock on screenshot
)

// HiddenFolder represents a biometrically-protected folder for sensitive conversations
type HiddenFolder struct {
	ID                       string           `json:"id"`
	Name                     string           `json:"name"`
	CreatedAt                time.Time        `json:"created_at"`
	LastAccessedAt           time.Time        `json:"last_accessed_at"`
	AutoLockTimeout          time.Duration    `json:"auto_lock_timeout"` // seconds
	NotificationMode         NotificationMode `json:"notification_mode"`
	ScreenshotProtection     bool             `json:"screenshot_protection"`
	WipeLockoutThreshold     int              `json:"wipe_lockout_threshold"` // default: 10
	IsDecoy                  bool             `json:"is_decoy"`               // true = duress decoy folder
	SecurityTier             SecurityTier     `json:"security_tier"`
	MediaQuotaBytes          int64            `json:"media_quota_bytes"` // default: 2GB
	MediaUsedBytes           int64            `json:"media_used_bytes"`
	EncryptedConversationIDs []string         `json:"encrypted_conversation_ids"` // Note: encrypted in production
	UserID                   string           `json:"user_id"`                    // Device-local only
}

// HiddenMessage represents a message stored in a hidden folder (Layer 2 encrypted)
type HiddenMessage struct {
	ID                string     `json:"id"` // Same as original message ID
	FolderID          string     `json:"folder_id"`
	ConversationID    string     `json:"conversation_id"`
	Timestamp         time.Time  `json:"timestamp"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"` // Disappearing message expiry
	EncryptedContent  []byte     `json:"encrypted_content,omitempty"`
	EncryptedMetadata []byte     `json:"encrypted_metadata,omitempty"`
	ContentType       string     `json:"content_type"` // "text", "image", etc.
	IsRead            bool       `json:"is_read"`
	DeliveryStatus    string     `json:"delivery_status"`
}

// HiddenFolderRoute maps conversations to their hidden folders
type HiddenFolderRoute struct {
	FolderID       string `json:"folder_id"`
	ConversationID string `json:"conversation_id"`
}

// SecurityEvent represents security-related events (lockouts, wipes, biometric attempts)
type SecurityEvent struct {
	ID        string    `json:"id"`
	FolderID  string    `json:"folder_id"`
	EventType string    `json:"event_type"` // "biometric_attempt", "lockout", "wipe", "screenshot_attempt"
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// BiometricLockoutTracker tracks failed biometric attempts
type BiometricLockoutTracker struct {
	FolderID         string     `json:"folder_id"`
	FailedAttempts   int        `json:"failed_attempts"`
	LastFailureTime  *time.Time `json:"last_failure_time,omitempty"`
	IsLockedOut      bool       `json:"is_locked_out"`
	LockoutExpiresAt *time.Time `json:"lockout_expires_at,omitempty"`
	WipeScheduledAt  *time.Time `json:"wipe_scheduled_at,omitempty"` // If 10th attempt reached
}

// NewHiddenFolder creates a new hidden folder with default values
func NewHiddenFolder(name string, tier SecurityTier) *HiddenFolder {
	now := time.Now()

	folder := &HiddenFolder{
		ID:                   generateFolderID(),
		Name:                 name,
		CreatedAt:            now,
		LastAccessedAt:       now,
		NotificationMode:     NotificationSuppressed,
		ScreenshotProtection: true,
		IsDecoy:              false,
		SecurityTier:         tier,
		MediaQuotaBytes:      2_147_483_648, // 2GB
		MediaUsedBytes:       0,
	}

	// Apply tier defaults
	switch tier {
	case SecurityTierStandard:
		folder.AutoLockTimeout = 5 * time.Minute
		folder.WipeLockoutThreshold = 0 // Disabled
	case SecurityTierElevated:
		folder.AutoLockTimeout = 30 * time.Second
		folder.WipeLockoutThreshold = 15
	case SecurityTierMaximum:
		folder.AutoLockTimeout = 0 // Immediate
		folder.WipeLockoutThreshold = 10
	}

	return folder
}

// GetAutoLockDuration returns the auto-lock timeout in seconds
func (hf *HiddenFolder) GetAutoLockDuration() int {
	return int(hf.AutoLockTimeout.Seconds())
}

// CanAccessWithoutBiometric returns false for all hidden folders (always requires biometric)
func (hf *HiddenFolder) CanAccessWithoutBiometric() bool {
	return false
}

// IsAtMediaQuota returns true if media usage is at or above the quota
func (hf *HiddenFolder) IsAtMediaQuota() bool {
	return hf.MediaUsedBytes >= hf.MediaQuotaBytes
}

// MediaQuotaPercentage returns the media usage as a percentage of quota
func (hf *HiddenFolder) MediaQuotaPercentage() float64 {
	if hf.MediaQuotaBytes == 0 {
		return 0
	}
	return float64(hf.MediaUsedBytes) / float64(hf.MediaQuotaBytes) * 100
}

// HasConversation returns true if the folder contains a specific conversation
func (hf *HiddenFolder) HasConversation(conversationID string) bool {
	for _, id := range hf.EncryptedConversationIDs {
		if id == conversationID {
			return true
		}
	}
	return false
}

// generateFolderID generates a unique folder ID
func generateFolderID() string {
	// In production, use a proper UUID or cryptographic random
	return "folder_" + randString(16)
}

// Helper function for ID generation
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seeded := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seeded.Intn(len(charset))]
	}
	return string(result)
}

// NewHiddenMessage creates a new hidden message
func NewHiddenMessage(id, folderID, conversationID, contentType string) *HiddenMessage {
	return &HiddenMessage{
		ID:             id,
		FolderID:       folderID,
		ConversationID: conversationID,
		ContentType:    contentType,
		Timestamp:      time.Now(),
		IsRead:         false,
		DeliveryStatus: "delivered",
	}
}

// NewSecurityEvent creates a new security event
func NewSecurityEvent(folderID, eventType, details string) *SecurityEvent {
	return &SecurityEvent{
		ID:        generateEventID(),
		FolderID:  folderID,
		EventType: eventType,
		Timestamp: time.Now(),
		Details:   details,
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return "event_" + randString(12)
}

// NewBiometricLockoutTracker creates a new lockout tracker
func NewBiometricLockoutTracker(folderID string) *BiometricLockoutTracker {
	return &BiometricLockoutTracker{
		FolderID:       folderID,
		FailedAttempts: 0,
		IsLockedOut:    false,
	}
}

// RecordFailedAttempt records a failed biometric attempt and returns the new count
func (blt *BiometricLockoutTracker) RecordFailedAttempt() int {
	blt.FailedAttempts++
	now := time.Now()
	blt.LastFailureTime = &now
	return blt.FailedAttempts
}

// ResetFailedAttempts resets the counter after successful biometric
func (blt *BiometricLockoutTracker) ResetFailedAttempts() {
	blt.FailedAttempts = 0
	blt.IsLockedOut = false
	blt.LastFailureTime = nil
	blt.LockoutExpiresAt = nil
	blt.WipeScheduledAt = nil
}

// IsWipeTriggered returns true if 10 failed attempts have been recorded
func (blt *BiometricLockoutTracker) IsWipeTriggered() bool {
	return blt.FailedAttempts >= 10
}

// GetLockoutStatus returns a human-readable lockout status
func (blt *BiometricLockoutTracker) GetLockoutStatus() string {
	switch blt.FailedAttempts {
	case 0, 1, 2, 3:
		return "No lockout"
	case 4, 5:
		return "30-second cooldown active"
	case 6, 7, 8:
		return "5-minute cooldown active"
	case 9:
		return "Final warning: one more failure will wipe this folder"
	default:
		if blt.FailedAttempts >= 10 {
			return "Folder wiped after 10 failed attempts"
		}
		return "Unknown"
	}
}
