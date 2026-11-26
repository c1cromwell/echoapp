// Package wallet holds the account recovery key.
//
// The name is historical. When an account is created the app derives a secp256k1 key from the
// recovery phrase and links that key's address to the account. Restoring an account on a new
// device is proven by signing a challenge with that key (POST /v1/auth/restore-did). The address
// and signature formats are the ones the app has always used ("DAG..." addresses); they are not
// tied to any chain, and nothing here holds or moves anything of value.
package wallet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotLinked means no account is linked to the address.
var ErrNotLinked = errors.New("recovery address is not linked to an account")

// ErrLinkedElsewhere means the address already belongs to a different account.
var ErrLinkedElsewhere = errors.New("recovery address is already linked to another account")

// Store remembers which account each recovery address belongs to.
type Store interface {
	// Link binds address to did. An address belongs to one account: linking one that another
	// account holds returns ErrLinkedElsewhere. Linking again for the same account is a no-op.
	Link(ctx context.Context, did, address string) error
	// AccountFor returns the account an address is linked to, or ErrNotLinked.
	AccountFor(ctx context.Context, address string) (string, error)
}

// ServerDerivableAddress is the placeholder address handed out at enrollment before the app has
// linked its real one: "DAG" + the first 36 hex characters of SHA-256(did). Anyone can compute
// it, no private key corresponds to it, and so it can never pass a restore.
func ServerDerivableAddress(did string) string {
	sum := sha256.Sum256([]byte(did))
	return "DAG" + hex.EncodeToString(sum[:])[:36]
}

// MemStore is the in-process store, used when no database is configured.
type MemStore struct {
	mu        sync.Mutex
	byAddress map[string]string
}

// NewMemStore creates an empty store.
func NewMemStore() *MemStore { return &MemStore{byAddress: map[string]string{}} }

// Link implements Store.
func (m *MemStore) Link(_ context.Context, did, address string) error {
	if did == "" || address == "" {
		return errors.New("did and address are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if owner, ok := m.byAddress[address]; ok && owner != did {
		return ErrLinkedElsewhere
	}
	// One address per account: a new link replaces the account's earlier one.
	for addr, owner := range m.byAddress {
		if owner == did && addr != address {
			delete(m.byAddress, addr)
		}
	}
	m.byAddress[address] = did
	return nil
}

// AccountFor implements Store.
func (m *MemStore) AccountFor(_ context.Context, address string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if did, ok := m.byAddress[address]; ok && address != "" {
		return did, nil
	}
	return "", ErrNotLinked
}

// PGStore is the durable store (table wallet_accounts, migration 024).
type PGStore struct {
	pool *pgxpool.Pool
}

// NewPGStore wraps a pgx pool.
func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

// Link implements Store.
func (s *PGStore) Link(ctx context.Context, did, address string) error {
	if did == "" || address == "" {
		return errors.New("did and address are required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	// Serialise links for one address so two accounts cannot both claim it.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, address); err != nil {
		return err
	}
	var owner string
	err = tx.QueryRow(ctx, `SELECT did FROM wallet_accounts WHERE dag_address = $1 AND did <> $2 LIMIT 1`, address, did).Scan(&owner)
	if err == nil {
		return ErrLinkedElsewhere
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_accounts (did, dag_address) VALUES ($1, $2)
		ON CONFLICT (did) DO UPDATE SET dag_address = EXCLUDED.dag_address, linked_at = NOW()`, did, address); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// AccountFor implements Store.
func (s *PGStore) AccountFor(ctx context.Context, address string) (string, error) {
	var did string
	err := s.pool.QueryRow(ctx, `SELECT did FROM wallet_accounts WHERE dag_address = $1 LIMIT 1`, address).Scan(&did)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotLinked
	}
	return did, err
}
