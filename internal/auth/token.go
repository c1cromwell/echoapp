package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RedisBackend is the subset of the Redis client used for durable token state
// (revocation blocklist + single-use nonces). It is satisfied structurally by
// *infra.RedisClient, so no adapter is needed. When unset, TokenService falls
// back to in-memory maps (dev/test) — which do NOT survive a restart.
type RedisBackend interface {
	BlocklistToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsBlocklisted(ctx context.Context, jti string) (bool, error)
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
}

const (
	AccessTokenTTL          = 15 * time.Minute
	TempAccessTokenTTL      = 5 * time.Minute
	RefreshTokenTTL         = 30 * 24 * time.Hour // 30 days
	ElevatedTokenTTLDefault = 5 * time.Minute
	ElevatedTokenTTLShort   = 2 * time.Minute
	Issuer                  = "https://api.echo.app"
)

// TokenService manages JWT creation, validation, and refresh token lifecycle.
type TokenService struct {
	mu         sync.RWMutex
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	keyID      string

	// redis, when set, is the durable backing store for revocation + nonces.
	// When nil, the in-memory maps below are used (dev/test; not restart-safe).
	redis RedisBackend

	// In-memory stores (fallback when redis is nil).
	blocklist     map[string]time.Time           // jti -> expiry
	refreshTokens map[string]*RefreshTokenRecord // token_hash -> record
	usedNonces    map[string]time.Time           // nonce -> expiry
}

// SetRedisBackend wires a durable store for the blocklist and nonce sets so
// revocation and anti-replay survive process restarts. Safe to call once at
// startup before the service handles traffic.
func (ts *TokenService) SetRedisBackend(r RedisBackend) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.redis = r
}

// NewTokenService creates a token service whose ES256 signing key is loaded
// from the JWT_SIGNING_KEY env var (hex-encoded 32-byte P-256 scalar) so tokens
// survive process restarts. When the var is unset:
//   - in production (ENVIRONMENT=production) this is a fatal misconfiguration;
//   - otherwise an ephemeral key is generated and a warning is logged (dev/test).
//
// Generate a key for ops with: openssl rand -hex 32
func NewTokenService() (*TokenService, error) {
	privateKey, persistent, err := loadOrGenerateSigningKey()
	if err != nil {
		return nil, err
	}
	if !persistent {
		log.Println("⚠ auth: JWT_SIGNING_KEY unset — using an EPHEMERAL signing key; all tokens are invalidated on restart. Set JWT_SIGNING_KEY (openssl rand -hex 32) for persistence.")
	}

	return &TokenService{
		privateKey:    privateKey,
		publicKey:     &privateKey.PublicKey,
		keyID:         deriveKeyID(&privateKey.PublicKey),
		blocklist:     make(map[string]time.Time),
		refreshTokens: make(map[string]*RefreshTokenRecord),
		usedNonces:    make(map[string]time.Time),
	}, nil
}

// loadOrGenerateSigningKey returns the configured persistent key, or generates an
// ephemeral one. The bool reports whether the key is persistent (loaded from env).
func loadOrGenerateSigningKey() (*ecdsa.PrivateKey, bool, error) {
	if hexKey := os.Getenv("JWT_SIGNING_KEY"); hexKey != "" {
		key, err := parseSigningKeyHex(hexKey)
		if err != nil {
			return nil, false, fmt.Errorf("JWT_SIGNING_KEY invalid: %w", err)
		}
		return key, true, nil
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		return nil, false, fmt.Errorf("JWT_SIGNING_KEY is required in production (generate with: openssl rand -hex 32)")
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, false, fmt.Errorf("generate signing key: %w", err)
	}
	return key, false, nil
}

// parseSigningKeyHex reconstructs a P-256 private key from a hex-encoded 32-byte
// scalar (the private value d). The public key is recomputed as d·G.
func parseSigningKeyHex(h string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, fmt.Errorf("not valid hex: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("expected 32 bytes (64 hex chars), got %d", len(b))
	}
	d := new(big.Int).SetBytes(b)
	curve := elliptic.P256()
	if d.Sign() <= 0 || d.Cmp(curve.Params().N) >= 0 {
		return nil, fmt.Errorf("scalar out of range for P-256")
	}
	x, y := curve.ScalarBaseMult(b)
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: curve, X: x, Y: y},
		D:         d,
	}, nil
}

// deriveKeyID returns a stable key id (first 8 hex of SHA-256 of the public point)
// so the same signing key yields the same kid across restarts.
func deriveKeyID(pub *ecdsa.PublicKey) string {
	marshaled := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	sum := sha256.Sum256(marshaled)
	return hex.EncodeToString(sum[:])[:8]
}

