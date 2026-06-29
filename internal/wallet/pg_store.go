package wallet

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists wallet balances, staking positions, and DAG address links.
type Store interface {
	GetBalance(ctx context.Context, did string) (*BalanceInfo, error)
	EnsureGenesisCredit(ctx context.Context, did string) error
	ListLocks(ctx context.Context, did string) ([]TokenLockPos, error)
	ListDelegations(ctx context.Context, did string) ([]DelegationPos, error)
	InsertLock(ctx context.Context, did string, pos TokenLockPos) error
	InsertDelegation(ctx context.Context, did string, d DelegationPos) error
	ApplyStake(ctx context.Context, did string, amount int64) error
	ApplyUnstake(ctx context.Context, did string, amount int64) error
	CreditRewards(ctx context.Context, did string, amount int64) error
	GetDAGAddress(ctx context.Context, did string) (string, error)
	LinkDAGAddress(ctx context.Context, did, address string) error
	UpsertValidators(ctx context.Context, validators []ValidatorInfo) error
	ListValidators(ctx context.Context) ([]ValidatorInfo, error)
}

// PGStore implements Store on PostgreSQL (migrations 003 + 024).
type PGStore struct {
	pool *pgxpool.Pool
}

// NewPGStore returns a PostgreSQL-backed wallet store.
func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func genesisCreditDatum() int64 {
	raw := os.Getenv("ECHO_WALLET_GENESIS_ECHO")
	if raw == "" {
		raw = "1000"
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || f <= 0 {
		return 1000 * DatumPerECHO
	}
	return int64(f * float64(DatumPerECHO))
}

func genesisAutoEnabled() bool {
	return os.Getenv("ECHO_WALLET_GENESIS_AUTO") == "1" ||
		os.Getenv("ENVIRONMENT") != "production"
}

// EnsureGenesisCredit seeds dev/TestFlight balance when cache row is missing.
func (s *PGStore) EnsureGenesisCredit(ctx context.Context, did string) error {
	if !genesisAutoEnabled() {
		return nil
	}
	credit := genesisCreditDatum()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO wallet_balance_cache (did, total_balance, available, staked, pending_rewards)
		VALUES ($1, $2, $2, 0, 0)
		ON CONFLICT (did) DO NOTHING
	`, did, credit)
	return err
}

func (s *PGStore) GetBalance(ctx context.Context, did string) (*BalanceInfo, error) {
	if err := s.EnsureGenesisCredit(ctx, did); err != nil {
		return nil, err
	}
	var total, available int64
	err := s.pool.QueryRow(ctx, `
		SELECT total_balance, available FROM wallet_balance_cache WHERE did = $1
	`, did).Scan(&total, &available)
	if errors.Is(err, pgx.ErrNoRows) {
		return &BalanceInfo{Total: 0, Available: 0}, nil
	}
	if err != nil {
		return nil, err
	}
	return &BalanceInfo{Total: total, Available: available}, nil
}

func (s *PGStore) ListLocks(ctx context.Context, did string) ([]TokenLockPos, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, amount, tier, locked_until, COALESCE(vesting_type, ''), COALESCE(delegated_to, '')
		FROM staking_positions WHERE did = $1 ORDER BY locked_until ASC
	`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TokenLockPos
	for rows.Next() {
		var p TokenLockPos
		var vesting, delegated string
		if err := rows.Scan(&p.ID, &p.Amount, &p.Tier, &p.LockedUntil, &vesting, &delegated); err != nil {
			return nil, err
		}
		p.VestingType = vesting
		p.DelegatedTo = delegated
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PGStore) ListDelegations(ctx context.Context, did string) ([]DelegationPos, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, stake_id, validator_id, amount, created_at
		FROM staking_delegations WHERE did = $1 ORDER BY created_at DESC
	`, did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DelegationPos
	for rows.Next() {
		var d DelegationPos
		if err := rows.Scan(&d.ID, &d.StakeID, &d.ValidatorID, &d.Amount, &d.Since); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *PGStore) InsertLock(ctx context.Context, did string, pos TokenLockPos) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO staking_positions (id, did, amount, tier, locked_until, vesting_type, delegated_to, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NOW())
		ON CONFLICT (id) DO UPDATE SET
			amount = EXCLUDED.amount,
			tier = EXCLUDED.tier,
			locked_until = EXCLUDED.locked_until,
			updated_at = NOW()
	`, pos.ID, did, pos.Amount, pos.Tier, pos.LockedUntil, pos.VestingType, pos.DelegatedTo)
	return err
}

func (s *PGStore) InsertDelegation(ctx context.Context, did string, d DelegationPos) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO staking_delegations (id, did, stake_id, validator_id, amount, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			stake_id = EXCLUDED.stake_id,
			validator_id = EXCLUDED.validator_id,
			amount = EXCLUDED.amount
	`, d.ID, did, d.StakeID, d.ValidatorID, d.Amount)
	return err
}

func (s *PGStore) ApplyStake(ctx context.Context, did string, amount int64) error {
	res, err := s.pool.Exec(ctx, `
		UPDATE wallet_balance_cache
		SET available = available - $2,
		    staked = staked + $2,
		    updated_at = NOW()
		WHERE did = $1 AND available >= $2
	`, did, amount)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficientBalance
	}
	return nil
}

func (s *PGStore) ApplyUnstake(ctx context.Context, did string, amount int64) error {
	res, err := s.pool.Exec(ctx, `
		UPDATE wallet_balance_cache
		SET available = available + $2,
		    staked = GREATEST(0, staked - $2),
		    updated_at = NOW()
		WHERE did = $1
	`, did, amount)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficientBalance
	}
	return nil
}

func (s *PGStore) CreditRewards(ctx context.Context, did string, amount int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO wallet_balance_cache (did, total_balance, available, staked, pending_rewards)
		VALUES ($1, $2, $2, 0, 0)
		ON CONFLICT (did) DO UPDATE SET
			total_balance = wallet_balance_cache.total_balance + $2,
			available = wallet_balance_cache.available + $2,
			updated_at = NOW()
	`, did, amount)
	return err
}

