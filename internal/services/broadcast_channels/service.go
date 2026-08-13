package broadcast_channels

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ChannelService manages broadcast channels
type ChannelService struct {
	mu sync.RWMutex

	store ChannelStore

	// Core stores
	channels         map[string]*Channel
	posts            map[string]*ChannelPost
	subscribers      map[string]*ChannelSubscriber
	postIndex        map[string][]string // channelID -> postIDs
	moderations      map[string]*ModerationAction
	analytics        map[string]*ChannelAnalytics
	subscriberLookup map[string]map[string]*ChannelSubscriber   // channelID -> subscriberID -> subscriber
	creatorChannels  map[string][]string                        // creatorID -> channelIDs
	joinRequests     map[string]map[string]*ChannelJoinRequest  // channelID -> subscriberID -> request (in-memory fallback)
}

// ChannelServiceOption configures optional durable channel storage.
type ChannelServiceOption func(*ChannelService)

// WithStore enables durable channel storage while retaining the in-memory hot cache.
func WithStore(store ChannelStore) ChannelServiceOption {
	return func(service *ChannelService) {
		service.store = store
	}
}

// NewChannelService creates a new channel service
func NewChannelService(opts ...ChannelServiceOption) *ChannelService {
	service := &ChannelService{
		channels:         make(map[string]*Channel),
		posts:            make(map[string]*ChannelPost),
		subscribers:      make(map[string]*ChannelSubscriber),
		postIndex:        make(map[string][]string),
		moderations:      make(map[string]*ModerationAction),
		analytics:        make(map[string]*ChannelAnalytics),
		subscriberLookup: make(map[string]map[string]*ChannelSubscriber),
		creatorChannels:  make(map[string][]string),
		joinRequests:     make(map[string]map[string]*ChannelJoinRequest),
	}
	for _, opt := range opts {
		opt(service)
	}
	if service.store != nil {
		service.loadWarmCache()
	}
	return service
}

func (cs *ChannelService) loadWarmCache() {
	channels, posts, subs, err := cs.store.LoadAll(context.Background())
	if err != nil {
		log.Printf("broadcast channel cache warm-up failed: %v", err)
		return
	}
	for _, channel := range channels {
		cs.channels[channel.ID] = channel
		cs.postIndex[channel.ID] = []string{}
		cs.subscriberLookup[channel.ID] = make(map[string]*ChannelSubscriber)
		cs.creatorChannels[channel.CreatorID] = append(cs.creatorChannels[channel.CreatorID], channel.ID)
	}
	for _, post := range posts {
		cs.posts[post.ID] = post
		cs.postIndex[post.ChannelID] = append(cs.postIndex[post.ChannelID], post.ID)
	}
	for _, sub := range subs {
		cs.subscribers[sub.ID] = sub
		if cs.subscriberLookup[sub.ChannelID] == nil {
			cs.subscriberLookup[sub.ChannelID] = make(map[string]*ChannelSubscriber)
		}
		cs.subscriberLookup[sub.ChannelID][sub.SubscriberID] = sub
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

	// Creator is auto-subscribed as admin so they can post immediately.
	sub := NewChannelSubscriber(channel.ID, creatorID)
	sub.Role = SubscriberRoleAdmin
	cs.subscriberLookup[channel.ID][creatorID] = sub
	cs.subscribers[sub.ID] = sub
	channel.SubscriberCount = 1

	if cs.store != nil {
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return nil, fmt.Errorf("persist channel: %w", err)
		}
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return nil, fmt.Errorf("persist creator subscription: %w", err)
		}
	}
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

	if cs.store != nil {
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return nil, fmt.Errorf("persist channel: %w", err)
		}
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

	delete(cs.subscriberLookup, channelID)

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
	if cs.store != nil {
		if err := cs.store.DeleteChannel(context.Background(), channelID); err != nil {
			return fmt.Errorf("delete persisted channel: %w", err)
		}
	}
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

	if cs.store != nil {
		if err := cs.store.SavePost(context.Background(), post); err != nil {
			return nil, fmt.Errorf("persist post: %w", err)
		}
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return nil, fmt.Errorf("persist channel post count: %w", err)
		}
	}
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
		if cs.store != nil {
			if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
				return nil, fmt.Errorf("persist channel post count: %w", err)
			}
		}
	}

	if cs.store != nil {
		if err := cs.store.SavePost(context.Background(), post); err != nil {
			return nil, fmt.Errorf("persist post: %w", err)
		}
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
	if cs.store != nil {
		if err := cs.store.DeletePost(context.Background(), postID); err != nil {
			return fmt.Errorf("delete persisted post: %w", err)
		}
	}
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
	if existing, exists := subscribers[subscriberID]; exists {
		return existing, nil
	}

	sub := NewChannelSubscriber(channelID, subscriberID)
	subscribers[subscriberID] = sub
	cs.subscribers[sub.ID] = sub
	channel.SubscriberCount++

	if cs.store != nil {
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return nil, fmt.Errorf("persist subscription: %w", err)
		}
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return nil, fmt.Errorf("persist channel subscriber count: %w", err)
		}
	}
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

	sub, exists := subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("not subscribed")
	}

	delete(subscribers, subscriberID)
	delete(cs.subscribers, sub.ID)
	channel.SubscriberCount--

	if cs.store != nil {
		if err := cs.store.DeleteSubscriber(context.Background(), channelID, subscriberID); err != nil {
			return fmt.Errorf("delete persisted subscription: %w", err)
		}
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return fmt.Errorf("persist channel subscriber count: %w", err)
		}
	}
	return nil
}

