// Package validation provides server-side pre-validation for messaging endpoints (WO-35).
package validation

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

const (
	maxGroupIDLen       = 128
	maxGroupNameLen     = 100
	maxGroupDescLen     = 500
	maxDeviceIDLen      = 128
	maxSyncCiphertext   = 4 << 20 // 4 MiB opaque blob cap per push
	maxBackupCiphertext = 64 << 20
)

// ValidateDIDKey ensures a non-empty did:key URI parses under pkg/didkey.
func ValidateDIDKey(did string) error {
	did = strings.TrimSpace(did)
	if did == "" {
		return errors.New("did is required")
	}
	if !strings.HasPrefix(did, "did:key:") {
		return errors.New("only did:key identities are supported in Phase 1–3")
	}
	if _, err := didkey.Parse(did); err != nil {
		return fmt.Errorf("invalid did:key: %w", err)
	}
	return nil
}

// ValidateGroupID checks the client-supplied group identifier.
func ValidateGroupID(groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return errors.New("group id is required")
	}
	if len(groupID) > maxGroupIDLen {
		return fmt.Errorf("group id exceeds %d characters", maxGroupIDLen)
	}
	for _, r := range groupID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == ':' {
			continue
		}
		return errors.New("group id contains invalid characters")
	}
	return nil
}

// ValidateGroupCreate validates group creation input before business rules run.
func ValidateGroupCreate(groupID, ownerDID string, groupType groups.GroupType, name, description string) error {
	if err := ValidateGroupID(groupID); err != nil {
		return err
	}
	if err := ValidateDIDKey(ownerDID); err != nil {
		return fmt.Errorf("owner: %w", err)
	}
	switch groupType {
	case groups.GroupTypePublic, groups.GroupTypePrivate, groups.GroupTypeSecret:
	default:
		return groups.ErrInvalidGroupType
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("group name is required")
	}
	if utf8.RuneCountInString(name) > maxGroupNameLen {
		return fmt.Errorf("group name exceeds %d characters", maxGroupNameLen)
	}
	if utf8.RuneCountInString(description) > maxGroupDescLen {
		return fmt.Errorf("group description exceeds %d characters", maxGroupDescLen)
	}
	return nil
}

// ValidateGroupMemberDID validates a member DID for add/remove operations.
func ValidateGroupMemberDID(memberDID string) error {
	return ValidateDIDKey(memberDID)
}

// ValidateDeviceID checks a linked-device identifier for sync streams.
func ValidateDeviceID(deviceID string) error {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return errors.New("device id is required")
	}
	if len(deviceID) > maxDeviceIDLen {
		return fmt.Errorf("device id exceeds %d characters", maxDeviceIDLen)
	}
	return nil
}

// ValidateSyncPush validates opaque history-sync push payloads (WO-CA3).
func ValidateSyncPush(targetDeviceID string, ciphertext []byte) error {
	if err := ValidateDeviceID(targetDeviceID); err != nil {
		return err
	}
	if len(ciphertext) == 0 {
		return errors.New("ciphertext is required")
	}
	if len(ciphertext) > maxSyncCiphertext {
		return fmt.Errorf("ciphertext exceeds %d bytes", maxSyncCiphertext)
	}
	return nil
}

// ValidateMessageID checks client message identifiers for message-ops endpoints.
func ValidateMessageID(messageID string) error {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return errors.New("message id is required")
	}
	if len(messageID) > 128 {
		return fmt.Errorf("message id exceeds %d characters", 128)
	}
	for _, r := range messageID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == ':' {
			continue
		}
		return errors.New("message id contains invalid characters")
	}
	return nil
}

// ValidateConversationID checks dm: or group: conversation identifiers.
func ValidateConversationID(conversationID string) error {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return errors.New("conversation id is required")
	}
	if len(conversationID) > 256 {
		return fmt.Errorf("conversation id exceeds %d characters", 256)
	}
	if strings.HasPrefix(conversationID, "dm:") || strings.HasPrefix(conversationID, "group:") {
		return nil
	}
	return errors.New("conversation id must start with dm: or group:")
}

// ValidateMessageRefs validates durable reply/forward metadata (WO-59 / S2).
func ValidateMessageRefs(messageID, conversationID, replyTo, forwardedFrom, forwardedFromConv string) error {
	if err := ValidateMessageID(messageID); err != nil {
		return err
	}
	if err := ValidateConversationID(conversationID); err != nil {
		return err
	}
	if replyTo == "" && forwardedFrom == "" {
		return errors.New("reply_to_message_id or forwarded_from_message_id required")
	}
	if replyTo != "" {
		if err := ValidateMessageID(replyTo); err != nil {
			return fmt.Errorf("reply_to_message_id: %w", err)
		}
	}
	if forwardedFrom != "" {
		if err := ValidateMessageID(forwardedFrom); err != nil {
			return fmt.Errorf("forwarded_from_message_id: %w", err)
		}
	}
	if forwardedFromConv != "" {
		if err := ValidateConversationID(forwardedFromConv); err != nil {
			return fmt.Errorf("forwarded_from_conversation_id: %w", err)
		}
	}
	return nil
}

// ValidateBackupCiphertext caps encrypted backup blob size (WO-64).
func ValidateBackupCiphertext(ciphertext []byte) error {
	if len(ciphertext) == 0 {
		return errors.New("ciphertext is required")
	}
	if len(ciphertext) > maxBackupCiphertext {
		return fmt.Errorf("backup ciphertext exceeds %d bytes", maxBackupCiphertext)
	}
	return nil
}
