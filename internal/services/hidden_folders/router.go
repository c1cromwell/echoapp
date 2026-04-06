package hidden_folders

import (
	"fmt"
	"sync"
	"time"
)

// MessageRouter determines message routing to hidden or main storage
type MessageRouter struct {
	mu                   sync.RWMutex
	hiddenFolderService  *HiddenFolderService
	routingIndex         map[string]string // conversationID → folderID (cache for fast lookups)
	notificationHandlers map[string]NotificationHandler
}

// NotificationHandler defines how to handle notifications for hidden messages
type NotificationHandler interface {
	SendSuppressed(conversationID string)
	SendRedacted(conversationID string)
	SendNormal(conversationID, senderID, preview string)
}

// RoutingDecision represents the result of routing analysis
type RoutingDecision struct {
	IsHidden           bool
	FolderID           string
	ShouldNotify       bool
	NotificationType   NotificationMode
	RequiresBiometric  bool
	EncryptionRequired bool
}

// NewMessageRouter creates a new message router
func NewMessageRouter(hiddenFolderService *HiddenFolderService) *MessageRouter {
	return &MessageRouter{
		hiddenFolderService:  hiddenFolderService,
		routingIndex:         make(map[string]string),
		notificationHandlers: make(map[string]NotificationHandler),
	}
}

// RouteMessage determines where an incoming message should be stored
func (mr *MessageRouter) RouteMessage(conversationID, senderID, contentPreview string) (*RoutingDecision, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Check if conversation is routed to a hidden folder
	folderID, isHidden := mr.routingIndex[conversationID]

	decision := &RoutingDecision{
		IsHidden:           isHidden,
		FolderID:           folderID,
		EncryptionRequired: true,
	}

	if !isHidden {
		// Normal message flow
		decision.ShouldNotify = true
		decision.NotificationType = NotificationSuppressed
		return decision, nil
	}

	// Message is for a hidden conversation
	folder, err := mr.hiddenFolderService.GetFolder(folderID)
	if err != nil {
		return nil, fmt.Errorf("folder lookup failed: %w", err)
	}

	// Determine notification behavior
	decision.NotificationType = folder.NotificationMode
	decision.RequiresBiometric = true

	switch folder.NotificationMode {
	case NotificationSuppressed:
		decision.ShouldNotify = false

	case NotificationRedacted:
		decision.ShouldNotify = true
		// Notification will be sent with no sender name or preview

	case NotificationUnlockedOnly:
		// Only notify if folder is currently unlocked
		decision.ShouldNotify = mr.hiddenFolderService.IsFolderUnlocked(folderID)
	}

	return decision, nil
}

// SyncRoutingIndex rebuilds the routing index from the hidden folder service
func (mr *MessageRouter) SyncRoutingIndex() {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Clear and rebuild
	mr.routingIndex = make(map[string]string)

	folders := mr.hiddenFolderService.ListFolders()
	for _, _ = range folders {
		// In a real implementation, this would iterate through all conversations
		// assigned to this folder and build the index
		// For now, we'll get this from the service's internal routes
		// This is a simplified version — in production, you'd query the database
	}
}

// RegisterConversationRoute registers a conversation to be routed to a hidden folder
func (mr *MessageRouter) RegisterConversationRoute(conversationID, folderID string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Verify folder exists
	if _, err := mr.hiddenFolderService.GetFolder(folderID); err != nil {
		return fmt.Errorf("folder lookup failed: %w", err)
	}

	mr.routingIndex[conversationID] = folderID
	return mr.hiddenFolderService.RouteConversation(conversationID, folderID)
}

// UnregisterConversationRoute removes a conversation from hidden folder routing
func (mr *MessageRouter) UnregisterConversationRoute(conversationID string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	delete(mr.routingIndex, conversationID)
	return mr.hiddenFolderService.UnrouteConversation(conversationID)
}

// GetRoutingDecisionForSearch determines if a conversation should be searchable
func (mr *MessageRouter) GetRoutingDecisionForSearch(conversationID string) *RoutingDecision {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, isHidden := mr.routingIndex[conversationID]

	return &RoutingDecision{
		IsHidden:          isHidden,
		FolderID:          folderID,
		ShouldNotify:      false,
		RequiresBiometric: isHidden,
	}
}

// GetRoutingDecisionForForward determines if a message can be forwarded
func (mr *MessageRouter) GetRoutingDecisionForForward(messageID, conversationID string) (*RoutingDecision, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, isHidden := mr.routingIndex[conversationID]

	if !isHidden {
		return &RoutingDecision{
			IsHidden: false,
		}, nil
	}

	// Message is in hidden folder — forwarding restrictions apply
	decision := &RoutingDecision{
		IsHidden: true,
		FolderID: folderID,
	}

	return decision, nil
}

// CanForwardHiddenMessage returns true if a hidden message can be forwarded to another hidden folder
func (mr *MessageRouter) CanForwardHiddenMessage(sourceConversationID, targetConversationID string) (bool, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	_, sourceIsHidden := mr.routingIndex[sourceConversationID]
	_, targetIsHidden := mr.routingIndex[targetConversationID]

	if !sourceIsHidden {
		return true, nil // Normal message, can forward anywhere
	}

	if sourceIsHidden && targetIsHidden {
		return true, nil // Both are hidden, allow forward
	}

	// Source is hidden but target is not — block
	return false, fmt.Errorf("cannot forward message from hidden folder to non-hidden conversation")
}

