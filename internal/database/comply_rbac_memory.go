package database

import (
	"context"
	"time"
)

func (m *MemoryDB) UpsertOrgMember(_ context.Context, member *OrgMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyMembers == nil {
		m.complyMembers = make(map[string]*OrgMember)
	}
	stored := *member
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = time.Now().UTC()
	}
	m.complyMembers[member.OrgDID+"|"+member.MemberDID] = &stored
	return nil
}

func (m *MemoryDB) GetOrgMember(_ context.Context, orgDID, memberDID string) (*OrgMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mem, ok := m.complyMembers[orgDID+"|"+memberDID]
	if !ok {
		return nil, ErrComplyOrgForbidden
	}
	copy := *mem
	return &copy, nil
}

func (m *MemoryDB) RecordDEFingerprint(_ context.Context, rec *DEFingerprintRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyDEFP == nil {
		m.complyDEFP = make(map[string]*DEFingerprintRecord)
	}
	stored := *rec
	if stored.RecordedAt.IsZero() {
		stored.RecordedAt = time.Now().UTC()
	}
	m.complyDEFP[rec.OrgDID+"|"+rec.MessageID] = &stored
	return nil
}

func (m *MemoryDB) CountDEFingerprints(_ context.Context, orgDID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, rec := range m.complyDEFP {
		if rec.OrgDID == orgDID {
			n++
		}
	}
	return n, nil
}

func (m *MemoryDB) CountOrgScopedMessages(ctx context.Context, orgDID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	convSet := make(map[string]struct{})
	for _, b := range m.complyBindings {
		if b.OrgDID == orgDID {
			convSet[b.ConversationID] = struct{}{}
		}
	}
	if len(convSet) == 0 {
		return 0, nil
	}
	seen := make(map[string]struct{})
	for _, msg := range m.messages {
		if _, ok := convSet[msg.ConversationID]; !ok {
			continue
		}
		seen[msg.MessageID] = struct{}{}
	}
	return len(seen), nil
}
