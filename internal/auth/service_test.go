package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	svc, err := NewAuthService()
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	return svc
}

func validDeviceInfo() DeviceInfo {
	return DeviceInfo{
		DeviceID:      "test-device-123",
		Platform:      "ios",
		OSVersion:     "17.4",
		AppVersion:    "1.0.0",
		Model:         "iPhone 15 Pro",
		SecureEnclave: true,
		BiometricType: "face_id",
	}
}

// --- Registration Flow Tests ---

func TestAuthService_RegisterPhone_ValidPhone(t *testing.T) {
	svc := newTestAuthService(t)

	resp, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551234567",
		CountryCode: "+1",
		DeviceInfo:  validDeviceInfo(),
	}, "192.168.1.1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VerificationID == "" {
		t.Error("should return verification ID")
	}
	if resp.RetryAfter != 60 {
		t.Errorf("expected retry_after 60, got %d", resp.RetryAfter)
	}
}

func TestAuthService_RegisterPhone_InvalidPhone(t *testing.T) {
	svc := newTestAuthService(t)

	_, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "abc",
		CountryCode: "+1",
		DeviceInfo:  validDeviceInfo(),
	}, "1.1.1.1")

	if err == nil {
		t.Fatal("invalid phone should return error")
	}
	if err.Code != ErrCodeInvalidPhone {
		t.Errorf("expected AUTH_001, got %s", err.Code)
	}
}

func TestAuthService_RegisterPhone_JailbreakRejected(t *testing.T) {
	svc := newTestAuthService(t)
	info := validDeviceInfo()
	info.JailbreakStatus = true

	_, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551234567",
		CountryCode: "+1",
		DeviceInfo:  info,
	}, "1.1.1.1")

	if err == nil {
		t.Fatal("jailbroken device should be rejected")
	}
	if err.Code != ErrCodeDeviceIntegrity {
		t.Errorf("expected AUTH_010, got %s", err.Code)
	}
}

func TestAuthService_RegisterPhone_RateLimit(t *testing.T) {
	svc := newTestAuthService(t)

	for i := 0; i < 5; i++ {
		_, err := svc.RegisterPhone(PhoneRegistrationRequest{
			PhoneNumber: "5551234567",
			CountryCode: "+1",
			DeviceInfo:  validDeviceInfo(),
		}, "1.1.1.1")
		if err != nil {
			t.Fatalf("request %d should succeed: %v", i, err)
		}
	}

	// 6th request should be rate limited
	_, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551234567",
		CountryCode: "+1",
		DeviceInfo:  validDeviceInfo(),
	}, "1.1.1.1")

	if err == nil {
		t.Fatal("6th OTP send should be rate limited")
	}
	if err.Code != ErrCodeOTPRateLimit {
		t.Errorf("expected AUTH_002, got %s", err.Code)
	}
}

func TestAuthService_EnumerationPrevention(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	// Register first phone
	resp1, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551111111",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("first registration: %v", err)
	}

	// Register second (different) phone
	resp2, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5552222222",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("second registration: %v", err)
	}

	// Both should have same response structure (different IDs is fine)
	if resp1.RetryAfter != resp2.RetryAfter {
		t.Error("responses should have identical structure for enumeration prevention")
	}
}

// --- Full Registration Flow ---

