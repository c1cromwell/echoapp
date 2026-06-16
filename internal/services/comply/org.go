package comply

import (
	"context"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// GetOrgProfile returns tier/seat metadata for WO-310 (defaults when unset).
func (s *Service) GetOrgProfile(ctx context.Context, orgDID string) (*database.ComplyOrgProfile, error) {
	profile, err := s.store.GetOrgProfile(ctx, orgDID)
	if err != nil {
		return &database.ComplyOrgProfile{
			OrgDID:    orgDID,
			Tier:      "starter",
			Seats:     10,
			UpdatedAt: time.Now().UTC(),
		}, nil
	}
	return profile, nil
}

// ListLitigationHolds returns org-scoped matters for the portal (WO-311 / WO-312).
func (s *Service) ListLitigationHolds(ctx context.Context, orgDID string, activeOnly bool) ([]*database.LitigationMatter, error) {
	return s.store.ListLitigationMatters(ctx, orgDID, activeOnly)
}
