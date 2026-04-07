import XCTest
import XCTest
@testable import Echo

// MARK: - AuthCoordinatorVM Tests

@MainActor
final class AuthCoordinatorVMTests: XCTestCase {

    var coordinator: AuthCoordinatorVM!
    var mockKeychain: MockKeychainManager!
    var mockAPIClient: MockAuthAPIClient!
    var mockPasskeyManager: MockPasskeyManager!
    var tokenManager: TokenManager!
    var biometricService: BiometricIntegrityService!

    override func setUp() {
        super.setUp()
        mockKeychain = MockKeychainManager()
        mockAPIClient = MockAuthAPIClient()
        mockPasskeyManager = MockPasskeyManager()
        tokenManager = TokenManager(keychain: mockKeychain, apiClient: mockAPIClient)
        biometricService = BiometricIntegrityService(keychain: mockKeychain)
        coordinator = AuthCoordinatorVM(
            tokenManager: tokenManager,
            passkeyManager: mockPasskeyManager,
            biometricService: biometricService,
            apiClient: mockAPIClient
        )
    }

    func testInitialState() {
        XCTAssertEqual(coordinator.state, .unauthenticated)
    }

    func testPhoneSubmittedTransition() {
        coordinator.handle(.phoneSubmitted(phone: "+15551234567", verificationId: "vid-1"))
        XCTAssertEqual(coordinator.state, .otpVerification(phone: "+15551234567", verificationId: "vid-1"))
    }

    func testOTPVerifiedTransition() {
        coordinator.handle(.phoneSubmitted(phone: "+15551234567", verificationId: "vid-1"))
        coordinator.handle(.otpVerified(tempToken: "temp-token"))
        XCTAssertEqual(coordinator.state, .passkeySetup(tempToken: "temp-token"))
    }

    func testPasskeyCreatedTransition() {
        coordinator.handle(.phoneSubmitted(phone: "+1", verificationId: "v"))
        coordinator.handle(.otpVerified(tempToken: "t"))
        coordinator.handle(.passkeyCreated)
        XCTAssertEqual(coordinator.state, .profileSetup)
    }

    func testProfileCompletedTransition() {
        coordinator.handle(.phoneSubmitted(phone: "+1", verificationId: "v"))
        coordinator.handle(.otpVerified(tempToken: "t"))
        coordinator.handle(.passkeyCreated)
        coordinator.handle(.profileCompleted)
        XCTAssertEqual(coordinator.state, .trustIntro)
    }

    func testLoginSucceededFromAnyState() {
        let user = UserProfile(id: "u1", did: "did:dag:u1", displayName: "Test", username: "test", trustScore: 50, trustTier: 2)
        coordinator.handle(.loginSucceeded(user))
        XCTAssertEqual(coordinator.state, .authenticated(user: user))
    }

    func testSessionExpiredClearsTokens() {
        let user = UserProfile(id: "u1", did: "did:dag:u1", displayName: "Test", username: "test", trustScore: 50, trustTier: 2)
        coordinator.handle(.loginSucceeded(user))

        // Store a refresh token first
        try? mockKeychain.save(Data("rt".utf8), for: TokenManager.refreshTokenKey)

        coordinator.handle(.sessionExpired)
        XCTAssertEqual(coordinator.state, .unauthenticated)
        XCTAssertNil(try? mockKeychain.load(for: TokenManager.refreshTokenKey))
    }

    func testLoggedOutClearsTokens() {
        let user = UserProfile(id: "u1", did: "did:dag:u1", displayName: "Test", username: "test", trustScore: 50, trustTier: 2)
        coordinator.handle(.loginSucceeded(user))
        coordinator.handle(.loggedOut)
        XCTAssertEqual(coordinator.state, .unauthenticated)
    }

    func testAccountLockedTransition() {
        let retryAfter = Date().addingTimeInterval(3600)
        coordinator.handle(.accountLocked(.tooManyAttempts, retryAfter: retryAfter))
        if case .locked(let reason, _) = coordinator.state {
            XCTAssertEqual(reason, .tooManyAttempts)
        } else {
            XCTFail("Expected locked state")
        }
    }

