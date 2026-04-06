package trustnet

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// BlockchainAnchor represents a trust event anchored to blockchain
type BlockchainAnchor struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"` // "trust_score", "endorsement", "dispute_resolution"
	UserDID         string     `json:"user_did"`
	Timestamp       time.Time  `json:"timestamp"`
	DataHash        string     `json:"data_hash"`                 // ZK-friendly hash of the data
	CardanoTxHash   string     `json:"cardano_tx_hash,omitempty"` // Transaction hash when committed
	MetagraphRef    string     `json:"metagraph_ref,omitempty"`   // Reference in metagraph L1/L0
	ZKProofHash     string     `json:"zk_proof_hash,omitempty"`
	CommittedAt     *time.Time `json:"committed_at,omitempty"`
	VerificationURL string     `json:"verification_url,omitempty"` // URL to verify commitment

	// Metadata for different types
	EndorserDID    string  `json:"endorser_did,omitempty"`    // For endorsements
	TrustScore     float64 `json:"trust_score,omitempty"`     // For trust scores
	EndorsementID  string  `json:"endorsement_id,omitempty"`  // For endorsements
	DisputeID      string  `json:"dispute_id,omitempty"`      // For disputes
	DisputeOutcome string  `json:"dispute_outcome,omitempty"` // "upheld", "dismissed"
}

// ZKCommitment represents a zero-knowledge proof commitment
type ZKCommitment struct {
	ID             string     `json:"id"`
	CommitmentHash string     `json:"commitment_hash"`
	Salt           string     `json:"salt"`
	CreatedAt      time.Time  `json:"created_at"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`

	// The commitment reveals nothing but can be verified
	RevealedValue string `json:"revealed_value,omitempty"` // Only after verification requested
	RevealSalt    string `json:"reveal_salt,omitempty"`    // To verify commitment
}

// CardanoConfig holds configuration for Cardano integration
type CardanoConfig struct {
	Enabled         bool
	NetworkID       string // "mainnet", "testnet", "preview"
	ContractAddress string
	WalletAddress   string
	APIEndpoint     string
	MinFeeLovelace  int64
}

// BlockchainAnchorService manages trust event anchoring
type BlockchainAnchorService struct {
	mu              sync.RWMutex
	config          CardanoConfig
	anchors         map[string]*BlockchainAnchor // by ID
	zkCommitments   map[string]*ZKCommitment     // by ID
	userAnchors     map[string][]string          // userDID -> anchorIDs
	pendingCommits  map[string]*BlockchainAnchor // waiting for blockchain confirmation
	commitmentQueue []string                     // queue of anchor IDs waiting to be committed
}

// NewBlockchainAnchorService creates a new blockchain anchor service
func NewBlockchainAnchorService(config CardanoConfig) *BlockchainAnchorService {
	return &BlockchainAnchorService{
		config:         config,
		anchors:        make(map[string]*BlockchainAnchor),
		zkCommitments:  make(map[string]*ZKCommitment),
		userAnchors:    make(map[string][]string),
		pendingCommits: make(map[string]*BlockchainAnchor),
	}
}

// CreateTrustScoreAnchor creates an anchor for a trust score update
func (bas *BlockchainAnchorService) CreateTrustScoreAnchor(userDID string, score float64, previousScore float64) (*BlockchainAnchor, error) {
	if !bas.config.Enabled {
		return nil, fmt.Errorf("blockchain anchoring disabled")
	}

	bas.mu.Lock()
	defer bas.mu.Unlock()

	data := fmt.Sprintf("score:%s:%.1f:%.1f:%d", userDID, score, previousScore, time.Now().Unix())
	dataHash := bas.hashData(data)

	anchor := &BlockchainAnchor{
		ID:         generateID("anchor_score"),
		Type:       "trust_score",
		UserDID:    userDID,
		TrustScore: score,
		DataHash:   dataHash,
		Timestamp:  time.Now(),
	}

	bas.anchors[anchor.ID] = anchor
	bas.userAnchors[userDID] = append(bas.userAnchors[userDID], anchor.ID)
	bas.commitmentQueue = append(bas.commitmentQueue, anchor.ID)

	return anchor, nil
}

