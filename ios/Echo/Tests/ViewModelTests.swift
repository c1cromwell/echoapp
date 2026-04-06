import XCTest
@testable import Echo

// MARK: - AuthViewModel Tests

final class AuthViewModelTests: XCTestCase {

    var mockAuth: MockAuthService!
    var mockKeychain: MockKeychainManager!
    var vm: AuthViewModel!

    override func setUp() {
        super.setUp()
        mockAuth = MockAuthService()
        mockKeychain = MockKeychainManager()
        vm = AuthViewModel(authService: mockAuth, keychainManager: mockKeychain)
    }

    // MARK: - OTP Request

    @MainActor
    func testRequestOTP_success_transitionsToOTPVerification() async throws {
        vm.requestOTP(phone: "+15550100")

        // Allow Task to complete
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.authState, .otpVerification)
        XCTAssertEqual(vm.phone, "+15550100")
        XCTAssertFalse(vm.isLoading)
        XCTAssertNil(vm.errorMessage)
        XCTAssertEqual(mockAuth.requestOTPCallCount, 1)
        XCTAssertEqual(mockAuth.lastRequestedPhone, "+15550100")
    }

    @MainActor
    func testRequestOTP_failure_setsError() async throws {
        mockAuth.requestOTPResult = .failure(NSError(domain: "", code: -1, userInfo: [NSLocalizedDescriptionKey: "Network error"]))
        vm.requestOTP(phone: "+15550100")

        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.authState, .welcome) // should not transition
        XCTAssertNotNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
    }

    // MARK: - OTP Verification

    @MainActor
    func testVerifyOTP_success_savesTokenAndTransitions() async throws {
        vm.phone = "+15550100"
        vm.verifyOTP("123456")

        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.authState, .passkeySetup)
        XCTAssertEqual(vm.otp, "123456")
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockKeychain.storedToken, "mock-token-123")
        XCTAssertEqual(mockKeychain.saveTokenCallCount, 1)
        XCTAssertEqual(mockAuth.verifyOTPCallCount, 1)
    }

    @MainActor
    func testVerifyOTP_failure_setsError() async throws {
        mockAuth.verifyOTPResult = .failure(NSError(domain: "", code: -1, userInfo: [NSLocalizedDescriptionKey: "Invalid code"]))
        vm.verifyOTP("000000")

        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertNotNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockKeychain.saveTokenCallCount, 0) // token should NOT be saved
    }

    // MARK: - Passkey Setup

    @MainActor
    func testSetupPasskey_success_transitionsToProfileSetup() async throws {
        vm.setupPasskey()

        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.authState, .profileSetup)
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockAuth.registerPasskeyCallCount, 1)
    }

    // MARK: - Sign Out

    @MainActor
    func testSignOut_clearsState() {
        vm.isAuthenticated = true
        vm.phone = "+15550100"
        vm.otp = "123456"
        mockKeychain.storedToken = "some-token"

        vm.signOut()

        XCTAssertFalse(vm.isAuthenticated)
        XCTAssertEqual(vm.authState, .welcome)
        XCTAssertEqual(vm.phone, "")
        XCTAssertEqual(vm.otp, "")
        XCTAssertEqual(mockKeychain.clearAllCallCount, 1)
        XCTAssertNil(mockKeychain.storedToken)
    }
}

// MARK: - MessagingViewModel Tests

final class MessagingViewModelTests: XCTestCase {

    var mockService: MockMessagingService!
    var vm: MessagingViewModel!

    override func setUp() {
        super.setUp()
        mockService = MockMessagingService()
        vm = MessagingViewModel(messagingService: mockService)
    }

    @MainActor
    func testFetchConversations_populatesList() async throws {
        mockService.conversations = [
            ConversationModel(
                id: "conv-1",
                participantId: "user-2",
                participantName: "Alice",
                lastMessage: "Hey!",
                unreadCount: 1,
                updatedAt: Date()
            ),
            ConversationModel(
                id: "conv-2",
                participantId: "user-3",
                participantName: "Bob",
                lastMessage: nil,
                unreadCount: 0,
                updatedAt: Date()
            ),
        ]

        vm.fetchConversations()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.conversations.count, 2)
        XCTAssertEqual(vm.conversations[0].participantName, "Alice")
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockService.fetchConversationsCallCount, 1)
    }

    @MainActor
    func testSelectConversation_setsCurrent() async throws {
        let conv = ConversationModel(
            id: "conv-1",
            participantId: "user-2",
            participantName: "Alice",
            lastMessage: "Hey!",
            unreadCount: 1,
            updatedAt: Date()
        )

        mockService.messages["conv-1"] = [
            MessageModel(id: "msg-1", conversationId: "conv-1", senderId: "user-2", content: "Hey!", status: .delivered, createdAt: Date()),
        ]

        vm.selectConversation(conv)
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.selectedConversation?.id, "conv-1")
        XCTAssertEqual(vm.messages.count, 1)
        XCTAssertEqual(vm.messages[0].content, "Hey!")
    }

    @MainActor
    func testSendMessage_callsServiceAndRefreshes() async throws {
        mockService.messages["conv-1"] = []

        vm.sendMessage("Hello Echo!", to: "conv-1")
        try await Task.sleep(nanoseconds: 200_000_000)

        XCTAssertEqual(mockService.sendMessageCallCount, 1)
        XCTAssertEqual(mockService.lastSentContent, "Hello Echo!")
        XCTAssertEqual(mockService.lastSentConversationId, "conv-1")
        // After send, it refetches messages
        XCTAssertGreaterThanOrEqual(mockService.fetchMessagesCallCount, 1)
    }

    @MainActor
    func testSendMessage_failure_setsError() async throws {
        mockService.sendMessageResult = .failure(NSError(domain: "", code: -1, userInfo: [NSLocalizedDescriptionKey: "Send failed"]))

        vm.sendMessage("Hello", to: "conv-1")
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertNotNil(vm.errorMessage)
    }
}

