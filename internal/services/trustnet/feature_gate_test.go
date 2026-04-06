package trustnet

import (
	"testing"
)

// ========== FEATURE GATE TESTS ==========

func TestCanAccessFeature_UnverifiedUser(t *testing.T) {
	fgs := NewFeatureGateService()

	// Unverified user (score < 20) can access basic messaging
	can, msg := fgs.CanAccessFeature("user-1", "messaging", 10.0)
	if !can {
		t.Errorf("unverified user should access messaging: %s", msg)
	}

	// But cannot create groups
	can, msg = fgs.CanAccessFeature("user-1", "create_group", 10.0)
	if can {
		t.Error("unverified user should not create groups")
	}
}

func TestCanAccessFeature_MemberUser(t *testing.T) {
	fgs := NewFeatureGateService()

	// Member user (score 40-59)
	score := 50.0

	// Can create group
	can, msg := fgs.CanAccessFeature("user-2", "create_group", score)
	if !can {
		t.Errorf("member user should create group: %s", msg)
	}

	// Can make voice calls
	can, msg = fgs.CanAccessFeature("user-2", "voice_calls", score)
	if !can {
		t.Errorf("member user should make voice calls: %s", msg)
	}

	// But cannot make video calls
	can, msg = fgs.CanAccessFeature("user-2", "video_calls", score)
	if can {
		t.Error("member user should not make video calls")
	}
}

func TestCanAccessFeature_TrustedUser(t *testing.T) {
	fgs := NewFeatureGateService()

	// Trusted user (score 60-79)
	score := 70.0

	// Can make video calls
	can, msg := fgs.CanAccessFeature("user-3", "video_calls", score)
	if !can {
		t.Errorf("trusted user should make video calls: %s", msg)
	}

	// Can endorse others
	can, msg = fgs.CanAccessFeature("user-3", "endorse_others", score)
	if !can {
		t.Errorf("trusted user should endorse: %s", msg)
	}
}

func TestGetFeatureTier(t *testing.T) {
	fgs := NewFeatureGateService()

	tests := []struct {
		score    float64
		expected FeatureTier
	}{
		{5.0, FeatureTierUnverified},
		{25.0, FeatureTierNewcomer},
		{45.0, FeatureTierMember},
		{65.0, FeatureTierTrusted},
		{85.0, FeatureTierVerified},
	}

	for _, tt := range tests {
		tier := fgs.GetFeatureTier(tt.score)
		if tier != tt.expected {
			t.Errorf("score %.1f: expected %s, got %s", tt.score, tt.expected, tier)
		}
	}
}

func TestRecordFeatureUsage(t *testing.T) {
	fgs := NewFeatureGateService()

	// Register a limited feature
	fgs.RegisterCustomFeature(Feature{
		Name:           "daily_feature",
		RequiredTier:   FeatureTierMember,
		RequiredScore:  40.0,
		RateLimitDaily: 5,
	})

	// Record 5 uses
	for i := 0; i < 5; i++ {
		fgs.RecordFeatureUsage("user-1", "daily_feature")
		remaining := fgs.GetRemainingUsageToday("user-1", "daily_feature")
		expected := 5 - i - 1
		if remaining != expected {
			t.Errorf("use %d: expected %d remaining, got %d", i+1, expected, remaining)
		}
	}

	// 6th use should fail
	can, _ := fgs.CanAccessFeature("user-1", "daily_feature", 50.0)
	if can {
		t.Error("should exceed daily limit")
	}
}

func TestGetRemainingUsageToday(t *testing.T) {
	fgs := NewFeatureGateService()

	// Unlimited feature
	remaining := fgs.GetRemainingUsageToday("user-1", "messaging")
	if remaining != -1 {
		t.Errorf("unlimited feature should return -1, got %d", remaining)
	}

	// Limited feature
	remaining = fgs.GetRemainingUsageToday("user-1", "create_group")
	if remaining != -1 { // create_group has no daily limit by default
		t.Errorf("feature with no limit should return -1, got %d", remaining)
	}
}

