// Package database provides a database abstraction layer for the Echo backend.
package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a duplicate record is detected.
var ErrDuplicate = errors.New("duplicate record")

// --- Models ---

// User represents an Echo user account.
type User struct {
	UserID              string    `json:"userId"`
	DID                 string    `json:"did"`
	Username            string    `json:"username"`
	TrustTier           int       `json:"trustTier"`
	PhoneDiscoveryOptIn *bool     `json:"phoneDiscoveryOptIn,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// TrustScore represents a cached trust score for a DID.
type TrustScore struct {
	DID       string    `json:"did"`
	Score     float64   `json:"score"`
	Tier      int       `json:"tier"`
	ExpiresAt time.Time `json:"expiresAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Credential represents a verifiable credential.
type Credential struct {
	CredentialID string    `json:"credentialId"`
	IssuerDID    string    `json:"issuerDid"`
	SubjectDID   string    `json:"subjectDid"`
	SchemaID     string    `json:"schemaId"`
	Status       string    `json:"status"`
	IssuedAt     time.Time `json:"issuedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// QueuedMessage represents a message in the offline queue.
type QueuedMessage struct {
	MessageID      string     `json:"messageId"`
	ConversationID string     `json:"conversationId,omitempty"`
	SenderDID      string     `json:"senderDid"`
	RecipientDID   string     `json:"recipientDid"`
	Payload        []byte     `json:"payload"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	DeliveredAt    *time.Time `json:"deliveredAt,omitempty"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
}

// MessageMeta is the delivery-state + routing view of a queued message, used to
// drive read receipts (WO-192) and server-authoritative reaction fan-out (WO-10).
// It carries no message content — only metadata the relay already holds.
type MessageMeta struct {
	MessageID      string     `json:"messageId"`
	ConversationID string     `json:"conversationId,omitempty"`
	SenderDID      string     `json:"senderDid"`
	RecipientDID   string     `json:"recipientDid"`
	Status         string     `json:"status"`
	DeliveredAt    *time.Time `json:"deliveredAt,omitempty"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
}

// MerkleBatch represents a batch of message hashes for L1 anchoring.
type MerkleBatch struct {
	BatchID   string    `json:"batchId"`
	RootHash  string    `json:"rootHash"`
	LeafCount int       `json:"leafCount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// Contact represents a contact relationship.
type Contact struct {
	OwnerDID   string    `json:"ownerDid"`
	ContactDID string    `json:"contactDid"`
	AddedVia   string    `json:"addedVia"`
	Blocked    bool      `json:"blocked"`
	TrustBadge string    `json:"trustBadge"`
	CreatedAt  time.Time `json:"createdAt"`
}

// InviteLink represents a contact invite link.
type InviteLink struct {
	Code       string    `json:"code"`
	CreatorDID string    `json:"creatorDid"`
	AcceptedBy string    `json:"acceptedBy,omitempty"`
	Accepted   bool      `json:"accepted"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// MediaFile represents encrypted media metadata.
type MediaFile struct {
	FileID        string    `json:"fileId"`
	UploaderDID   string    `json:"uploaderDid"`
	ContentType   string    `json:"contentType"`
	EncryptedSize int64     `json:"encryptedSize"`
	ChunkCount    int       `json:"chunkCount"`
	ScanStatus    string    `json:"scanStatus"`
	CreatedAt     time.Time `json:"createdAt"`
}

// MediaChunk represents a single chunk of an encrypted file.
type MediaChunk struct {
	ChunkID  string `json:"chunkId"`
	FileID   string `json:"fileId"`
	Index    int    `json:"index"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

// UserDevice represents a registered device for push notifications.
type UserDevice struct {
	DeviceID    string    `json:"deviceId"`
	DID         string    `json:"did"`
	DeviceLabel string    `json:"deviceLabel"`
	PublicKey   string    `json:"publicKey"`
	APNsToken   string    `json:"apnsToken"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NotificationPrefs stores notification preferences for a user.
type NotificationPrefs struct {
	DID                  string `json:"did"`
	PushEnabled          bool   `json:"pushEnabled"`
	MessagePreview       bool   `json:"messagePreview"`
	GroupNotifications   bool   `json:"groupNotifications"`
	ChannelNotifications bool   `json:"channelNotifications"`
	QuietHoursStart      int    `json:"quietHoursStart"`
	QuietHoursEnd        int    `json:"quietHoursEnd"`
}

// LogIndexEntry represents a log entry index for DAG publishing.
type LogIndexEntry struct {
	EntryID       string    `json:"entryId"`
	DID           string    `json:"did"`
	EventType     string    `json:"eventType"`
	TimeRangeFrom time.Time `json:"timeRangeFrom"`
	TimeRangeTo   time.Time `json:"timeRangeTo"`
	CreatedAt     time.Time `json:"createdAt"`
}

// --- Store Interfaces ---

type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByDID(ctx context.Context, did string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdatePhoneDiscoveryOptIn(ctx context.Context, did string, optIn *bool) error
}

type TrustScoreStore interface {
	SetTrustScore(ctx context.Context, ts *TrustScore) error
	GetTrustScore(ctx context.Context, did string) (*TrustScore, error)
	GetTrustScoreBatch(ctx context.Context, dids []string) ([]*TrustScore, error)
}

type CredentialStore interface {
	StoreCredential(ctx context.Context, cred *Credential) error
	GetCredentialsByDID(ctx context.Context, did string) ([]*Credential, error)
	RevokeCredential(ctx context.Context, credentialID string) error
}

type MessageQueueStore interface {
	Enqueue(ctx context.Context, msg *QueuedMessage) error
	Dequeue(ctx context.Context, recipientDID string, limit int) ([]*QueuedMessage, error)
	MarkDelivered(ctx context.Context, messageID string) error
	// MarkRead records that the recipient has read the message (WO-192). It is a
	// monotonic upgrade of delivery state: a read message is also delivered.
	MarkRead(ctx context.Context, messageID string) error
	// GetMessageMeta returns the delivery-state + routing metadata for a message
	// (no content). Used to sync receipts on reconnect and to route reaction
	// signals to the counterparty. Returns ErrNotFound if the message is unknown.
	GetMessageMeta(ctx context.Context, messageID string) (*MessageMeta, error)
	PurgeExpired(ctx context.Context) (int, error)
}

type MerkleBatchStore interface {
	CreateBatch(ctx context.Context, batch *MerkleBatch) error
	GetBatch(ctx context.Context, batchID string) (*MerkleBatch, error)
	FinalizeBatch(ctx context.Context, batchID, rootHash string, leafCount int) error
}

type ContactStore interface {
	AddContact(ctx context.Context, contact *Contact) error
	GetContacts(ctx context.Context, ownerDID string) ([]*Contact, error)
	RemoveContact(ctx context.Context, ownerDID, contactDID string) error
	SetBlocked(ctx context.Context, ownerDID, contactDID string, blocked bool) error
	GetContactCount(ctx context.Context, ownerDID string) (int, error)
}

type InviteStore interface {
	CreateInvite(ctx context.Context, invite *InviteLink) error
	GetInvite(ctx context.Context, code string) (*InviteLink, error)
	AcceptInvite(ctx context.Context, code, acceptedBy string) error
}

type MediaStore interface {
	StoreMediaFile(ctx context.Context, file *MediaFile) error
	GetMediaFile(ctx context.Context, fileID string) (*MediaFile, error)
	UpdateScanStatus(ctx context.Context, fileID, status string) error
	StoreChunk(ctx context.Context, chunk *MediaChunk) error
	GetChunks(ctx context.Context, fileID string) ([]*MediaChunk, error)
}

type NotificationStore interface {
	RegisterDevice(ctx context.Context, device *UserDevice) error
	GetDevicesByDID(ctx context.Context, did string) ([]*UserDevice, error)
	UpdateAPNsToken(ctx context.Context, deviceID, token string) error
	GetNotificationPrefs(ctx context.Context, did string) (*NotificationPrefs, error)
	UpdateNotificationPrefs(ctx context.Context, prefs *NotificationPrefs) error
}

type LogIndexStore interface {
	StoreLogIndex(ctx context.Context, entry *LogIndexEntry) error
	GetLogIndex(ctx context.Context, from, to time.Time) ([]*LogIndexEntry, error)
}

// DB is the composite database interface.
// SMSRecoveryStore stores the phone commitment (H(phone)) for SMS-based recovery.
// The raw phone number is never stored — only the SHA-256 hex commitment.
type SMSRecoveryStore interface {
	// SetSMSRecovery stores {did → phone_hash} (upsert).
	SetSMSRecovery(ctx context.Context, did, phoneHash string) error
	// GetSMSRecoveryByPhoneHash returns the DID registered for the given phone hash.
	GetSMSRecoveryByPhoneHash(ctx context.Context, phoneHash string) (string, error)
}

type DB interface {
	UserStore
	TrustScoreStore
	CredentialStore
	MessageQueueStore
	MerkleBatchStore
	ContactStore
	InviteStore
	MediaStore
	NotificationStore
	LogIndexStore
	SMSRecoveryStore
	ReactionStore
	ContactDiscoveryStore
	MessageOpsStore
	ConversationArchiveStore
	DeviceSyncStore
	ComplyExtendedStore
	ConversationIndex
}

// ReactionRow is a single emoji reaction by one reactor on one message.
type ReactionRow struct {
	MessageID  string
	ReactorDID string
	Emoji      string
}

// ReactionStore persists emoji reactions (one emoji per reactor per message).
type ReactionStore interface {
	// AddReaction upserts the reactor's emoji on a message (replaces any prior).
	AddReaction(ctx context.Context, messageID, reactorDID, emoji string) error
	// RemoveReaction clears the reactor's reaction on a message (no-op if absent).
	RemoveReaction(ctx context.Context, messageID, reactorDID string) error
	// GetReactions returns all reaction rows for a message (unaggregated).
	GetReactions(ctx context.Context, messageID string) ([]ReactionRow, error)
}

// ContactDiscoveryStore persists the OPRF contact-discovery index:
// hex(OPRF_k(phone)) -> DID. Raw phone numbers are never stored.
type ContactDiscoveryStore interface {
	// PutDiscoveryKey upserts an OPRF index key -> DID binding.
	PutDiscoveryKey(ctx context.Context, oprfKey, did string) error
	// AllDiscoveryKeys returns the whole index for client-side matching.
	AllDiscoveryKeys(ctx context.Context) (map[string]string, error)
	// DeleteDiscoveryKeysForDID removes all PSI index entries for a user.
	DeleteDiscoveryKeysForDID(ctx context.Context, did string) error
}

// --- In-Memory Implementation ---

const (
	maxQueuePerRecipient = 1000
	trustScoreTTL        = 60 * time.Second
)

type MemoryDB struct {
	mu sync.RWMutex

	users  map[string]*User
	byUser map[string]*User

	trustScores map[string]*TrustScore

	credentials map[string]*Credential
	credsByDID  map[string][]*Credential

	messages   map[string]*QueuedMessage
	msgByRecip map[string][]*QueuedMessage

	batches map[string]*MerkleBatch

	contacts map[string][]*Contact

	invites map[string]*InviteLink

	mediaFiles  map[string]*MediaFile
	mediaChunks map[string][]*MediaChunk

	devices           map[string]*UserDevice
	devicesByDID      map[string][]*UserDevice
	notificationPrefs map[string]*NotificationPrefs

	logIndex []*LogIndexEntry

	smsRecoveryByDID   map[string]string // did → phone_hash
	smsRecoveryByPhone map[string]string // phone_hash → did

	reactions      map[string]map[string]string // messageID → reactorDID → emoji
	discoveryIndex map[string]string            // hex(OPRF_k(phone)) → did

	// M1 message ops (WO-25/84/59). Content stays opaque; metadata only.
	editVersions map[string][]*MessageEdit   // messageID → ordered immutable versions (retained convos)
	deletedMsgs  map[string]bool             // messageID → tombstoned
	pins         map[string][]*PinnedMessage // conversationID → pinned messages (max 5)
	retained     map[string]bool             // conversationID → retention/litigation-hold flag
	disappearing map[string]int              // conversationID → disappearing TTL seconds (0 = off)
	archived     map[string]bool             // conversationID → archived (WO-198)

	// WO-250 Comply policy metadata (no message content).
	complyPolicies map[string]*RetentionPolicy
	complyBindings map[string]*ConversationPolicyBinding
	complyOrgs     map[string]*ComplyOrgProfile
	// WO-251 litigation / eDiscovery metadata.
	complyMatters       map[string]*LitigationMatter
	complyCustodians    map[string][]*LitigationCustodianBinding
	complyCustodianKeys map[string]struct{}
	complyExports       map[string]*EDiscoveryExport
	complyAudit         []*AuditEvent
	complyMembers       map[string]*OrgMember
	complyDEFP          map[string]*DEFingerprintRecord

	// M3 device sync (WO-CA3): per-device addressed ciphertext streams, content-blind.
	syncStreams map[string]map[string][]*SyncEntry // controllerDID → targetDeviceID → ordered entries
	syncSeq     map[string]map[string]int64        // controllerDID → targetDeviceID → last seq
	syncRevoked map[string]map[string]bool         // controllerDID → targetDeviceID → revoked
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:               make(map[string]*User),
		byUser:              make(map[string]*User),
		trustScores:         make(map[string]*TrustScore),
		credentials:         make(map[string]*Credential),
		credsByDID:          make(map[string][]*Credential),
		messages:            make(map[string]*QueuedMessage),
		msgByRecip:          make(map[string][]*QueuedMessage),
		batches:             make(map[string]*MerkleBatch),
		contacts:            make(map[string][]*Contact),
		invites:             make(map[string]*InviteLink),
		mediaFiles:          make(map[string]*MediaFile),
		mediaChunks:         make(map[string][]*MediaChunk),
		devices:             make(map[string]*UserDevice),
		devicesByDID:        make(map[string][]*UserDevice),
		notificationPrefs:   make(map[string]*NotificationPrefs),
		smsRecoveryByDID:    make(map[string]string),
		smsRecoveryByPhone:  make(map[string]string),
		reactions:           make(map[string]map[string]string),
		discoveryIndex:      make(map[string]string),
		editVersions:        make(map[string][]*MessageEdit),
		deletedMsgs:         make(map[string]bool),
		pins:                make(map[string][]*PinnedMessage),
		retained:            make(map[string]bool),
		disappearing:        make(map[string]int),
		archived:            make(map[string]bool),
		complyPolicies:      make(map[string]*RetentionPolicy),
		complyBindings:      make(map[string]*ConversationPolicyBinding),
		complyOrgs:          make(map[string]*ComplyOrgProfile),
		complyMatters:       make(map[string]*LitigationMatter),
		complyCustodians:    make(map[string][]*LitigationCustodianBinding),
		complyCustodianKeys: make(map[string]struct{}),
		complyExports:       make(map[string]*EDiscoveryExport),
		complyMembers:       make(map[string]*OrgMember),
		complyDEFP:          make(map[string]*DEFingerprintRecord),
		syncStreams:         make(map[string]map[string][]*SyncEntry),
		syncSeq:             make(map[string]map[string]int64),
		syncRevoked:         make(map[string]map[string]bool),
	}
}

// --- Reactions (MemoryDB) ---

func (m *MemoryDB) AddReaction(_ context.Context, messageID, reactorDID, emoji string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reactions[messageID] == nil {
		m.reactions[messageID] = make(map[string]string)
	}
	m.reactions[messageID][reactorDID] = emoji
	return nil
}

func (m *MemoryDB) RemoveReaction(_ context.Context, messageID, reactorDID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r := m.reactions[messageID]; r != nil {
		delete(r, reactorDID)
		if len(r) == 0 {
			delete(m.reactions, messageID)
		}
	}
	return nil
}

func (m *MemoryDB) GetReactions(_ context.Context, messageID string) ([]ReactionRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rows := make([]ReactionRow, 0, len(m.reactions[messageID]))
	for did, emoji := range m.reactions[messageID] {
		rows = append(rows, ReactionRow{MessageID: messageID, ReactorDID: did, Emoji: emoji})
	}
	return rows, nil
}

// --- Contact discovery index (MemoryDB) ---

func (m *MemoryDB) PutDiscoveryKey(_ context.Context, oprfKey, did string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.discoveryIndex[oprfKey] = did
	return nil
}

func (m *MemoryDB) AllDiscoveryKeys(_ context.Context) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.discoveryIndex))
	for k, v := range m.discoveryIndex {
		out[k] = v
	}
	return out, nil
}

