package broadcast_channels

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ChannelService manages broadcast channels
type ChannelService struct {
	mu sync.RWMutex

	// Core stores
	channels         map[string]*Channel
	posts            map[string]*ChannelPost
	subscribers      map[string]*ChannelSubscriber
	postIndex        map[string][]string // channelID -> postIDs
	moderations      map[string]*ModerationAction
	analytics        map[string]*ChannelAnalytics
	subscriberLookup map[string]map[string]*ChannelSubscriber // channelID -> subscriberID -> subscriber
	creatorChannels  map[string][]string                      // creatorID -> channelIDs
}

// NewChannelService creates a new channel service
func NewChannelService() *ChannelService {
	return &ChannelService{
		channels:         make(map[string]*Channel),
		posts:            make(map[string]*ChannelPost),
		subscribers:      make(map[string]*ChannelSubscriber),
		postIndex:        make(map[string][]string),
		moderations:      make(map[string]*ModerationAction),
		analytics:        make(map[string]*ChannelAnalytics),
		subscriberLookup: make(map[string]map[string]*ChannelSubscriber),
		creatorChannels:  make(map[string][]string),
	}
}

// ========== CHANNEL OPERATIONS ==========

func (cs *ChannelService) CreateChannel(name, topic string, creatorID string, channelType ChannelType) (*Channel, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Validate name
	if name == "" || len(name) > 100 {
		return nil, fmt.Errorf("channel name must be 1-100 characters")
	}

	// Check creator limit (max 20 channels per creator)
	if len(cs.creatorChannels[creatorID]) >= 20 {
		return nil, fmt.Errorf("creator has reached maximum of 20 channels")
	}

	channel := NewChannel(name, topic, creatorID, channelType)
	cs.channels[channel.ID] = channel
	cs.creatorChannels[creatorID] = append(cs.creatorChannels[creatorID], channel.ID)
	cs.subscriberLookup[channel.ID] = make(map[string]*ChannelSubscriber)
	cs.postIndex[channel.ID] = []string{}

	return channel, nil
}

func (cs *ChannelService) GetChannel(channelID string) (*Channel, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found")
	}
	return channel, nil
}

func (cs *ChannelService) UpdateChannel(channelID string, updates map[string]interface{}) (*Channel, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found")
	}

	// Apply updates
	if name, ok := updates["name"]; ok {
		channel.Name = name.(string)
	}
	if visibility, ok := updates["visibility_mode"]; ok {
		channel.VisibilityMode = visibility.(VisibilityMode)
	}
	if trustScore, ok := updates["trust_score"]; ok {
		channel.TrustScore = trustScore.(float64)
	}

	return channel, nil
}

func (cs *ChannelService) DeleteChannel(channelID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found")
	}

	// Cascade delete posts and subscribers
	if postIDs, ok := cs.postIndex[channelID]; ok {
		for _, postID := range postIDs {
			delete(cs.posts, postID)
		}
		delete(cs.postIndex, channelID)
	}

	delete(cs.subscriberLookup[channelID], "subscribers")

	// Remove from creator's channels
	for i, id := range cs.creatorChannels[channel.CreatorID] {
		if id == channelID {
			cs.creatorChannels[channel.CreatorID] = append(
				cs.creatorChannels[channel.CreatorID][:i],
				cs.creatorChannels[channel.CreatorID][i+1:]...,
			)
			break
		}
	}

	delete(cs.channels, channelID)
	return nil
}

func (cs *ChannelService) ListChannels(limit int, offset int) []*Channel {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]*Channel, 0)
	count := 0

	for _, channel := range cs.channels {
		if channel.VisibilityMode == VisibilityPublic && channel.IsActive {
			if count >= offset && count < offset+limit {
				result = append(result, channel)
			}
			count++
		}
	}

	return result
}

func (cs *ChannelService) GetCreatorChannels(creatorID string) []*Channel {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	channelIDs := cs.creatorChannels[creatorID]
	result := make([]*Channel, 0, len(channelIDs))

	for _, channelID := range channelIDs {
		if channel, exists := cs.channels[channelID]; exists {
			result = append(result, channel)
		}
	}

	return result
}

// ========== POST OPERATIONS ==========

