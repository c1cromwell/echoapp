package broadcast_channels

import (
	"math/rand"
	"time"
)

// ========== ENUMS ==========

type VisibilityMode string

const (
	VisibilityPublic      VisibilityMode = "public"       // Discoverable, anyone can join
	VisibilitySemiPrivate VisibilityMode = "semi_private" // Discoverable, requires approval
	VisibilityPrivate     VisibilityMode = "private"      // Invite-only, hidden from search
)

type ChannelType string

const (
	ChannelTypeNews         ChannelType = "news"         // Media/news organization
	ChannelTypeAnnouncement ChannelType = "announcement" // Business announcements
	ChannelTypeEducational  ChannelType = "educational"  // Course/tutorial content
	ChannelTypeCommunity    ChannelType = "community"    // Community discussion
)

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypePoll  ContentType = "poll"
	ContentTypeFile  ContentType = "file"
	ContentTypeMixed ContentType = "mixed"
)

type PublishStatus string

const (
	PublishStatusDraft     PublishStatus = "draft"
	PublishStatusPending   PublishStatus = "pending"
	PublishStatusScheduled PublishStatus = "scheduled"
	PublishStatusPublished PublishStatus = "published"
	PublishStatusArchived  PublishStatus = "archived"
	PublishStatusDeleted   PublishStatus = "deleted"
)

type SubscriptionTier string

const (
	SubscriptionTierFree   SubscriptionTier = "free"
	SubscriptionTierBronze SubscriptionTier = "bronze"
	SubscriptionTierSilver SubscriptionTier = "silver"
	SubscriptionTierGold   SubscriptionTier = "gold"
)

type SubscriberRole string

const (
	SubscriberRoleSubscriber SubscriberRole = "subscriber"
	SubscriberRoleModerator  SubscriberRole = "moderator"
	SubscriberRoleAdmin      SubscriberRole = "admin"
)

type NotificationMode string

const (
	NotificationModeAll       NotificationMode = "all"
	NotificationModeImportant NotificationMode = "important"
	NotificationModeNone      NotificationMode = "none"
)

type ModStatus string

const (
	ModStatusApproved ModStatus = "approved"
	ModStatusPending  ModStatus = "pending"
	ModStatusRejected ModStatus = "rejected"
	ModStatusAppealed ModStatus = "appealed"
)

type ActionType string

const (
	ActionTypeDelete  ActionType = "delete"
	ActionTypeHide    ActionType = "hide"
	ActionTypeMute    ActionType = "mute"
	ActionTypeBlock   ActionType = "block"
	ActionTypeWarn    ActionType = "warn"
	ActionTypeRestore ActionType = "restore"
)

type ReasonCode string

const (
	ReasonSpam           ReasonCode = "spam"
	ReasonAbuse          ReasonCode = "abuse"
	ReasonMisinformation ReasonCode = "misinformation"
	ReasonAdult          ReasonCode = "adult"
	ReasonManipulation   ReasonCode = "manipulation"
	ReasonOther          ReasonCode = "other"
)

type VerificationStatus string

const (
	VerificationUnverified VerificationStatus = "unverified"
	VerificationVerified   VerificationStatus = "verified"
	VerificationPremium    VerificationStatus = "premium"
)

// ========== CORE MODELS ==========

type Channel struct {
	// Identity
	ID          string `json:"id"`
	CreatorID   string `json:"creator_id"`
	Name        string `json:"name"`
	Topic       string `json:"topic"`
	Description string `json:"description"`

	// Configuration
	VisibilityMode VisibilityMode `json:"visibility_mode"`
	ChannelType    ChannelType    `json:"channel_type"`
	CoverImageURL  string         `json:"cover_image_url,omitempty"`
	Tags           []string       `json:"tags"`

	// State
	IsActive   bool       `json:"is_active"`
	IsMuted    bool       `json:"is_muted"`
	CreatedAt  time.Time  `json:"created_at"`
	LastPostAt *time.Time `json:"last_post_at,omitempty"`

	// Counts
	SubscriberCount int64 `json:"subscriber_count"`
	TotalPostCount  int64 `json:"total_post_count"`

	// Metadata
	Language           string             `json:"language"`
	Website            string             `json:"website,omitempty"`
	TrustScore         float64            `json:"trust_score"`
	VerificationStatus VerificationStatus `json:"verification_status"`

	// Configuration flags
	AllowComments   bool `json:"allow_comments"`
	AllowPolls      bool `json:"allow_polls"`
	AllowLinks      bool `json:"allow_links"`
	AllowMedia      bool `json:"allow_media"`
	RequireApproval bool `json:"require_approval"`
}

