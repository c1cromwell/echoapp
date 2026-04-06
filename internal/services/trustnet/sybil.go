package trustnet

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"sync"
	"time"
)

// SybilRiskLevel represents the assessed risk of an account being a sybil
type SybilRiskLevel string

const (
	SybilRiskNone   SybilRiskLevel = "none"
	SybilRiskLow    SybilRiskLevel = "low"
	SybilRiskMedium SybilRiskLevel = "medium"
	SybilRiskHigh   SybilRiskLevel = "high"
)

const (
	// ClusterThreshold is the minimum similarity score to flag a cluster
	ClusterThreshold = 0.7

	// MinGraphConnections is the minimum unique connections to not be flagged
	MinGraphConnections = 3

	// MaxAccountsPerDevice is the max accounts allowed from one device fingerprint
	MaxAccountsPerDevice = 2
)

// DeviceFingerprint represents a device identity signal
type DeviceFingerprint struct {
	UserDID       string
	FingerprintHash string
	IPAddress     string
	UserAgent     string
	CreatedAt     time.Time
}

// SybilAssessment is the result of a sybil check
type SybilAssessment struct {
	UserDID         string
	RiskLevel       SybilRiskLevel
	Score           float64 // 0.0 (safe) to 1.0 (definitely sybil)
	Flags           []string
	ClusterID       string // empty if no cluster detected
	AssessedAt      time.Time
}

// AccountCluster represents a group of potentially linked accounts
type AccountCluster struct {
	ID        string
	Members   []string // user DIDs
	Signals   []string // what triggered the cluster
	DetectedAt time.Time
	Reviewed  bool
}

// SybilService detects and manages sybil accounts
type SybilService struct {
	mu           sync.RWMutex
	fingerprints map[string]*DeviceFingerprint // userDID -> fingerprint
	byDevice     map[string][]string           // fingerprintHash -> []userDID
	byIP         map[string][]string           // ipAddress -> []userDID
	clusters     map[string]*AccountCluster    // clusterID -> cluster
	assessments  map[string]*SybilAssessment   // userDID -> latest assessment
	connections  map[string]map[string]bool     // userDID -> set of connected DIDs
}

// NewSybilService creates a new sybil detection service
func NewSybilService() *SybilService {
	return &SybilService{
		fingerprints: make(map[string]*DeviceFingerprint),
		byDevice:     make(map[string][]string),
		byIP:         make(map[string][]string),
		clusters:     make(map[string]*AccountCluster),
		assessments:  make(map[string]*SybilAssessment),
		connections:  make(map[string]map[string]bool),
	}
}

// RegisterDevice registers a device fingerprint for a user
func (s *SybilService) RegisterDevice(userDID, fingerprint, ipAddress, userAgent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fpHash := hashFingerprint(fingerprint)

	// Check max accounts per device
	existing := s.byDevice[fpHash]
	for _, did := range existing {
		if did == userDID {
			// Already registered, update
			s.fingerprints[userDID] = &DeviceFingerprint{
				UserDID:         userDID,
				FingerprintHash: fpHash,
				IPAddress:       ipAddress,
				UserAgent:       userAgent,
				CreatedAt:       time.Now(),
			}
			return nil
		}
	}

	if len(existing) >= MaxAccountsPerDevice {
		return ErrDeviceClusterDetected
	}

	fp := &DeviceFingerprint{
		UserDID:         userDID,
		FingerprintHash: fpHash,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		CreatedAt:       time.Now(),
	}

	s.fingerprints[userDID] = fp
	s.byDevice[fpHash] = append(s.byDevice[fpHash], userDID)
	s.byIP[ipAddress] = append(s.byIP[ipAddress], userDID)

	return nil
}

// RecordConnection records a social graph connection between users
func (s *SybilService) RecordConnection(userDID, connectedDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connections[userDID] == nil {
		s.connections[userDID] = make(map[string]bool)
	}
	s.connections[userDID][connectedDID] = true

	if s.connections[connectedDID] == nil {
		s.connections[connectedDID] = make(map[string]bool)
	}
	s.connections[connectedDID][userDID] = true
}

