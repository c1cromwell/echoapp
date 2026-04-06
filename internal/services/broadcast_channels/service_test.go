package broadcast_channels

import (
	"fmt"
	"testing"
)

// ========== MODEL TESTS ==========

func TestNewChannel(t *testing.T) {
	channel := NewChannel("Test Channel", "test", "creator-1", ChannelTypeNews)

	if channel.Name != "Test Channel" {
		t.Errorf("expected Test Channel, got %s", channel.Name)
	}
	if channel.ChannelType != ChannelTypeNews {
		t.Errorf("expected news type, got %s", channel.ChannelType)
	}
	if channel.IsActive != true {
		t.Error("new channel should be active")
	}
	if channel.SubscriberCount != 0 {
		t.Error("new channel should have 0 subscribers")
	}
}

func TestChannelTypeDefaults(t *testing.T) {
	tests := []struct {
		name           string
		channelType    ChannelType
		expectComments bool
		expectApproval bool
	}{
		{"news", ChannelTypeNews, true, false},
		{"announcement", ChannelTypeAnnouncement, false, false},
		{"educational", ChannelTypeEducational, true, true},
		{"community", ChannelTypeCommunity, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := NewChannel("Test", "test", "creator-1", tt.channelType)

			if channel.AllowComments != tt.expectComments {
				t.Errorf("comments: expected %v, got %v", tt.expectComments, channel.AllowComments)
			}
			if channel.RequireApproval != tt.expectApproval {
				t.Errorf("approval: expected %v, got %v", tt.expectApproval, channel.RequireApproval)
			}
		})
	}
}

func TestNewChannelPost(t *testing.T) {
	post := NewChannelPost("ch-1", "user-1", "Hello world", ContentTypeText)

	if post.Content != "Hello world" {
		t.Errorf("expected Hello world, got %s", post.Content)
	}
	if post.ContentType != ContentTypeText {
		t.Errorf("expected text type, got %s", post.ContentType)
	}
	if post.PublishStatus != PublishStatusDraft {
		t.Errorf("expected draft, got %s", post.PublishStatus)
	}
}

func TestNewChannelSubscriber(t *testing.T) {
	sub := NewChannelSubscriber("ch-1", "user-1")

	if sub.SubscriptionTier != SubscriptionTierFree {
		t.Error("should default to free tier")
	}
	if sub.Role != SubscriberRoleSubscriber {
		t.Error("should default to subscriber role")
	}
}

func TestCanPublish(t *testing.T) {
	channel := NewChannel("Test", "test", "creator-1", ChannelTypeNews)
	post := NewChannelPost("ch-1", "user-1", "Content", ContentTypeText)

	can, msg := channel.CanPublish(post)
	if !can {
		t.Errorf("should allow publish: %s", msg)
	}

	// Test empty content
	emptyPost := NewChannelPost("ch-1", "user-1", "", ContentTypeText)
	can, _ = channel.CanPublish(emptyPost)
	if can {
		t.Error("should reject empty post")
	}
}

// ========== SERVICE TESTS ==========

func TestCreateChannel(t *testing.T) {
	cs := NewChannelService()

	channel, err := cs.CreateChannel("Test Channel", "test", "creator-1", ChannelTypeNews)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if channel.Name != "Test Channel" {
		t.Errorf("expected Test Channel, got %s", channel.Name)
	}

	// Verify it's stored
	retrieved, err := cs.GetChannel(channel.ID)
	if err != nil {
		t.Fatalf("failed to retrieve channel: %v", err)
	}

	if retrieved.ID != channel.ID {
		t.Error("channel mismatch")
	}
}

func TestCreateChannelValidation(t *testing.T) {
	cs := NewChannelService()

	tests := []struct {
		name        string
		channelName string
		shouldFail  bool
	}{
		{"empty_name", "", true},
		{"valid_name", "Test Channel", false},
		{"long_name", "x" + fmt.Sprintf("%0101d", 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cs.CreateChannel(tt.channelName, "test", "creator-1", ChannelTypeNews)
			if (err != nil) != tt.shouldFail {
				t.Errorf("expected fail=%v, got error=%v", tt.shouldFail, err)
			}
		})
	}
}

