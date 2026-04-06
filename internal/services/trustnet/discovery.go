package trustnet

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// QRCodeExpiry is how long a QR code is valid
	QRCodeExpiry = 15 * time.Minute

	// QRCodePrefix identifies Echo QR codes
	QRCodePrefix = "echo://"

	// MinSearchLength is the minimum query length for username search
	MinSearchLength = 2
)

// UserProfile represents a searchable user profile in the discovery system
type UserProfile struct {
	DID          string
	Username     string
	DisplayName  string
	TrustScore   int
	Verified     bool
	CreatedAt    time.Time
}

// QRCode represents a generated QR code payload
type QRCode struct {
	ID        string
	UserDID   string
	Payload   string // the encoded QR content
	Signature string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// QRPayload is the data encoded in the QR code
type QRPayload struct {
	DID       string `json:"did"`
	Username  string `json:"username"`
	Timestamp int64  `json:"ts"`
	Nonce     string `json:"nonce"`
}

// SearchResult represents a user found via search
type SearchResult struct {
	DID         string
	Username    string
	DisplayName string
	TrustScore  int
	Verified    bool
	MutualCount int
}

// DiscoveryService handles contact discovery via QR codes, username, and DID search
type DiscoveryService struct {
	mu       sync.RWMutex
	profiles map[string]*UserProfile // DID -> profile
	byName   map[string]string       // username (lowercase) -> DID
	qrCodes  map[string]*QRCode      // qrID -> code
	circles  *CircleService
	hmacKey  []byte
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(circles *CircleService, hmacKey []byte) *DiscoveryService {
	if len(hmacKey) == 0 {
		hmacKey = []byte("default-dev-key-change-in-production")
	}
	return &DiscoveryService{
		profiles: make(map[string]*UserProfile),
		byName:   make(map[string]string),
		qrCodes:  make(map[string]*QRCode),
		circles:  circles,
		hmacKey:  hmacKey,
	}
}

// RegisterProfile registers or updates a user profile for discovery
func (s *DiscoveryService) RegisterProfile(did, username, displayName string, trustScore int, verified bool) (*UserProfile, error) {
	if did == "" || username == "" {
		return nil, ErrUserNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check username uniqueness (skip if same user updating)
	lowerName := strings.ToLower(username)
	if existingDID, taken := s.byName[lowerName]; taken && existingDID != did {
		return nil, ErrPersonaDuplicate // username taken
	}

	// Remove old username mapping if updating
	if existing, ok := s.profiles[did]; ok {
		delete(s.byName, strings.ToLower(existing.Username))
	}

	profile := &UserProfile{
		DID:         did,
		Username:    username,
		DisplayName: displayName,
		TrustScore:  trustScore,
		Verified:    verified,
		CreatedAt:   time.Now(),
	}

	s.profiles[did] = profile
	s.byName[lowerName] = did
	return profile, nil
}

// GetProfile returns a user profile by DID
func (s *DiscoveryService) GetProfile(did string) (*UserProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.profiles[did]
	if !ok {
		return nil, ErrUserNotFound
	}
	return p, nil
}

// GenerateQRCode generates a QR code payload for a user
func (s *DiscoveryService) GenerateQRCode(userDID, username string) (*QRCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, ok := s.profiles[userDID]
	if !ok {
		return nil, ErrUserNotFound
	}

	nonce := uuid.New().String()[:8]
	payload := QRPayload{
		DID:       userDID,
		Username:  profile.Username,
		Timestamp: time.Now().Unix(),
		Nonce:     nonce,
	}

	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.URLEncoding.EncodeToString(payloadBytes)
	signature := s.sign(payloadBytes)

	qr := &QRCode{
		ID:        uuid.New().String(),
		UserDID:   userDID,
		Payload:   QRCodePrefix + encoded,
		Signature: signature,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(QRCodeExpiry),
	}

	s.qrCodes[qr.ID] = qr
	return qr, nil
}

// ParseQRCode parses and validates a scanned QR code
func (s *DiscoveryService) ParseQRCode(raw string) (*UserProfile, error) {
	if !strings.HasPrefix(raw, QRCodePrefix) {
		return nil, ErrQRCodeInvalid
	}

	encoded := strings.TrimPrefix(raw, QRCodePrefix)
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrQRCodeInvalid
	}

	var payload QRPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, ErrQRCodeInvalid
	}

	// Check expiry (15 minute window)
	created := time.Unix(payload.Timestamp, 0)
	if time.Since(created) > QRCodeExpiry {
		return nil, ErrQRCodeExpired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.profiles[payload.DID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return profile, nil
}

// SearchByUsername searches for users by username prefix
func (s *DiscoveryService) SearchByUsername(query, callerDID string) []SearchResult {
	if len(query) < MinSearchLength {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	lowerQuery := strings.ToLower(query)
	var results []SearchResult

	for _, profile := range s.profiles {
		if profile.DID == callerDID {
			continue // don't show self in results
		}

		lowerUsername := strings.ToLower(profile.Username)
		lowerDisplay := strings.ToLower(profile.DisplayName)

		if strings.Contains(lowerUsername, lowerQuery) || strings.Contains(lowerDisplay, lowerQuery) {
			mutualCount := s.circles.CountMutualContacts(callerDID, profile.DID)
			results = append(results, SearchResult{
				DID:         profile.DID,
				Username:    profile.Username,
				DisplayName: profile.DisplayName,
				TrustScore:  profile.TrustScore,
				Verified:    profile.Verified,
				MutualCount: mutualCount,
			})
		}
	}

	// Sort: exact matches first, then verified, then by trust score
	sortSearchResults(results, lowerQuery)
	return results
}

// SearchByDID looks up a user by their exact DID
func (s *DiscoveryService) SearchByDID(did, callerDID string) (*SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.profiles[did]
	if !ok {
		return nil, ErrUserNotFound
	}

	mutualCount := s.circles.CountMutualContacts(callerDID, did)
	return &SearchResult{
		DID:         profile.DID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		TrustScore:  profile.TrustScore,
		Verified:    profile.Verified,
		MutualCount: mutualCount,
	}, nil
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// ValidateUsername checks if a username is valid
func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// IsUsernameAvailable checks if a username is available
func (s *DiscoveryService) IsUsernameAvailable(username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, taken := s.byName[strings.ToLower(username)]
	return !taken
}

func (s *DiscoveryService) sign(data []byte) string {
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write(data)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// sortSearchResults sorts results: exact username match first, then verified, then by trust
func sortSearchResults(results []SearchResult, query string) {
	// Simple insertion sort (results typically small)
	for i := 1; i < len(results); i++ {
		j := i
		for j > 0 && searchResultLess(results[j], results[j-1], query) {
			results[j], results[j-1] = results[j-1], results[j]
			j--
		}
	}
}

func searchResultLess(a, b SearchResult, query string) bool {
	aExact := strings.ToLower(a.Username) == query
	bExact := strings.ToLower(b.Username) == query
	if aExact != bExact {
		return aExact
	}
	if a.Verified != b.Verified {
		return a.Verified
	}
	return a.TrustScore > b.TrustScore
}