// --- SMS Recovery Store ---

func (m *MemoryDB) SetSMSRecovery(_ context.Context, did, phoneHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Remove old phone hash mapping if DID was previously registered.
	if old, ok := m.smsRecoveryByDID[did]; ok {
		delete(m.smsRecoveryByPhone, old)
	}
	m.smsRecoveryByDID[did] = phoneHash
	m.smsRecoveryByPhone[phoneHash] = did
	return nil
}

func (m *MemoryDB) GetSMSRecoveryByPhoneHash(_ context.Context, phoneHash string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	did, ok := m.smsRecoveryByPhone[phoneHash]
	if !ok {
		return "", fmt.Errorf("phone hash not registered")
	}
	return did, nil
}

// --- User Store ---

func (m *MemoryDB) CreateUser(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.DID]; exists {
		return ErrDuplicate
	}
	if _, exists := m.byUser[user.Username]; exists {
		return ErrDuplicate
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if user.TrustTier == 0 {
		user.TrustTier = 1
	}
	m.users[user.DID] = user
	m.byUser[user.Username] = user
	return nil
}

func (m *MemoryDB) GetUserByDID(ctx context.Context, did string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[did]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *MemoryDB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.byUser[username]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *MemoryDB) UpdatePhoneDiscoveryOptIn(_ context.Context, did string, optIn *bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[did]
	if !ok {
		return ErrNotFound
	}
	if optIn == nil {
		u.PhoneDiscoveryOptIn = nil
	} else {
		v := *optIn
		u.PhoneDiscoveryOptIn = &v
	}
	u.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryDB) DeleteDiscoveryKeysForDID(_ context.Context, did string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, owner := range m.discoveryIndex {
		if owner == did {
			delete(m.discoveryIndex, key)
		}
	}
	return nil
}