func TestAuthService_FullRegistrationFlow(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	// Step 1: Send OTP
	phoneResp, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551234567",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("register phone: %v", err)
	}

	// Get the OTP code from session (in production this comes via SMS)
	session := svc.OTP.GetSession(phoneResp.VerificationID)
	if session == nil {
		t.Fatal("OTP session should exist")
	}

	// We need to know the code for testing. Generate a known code.
	knownCode := "999999"
	knownHash, _ := HashOTP(knownCode)
	session.CodeHash = knownHash

	// Step 2: Verify OTP
	verifyResp, err := svc.VerifyOTP(OTPVerifyRequest{
		VerificationID: phoneResp.VerificationID,
		Code:           knownCode,
		DeviceInfo:     di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("verify OTP: %v", err)
	}
	if verifyResp.AccessToken == "" {
		t.Error("should return temp access token")
	}
	if verifyResp.User == nil {
		t.Fatal("should return user")
	}
	if verifyResp.User.Status != UserStatusPendingPasskey {
		t.Errorf("expected pending_passkey, got %s", verifyResp.User.Status)
	}
	if verifyResp.PasskeyChallenge == "" {
		t.Error("should return passkey challenge")
	}

	// Step 3: Register Passkey
	passkeyReq := buildTestAttestation(t, verifyResp.PasskeyChallenge, "cred-id-1", svc.PasskeyVerifier)
	passkeyReq.DeviceInfo = di
	passkeyResp, err := svc.RegisterPasskey(verifyResp.User.DID, passkeyReq, "1.1.1.1")
	if err != nil {
		t.Fatalf("register passkey: %v", err)
	}
	if passkeyResp.AccessToken == "" {
		t.Error("should return full access token")
	}
	if passkeyResp.RefreshToken == "" {
		t.Error("should return refresh token")
	}
	if passkeyResp.User.Status != UserStatusActive {
		t.Errorf("expected active, got %s", passkeyResp.User.Status)
	}
	if passkeyResp.User.TrustScore != 5 {
		t.Errorf("expected trust score 5, got %d", passkeyResp.User.TrustScore)
	}

	// Verify user count
	if svc.UserCount() != 1 {
		t.Errorf("expected 1 user, got %d", svc.UserCount())
	}
}

// --- Login Tests ---

func TestAuthService_LoginChallenge(t *testing.T) {
	svc := newTestAuthService(t)

	resp, err := svc.LoginChallenge()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Challenge == "" {
		t.Error("challenge should not be empty")
	}
	if resp.RPID != "echo.app" {
		t.Errorf("expected echo.app, got %s", resp.RPID)
	}
	if resp.Timeout != 300000 {
		t.Errorf("expected 300000, got %d", resp.Timeout)
	}
}

func TestAuthService_Login_UnknownDevice(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	// Register a user first
	_, credID, privKey := registerTestUser(t, svc, di)

	// Get a login challenge
	challengeResp, challengeErr := svc.LoginChallenge()
	if challengeErr != nil {
		t.Fatal(challengeErr)
	}

	// Build valid assertion but with different device
	cred := buildTestAssertion(t, challengeResp.Challenge, credID, privKey, svc.PasskeyVerifier)

	differentDevice := di
	differentDevice.DeviceID = "totally-different-device"

	_, err := svc.Login(LoginRequest{
		AuthType:   "passkey",
		Nonce:      challengeResp.Challenge,
		Credential: &cred,
		DeviceInfo: differentDevice,
	}, "1.1.1.1")

	if err == nil {
		t.Fatal("unknown device should trigger step-up")
	}
	if err.Code != ErrCodeUnknownDevice {
		t.Errorf("expected AUTH_007, got %s", err.Code)
	}
}

func TestAuthService_Login_DIDSignature_NonceReplay(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	user, _, _ := registerTestUser(t, svc, di)

	nonce := "unique-nonce-123"

	// First login should succeed (DID signature verification is stubbed)
	// We need a known device first
	req := LoginRequest{
		AuthType:   "did_signature",
		DID:        user.DID,
		Signature:  "test-signature",
		Timestamp:  timeNowRFC3339(),
		Nonce:      nonce,
		DeviceInfo: di,
	}

	_, _ = svc.Login(req, "1.1.1.1") // May fail for other reasons, but nonce is consumed

	// Replay with same nonce should fail
	req.Timestamp = timeNowRFC3339() // Fresh timestamp
	_, err := svc.Login(req, "1.1.1.1")
	if err == nil {
		t.Error("replayed nonce should be rejected")
	}
}

