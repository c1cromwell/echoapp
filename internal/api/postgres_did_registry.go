package api

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDIDRegistry is a WO-273 persistent multi-device registry backed by
// the identity_device table (migration 008).
type PostgresDIDRegistry struct {
	pool *pgxpool.Pool
}

// NewPostgresDIDRegistry constructs a registry that uses the given pool.
func NewPostgresDIDRegistry(pool *pgxpool.Pool) *PostgresDIDRegistry {
	return &PostgresDIDRegistry{pool: pool}
}

func (p *PostgresDIDRegistry) Register(ctx context.Context, did, publicKeyHex string) (*DIDBinding, bool, error) {
	var devLabel string
	var regAt time.Time
	err := p.pool.QueryRow(ctx,
		`SELECT device_label, created_at FROM identity_device WHERE did = $1 AND public_key_hex = $2`,
		did, publicKeyHex,
	).Scan(&devLabel, &regAt)
	if err == nil {
		return &DIDBinding{
			DID: did, PublicKeyHex: publicKeyHex, DeviceLabel: devLabel, RegisteredAt: regAt,
		}, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	var n int
	if err := p.pool.QueryRow(ctx, `SELECT count(*)::int FROM identity_device WHERE did = $1`, did).Scan(&n); err != nil {
		return nil, false, err
	}
	if n > 0 {
		return nil, false, ErrDIDConflict
	}

	err = p.pool.QueryRow(ctx,
		`INSERT INTO identity_device (did, public_key_hex, device_label)
		 VALUES ($1, $2, 'primary') RETURNING created_at`,
		did, publicKeyHex,
	).Scan(&regAt)
	if err != nil {
		return nil, false, err
	}
	return &DIDBinding{
		DID: did, PublicKeyHex: publicKeyHex, DeviceLabel: "primary", RegisteredAt: regAt,
	}, true, nil
}

func (p *PostgresDIDRegistry) Lookup(ctx context.Context, did string) (*DIDBinding, error) {
	var publicKeyHex, devLabel string
	var regAt time.Time
	err := p.pool.QueryRow(ctx,
		`SELECT public_key_hex, device_label, created_at
		 FROM identity_device WHERE did = $1 ORDER BY created_at ASC LIMIT 1`,
		did,
	).Scan(&publicKeyHex, &devLabel, &regAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBindingNotFound
	}
	if err != nil {
		return nil, err
	}
	return cloneBinding(&DIDBinding{
		DID: did, PublicKeyHex: publicKeyHex, DeviceLabel: devLabel, RegisteredAt: regAt,
	}), nil
}

func (p *PostgresDIDRegistry) ListDevices(ctx context.Context, did string) ([]*DIDBinding, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT public_key_hex, device_label, created_at
		 FROM identity_device WHERE did = $1 ORDER BY created_at ASC`,
		did,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*DIDBinding
	for rows.Next() {
		var publicKeyHex, devLabel string
		var regAt time.Time
		if err := rows.Scan(&publicKeyHex, &devLabel, &regAt); err != nil {
			return nil, err
		}
		out = append(out, cloneBinding(&DIDBinding{
			DID: did, PublicKeyHex: publicKeyHex, DeviceLabel: devLabel, RegisteredAt: regAt,
		}))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrBindingNotFound
	}
	return out, nil
}

func (p *PostgresDIDRegistry) RegisterAdditionalDevice(ctx context.Context, subjectDID, newPublicKeyHex, deviceLabel string) (*DIDBinding, error) {
	var n int
	if err := p.pool.QueryRow(ctx, `SELECT count(*)::int FROM identity_device WHERE did = $1`, subjectDID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrBindingNotFound
	}

	var regAt time.Time
	err := p.pool.QueryRow(ctx,
		`INSERT INTO identity_device (did, public_key_hex, device_label)
		 VALUES ($1, $2, $3) RETURNING created_at`,
		subjectDID, newPublicKeyHex, deviceLabel,
	).Scan(&regAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateDeviceKey
		}
		return nil, err
	}
	return cloneBinding(&DIDBinding{
		DID: subjectDID, PublicKeyHex: newPublicKeyHex, DeviceLabel: deviceLabel, RegisteredAt: regAt,
	}), nil
}
