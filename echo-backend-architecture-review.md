# Echo Backend Architecture Review & Recommendations

## Executive Summary

The current blueprint provides a solid foundation for the Echo backend architecture, establishing Go REST services as the intermediary layer between the iOS app and Constellation metagraph. However, several critical areas require additional detail for production readiness at scale.

**Overall Assessment**: 🟡 Good Foundation, Needs Enhancement

| Area | Current State | Recommendation Priority |
|------|---------------|------------------------|
| Service Architecture | Basic outline | 🔴 Critical |
| Database Strategy | Not specified | 🔴 Critical |
| Caching Layer | Not specified | 🔴 Critical |
| Rate Limiting | Basic concept | 🟡 High |
| Authentication | Outlined | 🟡 High |
| Monitoring/Observability | Basic logging | 🟡 High |
| Deployment/Infrastructure | Not specified | 🔴 Critical |
| Disaster Recovery | Not specified | 🔴 Critical |

---

## 1. Service Architecture Gaps & Recommendations

### Current Gap
The blueprint mentions "collection of Go REST services" but doesn't define:
- Service boundaries
- Inter-service communication patterns
- Service discovery mechanisms
- API gateway strategy

### Recommended Microservices Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API Gateway Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Kong /    │  │    Rate     │  │   Auth      │  │   Request   │        │
│  │   Traefik   │  │   Limiter   │  │   Filter    │  │   Router    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Core Services Layer                               │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Identity   │  │   Messaging  │  │    Trust     │  │   Rewards    │    │
│  │   Service    │  │   Service    │  │   Service    │  │   Service    │    │
│  │              │  │              │  │              │  │              │    │
│  │ • Auth       │  │ • Send/Recv  │  │ • Score Calc │  │ • Token Mgmt │    │
│  │ • Passkeys   │  │ • Groups     │  │ • Verify     │  │ • Staking    │    │
│  │ • DID Mgmt   │  │ • Encryption │  │ • Reports    │  │ • Referrals  │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Contacts   │  │  Metagraph   │  │ Notification │  │    Media     │    │
│  │   Service    │  │   Gateway    │  │   Service    │  │   Service    │    │
│  │              │  │              │  │              │  │              │    │
│  │ • CRUD       │  │ • L0 Queries │  │ • Push/APNS  │  │ • Upload     │    │
│  │ • Circles    │  │ • L1 Submit  │  │ • In-App     │  │ • Processing │    │
│  │ • Discovery  │  │ • Tx Monitor │  │ • Email      │  │ • CDN Sync   │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Data & Infrastructure Layer                         │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  PostgreSQL  │  │    Redis     │  │   Message    │  │  Metagraph   │    │
│  │   Cluster    │  │   Cluster    │  │    Queue     │  │    Nodes     │    │
│  │              │  │              │  │   (NATS)     │  │              │    │
│  │ • Primary    │  │ • Cache      │  │              │  │ • L0 (read)  │    │
│  │ • Replicas   │  │ • Sessions   │  │ • Async Jobs │  │ • L1 (write) │    │
│  │ • Sharding   │  │ • Pub/Sub    │  │ • Events     │  │ • Data L1    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Recommended Service Definitions

```go
// Service Registry - services.go
package services

type ServiceDefinition struct {
    Name           string
    Version        string
    Port           int
    HealthEndpoint string
    Dependencies   []string
    Replicas       ReplicaConfig
}

var ServiceRegistry = map[string]ServiceDefinition{
    "identity": {
        Name:           "identity-service",
        Version:        "v1",
        Port:           8001,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "redis", "cardano-api"},
        Replicas:       ReplicaConfig{Min: 3, Max: 20, CPUThreshold: 70},
    },
    "messaging": {
        Name:           "messaging-service",
        Version:        "v1",
        Port:           8002,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "redis", "metagraph", "kinnami"},
        Replicas:       ReplicaConfig{Min: 5, Max: 50, CPUThreshold: 60},
    },
    "trust": {
        Name:           "trust-service",
        Version:        "v1",
        Port:           8003,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "redis", "metagraph"},
        Replicas:       ReplicaConfig{Min: 3, Max: 15, CPUThreshold: 70},
    },
    "rewards": {
        Name:           "rewards-service",
        Version:        "v1",
        Port:           8004,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "redis", "metagraph"},
        Replicas:       ReplicaConfig{Min: 2, Max: 10, CPUThreshold: 70},
    },
    "contacts": {
        Name:           "contacts-service",
        Version:        "v1",
        Port:           8005,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "redis"},
        Replicas:       ReplicaConfig{Min: 3, Max: 20, CPUThreshold: 70},
    },
    "metagraph-gateway": {
        Name:           "metagraph-gateway",
        Version:        "v1",
        Port:           8006,
        HealthEndpoint: "/health",
        Dependencies:   []string{"metagraph-l0", "metagraph-l1", "redis"},
        Replicas:       ReplicaConfig{Min: 3, Max: 10, CPUThreshold: 60},
    },
    "notification": {
        Name:           "notification-service",
        Version:        "v1",
        Port:           8007,
        HealthEndpoint: "/health",
        Dependencies:   []string{"redis", "apns", "nats"},
        Replicas:       ReplicaConfig{Min: 2, Max: 15, CPUThreshold: 70},
    },
    "media": {
        Name:           "media-service",
        Version:        "v1",
        Port:           8008,
        HealthEndpoint: "/health",
        Dependencies:   []string{"postgres", "s3", "cdn"},
        Replicas:       ReplicaConfig{Min: 2, Max: 20, CPUThreshold: 60},
    },
}
```

---

## 2. Database Strategy (CRITICAL GAP)

### Current Gap
No database strategy is defined. This is critical for:
- User data persistence
- Message storage (even if encrypted)
- Trust score caching
- Session management
- Rate limit tracking

