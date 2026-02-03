package oidc4vc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// FlowManager manages OIDC4VC authorization flows
type FlowManager struct {
	config             *Config
	authorizationCodes map[string]*AuthorizationCode
	preAuthorizedCodes map[string]*PreAuthorizedCode
	accessTokens       map[string]*AccessToken
	tokenMutex         sync.RWMutex
}

// Config represents OIDC4VC flow configuration
type Config struct {
	IssuerDID                string
	IssuerBaseURL            string
	AuthorizationCodeTTL     time.Duration
	PreAuthorizedCodeTTL     time.Duration
	AccessTokenTTL           time.Duration
	EnablePKCE               bool
	RequireProofOfPossession bool
}

// NewFlowManager creates new flow manager
func NewFlowManager(config *Config) *FlowManager {
	fm := &FlowManager{
		config:             config,
		authorizationCodes: make(map[string]*AuthorizationCode),
		preAuthorizedCodes: make(map[string]*PreAuthorizedCode),
		accessTokens:       make(map[string]*AccessToken),
	}

	// Start cleanup routine
	go fm.cleanupExpiredTokens()

	return fm
}

// CreateAuthorizationCode creates authorization code for auth code flow
func (fm *FlowManager) CreateAuthorizationCode(clientID, redirectURI, scope, state string, challenge string) (string, error) {
	code, err := generateRandomCode(32)
	if err != nil {
		return "", err
	}

	fm.tokenMutex.Lock()
	defer fm.tokenMutex.Unlock()

	fm.authorizationCodes[code] = &AuthorizationCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		ExpiresAt:           time.Now().Add(fm.config.AuthorizationCodeTTL),
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}

	return code, nil
}

