package database

import (
	"context"
	"fmt"
)

// accountPurgeStep is one DELETE issued when an account is erased. Query takes
// the account DID as $1.
type accountPurgeStep struct {
	Table string
	Query string
}

// accountPurgeSteps lists every server-side row that is personal to one account
// and is erased by PurgeAccount, children before parents.
//
// Deliberately NOT erased here — each needs a product or legal decision before
// it can be automated, so they are left visible in one place:
//
//   - message_queue rows the account SENT (sender_did): undelivered ciphertext
//     addressed to other people; it expires through the queue TTL.
//   - message_refs / message_edits / message_tombstones / evidence_fingerprints:
//     per-message metadata shared with the other members of a conversation.
//   - echo_groups / broadcast_channels the account OWNS, and its channel posts:
//     ownership has to be transferred or the space closed first.
//   - votes, staking_positions, staking_delegations: governance tallies and
//     token ledger mirrors of on-chain state.
//   - invite_links where the account was the invitee, and
//     passport_recovery_share rows where it is a guardian for someone else.
//   - comply_*: organisation compliance records under their own retention and
//     legal-hold rules.
//   - media object bytes in S3/IPFS: only the index rows are removed here.
var accountPurgeSteps = []accountPurgeStep{
	// Devices, push tokens and key bindings.
	{"user_devices", `DELETE FROM user_devices WHERE did = $1 OR user_id IN (SELECT user_id FROM users WHERE did = $1)`},
	{"identity_device", `DELETE FROM identity_device WHERE did = $1`},
	{"device_sync_log", `DELETE FROM device_sync_log WHERE controller_did = $1`},
	{"device_sync_revoked", `DELETE FROM device_sync_revoked WHERE controller_did = $1`},

	// Discovery, recovery and contacts (both directions: the address book the
	// account kept, and other people's entries pointing at it).
	{"contact_discovery_index", `DELETE FROM contact_discovery_index WHERE did = $1`},
	{"sms_recovery", `DELETE FROM sms_recovery WHERE did = $1`},
	{"contacts", `DELETE FROM contacts WHERE owner_did = $1 OR contact_did = $1`},
	{"safety_reports", `DELETE FROM safety_reports WHERE reporter_did = $1 OR subject_did = $1`},
	{"invite_links", `DELETE FROM invite_links WHERE inviter_did = $1`},

	// Messaging state held for the account.
	{"message_queue", `DELETE FROM message_queue WHERE recipient_did = $1`},
	{"delivery_receipts", `DELETE FROM delivery_receipts WHERE recipient_did = $1`},
	{"message_reactions", `DELETE FROM message_reactions WHERE reactor_did = $1`},
	{"message_pins", `DELETE FROM message_pins WHERE pinner_did = $1`},
	{"message_backup_blob", `DELETE FROM message_backup_blob WHERE holder_did = $1`},
	{"media_chunks", `DELETE FROM media_chunks WHERE file_id IN (SELECT file_id FROM media_files WHERE uploader_did = $1)`},
	{"media_files", `DELETE FROM media_files WHERE uploader_did = $1`},

	// Memberships and community activity.
	{"echo_group_members", `DELETE FROM echo_group_members WHERE member_id = $1`},
	{"broadcast_channel_subscribers", `DELETE FROM broadcast_channel_subscribers WHERE subscriber_id = $1`},
	{"broadcast_channel_join_requests", `DELETE FROM broadcast_channel_join_requests WHERE subscriber_id = $1`},
	{"broadcast_channel_comments", `DELETE FROM broadcast_channel_comments WHERE author_id = $1`},

	// Preferences.
	{"user_prefs", `DELETE FROM user_prefs WHERE did = $1`},
	{"chat_folders", `DELETE FROM chat_folders WHERE did = $1`},
	{"notification_preferences", `DELETE FROM notification_preferences WHERE did = $1`},

	// Credentials, passport and trust.
	{"passport_recovery_share", `DELETE FROM passport_recovery_share WHERE holder_did = $1`},
	{"passport_recovery_session", `DELETE FROM passport_recovery_session WHERE holder_did = $1`},
	{"passport_recovery_policy", `DELETE FROM passport_recovery_policy WHERE holder_did = $1`},
	{"passport_credential_ref", `DELETE FROM passport_credential_ref WHERE holder_did = $1`},
	{"passport_sync_blob", `DELETE FROM passport_sync_blob WHERE holder_did = $1`},
	{"credentials", `DELETE FROM credentials WHERE did = $1`},
	{"trust_scores", `DELETE FROM trust_scores WHERE did = $1`},

	// Rewards and the wallet link (the DID-to-address mapping and cached
	// balances; nothing on-chain is touched).
	{"daily_reward_caps", `DELETE FROM daily_reward_caps WHERE did = $1`},
	{"quest_completions", `DELETE FROM quest_completions WHERE did = $1`},
	{"wallet_balance_cache", `DELETE FROM wallet_balance_cache WHERE did = $1`},
	{"wallet_accounts", `DELETE FROM wallet_accounts WHERE did = $1`},

	// The account row itself, last.
	{"users", `DELETE FROM users WHERE did = $1`},
}

// PurgeAccount erases the server-side personal data held for did in a single
// transaction and returns the number of rows removed per table. A table that
// does not exist yet (older schema) is skipped, so the purge keeps working
// across migrations. Either every step commits or none does.
func (p *PostgresDB) PurgeAccount(ctx context.Context, did string) (map[string]int64, error) {
	if did == "" {
		return nil, fmt.Errorf("purge account: empty did")
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("purge account: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	removed := make(map[string]int64, len(accountPurgeSteps))
	for _, step := range accountPurgeSteps {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT to_regclass('public.' || $1) IS NOT NULL`, step.Table).Scan(&exists); err != nil {
			return nil, fmt.Errorf("purge account: probe %s: %w", step.Table, err)
		}
		if !exists {
			continue
		}
		tag, err := tx.Exec(ctx, step.Query, did)
		if err != nil {
			return nil, fmt.Errorf("purge account: %s: %w", step.Table, err)
		}
		if n := tag.RowsAffected(); n > 0 {
			removed[step.Table] += n
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("purge account: commit: %w", err)
	}
	return removed, nil
}
