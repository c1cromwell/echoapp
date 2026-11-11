package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// With CURRENCY_L1_URL unset, main passes a nil *MetagraphClient as the
// CurrencySubmitter interface. That typed nil must be treated as "no Currency
// L1": the reconciler used to dereference it on its first tick and crash the
// server five minutes after start.
func TestLedgerQuerier_TypedNilSubmitterIsTreatedAsAbsent(t *testing.T) {
	var client *metagraph.MetagraphClient // nil pointer, non-nil once boxed in the interface
	q := NewLedgerQuerier(nil, client)
	if q.submitter != nil {
		t.Fatal("a nil *MetagraphClient must be normalised to a nil submitter")
	}

	done := make(chan struct{})
	go func() {
		NewReconciler(q).Run(context.Background()) // must return at once: nothing to reconcile against
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reconciler should skip when there is no Currency L1 source")
	}
}