func TestCreateChannelLimit(t *testing.T) {
	cs := NewChannelService()

	for i := 0; i < 20; i++ {
		_, err := cs.CreateChannel(fmt.Sprintf("Channel %d", i), "test", "creator-1", ChannelTypeNews)
		if err != nil {
			t.Fatalf("failed to create channel %d: %v", i, err)
		}
	}

	// 21st should fail
	_, err := cs.CreateChannel("Channel 21", "test", "creator-1", ChannelTypeNews)
	if err == nil {
		t.Error("should reject > 20 channels per creator")
	}
}

func TestUpdateChannel(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	updates := map[string]interface{}{
		"name": "Updated Name",
	}

	updated, err := cs.UpdateChannel(channel.ID, updates)
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected Updated Name, got %s", updated.Name)
	}
}

func TestDeleteChannel(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	err := cs.DeleteChannel(channel.ID)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = cs.GetChannel(channel.ID)
	if err == nil {
		t.Error("should not retrieve deleted channel")
	}
}

func TestListChannels(t *testing.T) {
	cs := NewChannelService()

	for i := 0; i < 5; i++ {
		cs.CreateChannel(fmt.Sprintf("Channel %d", i), "test", "creator-1", ChannelTypeNews)
	}

	channels := cs.ListChannels(10, 0)
	if len(channels) != 5 {
		t.Errorf("expected 5 channels, got %d", len(channels))
	}
}

// ========== POST TESTS ==========

func TestCreatePost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")

	post, err := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	if post.Content != "Hello" {
		t.Errorf("expected Hello, got %s", post.Content)
	}

	// Should auto-publish on news channel
	if post.PublishStatus != PublishStatusPublished {
		t.Errorf("expected published, got %s", post.PublishStatus)
	}
}

func TestCreatePostRequiresSubscription(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	// Non-subscriber can't post
	_, err := cs.CreatePost(channel.ID, "user-2", "Hello", ContentTypeText)
	if err == nil {
		t.Error("should require subscription to post")
	}
}

func TestPublishPost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeEducational)
	cs.Subscribe(channel.ID, "creator-1")

	post, _ := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)

	if post.PublishStatus != PublishStatusPending {
		t.Errorf("educational posts should be pending, got %s", post.PublishStatus)
	}

	published, err := cs.PublishPost(post.ID)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	if published.PublishStatus != PublishStatusPublished {
		t.Error("should be published after PublishPost")
	}
}

func TestDeletePost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")
	post, _ := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)

	err := cs.DeletePost(post.ID)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	retrieved, _ := cs.GetPost(post.ID)
	if retrieved.PublishStatus != PublishStatusDeleted {
		t.Error("post should be marked deleted")
	}
}

func TestGetChannelPosts(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")

	for i := 0; i < 5; i++ {
		cs.CreatePost(channel.ID, "creator-1", fmt.Sprintf("Post %d", i), ContentTypeText)
	}

	posts := cs.GetChannelPosts(channel.ID, 10, 0)
	if len(posts) != 5 {
		t.Errorf("expected 5 posts, got %d", len(posts))
	}
}

// ========== SUBSCRIBER TESTS ==========

func TestSubscribe(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	sub, err := cs.Subscribe(channel.ID, "user-1")
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	if sub.SubscriberID != "user-1" {
		t.Errorf("expected user-1, got %s", sub.SubscriberID)
	}

	if channel.SubscriberCount != 1 {
		t.Error("channel subscriber count should be 1")
	}
}

func TestSubscribeDuplicate(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	cs.Subscribe(channel.ID, "user-1")
	_, err := cs.Subscribe(channel.ID, "user-1")
	if err == nil {
		t.Error("should reject duplicate subscription")
	}
}

func TestUnsubscribe(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "user-1")

	err := cs.Unsubscribe(channel.ID, "user-1")
	if err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}

	if channel.SubscriberCount != 0 {
		t.Error("channel subscriber count should be 0")
	}
}

func TestGetSubscriber(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "user-1")

	sub, err := cs.GetSubscriber(channel.ID, "user-1")
	if err != nil {
		t.Fatalf("failed to get subscriber: %v", err)
	}

	if sub.SubscriberID != "user-1" {
		t.Error("subscriber mismatch")
	}
}

