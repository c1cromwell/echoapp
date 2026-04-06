package notification

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func setupTest() (*Service, database.DB) {
	db := database.NewMemoryDB()
	svc := NewService(db)
	return svc, db
}

func TestRegisterDevice(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	device, err := svc.RegisterDevice(ctx, "did:alice", "iPhone 15", "pubkey123", "apnstoken1234")
	if err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}
	if device.DeviceID == "" {
		t.Error("expected non-empty device ID")
	}
	if device.DID != "did:alice" {
		t.Errorf("expected did:alice, got %s", device.DID)
	}
}

func TestRegisterDeviceInvalidToken(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	_, err := svc.RegisterDevice(ctx, "did:alice", "iPhone", "pk", "short")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestUpdateAPNsToken(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{
		DeviceID: "dev-1", DID: "did:alice", APNsToken: "oldtoken12345",
	})

	if err := svc.UpdateAPNsToken(ctx, "dev-1", "newtoken1234"); err != nil {
		t.Fatalf("UpdateAPNsToken: %v", err)
	}
}

func TestUpdateAPNsTokenInvalid(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	err := svc.UpdateAPNsToken(ctx, "dev-1", "short")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestSendNotification(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{
		DeviceID: "dev-1", DID: "did:bob", APNsToken: "tokenabcd1234",
	})

	payload := PushPayload{
		Type:           TypeMessage,
		ConversationID: "conv-1",
		SenderDID:      "did:alice",
	}

	result, err := svc.Send(ctx, "did:bob", payload)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if result.DeviceCount != 1 {
		t.Errorf("expected 1 device, got %d", result.DeviceCount)
	}
	if result.Delivered != 1 {
		t.Errorf("expected 1 delivered, got %d", result.Delivered)
	}
}

func TestSendNoDevices(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	_, err := svc.Send(ctx, "did:nobody", PushPayload{Type: TypeMessage})
	if err != ErrNoDevices {
		t.Errorf("expected ErrNoDevices, got %v", err)
	}
}

func TestSendPushDisabled(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{
		DeviceID: "dev-1", DID: "did:bob", APNsToken: "token1234567",
	})
	db.UpdateNotificationPrefs(ctx, &database.NotificationPrefs{
		DID:         "did:bob",
		PushEnabled: false,
	})

	_, err := svc.Send(ctx, "did:bob", PushPayload{Type: TypeMessage})
	if err != ErrPushDisabled {
		t.Errorf("expected ErrPushDisabled, got %v", err)
	}
}

func TestSendSystemAlwaysDelivered(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{
		DeviceID: "dev-1", DID: "did:bob", APNsToken: "token1234567",
	})
	db.UpdateNotificationPrefs(ctx, &database.NotificationPrefs{
		DID:         "did:bob",
		PushEnabled: false,
	})

	result, err := svc.Send(ctx, "did:bob", PushPayload{Type: TypeSystem})
	if err != nil {
		t.Fatalf("system notifications should always be delivered: %v", err)
	}
	if result.Delivered != 1 {
		t.Errorf("expected system notification delivered")
	}
}

func TestGroupNotificationsDisabled(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{
		DeviceID: "dev-1", DID: "did:bob", APNsToken: "token1234567",
	})
	db.UpdateNotificationPrefs(ctx, &database.NotificationPrefs{
		DID:                "did:bob",
		PushEnabled:        true,
		GroupNotifications: false,
	})

	_, err := svc.Send(ctx, "did:bob", PushPayload{Type: TypeGroup})
	if err != ErrPushDisabled {
		t.Errorf("expected ErrPushDisabled for disabled group notifs, got %v", err)
	}
}

func TestPreferences(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	prefs, err := svc.GetPreferences(ctx, "did:alice")
	if err != nil {
		t.Fatalf("GetPreferences: %v", err)
	}
	if !prefs.PushEnabled {
		t.Error("expected push enabled by default")
	}

	prefs.PushEnabled = false
	if err := svc.UpdatePreferences(ctx, prefs); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	got, _ := svc.GetPreferences(ctx, "did:alice")
	if got.PushEnabled {
		t.Error("expected push disabled after update")
	}
}

func TestSendBatch(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{DeviceID: "d1", DID: "did:a", APNsToken: "token1234567"})
	db.RegisterDevice(ctx, &database.UserDevice{DeviceID: "d2", DID: "did:b", APNsToken: "token2345678"})

	results, err := svc.SendBatch(ctx, []string{"did:a", "did:b", "did:c"}, PushPayload{Type: TypeMessage})
	if err != nil {
		t.Fatalf("SendBatch: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (did:c has no device), got %d", len(results))
	}
}

func TestQueueDepth(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.RegisterDevice(ctx, &database.UserDevice{DeviceID: "d1", DID: "did:alice", APNsToken: "t1234567"})
	db.RegisterDevice(ctx, &database.UserDevice{DeviceID: "d2", DID: "did:alice", APNsToken: "t2345678"})

	depth, _ := svc.QueueDepth(ctx, "did:alice")
	if depth != 2 {
		t.Errorf("expected queue depth 2, got %d", depth)
	}
}
