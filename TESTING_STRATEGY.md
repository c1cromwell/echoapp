# ECHO Platform — Testing Strategy & Implementation Guide

## Overview

This document provides comprehensive testing strategy for the ECHO platform, aligned with DATA_LAYER_ARCHITECTURE_v3 and OpenAPI specification.

**Coverage Goals:**
- Unit tests: 90%+ coverage for all services
- Integration tests: 80%+ coverage for chain interactions
- E2E tests: 70%+ coverage for critical user flows
- Stress tests: Key resilience scenarios

**Timeline:** Phase 1–2 (parallel with backend implementation)

---

## 1. Test Environment Setup

### 1.1 Local Development Environment

**Docker Compose Services:**
```yaml
version: '3.9'
services:
  # Backend
  echo-backend:
    image: echo-backend:dev
    ports:
      - "8080:8080"
    environment:
      - METAGRAPH_ENDPOINT=http://metagraph-validator:9001
      - CARDANO_ENDPOINT=http://cardano-preview:8000
      - REDIS_URL=redis://redis:6379
      - DATABASE_URL=postgres://postgres:postgres@postgres:5432/echo

  # Blockchain simulators
  metagraph-validator:
    image: metagraph-simulator:latest
    ports:
      - "9001:9001"  # gRPC
    environment:
      - NETWORK_TYPE=local

  cardano-preview:
    image: cardano-node:8.0
    ports:
      - "8000:8000"
      - "8001:8001"
    volumes:
      - ./testnet-config:/config

  # Cache & queue
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # Database
  postgres:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=echo_test
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - ./migrations:/docker-entrypoint-initdb.d

  # IPFS (optional for local)
  ipfs:
    image: ipfs/go-ipfs:latest
    ports:
      - "5001:5001"
      - "8080:8080"

  # APNs mock
  apns-mock:
    image: apns-simulator:latest
    ports:
      - "8443:8443"

  # Test runner
  test-runner:
    image: echo-backend:dev
    command: /bin/bash
    stdin_open: true
    tty: true
    depends_on:
      - echo-backend
      - metagraph-validator
      - cardano-preview
      - redis
      - postgres
```

### 1.2 Test Database Setup

**PostgreSQL Migration (Phase 1):**
```sql
-- migrations/001_initial_schema.sql

CREATE TABLE users (
    id UUID PRIMARY KEY,
    did TEXT UNIQUE NOT NULL,
    public_key BYTEA NOT NULL,
    email TEXT UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE message_queue (
    id UUID PRIMARY KEY,
    recipient_did TEXT NOT NULL,
    sender_did TEXT NOT NULL,
    ciphertext BYTEA NOT NULL,  -- E2E encrypted
    signature BYTEA NOT NULL,
    commitment TEXT NOT NULL,   -- SHA256 hash
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'queued'
);

CREATE INDEX idx_message_queue_recipient ON message_queue(recipient_did, created_at DESC);
CREATE INDEX idx_message_queue_expires ON message_queue(expires_at);

CREATE TABLE reward_claims (
    id UUID PRIMARY KEY,
    user_did TEXT NOT NULL,
    claim_type VARCHAR(50) NOT NULL,  -- "message", "referral", "verification"
    amount DECIMAL(10, 2) NOT NULL,
    trust_tier INT NOT NULL,
    multiplier DECIMAL(3, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',  -- pending, submitted, confirmed
    metagraph_block_height INT,
    created_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP
);

CREATE INDEX idx_reward_claims_user_date ON reward_claims(user_did, created_at DESC);
CREATE INDEX idx_reward_claims_status ON reward_claims(status);

CREATE TABLE offline_queue_stats (
    recipient_did TEXT PRIMARY KEY,
    queue_depth INT NOT NULL DEFAULT 0,
    oldest_message_age_ms BIGINT,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE relay_metrics (
    id BIGSERIAL PRIMARY KEY,
    time_bucket TIMESTAMP NOT NULL,
    message_count INT NOT NULL,
    avg_latency_ms FLOAT NOT NULL,
    max_latency_ms FLOAT NOT NULL,
    rate_limit_hits INT NOT NULL,
    circuit_breaker_state VARCHAR(20),
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_relay_metrics_time ON relay_metrics(time_bucket DESC);
```

### 1.3 Test Data Fixtures