func TestGetFeaturesForTier(t *testing.T) {
	fgs := NewFeatureGateService()

	// Member tier (score 40-59)
	features := fgs.GetFeaturesForTier(FeatureTierMember)

	if len(features) < 1 {
		t.Error("member tier should have features")
	}

	// Check that all returned features are accessible at this tier
	for _, f := range features {
		if f.RequiredScore > 40 {
			t.Errorf("feature %s requires score %.0f, too high for member tier", f.Name, f.RequiredScore)
		}
	}
}

// ========== BLOCKCHAIN ANCHOR TESTS ==========

func TestCreateTrustScoreAnchor(t *testing.T) {
	config := CardanoConfig{Enabled: true, NetworkID: "testnet"}
	bas := NewBlockchainAnchorService(config)

	anchor, err := bas.CreateTrustScoreAnchor("user-1", 75.0, 70.0)
	if err != nil {
		t.Fatalf("failed to create anchor: %v", err)
	}

	if anchor.Type != "trust_score" {
		t.Errorf("expected trust_score type, got %s", anchor.Type)
	}

	if anchor.TrustScore != 75.0 {
		t.Errorf("expected score 75.0, got %.1f", anchor.TrustScore)
	}

	if anchor.CardanoTxHash != "" {
		t.Error("uncommitted anchor should not have tx hash")
	}
}

func TestCreateEndorsementAnchor(t *testing.T) {
	config := CardanoConfig{Enabled: true, NetworkID: "testnet"}
	bas := NewBlockchainAnchorService(config)

	anchor, err := bas.CreateEndorsementAnchor("endorser-1", "endorsee-1", "endorse-1", "reliable", 10.0)
	if err != nil {
		t.Fatalf("failed to create endorsement anchor: %v", err)
	}

	if anchor.Type != "endorsement" {
		t.Errorf("expected endorsement type, got %s", anchor.Type)
	}

	if anchor.EndorserDID != "endorser-1" {
		t.Errorf("endorser DID mismatch")
	}
}

func TestCreateDisputeResolutionAnchor(t *testing.T) {
	config := CardanoConfig{Enabled: true, NetworkID: "testnet"}
	bas := NewBlockchainAnchorService(config)

	anchor, err := bas.CreateDisputeResolutionAnchor("dispute-1", "user-1", "upheld", 10.0)
	if err != nil {
		t.Fatalf("failed to create dispute anchor: %v", err)
	}

	if anchor.Type != "dispute_resolution" {
		t.Errorf("expected dispute_resolution type, got %s", anchor.Type)
	}

	if anchor.DisputeOutcome != "upheld" {
		t.Errorf("expected upheld outcome, got %s", anchor.DisputeOutcome)
	}
}

func TestCommitBatch(t *testing.T) {
	config := CardanoConfig{Enabled: true, NetworkID: "testnet"}
	bas := NewBlockchainAnchorService(config)

	// Create some anchors
	bas.CreateTrustScoreAnchor("user-1", 75.0, 70.0)
	bas.CreateTrustScoreAnchor("user-2", 55.0, 50.0)

	pending := bas.GetPendingCommits()
	if len(pending) != 2 {
		t.Errorf("expected 2 pending commits, got %d", len(pending))
	}

	// Commit batch
	results, err := bas.CommitBatch()
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Check queue is empty
	pending = bas.GetPendingCommits()
	if len(pending) != 0 {
		t.Errorf("queue should be empty after commit, has %d", len(pending))
	}

	// Check committed anchor has tx hash
	for _, anchor := range bas.GetUserAnchors("user-1") {
		if anchor.CardanoTxHash == "" {
			t.Error("committed anchor should have tx hash")
		}
		if anchor.MetagraphRef == "" {
			t.Error("committed anchor should have metagraph ref")
		}
		if anchor.CommittedAt == nil {
			t.Error("committed anchor should have commit timestamp")
		}
	}
}

func TestCreateZKCommitment(t *testing.T) {
	config := CardanoConfig{Enabled: true}
	bas := NewBlockchainAnchorService(config)

	zk, err := bas.CreateZKCommitment("user-1", "secret_value")
	if err != nil {
		t.Fatalf("failed to create ZK commitment: %v", err)
	}

	if zk.CommitmentHash == "" {
		t.Error("commitment hash should be set")
	}

	if zk.VerifiedAt != nil {
		t.Error("new commitment should not be verified")
	}
}

