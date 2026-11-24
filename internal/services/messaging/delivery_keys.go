package messaging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Delivery keys are how a message reaches someone without the server learning who sent it.
//
// Each account makes a random 32-byte delivery key and gives it, inside end-to-end encrypted
// messages, to the people it talks to. The server is told only SHA-256(key). To send, a client
// presents the recipient's key on a request that carries no login; the server hashes it, finds the
// recipient, and delivers. Holding the key is the permission to send; the server never needs to
// know, and is never told, who holds it.
//
// An account keeps its two most recent keys valid, so rotating does not strand a contact who has
// not yet received the new one.

// ErrBadDeliveryKeyHash means the value is not 64 lowercase hex characters.
var ErrBadDeliveryKeyHash = errors.New("delivery key hash must be 64 lowercase hex characters")

// keptPerAccount is how many delivery keys stay valid per account.
const keptPerAccount = 2

// DeliveryKeyStore maps the hash of a delivery key to the account it delivers to.
type DeliveryKeyStore interface {
	// Register makes hash the account's newest delivery key. Older keys beyond the last two stop
	// working; with revokeOthers every other key stops working at once.
	Register(ctx context.Context, did, hash string, revokeOthers bool) error
	// Resolve returns the account a delivery key hash delivers to.
	Resolve(ctx context.Context, hash string) (did string, ok bool, err error)
}

// HashDeliveryKey is the value the server stores for a delivery key.
func HashDeliveryKey(key []byte) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:])
}

func validHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, c := range hash {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// MemoryDeliveryKeyStore is the in-process store, used when no database is configured.
type MemoryDeliveryKeyStore struct {
	mu     sync.Mutex
	byHash map[string]string   // hash -> did
	byDID  map[string][]string // did -> hashes, newest last
}

// NewMemoryDeliveryKeyStore creates an empty store.
func NewMemoryDeliveryKeyStore() *MemoryDeliveryKeyStore {
	return &MemoryDeliveryKeyStore{byHash: map[string]string{}, byDID: map[string][]string{}}
}

// Register implements DeliveryKeyStore.
func (s *MemoryDeliveryKeyStore) Register(_ context.Context, did, hash string, revokeOthers bool) error {
	if did == "" {
		return errors.New("delivery key needs an account")
	}
	if !validHash(hash) {
		return ErrBadDeliveryKeyHash
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := s.byDID[did]
	if owner, taken := s.byHash[hash]; taken {
		if owner != did {
			// A key someone else registered. Refuse without saying whose it is.
			return errors.New("delivery key is not usable")
		}
	} else {
		s.byHash[hash] = did
		keys = append(keys, hash)
	}
	keep := keptPerAccount
	if revokeOthers {
		for _, k := range keys {
			if k != hash {
				delete(s.byHash, k)
			}
		}
		keys = []string{hash}
	}
	for len(keys) > keep {
		delete(s.byHash, keys[0])
		keys = keys[1:]
	}
	s.byDID[did] = keys
	return nil
}

// Resolve implements DeliveryKeyStore.
func (s *MemoryDeliveryKeyStore) Resolve(_ context.Context, hash string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	did, ok := s.byHash[hash]
	return did, ok, nil
}

// PostgresDeliveryKeyStore is the durable store (migration 038_delivery_keys).
type PostgresDeliveryKeyStore struct {
	pool *pgxpool.Pool
}

// NewPostgresDeliveryKeyStore wraps a pgx pool.
func NewPostgresDeliveryKeyStore(pool *pgxpool.Pool) *PostgresDeliveryKeyStore {
	return &PostgresDeliveryKeyStore{pool: pool}
}

// Register implements DeliveryKeyStore.
func (s *PostgresDeliveryKeyStore) Register(ctx context.Context, did, hash string, revokeOthers bool) error {
	if did == "" {
		return errors.New("delivery key needs an account")
	}
	if !validHash(hash) {
		return ErrBadDeliveryKeyHash
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	tag, err := tx.Exec(ctx, `INSERT INTO delivery_keys (key_hash, did) VALUES ($1, $2)
		ON CONFLICT (key_hash) DO UPDATE SET did = delivery_keys.did WHERE delivery_keys.did = EXCLUDED.did`, hash, did)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("delivery key is not usable")
	}
	if revokeOthers {
		if _, err := tx.Exec(ctx, `DELETE FROM delivery_keys WHERE did = $1 AND key_hash <> $2`, did, hash); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM delivery_keys WHERE did = $1 AND key_hash NOT IN (
		SELECT key_hash FROM delivery_keys WHERE did = $1 ORDER BY created_at DESC, key_hash LIMIT $2)`, did, keptPerAccount); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Resolve implements DeliveryKeyStore.
func (s *PostgresDeliveryKeyStore) Resolve(ctx context.Context, hash string) (string, bool, error) {
	var did string
	err := s.pool.QueryRow(ctx, `SELECT did FROM delivery_keys WHERE key_hash = $1`, hash).Scan(&did)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return did, true, nil
}