    func testRecoveryInitiatedTransition() {
        coordinator.handle(.recoveryInitiated(.recoveryPhrase))
        XCTAssertEqual(coordinator.state, .recovery(method: .recoveryPhrase))
    }

    func testRecoveryCompletedTransition() {
        coordinator.handle(.recoveryInitiated(.trustedContacts))
        coordinator.handle(.recoveryCompleted)
        XCTAssertEqual(coordinator.state, .passkeySetup(tempToken: ""))
    }

    func testInvalidTransitionIgnored() {
        // Can't go from unauthenticated to passkeyCreated
        coordinator.handle(.passkeyCreated)
        XCTAssertEqual(coordinator.state, .unauthenticated)
    }

    func testStepUpRequestShowsSheet() {
        var receivedToken: String?
        coordinator.requestStepUp(for: .revokeDevice) { token in
            receivedToken = token
        }
        XCTAssertTrue(coordinator.showStepUpSheet)
        XCTAssertEqual(coordinator.stepUpAction, .revokeDevice)

        coordinator.completeStepUp(elevatedToken: "elevated-123")
        XCTAssertFalse(coordinator.showStepUpSheet)
        XCTAssertEqual(receivedToken, "elevated-123")
        XCTAssertNil(coordinator.stepUpAction)
    }
}

// MARK: - TokenManager Tests

@MainActor
final class TokenManagerTests: XCTestCase {

    var tokenManager: TokenManager!
    var mockKeychain: MockKeychainManager!
    var mockAPIClient: MockAuthAPIClient!

    override func setUp() {
        super.setUp()
        mockKeychain = MockKeychainManager()
        mockAPIClient = MockAuthAPIClient()
        tokenManager = TokenManager(keychain: mockKeychain, apiClient: mockAPIClient)
    }

    func testStoreTokens() throws {
        let response = AuthTokenResponse(
            accessToken: "access-123",
            refreshToken: "refresh-456",
            expiresAt: Date().addingTimeInterval(900),
            user: nil,
            passkeyChallenge: nil
        )

        try tokenManager.storeTokens(response)

        XCTAssertTrue(tokenManager.isAuthenticated)
        XCTAssertEqual(tokenManager.currentAccessToken, "access-123")

        let storedRefresh = try mockKeychain.load(for: TokenManager.refreshTokenKey)
        XCTAssertEqual(String(data: storedRefresh!, encoding: .utf8), "refresh-456")
    }

    func testStoreTokensWithoutRefresh() throws {
        let response = AuthTokenResponse(
            accessToken: "access-123",
            refreshToken: nil,
            expiresAt: Date().addingTimeInterval(900),
            user: nil,
            passkeyChallenge: nil
        )

        try tokenManager.storeTokens(response)
        XCTAssertTrue(tokenManager.isAuthenticated)
        XCTAssertNil(try mockKeychain.load(for: TokenManager.refreshTokenKey))
    }

    func testClearTokens() throws {
        let response = AuthTokenResponse(
            accessToken: "access-123",
            refreshToken: "refresh-456",
            expiresAt: Date().addingTimeInterval(900),
            user: nil,
            passkeyChallenge: nil
        )
        try tokenManager.storeTokens(response)

        tokenManager.clearTokens()

        XCTAssertFalse(tokenManager.isAuthenticated)
        XCTAssertNil(tokenManager.currentAccessToken)
        XCTAssertNil(try mockKeychain.load(for: TokenManager.refreshTokenKey))
    }

    func testGetValidAccessTokenReturnsCurrentIfValid() async throws {
        let response = AuthTokenResponse(
            accessToken: "valid-token",
            refreshToken: "rt",
            expiresAt: Date().addingTimeInterval(900),
            user: nil,
            passkeyChallenge: nil
        )
        try tokenManager.storeTokens(response)

        let token = try await tokenManager.getValidAccessToken()
        XCTAssertEqual(token, "valid-token")
    }

