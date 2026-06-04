package contacts

import "context"

// MutualContactSummary is a privacy-safe mutual contact entry (DID + optional username).
type MutualContactSummary struct {
	DID      string `json:"did"`
	Username string `json:"username,omitempty"`
}

// MutualContacts returns contacts present in both users' contact lists (max limit).
func (s *Service) MutualContacts(ctx context.Context, callerDID, peerDID string, limit int) ([]MutualContactSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	a, err := s.db.GetContacts(ctx, callerDID)
	if err != nil {
		return nil, err
	}
	b, err := s.db.GetContacts(ctx, peerDID)
	if err != nil {
		return nil, err
	}
	inB := make(map[string]struct{}, len(b))
	for _, c := range b {
		if c.ContactDID != "" {
			inB[c.ContactDID] = struct{}{}
		}
	}
	var out []MutualContactSummary
	for _, c := range a {
		if c.ContactDID == "" || c.ContactDID == callerDID || c.ContactDID == peerDID {
			continue
		}
		if _, ok := inB[c.ContactDID]; !ok {
			continue
		}
		entry := MutualContactSummary{DID: c.ContactDID}
		if user, err := s.db.GetUserByDID(ctx, c.ContactDID); err == nil && user.Username != "" {
			entry.Username = user.Username
		}
		out = append(out, entry)
		if len(out) >= limit {
			break
		}
	}
	if out == nil {
		out = []MutualContactSummary{}
	}
	return out, nil
}
