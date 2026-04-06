package database

import (
	"context"
	"testing"
	"time"
)

func TestUserCRUD(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	user := &User{
		UserID:   "user-1",
		DID:      "did:test:alice",
		Username: "alice",
	}

	if err := db.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.TrustTier != 1 {
		t.Errorf("expected default tier 1, got %d", user.TrustTier)
	}

	got, err := db.GetUserByDID(ctx, "did:test:alice")
	if err != nil {
		t.Fatalf("GetUserByDID: %v", err)
	}
	if got.Username != "alice" {
		t.Errorf("expected alice, got %s", got.Username)
	}

	got2, err := db.GetUserByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if got2.DID != "did:test:alice" {
		t.Errorf("expected did:test:alice, got %s", got2.DID)
	}

	// Duplicate should fail
	if err := db.CreateUser(ctx, user); err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}

	// Not found
	_, err = db.GetUserByDID(ctx, "did:test:unknown")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTrustScores(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	ts := &TrustScore{
		DID:   "did:test:alice",
		Score: 85.5,
		Tier:  4,
	}
	if err := db.SetTrustScore(ctx, ts); err != nil {
		t.Fatalf("SetTrustScore: %v", err)
	}

	got, err := db.GetTrustScore(ctx, "did:test:alice")
	if err != nil {
		t.Fatalf("GetTrustScore: %v", err)
	}
	if got.Score != 85.5 {
		t.Errorf("expected 85.5, got %f", got.Score)
	}

	// Expired score
	ts2 := &TrustScore{
		DID:       "did:test:bob",
		Score:     50.0,
		Tier:      2,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	db.SetTrustScore(ctx, ts2)
	_, err = db.GetTrustScore(ctx, "did:test:bob")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for expired score, got %v", err)
	}

	// Batch
	scores, err := db.GetTrustScoreBatch(ctx, []string{"did:test:alice", "did:test:bob"})
	if err != nil {
		t.Fatalf("GetTrustScoreBatch: %v", err)
	}
	if len(scores) != 1 {
		t.Errorf("expected 1 non-expired score, got %d", len(scores))
	}
}

func TestCredentials(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	cred := &Credential{
		CredentialID: "cred-1",
		IssuerDID:    "did:issuer",
		SubjectDID:   "did:subject",
		SchemaID:     "schema-1",
		Status:       "active",
	}
	if err := db.StoreCredential(ctx, cred); err != nil {
		t.Fatalf("StoreCredential: %v", err)
	}

	creds, err := db.GetCredentialsByDID(ctx, "did:subject")
	if err != nil {
		t.Fatalf("GetCredentialsByDID: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}

	if err := db.RevokeCredential(ctx, "cred-1"); err != nil {
		t.Fatalf("RevokeCredential: %v", err)
	}
	if cred.Status != "revoked" {
		t.Errorf("expected revoked, got %s", cred.Status)
	}

	// Empty list for unknown DID
	empty, _ := db.GetCredentialsByDID(ctx, "did:unknown")
	if len(empty) != 0 {
		t.Errorf("expected empty list, got %d", len(empty))
	}
}

func TestMessageQueue(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	msg := &QueuedMessage{
		MessageID:    "msg-1",
		SenderDID:    "did:alice",
		RecipientDID: "did:bob",
		Payload:      []byte("hello"),
	}
	if err := db.Enqueue(ctx, msg); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if msg.Status != "queued" {
		t.Errorf("expected queued, got %s", msg.Status)
	}

	msgs, err := db.Dequeue(ctx, "did:bob", 10)
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Status != "delivered" {
		t.Errorf("expected delivered, got %s", msgs[0].Status)
	}

	// MarkDelivered
	msg2 := &QueuedMessage{MessageID: "msg-2", SenderDID: "did:a", RecipientDID: "did:b", Payload: []byte("x")}
	db.Enqueue(ctx, msg2)
	if err := db.MarkDelivered(ctx, "msg-2"); err != nil {
		t.Fatalf("MarkDelivered: %v", err)
	}

	// Not found
	if err := db.MarkDelivered(ctx, "msg-unknown"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMessageQueueOverflow(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	for i := 0; i < 1001; i++ {
		msg := &QueuedMessage{
			MessageID:    "msg-" + time.Now().Format("150405.000000000") + "-" + string(rune(i)),
			SenderDID:    "did:alice",
			RecipientDID: "did:bob",
			Payload:      []byte("data"),
		}
		db.Enqueue(ctx, msg)
	}

	// Queue should be capped at 1000
	msgs, _ := db.Dequeue(ctx, "did:bob", 2000)
	if len(msgs) > 1000 {
		t.Errorf("expected at most 1000 messages, got %d", len(msgs))
	}
}

func TestMerkleBatches(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	batch := &MerkleBatch{BatchID: "batch-1"}
	if err := db.CreateBatch(ctx, batch); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	if batch.Status != "open" {
		t.Errorf("expected open, got %s", batch.Status)
	}

	if err := db.FinalizeBatch(ctx, "batch-1", "abc123", 5); err != nil {
		t.Fatalf("FinalizeBatch: %v", err)
	}

	got, _ := db.GetBatch(ctx, "batch-1")
	if got.Status != "finalized" {
		t.Errorf("expected finalized, got %s", got.Status)
	}
	if got.LeafCount != 5 {
		t.Errorf("expected 5 leaves, got %d", got.LeafCount)
	}
}

func TestContacts(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	c := &Contact{OwnerDID: "did:alice", ContactDID: "did:bob", AddedVia: "manual"}
	if err := db.AddContact(ctx, c); err != nil {
		t.Fatalf("AddContact: %v", err)
	}

	// Duplicate
	if err := db.AddContact(ctx, c); err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}

	contacts, _ := db.GetContacts(ctx, "did:alice")
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}

	count, _ := db.GetContactCount(ctx, "did:alice")
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Block/unblock
	db.SetBlocked(ctx, "did:alice", "did:bob", true)
	if !contacts[0].Blocked {
		t.Errorf("expected blocked")
	}
	db.SetBlocked(ctx, "did:alice", "did:bob", false)
	if contacts[0].Blocked {
		t.Errorf("expected unblocked")
	}

	// Remove
	if err := db.RemoveContact(ctx, "did:alice", "did:bob"); err != nil {
		t.Fatalf("RemoveContact: %v", err)
	}
	contacts, _ = db.GetContacts(ctx, "did:alice")
	if len(contacts) != 0 {
		t.Errorf("expected 0 contacts after remove, got %d", len(contacts))
	}
}

func TestInviteLinks(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	inv := &InviteLink{Code: "abc123", CreatorDID: "did:alice"}
	if err := db.CreateInvite(ctx, inv); err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}

	got, err := db.GetInvite(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetInvite: %v", err)
	}
	if got.CreatorDID != "did:alice" {
		t.Errorf("expected did:alice, got %s", got.CreatorDID)
	}

	if err := db.AcceptInvite(ctx, "abc123", "did:bob"); err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if !got.Accepted {
		t.Errorf("expected accepted")
	}

	// Double accept
	if err := db.AcceptInvite(ctx, "abc123", "did:charlie"); err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate on double accept, got %v", err)
	}
}

