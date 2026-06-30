package wallet

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"
)

// Reconciler periodically re-indexes authoritative metagraph state into the PG
// cache, turning the cache into a correctable mirror rather than a silent fork.
//
// Today it reconciles the validator set — the only state the metagraph client
// can currently query. Balance/lock/delegation reconciliation is gated on the
// Currency L0 calculated-state query API (post-launch, see
// docs/ECHO_WALLET_STAKING_LAUNCH.md "Out of scope"); the hook is marked in
// reconcileOnce. Until then PG remains authoritative for those.
type Reconciler struct {
	querier  *LedgerQuerier
	interval time.Duration
}

// NewReconciler builds a reconciler over the given ledger querier.
func NewReconciler(querier *LedgerQuerier) *Reconciler {
	return &Reconciler{querier: querier, interval: reconcileInterval()}
}

func reconcileInterval() time.Duration {
	if raw := os.Getenv("ECHO_WALLET_RECONCILE_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 5 * time.Minute
}

// Run blocks until ctx is cancelled, reconciling on each tick. It no-ops when no
// Currency L1 source is wired (nothing authoritative to reconcile against).
func (r *Reconciler) Run(ctx context.Context) {
	if r.querier == nil || r.querier.submitter == nil {
		log.Println("wallet reconciler: no Currency L1 source; skipping")
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	log.Printf("wallet reconciler: started (interval=%s)", r.interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("wallet reconciler: stopped")
			return
		case <-ticker.C:
			r.reconcileOnce(ctx)
		}
	}
}

func (r *Reconciler) reconcileOnce(ctx context.Context) {
	// Validators: pull the authoritative snapshot and refresh the cache,
	// logging divergence so a drifting mirror is visible rather than silent.
	snaps, err := r.querier.submitter.QueryValidators(ctx)
	if err != nil {
		log.Printf("wallet reconciler: validator query failed: %v", err)
		return
	}
	if cached, err := r.querier.store.ListValidators(ctx); err == nil && len(cached) != len(snaps) {
		log.Printf("wallet reconciler: validator divergence chain=%d cache=%d; refreshing", len(snaps), len(cached))
	}
	if err := r.querier.store.UpsertValidators(ctx, validatorsFromSnapshots(snaps)); err != nil {
		log.Printf("wallet reconciler: validator cache refresh failed: %v", err)
	}

	// TODO(currency-l0): reconcile balances / locks / delegations from the L0
	// calculated state once that query API lands. Until then those remain
	// PG-authoritative and cannot diverge from a source we cannot yet read.
}