// CreateEndorsementAnchor creates an anchor for an endorsement
func (bas *BlockchainAnchorService) CreateEndorsementAnchor(endorserDID, endorseeDID, endorsementID, category string, weight float64) (*BlockchainAnchor, error) {
	if !bas.config.Enabled {
		return nil, fmt.Errorf("blockchain anchoring disabled")
	}

	bas.mu.Lock()
	defer bas.mu.Unlock()

	data := fmt.Sprintf("endorse:%s:%s:%s:%s:%.1f:%d",
		endorserDID, endorseeDID, endorsementID, category, weight, time.Now().Unix())
	dataHash := bas.hashData(data)

	anchor := &BlockchainAnchor{
		ID:            generateID("anchor_endorse"),
		Type:          "endorsement",
		UserDID:       endorseeDID,
		EndorserDID:   endorserDID,
		EndorsementID: endorsementID,
		DataHash:      dataHash,
		Timestamp:     time.Now(),
	}

	bas.anchors[anchor.ID] = anchor
	bas.userAnchors[endorseeDID] = append(bas.userAnchors[endorseeDID], anchor.ID)
	bas.commitmentQueue = append(bas.commitmentQueue, anchor.ID)

	return anchor, nil
}

// CreateDisputeResolutionAnchor creates an anchor for dispute resolution
func (bas *BlockchainAnchorService) CreateDisputeResolutionAnchor(disputeID, userDID string, outcome string, scoreAdjustment float64) (*BlockchainAnchor, error) {
	if !bas.config.Enabled {
		return nil, fmt.Errorf("blockchain anchoring disabled")
	}

	bas.mu.Lock()
	defer bas.mu.Unlock()

	data := fmt.Sprintf("dispute_resolve:%s:%s:%s:%.1f:%d",
		disputeID, userDID, outcome, scoreAdjustment, time.Now().Unix())
	dataHash := bas.hashData(data)

	anchor := &BlockchainAnchor{
		ID:             generateID("anchor_dispute"),
		Type:           "dispute_resolution",
		UserDID:        userDID,
		DisputeID:      disputeID,
		DisputeOutcome: outcome,
		DataHash:       dataHash,
		Timestamp:      time.Now(),
	}

	bas.anchors[anchor.ID] = anchor
	bas.userAnchors[userDID] = append(bas.userAnchors[userDID], anchor.ID)
	bas.commitmentQueue = append(bas.commitmentQueue, anchor.ID)

	return anchor, nil
}

// CreateZKCommitment creates a zero-knowledge commitment for a value
func (bas *BlockchainAnchorService) CreateZKCommitment(userDID string, value string) (*ZKCommitment, error) {
	bas.mu.Lock()
	defer bas.mu.Unlock()

	salt := generateID("salt")
	commitment := fmt.Sprintf("%s:%s:%s", userDID, value, salt)
	commitmentHash := bas.hashData(commitment)

	zk := &ZKCommitment{
		ID:             generateID("zk_commit"),
		CommitmentHash: commitmentHash,
		Salt:           salt,
		CreatedAt:      time.Now(),
	}

	bas.zkCommitments[zk.ID] = zk
	return zk, nil
}

// VerifyZKCommitment verifies a ZK commitment by revealing the original value
func (bas *BlockchainAnchorService) VerifyZKCommitment(commitmentID string, revealValue string, revealSalt string) (bool, error) {
	bas.mu.Lock()
	defer bas.mu.Unlock()

	zk, exists := bas.zkCommitments[commitmentID]
	if !exists {
		return false, fmt.Errorf("commitment not found")
	}

	// Reconstruct the hash
	reconstructed := fmt.Sprintf("%s:%s:%s", "", revealValue, revealSalt)
	reconstructedHash := bas.hashData(reconstructed)

	if reconstructedHash == zk.CommitmentHash {
		now := time.Now()
		zk.VerifiedAt = &now
		zk.RevealedValue = revealValue
		zk.RevealSalt = revealSalt
		return true, nil
	}

	return false, fmt.Errorf("commitment verification failed")
}

