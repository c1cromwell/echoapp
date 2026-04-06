package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditLogger records authentication events for security monitoring.
type AuditLogger struct {
	mu      sync.RWMutex
	entries []*AuditLogEntry
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		entries: make([]*AuditLogEntry, 0),
	}
}

// Log records an authentication event.
func (al *AuditLogger) Log(userID string, eventType AuditEventType, result AuditResult, ip string, deviceID string, deviceInfo *DeviceInfo, errorCode string, metadata map[string]interface{}) {
	entry := &AuditLogEntry{
		ID:         uuid.New().String(),
		UserID:     userID,
		EventType:  eventType,
		Result:     result,
		IPAddress:  ip,
		DeviceID:   deviceID,
		DeviceInfo: deviceInfo,
		ErrorCode:  errorCode,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
	}

	al.mu.Lock()
	al.entries = append(al.entries, entry)
	al.mu.Unlock()
}

// GetByUser returns audit entries for a user, most recent first.
func (al *AuditLogger) GetByUser(userID string, limit int) []*AuditLogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []*AuditLogEntry
	// Iterate in reverse for most-recent-first
	for i := len(al.entries) - 1; i >= 0; i-- {
		if al.entries[i].UserID == userID {
			result = append(result, al.entries[i])
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// GetByIP returns audit entries for an IP address.
func (al *AuditLogger) GetByIP(ip string, limit int) []*AuditLogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []*AuditLogEntry
	for i := len(al.entries) - 1; i >= 0; i-- {
		if al.entries[i].IPAddress == ip {
			result = append(result, al.entries[i])
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// GetByEventType returns entries of a specific event type.
func (al *AuditLogger) GetByEventType(eventType AuditEventType, limit int) []*AuditLogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var result []*AuditLogEntry
	for i := len(al.entries) - 1; i >= 0; i-- {
		if al.entries[i].EventType == eventType {
			result = append(result, al.entries[i])
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// Count returns the total number of audit entries.
func (al *AuditLogger) Count() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.entries)
}

// CountByResult returns the number of entries with a specific result.
func (al *AuditLogger) CountByResult(result AuditResult) int {
	al.mu.RLock()
	defer al.mu.RUnlock()

	count := 0
	for _, entry := range al.entries {
		if entry.Result == result {
			count++
		}
	}
	return count
}

// RecentFailedLogins returns the number of failed login attempts for a user
// within the given duration.
func (al *AuditLogger) RecentFailedLogins(userID string, within time.Duration) int {
	al.mu.RLock()
	defer al.mu.RUnlock()

	cutoff := time.Now().Add(-within)
	count := 0
	for _, entry := range al.entries {
		if entry.UserID == userID &&
			entry.EventType == AuditEventLogin &&
			entry.Result == AuditResultFailed &&
			entry.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count
}