func TestMediaFiles(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	file := &MediaFile{FileID: "file-1", UploaderDID: "did:alice", ContentType: "image/png", EncryptedSize: 1024}
	if err := db.StoreMediaFile(ctx, file); err != nil {
		t.Fatalf("StoreMediaFile: %v", err)
	}
	if file.ScanStatus != "pending" {
		t.Errorf("expected pending, got %s", file.ScanStatus)
	}

	got, _ := db.GetMediaFile(ctx, "file-1")
	if got.ContentType != "image/png" {
		t.Errorf("expected image/png, got %s", got.ContentType)
	}

	db.UpdateScanStatus(ctx, "file-1", "clean")
	if got.ScanStatus != "clean" {
		t.Errorf("expected clean, got %s", got.ScanStatus)
	}

	// Chunks
	chunk := &MediaChunk{ChunkID: "chunk-1", FileID: "file-1", Index: 0, Size: 512}
	db.StoreChunk(ctx, chunk)
	chunks, _ := db.GetChunks(ctx, "file-1")
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestNotifications(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	device := &UserDevice{
		DeviceID:    "dev-1",
		DID:         "did:alice",
		DeviceLabel: "iPhone",
		APNsToken:   "token123456",
	}
	if err := db.RegisterDevice(ctx, device); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	devices, _ := db.GetDevicesByDID(ctx, "did:alice")
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}

	db.UpdateAPNsToken(ctx, "dev-1", "newtoken123")
	if devices[0].APNsToken != "newtoken123" {
		t.Errorf("expected newtoken123, got %s", devices[0].APNsToken)
	}

	// Prefs - defaults
	prefs, _ := db.GetNotificationPrefs(ctx, "did:alice")
	if !prefs.PushEnabled {
		t.Errorf("expected push enabled by default")
	}

	// Update prefs
	prefs.PushEnabled = false
	db.UpdateNotificationPrefs(ctx, prefs)
	got, _ := db.GetNotificationPrefs(ctx, "did:alice")
	if got.PushEnabled {
		t.Errorf("expected push disabled after update")
	}
}

func TestLogIndex(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	now := time.Now()
	entry := &LogIndexEntry{
		EntryID:       "log-1",
		DID:           "did:alice",
		EventType:     "message",
		TimeRangeFrom: now.Add(-1 * time.Hour),
		TimeRangeTo:   now,
	}
	db.StoreLogIndex(ctx, entry)

	results, _ := db.GetLogIndex(ctx, now.Add(-2*time.Hour), now.Add(time.Hour))
	if len(results) != 1 {
		t.Errorf("expected 1 log entry, got %d", len(results))
	}
}

func TestPurgeExpired(t *testing.T) {
	db := NewMemoryDB()
	ctx := context.Background()

	msg := &QueuedMessage{
		MessageID:    "msg-expired",
		SenderDID:    "did:a",
		RecipientDID: "did:b",
		Payload:      []byte("x"),
	}
	db.Enqueue(ctx, msg)
	// Force expiry
	msg.ExpiresAt = time.Now().Add(-1 * time.Hour)

	count, _ := db.PurgeExpired(ctx)
	if count != 1 {
		t.Errorf("expected 1 purged, got %d", count)
	}
}
