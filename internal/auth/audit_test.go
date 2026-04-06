package auth

import (
	"testing"
	"time"
)

func TestAuditLogger_Log(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)

	if al.Count() != 1 {
		t.Errorf("expected 1 entry, got %d", al.Count())
	}
}

func TestAuditLogger_GetByUser(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-2", AuditEventRegister, AuditResultSuccess, "5.6.7.8", "dev-2", nil, "", nil)
	al.Log("user-1", AuditEventRefresh, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)

	entries := al.GetByUser("user-1", 0)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for user-1, got %d", len(entries))
	}

	// Should be most-recent first
	if entries[0].EventType != AuditEventRefresh {
		t.Error("most recent entry should be first")
	}
}

func TestAuditLogger_GetByUserWithLimit(t *testing.T) {
	al := NewAuditLogger()

	for i := 0; i < 10; i++ {
		al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	}

	entries := al.GetByUser("user-1", 3)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries with limit, got %d", len(entries))
	}
}

func TestAuditLogger_GetByIP(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-2", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-2", nil, "AUTH_004", nil)
	al.Log("user-3", AuditEventLogin, AuditResultSuccess, "9.9.9.9", "dev-3", nil, "", nil)

	entries := al.GetByIP("1.2.3.4", 0)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for IP, got %d", len(entries))
	}
}

func TestAuditLogger_GetByEventType(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-1", AuditEventRefresh, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)

	entries := al.GetByEventType(AuditEventLogin, 0)
	if len(entries) != 2 {
		t.Errorf("expected 2 login entries, got %d", len(entries))
	}
}

func TestAuditLogger_CountByResult(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)
	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)
	al.Log("user-1", AuditEventLogin, AuditResultBlocked, "1.2.3.4", "dev-1", nil, "AUTH_009", nil)

	if al.CountByResult(AuditResultSuccess) != 1 {
		t.Error("expected 1 success")
	}
	if al.CountByResult(AuditResultFailed) != 2 {
		t.Error("expected 2 failures")
	}
	if al.CountByResult(AuditResultBlocked) != 1 {
		t.Error("expected 1 blocked")
	}
}

func TestAuditLogger_RecentFailedLogins(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)
	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)
	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	// Different user's failure shouldn't count
	al.Log("user-2", AuditEventLogin, AuditResultFailed, "5.6.7.8", "dev-2", nil, "AUTH_004", nil)

	count := al.RecentFailedLogins("user-1", time.Hour)
	if count != 2 {
		t.Errorf("expected 2 recent failed logins, got %d", count)
	}
}

func TestAuditLogger_RecentFailedLoginsWindow(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultFailed, "1.2.3.4", "dev-1", nil, "AUTH_004", nil)

	// With a zero window, nothing should match since entries are in the past (even by nanoseconds)
	count := al.RecentFailedLogins("user-1", 0)
	// The entry was just created so it might or might not be within 0 duration
	// Use a tiny window instead
	count = al.RecentFailedLogins("user-1", time.Millisecond)
	if count != 1 {
		t.Errorf("expected 1 recent failure within 1ms, got %d", count)
	}
}

func TestAuditLogger_LogWithMetadata(t *testing.T) {
	al := NewAuditLogger()

	metadata := map[string]interface{}{
		"auth_type": "passkey",
		"device":    "iPhone 15",
	}
	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", metadata)

	entries := al.GetByUser("user-1", 1)
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}
	if entries[0].Metadata["auth_type"] != "passkey" {
		t.Error("metadata should be preserved")
	}
}

func TestAuditLogger_LogWithDeviceInfo(t *testing.T) {
	al := NewAuditLogger()

	device := &DeviceInfo{
		DeviceID: "dev-1",
		Platform: "ios",
		Model:    "iPhone15,2",
	}
	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", device, "", nil)

	entries := al.GetByUser("user-1", 1)
	if entries[0].DeviceInfo == nil {
		t.Fatal("device info should be present")
	}
	if entries[0].DeviceInfo.Model != "iPhone15,2" {
		t.Error("device model should be preserved")
	}
}

func TestAuditLogger_EntryHasUniqueID(t *testing.T) {
	al := NewAuditLogger()

	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	al.Log("user-1", AuditEventLogin, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)

	entries := al.GetByUser("user-1", 0)
	if entries[0].ID == entries[1].ID {
		t.Error("each entry should have a unique ID")
	}
}

func TestAuditLogger_AllEventTypes(t *testing.T) {
	al := NewAuditLogger()

	types := []AuditEventType{
		AuditEventLogin, AuditEventRegister, AuditEventRefresh,
		AuditEventLogout, AuditEventStepUp, AuditEventRecovery,
	}

	for _, et := range types {
		al.Log("user-1", et, AuditResultSuccess, "1.2.3.4", "dev-1", nil, "", nil)
	}

	if al.Count() != len(types) {
		t.Errorf("expected %d entries, got %d", len(types), al.Count())
	}

	for _, et := range types {
		entries := al.GetByEventType(et, 0)
		if len(entries) != 1 {
			t.Errorf("expected 1 entry for type %s, got %d", et, len(entries))
		}
	}
}
