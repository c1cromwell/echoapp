package infra

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// rateLimitScript is an atomic sliding-window-log limiter over a sorted set:
// drop entries older than the window, count what remains, and admit (recording a
// new entry) only if under the limit. Running it server-side in one EVAL makes the
// decision correct across many API instances sharing one Redis.
//
// KEYS[1] = bucket key
// ARGV: now(ms), window(ms), limit, unique-member
// returns {allowed(0|1), remaining, retryAfter(ms)}
var rateLimitScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, now - window)
local count = redis.call('ZCARD', KEYS[1])
if count < limit then
  redis.call('ZADD', KEYS[1], now, ARGV[4])
  redis.call('PEXPIRE', KEYS[1], window)
  return {1, limit - count - 1, 0}
end
local oldest = redis.call('ZRANGE', KEYS[1], 0, 0, 'WITHSCORES')
local retry = window
if oldest[2] then
  retry = (tonumber(oldest[2]) + window) - now
  if retry < 0 then retry = 0 end
end
return {0, 0, retry}
`)

var rateLimitSeq uint64

// AllowRate atomically records a hit against `key` and reports whether it is within
// `limit` requests per `window`, sharing state across instances via Redis. Returns
// the remaining allowance and, when denied, how long to wait. Satisfies the
// RateLimitStore interfaces used by the in-process limiters.
func (r *RedisClient) AllowRate(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	if limit <= 0 || window <= 0 {
		return true, 0, 0, nil
	}
	now := time.Now().UnixMilli()
	// Unique member so two hits in the same millisecond don't collide in the set.
	member := strconv.FormatInt(now, 10) + ":" + strconv.FormatUint(atomic.AddUint64(&rateLimitSeq, 1), 10)
	res, err := rateLimitScript.Run(ctx, r.client, []string{"rl:" + key},
		now, window.Milliseconds(), limit, member).Result()
	if err != nil {
		return false, 0, 0, err
	}
	vals, ok := res.([]interface{})
	if !ok || len(vals) != 3 {
		return false, 0, 0, fmt.Errorf("ratelimit: unexpected script result %v", res)
	}
	allowed := luaInt(vals[0]) == 1
	remaining := int(luaInt(vals[1]))
	retry := time.Duration(luaInt(vals[2])) * time.Millisecond
	return allowed, remaining, retry, nil
}

// luaInt coerces a Lua number (go-redis returns int64) to int64.
func luaInt(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

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

// SetNX atomically sets key=value with TTL only if the key does not already
// exist. It returns true when the key was set (i.e. it was absent) — the
// building block for single-use nonce / anti-replay storage.
func (r *RedisClient) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

// --- Refresh token storage (durable rotation + reuse detection) ---

// RefreshPut stores (or overwrites) a refresh-token record keyed by its hash.
func (r *RedisClient) RefreshPut(ctx context.Context, tokenHash string, record []byte, ttl time.Duration) error {
	return r.client.Set(ctx, "refresh:"+tokenHash, record, ttl).Err()
}

// RefreshGet returns a refresh-token record by hash; ok is false when absent.
func (r *RedisClient) RefreshGet(ctx context.Context, tokenHash string) (record []byte, ok bool, err error) {
	val, gerr := r.client.Get(ctx, "refresh:"+tokenHash).Bytes()
	if gerr == redis.Nil {
		return nil, false, nil
	}
	if gerr != nil {
		return nil, false, gerr
	}
	return val, true, nil
}

// RefreshAddToUser indexes a token hash under its user so all of a user's tokens
// can be enumerated for revocation. The index TTL tracks the token lifetime.
func (r *RedisClient) RefreshAddToUser(ctx context.Context, userID, tokenHash string, ttl time.Duration) error {
	key := "urt:" + userID
	if err := r.client.SAdd(ctx, key, tokenHash).Err(); err != nil {
		return err
	}
	return r.client.Expire(ctx, key, ttl).Err()
}

// RefreshUserHashes returns all token hashes indexed for a user.
func (r *RedisClient) RefreshUserHashes(ctx context.Context, userID string) ([]string, error) {
	return r.client.SMembers(ctx, "urt:"+userID).Result()
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

// --- DID Device Key Cache (WO-1 passkey auth) ---

const didKeyPrefix = "did:keys:"
const DIDKeyCacheTTL = 60 * time.Second

// GetDIDDeviceKeys returns the cached hex-encoded P-256 public keys for a DID,
// or nil if the entry is not cached.
func (r *RedisClient) GetDIDDeviceKeys(ctx context.Context, did string) ([]string, error) {
	val, err := r.client.LRange(ctx, didKeyPrefix+did, 0, -1).Result()
	if err == redis.Nil || len(val) == 0 {
		return nil, nil
	}
	return val, err
}

// SetDIDDeviceKeys caches the hex-encoded P-256 public keys for a DID with a 60s TTL.
func (r *RedisClient) SetDIDDeviceKeys(ctx context.Context, did string, keys []string) error {
	key := didKeyPrefix + did
	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	for _, k := range keys {
		pipe.RPush(ctx, key, k)
	}
	pipe.Expire(ctx, key, DIDKeyCacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteDIDDeviceKeys removes the DID key cache entry, forcing re-population on next auth.
func (r *RedisClient) DeleteDIDDeviceKeys(ctx context.Context, did string) error {
	return r.client.Del(ctx, didKeyPrefix+did).Err()
}

// --- Device Registration Tokens (WO-273) ---

const deviceRegTokenPrefix = "device_reg_token:"

// SetDeviceRegToken stores a single-use device registration token with a TTL.
func (r *RedisClient) SetDeviceRegToken(ctx context.Context, token string, record []byte, ttl time.Duration) error {
	return r.client.Set(ctx, deviceRegTokenPrefix+token, record, ttl).Err()
}

// GetDeviceRegToken retrieves the device registration token record.
// Returns redis.Nil-wrapped error if missing or expired.
func (r *RedisClient) GetDeviceRegToken(ctx context.Context, token string) ([]byte, error) {
	val, err := r.client.Get(ctx, deviceRegTokenPrefix+token).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("token not found or expired")
	}
	return val, err
}

// DeleteDeviceRegToken deletes a device registration token (single-use enforcement).
func (r *RedisClient) DeleteDeviceRegToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, deviceRegTokenPrefix+token).Err()
}

// --- SMS OTP Sessions (Wave 12) ---

const smsOTPPrefix = "sms_otp:"
const SMSOTPSessionTTL = 5 * time.Minute

// SetSMSOTPSession stores the OTP session (JSON blob) keyed by session token.
func (r *RedisClient) SetSMSOTPSession(ctx context.Context, sessionToken string, record []byte) error {
	return r.client.Set(ctx, smsOTPPrefix+sessionToken, record, SMSOTPSessionTTL).Err()
}

// GetSMSOTPSession retrieves an OTP session by token; returns error if missing/expired.
func (r *RedisClient) GetSMSOTPSession(ctx context.Context, sessionToken string) ([]byte, error) {
	val, err := r.client.Get(ctx, smsOTPPrefix+sessionToken).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("OTP session not found or expired")
	}
	return val, err
}

// DeleteSMSOTPSession removes an OTP session (single-use enforcement).
func (r *RedisClient) DeleteSMSOTPSession(ctx context.Context, sessionToken string) error {
	return r.client.Del(ctx, smsOTPPrefix+sessionToken).Err()
}