**Fixture Files:**
```go
// testdata/fixtures.go

package testdata

import (
    "crypto/rand"
    "encoding/hex"
    "testing"
    "time"
)

// UserFixture creates a test user
func UserFixture(t *testing.T, overrides ...map[string]interface{}) *User {
    u := &User{
        ID:        uuid.New().String(),
        DID:       randomDID(),
        PublicKey: randomBytes(32),
        Email:     "test+" + randomString(8) + "@example.com",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if len(overrides) > 0 {
        // Apply overrides
    }
    
    return u
}

// MessageFixture creates a test message
func MessageFixture(t *testing.T, senderDID, recipientDID string) *Message {
    return &Message{
        ID:             uuid.New().String(),
        ConversationID: uuid.New().String(),
        SenderDID:      senderDID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
        Timestamp:      time.Now(),
        Status:         "sent",
        CreatedAt:      time.Now(),
    }
}

// RewardClaimFixture creates a test reward claim
func RewardClaimFixture(t *testing.T, userDID string) *RewardClaim {
    return &RewardClaim{
        ID:        uuid.New().String(),
        UserDID:   userDID,
        ClaimType: "message",
        Amount:    0.01,
        TrustTier: 1,
        Multiplier: 1.0,
        Status:    "pending",
        CreatedAt: time.Now(),
    }
}

// randomDID generates a test DID
func randomDID() string {
    return "did:prism:" + randomString(56)
}

// randomHash generates a random SHA256 hash
func randomHash() string {
    return hex.EncodeToString(randomBytes(32))
}

func randomBytes(n int) []byte {
    b := make([]byte, n)
    rand.Read(b)
    return b
}

func randomString(n int) string {
    return hex.EncodeToString(randomBytes(n / 2))
}
```

---

## 2. Unit Test Examples

### 2.1 AuthService Tests

**File: `internal/service/auth_service_test.go`**