// --- Trust Score Store ---

func (m *MemoryDB) SetTrustScore(ctx context.Context, ts *TrustScore) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ts.UpdatedAt = time.Now()
	if ts.ExpiresAt.IsZero() {
		ts.ExpiresAt = time.Now().Add(trustScoreTTL)
	}
	m.trustScores[ts.DID] = ts
	return nil
}

func (m *MemoryDB) GetTrustScore(ctx context.Context, did string) (*TrustScore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ts, ok := m.trustScores[did]
	if !ok {
		return nil, ErrNotFound
	}
	if time.Now().After(ts.ExpiresAt) {
		return nil, ErrNotFound
	}
	return ts, nil
}

func (m *MemoryDB) GetTrustScoreBatch(ctx context.Context, dids []string) ([]*TrustScore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*TrustScore
	now := time.Now()
	for _, did := range dids {
		if ts, ok := m.trustScores[did]; ok && now.Before(ts.ExpiresAt) {
			result = append(result, ts)
		}
	}
	return result, nil
}

// --- Credential Store ---

func (m *MemoryDB) StoreCredential(ctx context.Context, cred *Credential) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cred.IssuedAt = time.Now()
	m.credentials[cred.CredentialID] = cred
	m.credsByDID[cred.SubjectDID] = append(m.credsByDID[cred.SubjectDID], cred)
	return nil
}