func TestUpdateSubscriberRole(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "user-1")

	err := cs.UpdateSubscriberRole(channel.ID, "user-1", SubscriberRoleModerator)
	if err != nil {
		t.Fatalf("failed to update role: %v", err)
	}

	sub, _ := cs.GetSubscriber(channel.ID, "user-1")
	if sub.Role != SubscriberRoleModerator {
		t.Errorf("expected moderator, got %s", sub.Role)
	}
}

func TestMuteSubscriber(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "user-1")

	err := cs.MuteSubscriber(channel.ID, "user-1")
	if err != nil {
		t.Fatalf("failed to mute: %v", err)
	}

	sub, _ := cs.GetSubscriber(channel.ID, "user-1")
	if !sub.IsMuted {
		t.Error("subscriber should be muted")
	}
}

func TestBlockSubscriber(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "user-1")

	err := cs.BlockSubscriber(channel.ID, "user-1")
	if err != nil {
		t.Fatalf("failed to block: %v", err)
	}

	sub, _ := cs.GetSubscriber(channel.ID, "user-1")
	if !sub.IsBlocked {
		t.Error("subscriber should be blocked")
	}
}

// ========== MODERATION TESTS ==========

func TestReportPost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")
	post, _ := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)

	mod, err := cs.ReportPost(channel.ID, post.ID, "reporter-1", ReasonSpam)
	if err != nil {
		t.Fatalf("failed to report: %v", err)
	}

	if mod.ReasonCode != ReasonSpam {
		t.Errorf("expected spam, got %s", mod.ReasonCode)
	}

	if post.FlagCount != 1 {
		t.Error("post flag count should be 1")
	}
}

func TestApprovePost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeEducational)
	cs.Subscribe(channel.ID, "creator-1")
	post, _ := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)

	err := cs.ApprovePost(post.ID)
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}

	approved, _ := cs.GetPost(post.ID)
	if approved.PublishStatus != PublishStatusPublished {
		t.Error("post should be published")
	}
}

func TestRejectPost(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeEducational)
	cs.Subscribe(channel.ID, "creator-1")
	post, _ := cs.CreatePost(channel.ID, "creator-1", "Hello", ContentTypeText)

	err := cs.RejectPost(post.ID, "Violates policy")
	if err != nil {
		t.Fatalf("failed to reject: %v", err)
	}

	rejected, _ := cs.GetPost(post.ID)
	if rejected.ModStatus != ModStatusRejected {
		t.Error("post should be rejected")
	}
}

// ========== ANALYTICS TESTS ==========

func TestRecordView(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	err := cs.RecordView(channel.ID)
	if err != nil {
		t.Fatalf("failed to record view: %v", err)
	}

	analytics, _ := cs.GetAnalytics(channel.ID, "daily")
	if analytics.ViewCount != 1 {
		t.Errorf("expected 1 view, got %d", analytics.ViewCount)
	}
}

func TestGetAnalytics(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	for i := 0; i < 5; i++ {
		cs.RecordView(channel.ID)
	}

	analytics, err := cs.GetAnalytics(channel.ID, "daily")
	if err != nil {
		t.Fatalf("failed to get analytics: %v", err)
	}

	if analytics.ViewCount != 5 {
		t.Errorf("expected 5 views, got %d", analytics.ViewCount)
	}
}

// ========== SEARCH & DISCOVERY TESTS ==========

func TestSearchChannels(t *testing.T) {
	cs := NewChannelService()
	cs.CreateChannel("Breaking News", "news", "creator-1", ChannelTypeNews)
	cs.CreateChannel("Product Launch", "announcement", "creator-1", ChannelTypeAnnouncement)

	results := cs.SearchChannels("news", 10)
	if len(results) < 1 {
		t.Error("should find Breaking News")
	}

	results = cs.SearchChannels("launch", 10)
	if len(results) < 1 {
		t.Error("should find Product Launch")
	}
}