```go
package service

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAuthService_Register(t *testing.T) {
    service := setupAuthService(t)
    
    req := &RegisterRequest{
        PublicKey: hex.EncodeToString(randomBytes(32)),
        DisplayName: "Test User",
        DeviceInfo: &DeviceInfo{
            DeviceID: uuid.New().String(),
            DeviceName: "iPhone 15",
            OSVersion: "17.2",
            AppVersion: "1.0.0",
        },
    }
    
    resp, err := service.Register(context.Background(), req)
    
    require.NoError(t, err)
    assert.NotEmpty(t, resp.DID)
    assert.True(t, strings.HasPrefix(resp.DID, "did:prism:"))
    assert.NotEmpty(t, resp.ServiceEndpoints)
    
    // Verify user persisted
    user, err := service.db.GetUserByDID(resp.DID)
    require.NoError(t, err)
    assert.Equal(t, req.PublicKey, hex.EncodeToString(user.PublicKey))
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
    service := setupAuthService(t)
    
    // Register first user
    req1 := &RegisterRequest{
        PublicKey: hex.EncodeToString(randomBytes(32)),
        DisplayName: "User 1",
    }
    resp1, err := service.Register(context.Background(), req1)
    require.NoError(t, err)
    
    // Register second user with same email (should succeed, emails optional)
    req2 := &RegisterRequest{
        PublicKey: hex.EncodeToString(randomBytes(32)),
        DisplayName: "User 2",
    }
    resp2, err := service.Register(context.Background(), req2)
    require.NoError(t, err)
    assert.NotEqual(t, resp1.DID, resp2.DID)
}

func TestAuthService_RequestChallenge(t *testing.T) {
    service := setupAuthService(t)
    
    // Register user first
    user := createTestUser(t, service)
    
    // Request challenge
    chalResp, err := service.RequestChallenge(context.Background(), user.DID)
    require.NoError(t, err)
    assert.NotEmpty(t, chalResp.Challenge)
    assert.True(t, chalResp.ExpiresAt.After(time.Now()))
    
    // Challenge should expire in ~5 minutes
    assert.Less(t, time.Until(chalResp.ExpiresAt), 6*time.Minute)
    assert.Greater(t, time.Until(chalResp.ExpiresAt), 4*time.Minute)
}

func TestAuthService_RequestChallenge_UnregisteredDID(t *testing.T) {
    service := setupAuthService(t)
    
    _, err := service.RequestChallenge(context.Background(), "did:prism:fake")
    require.Error(t, err)
    assert.Equal(t, ErrUserNotFound, err)
}

func TestAuthService_VerifyChallenge(t *testing.T) {
    service := setupAuthService(t)
    
    // Setup: user with registered keypair
    user := createTestUser(t, service)
    
    // Request challenge
    chalResp, err := service.RequestChallenge(context.Background(), user.DID)
    require.NoError(t, err)
    
    // Sign challenge with private key
    signature := signWithTestKey(t, chalResp.Challenge)
    
    // Verify challenge
    authResp, err := service.VerifyChallenge(context.Background(), &VerifyChallengeRequest{
        DID:       user.DID,
        Challenge: chalResp.Challenge,
        Signature: signature,
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, authResp.AccessToken)
    assert.NotEmpty(t, authResp.RefreshToken)
    assert.NotEmpty(t, authResp.User.DID)
}

func TestAuthService_VerifyChallenge_InvalidSignature(t *testing.T) {
    service := setupAuthService(t)
    
    user := createTestUser(t, service)
    chalResp, _ := service.RequestChallenge(context.Background(), user.DID)
    
    // Sign with wrong key
    wrongKey := randomBytes(32)
    invalidSig := signWithKey(t, chalResp.Challenge, wrongKey)
    
    _, err := service.VerifyChallenge(context.Background(), &VerifyChallengeRequest{
        DID:       user.DID,
        Challenge: chalResp.Challenge,
        Signature: invalidSig,
    })
    
    require.Error(t, err)
    assert.Equal(t, ErrInvalidSignature, err)
}

func TestAuthService_VerifyChallenge_ExpiredChallenge(t *testing.T) {
    service := setupAuthService(t)
    
    user := createTestUser(t, service)
    chalResp, _ := service.RequestChallenge(context.Background(), user.DID)
    
    // Simulate expired challenge
    mockTime := func() time.Time { return chalResp.ExpiresAt.Add(1 * time.Minute) }
    service.clock = mockTime
    
    signature := signWithTestKey(t, chalResp.Challenge)
    
    _, err := service.VerifyChallenge(context.Background(), &VerifyChallengeRequest{
        DID:       user.DID,
        Challenge: chalResp.Challenge,
        Signature: signature,
    })
    
    require.Error(t, err)
    assert.Equal(t, ErrChallengeExpired, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
    service := setupAuthService(t)
    
    user := createTestUser(t, service)
    
    // Get initial tokens
    chalResp, _ := service.RequestChallenge(context.Background(), user.DID)
    authResp, _ := service.VerifyChallenge(context.Background(), &VerifyChallengeRequest{
        DID:       user.DID,
        Challenge: chalResp.Challenge,
        Signature: signWithTestKey(t, chalResp.Challenge),
    })
    
    // Refresh with refresh token
    newAuthResp, err := service.RefreshToken(context.Background(), &RefreshTokenRequest{
        RefreshToken: authResp.RefreshToken,
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, newAuthResp.AccessToken)
    assert.NotEqual(t, authResp.AccessToken, newAuthResp.AccessToken)
}

func TestAuthService_RefreshToken_ExpiredToken(t *testing.T) {
    service := setupAuthService(t)
    
    _, err := service.RefreshToken(context.Background(), &RefreshTokenRequest{
        RefreshToken: "eyJhbGc...invalid",
    })
    
    require.Error(t, err)
    assert.Equal(t, ErrInvalidToken, err)
}

func TestAuthService_Logout(t *testing.T) {
    service := setupAuthService(t)
    
    user := createTestUser(t, service)
    authResp := getTestAuthToken(t, service, user)
    
    // Logout
    err := service.Logout(context.Background(), authResp.AccessToken)
    require.NoError(t, err)
    
    // Subsequent request with token should fail
    _, err = service.ValidateToken(context.Background(), authResp.AccessToken)
    require.Error(t, err)
    assert.Equal(t, ErrTokenRevoked, err)
}

// Helper functions
func setupAuthService(t *testing.T) *AuthService {
    db := setupTestDB(t)
    cardano := setupTestCardanoClient(t)
    cache := setupTestRedis(t)
    
    return NewAuthService(db, cardano, cache)
}

func createTestUser(t *testing.T, service *AuthService) *User {
    resp, err := service.Register(context.Background(), &RegisterRequest{
        PublicKey: hex.EncodeToString(randomBytes(32)),
        DisplayName: "Test User",
    })
    require.NoError(t, err)
    
    user, err := service.db.GetUserByDID(resp.DID)
    require.NoError(t, err)
    return user
}
```