### Recommended Database Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Database Architecture                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              PostgreSQL Cluster (Primary)                │    │
│  │                                                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │   users      │  │   contacts   │  │   messages   │   │    │
│  │  │   Schema     │  │   Schema     │  │   Schema     │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘   │    │
│  │                                                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │   trust      │  │   rewards    │  │   audit      │   │    │
│  │  │   Schema     │  │   Schema     │  │   Schema     │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                   │
│                    ┌─────────┴─────────┐                        │
│                    ▼                   ▼                        │
│  ┌──────────────────────┐  ┌──────────────────────┐            │
│  │   Read Replica #1    │  │   Read Replica #2    │            │
│  │   (US-East)          │  │   (US-West)          │            │
│  └──────────────────────┘  └──────────────────────┘            │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              TimescaleDB (Time-Series)                   │    │
│  │                                                          │    │
│  │  • Trust score history    • Reward transactions          │    │
│  │  • API metrics            • Metagraph sync status        │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Recommended Schema Design

```sql
-- Core User Schema
CREATE SCHEMA users;

CREATE TABLE users.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    did VARCHAR(256) UNIQUE NOT NULL,
    phone_hash VARCHAR(64) UNIQUE, -- Hashed for privacy
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active',
    verification_level INTEGER DEFAULT 0,
    metagraph_address VARCHAR(128),
    
    -- Partitioning key for sharding
    shard_key INTEGER GENERATED ALWAYS AS (
        abs(hashtext(id::text)) % 16
    ) STORED
);

CREATE TABLE users.passkeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users.accounts(id) ON DELETE CASCADE,
    credential_id BYTEA UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    sign_count INTEGER DEFAULT 0,
    device_name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    is_primary BOOLEAN DEFAULT FALSE
);

CREATE TABLE users.profiles (
    user_id UUID PRIMARY KEY REFERENCES users.accounts(id) ON DELETE CASCADE,
    display_name VARCHAR(100),
    username VARCHAR(50) UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(500),
    status_message VARCHAR(200),
    visibility_settings JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trust Schema
CREATE SCHEMA trust;

CREATE TABLE trust.scores (
    user_id UUID PRIMARY KEY REFERENCES users.accounts(id),
    current_score INTEGER DEFAULT 0 CHECK (current_score >= 0 AND current_score <= 100),
    verification_points INTEGER DEFAULT 0,
    network_points INTEGER DEFAULT 0,
    behavior_points INTEGER DEFAULT 0,
    penalty_points INTEGER DEFAULT 0,
    multiplier DECIMAL(3,2) DEFAULT 1.0,
    level VARCHAR(20) DEFAULT 'newcomer',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    metagraph_hash VARCHAR(128) -- Last synced hash
);

CREATE TABLE trust.history (
    id BIGSERIAL,
    user_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    points_change INTEGER NOT NULL,
    reason VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metagraph_tx_hash VARCHAR(128),
    
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE trust.history_2025_01 PARTITION OF trust.history
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE trust.history_2025_02 PARTITION OF trust.history
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- Continue for each month...

-- Messages Schema (metadata only - content encrypted)
CREATE SCHEMA messages;

CREATE TABLE messages.conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL, -- 'direct', 'group'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_message_at TIMESTAMPTZ,
    metagraph_channel_id VARCHAR(128)
);

CREATE TABLE messages.participants (
    conversation_id UUID REFERENCES messages.conversations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users.accounts(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_read_at TIMESTAMPTZ,
    muted_until TIMESTAMPTZ,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE messages.message_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID REFERENCES messages.conversations(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users.accounts(id),
    message_type VARCHAR(20) NOT NULL, -- 'text', 'image', 'voice', 'file'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    metagraph_tx_hash VARCHAR(128),
    encrypted_content_ref VARCHAR(256), -- Reference to encrypted content
    
    -- Partitioning for scale
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

-- Rewards Schema
CREATE SCHEMA rewards;

CREATE TABLE rewards.balances (
    user_id UUID PRIMARY KEY REFERENCES users.accounts(id),
    available_balance DECIMAL(18,8) DEFAULT 0,
    staked_balance DECIMAL(18,8) DEFAULT 0,
    pending_rewards DECIMAL(18,8) DEFAULT 0,
    lifetime_earned DECIMAL(18,8) DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    metagraph_balance_hash VARCHAR(128)
);

CREATE TABLE rewards.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users.accounts(id),
    type VARCHAR(30) NOT NULL, -- 'earn', 'stake', 'unstake', 'transfer', 'referral'
    amount DECIMAL(18,8) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    metagraph_tx_hash VARCHAR(128),
    metadata JSONB
);

CREATE TABLE rewards.staking_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users.accounts(id),
    amount DECIMAL(18,8) NOT NULL,
    tier VARCHAR(20) NOT NULL, -- 'bronze', 'silver', 'gold', 'diamond'
    apy_rate DECIMAL(5,2) NOT NULL,
    lock_period_days INTEGER NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    unlocks_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) DEFAULT 'active'
);

-- Indexes for Performance
CREATE INDEX idx_users_phone_hash ON users.accounts(phone_hash);
CREATE INDEX idx_users_did ON users.accounts(did);
CREATE INDEX idx_trust_scores_level ON trust.scores(level);
CREATE INDEX idx_messages_conversation ON messages.message_metadata(conversation_id, created_at DESC);
CREATE INDEX idx_rewards_user_transactions ON rewards.transactions(user_id, created_at DESC);
```

---

## 3. Caching Layer (CRITICAL GAP)

### Current Gap
No caching strategy defined. This is essential for:
- Reducing database load
- Improving response times
- Session management
- Rate limit tracking

### Recommended Redis Architecture