// MARK: - ProfileViewModel Tests

final class ProfileViewModelTests: XCTestCase {

    var mockService: MockProfileService!
    var vm: ProfileViewModel!

    override func setUp() {
        super.setUp()
        mockService = MockProfileService()
        vm = ProfileViewModel(profileService: mockService)
    }

    @MainActor
    func testFetchProfile_populatesData() async throws {
        vm.fetchProfile()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.profile.displayName, "Test User")
        XCTAssertEqual(vm.profile.username, "testuser")
        XCTAssertEqual(vm.profile.trustScore, 75)
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockService.fetchProfileCallCount, 1)
    }

    @MainActor
    func testFetchProfile_failure_setsError() async throws {
        mockService.shouldFail = true
        vm.fetchProfile()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertNotNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
    }

    @MainActor
    func testBeginEditProfile_copiesCurrentValues() {
        vm.profile = ProfileData(
            displayName: "Alice",
            username: "alice",
            bio: "Hi there",
            status: "Available",
            website: "https://alice.dev"
        )

        vm.beginEditProfile()

        XCTAssertEqual(vm.editDisplayName, "Alice")
        XCTAssertEqual(vm.editUsername, "alice")
        XCTAssertEqual(vm.editBio, "Hi there")
        XCTAssertEqual(vm.editWebsite, "https://alice.dev")
    }

    @MainActor
    func testSaveProfile_updatesProfileData() async throws {
        vm.profile = ProfileData(displayName: "Old Name", username: "old")
        vm.beginEditProfile()
        vm.editDisplayName = "New Name"
        vm.editUsername = "newuser"

        vm.saveProfile()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.profile.displayName, "New Name")
        XCTAssertEqual(vm.profile.username, "newuser")
        XCTAssertFalse(vm.isSaving)
        XCTAssertEqual(vm.successMessage, "Profile updated")
        XCTAssertEqual(mockService.updateProfileCallCount, 1)
    }

    @MainActor
    func testCheckUsernameAvailability_available() async throws {
        vm.profile = ProfileData(username: "oldname")
        vm.editUsername = "newname"
        mockService.usernameAvailable = true

        vm.checkUsernameAvailability()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.isUsernameAvailable, true)
        XCTAssertFalse(vm.isCheckingUsername)
        XCTAssertEqual(mockService.lastCheckedUsername, "newname")
    }

    @MainActor
    func testCheckUsernameAvailability_skipsIfSameAsCurrentUsername() {
        vm.profile = ProfileData(username: "myname")
        vm.editUsername = "myname"

        vm.checkUsernameAvailability()

        // Should not call service — username unchanged
        XCTAssertEqual(mockService.checkUsernameCallCount, 0)
        XCTAssertNil(vm.isUsernameAvailable)
    }
}

// MARK: - RewardsViewModel Tests

final class RewardsViewModelTests: XCTestCase {

    var mockService: MockRewardsService!
    var vm: RewardsViewModel!

    override func setUp() {
        super.setUp()
        mockService = MockRewardsService()
        vm = RewardsViewModel(rewardsService: mockService)
    }

    @MainActor
    func testFetchBalance_setsTokenBalance() async throws {
        mockService.balance = 500.0

        vm.fetchBalance()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.tokenBalance, 500.0)
        XCTAssertEqual(mockService.fetchBalanceCallCount, 1)
    }

    @MainActor
    func testFetchActivity_populatesActivities() async throws {
        vm.fetchActivity()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.activities.count, 2)
        XCTAssertEqual(vm.activities[0].type, "message_sent")
        XCTAssertEqual(mockService.fetchActivityCallCount, 1)
    }

    @MainActor
    func testClaimRewards_updatesBalance() async throws {
        mockService.claimResult = 300.0

        vm.claimRewards()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.tokenBalance, 300.0)
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockService.claimRewardsCallCount, 1)
    }

    @MainActor
    func testClaimRewards_failure_setsError() async throws {
        mockService.shouldFail = true

        vm.claimRewards()
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertNotNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
    }
}

// MARK: - TrustViewModel Tests

final class TrustViewModelTests: XCTestCase {

    var mockService: MockTrustService!
    var vm: TrustViewModel!

    override func setUp() {
        super.setUp()
        mockService = MockTrustService()
        vm = TrustViewModel(trustService: mockService)
    }

    @MainActor
    func testFetchTrustScore_populatesScoreAndBreakdown() async throws {
        vm.fetchTrustScore(userId: "user-1")
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertEqual(vm.trustScore, 85)
        XCTAssertEqual(vm.trustLevel, "Established")
        XCTAssertNotNil(vm.breakdown)
        XCTAssertEqual(vm.breakdown?.identity, 30)
        XCTAssertEqual(vm.breakdown?.behavior, 25)
        XCTAssertFalse(vm.isLoading)
        XCTAssertEqual(mockService.fetchTrustScoreCallCount, 1)
    }

    @MainActor
    func testFetchTrustScore_failure_setsError() async throws {
        mockService.shouldFail = true

        vm.fetchTrustScore(userId: "user-1")
        try await Task.sleep(nanoseconds: 100_000_000)

        XCTAssertNotNil(vm.errorMessage)
        XCTAssertFalse(vm.isLoading)
    }
}
