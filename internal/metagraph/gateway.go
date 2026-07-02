package metagraph

import (
	"context"
	"errors"
)

var ErrGatewayNotConfigured = errors.New("metagraph: gateway layer not configured")

// Gateway is the unified metagraph write facade for the Go API (WO-8 / WO-27).
// Production wiring uses one MetagraphClient per layer URL from main.go; this type
// documents the single integration surface for Data, Currency, and Identity L1.
type Gateway struct {
	Data     *MetagraphClient
	Currency *MetagraphClient
	Identity *MetagraphClient
}

// NewGateway groups layer clients. Pass the same client when URLs share a host.
func NewGateway(data, currency, identity *MetagraphClient) *Gateway {
	return &Gateway{Data: data, Currency: currency, Identity: identity}
}

// SubmitMessageMerkleRoot anchors a message-integrity batch on Data L1 (WO-15).
func (g *Gateway) SubmitMessageMerkleRoot(ctx context.Context, root string, leafCount int) (string, error) {
	if g == nil || g.Data == nil {
		return "", ErrGatewayNotConfigured
	}
	return g.Data.SubmitDataL1(ctx, DataL1MerkleRootUpdate{Root: root, LeafCount: leafCount})
}

// SubmitCurrency posts a Currency L1 transaction (rewards, wallet, staking).
func (g *Gateway) SubmitCurrency(ctx context.Context, tx CurrencyL1Transaction) (string, error) {
	if g == nil || g.Currency == nil {
		return "", ErrGatewayNotConfigured
	}
	return g.Currency.SubmitCurrencyL1(ctx, tx)
}

// SubmitIdentityData posts an Identity L1 data-application update (trust tier, VC).
func (g *Gateway) SubmitIdentityData(ctx context.Context, update interface{}) (string, error) {
	if g == nil || g.Identity == nil {
		return "", ErrGatewayNotConfigured
	}
	return g.Identity.SubmitDataL1(ctx, update)
}
