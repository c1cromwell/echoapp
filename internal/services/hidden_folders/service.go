package hidden_folders

import (
	"fmt"
	"sync"
	"time"
)

// HiddenFolderService manages hidden folder operations
type HiddenFolderService struct {
	mu                 sync.RWMutex
	folders            map[string]*HiddenFolder            // folderID → folder
	messages           map[string][]*HiddenMessage         // folderID → messages
	routes             map[string]string                   // conversationID → folderID
	lockoutTrackers    map[string]*BiometricLockoutTracker // folderID → tracker
	securityEvents     map[string][]*SecurityEvent         // folderID → events
	folderKeyCache     map[string]bool                     // folderID → isUnlocked
	unlockedFolderKeys map[string]time.Time                // folderID → unlock time
}

// NewHiddenFolderService creates a new service instance
func NewHiddenFolderService() *HiddenFolderService {
	return &HiddenFolderService{
		folders:            make(map[string]*HiddenFolder),
		messages:           make(map[string][]*HiddenMessage),
		routes:             make(map[string]string),
		lockoutTrackers:    make(map[string]*BiometricLockoutTracker),
		securityEvents:     make(map[string][]*SecurityEvent),
		folderKeyCache:     make(map[string]bool),
		unlockedFolderKeys: make(map[string]time.Time),
	}
}

// CreateFolder creates a new hidden folder
func (hfs *HiddenFolderService) CreateFolder(name string, tier SecurityTier) (*HiddenFolder, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name cannot be empty")
	}

	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	// Check max 10 folders per user
	folderCount := len(hfs.folders)
	if folderCount >= 10 {
		return nil, fmt.Errorf("maximum 10 hidden folders allowed")
	}

	folder := NewHiddenFolder(name, tier)
	hfs.folders[folder.ID] = folder
	hfs.messages[folder.ID] = []*HiddenMessage{}
	hfs.lockoutTrackers[folder.ID] = NewBiometricLockoutTracker(folder.ID)

	return folder, nil
}

// GetFolder retrieves a folder by ID
func (hfs *HiddenFolderService) GetFolder(folderID string) (*HiddenFolder, error) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	folder, ok := hfs.folders[folderID]
	if !ok {
		return nil, fmt.Errorf("folder not found: %s", folderID)
	}

	return folder, nil
}

// ListFolders returns all hidden folders
func (hfs *HiddenFolderService) ListFolders() []*HiddenFolder {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	folders := make([]*HiddenFolder, 0, len(hfs.folders))
	for _, folder := range hfs.folders {
		folders = append(folders, folder)
	}

	return folders
}

// DeleteFolder deletes a folder and all its messages
func (hfs *HiddenFolderService) DeleteFolder(folderID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	if _, ok := hfs.folders[folderID]; !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	// Remove folder
	delete(hfs.folders, folderID)
	delete(hfs.messages, folderID)
	delete(hfs.lockoutTrackers, folderID)
	delete(hfs.securityEvents, folderID)
	delete(hfs.folderKeyCache, folderID)
	delete(hfs.unlockedFolderKeys, folderID)

	// Remove all routes pointing to this folder
	for convID, fID := range hfs.routes {
		if fID == folderID {
			delete(hfs.routes, convID)
		}
	}

	return nil
}

// RouteConversation adds a routing rule to direct a conversation to a hidden folder
func (hfs *HiddenFolderService) RouteConversation(conversationID, folderID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	// Verify folder exists
	if _, ok := hfs.folders[folderID]; !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	hfs.routes[conversationID] = folderID
	return nil
}

// UnrouteConversation removes a routing rule (moves conversation out of hidden folder)
func (hfs *HiddenFolderService) UnrouteConversation(conversationID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	if _, ok := hfs.routes[conversationID]; !ok {
		return fmt.Errorf("conversation not routed to hidden folder: %s", conversationID)
	}

	delete(hfs.routes, conversationID)
	return nil
}

