package credentials

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgStatusList persists StatusList2021 slots and publish outbox (WO-274).
type pgStatusList struct {
	pool      *pgxpool.Pool
	issuerDID string
}

func newPGStatusList(pool *pgxpool.Pool, issuerDID string) *pgStatusList {
	if pool == nil || issuerDID == "" {
		return nil
	}
	return &pgStatusList{pool: pool, issuerDID: issuerDID}
}

func issuerAdvisoryKeys(issuer string) (int32, int32) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(issuer))
	sum := h.Sum32()
	return int32(sum & 0x7fff), int32((sum >> 15) & 0x7fff)
}

// allocateIndex returns an existing index or allocates the next slot for this issuer.
func (s *pgStatusList) allocateIndex(ctx context.Context, credentialID string) (int, error) {
	if s == nil {
		return 0, errors.New("nil pg status list")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var idx int
	err = tx.QueryRow(ctx,
		`SELECT status_list_index FROM credential_vc_status WHERE credential_id = $1`,
		credentialID,
	).Scan(&idx)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return idx, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	k1, k2 := issuerAdvisoryKeys(s.issuerDID)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, $2)`, k1, k2); err != nil {
		return 0, err
	}

	var mx int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(status_list_index), -1)::int FROM credential_vc_status WHERE issuer_did = $1`,
		s.issuerDID,
	).Scan(&mx)
	if err != nil {
		return 0, err
	}
	next := mx + 1
	if next >= statusListBitCount {
		return 0, fmt.Errorf("status list slots exhausted for issuer")
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO credential_vc_status (credential_id, issuer_did, status_list_index) VALUES ($1, $2, $3)`,
		credentialID, s.issuerDID, next,
	); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO status_list_l1_outbox (issuer_did, pending_publish) VALUES ($1, false)
		 ON CONFLICT (issuer_did) DO NOTHING`,
		s.issuerDID,
	); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return next, nil
}

func (s *pgStatusList) markRevoked(ctx context.Context, credentialID string) (bool, error) {
	if s == nil {
		return false, nil
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE credential_vc_status SET revoked_at = COALESCE(revoked_at, now()) WHERE credential_id = $1 AND issuer_did = $2`,
		credentialID, s.issuerDID,
	)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO status_list_l1_outbox (issuer_did, pending_publish) VALUES ($1, true)
		 ON CONFLICT (issuer_did) DO UPDATE SET pending_publish = true, updated_at = now()`,
		s.issuerDID,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *pgStatusList) publishPending(ctx context.Context) (bool, error) {
	if s == nil {
		return false, nil
	}
	var pending bool
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(pending_publish, false) FROM status_list_l1_outbox WHERE issuer_did = $1`,
		s.issuerDID,
	).Scan(&pending)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return pending, nil
}

var errStatusListNothingToPublish = errors.New("credentials: status list nothing to publish")

// buildBitVector loads the current outbox sequence, builds revoked bit positions, commits, and
// returns a wall-clock watermark after commit (revokes with revoked_at > watermark need another publish).
func (s *pgStatusList) buildBitVector(ctx context.Context) ([]byte, int64, time.Time, error) {
	if s == nil {
		return nil, 0, time.Time{}, errors.New("nil store")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	defer tx.Rollback(ctx)

	k1, k2 := issuerAdvisoryKeys(s.issuerDID)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1, $2)`, k1, k2); err != nil {
		return nil, 0, time.Time{}, err
	}

	var seq int64
	var pending bool
	err = tx.QueryRow(ctx,
		`SELECT last_published_sequence, pending_publish FROM status_list_l1_outbox WHERE issuer_did = $1 FOR UPDATE`,
		s.issuerDID,
	).Scan(&seq, &pending)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return nil, 0, time.Time{}, err
		}
		return nil, 0, time.Time{}, errStatusListNothingToPublish
	}
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	if !pending {
		if err := tx.Commit(ctx); err != nil {
			return nil, 0, time.Time{}, err
		}
		return nil, 0, time.Time{}, errStatusListNothingToPublish
	}

	rows, err := tx.Query(ctx,
		`SELECT status_list_index FROM credential_vc_status WHERE issuer_did = $1 AND revoked_at IS NOT NULL`,
		s.issuerDID,
	)
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	defer rows.Close()

	vec := make([]byte, statusListBitCount/8)
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err != nil {
			return nil, 0, time.Time{}, err
		}
		if idx >= 0 && idx < statusListBitCount {
			vec[idx/8] |= 1 << (idx % 8)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, time.Time{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, time.Time{}, err
	}
	wm := time.Now().UTC()
	return vec, seq, wm, nil
}

func (s *pgStatusList) afterPublishOK(ctx context.Context, postedSeq int64, watermark time.Time) error {
	if s == nil {
		return nil
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE status_list_l1_outbox AS o SET
		   last_published_sequence = o.last_published_sequence + 1,
		   last_publish_at = now(),
		   updated_at = now(),
		   pending_publish = EXISTS (
		     SELECT 1 FROM credential_vc_status c
		     WHERE c.issuer_did = o.issuer_did
		       AND c.revoked_at IS NOT NULL
		       AND c.revoked_at > $3
		   )
		 WHERE o.issuer_did = $1 AND o.last_published_sequence = $2`,
		s.issuerDID, postedSeq, watermark,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("status list outbox sequence changed unexpectedly for issuer %s", s.issuerDID)
	}
	return nil
}

// QueryCredentialVCStatus reads durable StatusList2021 slot state from Postgres (WO-274).
func QueryCredentialVCStatus(ctx context.Context, pool *pgxpool.Pool, credentialID string) (statusListIndex int, revoked bool, found bool, err error) {
	if pool == nil || credentialID == "" {
		return 0, false, false, nil
	}
	var idx int
	var revAt *time.Time
	err = pool.QueryRow(ctx,
		`SELECT status_list_index, revoked_at FROM credential_vc_status WHERE credential_id = $1`,
		credentialID,
	).Scan(&idx, &revAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, err
	}
	return idx, revAt != nil, true, nil
}
