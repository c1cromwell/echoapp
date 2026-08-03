package broadcast_channels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists broadcast channels using migration 028.
type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (p *PostgresStore) SaveChannel(ctx context.Context, channel *Channel) error {
	if channel == nil || channel.ID == "" {
		return errors.New("channel id required")
	}
	tags, err := json.Marshal(channel.Tags)
	if err != nil {
		return fmt.Errorf("marshal channel tags: %w", err)
	}
	_, err = p.pool.Exec(ctx, `
		INSERT INTO broadcast_channels (
			id, creator_id, name, topic, description, visibility_mode, channel_type,
			cover_image_url, tags, is_active, is_muted, created_at, last_post_at,
			subscriber_count, total_post_count, language, website, trust_score,
			verification_status, allow_comments, allow_polls, allow_links, allow_media, require_approval
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
		)
		ON CONFLICT (id) DO UPDATE SET
			creator_id=EXCLUDED.creator_id, name=EXCLUDED.name, topic=EXCLUDED.topic,
			description=EXCLUDED.description, visibility_mode=EXCLUDED.visibility_mode,
			channel_type=EXCLUDED.channel_type, cover_image_url=EXCLUDED.cover_image_url,
			tags=EXCLUDED.tags, is_active=EXCLUDED.is_active, is_muted=EXCLUDED.is_muted,
			last_post_at=EXCLUDED.last_post_at, subscriber_count=EXCLUDED.subscriber_count,
			total_post_count=EXCLUDED.total_post_count, language=EXCLUDED.language,
			website=EXCLUDED.website, trust_score=EXCLUDED.trust_score,
			verification_status=EXCLUDED.verification_status, allow_comments=EXCLUDED.allow_comments,
			allow_polls=EXCLUDED.allow_polls, allow_links=EXCLUDED.allow_links,
			allow_media=EXCLUDED.allow_media, require_approval=EXCLUDED.require_approval`,
		channel.ID, channel.CreatorID, channel.Name, channel.Topic, channel.Description,
		string(channel.VisibilityMode), string(channel.ChannelType), channel.CoverImageURL, tags,
		channel.IsActive, channel.IsMuted, channel.CreatedAt, channel.LastPostAt,
		channel.SubscriberCount, channel.TotalPostCount, channel.Language, channel.Website,
		channel.TrustScore, string(channel.VerificationStatus), channel.AllowComments,
		channel.AllowPolls, channel.AllowLinks, channel.AllowMedia, channel.RequireApproval,
	)
	return err
}

