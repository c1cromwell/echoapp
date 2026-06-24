package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// --- PostgresDB implementation of MessageOpsStore (WO-25/84/59) ---

func (p *PostgresDB) SetConversationRetention(ctx context.Context, conversationID string, retained bool) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO conversation_retention (conversation_id, retained, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (conversation_id) DO UPDATE SET retained = EXCLUDED.retained, updated_at = NOW()`,
		conversationID, retained)
	if err != nil {
		return fmt.Errorf("set conversation retention: %w", err)
	}
	return nil
}

func (p *PostgresDB) IsConversationRetained(ctx context.Context, conversationID string) (bool, error) {
	var retained bool
	err := p.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT retained FROM conversation_retention WHERE conversation_id = $1), false)`,
		conversationID).Scan(&retained)
	if err != nil {
		return false, fmt.Errorf("is conversation retained: %w", err)
	}
	return retained, nil
}

func (p *PostgresDB) AppendEditVersion(ctx context.Context, edit *MessageEdit) (int, error) {
	if edit == nil || edit.MessageID == "" {
		return 0, fmt.Errorf("edit message id required")
	}
	editedAt := edit.EditedAt
	if editedAt.IsZero() {
		editedAt = time.Now()
	}
	var version int
	err := p.pool.QueryRow(ctx,
		`INSERT INTO message_edits (message_id, version, conversation_id, editor_did, ciphertext, edited_at)
		 VALUES ($1,
		         (SELECT COALESCE(MAX(version), 0) + 1 FROM message_edits WHERE message_id = $1),
		         $2, $3, $4, $5)
		 RETURNING version`,
		edit.MessageID, edit.ConversationID, edit.EditorDID, edit.Ciphertext, editedAt).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("append edit version: %w", err)
	}
	return version, nil
}

func (p *PostgresDB) GetEditHistory(ctx context.Context, messageID string) ([]*MessageEdit, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT message_id, version, conversation_id, editor_did, ciphertext, edited_at
		   FROM message_edits WHERE message_id = $1 ORDER BY version ASC`, messageID)
	if err != nil {
		return nil, fmt.Errorf("get edit history: %w", err)
	}
	defer rows.Close()
	var out []*MessageEdit
	for rows.Next() {
		var e MessageEdit
		if err := rows.Scan(&e.MessageID, &e.Version, &e.ConversationID, &e.EditorDID, &e.Ciphertext, &e.EditedAt); err != nil {
			return nil, fmt.Errorf("scan edit: %w", err)
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (p *PostgresDB) MarkMessageDeleted(ctx context.Context, messageID string, retained bool) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO message_tombstones (message_id, deleted_at, retained)
		 VALUES ($1, NOW(), $2)
		 ON CONFLICT (message_id) DO UPDATE SET deleted_at = NOW(), retained = EXCLUDED.retained`,
		messageID, retained)
	if err != nil {
		return fmt.Errorf("mark message deleted: %w", err)
	}
	if !retained {
		// Litigation hold absent: edit history may be purged with the message.
		if _, err := p.pool.Exec(ctx, `DELETE FROM message_edits WHERE message_id = $1`, messageID); err != nil {
			return fmt.Errorf("purge edits on delete: %w", err)
		}
	}
	return nil
}

func (p *PostgresDB) IsMessageDeleted(ctx context.Context, messageID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM message_tombstones WHERE message_id = $1)`, messageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("is message deleted: %w", err)
	}
	return exists, nil
}

func (p *PostgresDB) PinMessage(ctx context.Context, conversationID, messageID, pinnerDID string) error {
	// Already pinned? idempotent success.
	var pinned bool
	if err := p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM message_pins WHERE conversation_id = $1 AND message_id = $2)`,
		conversationID, messageID).Scan(&pinned); err != nil {
		return fmt.Errorf("check pin: %w", err)
	}
	if pinned {
		return nil
	}
	var count int
	if err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM message_pins WHERE conversation_id = $1`, conversationID).Scan(&count); err != nil {
		return fmt.Errorf("count pins: %w", err)
	}
	if count >= MaxPinnedPerConversation {
		return ErrPinLimitReached
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO message_pins (conversation_id, message_id, pinner_did, pinned_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (conversation_id, message_id) DO NOTHING`,
		conversationID, messageID, pinnerDID)
	if err != nil {
		return fmt.Errorf("pin message: %w", err)
	}
	return nil
}