// GetFolderForConversation returns the hidden folder ID for a conversation, if any
func (hfs *HiddenFolderService) GetFolderForConversation(conversationID string) (string, bool) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	folderID, ok := hfs.routes[conversationID]
	return folderID, ok
}

// StoreMessage stores a message in a hidden folder
func (hfs *HiddenFolderService) StoreMessage(folderID string, message *HiddenMessage) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	if _, ok := hfs.folders[folderID]; !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	hfs.messages[folderID] = append(hfs.messages[folderID], message)
	return nil
}

// GetMessage retrieves a message from a hidden folder
func (hfs *HiddenFolderService) GetMessage(folderID, messageID string) (*HiddenMessage, error) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	messages, ok := hfs.messages[folderID]
	if !ok {
		return nil, fmt.Errorf("folder not found: %s", folderID)
	}

	for _, msg := range messages {
		if msg.ID == messageID {
			return msg, nil
		}
	}

	return nil, fmt.Errorf("message not found: %s", messageID)
}

// GetMessagesInConversation returns all messages in a conversation within a hidden folder
func (hfs *HiddenFolderService) GetMessagesInConversation(folderID, conversationID string) []*HiddenMessage {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	messages, ok := hfs.messages[folderID]
	if !ok {
		return []*HiddenMessage{}
	}

	var result []*HiddenMessage
	for _, msg := range messages {
		if msg.ConversationID == conversationID {
			result = append(result, msg)
		}
	}

	return result
}

// UnlockFolder marks a folder as unlocked (biometric succeeded)
func (hfs *HiddenFolderService) UnlockFolder(folderID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	if _, ok := hfs.folders[folderID]; !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	hfs.folderKeyCache[folderID] = true
	hfs.unlockedFolderKeys[folderID] = time.Now()

	// Reset lockout on successful unlock
	if tracker, ok := hfs.lockoutTrackers[folderID]; ok {
		tracker.ResetFailedAttempts()
	}

	return nil
}

// LockFolder marks a folder as locked
func (hfs *HiddenFolderService) LockFolder(folderID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	if _, ok := hfs.folders[folderID]; !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	hfs.folderKeyCache[folderID] = false
	delete(hfs.unlockedFolderKeys, folderID)

	return nil
}

// IsFolderUnlocked returns true if a folder is currently unlocked
func (hfs *HiddenFolderService) IsFolderUnlocked(folderID string) bool {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	return hfs.folderKeyCache[folderID]
}

// RecordFailedBiometricAttempt records a failed biometric attempt and returns the new count
func (hfs *HiddenFolderService) RecordFailedBiometricAttempt(folderID string) (int, error) {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	tracker, ok := hfs.lockoutTrackers[folderID]
	if !ok {
		return 0, fmt.Errorf("folder not found: %s", folderID)
	}

	count := tracker.RecordFailedAttempt()

	// Record security event
	event := NewSecurityEvent(folderID, "biometric_attempt", fmt.Sprintf("Failed attempt %d", count))
	if _, ok := hfs.securityEvents[folderID]; !ok {
		hfs.securityEvents[folderID] = []*SecurityEvent{}
	}
	hfs.securityEvents[folderID] = append(hfs.securityEvents[folderID], event)

	return count, nil
}

// CheckLockoutStatus returns the lockout status for a folder
func (hfs *HiddenFolderService) CheckLockoutStatus(folderID string) (string, bool, error) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	tracker, ok := hfs.lockoutTrackers[folderID]
	if !ok {
		return "", false, fmt.Errorf("folder not found: %s", folderID)
	}

	status := tracker.GetLockoutStatus()
	isWipeTriggered := tracker.IsWipeTriggered()

	return status, isWipeTriggered, nil
}

// WipeFolder securely wipes a folder after 10 failed attempts
func (hfs *HiddenFolderService) WipeFolder(folderID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	tracker, ok := hfs.lockoutTrackers[folderID]
	if !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	if !tracker.IsWipeTriggered() {
		return fmt.Errorf("wipe not triggered for folder: %s", folderID)
	}

	// Record security event
	event := NewSecurityEvent(folderID, "wipe", "Folder wiped after 10 failed biometric attempts")
	hfs.securityEvents[folderID] = append(hfs.securityEvents[folderID], event)

	// Clear all messages in folder
	hfs.messages[folderID] = []*HiddenMessage{}

	// The folder key would be deleted from Keychain on iOS (not applicable here)
	// Encrypted files would be overwritten with random bytes on iOS filesystem

	return nil
}