// GetHiddenFoldersCount returns the number of active hidden folders
func (mr *MessageRouter) GetHiddenFoldersCount() int {
	folders := mr.hiddenFolderService.ListFolders()
	return len(folders)
}

// GetRoutedConversationsCount returns the total number of conversations routed to hidden folders
func (mr *MessageRouter) GetRoutedConversationsCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	return len(mr.routingIndex)
}

// GetFolderIDForConversation returns the folder ID for a conversation, if hidden
func (mr *MessageRouter) GetFolderIDForConversation(conversationID string) (string, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, ok := mr.routingIndex[conversationID]
	return folderID, ok
}

// EnforceSystemExclusions removes any accidentally donated system integration data for hidden contacts
func (mr *MessageRouter) EnforceSystemExclusions(hiddenContactDIDs []string) error {
	// In a real implementation, this would:
	// 1. Call INInteraction.delete() for hidden contacts
	// 2. Delete from CSSearchableIndex for hidden content
	// 3. Clear any Spotlight data
	// 4. Remove from keyboard prediction suggestions
	// 5. Exclude from Siri donations
	//
	// This is an iOS-specific operation; here we just validate the concept

	if len(hiddenContactDIDs) == 0 {
		return nil
	}

	// Placeholder for enforcing system exclusions
	return nil
}

// GetConversationNotificationMode returns the notification mode for a hidden conversation
func (mr *MessageRouter) GetConversationNotificationMode(conversationID string) (NotificationMode, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, isHidden := mr.routingIndex[conversationID]
	if !isHidden {
		// Non-hidden conversations use standard notification mode
		return NotificationSuppressed, nil
	}

	folder, err := mr.hiddenFolderService.GetFolder(folderID)
	if err != nil {
		return "", fmt.Errorf("folder lookup failed: %w", err)
	}

	return folder.NotificationMode, nil
}

// LockAllAndExclude locks all hidden folders and ensures system exclusions are enforced
func (mr *MessageRouter) LockAllAndExclude() error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Lock all folders
	mr.hiddenFolderService.LockAllFolders()

	// Gather all hidden contact DIDs
	hiddenContacts := make([]string, 0)
	folders := mr.hiddenFolderService.ListFolders()
	for _, folder := range folders {
		// In real implementation, extract DIDs from conversations in this folder
		_ = folder
	}

	// Enforce system exclusions
	return mr.EnforceSystemExclusions(hiddenContacts)
}

// GetMessageStoreLocation returns where a message should be stored
func (mr *MessageRouter) GetMessageStoreLocation(conversationID string) StoreLocation {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if folderID, isHidden := mr.routingIndex[conversationID]; isHidden {
		return StoreLocation{
			IsHidden:  true,
			FolderID:  folderID,
			StorePath: "hidden_store/" + folderID,
		}
	}

	return StoreLocation{
		IsHidden:  false,
		StorePath: "main_store",
	}
}

// StoreLocation represents where a message should be stored
type StoreLocation struct {
	IsHidden  bool
	FolderID  string
	StorePath string
}

// MessagePreflightCheck performs pre-send checks for messages in hidden folders
func (mr *MessageRouter) MessagePreflightCheck(conversationID, messageContent string) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, isHidden := mr.routingIndex[conversationID]
	if !isHidden {
		return nil // Normal message, no special checks
	}

	folder, err := mr.hiddenFolderService.GetFolder(folderID)
	if err != nil {
		return fmt.Errorf("folder not found: %w", err)
	}

	// Check if folder is at media quota (for attachments)
	if folder.IsAtMediaQuota() {
		return fmt.Errorf("hidden folder media quota exceeded")
	}

	// Check if folder is locked
	if !mr.hiddenFolderService.IsFolderUnlocked(folderID) {
		return fmt.Errorf("hidden folder is locked; unlock to send messages")
	}

	// Check for other pre-send conditions
	// In a real implementation, check encryption keys, network status, etc.

	return nil
}

// GetAutoLockTimeout returns the auto-lock timeout for a hidden conversation
func (mr *MessageRouter) GetAutoLockTimeout(conversationID string) (*time.Duration, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	folderID, isHidden := mr.routingIndex[conversationID]
	if !isHidden {
		return nil, nil
	}

	folder, err := mr.hiddenFolderService.GetFolder(folderID)
	if err != nil {
		return nil, fmt.Errorf("folder not found: %w", err)
	}

	return &folder.AutoLockTimeout, nil
}

// BroadcastFolderLocked notifies all routes when a folder locks
func (mr *MessageRouter) BroadcastFolderLocked(folderID string) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Find all conversations routed to this folder
	affectedConversations := make([]string, 0)
	for convID, fID := range mr.routingIndex {
		if fID == folderID {
			affectedConversations = append(affectedConversations, convID)
		}
	}

	// In a real implementation, notify UI to close these conversations
	_ = affectedConversations

	return nil
}

// BroadcastFolderUnlocked notifies route listeners when a folder unlocks
func (mr *MessageRouter) BroadcastFolderUnlocked(folderID string) error {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Find all conversations routed to this folder
	affectedConversations := make([]string, 0)
	for convID, fID := range mr.routingIndex {
		if fID == folderID {
			affectedConversations = append(affectedConversations, convID)
		}
	}

	// In a real implementation, notify UI to display queued messages
	_ = affectedConversations

	return nil
}
