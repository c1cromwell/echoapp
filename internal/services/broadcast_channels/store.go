package broadcast_channels

import "context"

// ChannelStore persists channel data. The service maintains a hot in-memory
// cache so the same behavior is available when no durable store is configured.
type ChannelStore interface {
	SaveChannel(ctx context.Context, channel *Channel) error
	GetChannel(ctx context.Context, id string) (*Channel, error)
	ListChannels(ctx context.Context, limit, offset int) ([]*Channel, error)
	DeleteChannel(ctx context.Context, id string) error

	SavePost(ctx context.Context, post *ChannelPost) error
	GetPost(ctx context.Context, id string) (*ChannelPost, error)
	ListPosts(ctx context.Context, channelID string, limit, offset int) ([]*ChannelPost, error)
	DeletePost(ctx context.Context, id string) error

	SaveSubscriber(ctx context.Context, subscriber *ChannelSubscriber) error
	GetSubscriber(ctx context.Context, channelID, subscriberID string) (*ChannelSubscriber, error)
	ListSubscribers(ctx context.Context, channelID string) ([]*ChannelSubscriber, error)
	DeleteSubscriber(ctx context.Context, channelID, subscriberID string) error

	LoadAll(ctx context.Context) (channels []*Channel, posts []*ChannelPost, subs []*ChannelSubscriber, err error)
}