// GetSecurityEvents returns all security events for a folder
func (hfs *HiddenFolderService) GetSecurityEvents(folderID string) []*SecurityEvent {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	events, ok := hfs.securityEvents[folderID]
	if !ok {
		return []*SecurityEvent{}
	}

	result := make([]*SecurityEvent, len(events))
	copy(result, events)
	return result
}

// UpdateFolderSettings updates notification mode and auto-lock settings
func (hfs *HiddenFolderService) UpdateFolderSettings(folderID string, notificationMode NotificationMode, autoLockTimeout time.Duration, screenshotProtection bool) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	folder, ok := hfs.folders[folderID]
	if !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	folder.NotificationMode = notificationMode
	folder.AutoLockTimeout = autoLockTimeout
	folder.ScreenshotProtection = screenshotProtection
	folder.LastAccessedAt = time.Now()

	return nil
}

// UpdateMediaUsage updates the media usage for a folder
func (hfs *HiddenFolderService) UpdateMediaUsage(folderID string, additionalBytes int64) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	folder, ok := hfs.folders[folderID]
	if !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	folder.MediaUsedBytes += additionalBytes
	if folder.MediaUsedBytes < 0 {
		folder.MediaUsedBytes = 0
	}

	return nil
}

// IsConversationHidden returns true if a conversation is in any hidden folder
func (hfs *HiddenFolderService) IsConversationHidden(conversationID string) bool {
	_, hidden := hfs.GetFolderForConversation(conversationID)
	return hidden
}

// GetConversationCount returns the number of conversations in a folder
func (hfs *HiddenFolderService) GetConversationCount(folderID string) (int, error) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	if _, ok := hfs.folders[folderID]; !ok {
		return 0, fmt.Errorf("folder not found: %s", folderID)
	}

	conversationSet := make(map[string]bool)
	for convID, fID := range hfs.routes {
		if fID == folderID {
			conversationSet[convID] = true
		}
	}

	return len(conversationSet), nil
}

// GetMessageCount returns the total number of messages in a folder
func (hfs *HiddenFolderService) GetMessageCount(folderID string) (int, error) {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	messages, ok := hfs.messages[folderID]
	if !ok {
		return 0, fmt.Errorf("folder not found: %s", folderID)
	}

	return len(messages), nil
}

// LockAllFolders locks all folders (e.g., on app background)
func (hfs *HiddenFolderService) LockAllFolders() {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	for folderID := range hfs.folderKeyCache {
		hfs.folderKeyCache[folderID] = false
		delete(hfs.unlockedFolderKeys, folderID)
	}
}

// DeleteMessage removes a message from a folder
func (hfs *HiddenFolderService) DeleteMessage(folderID, messageID string) error {
	hfs.mu.Lock()
	defer hfs.mu.Unlock()

	messages, ok := hfs.messages[folderID]
	if !ok {
		return fmt.Errorf("folder not found: %s", folderID)
	}

	newMessages := make([]*HiddenMessage, 0, len(messages))
	found := false

	for _, msg := range messages {
		if msg.ID != messageID {
			newMessages = append(newMessages, msg)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("message not found: %s", messageID)
	}

	hfs.messages[folderID] = newMessages
	return nil
}

// GetUnlockedDuration returns how long a folder has been unlocked
func (hfs *HiddenFolderService) GetUnlockedDuration(folderID string) *time.Duration {
	hfs.mu.RLock()
	defer hfs.mu.RUnlock()

	unlockedAt, ok := hfs.unlockedFolderKeys[folderID]
	if !ok {
		return nil
	}

	duration := time.Since(unlockedAt)
	return &duration
}