func (s *PGStore) GetDAGAddress(ctx context.Context, did string) (string, error) {
	var addr string
	err := s.pool.QueryRow(ctx, `SELECT dag_address FROM wallet_accounts WHERE did = $1`, did).Scan(&addr)
	return addr, err
}

func (s *PGStore) LinkDAGAddress(ctx context.Context, did, address string) error {
	if did == "" || address == "" {
		return fmt.Errorf("did and dag_address required")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO wallet_accounts (did, dag_address) VALUES ($1, $2)
		ON CONFLICT (did) DO UPDATE SET dag_address = EXCLUDED.dag_address, linked_at = NOW()
	`, did, address)
	return err
}

func (s *PGStore) UpsertValidators(ctx context.Context, validators []ValidatorInfo) error {
	for _, v := range validators {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO validators (id, address, layer, uptime_percent, commission_pct, total_delegated, delegator_count, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (id) DO UPDATE SET
				address = EXCLUDED.address,
				uptime_percent = EXCLUDED.uptime_percent,
				commission_pct = EXCLUDED.commission_pct,
				total_delegated = EXCLUDED.total_delegated,
				delegator_count = EXCLUDED.delegator_count,
				updated_at = NOW()
		`, v.ID, v.Address, v.Layer, v.Uptime, v.Commission, v.TotalDelegated, v.DelegatorCount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) ListValidators(ctx context.Context) ([]ValidatorInfo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, address, layer, uptime_percent, commission_pct, total_delegated, delegator_count
		FROM validators ORDER BY total_delegated DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ValidatorInfo
	for rows.Next() {
		var v ValidatorInfo
		if err := rows.Scan(&v.ID, &v.Address, &v.Layer, &v.Uptime, &v.Commission, &v.TotalDelegated, &v.DelegatorCount); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// MemStore is an in-memory Store for unit tests.
type MemStore struct {
	balances    map[string]*BalanceInfo
	locks       map[string][]TokenLockPos
	delegations map[string][]DelegationPos
	addresses   map[string]string
	validators  []ValidatorInfo
}

// NewMemStore creates an empty in-memory wallet store.
func NewMemStore() *MemStore {
	return &MemStore{
		balances:    make(map[string]*BalanceInfo),
		locks:       make(map[string][]TokenLockPos),
		delegations: make(map[string][]DelegationPos),
		addresses:   make(map[string]string),
	}
}

func (m *MemStore) EnsureGenesisCredit(_ context.Context, did string) error {
	if _, ok := m.balances[did]; !ok && genesisAutoEnabled() {
		c := genesisCreditDatum()
		m.balances[did] = &BalanceInfo{Total: c, Available: c}
	}
	return nil
}

func (m *MemStore) GetBalance(ctx context.Context, did string) (*BalanceInfo, error) {
	_ = m.EnsureGenesisCredit(ctx, did)
	if b, ok := m.balances[did]; ok {
		return &BalanceInfo{Total: b.Total, Available: b.Available}, nil
	}
	return &BalanceInfo{}, nil
}

func (m *MemStore) ListLocks(_ context.Context, did string) ([]TokenLockPos, error) {
	return append([]TokenLockPos{}, m.locks[did]...), nil
}

func (m *MemStore) ListDelegations(_ context.Context, did string) ([]DelegationPos, error) {
	return append([]DelegationPos{}, m.delegations[did]...), nil
}

func (m *MemStore) InsertLock(_ context.Context, did string, pos TokenLockPos) error {
	m.locks[did] = append(m.locks[did], pos)
	return nil
}

func (m *MemStore) InsertDelegation(_ context.Context, did string, d DelegationPos) error {
	m.delegations[did] = append(m.delegations[did], d)
	return nil
}

func (m *MemStore) ApplyStake(_ context.Context, did string, amount int64) error {
	b := m.balances[did]
	if b == nil || b.Available < amount {
		return ErrInsufficientBalance
	}
	b.Available -= amount
	return nil
}

func (m *MemStore) ApplyUnstake(_ context.Context, did string, amount int64) error {
	b := m.balances[did]
	if b == nil {
		return ErrInsufficientBalance
	}
	b.Available += amount
	return nil
}

func (m *MemStore) CreditRewards(_ context.Context, did string, amount int64) error {
	if m.balances[did] == nil {
		m.balances[did] = &BalanceInfo{}
	}
	m.balances[did].Total += amount
	m.balances[did].Available += amount
	return nil
}

func (m *MemStore) GetDAGAddress(_ context.Context, did string) (string, error) {
	return m.addresses[did], nil
}

func (m *MemStore) LinkDAGAddress(_ context.Context, did, address string) error {
	m.addresses[did] = address
	return nil
}

func (m *MemStore) UpsertValidators(_ context.Context, validators []ValidatorInfo) error {
	m.validators = validators
	return nil
}

func (m *MemStore) ListValidators(context.Context) ([]ValidatorInfo, error) {
	return append([]ValidatorInfo{}, m.validators...), nil
}

// NewLockID returns a unique staking position id.
func NewLockID() string {
	return "lock-" + uuid.New().String()
}

// LockUntil computes unlock time from tier lock days.
func LockUntil(days int) time.Time {
	return time.Now().UTC().AddDate(0, 0, days)
}