```go
// redis/client.go
package redis

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type CacheConfig struct {
    // Cluster configuration
    Addresses     []string
    Password      string
    PoolSize      int
    MinIdleConns  int
    MaxRetries    int
    
    // TTL configurations
    SessionTTL          time.Duration
    UserProfileTTL      time.Duration
    TrustScoreTTL       time.Duration
    RateLimitWindowTTL  time.Duration
    MetagraphStateTTL   time.Duration
}

var DefaultConfig = CacheConfig{
    PoolSize:            100,
    MinIdleConns:        20,
    MaxRetries:          3,
    SessionTTL:          24 * time.Hour,
    UserProfileTTL:      15 * time.Minute,
    TrustScoreTTL:       5 * time.Minute,
    RateLimitWindowTTL:  1 * time.Minute,
    MetagraphStateTTL:   30 * time.Second,
}

// Cache key patterns
const (
    // Session keys
    KeySessionPrefix      = "session:"           // session:{session_id}
    KeyUserSessionsPrefix = "user:sessions:"     // user:sessions:{user_id}
    
    // User data keys
    KeyUserProfilePrefix  = "user:profile:"      // user:profile:{user_id}
    KeyUserTrustPrefix    = "user:trust:"        // user:trust:{user_id}
    KeyUserBalancePrefix  = "user:balance:"      // user:balance:{user_id}
    
    // Rate limiting keys
    KeyRateLimitPrefix    = "ratelimit:"         // ratelimit:{user_id}:{endpoint}
    KeyRateLimitIPPrefix  = "ratelimit:ip:"      // ratelimit:ip:{ip_address}
    
    // Metagraph state keys
    KeyMetagraphState     = "metagraph:state"
    KeyMetagraphTxPrefix  = "metagraph:tx:"      // metagraph:tx:{tx_hash}
    
    // Pub/Sub channels
    ChannelUserEvents     = "events:user"
    ChannelMessageEvents  = "events:messages"
    ChannelTrustEvents    = "events:trust"
)

type CacheClient struct {
    cluster *redis.ClusterClient
    config  CacheConfig
}

func NewCacheClient(config CacheConfig) (*CacheClient, error) {
    cluster := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:        config.Addresses,
        Password:     config.Password,
        PoolSize:     config.PoolSize,
        MinIdleConns: config.MinIdleConns,
        MaxRetries:   config.MaxRetries,
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := cluster.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis connection failed: %w", err)
    }
    
    return &CacheClient{cluster: cluster, config: config}, nil
}

// Session Management
func (c *CacheClient) SetSession(ctx context.Context, sessionID string, userID string, data map[string]interface{}) error {
    pipe := c.cluster.Pipeline()
    
    // Store session data
    sessionKey := KeySessionPrefix + sessionID
    pipe.HSet(ctx, sessionKey, data)
    pipe.Expire(ctx, sessionKey, c.config.SessionTTL)
    
    // Add to user's sessions set
    userSessionsKey := KeyUserSessionsPrefix + userID
    pipe.SAdd(ctx, userSessionsKey, sessionID)
    pipe.Expire(ctx, userSessionsKey, c.config.SessionTTL)
    
    _, err := pipe.Exec(ctx)
    return err
}

func (c *CacheClient) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
    return c.cluster.HGetAll(ctx, KeySessionPrefix+sessionID).Result()
}

func (c *CacheClient) InvalidateSession(ctx context.Context, sessionID string, userID string) error {
    pipe := c.cluster.Pipeline()
    pipe.Del(ctx, KeySessionPrefix+sessionID)
    pipe.SRem(ctx, KeyUserSessionsPrefix+userID, sessionID)
    _, err := pipe.Exec(ctx)
    return err
}

// User Profile Caching
func (c *CacheClient) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
    data, err := c.cluster.Get(ctx, KeyUserProfilePrefix+userID).Bytes()
    if err == redis.Nil {
        return nil, nil // Cache miss
    }
    if err != nil {
        return nil, err
    }
    
    var profile UserProfile
    if err := json.Unmarshal(data, &profile); err != nil {
        return nil, err
    }
    return &profile, nil
}

func (c *CacheClient) SetUserProfile(ctx context.Context, userID string, profile *UserProfile) error {
    data, err := json.Marshal(profile)
    if err != nil {
        return err
    }
    return c.cluster.Set(ctx, KeyUserProfilePrefix+userID, data, c.config.UserProfileTTL).Err()
}

// Trust Score Caching
func (c *CacheClient) GetTrustScore(ctx context.Context, userID string) (*TrustScore, error) {
    data, err := c.cluster.Get(ctx, KeyUserTrustPrefix+userID).Bytes()
    if err == redis.Nil {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    var score TrustScore
    if err := json.Unmarshal(data, &score); err != nil {
        return nil, err
    }
    return &score, nil
}

func (c *CacheClient) SetTrustScore(ctx context.Context, userID string, score *TrustScore) error {
    data, err := json.Marshal(score)
    if err != nil {
        return err
    }
    return c.cluster.Set(ctx, KeyUserTrustPrefix+userID, data, c.config.TrustScoreTTL).Err()
}

// Rate Limiting with Sliding Window
func (c *CacheClient) CheckRateLimit(ctx context.Context, userID string, endpoint string, limit int, window time.Duration) (bool, int, error) {
    key := fmt.Sprintf("%s%s:%s", KeyRateLimitPrefix, userID, endpoint)
    now := time.Now().UnixNano()
    windowStart := now - int64(window)
    
    pipe := c.cluster.Pipeline()
    
    // Remove old entries outside the window
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
    
    // Count current requests in window
    pipe.ZCard(ctx, key)
    
    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    
    // Set expiry
    pipe.Expire(ctx, key, window)
    
    results, err := pipe.Exec(ctx)
    if err != nil {
        return false, 0, err
    }
    
    currentCount := results[1].(*redis.IntCmd).Val()
    remaining := limit - int(currentCount)
    
    if currentCount >= int64(limit) {
        return false, 0, nil // Rate limited
    }
    
    return true, remaining, nil
}

// Pub/Sub for Real-time Events
func (c *CacheClient) PublishEvent(ctx context.Context, channel string, event interface{}) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    return c.cluster.Publish(ctx, channel, data).Err()
}

func (c *CacheClient) SubscribeEvents(ctx context.Context, channels ...string) *redis.PubSub {
    return c.cluster.Subscribe(ctx, channels...)
}
```