    func testIsAuthenticatedOnInit() {
        // No refresh token stored — should not be authenticated
        XCTAssertFalse(tokenManager.isAuthenticated)

        // Store refresh token and create new manager
        try? mockKeychain.save(Data("rt".utf8), for: TokenManager.refreshTokenKey)
        let newManager = TokenManager(keychain: mockKeychain, apiClient: mockAPIClient)
        XCTAssertTrue(newManager.isAuthenticated)
    }
}

// MARK: - PhoneEntryViewModel Tests

@MainActor
final class PhoneEntryViewModelTests: XCTestCase {

    var viewModel: PhoneEntryViewModel!
    var mockAPIClient: MockAuthAPIClient!

    override func setUp() {
        super.setUp()
        mockAPIClient = MockAuthAPIClient()
        viewModel = PhoneEntryViewModel(
            apiClient: mockAPIClient,
            deviceService: DeviceFingerprintService()
        )
    }

    func testValidPhoneNumber() {
        viewModel.phoneNumber = "5551234567"
        XCTAssertTrue(viewModel.isValid)
    }

    func testInvalidPhoneTooShort() {
        viewModel.phoneNumber = "555"
        XCTAssertFalse(viewModel.isValid)
    }

    func testInvalidPhoneTooLong() {
        viewModel.phoneNumber = "1234567890123456"
        XCTAssertFalse(viewModel.isValid)
    }

    func testFormattedDisplayUS() {
        viewModel.countryCode = "+1"
        viewModel.phoneNumber = "5551234567"
        XCTAssertEqual(viewModel.formattedDisplay, "(555) 123-4567")
    }

    func testFormattedDisplayNonUS() {
        viewModel.countryCode = "+44"
        viewModel.phoneNumber = "5551234567"
        // Non-US numbers aren't formatted
        XCTAssertEqual(viewModel.formattedDisplay, "5551234567")
    }

    func testSendOTPSuccess() async {
        viewModel.phoneNumber = "5551234567"
        mockAPIClient.phoneResponse = PhoneRegistrationResponse(
            verificationId: "vid-abc",
            expiresAt: Date().addingTimeInterval(600),
            retryAfter: 60
        )

        let result = await viewModel.sendOTP()
        XCTAssertNotNil(result)
        XCTAssertEqual(result?.verificationId, "vid-abc")
        XCTAssertTrue(result!.phone.contains("5551234567"))
        XCTAssertNil(viewModel.errorMessage)
    }

    func testSendOTPFailure() async {
        viewModel.phoneNumber = "5551234567"
        mockAPIClient.errorToThrow = AuthAPIError(
            code: "AUTH_001",
            userMessage: "Invalid phone number.",
            httpStatus: 400
        )

        let result = await viewModel.sendOTP()
        XCTAssertNil(result)
        XCTAssertEqual(viewModel.errorMessage, "Invalid phone number.")
    }

    func testSendOTPInvalidReturnsNil() async {
        viewModel.phoneNumber = "123"
        let result = await viewModel.sendOTP()
        XCTAssertNil(result)
    }
}

// MARK: - OTPViewModel Tests

@MainActor
final class OTPViewModelTests: XCTestCase {

    var viewModel: OTPViewModel!
    var mockAPIClient: MockAuthAPIClient!

    override func setUp() {
        super.setUp()
        mockAPIClient = MockAuthAPIClient()
        viewModel = OTPViewModel(
            phoneNumber: "+15551234567",
            verificationId: "vid-1",
            apiClient: mockAPIClient,
            deviceService: DeviceFingerprintService()
        )
    }

    func testInitialState() {
        XCTAssertEqual(viewModel.fullCode, "")
        XCTAssertFalse(viewModel.isComplete)
        XCTAssertEqual(viewModel.resendCountdown, 60)
        XCTAssertFalse(viewModel.canResend)
    }

    func testCodeCompletion() {
        viewModel.code = ["1", "2", "3", "4", "5", "6"]
        XCTAssertTrue(viewModel.isComplete)
        XCTAssertEqual(viewModel.fullCode, "123456")
    }