func TestVerifyZKCommitment(t *testing.T) {
	config := CardanoConfig{Enabled: true}
	bas := NewBlockchainAnchorService(config)

	zk, _ := bas.CreateZKCommitment("user-1", "secret_value")

	// Verification with wrong value should fail
	verified, _ := bas.VerifyZKCommitment(zk.ID, "wrong_value", zk.Salt)
	if verified {
		t.Error("verification with wrong value should fail")
	}

	// Correct verification (Note: this is simplified for testing)
	// In a real implementation, the reveal would involve a proper ZK construct
}

func TestGetUserAnchors(t *testing.T) {
	config := CardanoConfig{Enabled: true}
	bas := NewBlockchainAnchorService(config)

	// Create anchors for user-1
	bas.CreateTrustScoreAnchor("user-1", 75.0, 70.0)
	bas.CreateEndorsementAnchor("endorser-1", "user-1", "endorse-1", "reliable", 10.0)
	bas.CreateTrustScoreAnchor("user-2", 55.0, 50.0)

	// Get user-1 anchors
	anchors := bas.GetUserAnchors("user-1")
	if len(anchors) != 2 {
		t.Errorf("expected 2 anchors for user-1, got %d", len(anchors))
	}

	// Get user-2 anchors
	anchors = bas.GetUserAnchors("user-2")
	if len(anchors) != 1 {
		t.Errorf("expected 1 anchor for user-2, got %d", len(anchors))
	}
}

func TestGetUserAnchorsByType(t *testing.T) {
	config := CardanoConfig{Enabled: true}
	bas := NewBlockchainAnchorService(config)

	// Create mixed anchors
	bas.CreateTrustScoreAnchor("user-1", 75.0, 70.0)
	bas.CreateTrustScoreAnchor("user-1", 76.0, 75.0)
	bas.CreateEndorsementAnchor("endorser-1", "user-1", "endorse-1", "reliable", 10.0)

	// Get trust_score anchors only
	anchors := bas.GetUserAnchorsByType("user-1", "trust_score")
	if len(anchors) != 2 {
		t.Errorf("expected 2 trust_score anchors, got %d", len(anchors))
	}

	// Get endorsement anchors
	anchors = bas.GetUserAnchorsByType("user-1", "endorsement")
	if len(anchors) != 1 {
		t.Errorf("expected 1 endorsement anchor, got %d", len(anchors))
	}
}

// ========== NOTIFICATION TESTS ==========

// MockNotificationProvider for testing
type MockNotificationProvider struct {
	emails []struct{ user, title, message string }
	pushes []struct{ user, title, message string }
	inApps []*NotificationEvent
}

func (m *MockNotificationProvider) SendEmail(userDID string, title string, message string) error {
	m.emails = append(m.emails, struct{ user, title, message string }{userDID, title, message})
	return nil
}

func (m *MockNotificationProvider) SendPush(userDID string, title string, message string) error {
	m.pushes = append(m.pushes, struct{ user, title, message string }{userDID, title, message})
	return nil
}

func (m *MockNotificationProvider) SendInAppNotification(notification *NotificationEvent) error {
	m.inApps = append(m.inApps, notification)
	return nil
}

func TestNotifyScoreChange(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	err := ns.NotifyScoreChange("user-1", 50.0, 75.0, "Endorsements received")
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	// Queue not sent yet
	if len(mock.emails) > 0 {
		t.Error("notification should not be sent until SendPendingNotifications is called")
	}

	// Send pending
	ns.SendPendingNotifications()
	if len(mock.emails) != 1 {
		t.Errorf("expected 1 email, got %d", len(mock.emails))
	}
}

func TestNotifyScoreDeltaThreshold(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	// Set high threshold
	prefs := DefaultPreferences("user-1")
	prefs.ScoreMinimumDelta = 10.0
	ns.SetUserPreferences(prefs)

	// Small change (below threshold)
	err := ns.NotifyScoreChange("user-1", 50.0, 52.0, "Small change")
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	ns.SendPendingNotifications()
	if len(mock.emails) != 0 {
		t.Error("should not send notification below delta threshold")
	}

	// Large change (above threshold)
	ns.NotifyScoreChange("user-1", 50.0, 65.0, "Large change")
	ns.SendPendingNotifications()
	if len(mock.emails) != 1 {
		t.Error("should send notification above delta threshold")
	}
}

