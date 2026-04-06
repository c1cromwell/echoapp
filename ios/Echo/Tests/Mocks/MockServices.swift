import Foundation
@testable import Echo

// MARK: - Mock Auth Service

class MockAuthService: AuthServiceProtocol {
    var requestOTPResult: Result<OTPResponse, Error> = .success(
        OTPResponse(expiresIn: 300, phone: "+15550100")
    )
    var verifyOTPResult: Result<AuthResponse, Error> = .success(
        AuthResponse(
            token: "mock-token-123",
            refreshToken: "mock-refresh-456",
            did: "did:prism:mock123",
            user: UserProfile(
                id: "user-1",
                phone: "+15550100",
                displayName: "Test User",
                username: "testuser",
                avatarURL: nil
            )
        )
    )
    var registerPasskeyResult: Result<Void, Error> = .success(())
    var authenticateWithPasskeyResult: Result<AuthResponse, Error>?
    var refreshTokenResult: Result<String, Error> = .success("new-token")

    // Call tracking
    var requestOTPCallCount = 0
    var verifyOTPCallCount = 0
    var registerPasskeyCallCount = 0
    var lastRequestedPhone: String?
    var lastVerifiedCode: String?

    func requestOTP(phone: String) async throws -> OTPResponse {
        requestOTPCallCount += 1
        lastRequestedPhone = phone
        return try requestOTPResult.get()
    }

    func verifyOTP(phone: String, code: String) async throws -> AuthResponse {
        verifyOTPCallCount += 1
        lastVerifiedCode = code
        return try verifyOTPResult.get()
    }

    func registerPasskey() async throws {
        registerPasskeyCallCount += 1
        try registerPasskeyResult.get()
    }

    func authenticateWithPasskey() async throws -> AuthResponse {
        return try (authenticateWithPasskeyResult ?? verifyOTPResult).get()
    }

    func refreshToken() async throws -> String {
        return try refreshTokenResult.get()
    }
}

// MARK: - Mock Keychain Manager

class MockKeychainManager: KeychainManagerProtocol {
    var storedToken: String?
    var saveTokenCallCount = 0
    var clearAllCallCount = 0

    func saveToken(_ token: String) throws {
        saveTokenCallCount += 1
        storedToken = token
    }

    func retrieveToken() -> String? {
        return storedToken
    }

    func clearAll() {
        clearAllCallCount += 1
        storedToken = nil
    }
}

// MARK: - Mock Messaging Service

class MockMessagingService: MessagingServiceProtocol {
    var conversations: [ConversationModel] = []
    var messages: [String: [MessageModel]] = [:]
    var sendMessageResult: Result<Void, Error> = .success(())

    // Call tracking
    var fetchConversationsCallCount = 0
    var fetchMessagesCallCount = 0
    var sendMessageCallCount = 0
    var markAsReadCallCount = 0
    var lastSentContent: String?
    var lastSentConversationId: String?

    func fetchConversations() async throws -> [ConversationModel] {
        fetchConversationsCallCount += 1
        return conversations
    }

    func fetchMessages(conversationId: String) async throws -> [MessageModel] {
        fetchMessagesCallCount += 1
        return messages[conversationId] ?? []
    }

    func sendMessage(_ content: String, to conversationId: String) async throws {
        sendMessageCallCount += 1
        lastSentContent = content
        lastSentConversationId = conversationId
        try sendMessageResult.get()

        // Add message to local store so fetchMessages returns it
        let msg = MessageModel(
            id: UUID().uuidString,
            conversationId: conversationId,
            senderId: "current-user",
            content: content,
            status: .sent,
            createdAt: Date()
        )
        messages[conversationId, default: []].append(msg)
    }

    func markAsRead(conversationId: String) async throws {
        markAsReadCallCount += 1
    }
}

// MARK: - Mock Profile Service

class MockProfileService: ProfileServiceProtocol {
    var profileData = ProfileData(
        displayName: "Test User",
        username: "testuser",
        bio: "Hello world",
        status: "Available",
        trustScore: 75,
        trustLevel: "Established",
        isVerified: true,
        messagesSent: 42,
        contactsCount: 10,
        echoRewards: 150.0
    )
    var personas: [Persona] = []
    var settings = ProfileSettings()
    var storageInfo = StorageInfo()
    var usernameAvailable = true
    var shouldFail = false
    var failError: Error = NSError(domain: "MockError", code: -1, userInfo: [NSLocalizedDescriptionKey: "Mock error"])

    // Call tracking
    var fetchProfileCallCount = 0
    var updateProfileCallCount = 0
    var checkUsernameCallCount = 0
    var lastCheckedUsername: String?