    func testCodeNotCompleteWithLetters() {
        viewModel.code = ["1", "2", "3", "4", "5", "a"]
        XCTAssertFalse(viewModel.isComplete)
    }

    func testVerifyOTPSuccess() async {
        viewModel.code = ["1", "2", "3", "4", "5", "6"]
        let token = await viewModel.verifyOTP()
        XCTAssertNotNil(token)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testVerifyOTPFailureClearsCode() async {
        viewModel.code = ["1", "2", "3", "4", "5", "6"]
        mockAPIClient.errorToThrow = AuthAPIError(
            code: "AUTH_003",
            userMessage: "Incorrect code.",
            httpStatus: 400
        )

        let token = await viewModel.verifyOTP()
        XCTAssertNil(token)
        XCTAssertEqual(viewModel.errorMessage, "Incorrect code.")
        XCTAssertEqual(viewModel.fullCode, "") // Code cleared
        XCTAssertEqual(viewModel.focusedIndex, 0) // Focus reset
    }

    func testVerifyOTPIncompleteReturnsNil() async {
        viewModel.code = ["1", "2", "3", "", "", ""]
        let token = await viewModel.verifyOTP()
        XCTAssertNil(token)
    }

    func testHandleDigitInputAdvancesFocus() {
        viewModel.handleDigitInput(at: 0, value: "1")
        XCTAssertEqual(viewModel.focusedIndex, 1)

        viewModel.handleDigitInput(at: 4, value: "5")
        XCTAssertEqual(viewModel.focusedIndex, 5)
    }

    func testHandleDigitInputDoesNotAdvancePastLast() {
        viewModel.handleDigitInput(at: 5, value: "6")
        // Should stay at 5, not advance past
        XCTAssertEqual(viewModel.focusedIndex, 5)
    }
}

// MARK: - LoginViewModel Tests

@MainActor
final class LoginViewModelTests: XCTestCase {

    var viewModel: LoginViewModel!
    var mockAPIClient: MockAuthAPIClient!
    var mockPasskeyManager: MockPasskeyManager!
    var mockKeychain: MockKeychainManager!
    var tokenManager: TokenManager!

    override func setUp() {
        super.setUp()
        mockAPIClient = MockAuthAPIClient()
        mockPasskeyManager = MockPasskeyManager()
        mockKeychain = MockKeychainManager()
        tokenManager = TokenManager(keychain: mockKeychain, apiClient: mockAPIClient)
        viewModel = LoginViewModel(
            passkeyManager: mockPasskeyManager,
            tokenManager: tokenManager,
            apiClient: mockAPIClient,
            deviceService: DeviceFingerprintService()
        )
    }

    func testInitialState() {
        XCTAssertFalse(viewModel.isAuthenticating)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testLoadStoredAccountInfo() {
        UserDefaults.standard.set("testuser", forKey: "echo.display.username")
        UserDefaults.standard.set("***-1234", forKey: "echo.display.masked_phone")

        viewModel.loadStoredAccountInfo()

        XCTAssertEqual(viewModel.username, "testuser")
        XCTAssertEqual(viewModel.maskedPhone, "***-1234")

        // Cleanup
        UserDefaults.standard.removeObject(forKey: "echo.display.username")
        UserDefaults.standard.removeObject(forKey: "echo.display.masked_phone")
    }

    func testLoginWithPasskeySuccess() async {
        let expectedUser = UserProfile(
            id: "user-1", did: "did:dag:u1", displayName: "Test",
            username: "testuser", trustScore: 50, trustTier: 2
        )
        mockAPIClient.tokenResponse = AuthTokenResponse(
            accessToken: "at", refreshToken: "rt",
            expiresAt: Date().addingTimeInterval(900),
            user: expectedUser, passkeyChallenge: nil
        )

        let user = await viewModel.loginWithPasskey()
        XCTAssertNotNil(user)
        XCTAssertEqual(user?.id, "user-1")
        XCTAssertNil(viewModel.errorMessage)
        XCTAssertTrue(tokenManager.isAuthenticated)
    }

    func testLoginWithPasskeyNewDevice() async {
        mockAPIClient.errorToThrow = AuthAPIError(
            code: "AUTH_007",
            userMessage: "New device detected.",
            httpStatus: 403
        )

        let user = await viewModel.loginWithPasskey()
        XCTAssertNil(user)
        XCTAssertEqual(viewModel.errorMessage, "New device detected. Additional verification needed.")
    }

    func testLoginWithPasskeyAccountLocked() async {
        mockAPIClient.errorToThrow = AuthAPIError(
            code: "AUTH_009",
            userMessage: "Account locked.",
            httpStatus: 429
        )

        let user = await viewModel.loginWithPasskey()
        XCTAssertNil(user)
        XCTAssertEqual(viewModel.errorMessage, "Account temporarily locked. Please try again later.")
    }
}

// MARK: - DeviceManagementViewModel Tests

@MainActor
final class DeviceManagementViewModelTests: XCTestCase {