// IssueAccessToken creates a signed access token for the given claims.
func (ts *TokenService) IssueAccessToken(userDID string, deviceID string, trustTier int, scope string) (string, *TokenClaims, error) {
	now := time.Now()
	claims := &TokenClaims{
		Subject:   userDID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(AccessTokenTTL).Unix(),
		TokenID:   uuid.New().String(),
		DeviceID:  deviceID,
		TrustTier: trustTier,
		Scope:     scope,
		Elevated:  false,
	}

	token, err := ts.signClaims(claims)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

// IssueTempAccessToken creates a short-lived token for passkey registration.
func (ts *TokenService) IssueTempAccessToken(userDID string, deviceID string) (string, *TokenClaims, error) {
	now := time.Now()
	claims := &TokenClaims{
		Subject:   userDID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(TempAccessTokenTTL).Unix(),
		TokenID:   uuid.New().String(),
		DeviceID:  deviceID,
		TrustTier: 0,
		Scope:     "passkey_registration",
		Elevated:  false,
	}

	token, err := ts.signClaims(claims)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

// IssueElevatedToken creates a short-lived elevated token for step-up auth.
func (ts *TokenService) IssueElevatedToken(userDID string, deviceID string, trustTier int, action string, ttl time.Duration) (string, *TokenClaims, error) {
	now := time.Now()
	claims := &TokenClaims{
		Subject:        userDID,
		IssuedAt:       now.Unix(),
		ExpiresAt:      now.Add(ttl).Unix(),
		TokenID:        uuid.New().String(),
		DeviceID:       deviceID,
		TrustTier:      trustTier,
		Scope:          "elevated",
		Elevated:       true,
		ElevatedAction: action,
	}

	token, err := ts.signClaims(claims)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

// signClaims signs token claims with the ES256 private key.
// This is a simplified JWT-like signing (production uses golang-jwt/jwt/v5).
func (ts *TokenService) signClaims(claims *TokenClaims) (string, error) {
	// Encode claims as JSON
	payload := fmt.Sprintf(
		`{"iss":"%s","sub":"%s","iat":%d,"exp":%d,"jti":"%s","device_id":"%s","trust_tier":%d,"scope":"%s","elevated":%t,"elevated_action":"%s"}`,
		Issuer, claims.Subject, claims.IssuedAt, claims.ExpiresAt,
		claims.TokenID, claims.DeviceID, claims.TrustTier, claims.Scope,
		claims.Elevated, claims.ElevatedAction,
	)

	// ES256 sign
	hash := sha256.Sum256([]byte(payload))
	r, s, err := ecdsa.Sign(rand.Reader, ts.privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	sig := append(r.Bytes(), s.Bytes()...)
	header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"ES256","kid":"%s"}`, ts.keyID)))
	body := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(sig)

	return header + "." + body + "." + signature, nil
}

// ValidateAccessToken verifies and decodes an access token.
func (ts *TokenService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	claims, err := ts.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if claims.ExpiresAt < now {
		return nil, fmt.Errorf("token expired")
	}

	if ts.IsBlocklisted(claims.TokenID) {
		return nil, fmt.Errorf("token revoked")
	}

	return claims, nil
}

// parseToken decodes and verifies the token signature.
func (ts *TokenService) parseToken(tokenString string) (*TokenClaims, error) {
	// Split into parts
	parts := splitToken(tokenString)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed token")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	// Verify signature
	hash := sha256.Sum256(payloadBytes)
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	if len(sigBytes) < 64 {
		return nil, fmt.Errorf("invalid signature length")
	}

	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])

	if !ecdsa.Verify(ts.publicKey, hash[:], r, s) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Parse claims from JSON (simplified — production uses proper JSON decoder)
	claims, err := parseClaimsJSON(payloadBytes)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func splitToken(token string) []string {
	var parts []string
	start := 0
	for i, c := range token {
		if c == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

// parseClaimsJSON extracts claims from JSON bytes.
func parseClaimsJSON(data []byte) (*TokenClaims, error) {
	// Use encoding/json for proper parsing
	claims := &TokenClaims{}
	if err := jsonUnmarshal(data, claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}
	return claims, nil
}

// BlocklistToken adds a token ID to the blocklist. Durable when a Redis backend
// is configured; otherwise in-memory (lost on restart).
func (ts *TokenService) BlocklistToken(jti string, expiresAt time.Time) {
	if r := ts.backend(); r != nil {
		if err := r.BlocklistToken(context.Background(), jti, expiresAt); err == nil {
			return
		} else {
			log.Printf("auth: redis blocklist write failed for jti=%s, falling back to memory: %v", jti, err)
		}
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.blocklist[jti] = expiresAt
}

// IsBlocklisted checks if a token ID has been revoked.
func (ts *TokenService) IsBlocklisted(jti string) bool {
	if r := ts.backend(); r != nil {
		if blocked, err := r.IsBlocklisted(context.Background(), jti); err == nil {
			return blocked
		} else {
			log.Printf("auth: redis blocklist check failed for jti=%s, falling back to memory: %v", jti, err)
		}
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	expiry, exists := ts.blocklist[jti]
	if !exists {
		return false
	}
	// Auto-clean expired entries
	if time.Now().After(expiry) {
		return false
	}
	return true
}

// backend returns the configured Redis backend (read under lock).
func (ts *TokenService) backend() RedisBackend {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.redis
}

// --- Refresh Token Management ---

// GenerateRefreshToken creates a new opaque refresh token.
func GenerateRefreshToken() string {
	return "rt_" + uuid.New().String()
}

// HashRefreshToken hashes a refresh token for storage.
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// StoreRefreshToken persists a new refresh token record.
func (ts *TokenService) StoreRefreshToken(userID, token, deviceID string) *RefreshTokenRecord {
	record := &RefreshTokenRecord{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: HashRefreshToken(token),
		DeviceID:  deviceID,
		Status:    RefreshTokenActive,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
		CreatedAt: time.Now(),
	}

	ts.mu.Lock()
	ts.refreshTokens[record.TokenHash] = record
	ts.mu.Unlock()

	return record
}

// RotateRefreshToken validates the old token and issues a new one.
// Implements single-use enforcement with replay detection.
func (ts *TokenService) RotateRefreshToken(oldToken, deviceID string) (newToken string, record *RefreshTokenRecord, err *AuthError) {
	oldHash := HashRefreshToken(oldToken)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	existing, exists := ts.refreshTokens[oldHash]
	if !exists {
		return "", nil, NewAuthError(ErrCodeRefreshInvalid, 401)
	}

	// Check expiry
	if time.Now().After(existing.ExpiresAt) {
		existing.Status = RefreshTokenRevoked
		return "", nil, NewAuthError(ErrCodeRefreshInvalid, 401)
	}

	// Device binding check
	if existing.DeviceID != deviceID {
		return "", nil, NewAuthError(ErrCodeUnknownDevice, 403)
	}

	// REPLAY DETECTION: If already used, revoke ALL user tokens
	if existing.Status == RefreshTokenUsed {
		ts.revokeAllUserTokensLocked(existing.UserID)
		return "", nil, NewAuthError(ErrCodeRefreshInvalid, 401)
	}

	if existing.Status == RefreshTokenRevoked {
		return "", nil, NewAuthError(ErrCodeRefreshInvalid, 401)
	}

	// Mark old token as used
	now := time.Now()
	existing.Status = RefreshTokenUsed
	existing.UsedAt = &now

	// Generate new token
	newTokenStr := GenerateRefreshToken()
	newRecord := &RefreshTokenRecord{
		ID:        uuid.New().String(),
		UserID:    existing.UserID,
		TokenHash: HashRefreshToken(newTokenStr),
		DeviceID:  deviceID,
		Status:    RefreshTokenActive,
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
		CreatedAt: time.Now(),
	}
	ts.refreshTokens[newRecord.TokenHash] = newRecord

	return newTokenStr, newRecord, nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (ts *TokenService) RevokeAllUserTokens(userID string) int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.revokeAllUserTokensLocked(userID)
}

func (ts *TokenService) revokeAllUserTokensLocked(userID string) int {
	count := 0
	for _, record := range ts.refreshTokens {
		if record.UserID == userID && record.Status == RefreshTokenActive {
			record.Status = RefreshTokenRevoked
			count++
		}
	}
	return count
}

// GetActiveRefreshTokenCount returns the number of active refresh tokens for a user.
func (ts *TokenService) GetActiveRefreshTokenCount(userID string) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	count := 0
	for _, record := range ts.refreshTokens {
		if record.UserID == userID && record.Status == RefreshTokenActive {
			count++
		}
	}
	return count
}

// --- Nonce Management (DID Signature Anti-Replay) ---

// CheckAndStoreNonce returns true if the nonce is fresh (not previously used),
// atomically marking it used. Durable + atomic when a Redis backend is set.
func (ts *TokenService) CheckAndStoreNonce(nonce string) bool {
	return ts.CheckAndStoreNonceTTL(nonce, 10*time.Minute)
}

// CheckAndStoreNonceTTL is CheckAndStoreNonce with a caller-specified TTL, used
// by request-replay protection where the freshness window is short.
func (ts *TokenService) CheckAndStoreNonceTTL(nonce string, ttl time.Duration) bool {
	if r := ts.backend(); r != nil {
		fresh, err := r.SetNX(context.Background(), "nonce:"+nonce, []byte("1"), ttl)
		if err == nil {
			return fresh
		}
		log.Printf("auth: redis nonce check failed, falling back to memory: %v", err)
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, used := ts.usedNonces[nonce]; used {
		return false
	}
	ts.usedNonces[nonce] = time.Now().Add(ttl)
	return true
}

// CleanExpiredNonces removes expired nonce entries.
func (ts *TokenService) CleanExpiredNonces() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	count := 0
	for nonce, expiry := range ts.usedNonces {
		if now.After(expiry) {
			delete(ts.usedNonces, nonce)
			count++
		}
	}
	return count
}

// PublicKey returns the verification public key.
func (ts *TokenService) PublicKey() *ecdsa.PublicKey {
	return ts.publicKey
}

// KeyID returns the current key ID.
func (ts *TokenService) KeyID() string {
	return ts.keyID
}