// CommitBatch simulates committing pending anchors to blockchain
// In a real implementation, this would submit a Cardano transaction
func (bas *BlockchainAnchorService) CommitBatch() (map[string]string, error) {
	if !bas.config.Enabled {
		return nil, fmt.Errorf("blockchain anchoring disabled")
	}

	bas.mu.Lock()
	defer bas.mu.Unlock()

	if len(bas.commitmentQueue) == 0 {
		return make(map[string]string), nil
	}

	results := make(map[string]string)

	// In production, this would submit to Cardano
	for _, anchorID := range bas.commitmentQueue {
		anchor := bas.anchors[anchorID]
		if anchor == nil {
			continue
		}

		// Simulate blockchain commitment
		txHash := fmt.Sprintf("tx_%s_%d", anchorID, time.Now().UnixNano())
		metagraphRef := fmt.Sprintf("mg_ref_%s", anchorID)

		anchor.CardanoTxHash = txHash
		anchor.MetagraphRef = metagraphRef
		now := time.Now()
		anchor.CommittedAt = &now
		anchor.VerificationURL = fmt.Sprintf("https://cardano.example.com/tx/%s", txHash)

		results[anchorID] = txHash
	}

	bas.commitmentQueue = []string{} // Clear queue

	return results, nil
}

// GetAnchor retrieves an anchor by ID
func (bas *BlockchainAnchorService) GetAnchor(anchorID string) (*BlockchainAnchor, error) {
	bas.mu.RLock()
	defer bas.mu.RUnlock()

	anchor, exists := bas.anchors[anchorID]
	if !exists {
		return nil, fmt.Errorf("anchor not found")
	}

	return anchor, nil
}

// GetUserAnchors retrieves all anchors for a user
func (bas *BlockchainAnchorService) GetUserAnchors(userDID string) []*BlockchainAnchor {
	bas.mu.RLock()
	defer bas.mu.RUnlock()

	anchorIDs := bas.userAnchors[userDID]
	result := make([]*BlockchainAnchor, 0, len(anchorIDs))

	for _, id := range anchorIDs {
		if anchor, exists := bas.anchors[id]; exists {
			result = append(result, anchor)
		}
	}

	return result
}

// GetUserAnchorsByType retrieves anchors of a specific type for a user
func (bas *BlockchainAnchorService) GetUserAnchorsByType(userDID string, typee string) []*BlockchainAnchor {
	bas.mu.RLock()
	defer bas.mu.RUnlock()

	anchorIDs := bas.userAnchors[userDID]
	result := make([]*BlockchainAnchor, 0)

	for _, id := range anchorIDs {
		if anchor, exists := bas.anchors[id]; exists && anchor.Type == typee {
			result = append(result, anchor)
		}
	}

	return result
}

// GetPendingCommits returns anchors waiting for blockchain commitment
func (bas *BlockchainAnchorService) GetPendingCommits() []*BlockchainAnchor {
	bas.mu.RLock()
	defer bas.mu.RUnlock()

	result := make([]*BlockchainAnchor, 0, len(bas.commitmentQueue))

	for _, id := range bas.commitmentQueue {
		if anchor, exists := bas.anchors[id]; exists {
			result = append(result, anchor)
		}
	}

	return result
}

// hashData creates a hash of data (ZK-friendly)
func (bas *BlockchainAnchorService) hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Helper function to generate unique IDs
func generateID(prefix string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), time.Now().UnixNano())))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:8]))
}
