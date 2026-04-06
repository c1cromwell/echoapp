package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// E.164 phone number validation pattern
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// AuthService orchestrates all authentication operations:
// registration, login, token management, device management, and recovery.
type AuthService struct {
	mu sync.RWMutex

	// Sub-services
	OTP         *OTPService
	Tokens      *TokenService
	Devices     *DeviceService
	RateLimiter *AuthRateLimiter
	StepUp      *StepUpService
	Recovery    *RecoveryService
	Audit       *AuditLogger

	// In-memory user store (production: Postgres)
	users       map[string]*User // key: user_id
	usersByDID  map[string]*User // key: did
	usersByPhone map[string]*User // key: phone_hash

	// Credential store (production: Postgres)
	credentials map[string]*CredentialRecord // key: credential_id
}

// NewAuthService creates a fully initialized auth service.
func NewAuthService() (*AuthService, error) {
	tokenService, err := NewTokenService()
	if err != nil {
		return nil, fmt.Errorf("create auth service: %w", err)
	}

	return &AuthService{
		OTP:          NewOTPService(),
		Tokens:       tokenService,
		Devices:      NewDeviceService(),
		RateLimiter:  NewAuthRateLimiter(),
		StepUp:       NewStepUpService(),
		Recovery:     NewRecoveryService(),
		Audit:        NewAuditLogger(),
		users:        make(map[string]*User),
		usersByDID:   make(map[string]*User),
		usersByPhone: make(map[string]*User),
		credentials:  make(map[string]*CredentialRecord),
	}, nil
}

// --- Registration Flow ---

// RegisterPhone initiates phone registration by sending an OTP.
// Returns the same response shape regardless of whether the phone is new or existing
// (prevents enumeration).
func (s *AuthService) RegisterPhone(req PhoneRegistrationRequest, ip string) (*PhoneRegistrationResponse, *AuthError) {
	// 1. Validate phone format
	fullPhone := req.CountryCode + req.PhoneNumber
	if !e164Pattern.MatchString(fullPhone) {
		return nil, NewAuthError(ErrCodeInvalidPhone, 400)
	}

	// 2. Validate device
	if err := ValidateDeviceInfo(req.DeviceInfo); err != nil {
		return nil, err
	}
	if err := ValidateDeviceIntegrity(req.DeviceInfo); err != nil {
		return nil, err
	}

	phoneHash := hashPhone(fullPhone)

	// 3. Rate limit check
	key := FormatRateLimitKey("otp", phoneHash)
	if err := s.RateLimiter.Check(key, RateLimitOTPSend); err != nil {
		err.Code = ErrCodeOTPRateLimit
		err.Message = ErrCodeOTPRateLimit.UserFacingMessage()
		return nil, err
	}

	// 4. Generate OTP (even if phone is already registered — prevent enumeration)
	code, err := GenerateOTP()
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	codeHash, err := HashOTP(code)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	// 5. Store OTP session
	verificationID := uuid.New().String()
	s.OTP.CreateSession(verificationID, phoneHash, codeHash)
	s.OTP.RecordOTPSend(phoneHash)

	// 6. In production: send via silent push or Twilio SMS
	// For development, the code is logged (production would not do this)

	// 7. Audit
	s.Audit.Log("", AuditEventRegister, AuditResultSuccess, ip, "", &req.DeviceInfo, "", map[string]interface{}{
		"step": "phone_registration",
		"phone_hash": phoneHash[:8] + "...",
	})

	expiresAt := time.Now().Add(OTPExpiry).UTC().Format(time.RFC3339)

	return &PhoneRegistrationResponse{
		VerificationID: verificationID,
		ExpiresAt:      expiresAt,
		RetryAfter:     int(OTPCooldown.Seconds()),
	}, nil
}

