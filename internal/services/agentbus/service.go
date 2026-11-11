// Package agentbus is the headless messaging bus: scoped grants, sealed
// delivery, and audit metadata. It never stores plaintext or identity keys.
package agentbus

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	VerbSend         = "send"
	VerbReadMetadata = "read_metadata"
	VerbSearch       = "search"
	VerbPay          = "pay"
	VerbForward      = "forward"
	VerbRotateKey    = "rotate_key"
)

// Irreversible verbs are not relayed until the user passkey confirms them.
func Irreversible(verb string) bool {
	switch verb {
	case VerbPay, VerbForward, VerbRotateKey:
		return true
	default:
		return false
	}
}

var (
	ErrNotFound      = errors.New("not found")
	ErrExpired       = errors.New("grant expired")
	ErrVerb          = errors.New("verb not granted")
	ErrThread        = errors.New("thread not granted")
	ErrPlaintext     = errors.New("plaintext refused")
	ErrBadHint       = errors.New("render hint refused")
	ErrReplay        = errors.New("challenge already used")
	ErrNoReconstruct = errors.New("recovery reconstruction is client-only")
)

// Grant is a scoped capability. Token is the secret the agent presents.
// The identity key is never copied into a grant.
type Grant struct {
	ID          string
	Token       string
	UserDID     string
	AgentDID    string
	AgentName   string
	Verb        string
	ThreadID    string
	ExpiresAt   time.Time
	EmitSignals bool
}

// AuditRow records what an agent did. It has no message body.
type AuditRow struct {
	AgentDID string
	Verb     string
	ThreadID string
	At       time.Time
}

// Delivery is routing metadata beside an opaque blob.
type Delivery struct {
	To         string
	ThreadID   string
	Ciphertext []byte
	ExpiresAt  string
	RenderHint string
	Untrusted  bool
	FromUser   string
	AgentDID   string
}

// Challenge is a one-time user confirmation. Payload is opaque.
type Challenge struct {
	ID         string
	UserDID    string
	AgentDID   string
	Verb       string
	To         string
	ThreadID   string
	Ciphertext []byte
	ExpiresAt  string
	RenderHint string
	Used       bool
}

// ShareMeta is who holds a recovery share. Share bytes are never stored.
type ShareMeta struct {
	Index     int
	HolderDID string
	Status    string
}

type Service struct {
	mu         sync.Mutex
	keys       map[string]*ecdsa.PublicKey
	grants     map[string]*Grant // token -> grant
	byID       map[string]*Grant
	byUser     map[string][]string
	audit      []AuditRow
	auditLines []string
	queue      []Delivery
	contacts   map[string]map[string]bool
	challenges map[string]*Challenge
	shares     []ShareMeta
	keyGen     map[string]int
	now        func() time.Time
}

func New() *Service {
	return &Service{
		keys:       map[string]*ecdsa.PublicKey{},
		grants:     map[string]*Grant{},
		byID:       map[string]*Grant{},
		byUser:     map[string][]string{},
		contacts:   map[string]map[string]bool{},
		challenges: map[string]*Challenge{},
		keyGen:     map[string]int{},
		now:        time.Now,
	}
}

func (s *Service) RegisterDeviceKey(userDID string, pub *ecdsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[userDID] = pub
}

func (s *Service) DeviceKey(userDID string) *ecdsa.PublicKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.keys[userDID]
}

func DirectThread(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return "dm:" + a + ":" + b
}