### 2.2 MessageService Tests

**File: `internal/service/message_service_test.go`**

```go
package service

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMessageService_SendMessage(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
        ContentType:    "text",
    }
    
    resp, err := service.SendMessage(context.Background(), msg)
    
    require.NoError(t, err)
    assert.NotEmpty(t, resp.MessageID)
    assert.Equal(t, "relayed", resp.Status)
    
    // Verify message persisted
    stored, err := service.db.GetMessage(resp.MessageID)
    require.NoError(t, err)
    assert.Equal(t, msg.SenderDID, stored.SenderDID)
}

func TestMessageService_SendMessage_RateLimited(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    // Send 60 messages (rate limit for Tier 1)
    for i := 0; i < 60; i++ {
        msg := &SendMessageRequest{
            ConversationID: uuid.New().String(),
            SenderDID:      sender.DID,
            RecipientDID:   recipient.DID,
            Ciphertext:     randomBytes(256),
            Signature:      randomBytes(64),
            Commitment:     randomHash(),
        }
        _, err := service.SendMessage(context.Background(), msg)
        require.NoError(t, err)
    }
    
    // 61st message should be rate limited
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    _, err := service.SendMessage(context.Background(), msg)
    require.Error(t, err)
    assert.Equal(t, ErrRateLimited, err)
}

func TestMessageService_QueueOfflineMessage(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    
    // Simulate offline recipient
    resp, err := service.SendMessage(context.Background(), msg)
    require.NoError(t, err)
    
    // Verify queued
    queued, err := service.cache.GetOfflineMessageCount(recipient.DID)
    require.NoError(t, err)
    assert.Greater(t, queued, 0)
    
    // Verify APNs notification would be sent
    // (mock APNs client)
    notif := service.apnsQueue.Pop()
    require.NotNil(t, notif)
    assert.Equal(t, recipient.DID, notif.RecipientDID)
}

func TestMessageService_QueueOfflineMessage_Overflow(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    // Fill queue to 1000 messages
    for i := 0; i < 1000; i++ {
        msg := &SendMessageRequest{
            ConversationID: uuid.New().String(),
            SenderDID:      sender.DID,
            RecipientDID:   recipient.DID,
            Ciphertext:     randomBytes(256),
            Signature:      randomBytes(64),
            Commitment:     randomHash(),
        }
        service.QueueOfflineMessage(context.Background(), msg)
    }
    
    // 1001st message should evict oldest
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    service.QueueOfflineMessage(context.Background(), msg)
    
    // Verify queue still 1000
    count, err := service.cache.GetOfflineMessageCount(recipient.DID)
    require.NoError(t, err)
    assert.Equal(t, 1000, count)
}

func TestMessageService_GetMessages_Pagination(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    convID := uuid.New().String()
    
    // Create 150 messages
    for i := 0; i < 150; i++ {
        msg := &SendMessageRequest{
            ConversationID: convID,
            SenderDID:      sender.DID,
            RecipientDID:   recipient.DID,
            Ciphertext:     randomBytes(256),
            Signature:      randomBytes(64),
            Commitment:     randomHash(),
        }
        service.SendMessage(context.Background(), msg)
    }
    
    // Fetch first page
    page1, err := service.GetMessages(context.Background(), &GetMessagesRequest{
        ConversationID: convID,
        Limit:          50,
    })
    require.NoError(t, err)
    assert.Len(t, page1.Messages, 50)
    assert.NotEmpty(t, page1.NextCursor)
    
    // Fetch second page with cursor
    page2, err := service.GetMessages(context.Background(), &GetMessagesRequest{
        ConversationID: convID,
        Cursor:         page1.NextCursor,
        Limit:          50,
    })
    require.NoError(t, err)
    assert.Len(t, page2.Messages, 50)
    
    // Verify no overlap
    msg1IDs := map[string]bool{}
    for _, m := range page1.Messages {
        msg1IDs[m.ID] = true
    }
    for _, m := range page2.Messages {
        assert.False(t, msg1IDs[m.ID])
    }
}

func TestMessageService_MarkAsRead(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    
    sendResp, _ := service.SendMessage(context.Background(), msg)
    
    // Mark as read
    err := service.MarkAsRead(context.Background(), sendResp.MessageID)
    require.NoError(t, err)
    
    // Verify status updated
    stored, _ := service.db.GetMessage(sendResp.MessageID)
    assert.Equal(t, "read", stored.Status)
}

func TestMessageService_AddReaction(t *testing.T) {
    service := setupMessageService(t)
    sender := createTestUser(t, service)
    recipient := createTestUser(t, service)
    
    // Send message
    msg := &SendMessageRequest{
        ConversationID: uuid.New().String(),
        SenderDID:      sender.DID,
        RecipientDID:   recipient.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    sendResp, _ := service.SendMessage(context.Background(), msg)
    
    // Add reaction
    err := service.AddReaction(context.Background(), &AddReactionRequest{
        MessageID: sendResp.MessageID,
        UserDID:   recipient.DID,
        Emoji:     "👍",
    })
    require.NoError(t, err)
    
    // Verify reaction added
    stored, _ := service.db.GetMessage(sendResp.MessageID)
    assert.Len(t, stored.Reactions, 1)
    assert.Equal(t, "👍", stored.Reactions[0].Emoji)
}

// Setup helpers
func setupMessageService(t *testing.T) *MessageService {
    db := setupTestDB(t)
    cache := setupTestRedis(t)
    nats := setupTestNATS(t)
    cardano := setupTestCardanoClient(t)
    apns := setupMockAPNs(t)
    
    return NewMessageService(db, cache, nats, cardano, apns)
}
```