// ExchangeAuthorizationCode exchanges authorization code for access token
func (fm *FlowManager) ExchangeAuthorizationCode(code, clientID, codeVerifier, redirectURI string) (*TokenResponse, error) {
	fm.tokenMutex.Lock()
	defer fm.tokenMutex.Unlock()

	// Find authorization code
	authCode, exists := fm.authorizationCodes[code]
	if !exists {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Check expiration
	if time.Now().After(authCode.ExpiresAt) {
		delete(fm.authorizationCodes, code)
		return nil, fmt.Errorf("authorization code expired")
	}

	// Verify client
	if authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("client verification failed")
	}

	// Verify PKCE if enabled
	if fm.config.EnablePKCE && authCode.CodeChallenge != "" {
		if !verifyCodeChallenge(codeVerifier, authCode.CodeChallenge) {
			return nil, fmt.Errorf("PKCE verification failed")
		}
	}

	// Create access token
	accessToken, err := generateRandomCode(32)
	if err != nil {
		return nil, err
	}

	cnonce, _ := generateRandomCode(16)

	fm.accessTokens[accessToken] = &AccessToken{
		Token:           accessToken,
		ClientID:        clientID,
		Scope:           authCode.Scope,
		ExpiresAt:       time.Now().Add(fm.config.AccessTokenTTL),
		CNonce:          cnonce,
		CNonceExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Remove authorization code
	delete(fm.authorizationCodes, code)

	return &TokenResponse{
		AccessToken:     accessToken,
		TokenType:       "Bearer",
		ExpiresIn:       int64(fm.config.AccessTokenTTL.Seconds()),
		Scope:           authCode.Scope,
		CNonce:          cnonce,
		CNonceExpiresIn: 300,
	}, nil
}

// CreatePreAuthorizedCode creates pre-authorized code for direct issuance
func (fm *FlowManager) CreatePreAuthorizedCode(credentialType string, pinRequired bool, pinLength int) (string, error) {
	code, err := generateRandomCode(32)
	if err != nil {
		return "", err
	}

	fm.tokenMutex.Lock()
	defer fm.tokenMutex.Unlock()

	fm.preAuthorizedCodes[code] = &PreAuthorizedCode{
		Code:           code,
		CredentialType: credentialType,
		ExpiresAt:      time.Now().Add(fm.config.PreAuthorizedCodeTTL),
		PINRequired:    pinRequired,
		PINLength:      pinLength,
		MaxAttempts:    3,
	}

	return code, nil
}

// ExchangePreAuthorizedCode exchanges pre-authorized code for access token
func (fm *FlowManager) ExchangePreAuthorizedCode(code string, txCode string) (*TokenResponse, error) {
	fm.tokenMutex.Lock()
	defer fm.tokenMutex.Unlock()

	preAuthCode, exists := fm.preAuthorizedCodes[code]
	if !exists {
		return nil, fmt.Errorf("invalid pre-authorized code")
	}

	// Check expiration
	if time.Now().After(preAuthCode.ExpiresAt) {
		delete(fm.preAuthorizedCodes, code)
		return nil, fmt.Errorf("pre-authorized code expired")
	}

	// Verify PIN if required
	if preAuthCode.PINRequired && txCode == "" {
		return nil, fmt.Errorf("transaction code required")
	}

	// Create access token
	accessToken, err := generateRandomCode(32)
	if err != nil {
		return nil, err
	}

	cnonce, _ := generateRandomCode(16)

	fm.accessTokens[accessToken] = &AccessToken{
		Token:           accessToken,
		Scope:           fmt.Sprintf("credential_%s", preAuthCode.CredentialType),
		ExpiresAt:       time.Now().Add(fm.config.AccessTokenTTL),
		CNonce:          cnonce,
		CNonceExpiresAt: time.Now().Add(5 * time.Minute),
	}

	// Remove pre-authorized code
	delete(fm.preAuthorizedCodes, code)

	return &TokenResponse{
		AccessToken:     accessToken,
		TokenType:       "Bearer",
		ExpiresIn:       int64(fm.config.AccessTokenTTL.Seconds()),
		CNonce:          cnonce,
		CNonceExpiresIn: 300,
	}, nil
}

// ValidateAccessToken validates an access token
func (fm *FlowManager) ValidateAccessToken(token string) (*AccessToken, error) {
	fm.tokenMutex.RLock()
	defer fm.tokenMutex.RUnlock()

	accessToken, exists := fm.accessTokens[token]
	if !exists {
		return nil, fmt.Errorf("invalid access token")
	}

	// Check expiration
	if time.Now().After(accessToken.ExpiresAt) {
		return nil, fmt.Errorf("access token expired")
	}

	return accessToken, nil
}

// GenerateCNonce generates new credential nonce
func (fm *FlowManager) GenerateCNonce(accessToken string) (string, error) {
	fm.tokenMutex.Lock()
	defer fm.tokenMutex.Unlock()

	token, exists := fm.accessTokens[accessToken]
	if !exists {
		return "", fmt.Errorf("invalid access token")
	}

	// Check nonce expiration
	if time.Now().After(token.CNonceExpiresAt) {
		// Generate new nonce
		newNonce, err := generateRandomCode(16)
		if err != nil {
			return "", err
		}

		token.CNonce = newNonce
		token.CNonceExpiresAt = time.Now().Add(5 * time.Minute)
		return newNonce, nil
	}

	return token.CNonce, nil
}

// cleanupExpiredTokens removes expired tokens
func (fm *FlowManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		fm.tokenMutex.Lock()

		now := time.Now()

		// Clean authorization codes
		for code, authCode := range fm.authorizationCodes {
			if now.After(authCode.ExpiresAt) {
				delete(fm.authorizationCodes, code)
			}
		}

		// Clean pre-authorized codes
		for code, preAuthCode := range fm.preAuthorizedCodes {
			if now.After(preAuthCode.ExpiresAt) {
				delete(fm.preAuthorizedCodes, code)
			}
		}

		// Clean access tokens
		for token, accessToken := range fm.accessTokens {
			if now.After(accessToken.ExpiresAt) {
				delete(fm.accessTokens, token)
			}
		}

		fm.tokenMutex.Unlock()
	}
}

// Helper functions

// generateRandomCode generates a random code for tokens
func generateRandomCode(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// verifyCodeChallenge verifies PKCE code challenge
func verifyCodeChallenge(codeVerifier, codeChallenge string) bool {
	// Create S256 code challenge
	hash := sha256.Sum256([]byte(codeVerifier))
	computedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return computedChallenge == codeChallenge
}

// GenerateCodeChallenge generates PKCE code challenge
func GenerateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateCodeVerifier generates PKCE code verifier
func GenerateCodeVerifier() (string, error) {
	return generateRandomCode(32)
}