    func fetchProfile() async throws -> ProfileData {
        fetchProfileCallCount += 1
        if shouldFail { throw failError }
        return profileData
    }

    func updateProfile(_ profile: ProfileData) async throws -> ProfileData {
        updateProfileCallCount += 1
        if shouldFail { throw failError }
        profileData = profile
        return profile
    }

    func checkUsernameAvailability(_ username: String) async throws -> Bool {
        checkUsernameCallCount += 1
        lastCheckedUsername = username
        if shouldFail { throw failError }
        return usernameAvailable
    }

    func uploadAvatar(data: Data) async throws -> String {
        return "https://cdn.echo.local/avatar/mock.jpg"
    }

    func fetchPersonas() async throws -> [Persona] { return personas }
    func createPersona(_ persona: Persona) async throws -> Persona { return persona }
    func updatePersona(_ persona: Persona) async throws -> Persona { return persona }
    func deletePersona(id: String, options: PersonaDeletionOptions) async throws {}
    func recoverPersona(id: String) async throws -> Persona {
        return personas.first { $0.id == id }!
    }
    func setDefaultPersona(id: String) async throws {}
    func fetchSettings() async throws -> ProfileSettings { return settings }
    func updateNotificationSettings(_ settings: EnhancedNotificationSettings) async throws {}
    func updatePrivacySettings(_ settings: EnhancedPrivacySettings) async throws {}
    func updateAppearanceSettings(_ settings: AppearanceSettings) async throws {}
    func updatePersonaPrivacySettings(personaId: String, settings: PersonaPrivacySettings) async throws {}
    func updatePersonaNotificationSettings(personaId: String, settings: PersonaNotificationSettings) async throws {}
    func updatePersonaFeatureSettings(personaId: String, settings: PersonaFeatureSettings) async throws {}
    func fetchStorageInfo() async throws -> StorageInfo { return storageInfo }
    func clearCache() async throws {}
    func backUpNow() async throws {}
    func deleteAccount() async throws {}
    func grantAccess(personaId: String, contactId: String, permissions: AccessPermissions) async throws -> AccessGrant {
        fatalError("Not implemented in mock")
    }
    func revokeAccess(grantId: String) async throws {}
    func fetchVisibilityMatrix() async throws -> [VisibilityMatrixEntry] { return [] }
    func validatePersonaSwitch(from: String?, to: String, contactId: String) async throws -> PersonaSwitchContext {
        fatalError("Not implemented in mock")
    }
    func fetchPersonaConversations(personaId: String) async throws -> [PersonaConversation] { return [] }
    func addPersonaBadge(personaId: String, badge: PersonaBadge) async throws -> PersonaBadge { return badge }
    func removePersonaBadge(personaId: String, badgeId: String) async throws {}
    func exportPersonaData(personaId: String) async throws -> URL {
        return URL(fileURLWithPath: "/tmp/export.json")
    }
}

// MARK: - Mock Rewards Service

class MockRewardsService: RewardsServiceProtocol {
    var balance: Double = 250.0
    var activities: [RewardActivityModel] = [
        RewardActivityModel(id: "1", type: "message_sent", amount: 0.5, description: "Sent a message", date: Date()),
        RewardActivityModel(id: "2", type: "daily_login", amount: 1.0, description: "Daily login bonus", date: Date()),
    ]
    var claimResult: Double = 275.0
    var shouldFail = false

    var fetchBalanceCallCount = 0
    var fetchActivityCallCount = 0
    var claimRewardsCallCount = 0

    func fetchBalance() async throws -> Double {
        fetchBalanceCallCount += 1
        if shouldFail { throw NSError(domain: "MockError", code: -1) }
        return balance
    }

    func fetchActivity() async throws -> [RewardActivityModel] {
        fetchActivityCallCount += 1
        return activities
    }

    func stakeTokens(amount: Double, period: Int) async throws {}

    func claimRewards() async throws -> Double {
        claimRewardsCallCount += 1
        if shouldFail { throw NSError(domain: "MockError", code: -1) }
        balance = claimResult
        return claimResult
    }
}

// MARK: - Mock Trust Service

class MockTrustService: TrustServiceProtocol {
    var trustResult = TrustScoreResult(
        score: 85,
        level: "Established",
        breakdown: TrustBreakdown(identity: 30, behavior: 25, network: 20, activity: 10)
    )
    var shouldFail = false

    var fetchTrustScoreCallCount = 0

    func fetchTrustScore(userId: String) async throws -> TrustScoreResult {
        fetchTrustScoreCallCount += 1
        if shouldFail { throw NSError(domain: "MockError", code: -1) }
        return trustResult
    }

    func submitVerification(documents: [URL], selfie: URL) async throws {}
    func updateTrustCircle(contactId: String, tier: String) async throws {}
}