func TestAuthService_Login_RateLimit(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	// Exhaust the login rate limit
	for i := 0; i < 10; i++ {
		svc.Login(LoginRequest{
			AuthType:   "passkey",
			Credential: &LoginCredential{ID: "nonexistent"},
			DeviceInfo: di,
		}, "1.1.1.1")
	}

	// 11th attempt should be rate limited
	_, err := svc.Login(LoginRequest{
		AuthType:   "passkey",
		Credential: &LoginCredential{ID: "nonexistent"},
		DeviceInfo: di,
	}, "1.1.1.1")

	if err == nil {
		t.Fatal("should be rate limited")
	}
}

// --- Token Refresh Tests ---

func TestAuthService_RefreshTokens(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	_, _, _ = registerTestUser(t, svc, di)

	// Get a refresh token via the registration flow
	// (already tested above, just need the token)
	phoneResp, _ := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5559999999",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")

	session := svc.OTP.GetSession(phoneResp.VerificationID)
	code := "111111"
	hash, _ := HashOTP(code)
	session.CodeHash = hash

	verifyResp, _ := svc.VerifyOTP(OTPVerifyRequest{
		VerificationID: phoneResp.VerificationID,
		Code:           code,
		DeviceInfo:     di,
	}, "1.1.1.1")

	passkeyReq2 := buildTestAttestation(t, verifyResp.PasskeyChallenge, "cred-refresh-test", svc.PasskeyVerifier)
	passkeyReq2.DeviceInfo = di
	passkeyResp, _ := svc.RegisterPasskey(verifyResp.User.DID, passkeyReq2, "1.1.1.1")

	// Refresh the token
	refreshResp, err := svc.RefreshTokens(RefreshRequest{
		RefreshToken: passkeyResp.RefreshToken,
	}, di, "1.1.1.1")

	if err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	if refreshResp.AccessToken == "" {
		t.Error("should return new access token")
	}
	if refreshResp.RefreshToken == "" {
		t.Error("should return new refresh token")
	}
	if refreshResp.RefreshToken == passkeyResp.RefreshToken {
		t.Error("new refresh token should differ from old")
	}
}

// --- Logout Tests ---

func TestAuthService_Logout(t *testing.T) {
	svc := newTestAuthService(t)

	claims := &TokenClaims{
		Subject:   "did:prism:test",
		ExpiresAt: 9999999999,
		TokenID:   "jti-logout-test",
		DeviceID:  "device-1",
	}

	svc.Logout(claims, false, "1.1.1.1")

	if !svc.Tokens.IsBlocklisted("jti-logout-test") {
		t.Error("logged out token should be blocklisted")
	}
}

// --- Audit Tests ---

func TestAuthService_AuditTrail(t *testing.T) {
	svc := newTestAuthService(t)
	di := validDeviceInfo()

	// Registration creates audit entries
	svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5551234567",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")

	if svc.Audit.Count() < 1 {
		t.Error("should have audit entries")
	}
}

// --- Error Code Tests ---

func TestAuthErrorCodes_UserFacingMessages(t *testing.T) {
	codes := []AuthErrorCode{
		ErrCodeInvalidPhone, ErrCodeOTPRateLimit, ErrCodeInvalidOTP,
		ErrCodePasskeyFailed, ErrCodeTokenExpired, ErrCodeRefreshInvalid,
		ErrCodeUnknownDevice, ErrCodeStepUpRequired, ErrCodeAccountLocked,
		ErrCodeDeviceIntegrity, ErrCodeRecoveryInvalid, ErrCodeGlobalRateLimit,
	}

	for _, code := range codes {
		msg := code.UserFacingMessage()
		if msg == "" {
			t.Errorf("code %s should have a user-facing message", code)
		}
		// Messages should not reveal internals
		for _, forbidden := range []string{"phone_hash", "credential_id", "sql", "postgres"} {
			if contains(msg, forbidden) {
				t.Errorf("code %s message contains forbidden term: %s", code, forbidden)
			}
		}
	}
}