func TestSearchPosts(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")

	cs.CreatePost(channel.ID, "creator-1", "Go is awesome", ContentTypeText)
	cs.CreatePost(channel.ID, "creator-1", "Rust is great", ContentTypeText)

	results := cs.SearchPosts(channel.ID, "Go", 10)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestGetTrendingChannels(t *testing.T) {
	cs := NewChannelService()

	ch1, _ := cs.CreateChannel("Channel 1", "test", "creator-1", ChannelTypeNews)
	ch2, _ := cs.CreateChannel("Channel 2", "test", "creator-1", ChannelTypeNews)

	// Give ch2 higher trust score
	cs.UpdateChannel(ch2.ID, map[string]interface{}{"trust_score": 80.0})
	cs.UpdateChannel(ch1.ID, map[string]interface{}{"trust_score": 30.0})

	trending := cs.GetTrendingChannels(10)
	if len(trending) < 2 {
		t.Error("should return trending channels")
	}

	if trending[0].ID != ch2.ID {
		t.Error("highest trust score should be first")
	}
}

// ========== INTEGRATION TESTS ==========

func TestFullChannelWorkflow(t *testing.T) {
	cs := NewChannelService()

	// 1. Create channel
	channel, err := cs.CreateChannel("News Channel", "news", "creator-1", ChannelTypeNews)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 2. Subscribe users
	cs.Subscribe(channel.ID, "user-1")
	cs.Subscribe(channel.ID, "user-2")

	// 3. Post content
	post, err := cs.CreatePost(channel.ID, "user-1", "Breaking news!", ContentTypeText)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	// 4. Get posts
	posts := cs.GetChannelPosts(channel.ID, 10, 0)
	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}

	// 5. Report content
	cs.ReportPost(channel.ID, post.ID, "user-2", ReasonMisinformation)
	post, _ = cs.GetPost(post.ID)
	if post.FlagCount != 1 {
		t.Error("report should increment flag count")
	}

	// 6. Delete post
	cs.DeletePost(post.ID)
	post, _ = cs.GetPost(post.ID)
	if post.PublishStatus != PublishStatusDeleted {
		t.Error("post should be deleted")
	}

	// 7. Unsubscribe
	cs.Unsubscribe(channel.ID, "user-1")
	if channel.SubscriberCount != 1 {
		t.Error("subscriber count should decrease")
	}
}

func TestEducationalChannelApprovalFlow(t *testing.T) {
	cs := NewChannelService()

	// Create educational channel (requires approval)
	channel, _ := cs.CreateChannel("Go Course", "education", "instructor-1", ChannelTypeEducational)
	cs.Subscribe(channel.ID, "instructor-1")

	// Create post (should be pending)
	post, _ := cs.CreatePost(channel.ID, "instructor-1", "Lesson 1: Introduction", ContentTypeText)
	if post.PublishStatus != PublishStatusPending {
		t.Error("educational posts should require approval")
	}

	// Approve post
	cs.ApprovePost(post.ID)
	post, _ = cs.GetPost(post.ID)
	if post.PublishStatus != PublishStatusPublished {
		t.Error("post should be published after approval")
	}
}

func TestChannelWithMultipleSubscribers(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Community", "discussion", "creator-1", ChannelTypeCommunity)

	// Multiple subscribers
	for i := 1; i <= 5; i++ {
		cs.Subscribe(channel.ID, fmt.Sprintf("user-%d", i))
	}

	subs := cs.GetChannelSubscribers(channel.ID)
	if len(subs) != 5 {
		t.Errorf("expected 5 subscribers, got %d", len(subs))
	}

	if channel.SubscriberCount != 5 {
		t.Errorf("channel count should be 5, got %d", channel.SubscriberCount)
	}
}

// ========== CONCURRENCY TESTS ==========

func TestConcurrentSubscriptions(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			userID := fmt.Sprintf("user-%d", index)
			cs.Subscribe(channel.ID, userID)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	subs := cs.GetChannelSubscribers(channel.ID)
	if len(subs) != 10 {
		t.Errorf("expected 10 subscribers, got %d", len(subs))
	}
}