func (p *PostgresDB) UnpinMessage(ctx context.Context, conversationID, messageID string) error {
	_, err := p.pool.Exec(ctx,
		`DELETE FROM message_pins WHERE conversation_id = $1 AND message_id = $2`,
		conversationID, messageID)
	if err != nil {
		return fmt.Errorf("unpin message: %w", err)
	}
	return nil
}

func (p *PostgresDB) SetDisappearingTTL(ctx context.Context, conversationID string, ttlSeconds int) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO conversation_retention (conversation_id, disappearing_ttl_seconds, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (conversation_id) DO UPDATE SET disappearing_ttl_seconds = EXCLUDED.disappearing_ttl_seconds, updated_at = NOW()`,
		conversationID, ttlSeconds)
	if err != nil {
		return fmt.Errorf("set disappearing ttl: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetDisappearingTTL(ctx context.Context, conversationID string) (int, error) {
	var ttl int
	err := p.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT disappearing_ttl_seconds FROM conversation_retention WHERE conversation_id = $1), 0)`,
		conversationID).Scan(&ttl)
	if err != nil {
		return 0, fmt.Errorf("get disappearing ttl: %w", err)
	}
	return ttl, nil
}

func (p *PostgresDB) GetPinnedMessages(ctx context.Context, conversationID string) ([]*PinnedMessage, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT conversation_id, message_id, pinner_did, pinned_at
		   FROM message_pins WHERE conversation_id = $1 ORDER BY pinned_at ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get pinned messages: %w", err)
	}
	defer rows.Close()
	var out []*PinnedMessage
	for rows.Next() {
		var pm PinnedMessage
		if err := rows.Scan(&pm.ConversationID, &pm.MessageID, &pm.PinnerDID, &pm.PinnedAt); err != nil {
			return nil, fmt.Errorf("scan pin: %w", err)
		}
		out = append(out, &pm)
	}
	return out, rows.Err()
}

func (p *PostgresDB) PutMessageRefs(ctx context.Context, refs *MessageRefs) error {
	if refs == nil || refs.MessageID == "" || refs.ConversationID == "" || refs.AuthorDID == "" {
		return fmt.Errorf("message id, conversation id, and author did required")
	}
	createdAt := refs.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO message_refs (
			message_id, conversation_id, author_did,
			reply_to_message_id, forwarded_from_message_id, forwarded_from_conversation_id,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (message_id) DO UPDATE SET
			conversation_id = EXCLUDED.conversation_id,
			author_did = EXCLUDED.author_did,
			reply_to_message_id = EXCLUDED.reply_to_message_id,
			forwarded_from_message_id = EXCLUDED.forwarded_from_message_id,
			forwarded_from_conversation_id = EXCLUDED.forwarded_from_conversation_id`,
		refs.MessageID, refs.ConversationID, refs.AuthorDID,
		nullIfEmpty(refs.ReplyToMessageID),
		nullIfEmpty(refs.ForwardedFromMessageID),
		nullIfEmpty(refs.ForwardedFromConversationID),
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("put message refs: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetMessageRefs(ctx context.Context, messageID string) (*MessageRefs, error) {
	var refs MessageRefs
	var replyTo, fwdMsg, fwdConv *string
	err := p.pool.QueryRow(ctx, `
		SELECT message_id, conversation_id, author_did,
		       reply_to_message_id, forwarded_from_message_id, forwarded_from_conversation_id,
		       created_at
		FROM message_refs WHERE message_id = $1`, messageID).Scan(
		&refs.MessageID, &refs.ConversationID, &refs.AuthorDID,
		&replyTo, &fwdMsg, &fwdConv, &refs.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get message refs: %w", err)
	}
	if replyTo != nil {
		refs.ReplyToMessageID = *replyTo
	}
	if fwdMsg != nil {
		refs.ForwardedFromMessageID = *fwdMsg
	}
	if fwdConv != nil {
		refs.ForwardedFromConversationID = *fwdConv
	}
	return &refs, nil
}