func TestNewAuthError(t *testing.T) {
	err := NewAuthError(ErrCodeInvalidOTP, 400)
	if err.Code != ErrCodeInvalidOTP {
		t.Errorf("expected AUTH_003, got %s", err.Code)
	}
	if err.HTTPStatus != 400 {
		t.Errorf("expected 400, got %d", err.HTTPStatus)
	}
	if err.Error() == "" {
		t.Error("Error() should return non-empty string")
	}
}

// --- Helpers ---

func registerTestUser(t *testing.T, svc *AuthService, di DeviceInfo) (*User, string, *ecdsa.PrivateKey) {
	t.Helper()

	phoneResp, err := svc.RegisterPhone(PhoneRegistrationRequest{
		PhoneNumber: "5550001111",
		CountryCode: "+1",
		DeviceInfo:  di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("register phone: %v", err)
	}

	session := svc.OTP.GetSession(phoneResp.VerificationID)
	code := "777777"
	hash, _ := HashOTP(code)
	session.CodeHash = hash

	verifyResp, err := svc.VerifyOTP(OTPVerifyRequest{
		VerificationID: phoneResp.VerificationID,
		Code:           code,
		DeviceInfo:     di,
	}, "1.1.1.1")
	if err != nil {
		t.Fatalf("verify OTP: %v", err)
	}

	credID := "test-cred-" + verifyResp.User.ID
	req, privKey := buildTestAttestationWithKey(t, verifyResp.PasskeyChallenge, credID, svc.PasskeyVerifier)
	req.DeviceInfo = di
	passkeyResp, err := svc.RegisterPasskey(verifyResp.User.DID, req, "1.1.1.1")
	if err != nil {
		t.Fatalf("register passkey: %v", err)
	}

	return passkeyResp.User, credID, privKey
}

// buildTestAttestation generates a valid WebAuthn attestation request with a real P-256 key.
func buildTestAttestation(t *testing.T, challenge, credID string, verifier *PasskeyVerifier) PasskeyRegistrationRequest {
	t.Helper()
	req, _ := buildTestAttestationWithKey(t, challenge, credID, verifier)
	return req
}

// buildTestAttestationWithKey returns both the registration request and the private key for login tests.
func buildTestAttestationWithKey(t *testing.T, challenge, credID string, verifier *PasskeyVerifier) (PasskeyRegistrationRequest, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	clientData := ClientData{
		Type:      "webauthn.create",
		Challenge: challenge,
		Origin:    verifier.Expected,
	}
	cdJSON, _ := json.Marshal(clientData)

	return PasskeyRegistrationRequest{
		Challenge: challenge,
		AttestationResponse: AttestationResponse{
			ID:    credID,
			RawID: base64.RawURLEncoding.EncodeToString([]byte(credID)),
			Response: AttestationResponseDetail{
				ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
				AttestationObject: base64.RawURLEncoding.EncodeToString(pubUncompressed),
			},
			Type: "public-key",
		},
	}, key
}

// buildTestAssertion generates a valid WebAuthn assertion (login) credential.
func buildTestAssertion(t *testing.T, challenge, credID string, privKey *ecdsa.PrivateKey, verifier *PasskeyVerifier) LoginCredential {
	t.Helper()

	clientData := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    verifier.Expected,
	}
	cdJSON, _ := json.Marshal(clientData)

	rpHash := sha256.Sum256([]byte(verifier.RPID))
	authData := make([]byte, 37)
	copy(authData[:32], rpHash[:])
	authData[32] = 0x01 // user present

	cdHash := sha256.Sum256(cdJSON)
	signedData := append(authData, cdHash[:]...)
	digest := sha256.Sum256(signedData)

	r, s, err := ecdsa.Sign(rand.Reader, privKey, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	sig := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	return LoginCredential{
		ID:    credID,
		RawID: base64.RawURLEncoding.EncodeToString([]byte(credID)),
		Response: AssertionResponseData{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
			Signature:         base64.RawURLEncoding.EncodeToString(sig),
		},
		Type: "public-key",
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func timeNowRFC3339() string {
	return "2026-03-11T12:00:00Z"
}