func (m *MemoryDB) GetCredentialsByDID(ctx context.Context, did string) ([]*Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	creds := m.credsByDID[did]
	if creds == nil {
		return []*Credential{}, nil
	}
	return creds, nil
}

func (m *MemoryDB) RevokeCredential(ctx context.Context, credentialID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cred, ok := m.credentials[credentialID]
	if !ok {
		return ErrNotFound
	}
	cred.Status = "revoked"
	return nil
}

// --- Message Queue Store ---

func (m *MemoryDB) Enqueue(ctx context.Context, msg *QueuedMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg.CreatedAt = time.Now()
	if msg.ExpiresAt.IsZero() {
		msg.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	}
	msg.Status = "queued"
	m.messages[msg.MessageID] = msg

	queue := m.msgByRecip[msg.RecipientDID]
	if len(queue) >= maxQueuePerRecipient {
		oldest := queue[0]
		oldest.Status = "expired"
		queue = queue[1:]
	}
	m.msgByRecip[msg.RecipientDID] = append(queue, msg)
	return nil
}

func (m *MemoryDB) Dequeue(ctx context.Context, recipientDID string, limit int) ([]*QueuedMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.msgByRecip[recipientDID]
	var result []*QueuedMessage
	var remaining []*QueuedMessage

	for _, msg := range queue {
		if msg.Status == "queued" && len(result) < limit {
			msg.Status = "delivered"
			result = append(result, msg)
		} else {
			remaining = append(remaining, msg)
		}
	}
	m.msgByRecip[recipientDID] = remaining
	return result, nil
}