// VerifyOTP verifies the OTP code and creates a user account.
func (s *AuthService) VerifyOTP(req OTPVerifyRequest, ip string) (*AuthResponse, *AuthError) {
	// 1. Validate device
	if err := ValidateDeviceInfo(req.DeviceInfo); err != nil {
		return nil, err
	}

	// 2. Verify OTP
	success, authErr := s.OTP.VerifyCode(req.VerificationID, req.Code)
	if authErr != nil {
		s.Audit.Log("", AuditEventRegister, AuditResultFailed, ip, "", &req.DeviceInfo, string(authErr.Code), nil)
		return nil, authErr
	}
	if !success {
		return nil, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	// 3. Get session to find phone hash
	session := s.OTP.GetSession(req.VerificationID)
	if session == nil {
		return nil, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	// 4. Check if user already exists for this phone
	s.mu.RLock()
	existingUser := s.usersByPhone[session.PhoneHash]
	s.mu.RUnlock()

	if existingUser != nil {
		// Phone already registered — issue temp token for existing user
		deviceHash := ComputeDeviceHash(req.DeviceInfo)
		token, claims, err := s.Tokens.IssueTempAccessToken(existingUser.DID, deviceHash)
		if err != nil {
			return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
		}

		s.Audit.Log(existingUser.ID, AuditEventRegister, AuditResultSuccess, ip, deviceHash, &req.DeviceInfo, "", map[string]interface{}{
			"step": "otp_verify_existing",
		})

		return &AuthResponse{
			AccessToken: token,
			ExpiresAt:   time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
			User:        existingUser,
		}, nil
	}

	// 5. Create new user
	userID := "usr_" + uuid.New().String()[:12]
	did := "did:prism:cardano:pending"

	user := &User{
		ID:         userID,
		DID:        did,
		PhoneHash:  session.PhoneHash,
		Status:     UserStatusPendingPasskey,
		TrustScore: 0,
		TrustTier:  0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.users[userID] = user
	s.usersByDID[did] = user
	s.usersByPhone[session.PhoneHash] = user
	s.mu.Unlock()

	// 6. Issue temp token for passkey registration
	deviceHash := ComputeDeviceHash(req.DeviceInfo)
	token, claims, err := s.Tokens.IssueTempAccessToken(did, deviceHash)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	// 7. Generate passkey challenge
	challenge, err := generateChallenge()
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	// 8. Audit
	s.Audit.Log(userID, AuditEventRegister, AuditResultSuccess, ip, deviceHash, &req.DeviceInfo, "", map[string]interface{}{
		"step": "otp_verify_new_user",
	})

	return &AuthResponse{
		AccessToken:      token,
		ExpiresAt:        time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
		User:             user,
		PasskeyChallenge: challenge,
	}, nil
}

// RegisterPasskey completes registration by storing the WebAuthn credential.
func (s *AuthService) RegisterPasskey(userDID string, req PasskeyRegistrationRequest, ip string) (*AuthResponse, *AuthError) {
	// 1. Validate device
	if err := ValidateDeviceInfo(req.DeviceInfo); err != nil {
		return nil, err
	}

	// 2. Find user
	s.mu.RLock()
	user := s.usersByDID[userDID]
	s.mu.RUnlock()

	if user == nil {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// 3. In production: verify WebAuthn attestation via go-webauthn library
	// For now, store the credential directly
	deviceHash := ComputeDeviceHash(req.DeviceInfo)
	credentialID := req.AttestationResponse.ID
	if credentialID == "" {
		credentialID = uuid.New().String()
	}

	credential := &CredentialRecord{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		CredentialID: credentialID,
		PublicKey:    []byte(req.AttestationResponse.Response.AttestationObject),
		SignCount:    0,
		DeviceID:     deviceHash,
		FriendlyName: req.DeviceInfo.Model,
		CreatedAt:    time.Now(),
	}

	s.mu.Lock()
	s.credentials[credentialID] = credential
	// Update user status and trust
	user.Status = UserStatusActive
	user.TrustScore = 5 // device-verified baseline
	user.DID = "did:prism:cardano:" + uuid.New().String()[:12]
	user.UpdatedAt = time.Now()
	s.usersByDID[user.DID] = user
	s.mu.Unlock()

	// 4. Register device
	s.Devices.RegisterDevice(user.ID, req.DeviceInfo, ip, credentialID)

	// 5. Issue full tokens
	scope := scopeForTrustTier(user.TrustTier)
	accessToken, claims, err := s.Tokens.IssueAccessToken(user.DID, deviceHash, user.TrustTier, scope)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	refreshToken := GenerateRefreshToken()
	s.Tokens.StoreRefreshToken(user.ID, refreshToken, deviceHash)

	// 6. Audit
	s.Audit.Log(user.ID, AuditEventRegister, AuditResultSuccess, ip, deviceHash, &req.DeviceInfo, "", map[string]interface{}{
		"step": "passkey_registered",
	})

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
		User:         user,
	}, nil
}

// --- Login Flow ---

// LoginChallenge generates a WebAuthn challenge for passkey login.
func (s *AuthService) LoginChallenge() (*LoginChallengeResponse, error) {
	challenge, err := generateChallenge()
	if err != nil {
		return nil, err
	}
	return &LoginChallengeResponse{
		Challenge: challenge,
		Timeout:   300000, // 5 minutes in ms
		RPID:      "echo.app",
	}, nil
}

// Login authenticates a user via passkey or DID signature.
func (s *AuthService) Login(req LoginRequest, ip string) (*AuthResponse, *AuthError) {
	// 1. Validate device
	if err := ValidateDeviceInfo(req.DeviceInfo); err != nil {
		return nil, err
	}
	if err := ValidateDeviceIntegrity(req.DeviceInfo); err != nil {
		return nil, err
	}

	// 2. Rate limit
	key := FormatRateLimitKey("login:ip", ip)
	if err := s.RateLimiter.Check(key, RateLimitLoginIP); err != nil {
		return nil, err
	}

	deviceHash := ComputeDeviceHash(req.DeviceInfo)

	switch req.AuthType {
	case "passkey":
		return s.loginPasskey(req, ip, deviceHash)
	case "did_signature":
		return s.loginDIDSignature(req, ip, deviceHash)
	default:
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}
}

func (s *AuthService) loginPasskey(req LoginRequest, ip string, deviceHash string) (*AuthResponse, *AuthError) {
	if req.Credential == nil {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// Look up credential
	s.mu.RLock()
	cred := s.credentials[req.Credential.ID]
	s.mu.RUnlock()

	if cred == nil {
		s.Audit.Log("", AuditEventLogin, AuditResultFailed, ip, deviceHash, &req.DeviceInfo, string(ErrCodePasskeyFailed), nil)
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// In production: verify WebAuthn assertion signature
	// For now: accept if credential exists

	// Check sign count (detect cloned passkeys)
	// In production: parse authenticator data for actual sign count

	// Device check
	s.mu.RLock()
	user := s.users[cred.UserID]
	s.mu.RUnlock()

	if user == nil {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	if !s.Devices.IsKnownDevice(user.ID, deviceHash) {
		s.Audit.Log(user.ID, AuditEventLogin, AuditResultFailed, ip, deviceHash, &req.DeviceInfo, string(ErrCodeUnknownDevice), nil)
		return nil, NewAuthError(ErrCodeUnknownDevice, 403)
	}

	// Update device activity
	s.Devices.UpdateDeviceActivity(user.ID, deviceHash, ip)

	// Update credential
	s.mu.Lock()
	cred.SignCount++
	now := time.Now()
	cred.LastUsedAt = &now
	s.mu.Unlock()

	// Issue tokens
	scope := scopeForTrustTier(user.TrustTier)
	accessToken, claims, err := s.Tokens.IssueAccessToken(user.DID, deviceHash, user.TrustTier, scope)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	refreshToken := GenerateRefreshToken()
	s.Tokens.StoreRefreshToken(user.ID, refreshToken, deviceHash)

	// Audit
	s.Audit.Log(user.ID, AuditEventLogin, AuditResultSuccess, ip, deviceHash, &req.DeviceInfo, "", nil)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
		User:         user,
	}, nil
}

func (s *AuthService) loginDIDSignature(req LoginRequest, ip string, deviceHash string) (*AuthResponse, *AuthError) {
	if req.DID == "" || req.Signature == "" || req.Nonce == "" || req.Timestamp == "" {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// 1. Anti-replay: check timestamp within 5-minute window
	ts, parseErr := time.Parse(time.RFC3339, req.Timestamp)
	if parseErr != nil {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}
	if time.Since(ts).Abs() > 5*time.Minute {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// 2. Anti-replay: check nonce uniqueness
	if !s.Tokens.CheckAndStoreNonce(req.Nonce) {
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// 3. Rate limit per DID
	key := FormatRateLimitKey("login:did", hashDID(req.DID))
	if err := s.RateLimiter.Check(key, RateLimitLoginAccount); err != nil {
		return nil, err
	}

	// 4. Find user by DID
	s.mu.RLock()
	user := s.usersByDID[req.DID]
	s.mu.RUnlock()

	if user == nil {
		s.Audit.Log("", AuditEventLogin, AuditResultFailed, ip, deviceHash, &req.DeviceInfo, string(ErrCodePasskeyFailed), nil)
		return nil, NewAuthError(ErrCodePasskeyFailed, 401)
	}

	// 5. In production: resolve DID document, verify Ed25519 signature
	// Signed message format: echo:auth:{did}:{timestamp}:{nonce}

	// 6. Device check
	if !s.Devices.IsKnownDevice(user.ID, deviceHash) {
		s.Audit.Log(user.ID, AuditEventLogin, AuditResultFailed, ip, deviceHash, &req.DeviceInfo, string(ErrCodeUnknownDevice), nil)
		return nil, NewAuthError(ErrCodeUnknownDevice, 403)
	}

	s.Devices.UpdateDeviceActivity(user.ID, deviceHash, ip)

	// 7. Issue tokens
	scope := scopeForTrustTier(user.TrustTier)
	accessToken, claims, err := s.Tokens.IssueAccessToken(user.DID, deviceHash, user.TrustTier, scope)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	refreshToken := GenerateRefreshToken()
	s.Tokens.StoreRefreshToken(user.ID, refreshToken, deviceHash)

	s.Audit.Log(user.ID, AuditEventLogin, AuditResultSuccess, ip, deviceHash, &req.DeviceInfo, "", nil)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
		User:         user,
	}, nil
}

// --- Token Refresh ---

// RefreshTokens rotates access and refresh tokens.
func (s *AuthService) RefreshTokens(req RefreshRequest, deviceInfo DeviceInfo, ip string) (*AuthResponse, *AuthError) {
	deviceHash := ComputeDeviceHash(deviceInfo)

	// Rate limit
	// We use the token hash as identifier since we don't know the user yet
	tokenHash := HashRefreshToken(req.RefreshToken)[:16]
	key := FormatRateLimitKey("refresh", tokenHash)
	if err := s.RateLimiter.Check(key, RateLimitRefresh); err != nil {
		return nil, err
	}

	// Rotate
	newToken, record, authErr := s.Tokens.RotateRefreshToken(req.RefreshToken, deviceHash)
	if authErr != nil {
		s.Audit.Log("", AuditEventRefresh, AuditResultFailed, ip, deviceHash, &deviceInfo, string(authErr.Code), nil)
		return nil, authErr
	}

	// Find user
	s.mu.RLock()
	user := s.users[record.UserID]
	s.mu.RUnlock()

	if user == nil {
		return nil, NewAuthError(ErrCodeRefreshInvalid, 401)
	}

	// Issue new access token
	scope := scopeForTrustTier(user.TrustTier)
	accessToken, claims, err := s.Tokens.IssueAccessToken(user.DID, deviceHash, user.TrustTier, scope)
	if err != nil {
		return nil, &AuthError{Code: "AUTH_INTERNAL", Message: "Internal error", HTTPStatus: 500}
	}

	s.Audit.Log(user.ID, AuditEventRefresh, AuditResultSuccess, ip, deviceHash, &deviceInfo, "", nil)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newToken,
		ExpiresAt:    time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
		User:         user,
	}, nil
}

// --- Logout ---

// Logout invalidates the current session or all sessions.
func (s *AuthService) Logout(claims *TokenClaims, allDevices bool, ip string) {
	// Blocklist the current access token
	s.Tokens.BlocklistToken(claims.TokenID, time.Unix(claims.ExpiresAt, 0))

	if allDevices {
		// Find user and revoke all refresh tokens
		s.mu.RLock()
		user := s.usersByDID[claims.Subject]
		s.mu.RUnlock()

		if user != nil {
			s.Tokens.RevokeAllUserTokens(user.ID)
		}
	}

	s.Audit.Log("", AuditEventLogout, AuditResultSuccess, ip, claims.DeviceID, nil, "", map[string]interface{}{
		"all_devices": allDevices,
	})
}

// --- User Lookups ---

// GetUser returns a user by ID.
func (s *AuthService) GetUser(userID string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[userID]
}

// GetUserByDID returns a user by DID.
func (s *AuthService) GetUserByDID(did string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.usersByDID[did]
}

// UserCount returns the total number of registered users.
func (s *AuthService) UserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// --- Helpers ---

func hashPhone(phone string) string {
	hash := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(hash[:])
}

func hashDID(did string) string {
	hash := sha256.Sum256([]byte(did))
	return hex.EncodeToString(hash[:16])
}

func generateChallenge() (string, error) {
	b := make([]byte, 32)
	if _, err := cryptoRandRead(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// cryptoRandRead is a variable for testing
var cryptoRandRead = cryptoRandReadDefault

func cryptoRandReadDefault(b []byte) (int, error) {
	return rand.Read(b)
}

func scopeForTrustTier(tier int) string {
	switch {
	case tier >= 3:
		return "messaging payments governance staking"
	case tier >= 2:
		return "messaging payments"
	case tier >= 1:
		return "messaging"
	default:
		return "messaging"
	}
}
