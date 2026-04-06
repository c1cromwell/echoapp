package trustnet

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// RequestStatus represents the state of a contact request
type RequestStatus string

const (
	RequestPending  RequestStatus = "pending"
	RequestAccepted RequestStatus = "accepted"
	RequestDeclined RequestStatus = "declined"
	RequestCancelled RequestStatus = "cancelled"
)

// ContactRequest represents a contact request from one user to another
type ContactRequest struct {
	ID             string
	FromDID        string
	ToDID          string
	Message        string // optional message with the request
	Status         RequestStatus
	SuggestedTier  CircleTier // the tier the accepter chooses
	MutualCount    int        // computed at request time
	FromTrustScore int
	CreatedAt      time.Time
	RespondedAt    *time.Time
}

// ContactRequestService manages contact requests
type ContactRequestService struct {
	mu       sync.RWMutex
	requests map[string]*ContactRequest // requestID -> request
	byFrom   map[string][]string        // fromDID -> []requestID (sent)
	byTo     map[string][]string        // toDID -> []requestID (received)
	circles  *CircleService
}

// NewContactRequestService creates a new contact request service
func NewContactRequestService(circles *CircleService) *ContactRequestService {
	return &ContactRequestService{
		requests: make(map[string]*ContactRequest),
		byFrom:   make(map[string][]string),
		byTo:     make(map[string][]string),
		circles:  circles,
	}
}

// SendRequest sends a contact request from one user to another
func (s *ContactRequestService) SendRequest(fromDID, toDID, message string, trustScore int) (*ContactRequest, error) {
	if fromDID == toDID {
		return nil, ErrRequestToSelf
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already contacts
	if s.circles.HasContact(fromDID, toDID) {
		return nil, ErrAlreadyContacts
	}

	// Check for existing pending request in either direction
	for _, id := range s.byFrom[fromDID] {
		r := s.requests[id]
		if r.ToDID == toDID && r.Status == RequestPending {
			return nil, ErrRequestAlreadySent
		}
	}
	for _, id := range s.byFrom[toDID] {
		r := s.requests[id]
		if r.ToDID == fromDID && r.Status == RequestPending {
			return nil, ErrRequestAlreadySent
		}
	}

	mutualCount := s.circles.CountMutualContacts(fromDID, toDID)

	req := &ContactRequest{
		ID:             uuid.New().String(),
		FromDID:        fromDID,
		ToDID:          toDID,
		Message:        message,
		Status:         RequestPending,
		MutualCount:    mutualCount,
		FromTrustScore: trustScore,
		CreatedAt:      time.Now(),
	}

	s.requests[req.ID] = req
	s.byFrom[fromDID] = append(s.byFrom[fromDID], req.ID)
	s.byTo[toDID] = append(s.byTo[toDID], req.ID)

	return req, nil
}

// AcceptRequest accepts a contact request and adds the contact to a circle
func (s *ContactRequestService) AcceptRequest(requestID string, tier CircleTier) (*CircleContact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return nil, ErrRequestNotFound
	}

	if req.Status != RequestPending {
		return nil, ErrRequestAlreadyHandled
	}

	// Add as contact in both directions
	contact, err := s.circles.AddContact(req.ToDID, req.FromDID, tier)
	if err != nil {
		return nil, err
	}

	// Also add the reverse direction as acquaintance
	s.circles.AddContact(req.FromDID, req.ToDID, CircleAcquaintance)

	now := time.Now()
	req.Status = RequestAccepted
	req.SuggestedTier = tier
	req.RespondedAt = &now

	return contact, nil
}

// DeclineRequest declines a contact request
func (s *ContactRequestService) DeclineRequest(requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return ErrRequestNotFound
	}

	if req.Status != RequestPending {
		return ErrRequestAlreadyHandled
	}

	now := time.Now()
	req.Status = RequestDeclined
	req.RespondedAt = &now
	return nil
}

// CancelRequest cancels a sent contact request
func (s *ContactRequestService) CancelRequest(requestID, fromDID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.requests[requestID]
	if !ok {
		return ErrRequestNotFound
	}

	if req.FromDID != fromDID {
		return ErrRequestNotFound
	}

	if req.Status != RequestPending {
		return ErrRequestAlreadyHandled
	}

	now := time.Now()
	req.Status = RequestCancelled
	req.RespondedAt = &now
	return nil
}

// GetRequest retrieves a request by ID
func (s *ContactRequestService) GetRequest(requestID string) (*ContactRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.requests[requestID]
	if !ok {
		return nil, ErrRequestNotFound
	}
	return req, nil
}

// GetPendingReceived returns pending requests received by a user
func (s *ContactRequestService) GetPendingReceived(toDID string) []*ContactRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ContactRequest
	for _, id := range s.byTo[toDID] {
		r := s.requests[id]
		if r.Status == RequestPending {
			result = append(result, r)
		}
	}
	return result
}

// GetPendingSent returns pending requests sent by a user
func (s *ContactRequestService) GetPendingSent(fromDID string) []*ContactRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ContactRequest
	for _, id := range s.byFrom[fromDID] {
		r := s.requests[id]
		if r.Status == RequestPending {
			result = append(result, r)
		}
	}
	return result
}

// CountPendingReceived returns the number of pending received requests
func (s *ContactRequestService) CountPendingReceived(toDID string) int {
	return len(s.GetPendingReceived(toDID))
}

// IsLowTrustWarning returns true if the sender has a low trust score
func IsLowTrustWarning(req *ContactRequest) bool {
	return req.FromTrustScore < 20
}
