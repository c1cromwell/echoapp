package groups

import (
	"context"
	"errors"
	"sync"
)

// ErrStoreConflict is returned when a group or member already exists.
var ErrStoreConflict = errors.New("group store conflict")

// GroupStore persists group metadata and membership. Implementations must be safe
// for concurrent use from multiple HTTP handlers.
type GroupStore interface {
	SaveGroup(ctx context.Context, g *Group) error
	GetGroup(ctx context.Context, groupID string) (*Group, error)
	GroupExists(ctx context.Context, groupID string) (bool, error)

	SaveMember(ctx context.Context, m *GroupMember) error
	UpdateMember(ctx context.Context, m *GroupMember) error
	GetMember(ctx context.Context, groupID, memberID string) (*GroupMember, error)
	DeleteMember(ctx context.Context, groupID, memberID string) error
	ListMembers(ctx context.Context, groupID string) ([]*GroupMember, error)
	ListGroupIDsForMember(ctx context.Context, memberID string) ([]string, error)
	CountOwnedGroups(ctx context.Context, ownerID string) (int, error)
}

type memoryStore struct {
	mu           sync.RWMutex
	groups       map[string]*Group
	memberships  map[string][]*GroupMember
	memberLookup map[string]*GroupMember
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		groups:       make(map[string]*Group),
		memberships:  make(map[string][]*GroupMember),
		memberLookup: make(map[string]*GroupMember),
	}
}

func memberKey(groupID, memberID string) string {
	return memberID + ":" + groupID
}

func (s *memoryStore) SaveGroup(_ context.Context, g *Group) error {
	if g == nil || g.GroupID == "" {
		return errors.New("group id required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *g
	s.groups[g.GroupID] = &cp
	return nil
}

func (s *memoryStore) GetGroup(_ context.Context, groupID string) (*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.groups[groupID]
	if !ok {
		return nil, ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}

func (s *memoryStore) GroupExists(_ context.Context, groupID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.groups[groupID]
	return ok, nil
}

func (s *memoryStore) SaveMember(_ context.Context, m *GroupMember) error {
	if m == nil || m.GroupID == "" || m.MemberID == "" {
		return errors.New("group and member id required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memberKey(m.GroupID, m.MemberID)
	if _, exists := s.memberLookup[key]; exists {
		return ErrStoreConflict
	}
	cp := *m
	s.memberships[m.GroupID] = append(s.memberships[m.GroupID], &cp)
	s.memberLookup[key] = &cp
	return nil
}

func (s *memoryStore) GetMember(_ context.Context, groupID, memberID string) (*GroupMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.memberLookup[memberKey(groupID, memberID)]
	if !ok {
		return nil, ErrMemberNotFound
	}
	cp := *m
	return &cp, nil
}

func (s *memoryStore) DeleteMember(_ context.Context, groupID, memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memberKey(groupID, memberID)
	if _, ok := s.memberLookup[key]; !ok {
		return ErrMemberNotFound
	}
	members := s.memberships[groupID]
	out := make([]*GroupMember, 0, len(members))
	for _, m := range members {
		if m.MemberID != memberID {
			out = append(out, m)
		}
	}
	s.memberships[groupID] = out
	delete(s.memberLookup, key)
	return nil
}

func (s *memoryStore) ListMembers(_ context.Context, groupID string) ([]*GroupMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members, ok := s.memberships[groupID]
	if !ok {
		return nil, ErrGroupNotFound
	}
	out := make([]*GroupMember, len(members))
	for i, m := range members {
		cp := *m
		out[i] = &cp
	}
	return out, nil
}

func (s *memoryStore) ListGroupIDsForMember(_ context.Context, memberID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
	for groupID, members := range s.memberships {
		for _, m := range members {
			if m.MemberID == memberID {
				ids = append(ids, groupID)
				break
			}
		}
	}
	return ids, nil
}

func (s *memoryStore) CountOwnedGroups(_ context.Context, ownerID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, g := range s.groups {
		if g.OwnerID == ownerID {
			n++
		}
	}
	return n, nil
}

func (s *memoryStore) UpdateMember(_ context.Context, m *GroupMember) error {
	return s.updateMember(m)
}

// updateMemberInMemory updates an existing member record (role changes, mute, ban).
func (s *memoryStore) updateMember(m *GroupMember) error {
	if m == nil || m.GroupID == "" || m.MemberID == "" {
		return errors.New("group and member id required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memberKey(m.GroupID, m.MemberID)
	stored, ok := s.memberLookup[key]
	if !ok {
		return ErrMemberNotFound
	}
	*stored = *m
	return nil
}