---

## 4. Enhanced Rate Limiting Strategy

### Current Gap
Basic rate limiting mentioned (100 req/min base, 2x-10x for VIP) but no implementation details for:
- Distributed rate limiting across service instances
- Different limits per endpoint type
- Burst handling
- Rate limit headers

### Recommended Implementation

```go
// ratelimit/limiter.go
package ratelimit

import (
    "context"
    "net/http"
    "strconv"
    "time"
)

type TierConfig struct {
    Tier           string
    RequestsPerMin int
    BurstSize      int
    CostMultiplier float64
}

var TierConfigs = map[string]TierConfig{
    "free": {
        Tier:           "free",
        RequestsPerMin: 100,
        BurstSize:      20,
        CostMultiplier: 1.0,
    },
    "basic": {
        Tier:           "basic", // $4.99/month
        RequestsPerMin: 200,
        BurstSize:      50,
        CostMultiplier: 0.8,
    },
    "premium": {
        Tier:           "premium", // $9.99/month
        RequestsPerMin: 500,
        BurstSize:      100,
        CostMultiplier: 0.5,
    },
    "enterprise": {
        Tier:           "enterprise",
        RequestsPerMin: 2000,
        BurstSize:      500,
        CostMultiplier: 0.3,
    },
}

// Endpoint-specific limits (multiplier on base rate)
var EndpointCosts = map[string]int{
    "GET /v1/messages":       1,
    "POST /v1/messages":      2,
    "GET /v1/contacts":       1,
    "POST /v1/contacts":      3,
    "GET /v1/trust/score":    1,
    "POST /v1/verify":        10, // Expensive operation
    "POST /v1/metagraph/tx":  5,  // Blockchain write
    "GET /v1/rewards":        1,
    "POST /v1/rewards/stake": 5,
}

type RateLimiter struct {
    cache  *CacheClient
    config map[string]TierConfig
}

func NewRateLimiter(cache *CacheClient) *RateLimiter {
    return &RateLimiter{
        cache:  cache,
        config: TierConfigs,
    }
}

func (r *RateLimiter) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            ctx := req.Context()
            
            // Get user from context (set by auth middleware)
            user := GetUserFromContext(ctx)
            if user == nil {
                // Fall back to IP-based limiting for unauthenticated requests
                r.handleIPRateLimit(w, req, next)
                return
            }
            
            // Get user's tier
            tierConfig := r.config[user.Tier]
            if tierConfig.Tier == "" {
                tierConfig = r.config["free"]
            }
            
            // Calculate cost for this endpoint
            endpointKey := req.Method + " " + req.URL.Path
            cost := EndpointCosts[endpointKey]
            if cost == 0 {
                cost = 1 // Default cost
            }
            
            // Apply tier multiplier
            effectiveCost := int(float64(cost) * tierConfig.CostMultiplier)
            if effectiveCost < 1 {
                effectiveCost = 1
            }
            
            // Check rate limit
            allowed, remaining, err := r.cache.CheckRateLimitWithCost(
                ctx,
                user.ID,
                tierConfig.RequestsPerMin,
                time.Minute,
                effectiveCost,
            )
            
            if err != nil {
                // On error, allow request but log
                log.Error("Rate limit check failed", "error", err, "user_id", user.ID)
                next.ServeHTTP(w, req)
                return
            }
            
            // Set rate limit headers
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(tierConfig.RequestsPerMin))
            w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
            w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
            
            if !allowed {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, req)
        })
    }
}

func (r *RateLimiter) handleIPRateLimit(w http.ResponseWriter, req *http.Request, next http.Handler) {
    ctx := req.Context()
    ip := getClientIP(req)
    
    // Stricter limits for unauthenticated requests
    allowed, remaining, _ := r.cache.CheckRateLimit(ctx, "ip:"+ip, "global", 30, time.Minute)
    
    w.Header().Set("X-RateLimit-Limit", "30")
    w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
    
    if !allowed {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    next.ServeHTTP(w, req)
}
```

---

## 5. Observability & Monitoring (Enhancement Needed)

### Current Gap
Blueprint mentions "centralized logging with batched, encrypted storage" but lacks:
- Structured logging standards
- Metrics collection
- Distributed tracing
- Alerting thresholds
- Dashboard specifications

### Recommended Observability Stack