func TestConcurrentPostCreation(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)

	for i := 1; i <= 5; i++ {
		cs.Subscribe(channel.ID, fmt.Sprintf("user-%d", i))
	}

	done := make(chan bool)

	for i := 0; i < 20; i++ {
		go func(index int) {
			userID := fmt.Sprintf("user-%d", index%5+1)
			cs.CreatePost(channel.ID, userID, fmt.Sprintf("Post %d", index), ContentTypeText)
			done <- true
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	posts := cs.GetChannelPosts(channel.ID, 100, 0)
	if len(posts) != 20 {
		t.Errorf("expected 20 posts, got %d", len(posts))
	}
}

// ========== EDGE CASES & VALIDATION ==========

func TestEdgeCases(t *testing.T) {
	cs := NewChannelService()

	// Non-existent channel
	_, err := cs.GetChannel("nonexistent")
	if err == nil {
		t.Error("should error on nonexistent channel")
	}

	// Non-existent post
	_, err = cs.GetPost("nonexistent")
	if err == nil {
		t.Error("should error on nonexistent post")
	}

	// Subscribe to nonexistent channel
	_, err = cs.Subscribe("nonexistent", "user-1")
	if err == nil {
		t.Error("should error subscribing to nonexistent channel")
	}
}

func TestEmptyLists(t *testing.T) {
	cs := NewChannelService()

	channels := cs.ListChannels(10, 0)
	if len(channels) != 0 {
		t.Error("should return empty list when no channels")
	}

	posts := cs.GetChannelPosts("nonexistent", 10, 0)
	if len(posts) != 0 {
		t.Error("should return empty list for nonexistent channel")
	}
}

func TestPagination(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")

	// Create 25 posts
	for i := 0; i < 25; i++ {
		cs.CreatePost(channel.ID, "creator-1", fmt.Sprintf("Post %d", i), ContentTypeText)
	}

	// Test pagination
	page1 := cs.GetChannelPosts(channel.ID, 10, 0)
	if len(page1) != 10 {
		t.Errorf("page 1: expected 10, got %d", len(page1))
	}

	page2 := cs.GetChannelPosts(channel.ID, 10, 10)
	if len(page2) != 10 {
		t.Errorf("page 2: expected 10, got %d", len(page2))
	}

	page3 := cs.GetChannelPosts(channel.ID, 10, 20)
	if len(page3) != 5 {
		t.Errorf("page 3: expected 5, got %d", len(page3))
	}
}

func TestModerationFlagAutoHide(t *testing.T) {
	cs := NewChannelService()
	channel, _ := cs.CreateChannel("Test", "test", "creator-1", ChannelTypeNews)
	cs.Subscribe(channel.ID, "creator-1")
	post, _ := cs.CreatePost(channel.ID, "creator-1", "Suspicious content", ContentTypeText)

	// Report 3 times
	cs.ReportPost(channel.ID, post.ID, "reporter-1", ReasonSpam)
	cs.ReportPost(channel.ID, post.ID, "reporter-2", ReasonSpam)
	cs.ReportPost(channel.ID, post.ID, "reporter-3", ReasonSpam)

	// Should auto-hide at 3 flags
	post, _ = cs.GetPost(post.ID)
	if post.PublishStatus != PublishStatusArchived {
		t.Errorf("expected archived after 3 flags, got %s", post.PublishStatus)
	}
}

func TestGetCreatorChannels(t *testing.T) {
	cs := NewChannelService()

	for i := 0; i < 3; i++ {
		cs.CreateChannel(fmt.Sprintf("Channel %d", i), "test", "creator-1", ChannelTypeNews)
	}

	for i := 0; i < 2; i++ {
		cs.CreateChannel(fmt.Sprintf("Other %d", i), "test", "creator-2", ChannelTypeNews)
	}

	channels := cs.GetCreatorChannels("creator-1")
	if len(channels) != 3 {
		t.Errorf("expected 3 channels for creator-1, got %d", len(channels))
	}
}

func TestTrustScoreOrdering(t *testing.T) {
	score1 := CalculateTrustScore(VerificationUnverified, 10.0, 15.0, 8.0)
	score2 := CalculateTrustScore(VerificationVerified, 10.0, 15.0, 8.0)
	score3 := CalculateTrustScore(VerificationPremium, 10.0, 15.0, 8.0)

	if score1 >= score2 {
		t.Error("verified should have higher score than unverified")
	}

	if score2 >= score3 {
		t.Error("premium should have highest score")
	}

	if score3 > 100 {
		t.Error("score should not exceed 100")
	}
}