func (m *MemoryDB) MarkDelivered(ctx context.Context, messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg, ok := m.messages[messageID]
	if !ok {
		return ErrNotFound
	}
	msg.Status = "delivered"
	if msg.DeliveredAt == nil {
		now := time.Now()
		msg.DeliveredAt = &now
	}
	return nil
}

// MarkRead records a read receipt (WO-192). Read implies delivered, so both
// timestamps are set if not already present.
func (m *MemoryDB) MarkRead(ctx context.Context, messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg, ok := m.messages[messageID]
	if !ok {
		return ErrNotFound
	}
	now := time.Now()
	if msg.DeliveredAt == nil {
		msg.DeliveredAt = &now
	}
	msg.Status = "read"
	if msg.ReadAt == nil {
		msg.ReadAt = &now
	}
	return nil
}

// GetMessageMeta returns delivery-state + routing metadata for a message.
func (m *MemoryDB) GetMessageMeta(ctx context.Context, messageID string) (*MessageMeta, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msg, ok := m.messages[messageID]
	if !ok {
		return nil, ErrNotFound
	}
	return &MessageMeta{
		MessageID:      msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderDID:      msg.SenderDID,
		RecipientDID:   msg.RecipientDID,
		Status:         msg.Status,
		DeliveredAt:    msg.DeliveredAt,
		ReadAt:         msg.ReadAt,
	}, nil
}