```go
// observability/telemetry.go
package observability

import (
    "context"
    "time"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/trace"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus Metrics
var (
    // Request metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "echo_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"service", "method", "endpoint", "status"},
    )
    
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "echo_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"service", "method", "endpoint"},
    )
    
    // Metagraph metrics
    MetagraphTransactionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "echo_metagraph_transactions_total",
            Help: "Total metagraph transactions",
        },
        []string{"type", "status"},
    )
    
    MetagraphTransactionLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "echo_metagraph_transaction_latency_seconds",
            Help:    "Metagraph transaction confirmation latency",
            Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
        },
        []string{"type"},
    )
    
    // Business metrics
    ActiveUsersGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "echo_active_users",
            Help: "Number of active users in the last 24h",
        },
    )
    
    MessagesSentTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "echo_messages_sent_total",
            Help: "Total messages sent",
        },
    )
    
    TrustScoreDistribution = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "echo_trust_score_distribution",
            Help:    "Distribution of user trust scores",
            Buckets: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
        },
        []string{"level"},
    )
    
    TokensStakedTotal = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "echo_tokens_staked_total",
            Help: "Total ECHO tokens staked",
        },
    )
    
    // Error tracking
    ErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "echo_errors_total",
            Help: "Total errors by type",
        },
        []string{"service", "error_type", "severity"},
    )
    
    // Cache metrics
    CacheHitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "echo_cache_hits_total",
            Help: "Cache hit count",
        },
        []string{"cache_type"},
    )
    
    CacheMissesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "echo_cache_misses_total",
            Help: "Cache miss count",
        },
        []string{"cache_type"},
    )
)

// Structured Logger
type Logger struct {
    service  string
    tracer   trace.Tracer
    fields   map[string]interface{}
}

func NewLogger(service string) *Logger {
    return &Logger{
        service: service,
        tracer:  otel.Tracer(service),
        fields:  make(map[string]interface{}),
    }
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
    newLogger := &Logger{
        service: l.service,
        tracer:  l.tracer,
        fields:  make(map[string]interface{}),
    }
    
    // Copy existing fields
    for k, v := range l.fields {
        newLogger.fields[k] = v
    }
    
    // Add trace context
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        newLogger.fields["trace_id"] = span.SpanContext().TraceID().String()
        newLogger.fields["span_id"] = span.SpanContext().SpanID().String()
    }
    
    return newLogger
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
    newLogger := &Logger{
        service: l.service,
        tracer:  l.tracer,
        fields:  make(map[string]interface{}),
    }
    
    for k, v := range l.fields {
        newLogger.fields[k] = v
    }
    newLogger.fields[key] = value
    
    return newLogger
}

func (l *Logger) Info(msg string) {
    l.log("INFO", msg)
}

func (l *Logger) Error(msg string, err error) {
    l.fields["error"] = err.Error()
    l.log("ERROR", msg)
    
    ErrorsTotal.WithLabelValues(l.service, categorizeError(err), "error").Inc()
}

func (l *Logger) log(level, msg string) {
    entry := map[string]interface{}{
        "timestamp": time.Now().UTC().Format(time.RFC3339Nano),
        "level":     level,
        "service":   l.service,
        "message":   msg,
    }
    
    for k, v := range l.fields {
        entry[k] = v
    }
    
    // Output as JSON (to be collected by log aggregator)
    jsonBytes, _ := json.Marshal(entry)
    fmt.Println(string(jsonBytes))
}

// Alert Thresholds Configuration
type AlertConfig struct {
    Name        string
    Metric      string
    Threshold   float64
    Duration    time.Duration
    Severity    string
    Description string
}

var AlertConfigs = []AlertConfig{
    {
        Name:        "HighErrorRate",
        Metric:      "rate(echo_errors_total[5m])",
        Threshold:   0.01, // 1% error rate
        Duration:    5 * time.Minute,
        Severity:    "critical",
        Description: "Error rate exceeds 1% for 5 minutes",
    },
    {
        Name:        "HighLatency",
        Metric:      "histogram_quantile(0.95, rate(echo_http_request_duration_seconds_bucket[5m]))",
        Threshold:   2.0, // 2 seconds P95
        Duration:    5 * time.Minute,
        Severity:    "warning",
        Description: "P95 latency exceeds 2 seconds",
    },
    {
        Name:        "MetagraphSyncLag",
        Metric:      "echo_metagraph_sync_lag_seconds",
        Threshold:   300, // 5 minutes
        Duration:    10 * time.Minute,
        Severity:    "critical",
        Description: "Metagraph sync lag exceeds 5 minutes",
    },
    {
        Name:        "LowCacheHitRate",
        Metric:      "echo_cache_hits_total / (echo_cache_hits_total + echo_cache_misses_total)",
        Threshold:   0.8, // Below 80% hit rate
        Duration:    15 * time.Minute,
        Severity:    "warning",
        Description: "Cache hit rate below 80%",
    },
    {
        Name:        "HighMemoryUsage",
        Metric:      "process_resident_memory_bytes",
        Threshold:   0.85, // 85% of limit
        Duration:    10 * time.Minute,
        Severity:    "warning",
        Description: "Memory usage exceeds 85%",
    },
}
```

---

## 6. Deployment & Infrastructure (CRITICAL GAP)

### Current Gap
No deployment strategy defined. Need:
- Container orchestration
- Auto-scaling policies
- Multi-region deployment
- Blue-green deployment strategy

### Recommended Kubernetes Architecture

```yaml
# kubernetes/echo-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: echo-production
  labels:
    app.kubernetes.io/name: echo
    environment: production

---
# kubernetes/identity-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: identity-service
  namespace: echo-production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: identity-service
  template:
    metadata:
      labels:
        app: identity-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: echo-service
      containers:
      - name: identity-service
        image: echo/identity-service:v1.0.0
        ports:
        - containerPort: 8001
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: SERVICE_NAME
          value: "identity-service"
        - name: POSTGRES_HOST
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: host
        - name: REDIS_CLUSTER
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: cluster-addresses
        - name: KINNAMI_KEY
          valueFrom:
            secretKeyRef:
              name: kinnami-credentials
              key: encryption-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8001
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          runAsNonRoot: true
          readOnlyRootFilesystem: true
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: identity-service
              topologyKey: topology.kubernetes.io/zone

---
# kubernetes/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: identity-service-hpa
  namespace: echo-production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: identity-service
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60

---
# kubernetes/pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: identity-service-pdb
  namespace: echo-production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: identity-service

---
# kubernetes/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-api-ingress
  namespace: echo-production
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.echo.app
    secretName: echo-api-tls
  rules:
  - host: api.echo.app
    http:
      paths:
      - path: /v1/identity
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8001
      - path: /v1/messages
        pathType: Prefix
        backend:
          service:
            name: messaging-service
            port:
              number: 8002
      - path: /v1/trust
        pathType: Prefix
        backend:
          service:
            name: trust-service
            port:
              number: 8003
      - path: /v1/rewards
        pathType: Prefix
        backend:
          service:
            name: rewards-service
            port:
              number: 8004
      - path: /v1/contacts
        pathType: Prefix
        backend:
          service:
            name: contacts-service
            port:
              number: 8005
```

