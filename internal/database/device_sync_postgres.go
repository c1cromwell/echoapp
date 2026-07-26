package database

import (
	"context"
	"fmt"
	"time"
)

// --- PostgresDB implementation of DeviceSyncStore (WO-CA3) ---

func (p *PostgresDB) AppendSyncEntry(ctx context.Context, entry *SyncEntry) (int64, error) {
	if entry == nil || entry.ControllerDID == "" || entry.TargetDeviceID == "" {
		return 0, fmt.Errorf("controller did and target device id required")
	}
	revoked, err := p.isStreamRevoked(ctx, entry.ControllerDID, entry.TargetDeviceID)
	if err != nil {
		return 0, err
	}
	if revoked {
		return 0, ErrDeviceRevoked
	}
	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	var seq int64
	err = p.pool.QueryRow(ctx,
		`INSERT INTO device_sync_log (controller_did, target_device_id, seq, entry_type, ciphertext, created_at)
		 VALUES ($1, $2,
		         (SELECT COALESCE(MAX(seq), 0) + 1 FROM device_sync_log WHERE controller_did = $1 AND target_device_id = $2),
		         $3, $4, $5)
		 RETURNING seq`,
		entry.ControllerDID, entry.TargetDeviceID, entry.EntryType, entry.Ciphertext, createdAt).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("append sync entry: %w", err)
	}
	// Bound pending stream (match MemSyncStore); oldest unacked entries fall off.
	_, _ = p.pool.Exec(ctx,
		`DELETE FROM device_sync_log
		  WHERE controller_did = $1 AND target_device_id = $2
		    AND seq <= (
		      SELECT COALESCE(MAX(seq), 0) - $3 FROM device_sync_log
		       WHERE controller_did = $1 AND target_device_id = $2
		    )`,
		entry.ControllerDID, entry.TargetDeviceID, MaxSyncEntriesPerDevice)
	return seq, nil
}

func (p *PostgresDB) AckSyncEntries(ctx context.Context, controllerDID, targetDeviceID string, throughSeq int64) error {
	if throughSeq <= 0 {
		return nil
	}
	if _, err := p.pool.Exec(ctx,
		`DELETE FROM device_sync_log
		  WHERE controller_did = $1 AND target_device_id = $2 AND seq <= $3`,
		controllerDID, targetDeviceID, throughSeq); err != nil {
		return fmt.Errorf("ack sync entries: %w", err)
	}
	return nil
}

func (p *PostgresDB) PullSyncEntries(ctx context.Context, controllerDID, targetDeviceID string, afterSeq int64, limit int) ([]*SyncEntry, error) {
	revoked, err := p.isStreamRevoked(ctx, controllerDID, targetDeviceID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, ErrDeviceRevoked
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := p.pool.Query(ctx,
		`SELECT controller_did, target_device_id, seq, entry_type, ciphertext, created_at
		   FROM device_sync_log
		  WHERE controller_did = $1 AND target_device_id = $2 AND seq > $3
		  ORDER BY seq ASC LIMIT $4`,
		controllerDID, targetDeviceID, afterSeq, limit)
	if err != nil {
		return nil, fmt.Errorf("pull sync entries: %w", err)
	}
	defer rows.Close()
	var out []*SyncEntry
	for rows.Next() {
		var e SyncEntry
		if err := rows.Scan(&e.ControllerDID, &e.TargetDeviceID, &e.Seq, &e.EntryType, &e.Ciphertext, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sync entry: %w", err)
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (p *PostgresDB) SyncHead(ctx context.Context, controllerDID, targetDeviceID string) (int64, error) {
	var seq int64
	err := p.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(seq), 0) FROM device_sync_log WHERE controller_did = $1 AND target_device_id = $2`,
		controllerDID, targetDeviceID).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("sync head: %w", err)
	}
	return seq, nil
}

func (p *PostgresDB) RevokeDeviceStream(ctx context.Context, controllerDID, targetDeviceID string) error {
	if _, err := p.pool.Exec(ctx,
		`INSERT INTO device_sync_revoked (controller_did, target_device_id, revoked_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (controller_did, target_device_id) DO UPDATE SET revoked_at = NOW()`,
		controllerDID, targetDeviceID); err != nil {
		return fmt.Errorf("revoke device stream: %w", err)
	}
	if _, err := p.pool.Exec(ctx,
		`DELETE FROM device_sync_log WHERE controller_did = $1 AND target_device_id = $2`,
		controllerDID, targetDeviceID); err != nil {
		return fmt.Errorf("purge sync log on revoke: %w", err)
	}
	return nil
}

func (p *PostgresDB) IsDeviceStreamActive(ctx context.Context, controllerDID, targetDeviceID string) (bool, error) {
	revoked, err := p.isStreamRevoked(ctx, controllerDID, targetDeviceID)
	if err != nil {
		return false, err
	}
	return !revoked, nil
}

func (p *PostgresDB) isStreamRevoked(ctx context.Context, controllerDID, targetDeviceID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM device_sync_revoked WHERE controller_did = $1 AND target_device_id = $2)`,
		controllerDID, targetDeviceID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check stream revoked: %w", err)
	}
	return exists, nil
}