func (s *Service) Issue(g Grant) (*Grant, error) {
	if g.UserDID == "" || g.AgentDID == "" || g.Verb == "" {
		return nil, ErrNotFound
	}
	if g.ThreadID == "" {
		g.ThreadID = "*"
	}
	if g.ExpiresAt.IsZero() {
		return nil, ErrExpired
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	g.Token = hex.EncodeToString(buf)
	id := make([]byte, 8)
	_, _ = rand.Read(id)
	g.ID = hex.EncodeToString(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := g
	s.grants[stored.Token] = &stored
	s.byID[stored.ID] = &stored
	s.byUser[stored.UserDID] = append(s.byUser[stored.UserDID], stored.ID)
	return &stored, nil
}

func (s *Service) Lookup(token string, verb string, now time.Time) (*Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.grants[token]
	if !ok {
		return nil, ErrNotFound
	}
	if !g.ExpiresAt.After(now) {
		return nil, ErrExpired
	}
	if g.Verb != verb {
		return nil, ErrVerb
	}
	return g, nil
}

func (s *Service) List(userDID string) []Grant {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Grant
	for _, id := range s.byUser[userDID] {
		if g, ok := s.byID[id]; ok {
			copy := *g
			copy.Token = ""
			out = append(out, copy)
		}
	}
	return out
}

func (s *Service) Revoke(userDID, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.byID[id]
	if !ok || g.UserDID != userDID {
		return false
	}
	delete(s.grants, g.Token)
	delete(s.byID, id)
	return true
}

func (s *Service) AllowsThread(g *Grant, threadID string) bool {
	return g.ThreadID == "*" || g.ThreadID == threadID
}

func (s *Service) MarkContact(owner, peer string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.contacts[owner] == nil {
		s.contacts[owner] = map[string]bool{}
	}
	s.contacts[owner][peer] = true
}

func (s *Service) IsContact(owner, peer string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.contacts[owner][peer]
}

func ValidHint(hint string) bool {
	switch hint {
	case "", "text", "spoken", "card", "silent":
		return true
	default:
		return false
	}
}

func (s *Service) Enqueue(d Delivery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, d)
	s.appendAudit(d.AgentDID, "send", d.ThreadID)
}

func (s *Service) Queue() []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Delivery, len(s.queue))
	copy(out, s.queue)
	return out
}

func (s *Service) appendAudit(agent, verb, thread string) {
	row := AuditRow{AgentDID: agent, Verb: verb, ThreadID: thread, At: s.now()}
	s.audit = append(s.audit, row)
	// Metadata only. Ciphertext is intentionally absent.
	line := "agent_audit agent=" + agent + " verb=" + verb + " thread=" + thread
	s.auditLines = append(s.auditLines, line)
}

func (s *Service) AuditLines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.auditLines))
	copy(out, s.auditLines)
	return out
}

func (s *Service) OpenChallenge(c Challenge) (*Challenge, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	c.ID = hex.EncodeToString(buf)
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := c
	s.challenges[stored.ID] = &stored
	return &stored, nil
}

func (s *Service) ConsumeChallenge(id, userDID string) (*Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.challenges[id]
	if !ok || c.UserDID != userDID {
		return nil, ErrNotFound
	}
	if c.Used {
		return nil, ErrReplay
	}
	c.Used = true
	if c.Verb == VerbRotateKey {
		s.keyGen[userDID]++
		s.appendAudit(c.AgentDID, c.Verb, c.ThreadID)
		return c, nil
	}
	s.queue = append(s.queue, Delivery{
		To:         c.To,
		ThreadID:   c.ThreadID,
		Ciphertext: append([]byte(nil), c.Ciphertext...),
		ExpiresAt:  c.ExpiresAt,
		RenderHint: c.RenderHint,
		Untrusted:  !s.contacts[c.To][userDID],
		FromUser:   userDID,
		AgentDID:   c.AgentDID,
	})
	s.appendAudit(c.AgentDID, c.Verb, c.ThreadID)
	return c, nil
}

func (s *Service) KeyGeneration(userDID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.keyGen[userDID]
}

func (s *Service) PutShareMeta(m ShareMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shares = append(s.shares, m)
}

func (s *Service) ShareMeta() []ShareMeta {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ShareMeta, len(s.shares))
	copy(out, s.shares)
	return out
}

// Reconstruct refuses to run. Shamir combine stays on the client.
func (s *Service) Reconstruct() error {
	return ErrNoReconstruct
}

func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "@")
	return strings.ToLower(strings.TrimSpace(name))
}
