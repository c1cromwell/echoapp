package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB implements DB using a pgx connection pool.
type PostgresDB struct {
	pool *pgxpool.Pool
}

// PostgresConfig holds database connection settings.
type PostgresConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (c PostgresConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, sslMode)
}

// NewPostgresDB connects to PostgreSQL and returns a DB implementation.
func NewPostgresDB(ctx context.Context, cfg PostgresConfig) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}
	return &PostgresDB{pool: pool}, nil
}

// Close closes the connection pool.
func (p *PostgresDB) Close() {
	p.pool.Close()
}

// Pool returns the underlying pgxpool for direct access (migrations, etc.).
func (p *PostgresDB) Pool() *pgxpool.Pool {
	return p.pool
}

// --- UserStore ---

func (p *PostgresDB) CreateUser(ctx context.Context, user *User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.TrustTier == 0 {
		user.TrustTier = 1
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO users (user_id, did, username, trust_tier, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.UserID, user.DID, user.Username, user.TrustTier, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetUserByDID(ctx context.Context, did string) (*User, error) {
	var u User
	err := p.pool.QueryRow(ctx,
		`SELECT user_id, did, username, trust_tier, created_at, updated_at
		 FROM users WHERE did = $1`, did).
		Scan(&u.UserID, &u.DID, &u.Username, &u.TrustTier, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by did: %w", err)
	}
	return &u, nil
}

func (p *PostgresDB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := p.pool.QueryRow(ctx,
		`SELECT user_id, did, username, trust_tier, created_at, updated_at
		 FROM users WHERE username = $1`, username).
		Scan(&u.UserID, &u.DID, &u.Username, &u.TrustTier, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

// --- TrustScoreStore ---

func (p *PostgresDB) SetTrustScore(ctx context.Context, ts *TrustScore) error {
	ts.UpdatedAt = time.Now()
	if ts.ExpiresAt.IsZero() {
		ts.ExpiresAt = time.Now().Add(trustScoreTTL)
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO trust_scores (did, score, tier, issued_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (did) DO UPDATE SET score=$2, tier=$3, issued_at=$4, expires_at=$5`,
		ts.DID, int(ts.Score), ts.Tier, ts.UpdatedAt, ts.ExpiresAt)
	if err != nil {
		return fmt.Errorf("set trust score: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetTrustScore(ctx context.Context, did string) (*TrustScore, error) {
	var ts TrustScore
	err := p.pool.QueryRow(ctx,
		`SELECT did, score, tier, issued_at, expires_at
		 FROM trust_scores WHERE did = $1 AND expires_at > NOW()`, did).
		Scan(&ts.DID, &ts.Score, &ts.Tier, &ts.UpdatedAt, &ts.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get trust score: %w", err)
	}
	return &ts, nil
}

func (p *PostgresDB) GetTrustScoreBatch(ctx context.Context, dids []string) ([]*TrustScore, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT did, score, tier, issued_at, expires_at
		 FROM trust_scores WHERE did = ANY($1) AND expires_at > NOW()`, dids)
	if err != nil {
		return nil, fmt.Errorf("get trust scores batch: %w", err)
	}
	defer rows.Close()
	var results []*TrustScore
	for rows.Next() {
		var ts TrustScore
		if err := rows.Scan(&ts.DID, &ts.Score, &ts.Tier, &ts.UpdatedAt, &ts.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan trust score: %w", err)
		}
		results = append(results, &ts)
	}
	return results, rows.Err()
}

// --- CredentialStore ---

func (p *PostgresDB) StoreCredential(ctx context.Context, cred *Credential) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO credentials (credential_id, did, issuer_did, credential_type, status, issued_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		cred.CredentialID, cred.SubjectDID, cred.IssuerDID, cred.SchemaID, cred.Status, cred.IssuedAt, cred.ExpiresAt)
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("store credential: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetCredentialsByDID(ctx context.Context, did string) ([]*Credential, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT credential_id, issuer_did, did, credential_type, status, issued_at, expires_at
		 FROM credentials WHERE did = $1`, did)
	if err != nil {
		return nil, fmt.Errorf("get credentials: %w", err)
	}
	defer rows.Close()
	var results []*Credential
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.CredentialID, &c.IssuerDID, &c.SubjectDID, &c.SchemaID, &c.Status, &c.IssuedAt, &c.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan credential: %w", err)
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (p *PostgresDB) RevokeCredential(ctx context.Context, credentialID string) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE credentials SET status = 'revoked', expires_at = NOW() WHERE credential_id = $1`, credentialID)
	if err != nil {
		return fmt.Errorf("revoke credential: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- MessageQueueStore ---

func (p *PostgresDB) Enqueue(ctx context.Context, msg *QueuedMessage) error {
	msg.CreatedAt = time.Now()
	if msg.ExpiresAt.IsZero() {
		msg.ExpiresAt = msg.CreatedAt.Add(30 * 24 * time.Hour)
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO message_queue (message_id, conversation_id, sender_did, recipient_did, encrypted_payload, status, queued_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		msg.MessageID, "", msg.SenderDID, msg.RecipientDID, msg.Payload, "queued", msg.CreatedAt, msg.ExpiresAt)
	if err != nil {
		return fmt.Errorf("enqueue message: %w", err)
	}
	return nil
}

func (p *PostgresDB) Dequeue(ctx context.Context, recipientDID string, limit int) ([]*QueuedMessage, error) {
	rows, err := p.pool.Query(ctx,
		`UPDATE message_queue SET status = 'delivered', delivered_at = NOW()
		 WHERE message_id IN (
		   SELECT message_id FROM message_queue
		   WHERE recipient_did = $1 AND status = 'queued' AND expires_at > NOW()
		   ORDER BY queued_at ASC LIMIT $2
		 )
		 RETURNING message_id, sender_did, recipient_did, encrypted_payload, status, queued_at, expires_at`,
		recipientDID, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeue messages: %w", err)
	}
	defer rows.Close()
	var results []*QueuedMessage
	for rows.Next() {
		var m QueuedMessage
		if err := rows.Scan(&m.MessageID, &m.SenderDID, &m.RecipientDID, &m.Payload, &m.Status, &m.CreatedAt, &m.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		results = append(results, &m)
	}
	return results, rows.Err()
}

func (p *PostgresDB) MarkDelivered(ctx context.Context, messageID string) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE message_queue SET status = 'delivered' WHERE message_id = $1`, messageID)
	if err != nil {
		return fmt.Errorf("mark delivered: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PostgresDB) PurgeExpired(ctx context.Context) (int, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM message_queue WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("purge expired: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// --- MerkleBatchStore ---

func (p *PostgresDB) CreateBatch(ctx context.Context, batch *MerkleBatch) error {
	batch.CreatedAt = time.Now()
	_, err := p.pool.Exec(ctx,
		`INSERT INTO merkle_batches (batch_id, merkle_root, commitment_count, time_range_from, time_range_to, status, created_at)
		 VALUES ($1, $2, $3, NOW(), NOW(), $4, $5)`,
		batch.BatchID, batch.RootHash, batch.LeafCount, "pending", batch.CreatedAt)
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetBatch(ctx context.Context, batchID string) (*MerkleBatch, error) {
	var b MerkleBatch
	err := p.pool.QueryRow(ctx,
		`SELECT batch_id, merkle_root, commitment_count, status, created_at
		 FROM merkle_batches WHERE batch_id = $1`, batchID).
		Scan(&b.BatchID, &b.RootHash, &b.LeafCount, &b.Status, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get batch: %w", err)
	}
	return &b, nil
}

func (p *PostgresDB) FinalizeBatch(ctx context.Context, batchID, rootHash string, leafCount int) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE merkle_batches SET merkle_root=$2, commitment_count=$3, status='finalized', finalized_at=NOW() WHERE batch_id=$1`,
		batchID, rootHash, leafCount)
	if err != nil {
		return fmt.Errorf("finalize batch: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- ContactStore ---

func (p *PostgresDB) AddContact(ctx context.Context, contact *Contact) error {
	contact.CreatedAt = time.Now()
	id := contact.OwnerDID + ":" + contact.ContactDID
	status := "active"
	if contact.Blocked {
		status = "blocked"
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO contacts (id, owner_did, contact_did, added_via, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, contact.OwnerDID, contact.ContactDID, contact.AddedVia, status, contact.CreatedAt)
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("add contact: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetContacts(ctx context.Context, ownerDID string) ([]*Contact, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT owner_did, contact_did, added_via, (status = 'blocked'), COALESCE(trust_tier::text, ''), created_at
		 FROM contacts WHERE owner_did = $1 AND status != 'removed'`, ownerDID)
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}
	defer rows.Close()
	var results []*Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.OwnerDID, &c.ContactDID, &c.AddedVia, &c.Blocked, &c.TrustBadge, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (p *PostgresDB) RemoveContact(ctx context.Context, ownerDID, contactDID string) error {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM contacts WHERE owner_did = $1 AND contact_did = $2`, ownerDID, contactDID)
	if err != nil {
		return fmt.Errorf("remove contact: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PostgresDB) SetBlocked(ctx context.Context, ownerDID, contactDID string, blocked bool) error {
	status := "active"
	if blocked {
		status = "blocked"
	}
	tag, err := p.pool.Exec(ctx,
		`UPDATE contacts SET status = $3, updated_at = NOW() WHERE owner_did = $1 AND contact_did = $2`,
		ownerDID, contactDID, status)
	if err != nil {
		return fmt.Errorf("set blocked: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PostgresDB) GetContactCount(ctx context.Context, ownerDID string) (int, error) {
	var count int
	err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM contacts WHERE owner_did = $1 AND status = 'active'`, ownerDID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get contact count: %w", err)
	}
	return count, nil
}

// --- InviteStore ---

func (p *PostgresDB) CreateInvite(ctx context.Context, invite *InviteLink) error {
	invite.CreatedAt = time.Now()
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = invite.CreatedAt.Add(30 * 24 * time.Hour)
	}
	id := "inv_" + invite.Code
	_, err := p.pool.Exec(ctx,
		`INSERT INTO invite_links (invite_id, inviter_did, invite_code, status, created_at, expires_at)
		 VALUES ($1, $2, $3, 'pending', $4, $5)`,
		id, invite.CreatorDID, invite.Code, invite.CreatedAt, invite.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetInvite(ctx context.Context, code string) (*InviteLink, error) {
	var inv InviteLink
	var acceptedBy *string
	err := p.pool.QueryRow(ctx,
		`SELECT invite_code, inviter_did, invitee_did, (status = 'accepted'), created_at, expires_at
		 FROM invite_links WHERE invite_code = $1`, code).
		Scan(&inv.Code, &inv.CreatorDID, &acceptedBy, &inv.Accepted, &inv.CreatedAt, &inv.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get invite: %w", err)
	}
	if acceptedBy != nil {
		inv.AcceptedBy = *acceptedBy
	}
	return &inv, nil
}

func (p *PostgresDB) AcceptInvite(ctx context.Context, code, acceptedBy string) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE invite_links SET status = 'accepted', invitee_did = $2, accepted_at = NOW()
		 WHERE invite_code = $1 AND status = 'pending' AND expires_at > NOW()`,
		code, acceptedBy)
	if err != nil {
		return fmt.Errorf("accept invite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- MediaStore ---

func (p *PostgresDB) StoreMediaFile(ctx context.Context, file *MediaFile) error {
	file.CreatedAt = time.Now()
	_, err := p.pool.Exec(ctx,
		`INSERT INTO media_files (file_id, uploader_did, content_type, encrypted_size, storage_backend, storage_key, chunk_count, scan_status, created_at)
		 VALUES ($1, $2, $3, $4, 'storj', $5, $6, $7, $8)`,
		file.FileID, file.UploaderDID, file.ContentType, file.EncryptedSize, file.FileID, file.ChunkCount, file.ScanStatus, file.CreatedAt)
	if err != nil {
		return fmt.Errorf("store media file: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetMediaFile(ctx context.Context, fileID string) (*MediaFile, error) {
	var f MediaFile
	err := p.pool.QueryRow(ctx,
		`SELECT file_id, uploader_did, content_type, encrypted_size, chunk_count, scan_status, created_at
		 FROM media_files WHERE file_id = $1`, fileID).
		Scan(&f.FileID, &f.UploaderDID, &f.ContentType, &f.EncryptedSize, &f.ChunkCount, &f.ScanStatus, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get media file: %w", err)
	}
	return &f, nil
}

func (p *PostgresDB) UpdateScanStatus(ctx context.Context, fileID, status string) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE media_files SET scan_status = $2 WHERE file_id = $1`, fileID, status)
	if err != nil {
		return fmt.Errorf("update scan status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PostgresDB) StoreChunk(ctx context.Context, chunk *MediaChunk) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO media_chunks (chunk_id, file_id, chunk_index, ipfs_cid, size_bytes)
		 VALUES ($1, $2, $3, $4, $5)`,
		chunk.ChunkID, chunk.FileID, chunk.Index, chunk.Checksum, chunk.Size)
	if err != nil {
		return fmt.Errorf("store chunk: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetChunks(ctx context.Context, fileID string) ([]*MediaChunk, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT chunk_id, file_id, chunk_index, size_bytes, ipfs_cid
		 FROM media_chunks WHERE file_id = $1 ORDER BY chunk_index`, fileID)
	if err != nil {
		return nil, fmt.Errorf("get chunks: %w", err)
	}
	defer rows.Close()
	var results []*MediaChunk
	for rows.Next() {
		var c MediaChunk
		if err := rows.Scan(&c.ChunkID, &c.FileID, &c.Index, &c.Size, &c.Checksum); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

// --- NotificationStore ---

func (p *PostgresDB) RegisterDevice(ctx context.Context, device *UserDevice) error {
	device.CreatedAt = time.Now()
	_, err := p.pool.Exec(ctx,
		`INSERT INTO user_devices (device_id, user_id, did, device_label, public_key, apns_token, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (device_id) DO UPDATE SET apns_token=$6, device_label=$4`,
		device.DeviceID, device.DID, device.DID, device.DeviceLabel, device.PublicKey, device.APNsToken, device.CreatedAt)
	if err != nil {
		return fmt.Errorf("register device: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetDevicesByDID(ctx context.Context, did string) ([]*UserDevice, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT device_id, did, device_label, public_key, apns_token, created_at
		 FROM user_devices WHERE did = $1`, did)
	if err != nil {
		return nil, fmt.Errorf("get devices: %w", err)
	}
	defer rows.Close()
	var results []*UserDevice
	for rows.Next() {
		var d UserDevice
		if err := rows.Scan(&d.DeviceID, &d.DID, &d.DeviceLabel, &d.PublicKey, &d.APNsToken, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}

func (p *PostgresDB) UpdateAPNsToken(ctx context.Context, deviceID, token string) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE user_devices SET apns_token = $2 WHERE device_id = $1`, deviceID, token)
	if err != nil {
		return fmt.Errorf("update apns token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PostgresDB) GetNotificationPrefs(ctx context.Context, did string) (*NotificationPrefs, error) {
	var np NotificationPrefs
	err := p.pool.QueryRow(ctx,
		`SELECT did, push_enabled, message_preview, group_notifications, channel_notifications,
		        EXTRACT(HOUR FROM quiet_hours_start)::int, EXTRACT(HOUR FROM quiet_hours_end)::int
		 FROM notification_preferences WHERE did = $1`, did).
		Scan(&np.DID, &np.PushEnabled, &np.MessagePreview, &np.GroupNotifications,
			&np.ChannelNotifications, &np.QuietHoursStart, &np.QuietHoursEnd)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get notification prefs: %w", err)
	}
	return &np, nil
}

func (p *PostgresDB) UpdateNotificationPrefs(ctx context.Context, prefs *NotificationPrefs) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO notification_preferences (did, push_enabled, message_preview, group_notifications, channel_notifications)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (did) DO UPDATE SET push_enabled=$2, message_preview=$3,
		   group_notifications=$4, channel_notifications=$5, updated_at=NOW()`,
		prefs.DID, prefs.PushEnabled, prefs.MessagePreview, prefs.GroupNotifications, prefs.ChannelNotifications)
	if err != nil {
		return fmt.Errorf("update notification prefs: %w", err)
	}
	return nil
}

// --- LogIndexStore ---

func (p *PostgresDB) StoreLogIndex(ctx context.Context, entry *LogIndexEntry) error {
	entry.CreatedAt = time.Now()
	_, err := p.pool.Exec(ctx,
		`INSERT INTO logs_index (batch_id, cid, time_range_from, time_range_to, digest, event_count, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.EntryID, entry.DID, entry.TimeRangeFrom, entry.TimeRangeTo, entry.EventType, 1, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("store log index: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetLogIndex(ctx context.Context, from, to time.Time) ([]*LogIndexEntry, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT batch_id, cid, time_range_from, time_range_to, created_at
		 FROM logs_index WHERE time_range_from >= $1 AND time_range_to <= $2
		 ORDER BY created_at`, from, to)
	if err != nil {
		return nil, fmt.Errorf("get log index: %w", err)
	}
	defer rows.Close()
	var results []*LogIndexEntry
	for rows.Next() {
		var e LogIndexEntry
		if err := rows.Scan(&e.EntryID, &e.DID, &e.TimeRangeFrom, &e.TimeRangeTo, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan log index: %w", err)
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// isDuplicateError checks for PostgreSQL unique violation (23505).
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrDuplicate) || (len(err.Error()) > 0 && contains23505(err.Error()))
}

func contains23505(s string) bool {
	return len(s) > 5 && (indexOf(s, "23505") >= 0 || indexOf(s, "duplicate key") >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