---

## 3. Integration Tests

### 3.1 Metagraph Integration

**File: `test/integration/metagraph_test.go`**

```go
package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestMessageAnchoringToMetagraph(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    // Setup
    backend := setupBackend(t)
    metagraph := setupMetagraphSimulator(t)
    defer backend.Stop()
    defer metagraph.Stop()
    
    sender := createTestUserWithBackend(t, backend)
    recipient := createTestUserWithBackend(t, backend)
    
    // Send messages
    commitments := []string{}
    for i := 0; i < 1000; i++ {
        msg := &SendMessageRequest{
            ConversationID: uuid.New().String(),
            SenderDID:      sender.DID,
            RecipientDID:   recipient.DID,
            Ciphertext:     randomBytes(256),
            Signature:      randomBytes(64),
            Commitment:     randomHash(),
        }
        resp, err := backend.MessageService.SendMessage(context.Background(), msg)
        require.NoError(t, err)
        
        // Extract commitment from backend's batch
        stored, _ := backend.DB.GetMessage(resp.MessageID)
        commitments = append(commitments, stored.Commitment)
    }
    
    // Wait for batch to be submitted (default: 5 minutes or 1000 commits)
    // Trigger batch early for testing
    err := backend.MessageRelayService.FlushBatch(context.Background())
    require.NoError(t, err)
    
    // Verify Merkle root submitted to metagraph
    merkleRoot, err := metagraph.GetLatestDataL1Block()
    require.NoError(t, err)
    assert.NotEmpty(t, merkleRoot)
    
    // Verify tree structure
    tree := backend.MerkleTree.GetLatestTree()
    require.NotNil(t, tree)
    assert.Equal(t, len(commitments), tree.LeafCount())
}

func TestCardanoCredentialIssuance(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    backend := setupBackend(t)
    cardano := setupCardanoSimulator(t)
    defer backend.Stop()
    defer cardano.Stop()
    
    user := createTestUserWithBackend(t, backend)
    
    // Simulate IDV callback
    idvCallback := &IDVCallback{
        Status:           "approved",
        ConfidenceScore:  0.98,
        DocumentType:    "passport",
        AgeOver18:       true,
        ReferenceID:     uuid.New().String(),
    }
    
    err := backend.IdentityService.ProcessIDVCallback(context.Background(), idvCallback, user.DID)
    require.NoError(t, err)
    
    // Verify credential created on Cardano
    cred, err := cardano.GetCredential(user.DID)
    require.NoError(t, err)
    assert.NotNil(t, cred)
    assert.Equal(t, "approved", cred.Status)
    
    // Verify trust tier updated
    tierDatum, err := cardano.GetTrustTierDatum(user.DID)
    require.NoError(t, err)
    assert.Equal(t, 3, tierDatum.Tier)
}
```

