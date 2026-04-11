package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// RedisClient wraps go-redis for cache, blocklist, and session operations.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient connects to Redis and verifies the connection.
func NewRedisClient(ctx context.Context, cfg RedisConfig) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisClient{client: client}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// --- Token Blocklist ---

// BlocklistToken adds a JTI to the blocklist with TTL matching token expiry.
func (r *RedisClient) BlocklistToken(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, "blocklist:"+jti, "1", ttl).Err()
}

// IsBlocklisted checks if a JTI is on the blocklist.
func (r *RedisClient) IsBlocklisted(ctx context.Context, jti string) (bool, error) {
	val, err := r.client.Exists(ctx, "blocklist:"+jti).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// --- Cache ---

// CacheSet stores a value with TTL.
func (r *RedisClient) CacheSet(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, "cache:"+key, value, ttl).Err()
}

// CacheGet retrieves a cached value.
func (r *RedisClient) CacheGet(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, "cache:"+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// CacheDelete removes a cached value.
func (r *RedisClient) CacheDelete(ctx context.Context, key string) error {
	return r.client.Del(ctx, "cache:"+key).Err()
}

// --- Offline Message Queue Overflow ---

// QueuePush adds an encrypted blob to a recipient's overflow queue.
func (r *RedisClient) QueuePush(ctx context.Context, recipientDID string, blob []byte) error {
	key := "queue:" + recipientDID
	if err := r.client.RPush(ctx, key, blob).Err(); err != nil {
		return err
	}
	// Set 30-day expiry on the queue key
	return r.client.Expire(ctx, key, 30*24*time.Hour).Err()
}

// QueueDrain retrieves and removes all queued blobs for a recipient.
func (r *RedisClient) QueueDrain(ctx context.Context, recipientDID string) ([][]byte, error) {
	key := "queue:" + recipientDID
	pipe := r.client.Pipeline()
	lrange := pipe.LRange(ctx, key, 0, -1)
	pipe.Del(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	vals, err := lrange.Result()
	if err != nil {
		return nil, err
	}
	result := make([][]byte, len(vals))
	for i, v := range vals {
		result[i] = []byte(v)
	}
	return result, nil
}

// QueueDepth returns the number of queued messages for a recipient.
func (r *RedisClient) QueueDepth(ctx context.Context, recipientDID string) (int64, error) {
	return r.client.LLen(ctx, "queue:"+recipientDID).Result()
}

// --- Session Store ---

// SessionSet stores a session value with TTL.
func (r *RedisClient) SessionSet(ctx context.Context, sessionID string, data []byte, ttl time.Duration) error {
	return r.client.Set(ctx, "session:"+sessionID, data, ttl).Err()
}

// SessionGet retrieves a session value.
func (r *RedisClient) SessionGet(ctx context.Context, sessionID string) ([]byte, error) {
	val, err := r.client.Get(ctx, "session:"+sessionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// SessionDelete removes a session.
func (r *RedisClient) SessionDelete(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, "session:"+sessionID).Err()
}