func (cs *ChannelService) CreatePost(channelID, creatorID, content string, contentType ContentType) (*ChannelPost, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found")
	}

	if !channel.IsActive {
		return nil, fmt.Errorf("channel is inactive")
	}

	// Check if creator is subscriber
	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return nil, fmt.Errorf("channel subscribers not initialized")
	}

	subscriber, exists := subscribers[creatorID]
	if !exists {
		return nil, fmt.Errorf("creator must be subscribed to post")
	}

	if subscriber.IsBlocked {
		return nil, fmt.Errorf("creator is blocked from posting")
	}

	post := NewChannelPost(channelID, creatorID, content, contentType)

	// Auto-publish based on channel type
	if !channel.RequireApproval {
		post.PublishStatus = PublishStatusPublished
		post.PublishedAt = &time.Time{}
		*post.PublishedAt = time.Now()
		channel.TotalPostCount++
	} else {
		post.PublishStatus = PublishStatusPending
	}

	cs.posts[post.ID] = post
	cs.postIndex[channelID] = append(cs.postIndex[channelID], post.ID)

	return post, nil
}

func (cs *ChannelService) GetPost(postID string) (*ChannelPost, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	post, exists := cs.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (cs *ChannelService) PublishPost(postID string) (*ChannelPost, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	post, exists := cs.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	post.PublishStatus = PublishStatusPublished
	now := time.Now()
	post.PublishedAt = &now

	// Update channel post count
	if channel, ok := cs.channels[post.ChannelID]; ok {
		channel.TotalPostCount++
	}

	return post, nil
}

func (cs *ChannelService) DeletePost(postID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	post, exists := cs.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.PublishStatus = PublishStatusDeleted
	return nil
}

func (cs *ChannelService) GetChannelPosts(channelID string, limit int, offset int) []*ChannelPost {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	postIDs, exists := cs.postIndex[channelID]
	if !exists {
		return []*ChannelPost{}
	}

	result := make([]*ChannelPost, 0)
	count := 0

	// Reverse iteration to get newest first
	for i := len(postIDs) - 1; i >= 0; i-- {
		postID := postIDs[i]
		if post, ok := cs.posts[postID]; ok && post.PublishStatus == PublishStatusPublished {
			if count >= offset && count < offset+limit {
				result = append(result, post)
			}
			count++
		}
	}

	return result
}

// ========== SUBSCRIBER OPERATIONS ==========

func (cs *ChannelService) Subscribe(channelID, subscriberID string) (*ChannelSubscriber, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found")
	}

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		subscribers = make(map[string]*ChannelSubscriber)
		cs.subscriberLookup[channelID] = subscribers
	}

	// Check for duplicate
	if _, exists := subscribers[subscriberID]; exists {
		return nil, fmt.Errorf("already subscribed")
	}

	sub := NewChannelSubscriber(channelID, subscriberID)
	subscribers[subscriberID] = sub
	channel.SubscriberCount++

	return sub, nil
}

func (cs *ChannelService) Unsubscribe(channelID, subscriberID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found")
	}

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return fmt.Errorf("no subscribers")
	}

	if _, exists := subscribers[subscriberID]; !exists {
		return fmt.Errorf("not subscribed")
	}

	delete(subscribers, subscriberID)
	channel.SubscriberCount--

	return nil
}

func (cs *ChannelService) GetSubscriber(channelID, subscriberID string) (*ChannelSubscriber, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return nil, fmt.Errorf("channel not found")
	}

	sub, exists := subscribers[subscriberID]
	if !exists {
		return nil, fmt.Errorf("subscriber not found")
	}

	return sub, nil
}

func (cs *ChannelService) UpdateSubscriberRole(channelID, subscriberID string, role SubscriberRole) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return fmt.Errorf("channel not found")
	}

	sub, exists := subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber not found")
	}

	sub.Role = role
	return nil
}

func (cs *ChannelService) GetChannelSubscribers(channelID string) []*ChannelSubscriber {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return []*ChannelSubscriber{}
	}

	result := make([]*ChannelSubscriber, 0, len(subscribers))
	for _, sub := range subscribers {
		result = append(result, sub)
	}

	return result
}

func (cs *ChannelService) MuteSubscriber(channelID, subscriberID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return fmt.Errorf("channel not found")
	}

	sub, exists := subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber not found")
	}

	sub.IsMuted = true
	return nil
}

func (cs *ChannelService) UnmuteSubscriber(channelID, subscriberID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return fmt.Errorf("channel not found")
	}

	sub, exists := subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber not found")
	}

	sub.IsMuted = false
	return nil
}