### Multi-Region Deployment Strategy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Global Traffic Management                            │
│                                                                              │
│                        ┌──────────────────────┐                             │
│                        │   CloudFlare / AWS   │                             │
│                        │   Global Accelerator │                             │
│                        └──────────┬───────────┘                             │
│                                   │                                          │
│                    ┌──────────────┼──────────────┐                          │
│                    ▼              ▼              ▼                          │
│  ┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────┐   │
│  │   US-East Region    │ │   US-West Region    │ │   EU Region         │   │
│  │   (Primary)         │ │   (Secondary)       │ │   (Secondary)       │   │
│  │                     │ │                     │ │                     │   │
│  │  ┌───────────────┐  │ │  ┌───────────────┐  │ │  ┌───────────────┐  │   │
│  │  │  K8s Cluster  │  │ │  │  K8s Cluster  │  │ │  │  K8s Cluster  │  │   │
│  │  │  (3 AZs)      │  │ │  │  (3 AZs)      │  │ │  │  (3 AZs)      │  │   │
│  │  └───────────────┘  │ │  └───────────────┘  │ │  └───────────────┘  │   │
│  │                     │ │                     │ │                     │   │
│  │  ┌───────────────┐  │ │  ┌───────────────┐  │ │  ┌───────────────┐  │   │
│  │  │  PostgreSQL   │  │ │  │  PostgreSQL   │  │ │  │  PostgreSQL   │  │   │
│  │  │  Primary      │──┼─┼──│  Read Replica │  │ │  │  Read Replica │  │   │
│  │  └───────────────┘  │ │  └───────────────┘  │ │  └───────────────┘  │   │
│  │                     │ │                     │ │                     │   │
│  │  ┌───────────────┐  │ │  ┌───────────────┐  │ │  ┌───────────────┐  │   │
│  │  │  Redis        │  │ │  │  Redis        │  │ │  │  Redis        │  │   │
│  │  │  Cluster      │  │ │  │  Cluster      │  │ │  │  Cluster      │  │   │
│  │  └───────────────┘  │ │  └───────────────┘  │ │  └───────────────┘  │   │
│  └─────────────────────┘ └─────────────────────┘ └─────────────────────┘   │
│                                                                              │
│                    Cross-Region Replication (Async)                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Disaster Recovery & Business Continuity (CRITICAL GAP)

### Current Gap
No disaster recovery strategy defined.

### Recommended DR Strategy

```yaml
# Disaster Recovery Configuration
disaster_recovery:
  # Recovery Point Objective (RPO)
  rpo_targets:
    user_data: 1_minute      # Max data loss acceptable
    messages: 5_minutes      # Messages can have slightly higher RPO
    trust_scores: 15_minutes # Cached/reconstructible from metagraph
    rewards: 0               # Zero data loss (metagraph is source of truth)
  
  # Recovery Time Objective (RTO)
  rto_targets:
    critical_services: 5_minutes   # Auth, core messaging
    standard_services: 15_minutes  # Contacts, profiles
    ancillary_services: 1_hour     # Analytics, reporting
  
  # Backup Strategy
  backups:
    database:
      type: continuous_wal_archiving
      retention: 30_days
      point_in_time_recovery: true
      cross_region_replication: true
      encryption: AES-256
    
    redis:
      type: rdb_snapshots
      frequency: 1_hour
      aof_persistence: true
      retention: 7_days
    
    secrets:
      type: vault_backup
      frequency: daily
      cross_region: true
  
  # Failover Configuration
  failover:
    automatic: true
    health_check_interval: 10_seconds
    failure_threshold: 3
    recovery_threshold: 2
    
    primary_region: us-east-1
    failover_regions:
      - us-west-2
      - eu-west-1
    
    dns_ttl: 60_seconds
    
  # Runbook Triggers
  runbooks:
    - name: database_failover
      trigger: primary_db_unreachable_3_minutes
      actions:
        - promote_replica
        - update_connection_strings
        - notify_oncall
        - validate_connectivity
    
    - name: region_failover
      trigger: region_unreachable_5_minutes
      actions:
        - activate_secondary_region
        - update_global_dns
        - scale_up_secondary
        - notify_stakeholders
    
    - name: metagraph_disconnect
      trigger: metagraph_unreachable_10_minutes
      actions:
        - switch_to_backup_nodes
        - enable_local_queue
        - alert_blockchain_team
```

### Backup Verification Script

```go
// dr/backup_verification.go
package dr

import (
    "context"
    "time"
)

type BackupVerification struct {
    db     *sql.DB
    s3     *s3.Client
    logger *Logger
}

func (bv *BackupVerification) RunDailyVerification(ctx context.Context) error {
    results := &VerificationReport{
        Timestamp: time.Now(),
        Tests:     make([]TestResult, 0),
    }
    
    // 1. Verify database backup exists and is recent
    dbBackup, err := bv.verifyDatabaseBackup(ctx)
    results.Tests = append(results.Tests, dbBackup)
    
    // 2. Verify backup can be restored (to test environment)
    restoreTest, err := bv.testBackupRestore(ctx)
    results.Tests = append(results.Tests, restoreTest)
    
    // 3. Verify data integrity after restore
    integrityTest, err := bv.verifyDataIntegrity(ctx)
    results.Tests = append(results.Tests, integrityTest)
    
    // 4. Verify cross-region replication lag
    replicationTest, err := bv.verifyReplicationLag(ctx)
    results.Tests = append(results.Tests, replicationTest)
    
    // 5. Verify metagraph state can be reconstructed
    metagraphTest, err := bv.verifyMetagraphReconstruction(ctx)
    results.Tests = append(results.Tests, metagraphTest)
    
    // Generate and store report
    if err := bv.storeReport(ctx, results); err != nil {
        return err
    }
    
    // Alert if any tests failed
    if results.HasFailures() {
        return bv.alertOnFailure(ctx, results)
    }
    
    return nil
}

func (bv *BackupVerification) verifyDatabaseBackup(ctx context.Context) (TestResult, error) {
    result := TestResult{
        Name:      "Database Backup Verification",
        StartTime: time.Now(),
    }
    
    // Check latest backup exists
    latestBackup, err := bv.s3.GetLatestBackup(ctx, "echo-db-backups")
    if err != nil {
        result.Status = "FAILED"
        result.Error = err.Error()
        return result, err
    }
    
    // Verify backup is recent (within RPO)
    if time.Since(latestBackup.Timestamp) > time.Minute {
        result.Status = "WARNING"
        result.Message = "Backup older than RPO target"
        return result, nil
    }
    
    // Verify backup size is reasonable
    if latestBackup.Size < 1000 { // Minimum expected size
        result.Status = "FAILED"
        result.Error = "Backup size suspiciously small"
        return result, fmt.Errorf("backup size too small")
    }
    
    result.Status = "PASSED"
    result.EndTime = time.Now()
    return result, nil
}
```

