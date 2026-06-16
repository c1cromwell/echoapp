package database

import (
	"context"
	"fmt"
	"time"
)

// --- MemoryDB conversation index (WO-251) ---

func (m *MemoryDB) ListConversationIDsForParticipant(_ context.Context, did string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[string]struct{})
	for _, msg := range m.messages {
		if msg.SenderDID == did || msg.RecipientDID == did {
			if msg.ConversationID != "" {
				seen[msg.ConversationID] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

func (m *MemoryDB) ListMessageManifest(_ context.Context, conversationIDs []string, from, to *time.Time) ([]*ExportManifestEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	allowed := make(map[string]struct{}, len(conversationIDs))
	for _, id := range conversationIDs {
		allowed[id] = struct{}{}
	}
	var out []*ExportManifestEntry
	for _, msg := range m.messages {
		if msg.ConversationID == "" {
			continue
		}
		if _, ok := allowed[msg.ConversationID]; !ok {
			continue
		}
		ts := msg.CreatedAt
		if from != nil && ts.Before(*from) {
			continue
		}
		if to != nil && ts.After(*to) {
			continue
		}
		out = append(out, &ExportManifestEntry{
			MessageID:      msg.MessageID,
			ConversationID: msg.ConversationID,
			SenderDID:      msg.SenderDID,
			RecipientDID:   msg.RecipientDID,
			Timestamp:      ts,
		})
	}
	return out, nil
}

// --- MemoryDB ComplyExtendedStore (WO-251) ---

func (m *MemoryDB) CreateLitigationMatter(_ context.Context, matter *LitigationMatter) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyMatters == nil {
		m.complyMatters = make(map[string]*LitigationMatter)
	}
	stored := *matter
	m.complyMatters[matter.MatterID] = &stored
	return nil
}

func (m *MemoryDB) GetLitigationMatter(_ context.Context, matterID string) (*LitigationMatter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mat, ok := m.complyMatters[matterID]
	if !ok {
		return nil, ErrComplyMatterNotFound
	}
	copy := *mat
	return &copy, nil
}

func (m *MemoryDB) UpdateLitigationMatter(_ context.Context, matter *LitigationMatter) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.complyMatters[matter.MatterID]; !ok {
		return ErrComplyMatterNotFound
	}
	stored := *matter
	m.complyMatters[matter.MatterID] = &stored
	return nil
}

func (m *MemoryDB) AddLitigationCustodian(_ context.Context, b *LitigationCustodianBinding) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyCustodians == nil {
		m.complyCustodians = make(map[string][]*LitigationCustodianBinding)
	}
	key := b.MatterID + "|" + b.CustodianDID + "|" + b.ConversationID
	if m.complyCustodianKeys == nil {
		m.complyCustodianKeys = make(map[string]struct{})
	}
	if _, dup := m.complyCustodianKeys[key]; dup {
		return nil
	}
	m.complyCustodianKeys[key] = struct{}{}
	stored := *b
	m.complyCustodians[b.MatterID] = append(m.complyCustodians[b.MatterID], &stored)
	return nil
}

func (m *MemoryDB) ListLitigationCustodians(_ context.Context, matterID string) ([]*LitigationCustodianBinding, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.complyCustodians[matterID]
	out := make([]*LitigationCustodianBinding, len(list))
	for i, b := range list {
		copy := *b
		out[i] = &copy
	}
	return out, nil
}

func (m *MemoryDB) CountActiveLitigationMatters(_ context.Context, orgDID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, mat := range m.complyMatters {
		if mat.OrgDID == orgDID && mat.Status == MatterActive {
			n++
		}
	}
	return n, nil
}

func (m *MemoryDB) ListLitigationMatters(_ context.Context, orgDID string, activeOnly bool) ([]*LitigationMatter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*LitigationMatter
	for _, mat := range m.complyMatters {
		if mat.OrgDID != orgDID {
			continue
		}
		if activeOnly && mat.Status != MatterActive {
			continue
		}
		copy := *mat
		out = append(out, &copy)
	}
	return out, nil
}

func (m *MemoryDB) CreateEDiscoveryExport(_ context.Context, e *EDiscoveryExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyExports == nil {
		m.complyExports = make(map[string]*EDiscoveryExport)
	}
	stored := *e
	m.complyExports[e.ExportID] = &stored
	return nil
}

func (m *MemoryDB) GetEDiscoveryExport(_ context.Context, exportID string) (*EDiscoveryExport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.complyExports[exportID]
	if !ok {
		return nil, ErrComplyExportNotFound
	}
	copy := *e
	return &copy, nil
}

func (m *MemoryDB) UpdateEDiscoveryExport(_ context.Context, e *EDiscoveryExport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.complyExports[e.ExportID]; !ok {
		return ErrComplyExportNotFound
	}
	stored := *e
	m.complyExports[e.ExportID] = &stored
	return nil
}

func (m *MemoryDB) CountPendingExports(_ context.Context, orgDID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, e := range m.complyExports {
		if e.OrgDID == orgDID && (e.Status == ExportPending || e.Status == ExportProcessing) {
			n++
		}
	}
	return n, nil
}

func (m *MemoryDB) ListEDiscoveryExports(_ context.Context, orgDID string, limit int) ([]*EDiscoveryExport, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	var out []*EDiscoveryExport
	for _, e := range m.complyExports {
		if e.OrgDID == orgDID {
			copy := *e
			out = append(out, &copy)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *MemoryDB) AppendAuditEvent(_ context.Context, e *AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.complyAudit == nil {
		m.complyAudit = make([]*AuditEvent, 0)
	}
	if e.ID == "" {
		return fmt.Errorf("audit event id required")
	}
	stored := *e
	m.complyAudit = append(m.complyAudit, &stored)
	return nil
}

func (m *MemoryDB) ListAuditEvents(_ context.Context, orgDID string, from, to time.Time) ([]*AuditEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*AuditEvent
	for _, e := range m.complyAudit {
		if e.OrgDID != orgDID {
			continue
		}
		if e.OccurredAt.Before(from) || e.OccurredAt.After(to) {
			continue
		}
		copy := *e
		out = append(out, &copy)
	}
	return out, nil
}