func (cs *ChannelService) BlockSubscriber(channelID, subscriberID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	subscribers, ok := cs.subscriberLookup[channelID]
	if !ok {
		return fmt.Errorf("channel not found")
	}

	sub, exists := subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber not found")
	}

	sub.IsBlocked = true
	return nil
}

// ========== MODERATION OPERATIONS ==========

func (cs *ChannelService) ReportPost(channelID, postID, reporterID string, reason ReasonCode) (*ModerationAction, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	post, exists := cs.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	post.FlagCount++

	// Auto-hide at 3+ flags
	if post.FlagCount >= 3 {
		post.PublishStatus = PublishStatusArchived
	}

	action := NewModerationAction(channelID, postID, "post", ActionTypeHide, reporterID)
	action.ReasonCode = reason
	cs.moderations[action.ID] = action

	return action, nil
}

func (cs *ChannelService) ApprovePost(postID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	post, exists := cs.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.PublishStatus = PublishStatusPublished
	post.ModStatus = ModStatusApproved
	now := time.Now()
	post.PublishedAt = &now

	// Update channel post count
	if channel, ok := cs.channels[post.ChannelID]; ok {
		channel.TotalPostCount++
	}

	return nil
}

func (cs *ChannelService) RejectPost(postID string, reason string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	post, exists := cs.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.PublishStatus = PublishStatusArchived
	post.ModStatus = ModStatusRejected
	post.ModNotes = reason

	return nil
}

// ========== ANALYTICS OPERATIONS ==========

func (cs *ChannelService) RecordView(channelID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	channel, exists := cs.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found")
	}

	// Get or create daily analytics
	key := channelID + ":daily"
	analytics, exists := cs.analytics[key]
	if !exists {
		analytics = NewChannelAnalytics(channelID, "daily")
		cs.analytics[key] = analytics
	}

	analytics.ViewCount++

	// Update trust score based on views
	channel.TrustScore = CalculateTrustScore(
		channel.VerificationStatus,
		float64(channel.TotalPostCount),
		float64(analytics.ViewCount)/100.0,
		float64(channel.SubscriberCount)/100.0,
	)

	return nil
}

func (cs *ChannelService) GetAnalytics(channelID string, periodType string) (*ChannelAnalytics, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	key := channelID + ":" + periodType
	analytics, exists := cs.analytics[key]
	if !exists {
		// Return empty analytics if none exist
		return NewChannelAnalytics(channelID, periodType), nil
	}

	return analytics, nil
}

// ========== SEARCH & DISCOVERY ==========

func (cs *ChannelService) SearchChannels(query string, limit int) []*Channel {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]*Channel, 0)
	query = strings.ToLower(query)

	for _, channel := range cs.channels {
		if !channel.IsActive {
			continue
		}

		// Skip private channels
		if channel.VisibilityMode == VisibilityPrivate {
			continue
		}

		// Search in name, topic, description
		if strings.Contains(strings.ToLower(channel.Name), query) ||
			strings.Contains(strings.ToLower(channel.Topic), query) ||
			strings.Contains(strings.ToLower(channel.Description), query) {
			result = append(result, channel)
		}

		if len(result) >= limit {
			break
		}
	}

	return result
}

func (cs *ChannelService) SearchPosts(channelID string, query string, limit int) []*ChannelPost {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]*ChannelPost, 0)
	query = strings.ToLower(query)

	postIDs, exists := cs.postIndex[channelID]
	if !exists {
		return result
	}

	for _, postID := range postIDs {
		if post, ok := cs.posts[postID]; ok {
			if post.PublishStatus == PublishStatusPublished &&
				strings.Contains(strings.ToLower(post.Content), query) {
				result = append(result, post)
			}

			if len(result) >= limit {
				break
			}
		}
	}

	return result
}

func (cs *ChannelService) GetTrendingChannels(limit int) []*Channel {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Get active public channels
	channels := make([]*Channel, 0)
	for _, channel := range cs.channels {
		if channel.IsActive && channel.VisibilityMode == VisibilityPublic {
			channels = append(channels, channel)
		}
	}

	// Sort by trust score descending
	for i := 0; i < len(channels); i++ {
		for j := i + 1; j < len(channels); j++ {
			if channels[j].TrustScore > channels[i].TrustScore {
				channels[i], channels[j] = channels[j], channels[i]
			}
		}
	}

	// Return top N
	if len(channels) > limit {
		return channels[:limit]
	}

	return channels
}