// RequestJoin adds the caller to the channel. If the channel requires approval,
// a pending join request is recorded (returns approved=false); otherwise the
// caller is subscribed immediately (approved=true). This wires channel
// RequireApproval to membership (previously it only gated posts).
func (cs *ChannelService) RequestJoin(channelID, subscriberID string) (approved bool, err error) {
	cs.mu.RLock()
	channel, exists := cs.channels[channelID]
	needsApproval := exists && channel.RequireApproval
	// Already a subscriber? Treat as approved (idempotent).
	if subs, ok := cs.subscriberLookup[channelID]; ok {
		if _, subbed := subs[subscriberID]; subbed {
			cs.mu.RUnlock()
			return true, nil
		}
	}
	cs.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("channel not found")
	}
	if !needsApproval {
		if _, err := cs.Subscribe(channelID, subscriberID); err != nil {
			return false, err
		}
		return true, nil
	}

	req := &ChannelJoinRequest{
		ChannelID:    channelID,
		SubscriberID: subscriberID,
		Status:       JoinRequestStatusPending,
		RequestedAt:  time.Now(),
	}
	cs.mu.Lock()
	if cs.joinRequests[channelID] == nil {
		cs.joinRequests[channelID] = make(map[string]*ChannelJoinRequest)
	}
	cs.joinRequests[channelID][subscriberID] = req
	cs.mu.Unlock()

	if cs.store != nil {
		if err := cs.store.SaveJoinRequest(context.Background(), req); err != nil {
			return false, fmt.Errorf("persist join request: %w", err)
		}
	}
	return false, nil
}

// ListPendingJoinRequests returns pending membership requests for a channel.
func (cs *ChannelService) ListPendingJoinRequests(channelID string) ([]*ChannelJoinRequest, error) {
	if cs.store != nil {
		return cs.store.ListJoinRequests(context.Background(), channelID, JoinRequestStatusPending)
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	var out []*ChannelJoinRequest
	for _, req := range cs.joinRequests[channelID] {
		if req.Status == JoinRequestStatusPending {
			out = append(out, req)
		}
	}
	return out, nil
}

// ApproveJoinRequest subscribes the requester and clears the pending request.
func (cs *ChannelService) ApproveJoinRequest(channelID, subscriberID string) error {
	if _, err := cs.Subscribe(channelID, subscriberID); err != nil {
		return err
	}
	cs.clearJoinRequest(channelID, subscriberID)
	return nil
}

// DenyJoinRequest clears a pending request without subscribing.
func (cs *ChannelService) DenyJoinRequest(channelID, subscriberID string) error {
	cs.clearJoinRequest(channelID, subscriberID)
	return nil
}

func (cs *ChannelService) clearJoinRequest(channelID, subscriberID string) {
	cs.mu.Lock()
	if reqs, ok := cs.joinRequests[channelID]; ok {
		delete(reqs, subscriberID)
	}
	cs.mu.Unlock()
	if cs.store != nil {
		_ = cs.store.DeleteJoinRequest(context.Background(), channelID, subscriberID)
	}
}

// RemoveSubscriber is an admin-initiated removal (kick). Delegates to Unsubscribe.
func (cs *ChannelService) RemoveSubscriber(channelID, subscriberID string) error {
	return cs.Unsubscribe(channelID, subscriberID)
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
	if cs.store != nil {
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return fmt.Errorf("persist subscriber role: %w", err)
		}
	}
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
	if cs.store != nil {
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return fmt.Errorf("persist muted subscriber: %w", err)
		}
	}
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
	if cs.store != nil {
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return fmt.Errorf("persist unmuted subscriber: %w", err)
		}
	}
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
	if cs.store != nil {
		if err := cs.store.SaveSubscriber(context.Background(), sub); err != nil {
			return fmt.Errorf("persist blocked subscriber: %w", err)
		}
	}
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

	if cs.store != nil {
		if err := cs.store.SavePost(context.Background(), post); err != nil {
			return nil, fmt.Errorf("persist moderated post: %w", err)
		}
	}
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
		if cs.store != nil {
			if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
				return fmt.Errorf("persist channel post count: %w", err)
			}
		}
	}

	if cs.store != nil {
		if err := cs.store.SavePost(context.Background(), post); err != nil {
			return fmt.Errorf("persist approved post: %w", err)
		}
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
	if cs.store != nil {
		if err := cs.store.SavePost(context.Background(), post); err != nil {
			return fmt.Errorf("persist rejected post: %w", err)
		}
	}

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

	if cs.store != nil {
		if err := cs.store.SaveChannel(context.Background(), channel); err != nil {
			return fmt.Errorf("persist channel view count: %w", err)
		}
	}
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
