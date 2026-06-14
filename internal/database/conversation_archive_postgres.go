package database

import (
	"context"
	"fmt"
)

func (p *PostgresDB) SetConversationArchived(ctx context.Context, conversationID string, archived bool) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO conversation_archive (conversation_id, archived, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (conversation_id) DO UPDATE SET archived = EXCLUDED.archived, updated_at = NOW()`,
		conversationID, archived)
	if err != nil {
		return fmt.Errorf("set conversation archived: %w", err)
	}
	return nil
}

func (p *PostgresDB) IsConversationArchived(ctx context.Context, conversationID string) (bool, error) {
	var archived bool
	err := p.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT archived FROM conversation_archive WHERE conversation_id = $1), false)`,
		conversationID).Scan(&archived)
	if err != nil {
		return false, fmt.Errorf("is conversation archived: %w", err)
	}
	return archived, nil
}