// AssessUser runs a sybil assessment on a user
func (s *SybilService) AssessUser(userDID string) *SybilAssessment {
	s.mu.Lock()
	defer s.mu.Unlock()

	assessment := &SybilAssessment{
		UserDID:    userDID,
		RiskLevel:  SybilRiskNone,
		Score:      0,
		AssessedAt: time.Now(),
	}

	var riskScore float64

	// Check device clustering
	if fp, ok := s.fingerprints[userDID]; ok {
		deviceAccounts := s.byDevice[fp.FingerprintHash]
		if len(deviceAccounts) > 1 {
			riskScore += 0.3
			assessment.Flags = append(assessment.Flags, "shared_device")
		}

		// Check IP clustering
		ipAccounts := s.byIP[fp.IPAddress]
		if len(ipAccounts) > 2 {
			riskScore += 0.2
			assessment.Flags = append(assessment.Flags, "shared_ip")
		}
	}

	// Check social graph sparseness
	connections := s.connections[userDID]
	if len(connections) < MinGraphConnections {
		riskScore += 0.2
		assessment.Flags = append(assessment.Flags, "sparse_graph")
	}

	// Check for endorsement ring patterns
	if s.hasEndorsementRing(userDID) {
		riskScore += 0.3
		assessment.Flags = append(assessment.Flags, "endorsement_ring")
	}

	// Clamp to 0-1
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	assessment.Score = math.Round(riskScore*100) / 100

	switch {
	case riskScore >= 0.7:
		assessment.RiskLevel = SybilRiskHigh
	case riskScore >= 0.4:
		assessment.RiskLevel = SybilRiskMedium
	case riskScore >= 0.2:
		assessment.RiskLevel = SybilRiskLow
	default:
		assessment.RiskLevel = SybilRiskNone
	}

	s.assessments[userDID] = assessment
	return assessment
}

// hasEndorsementRing detects if a user is part of a mutual endorsement ring
// (all members only endorse each other)
func (s *SybilService) hasEndorsementRing(userDID string) bool {
	conns := s.connections[userDID]
	if len(conns) < 2 {
		return false
	}

	// Check if all connections also connect to each other (clique)
	members := make([]string, 0, len(conns)+1)
	members = append(members, userDID)
	for did := range conns {
		members = append(members, did)
	}

	if len(members) > 5 {
		// Only check small groups for ring patterns
		return false
	}

	// Check if it's a complete subgraph
	for i, a := range members {
		for j, b := range members {
			if i == j {
				continue
			}
			if !s.connections[a][b] {
				return false
			}
		}
	}

	return true
}

// GetAssessment returns the latest sybil assessment for a user
func (s *SybilService) GetAssessment(userDID string) *SybilAssessment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assessments[userDID]
}

// DetectClusters scans all fingerprints for device clusters
func (s *SybilService) DetectClusters() []*AccountCluster {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newClusters []*AccountCluster

	// Check device-based clusters
	for fpHash, users := range s.byDevice {
		if len(users) <= 1 {
			continue
		}

		clusterID := "device-" + fpHash[:8]
		if _, exists := s.clusters[clusterID]; exists {
			continue
		}

		cluster := &AccountCluster{
			ID:         clusterID,
			Members:    make([]string, len(users)),
			Signals:    []string{"shared_device_fingerprint"},
			DetectedAt: time.Now(),
		}
		copy(cluster.Members, users)
		s.clusters[clusterID] = cluster
		newClusters = append(newClusters, cluster)
	}

	// Check IP-based clusters (more than 3 accounts)
	for ip, users := range s.byIP {
		if len(users) <= 3 {
			continue
		}

		clusterID := "ip-" + hashFingerprint(ip)[:8]
		if _, exists := s.clusters[clusterID]; exists {
			continue
		}

		cluster := &AccountCluster{
			ID:         clusterID,
			Members:    make([]string, len(users)),
			Signals:    []string{"shared_ip_address"},
			DetectedAt: time.Now(),
		}
		copy(cluster.Members, users)
		s.clusters[clusterID] = cluster
		newClusters = append(newClusters, cluster)
	}

	return newClusters
}

// GetCluster retrieves a cluster by ID
func (s *SybilService) GetCluster(clusterID string) *AccountCluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clusters[clusterID]
}

// GetAllClusters returns all detected clusters
func (s *SybilService) GetAllClusters() []*AccountCluster {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*AccountCluster
	for _, c := range s.clusters {
		result = append(result, c)
	}
	return result
}

// MarkClusterReviewed marks a cluster as manually reviewed
func (s *SybilService) MarkClusterReviewed(clusterID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clusters[clusterID]; ok {
		c.Reviewed = true
	}
}

// GetConnectionCount returns the number of unique connections for a user
func (s *SybilService) GetConnectionCount(userDID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections[userDID])
}

// EndorsementWeightPenalty returns a multiplier (0-1) for endorsement weight
// based on whether the endorser is in a suspected cluster
func (s *SybilService) EndorsementWeightPenalty(endorserDID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assessment := s.assessments[endorserDID]
	if assessment == nil {
		return 1.0
	}

	switch assessment.RiskLevel {
	case SybilRiskHigh:
		return 0.1
	case SybilRiskMedium:
		return 0.5
	case SybilRiskLow:
		return 0.8
	default:
		return 1.0
	}
}

func hashFingerprint(fp string) string {
	h := sha256.Sum256([]byte(fp))
	return hex.EncodeToString(h[:])
}