type ChannelPost struct {
	// Identity
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	CreatorID string `json:"creator_id"`

	// Content
	Content          string      `json:"content"`
	ContentType      ContentType `json:"content_type"`
	EncryptedContent []byte      `json:"encrypted_content"`

	// Engagement (denormalized)
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	ShareCount   int64 `json:"share_count"`

	// Status
	PublishedAt   *time.Time    `json:"published_at,omitempty"`
	ScheduledFor  *time.Time    `json:"scheduled_for,omitempty"`
	PublishStatus PublishStatus `json:"publish_status"`

	// Flags
	IsPinned     bool `json:"is_pinned"`
	IsFeatured   bool `json:"is_featured"`
	AllowReplies bool `json:"allow_replies"`
	IsSponsored  bool `json:"is_sponsored"`

	// Moderation
	FlagCount int       `json:"flag_count"`
	ModStatus ModStatus `json:"mod_status"`
	ModNotes  string    `json:"mod_notes,omitempty"`

	// Edit tracking
	EditCount int       `json:"edit_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChannelSubscriber struct {
	// Identity
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	SubscriberID string `json:"subscriber_id"`

	// Subscription
	JoinedAt              time.Time        `json:"joined_at"`
	SubscriptionTier      SubscriptionTier `json:"subscription_tier"`
	SubscriptionExpiresAt *time.Time       `json:"subscription_expires_at,omitempty"`
	AutoRenew             bool             `json:"auto_renew"`

	// Permissions
	Role SubscriberRole `json:"role"`

	// Engagement
	LastSeenAt       time.Time        `json:"last_seen_at"`
	NotificationMode NotificationMode `json:"notification_mode"`
	IsMuted          bool             `json:"is_muted"`
	IsBlocked        bool             `json:"is_blocked"`

	// Activity
	PostCount    int64 `json:"post_count"`
	CommentCount int64 `json:"comment_count"`
	LikeCount    int64 `json:"like_count"`

	// Trust
	TrustScore      float64 `json:"trust_score"`
	ModerationFlags int     `json:"moderation_flags"`
}

type ModerationAction struct {
	ID            string     `json:"id"`
	ChannelID     string     `json:"channel_id"`
	TargetID      string     `json:"target_id"`   // Post ID or Subscriber ID
	TargetType    string     `json:"target_type"` // "post" | "subscriber"
	ActionType    ActionType `json:"action_type"`
	ActionedByID  string     `json:"actioned_by_id"`
	Reason        string     `json:"reason"`
	ReasonCode    ReasonCode `json:"reason_code"`
	Evidence      []string   `json:"evidence"`
	AppealCount   int        `json:"appeal_count"`
	FinalStatus   bool       `json:"final_status"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ReversedAt    *time.Time `json:"reversed_at,omitempty"`
	ReverseReason string     `json:"reverse_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ChannelAnalytics struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channel_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	PeriodType  string    `json:"period_type"` // "daily" | "weekly" | "monthly"

	// Subscriber metrics
	TotalSubscribers int64 `json:"total_subscribers"`
	NewSubscribers   int64 `json:"new_subscribers"`
	UnsubscribeCount int64 `json:"unsubscribe_count"`

	// Engagement
	ViewCount         int64   `json:"view_count"`
	ClickCount        int64   `json:"click_count"`
	TotalLikes        int64   `json:"total_likes"`
	TotalComments     int64   `json:"total_comments"`
	TotalShares       int64   `json:"total_shares"`
	AverageEngagement float64 `json:"average_engagement"`
	PostCount         int64   `json:"post_count"`

	// Posts
	TopPostID         string `json:"top_post_id,omitempty"`
	TopPostEngagement int64  `json:"top_post_engagement"`

	// Monetization
	SubscriptionRevenue int64     `json:"subscription_revenue"` // In cents
	DonationRevenue     int64     `json:"donation_revenue"`
	TotalRevenue        int64     `json:"total_revenue"`
	CreatedAt           time.Time `json:"created_at"`
}

// ========== HELPER FUNCTIONS ==========

func NewChannel(name, topic string, creatorID string, channelType ChannelType) *Channel {
	now := time.Now()

	channel := &Channel{
		// Identity
		ID:        generateChannelID(),
		CreatorID: creatorID,
		Name:      name,
		Topic:     topic,

		// Configuration
		VisibilityMode: VisibilityPublic,
		ChannelType:    channelType,
		Tags:           []string{},

		// State
		IsActive:  true,
		IsMuted:   false,
		CreatedAt: now,

		// Counts
		SubscriberCount: 0,
		TotalPostCount:  0,

		// Metadata
		Language:           "en",
		TrustScore:         50.0, // Starting score
		VerificationStatus: VerificationUnverified,

		// Configuration flags
		AllowComments:   true,
		AllowPolls:      true,
		AllowLinks:      true,
		AllowMedia:      true,
		RequireApproval: false,
	}

	// Apply type-specific defaults
	switch channelType {
	case ChannelTypeNews:
		channel.AllowComments = true
		channel.AllowPolls = true
		channel.AllowMedia = true
		channel.RequireApproval = false
	case ChannelTypeAnnouncement:
		channel.AllowComments = false
		channel.AllowPolls = false
		channel.AllowMedia = true
		channel.RequireApproval = false
	case ChannelTypeEducational:
		channel.AllowComments = true
		channel.AllowPolls = true
		channel.AllowMedia = true
		channel.RequireApproval = true
	case ChannelTypeCommunity:
		channel.AllowComments = true
		channel.AllowPolls = true
		channel.AllowMedia = true
		channel.RequireApproval = false
	}

	return channel
}

func NewChannelPost(channelID, creatorID, content string, contentType ContentType) *ChannelPost {
	now := time.Now()

	return &ChannelPost{
		ID:            generatePostID(),
		ChannelID:     channelID,
		CreatorID:     creatorID,
		Content:       content,
		ContentType:   contentType,
		PublishStatus: PublishStatusDraft,
		AllowReplies:  true,
		ModStatus:     ModStatusApproved,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func NewChannelSubscriber(channelID, subscriberID string) *ChannelSubscriber {
	now := time.Now()

	return &ChannelSubscriber{
		ID:               generateSubscriberID(),
		ChannelID:        channelID,
		SubscriberID:     subscriberID,
		JoinedAt:         now,
		SubscriptionTier: SubscriptionTierFree,
		Role:             SubscriberRoleSubscriber,
		LastSeenAt:       now,
		NotificationMode: NotificationModeAll,
		IsMuted:          false,
		IsBlocked:        false,
		TrustScore:       50.0,
	}
}

func NewModerationAction(channelID, targetID, targetType string, actionType ActionType, actionedByID string) *ModerationAction {
	return &ModerationAction{
		ID:           generateActionID(),
		ChannelID:    channelID,
		TargetID:     targetID,
		TargetType:   targetType,
		ActionType:   actionType,
		ActionedByID: actionedByID,
		CreatedAt:    time.Now(),
		FinalStatus:  false,
		AppealCount:  0,
	}
}

func NewChannelAnalytics(channelID string, periodType string) *ChannelAnalytics {
	now := time.Now()
	periodStart := getPeriodStart(now, periodType)

	return &ChannelAnalytics{
		ID:          generateAnalyticsID(),
		ChannelID:   channelID,
		PeriodStart: periodStart,
		PeriodEnd:   now,
		PeriodType:  periodType,
		CreatedAt:   now,
	}
}

// Helper functions for ID generation

func generateChannelID() string {
	return "ch_" + randString(16)
}

func generatePostID() string {
	return "post_" + randString(16)
}

func generateSubscriberID() string {
	return "sub_" + randString(16)
}

func generateActionID() string {
	return "act_" + randString(16)
}

func generateAnalyticsID() string {
	return "ana_" + randString(16)
}

func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seeded := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seeded.Intn(len(charset))]
	}
	return string(result)
}

func getPeriodStart(t time.Time, periodType string) time.Time {
	switch periodType {
	case "daily":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "weekly":
		// Return start of week (Monday)
		daysToMonday := int(t.Weekday())
		if daysToMonday == 0 {
			daysToMonday = 7
		}
		return t.AddDate(0, 0, -daysToMonday+1)
	case "monthly":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// CalculateTrustScore computes channel trust score
func CalculateTrustScore(verificationStatus VerificationStatus, postConsistency float64, engagementRate float64, subscriberGrowth float64) float64 {
	score := 0.0

	// Verification status: 0-30 points
	switch verificationStatus {
	case VerificationPremium:
		score += 30
	case VerificationVerified:
		score += 20
	case VerificationUnverified:
		score += 0
	}

	// Post consistency: 0-20 points
	score += min(postConsistency, 20)

	// Engagement rate: 0-20 points
	score += min(engagementRate, 20)

	// Subscriber growth: 0-15 points
	score += min(subscriberGrowth, 15)

	// Content quality: 0-15 points (would need more detailed analysis)
	score += 0 // Placeholder

	return min(score, 100) // Cap at 100
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// CanPublish checks if post can be published based on channel settings
func (c *Channel) CanPublish(post *ChannelPost) (bool, string) {
	if !c.IsActive {
		return false, "channel is inactive"
	}

	if c.IsMuted {
		return false, "channel is muted"
	}

	if post.Content == "" {
		return false, "post cannot be empty"
	}

	if len(post.Content) > 5000 {
		return false, "post exceeds 5000 character limit"
	}

	// Check content type permissions
	switch post.ContentType {
	case ContentTypeImage, ContentTypeVideo, ContentTypeFile, ContentTypeMixed:
		if !c.AllowMedia {
			return false, "channel does not allow media"
		}
	case ContentTypePoll:
		if !c.AllowPolls {
			return false, "channel does not allow polls"
		}
	}

	return true, ""
}

// CanSubscribe checks if subscriber can join channel
func (c *Channel) CanSubscribe() (bool, string) {
	if !c.IsActive {
		return false, "channel is inactive"
	}

	return true, ""
}