### 3.2 End-to-End Tests

**File: `test/e2e/messaging_test.go`**

```go
package e2e

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestE2E_MessageSendReceiveOfflineReconnect(t *testing.T) {
    // Setup
    backend := setupBackend(t)
    client1 := setupIOSClient(t, backend)  // Sender
    client2 := setupIOSClient(t, backend)  // Recipient
    defer backend.Stop()
    defer client1.Stop()
    defer client2.Stop()
    
    // Register users
    user1 := client1.Register(t)
    user2 := client2.Register(t)
    
    // Authenticate
    token1 := client1.AuthenticateWithChallenge(t, user1.DID)
    token2 := client2.AuthenticateWithChallenge(t, user2.DID)
    
    // Create conversation
    convID := uuid.New().String()
    client1.CreateConversation(t, convID, user2.DID)
    client2.CreateConversation(t, convID, user1.DID)
    
    // Connect WebSocket
    ws1 := client1.ConnectWebSocket(t, token1)
    ws2 := client2.ConnectWebSocket(t, token2)
    
    // Send message (recipient online)
    msgID := uuid.New().String()
    msg := &Message{
        ID:             msgID,
        ConversationID: convID,
        SenderDID:      user1.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    
    client1.SendMessage(t, ws1, msg)
    
    // Recipient should receive immediately
    received := client2.WaitForMessage(t, ws2, 2*time.Second)
    require.NotNil(t, received)
    assert.Equal(t, msgID, received.ID)
    
    // Disconnect recipient
    client2.DisconnectWebSocket(t, ws2)
    
    // Send another message (recipient offline)
    msgID2 := uuid.New().String()
    msg2 := &Message{
        ID:             msgID2,
        ConversationID: convID,
        SenderDID:      user1.DID,
        Ciphertext:     randomBytes(256),
        Signature:      randomBytes(64),
        Commitment:     randomHash(),
    }
    client1.SendMessage(t, ws1, msg2)
    
    // Verify offline queue
    queueDepth := backend.GetOfflineQueueDepth(user2.DID)
    assert.Equal(t, 1, queueDepth)
    
    // Verify APNs push would be sent
    apnsCalls := backend.MockAPNs.GetCallHistory()
    assert.Greater(t, len(apnsCalls), 0)
    
    // Recipient reconnects
    ws2 = client2.ConnectWebSocket(t, token2)
    defer ws2.Close()
    
    // Should receive queued message
    queued := client2.WaitForMessage(t, ws2, 2*time.Second)
    require.NotNil(t, queued)
    assert.Equal(t, msgID2, queued.ID)
    
    // Queue should be empty
    queueDepth = backend.GetOfflineQueueDepth(user2.DID)
    assert.Equal(t, 0, queueDepth)
}

func TestE2E_MessageBatchAnchoring(t *testing.T) {
    backend := setupBackend(t)
    defer backend.Stop()
    
    client := setupIOSClient(t, backend)
    defer client.Stop()
    
    user := client.Register(t)
    token := client.AuthenticateWithChallenge(t, user.DID)
    
    // Send 1000 messages
    for i := 0; i < 1000; i++ {
        msg := &Message{
            ID:             uuid.New().String(),
            ConversationID: uuid.New().String(),
            SenderDID:      user.DID,
            Ciphertext:     randomBytes(256),
            Signature:      randomBytes(64),
            Commitment:     randomHash(),
        }
        client.SendMessage(t, token, msg)
    }
    
    // Trigger batch flush
    backend.MessageRelayService.FlushBatch(context.Background())
    
    // Wait for metagraph finality
    time.Sleep(15 * time.Second)
    
    // Verify Merkle root on metagraph
    root, err := backend.Metagraph.GetLatestDataL1Root()
    require.NoError(t, err)
    assert.NotEmpty(t, root)
}
```

---

## 4. Stress Test Examples

**File: `test/stress/throughput_test.go`**