    var viewModel: DeviceManagementViewModel!
    var mockAPIClient: MockAuthAPIClient!
    var mockKeychain: MockKeychainManager!
    var tokenManager: TokenManager!

    override func setUp() {
        super.setUp()
        mockAPIClient = MockAuthAPIClient()
        mockKeychain = MockKeychainManager()
        tokenManager = TokenManager(keychain: mockKeychain, apiClient: mockAPIClient)

        // Store tokens so getValidAccessToken works
        try? tokenManager.storeTokens(AuthTokenResponse(
            accessToken: "at", refreshToken: "rt",
            expiresAt: Date().addingTimeInterval(900),
            user: nil, passkeyChallenge: nil
        ))

        viewModel = DeviceManagementViewModel(
            apiClient: mockAPIClient,
            tokenManager: tokenManager
        )
    }

    func testLoadDevices() async {
        mockAPIClient.devices = [
            DeviceSession(
                id: "d1", friendlyName: "iPhone 15 Pro",
                platform: "ios", osVersion: "17.4",
                lastIP: "1.2.3.4", lastLocation: "San Francisco",
                lastActiveAt: Date(), isCurrentDevice: true
            ),
            DeviceSession(
                id: "d2", friendlyName: "iPad Air",
                platform: "ios", osVersion: "17.3",
                lastIP: "5.6.7.8", lastLocation: "New York",
                lastActiveAt: Date(), isCurrentDevice: false
            )
        ]

        await viewModel.loadDevices()

        XCTAssertFalse(viewModel.isLoading)
        XCTAssertNotNil(viewModel.currentDevice)
        XCTAssertEqual(viewModel.currentDevice?.id, "d1")
        XCTAssertEqual(viewModel.otherDevices.count, 1)
        XCTAssertEqual(viewModel.otherDevices.first?.id, "d2")
    }

    func testRevokeDevice() async {
        mockAPIClient.devices = [
            DeviceSession(
                id: "d2", friendlyName: "iPad",
                platform: "ios", osVersion: "17.3",
                lastIP: "5.6.7.8", lastLocation: nil,
                lastActiveAt: Date(), isCurrentDevice: false
            )
        ]
        await viewModel.loadDevices()

        let success = await viewModel.revokeDevice(id: "d2", elevatedToken: "elevated-123")
        XCTAssertTrue(success)
        XCTAssertTrue(viewModel.otherDevices.isEmpty)
    }

    func testLoadDevicesError() async {
        mockAPIClient.errorToThrow = AuthAPIError(
            code: "NETWORK", userMessage: "No connection", httpStatus: 0
        )

        await viewModel.loadDevices()
        XCTAssertNotNil(viewModel.errorMessage)
    }
}

// MARK: - AuthState Tests

final class AuthStateTests: XCTestCase {