func (m *MemoryDB) PurgeExpired(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0
	for id, msg := range m.messages {
		if now.After(msg.ExpiresAt) {
			msg.Status = "expired"
			delete(m.messages, id)
			count++
		}
	}
	for did, queue := range m.msgByRecip {
		var remaining []*QueuedMessage
		for _, msg := range queue {
			if msg.Status != "expired" {
				remaining = append(remaining, msg)
			}
		}
		m.msgByRecip[did] = remaining
	}
	return count, nil
}

// --- Merkle Batch Store ---

func (m *MemoryDB) CreateBatch(ctx context.Context, batch *MerkleBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	batch.CreatedAt = time.Now()
	batch.Status = "open"
	m.batches[batch.BatchID] = batch
	return nil
}

func (m *MemoryDB) GetBatch(ctx context.Context, batchID string) (*MerkleBatch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.batches[batchID]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}

func (m *MemoryDB) FinalizeBatch(ctx context.Context, batchID, rootHash string, leafCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.batches[batchID]
	if !ok {
		return ErrNotFound
	}
	b.RootHash = rootHash
	b.LeafCount = leafCount
	b.Status = "finalized"
	return nil
}

// --- Contact Store ---

func (m *MemoryDB) AddContact(ctx context.Context, contact *Contact) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.contacts[contact.OwnerDID] {
		if c.ContactDID == contact.ContactDID {
			return ErrDuplicate
		}
	}
	contact.CreatedAt = time.Now()
	m.contacts[contact.OwnerDID] = append(m.contacts[contact.OwnerDID], contact)
	return nil
}

func (m *MemoryDB) GetContacts(ctx context.Context, ownerDID string) ([]*Contact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contacts := m.contacts[ownerDID]
	if contacts == nil {
		return []*Contact{}, nil
	}
	return contacts, nil
}

