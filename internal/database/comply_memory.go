package database

import (
	"context"
	"fmt"
	"time"
)

// --- MemoryDB ComplyStore (WO-250) ---

func (m *MemoryDB) CreateRetentionPolicy(_ context.Context, p *RetentionPolicy) error {
	if p == nil || p.ID == "" || p.OrgDID == "" || !p.PolicyType.Valid() {
		return fmt.Errorf("invalid retention policy")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyPolicies == nil {
		m.complyPolicies = make(map[string]*RetentionPolicy)
	}
	stored := *p
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = time.Now()
	}
	if stored.EffectiveAt.IsZero() {
		stored.EffectiveAt = stored.CreatedAt
	}
	m.complyPolicies[p.ID] = &stored
	return nil
}

func (m *MemoryDB) ListRetentionPolicies(_ context.Context, orgDID string, activeOnly bool) ([]*RetentionPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*RetentionPolicy
	for _, p := range m.complyPolicies {
		if p.OrgDID != orgDID {
			continue
		}
		if activeOnly && !p.Active {
			continue
		}
		copy := *p
		out = append(out, &copy)
	}
	return out, nil
}

func (m *MemoryDB) GetRetentionPolicy(_ context.Context, policyID string) (*RetentionPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.complyPolicies[policyID]
	if !ok {
		return nil, ErrComplyPolicyNotFound
	}
	copy := *p
	return &copy, nil
}

func (m *MemoryDB) BindConversationPolicy(_ context.Context, b *ConversationPolicyBinding) error {
	if b == nil || b.ConversationID == "" || b.PolicyID == "" {
		return fmt.Errorf("invalid conversation policy binding")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyBindings == nil {
		m.complyBindings = make(map[string]*ConversationPolicyBinding)
	}
	stored := *b
	if stored.BoundAt.IsZero() {
		stored.BoundAt = time.Now()
	}
	m.complyBindings[b.ConversationID] = &stored
	return nil
}

func (m *MemoryDB) GetConversationBinding(_ context.Context, conversationID string) (*ConversationPolicyBinding, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.complyBindings[conversationID]
	if !ok {
		return nil, ErrComplyPolicyNotFound
	}
	copy := *b
	return &copy, nil
}

func (m *MemoryDB) UnbindConversation(_ context.Context, conversationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.complyBindings, conversationID)
	return nil
}

func (m *MemoryDB) UpsertOrgProfile(_ context.Context, profile *ComplyOrgProfile) error {
	if profile == nil || profile.OrgDID == "" {
		return fmt.Errorf("invalid org profile")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyOrgs == nil {
		m.complyOrgs = make(map[string]*ComplyOrgProfile)
	}
	stored := *profile
	if stored.UpdatedAt.IsZero() {
		stored.UpdatedAt = time.Now()
	}
	if stored.Tier == "" {
		stored.Tier = "starter"
	}
	if stored.Seats <= 0 {
		stored.Seats = 10
	}
	m.complyOrgs[profile.OrgDID] = &stored
	return nil
}

func (m *MemoryDB) GetOrgProfile(_ context.Context, orgDID string) (*ComplyOrgProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.complyOrgs[orgDID]
	if !ok {
		return nil, ErrComplyPolicyNotFound
	}
	copy := *p
	return &copy, nil
}

func (m *MemoryDB) CountActivePolicies(_ context.Context, orgDID string, policyType RetentionPolicyType) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, p := range m.complyPolicies {
		if p.OrgDID == orgDID && p.Active && (policyType == "" || p.PolicyType == policyType) {
			n++
		}
	}
	return n, nil
}