    func testAuthStateEquality() {
        XCTAssertEqual(AuthState.unauthenticated, .unauthenticated)
        XCTAssertEqual(AuthState.phoneEntry, .phoneEntry)
        XCTAssertEqual(
            AuthState.otpVerification(phone: "+1", verificationId: "v"),
            AuthState.otpVerification(phone: "+1", verificationId: "v")
        )
        XCTAssertNotEqual(
            AuthState.otpVerification(phone: "+1", verificationId: "v1"),
            AuthState.otpVerification(phone: "+1", verificationId: "v2")
        )
        XCTAssertEqual(AuthState.profileSetup, .profileSetup)
        XCTAssertEqual(AuthState.trustIntro, .trustIntro)
        XCTAssertNotEqual(AuthState.unauthenticated, .phoneEntry)
    }

    func testLockReasonEquality() {
        XCTAssertEqual(LockReason.tooManyAttempts, .tooManyAttempts)
        XCTAssertNotEqual(LockReason.tooManyAttempts, .suspiciousActivity)
    }

    func testRecoveryMethodCases() {
        XCTAssertEqual(RecoveryMethod.allCases.count, 3)
        XCTAssertEqual(RecoveryMethod.recoveryPhrase.rawValue, "recovery_phrase")
        XCTAssertEqual(RecoveryMethod.trustedContacts.rawValue, "trusted_contacts")
        XCTAssertEqual(RecoveryMethod.phoneReverification.rawValue, "phone")
    }

    func testStepUpActionProperties() {
        XCTAssertEqual(StepUpAction.revokeDevice.rawValue, "revoke_device")
        XCTAssertFalse(StepUpAction.revokeDevice.requiresPasskeyAndOTP)
        XCTAssertTrue(StepUpAction.exportRecoveryPhrase.requiresPasskeyAndOTP)
        XCTAssertTrue(StepUpAction.deleteAccount.requiresPasskeyAndOTP)
        XCTAssertFalse(StepUpAction.changePhone.requiresPasskeyAndOTP)
    }
}

// MARK: - AuthError Tests

final class AuthErrorTests: XCTestCase {

    func testErrorDescriptions() {
        XCTAssertNotNil(AuthError.passkeyCreationFailed.errorDescription)
        XCTAssertNotNil(AuthError.passkeyAssertionFailed.errorDescription)
        XCTAssertNotNil(AuthError.noRefreshToken.errorDescription)
        XCTAssertNotNil(AuthError.tokenStorageFailed.errorDescription)
        XCTAssertNotNil(AuthError.sessionRevoked.errorDescription)
        XCTAssertNotNil(AuthError.biometricChanged.errorDescription)
        XCTAssertNotNil(AuthError.deviceIntegrityFailed.errorDescription)
        XCTAssertNotNil(AuthError.networkUnavailable.errorDescription)
        XCTAssertNotNil(AuthError.invalidPhoneNumber.errorDescription)
        XCTAssertNotNil(AuthError.otpExpired.errorDescription)
        XCTAssertNotNil(AuthError.accountLocked(retryAfter: nil).errorDescription)
    }

    func testIsRetryable() {
        XCTAssertTrue(AuthError.networkUnavailable.isRetryable)
        XCTAssertTrue(AuthError.passkeyAssertionFailed.isRetryable)
        XCTAssertFalse(AuthError.sessionRevoked.isRetryable)
        XCTAssertFalse(AuthError.noRefreshToken.isRetryable)
        XCTAssertFalse(AuthError.deviceIntegrityFailed.isRetryable)
    }
}

// MARK: - AuthToken Model Tests

final class AuthTokenModelTests: XCTestCase {

    func testAuthTokenResponseDecodable() throws {
        let json = """
        {
            "access_token": "at-123",
            "refresh_token": "rt-456",
            "expires_at": "2026-03-11T12:00:00Z",
            "user": {
                "id": "u1",
                "did": "did:dag:u1",
                "display_name": "Test",
                "username": "test",
                "trust_score": 50,
                "trust_tier": 2
            }
        }
        """
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let response = try decoder.decode(AuthTokenResponse.self, from: Data(json.utf8))

        XCTAssertEqual(response.accessToken, "at-123")
        XCTAssertEqual(response.refreshToken, "rt-456")
        XCTAssertNotNil(response.user)
        XCTAssertEqual(response.user?.id, "u1")
    }

