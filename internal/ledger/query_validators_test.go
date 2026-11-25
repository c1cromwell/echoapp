package ledger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryValidators_RequiresL0URL(t *testing.T) {
	c := NewMetagraphClient(MetagraphConfig{CurrencyL1URL: "http://currency.invalid"})
	_, err := c.QueryValidators(context.Background())
	if err == nil || !strings.Contains(err.Error(), "GLOBAL_L0_URL") {
		t.Fatalf("want a configuration error naming GLOBAL_L0_URL, got %v", err)
	}
}

func TestQueryValidators_ReadsNodeParams(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`[{"peerId":"abc","node":"DAG1","totalAmountDelegated":500,"totalAddressesAssigned":2,"delegatedStakeRewardParameters":{"rewardFraction":5000000}}]`))
	}))
	defer srv.Close()

	c := NewMetagraphClient(MetagraphConfig{L0URL: srv.URL + "/"})
	got, err := c.QueryValidators(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/node-params" {
		t.Fatalf("path = %q, want /node-params", gotPath)
	}
	if len(got) != 1 || got[0].ID != "abc" || got[0].Address != "DAG1" || got[0].TotalDelegated != 500 || got[0].DelegatorCount != 2 || got[0].CommissionPercent != 5 {
		t.Fatalf("unexpected mapping: %+v", got)
	}
}

func TestQueryValidators_EmptyListIsValid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	got, err := NewMetagraphClient(MetagraphConfig{L0URL: srv.URL}).QueryValidators(context.Background())
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}