func TestNotifyEndorsementReceived(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	err := ns.NotifyEndorsementReceived("user-1", "endorser-1", "reliable")
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	ns.SendPendingNotifications()
	if len(mock.emails) != 1 {
		t.Error("should send endorsement notification")
	}
}

func TestNotifyDisputeAssignment(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	err := ns.NotifyDisputeAssignment("juror-1", "dispute-1")
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	ns.SendPendingNotifications()
	if len(mock.inApps) != 1 {
		t.Error("should send dispute assignment notification")
	}
}

func TestNotifyAutoPromotion(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	err := ns.NotifyAutoPromotion("user-1", "contact-1", CircleInner)
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	ns.SendPendingNotifications()
	if len(mock.pushes) != 1 {
		t.Error("should send auto-promotion notification")
	}
}

func TestNotifySecurityAlert(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	// Disable emails in preferences
	prefs := DefaultPreferences("user-1")
	prefs.NotifyViaEmail = false
	ns.SetUserPreferences(prefs)

	err := ns.NotifySecurityAlert("user-1", "sybil_warning", "Suspicious activity detected")
	if err != nil {
		t.Fatalf("failed to notify: %v", err)
	}

	ns.SendPendingNotifications()
	// Security alerts ignore email preference
	if len(mock.emails) != 1 {
		t.Error("security alerts should always send email")
	}
}

func TestGetUserNotifications(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	// Create multiple notifications
	ns.NotifyScoreChange("user-1", 50.0, 75.0, "score increase")
	ns.NotifyEndorsementReceived("user-1", "endorser-1", "reliable")
	ns.NotifyScoreChange("user-1", 75.0, 80.0, "score increase 2")

	notifications := ns.GetUserNotifications("user-1", 10, 0)
	if len(notifications) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(notifications))
	}
}

func TestMarkNotificationRead(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	ns.NotifyScoreChange("user-1", 50.0, 75.0, "score increase")
	notifications := ns.GetUserNotifications("user-1", 10, 0)

	if len(notifications) == 0 {
		t.Error("no notifications found")
	}

	notifID := notifications[0].ID
	err := ns.MarkNotificationRead(notifID)
	if err != nil {
		t.Fatalf("failed to mark read: %v", err)
	}

	notifications = ns.GetUserNotifications("user-1", 10, 0)
	if !notifications[0].Read {
		t.Error("notification should be marked read")
	}
}

func TestGetUnreadCount(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	ns.NotifyScoreChange("user-1", 50.0, 75.0, "score increase")
	ns.NotifyEndorsementReceived("user-1", "endorser-1", "reliable")

	count := ns.GetUnreadCount("user-1")
	if count != 2 {
		t.Errorf("expected 2 unread, got %d", count)
	}

	// Mark one as read
	notifications := ns.GetUserNotifications("user-1", 10, 0)
	ns.MarkNotificationRead(notifications[0].ID)

	count = ns.GetUnreadCount("user-1")
	if count != 1 {
		t.Errorf("expected 1 unread after marking 1 read, got %d", count)
	}
}

func TestClearOldNotifications(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	ns.NotifyScoreChange("user-1", 50.0, 75.0, "score increase")

	// Manually age a notification
	notifications := ns.GetUserNotifications("user-1", 10, 0)
	notifications[0].CreatedAt = notifications[0].CreatedAt.AddDate(0, 0, -90)

	// Clear notifications older than 30 days
	cleared := ns.ClearOldNotifications(30)
	if cleared != 1 {
		t.Errorf("expected to clear 1 notification, cleared %d", cleared)
	}

	remaining := ns.GetUserNotifications("user-1", 10, 0)
	if len(remaining) != 0 {
		t.Errorf("expected 0 notifications after clearing old ones, got %d", len(remaining))
	}
}

func TestDisabledNotifications(t *testing.T) {
	mock := &MockNotificationProvider{}
	ns := NewNotificationService(mock)

	prefs := DefaultPreferences("user-1")
	prefs.ScoreChangeAlerts = false
	ns.SetUserPreferences(prefs)

	// Try to send disabled notification
	ns.NotifyScoreChange("user-1", 50.0, 75.0, "score increase")
	ns.SendPendingNotifications()

	if len(mock.emails) != 0 {
		t.Error("should not send disabled notification type")
	}
}