    func testChallengeResponseData() {
        let challenge = ChallengeResponse(
            challenge: Data("test-challenge".utf8).base64EncodedString(),
            timeout: 60,
            rpId: "echo.app"
        )
        XCTAssertEqual(challenge.challengeData, Data("test-challenge".utf8))
    }

    func testChallengeResponseInvalidBase64() {
        let challenge = ChallengeResponse(
            challenge: "not-valid-base64!!!",
            timeout: 60,
            rpId: "echo.app"
        )
        XCTAssertEqual(challenge.challengeData, Data())
    }

    func testDeviceSessionDecodable() throws {
        let json = """
        {
            "id": "d1",
            "friendly_name": "iPhone 15",
            "platform": "ios",
            "os_version": "17.4",
            "last_ip": "1.2.3.4",
            "last_active_at": "2026-03-11T12:00:00Z",
            "is_current_device": true
        }
        """
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let device = try decoder.decode(DeviceSession.self, from: Data(json.utf8))

        XCTAssertEqual(device.id, "d1")
        XCTAssertEqual(device.friendlyName, "iPhone 15")
        XCTAssertTrue(device.isCurrentDevice)
        XCTAssertNil(device.lastLocation)
    }

    func testAuditLogEntryProperties() throws {
        let json = """
        {
            "id": "a1",
            "event_type": "login",
            "result": "success",
            "ip_address": "1.2.3.4",
            "created_at": "2026-03-11T12:00:00Z"
        }
        """
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        let entry = try decoder.decode(AuthAuditLogEntry.self, from: Data(json.utf8))

        XCTAssertEqual(entry.id, "a1")
        XCTAssertTrue(entry.isSuccess)
        XCTAssertFalse(entry.isFailed)
        XCTAssertFalse(entry.isBlocked)
    }
}

// MARK: - MockKeychainManager Tests

final class MockKeychainManagerTests: XCTestCase {

    func testSaveAndLoad() throws {
        let keychain = MockKeychainManager()
        let data = Data("test-value".utf8)

        try keychain.save(data, for: "test-key")
        let loaded = try keychain.load(for: "test-key")

        XCTAssertEqual(loaded, data)
    }

    func testDelete() throws {
        let keychain = MockKeychainManager()
        try keychain.save(Data("val".utf8), for: "key")
        try keychain.delete(for: "key")

        XCTAssertNil(try keychain.load(for: "key"))
    }

    func testLoadNonexistent() throws {
        let keychain = MockKeychainManager()
        XCTAssertNil(try keychain.load(for: "missing"))
    }
}

// MARK: - UserProfile Tests

final class UserProfileTests: XCTestCase {

    func testCodableRoundTrip() throws {
        let profile = UserProfile(
            id: "u1", did: "did:dag:u1", displayName: "Test User",
            username: "testuser", trustScore: 75, trustTier: 3
        )

        let encoder = JSONEncoder()
        let data = try encoder.encode(profile)
        let decoded = try JSONDecoder().decode(UserProfile.self, from: data)

        XCTAssertEqual(decoded.id, "u1")
        XCTAssertEqual(decoded.did, "did:dag:u1")
        XCTAssertEqual(decoded.displayName, "Test User")
        XCTAssertEqual(decoded.username, "testuser")
        XCTAssertEqual(decoded.trustScore, 75)
        XCTAssertEqual(decoded.trustTier, 3)
    }

    func testEquality() {
        let user1 = UserProfile(id: "u1", did: "did:dag:u1", displayName: "A", username: "a", trustScore: 1, trustTier: 1)
        let user2 = UserProfile(id: "u1", did: "did:dag:u1", displayName: "B", username: "b", trustScore: 2, trustTier: 2)
        let user3 = UserProfile(id: "u2", did: "did:dag:u2", displayName: "A", username: "a", trustScore: 1, trustTier: 1)

        // Equality based on ID only
        XCTAssertEqual(user1, user2)
        XCTAssertNotEqual(user1, user3)
    }
}