---

## 8. Security Enhancements

### Current Gap
Security mentioned but needs more detail on:
- Secret management
- API security
- Input validation
- Audit logging

### Recommended Security Implementation

```go
// security/middleware.go
package security

import (
    "context"
    "net/http"
    "strings"
)

// Security Headers Middleware
func SecurityHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Strict Transport Security
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
            
            // Content Security Policy
            w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
            
            // Prevent MIME sniffing
            w.Header().Set("X-Content-Type-Options", "nosniff")
            
            // Prevent clickjacking
            w.Header().Set("X-Frame-Options", "DENY")
            
            // XSS Protection
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            
            // Referrer Policy
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            
            // Remove server identification
            w.Header().Del("Server")
            w.Header().Del("X-Powered-By")
            
            next.ServeHTTP(w, r)
        })
    }
}

// Input Validation Middleware
func InputValidation() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Validate Content-Type for POST/PUT/PATCH
            if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
                contentType := r.Header.Get("Content-Type")
                if !strings.HasPrefix(contentType, "application/json") {
                    http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
                    return
                }
            }
            
            // Limit request body size (10MB max)
            r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
            
            // Validate required headers
            if r.Header.Get("X-Request-ID") == "" {
                r.Header.Set("X-Request-ID", generateRequestID())
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Audit Logging
type AuditLogger struct {
    storage AuditStorage
    kinnami *KinnamiClient
}

type AuditEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    UserID      string                 `json:"user_id,omitempty"`
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    ResourceID  string                 `json:"resource_id,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
    RequestID   string                 `json:"request_id"`
    Status      string                 `json:"status"` // success, failure, error
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Signature   string                 `json:"signature"` // For tamper detection
}

func (al *AuditLogger) Log(ctx context.Context, event AuditEvent) error {
    // Generate event ID
    event.ID = generateEventID()
    event.Timestamp = time.Now().UTC()
    
    // Sign the event for tamper detection
    eventBytes, _ := json.Marshal(event)
    event.Signature = al.kinnami.Sign(eventBytes)
    
    // Store locally first
    if err := al.storage.Store(ctx, event); err != nil {
        return err
    }
    
    // Batch for decentralized storage (IPFS/Storj)
    al.batchForDecentralizedStorage(event)
    
    return nil
}

