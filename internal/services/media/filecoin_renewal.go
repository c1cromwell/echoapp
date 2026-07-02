package media

import (
	"context"
	"log"
	"time"
)

// RenewalService renews Filecoin deals before expiry (WO-185).
type RenewalService struct {
	media    *Service
	interval time.Duration
	leadTime time.Duration
}

// NewRenewalService creates a background deal renewal worker.
func NewRenewalService(media *Service) *RenewalService {
	return &RenewalService{
		media:    media,
		interval: 6 * time.Hour,
		leadTime: 30 * 24 * time.Hour,
	}
}

// Run starts the renewal loop until ctx is cancelled.
func (r *RenewalService) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.renewExpiring(ctx)
		}
	}
}

func (r *RenewalService) renewExpiring(ctx context.Context) {
	ab, ok := r.media.StorageArchiver()
	if !ok {
		return
	}
	now := time.Now()
	for key, deal := range ab.AllDeals() {
		if deal == nil || deal.ExpiresAt.IsZero() {
			continue
		}
		if deal.ExpiresAt.Sub(now) > r.leadTime {
			continue
		}
		renewed, err := ab.RenewDeal(ctx, key, 180)
		if err != nil {
			log.Printf("filecoin: renewal failed for %s: %v", key, err)
			continue
		}
		log.Printf("filecoin: renewed deal %s cid=%s expires=%s", key, renewed.CID, renewed.ExpiresAt.Format(time.RFC3339))
	}
}