func (m *MemoryDB) RemoveContact(ctx context.Context, ownerDID, contactDID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	contacts := m.contacts[ownerDID]
	for i, c := range contacts {
		if c.ContactDID == contactDID {
			m.contacts[ownerDID] = append(contacts[:i], contacts[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MemoryDB) SetBlocked(ctx context.Context, ownerDID, contactDID string, blocked bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.contacts[ownerDID] {
		if c.ContactDID == contactDID {
			c.Blocked = blocked
			return nil
		}
	}
	return ErrNotFound
}

func (m *MemoryDB) GetContactCount(ctx context.Context, ownerDID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.contacts[ownerDID]), nil
}

// --- Invite Store ---

func (m *MemoryDB) CreateInvite(ctx context.Context, invite *InviteLink) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	invite.CreatedAt = time.Now()
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	m.invites[invite.Code] = invite
	return nil
}

func (m *MemoryDB) GetInvite(ctx context.Context, code string) (*InviteLink, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inv, ok := m.invites[code]
	if !ok {
		return nil, ErrNotFound
	}
	return inv, nil
}

func (m *MemoryDB) AcceptInvite(ctx context.Context, code, acceptedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv, ok := m.invites[code]
	if !ok {
		return ErrNotFound
	}
	if inv.Accepted {
		return ErrDuplicate
	}
	inv.Accepted = true
	inv.AcceptedBy = acceptedBy
	return nil
}

// --- Media Store ---

func (m *MemoryDB) StoreMediaFile(ctx context.Context, file *MediaFile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file.CreatedAt = time.Now()
	if file.ScanStatus == "" {
		file.ScanStatus = "pending"
	}
	m.mediaFiles[file.FileID] = file
	return nil
}

func (m *MemoryDB) GetMediaFile(ctx context.Context, fileID string) (*MediaFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, ok := m.mediaFiles[fileID]
	if !ok {
		return nil, ErrNotFound
	}
	return f, nil
}

func (m *MemoryDB) UpdateScanStatus(ctx context.Context, fileID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, ok := m.mediaFiles[fileID]
	if !ok {
		return ErrNotFound
	}
	f.ScanStatus = status
	return nil
}

func (m *MemoryDB) StoreChunk(ctx context.Context, chunk *MediaChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mediaChunks[chunk.FileID] = append(m.mediaChunks[chunk.FileID], chunk)
	return nil
}

func (m *MemoryDB) GetChunks(ctx context.Context, fileID string) ([]*MediaChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chunks := m.mediaChunks[fileID]
	if len(chunks) == 0 {
		return []*MediaChunk{}, nil
	}
	return chunks, nil
}

// --- Notification Store ---

func (m *MemoryDB) RegisterDevice(ctx context.Context, device *UserDevice) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	device.CreatedAt = time.Now()
	m.devices[device.DeviceID] = device
	m.devicesByDID[device.DID] = append(m.devicesByDID[device.DID], device)
	return nil
}

func (m *MemoryDB) GetDevicesByDID(ctx context.Context, did string) ([]*UserDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.devicesByDID[did], nil
}

func (m *MemoryDB) UpdateAPNsToken(ctx context.Context, deviceID, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.devices[deviceID]
	if !ok {
		return ErrNotFound
	}
	d.APNsToken = token
	return nil
}

func (m *MemoryDB) GetNotificationPrefs(ctx context.Context, did string) (*NotificationPrefs, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.notificationPrefs[did]
	if !ok {
		return &NotificationPrefs{
			DID:                  did,
			PushEnabled:          true,
			MessagePreview:       false,
			GroupNotifications:   true,
			ChannelNotifications: true,
		}, nil
	}
	return p, nil
}

func (m *MemoryDB) UpdateNotificationPrefs(ctx context.Context, prefs *NotificationPrefs) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notificationPrefs[prefs.DID] = prefs
	return nil
}

// --- Log Index Store ---

func (m *MemoryDB) StoreLogIndex(ctx context.Context, entry *LogIndexEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry.CreatedAt = time.Now()
	m.logIndex = append(m.logIndex, entry)
	return nil
}

func (m *MemoryDB) GetLogIndex(ctx context.Context, from, to time.Time) ([]*LogIndexEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*LogIndexEntry, 0)
	for _, e := range m.logIndex {
		if !e.TimeRangeFrom.Before(from) && !e.TimeRangeTo.After(to) {
			result = append(result, e)
		}
	}
	return result, nil
}