// Actions requiring audit logging
const (
    AuditActionLogin            = "user.login"
    AuditActionLogout           = "user.logout"
    AuditActionPasskeyCreate    = "passkey.create"
    AuditActionPasskeyDelete    = "passkey.delete"
    AuditActionIdentityVerify   = "identity.verify"
    AuditActionTrustChange      = "trust.change"
    AuditActionMessageSend      = "message.send"
    AuditActionMessageDelete    = "message.delete"
    AuditActionContactAdd       = "contact.add"
    AuditActionContactBlock     = "contact.block"
    AuditActionRewardsClaim     = "rewards.claim"
    AuditActionRewardsStake     = "rewards.stake"
    AuditActionRewardsTransfer  = "rewards.transfer"
    AuditActionSettingsChange   = "settings.change"
    AuditActionExportData       = "data.export"
    AuditActionDeleteAccount    = "account.delete"
)
```

---

## 9. Scale Projections & Capacity Planning

### Recommended Capacity Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Scale Projections                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  User Growth Projections:                                                    │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Month   │  Users    │  DAU      │  Messages/day  │  API Calls/day    │ │
│  ├──────────┼───────────┼───────────┼────────────────┼───────────────────┤ │
│  │  M1      │  10K      │  2K       │  50K           │  500K             │ │
│  │  M3      │  50K      │  10K      │  300K          │  3M               │ │
│  │  M6      │  200K     │  40K      │  1.5M          │  15M              │ │
│  │  M12     │  1M       │  200K     │  8M            │  80M              │ │
│  │  M24     │  5M       │  1M       │  50M           │  500M             │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  Infrastructure Scaling:                                                     │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Scale      │  API Pods  │  DB Size  │  Redis    │  Est. Cost/mo     │ │
│  ├─────────────┼────────────┼───────────┼───────────┼───────────────────┤ │
│  │  10K users  │  6-10      │  50GB     │  8GB      │  $2,000           │ │
│  │  100K users │  15-30     │  250GB    │  32GB     │  $8,000           │ │
│  │  500K users │  40-80     │  1TB      │  128GB    │  $25,000          │ │
│  │  1M users   │  80-150    │  2TB      │  256GB    │  $50,000          │ │
│  │  5M users   │  200-400   │  10TB     │  1TB      │  $150,000         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  Performance Targets:                                                        │
│  • API Response Time (P50): < 50ms                                          │
│  • API Response Time (P95): < 200ms                                         │
│  • API Response Time (P99): < 500ms                                         │
│  • Message Delivery Time: < 100ms                                           │
│  • Metagraph Confirmation: < 60s                                            │
│  • Availability Target: 99.95%                                              │
│  • Error Rate Target: < 0.1%                                                │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 10. Summary of Recommendations

### Critical (Must Address Before Launch)

| # | Gap | Recommendation | Effort |
|---|-----|----------------|--------|
| 1 | No database strategy | Implement PostgreSQL with read replicas and sharding plan | 2-3 weeks |
| 2 | No caching layer | Deploy Redis cluster with defined caching patterns | 1-2 weeks |
| 3 | No deployment strategy | Implement Kubernetes with auto-scaling | 2-3 weeks |
| 4 | No disaster recovery | Define RPO/RTO, implement backups and failover | 2-3 weeks |

### High Priority (Address Within 30 Days)

| # | Gap | Recommendation | Effort |
|---|-----|----------------|--------|
| 5 | Basic rate limiting | Implement distributed rate limiting with tiers | 1 week |
| 6 | Minimal observability | Deploy full observability stack (metrics, traces, logs) | 1-2 weeks |
| 7 | Service architecture unclear | Define service boundaries and communication patterns | 1 week |
| 8 | Security details missing | Implement comprehensive security middleware | 1-2 weeks |

### Medium Priority (Address Within 60 Days)

| # | Gap | Recommendation | Effort |
|---|-----|----------------|--------|
| 9 | No capacity planning | Define scaling thresholds and projections | 3-5 days |
| 10 | No API versioning strategy | Implement versioned endpoints with deprecation policy | 1 week |
| 11 | No circuit breaker patterns | Implement resilience patterns for external services | 1 week |

---

## 11. Recommended Technology Stack Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Recommended Technology Stack                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Core Services:                                                              │
│  • Language: Go 1.21+                                                        │
│  • Framework: Chi router + custom middleware                                 │
│  • API Spec: OpenAPI 3.0                                                     │
│                                                                              │
│  Data Layer:                                                                 │
│  • Primary DB: PostgreSQL 15+ (with TimescaleDB extension)                  │
│  • Cache: Redis 7+ Cluster                                                   │
│  • Message Queue: NATS JetStream                                            │
│  • Search: Elasticsearch (for contact discovery)                            │
│                                                                              │
│  Infrastructure:                                                             │
│  • Container Orchestration: Kubernetes (EKS/GKE)                            │
│  • Service Mesh: Istio (optional, for advanced traffic management)          │
│  • API Gateway: Kong or AWS API Gateway                                      │
│  • CDN: CloudFlare                                                           │
│  • Secrets: HashiCorp Vault                                                  │
│                                                                              │
│  Observability:                                                              │
│  • Metrics: Prometheus + Grafana                                             │
│  • Tracing: Jaeger / OpenTelemetry                                          │
│  • Logging: Loki or Elasticsearch                                           │
│  • Alerting: PagerDuty / Opsgenie                                           │
│                                                                              │
│  Security:                                                                   │
│  • TLS: 1.3 (cert-manager for K8s)                                          │
│  • Encryption: Kinnami (as specified)                                        │
│  • WAF: CloudFlare or AWS WAF                                               │
│  • Secrets: Vault with K8s integration                                       │
│                                                                              │
│  External Services:                                                          │
│  • Identity Verification: Prove, Daon, 1Kosmos, Darwinium                   │
│  • Push Notifications: APNs (direct)                                         │
│  • Decentralized Storage: IPFS/Storj (for audit logs)                       │
│  • Blockchain: Constellation Metagraph, Cardano                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Next Steps

1. **Immediate**: Review and approve database schema design
2. **Week 1**: Set up PostgreSQL cluster and Redis cluster in staging
3. **Week 2**: Implement core service skeletons with health checks
4. **Week 3**: Deploy to Kubernetes with basic auto-scaling
5. **Week 4**: Implement observability stack and alerting
6. **Week 5**: Load testing and capacity validation
7. **Week 6**: Disaster recovery testing and documentation

This document should be reviewed with the engineering team and updated based on specific requirements and constraints discovered during implementation.

---

# PART 5: CONSOLIDATED RECOMMENDATIONS (Continued)

## 5.2 Technology Stack Summary (Continued)

```
│  EXTERNAL SERVICES                                                           │
│  ├── Identity Verification: Prove, Daon, 1Kosmos, Darwinium                 │
│  ├── Push Notifications: APNs                                                │
│  ├── Blockchain APIs: Infura, Alchemy, QuickNode (reads)                    │
│  └── CDN: CloudFlare                                                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 5.3 Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Metagraph latency affects UX | High | High | Optimistic updates, clear status indicators |
| Cardano unavailable | Medium | High | Cache credentials, graceful degradation |
| Data inconsistency across chains | Medium | High | Explicit consistency model, conflict resolution |
| Scale beyond initial capacity | Medium | Medium | Auto-scaling, capacity monitoring |
| Security breach | Low | Critical | Defense in depth, audit logging, encryption |
| Key management failure | Low | Critical | HSM/Secure Enclave, backup procedures |

## 5.4 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Response Time (P95) | < 200ms | Prometheus histogram |
| Message Delivery Time | < 2 seconds | End-to-end tracing |
| Metagraph Confirmation | < 60 seconds | Transaction monitoring |
| App Crash Rate | < 0.1% | Crashlytics |
| Availability | 99.95% | Uptime monitoring |
| Error Rate | < 0.1% | Error tracking |
| Offline Sync Success | > 99% | Sync completion rate |

## 5.5 Next Steps

### Week 1-2: Foundation
- [ ] Finalize database schema
- [ ] Set up PostgreSQL cluster (staging)
- [ ] Set up Redis cluster (staging)
- [ ] Implement core service skeletons

### Week 3-4: Core Implementation
- [ ] Implement optimistic sync pattern
- [ ] Deploy to Kubernetes (staging)
- [ ] Implement health checks
- [ ] Set up observability stack

### Week 5-6: Integration & Testing
- [ ] End-to-end integration testing
- [ ] Load testing (target: 10K concurrent users)
- [ ] Security audit
- [ ] Disaster recovery testing

### Week 7-8: Hardening
- [ ] Performance optimization
- [ ] Documentation completion
- [ ] Runbook creation
- [ ] Production deployment preparation

---

## Document Information

| Field | Value |
|-------|-------|
| Version | 1.0 |
| Date | February 2025 |
| Status | Review Required |
| Authors | Architecture Team |
| Reviewers | Engineering, Security, Product |

---

*This document should be reviewed and updated as implementation progresses and requirements evolve.*