```go
package stress

import (
    "context"
    "fmt"
    "sync"
    "testing"
    "time"

    "github.com/prometheus/client_golang/prometheus"
)

func StressTest_MessageRelay_100K_Online_50MPS(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping stress test")
    }
    
    // Setup
    backend := setupBackend(t)
    defer backend.Stop()
    
    // Create 100K users (1K at a time)
    users := make([]*User, 100000)
    for i := 0; i < 100; i++ {
        for j := 0; j < 1000; j++ {
            users[i*1000+j] = createTestUserWithBackend(t, backend)
        }
        fmt.Printf("Created %d users\n", (i+1)*1000)
    }
    
    // Connect 100K to WebSocket
    wsConns := make([]*WebSocketConn, 100000)
    for i, user := range users {
        token := authenticateUser(t, backend, user)
        wsConns[i] = backend.ConnectWebSocket(t, token)
    }
    fmt.Println("Connected 100K users to WebSocket")
    
    // Send messages at 50 msg/sec for 60 seconds
    metricsCollector := NewMetricsCollector(backend.Metrics)
    
    var wg sync.WaitGroup
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // 50 concurrent senders, each sends 1 msg/sec
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(senderIdx int) {
            defer wg.Done()
            ticker := time.NewTicker(1 * time.Second)
            defer ticker.Stop()
            
            for {
                select {
                case <-ctx.Done():
                    return
                case <-ticker.C:
                    // Random sender + recipient
                    senderID := rand.Intn(100000)
                    recipientID := rand.Intn(100000)
                    if senderID == recipientID {
                        recipientID = (senderID + 1) % 100000
                    }
                    
                    startTime := time.Now()
                    
                    msg := &Message{
                        ID:             uuid.New().String(),
                        SenderDID:      users[senderID].DID,
                        RecipientDID:   users[recipientID].DID,
                        Ciphertext:     randomBytes(256),
                        Signature:      randomBytes(64),
                        Commitment:     randomHash(),
                    }
                    
                    err := backend.MessageService.SendMessage(context.Background(), msg)
                    latency := time.Since(startTime).Milliseconds()
                    
                    if err == nil {
                        metricsCollector.RecordSuccess(latency)
                    } else {
                        metricsCollector.RecordFailure(err)
                    }
                }
            }
        }(i)
    }
    
    // Wait for test completion
    wg.Wait()
    
    // Print results
    results := metricsCollector.GetResults()
    fmt.Printf("\n=== Message Relay Stress Test Results ===\n")
    fmt.Printf("Total messages: %d\n", results.TotalMessages)
    fmt.Printf("Success rate: %.2f%%\n", results.SuccessRate*100)
    fmt.Printf("P50 latency: %dms\n", results.P50Latency)
    fmt.Printf("P99 latency: %dms\n", results.P99Latency)
    fmt.Printf("Max latency: %dms\n", results.MaxLatency)
    fmt.Printf("Errors: %d\n", results.Errors)
    
    // Assertions
    require.Greater(t, results.SuccessRate, 0.99)  // 99% success
    require.Less(t, results.P50Latency, int64(100)) // P50 < 100ms
    require.Less(t, results.P99Latency, int64(500)) // P99 < 500ms
}
```

---

## 5. Test Execution Plan

### 5.1 Local Development

```bash
# Unit tests only
make test

# Unit + integration tests
make test-integration

# All tests including stress
make test-all

# With coverage report
make test-coverage
# Open coverage.html in browser
```

### 5.2 CI/CD Pipeline

**GitHub Actions (`.github/workflows/test.yml`):**
```yaml
name: Test Suite

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: 1.21
      - run: make test
      - run: make test-coverage
      - uses: codecov/codecov-action@v3

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
      redis:
        image: redis:7
      metagraph:
        image: metagraph-simulator:latest
      cardano:
        image: cardano-node:preview
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test-integration

  e2e-staging:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      - run: make test-e2e-staging
```

---

## 6. Test Maintenance

### 6.1 Test Data Cleanup

- Unit tests: Automatic cleanup (in-memory DB)
- Integration tests: Database transaction rollback
- E2E tests: Teardown all test users from backend

### 6.2 Flaky Test Detection

- Automatic retry on failure (max 2 retries)
- Metrics collection for timeout patterns
- Weekly flaky test report

### 6.3 Coverage Targets

| Layer | Target | Current | Status |
|-------|--------|---------|--------|
| Service layer | 90% | — | Phase 2 |
| Integration | 80% | — | Phase 2 |
| E2E | 70% | — | Phase 2 |

---

*ECHO Testing Strategy v1.0*
*Updated: February 23, 2026*
*Ready for Phase 1–2 implementation alongside backend development*