func (p *PostgresStore) GetChannel(ctx context.Context, id string) (*Channel, error) {
	channel, err := scanChannel(p.pool.QueryRow(ctx, channelSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return channel, err
}

func (p *PostgresStore) ListChannels(ctx context.Context, limit, offset int) ([]*Channel, error) {
	rows, err := p.pool.Query(ctx, channelSelect+` ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectChannels(rows)
}

func (p *PostgresStore) DeleteChannel(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM broadcast_channels WHERE id = $1`, id)
	return err
}

func (p *PostgresStore) SavePost(ctx context.Context, post *ChannelPost) error {
	if post == nil || post.ID == "" || post.ChannelID == "" {
		return errors.New("post and channel id required")
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO broadcast_channel_posts (
			id, channel_id, creator_id, content, content_type, encrypted_content,
			like_count, comment_count, share_count, published_at, scheduled_for,
			publish_status, is_pinned, is_featured, is_sponsored, allow_replies,
			flag_count, mod_status, mod_notes, edit_count, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22
		)
		ON CONFLICT (id) DO UPDATE SET
			channel_id=EXCLUDED.channel_id, creator_id=EXCLUDED.creator_id, content=EXCLUDED.content,
			content_type=EXCLUDED.content_type, encrypted_content=EXCLUDED.encrypted_content,
			like_count=EXCLUDED.like_count, comment_count=EXCLUDED.comment_count,
			share_count=EXCLUDED.share_count, published_at=EXCLUDED.published_at,
			scheduled_for=EXCLUDED.scheduled_for, publish_status=EXCLUDED.publish_status,
			is_pinned=EXCLUDED.is_pinned, is_featured=EXCLUDED.is_featured,
			is_sponsored=EXCLUDED.is_sponsored, allow_replies=EXCLUDED.allow_replies,
			flag_count=EXCLUDED.flag_count, mod_status=EXCLUDED.mod_status,
			mod_notes=EXCLUDED.mod_notes, edit_count=EXCLUDED.edit_count, updated_at=EXCLUDED.updated_at`,
		post.ID, post.ChannelID, post.CreatorID, post.Content, string(post.ContentType), post.EncryptedContent,
		post.LikeCount, post.CommentCount, post.ShareCount, post.PublishedAt, post.ScheduledFor,
		string(post.PublishStatus), post.IsPinned, post.IsFeatured, post.IsSponsored, post.AllowReplies,
		post.FlagCount, string(post.ModStatus), post.ModNotes, post.EditCount, post.CreatedAt, post.UpdatedAt,
	)
	return err
}

func (p *PostgresStore) GetPost(ctx context.Context, id string) (*ChannelPost, error) {
	post, err := scanPost(p.pool.QueryRow(ctx, postSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return post, err
}

func (p *PostgresStore) ListPosts(ctx context.Context, channelID string, limit, offset int) ([]*ChannelPost, error) {
	rows, err := p.pool.Query(ctx, postSelect+` WHERE channel_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPosts(rows)
}

func (p *PostgresStore) DeletePost(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM broadcast_channel_posts WHERE id = $1`, id)
	return err
}

func (p *PostgresStore) SaveSubscriber(ctx context.Context, subscriber *ChannelSubscriber) error {
	if subscriber == nil || subscriber.ChannelID == "" || subscriber.SubscriberID == "" {
		return errors.New("channel and subscriber id required")
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO broadcast_channel_subscribers (
			id, channel_id, subscriber_id, joined_at, subscription_tier,
			subscription_expires_at, auto_renew, role, last_seen_at, notification_mode,
			is_muted, is_blocked, post_count, comment_count, like_count, trust_score, moderation_flags
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (channel_id, subscriber_id) DO UPDATE SET
			id=EXCLUDED.id, joined_at=EXCLUDED.joined_at, subscription_tier=EXCLUDED.subscription_tier,
			subscription_expires_at=EXCLUDED.subscription_expires_at, auto_renew=EXCLUDED.auto_renew,
			role=EXCLUDED.role, last_seen_at=EXCLUDED.last_seen_at, notification_mode=EXCLUDED.notification_mode,
			is_muted=EXCLUDED.is_muted, is_blocked=EXCLUDED.is_blocked, post_count=EXCLUDED.post_count,
			comment_count=EXCLUDED.comment_count, like_count=EXCLUDED.like_count,
			trust_score=EXCLUDED.trust_score, moderation_flags=EXCLUDED.moderation_flags`,
		subscriber.ID, subscriber.ChannelID, subscriber.SubscriberID, subscriber.JoinedAt,
		string(subscriber.SubscriptionTier), subscriber.SubscriptionExpiresAt, subscriber.AutoRenew,
		string(subscriber.Role), subscriber.LastSeenAt, string(subscriber.NotificationMode),
		subscriber.IsMuted, subscriber.IsBlocked, subscriber.PostCount, subscriber.CommentCount,
		subscriber.LikeCount, subscriber.TrustScore, subscriber.ModerationFlags,
	)
	return err
}

func (p *PostgresStore) GetSubscriber(ctx context.Context, channelID, subscriberID string) (*ChannelSubscriber, error) {
	subscriber, err := scanSubscriber(p.pool.QueryRow(ctx, subscriberSelect+` WHERE channel_id = $1 AND subscriber_id = $2`, channelID, subscriberID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return subscriber, err
}

func (p *PostgresStore) ListSubscribers(ctx context.Context, channelID string) ([]*ChannelSubscriber, error) {
	rows, err := p.pool.Query(ctx, subscriberSelect+` WHERE channel_id = $1 ORDER BY joined_at`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectSubscribers(rows)
}

func (p *PostgresStore) DeleteSubscriber(ctx context.Context, channelID, subscriberID string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM broadcast_channel_subscribers WHERE channel_id = $1 AND subscriber_id = $2`, channelID, subscriberID)
	return err
}

func (p *PostgresStore) LoadAll(ctx context.Context) ([]*Channel, []*ChannelPost, []*ChannelSubscriber, error) {
	channels, err := p.ListChannels(ctx, 1_000_000, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, err := p.pool.Query(ctx, postSelect+` ORDER BY created_at`)
	if err != nil {
		return nil, nil, nil, err
	}
	posts, err := collectPosts(rows)
	rows.Close()
	if err != nil {
		return nil, nil, nil, err
	}
	rows, err = p.pool.Query(ctx, subscriberSelect+` ORDER BY joined_at`)
	if err != nil {
		return nil, nil, nil, err
	}
	subs, err := collectSubscribers(rows)
	rows.Close()
	return channels, posts, subs, err
}

const channelSelect = `SELECT id, creator_id, name, topic, description, visibility_mode, channel_type,
	cover_image_url, tags, is_active, is_muted, created_at, last_post_at, subscriber_count,
	total_post_count, language, website, trust_score, verification_status, allow_comments,
	allow_polls, allow_links, allow_media, require_approval FROM broadcast_channels`
const postSelect = `SELECT id, channel_id, creator_id, content, content_type, encrypted_content,
	like_count, comment_count, share_count, published_at, scheduled_for, publish_status, is_pinned,
	is_featured, is_sponsored, allow_replies, flag_count, mod_status, mod_notes, edit_count,
	created_at, updated_at FROM broadcast_channel_posts`
const subscriberSelect = `SELECT id, channel_id, subscriber_id, joined_at, subscription_tier,
	subscription_expires_at, auto_renew, role, last_seen_at, notification_mode, is_muted, is_blocked,
	post_count, comment_count, like_count, trust_score, moderation_flags FROM broadcast_channel_subscribers`

type channelRow interface{ Scan(...any) error }

func scanChannel(row channelRow) (*Channel, error) {
	var channel Channel
	var tags []byte
	var visibility, channelType, verification string
	err := row.Scan(&channel.ID, &channel.CreatorID, &channel.Name, &channel.Topic, &channel.Description,
		&visibility, &channelType, &channel.CoverImageURL, &tags, &channel.IsActive, &channel.IsMuted,
		&channel.CreatedAt, &channel.LastPostAt, &channel.SubscriberCount, &channel.TotalPostCount,
		&channel.Language, &channel.Website, &channel.TrustScore, &verification, &channel.AllowComments,
		&channel.AllowPolls, &channel.AllowLinks, &channel.AllowMedia, &channel.RequireApproval)
	if err == nil {
		channel.VisibilityMode = VisibilityMode(visibility)
		channel.ChannelType = ChannelType(channelType)
		channel.VerificationStatus = VerificationStatus(verification)
		err = json.Unmarshal(tags, &channel.Tags)
	}
	return &channel, err
}

func scanPost(row channelRow) (*ChannelPost, error) {
	var post ChannelPost
	var contentType, status, modStatus string
	err := row.Scan(&post.ID, &post.ChannelID, &post.CreatorID, &post.Content, &contentType,
		&post.EncryptedContent, &post.LikeCount, &post.CommentCount, &post.ShareCount,
		&post.PublishedAt, &post.ScheduledFor, &status, &post.IsPinned, &post.IsFeatured,
		&post.IsSponsored, &post.AllowReplies, &post.FlagCount, &modStatus, &post.ModNotes,
		&post.EditCount, &post.CreatedAt, &post.UpdatedAt)
	if err == nil {
		post.ContentType, post.PublishStatus, post.ModStatus = ContentType(contentType), PublishStatus(status), ModStatus(modStatus)
	}
	return &post, err
}

func scanSubscriber(row channelRow) (*ChannelSubscriber, error) {
	var subscriber ChannelSubscriber
	var tier, role, notification string
	err := row.Scan(&subscriber.ID, &subscriber.ChannelID, &subscriber.SubscriberID, &subscriber.JoinedAt,
		&tier, &subscriber.SubscriptionExpiresAt, &subscriber.AutoRenew, &role, &subscriber.LastSeenAt,
		&notification, &subscriber.IsMuted, &subscriber.IsBlocked, &subscriber.PostCount,
		&subscriber.CommentCount, &subscriber.LikeCount, &subscriber.TrustScore, &subscriber.ModerationFlags)
	if err == nil {
		subscriber.SubscriptionTier, subscriber.Role, subscriber.NotificationMode =
			SubscriptionTier(tier), SubscriberRole(role), NotificationMode(notification)
	}
	return &subscriber, err
}

func collectChannels(rows pgx.Rows) ([]*Channel, error) {
	defer rows.Close()
	var channels []*Channel
	for rows.Next() {
		channel, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, rows.Err()
}

func collectPosts(rows pgx.Rows) ([]*ChannelPost, error) {
	defer rows.Close()
	var posts []*ChannelPost
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func collectSubscribers(rows pgx.Rows) ([]*ChannelSubscriber, error) {
	defer rows.Close()
	var subscribers []*ChannelSubscriber
	for rows.Next() {
		subscriber, err := scanSubscriber(rows)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	return subscribers, rows.Err()
}
