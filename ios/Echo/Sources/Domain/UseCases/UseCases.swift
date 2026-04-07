import Foundation

// MARK: - Authenticate UseCase

struct AuthenticateUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute(email: String, password: String) async throws -> LoginResponse {
        return try await repository.login(email: email, password: password)
    }
}

// MARK: - Register UseCase

struct RegisterUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute(email: String, password: String, username: String) async throws -> LoginResponse {
        return try await repository.register(email: email, password: password, username: username)
    }
}

// MARK: - Create Passkey UseCase

struct CreatePasskeyUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> String {
        return try await repository.createPasskey()
    }
}

// MARK: - Verify Passkey UseCase

struct VerifyPasskeyUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute(credential: String) async throws -> Bool {
        return try await repository.verifyPasskey(credential: credential)
    }
}

// MARK: - Logout UseCase

struct LogoutUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute() async throws {
        return try await repository.logout()
    }
}

// MARK: - Send Message UseCase

struct SendMessageUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(conversationId: String, content: String) async throws -> Message {
        return try await repository.sendMessage(conversationId: conversationId, content: content)
    }
}

// MARK: - Fetch Messages UseCase

struct FetchMessagesUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(conversationId: String, limit: Int = 50) async throws -> [Message] {
        return try await repository.fetchMessages(conversationId: conversationId, limit: limit)
    }
}

// MARK: - Fetch Conversations UseCase

struct FetchConversationsUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(limit: Int = 20) async throws -> [Conversation] {
        return try await repository.fetchConversations(limit: limit)
    }
}

// MARK: - Create Conversation UseCase

struct CreateConversationUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(participantIds: [String]) async throws -> Conversation {
        return try await repository.createConversation(participantIds: participantIds)
    }
}

// MARK: - Delete Message UseCase

struct DeleteMessageUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(messageId: String) async throws {
        return try await repository.deleteMessage(id: messageId)
    }
}

// MARK: - Edit Message UseCase

struct EditMessageUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(messageId: String, content: String) async throws -> Message {
        return try await repository.editMessage(id: messageId, content: content)
    }
}

// MARK: - Add Reaction UseCase

struct AddReactionUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(messageId: String, emoji: String) async throws {
        return try await repository.addReaction(messageId: messageId, emoji: emoji)
    }
}

// MARK: - Mark As Read UseCase

struct MarkAsReadUseCase {
    private let repository: MessageRepository
    
    init(repository: MessageRepository) {
        self.repository = repository
    }
    
    func execute(messageId: String) async throws {
        return try await repository.markAsRead(messageId: messageId)
    }
}

// MARK: - Get Profile UseCase

struct GetProfileUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> User {
        return try await repository.getProfile()
    }
}

// MARK: - Update Profile UseCase

struct UpdateProfileUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(username: String, avatar: String?) async throws -> User {
        return try await repository.updateProfile(username: username, avatar: avatar)
    }
}

// MARK: - Search Users UseCase

struct SearchUsersUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(query: String) async throws -> [User] {
        return try await repository.searchUsers(query: query)
    }
}

// MARK: - Add Contact UseCase

struct AddContactUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(userId: String) async throws {
        return try await repository.addContact(userId: userId)
    }
}

// MARK: - Get Contacts UseCase

struct GetContactsUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> [Contact] {
        return try await repository.getContacts()
    }
}

// MARK: - Upload Avatar UseCase

struct UploadAvatarUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(imageData: Data) async throws -> String {
        return try await repository.uploadAvatar(imageData)
    }
}

// MARK: - Create DID UseCase

struct CreateDIDUseCase {
    private let userRepository: UserRepository
    private let secureEnclave: SecureEnclaveManager
    
    init(
        userRepository: UserRepository,
        secureEnclave: SecureEnclaveManager = .shared
    ) {
        self.userRepository = userRepository
        self.secureEnclave = secureEnclave
    }
    
    func execute() async throws -> DID {
        let keyId = "did-key-\(UUID().uuidString)"
        let publicKeyBase64 = try await secureEnclave.generateBiometricProtectedKey(id: keyId)
        
        let did = DID(
            id: UUID().uuidString,
            did: "did:echo:\(UUID().uuidString)",
            publicKey: publicKeyBase64,
            verificationMethod: "secureEnclave",
            proofType: "EcdsaSecp256r1Signature2019",
            created: Date(),
            updated: Date()
        )
        
        return did
    }
}

// MARK: - Verify Identity UseCase

struct VerifyIdentityUseCase {
    private let userRepository: UserRepository
    
    init(repository: UserRepository) {
        self.userRepository = repository
    }
    
    func execute(biometricChallenge: String) async throws -> Bool {
        // Perform biometric verification
        let bioManager = BiometricAuthManager()
        return await bioManager.authenticate(reason: "Verify your identity for credential issuance")
    }
}

// MARK: - Get Balance UseCase

struct GetBalanceUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> Token {
        return try await repository.getBalance()
    }
}

// MARK: - Get Transaction History UseCase

struct GetTransactionHistoryUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute(limit: Int = 50) async throws -> [Transaction] {
        return try await repository.getTransactionHistory(limit: limit)
    }
}

// MARK: - Send Tokens UseCase

struct SendTokensUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute(recipientId: String, amount: Decimal) async throws -> Transaction {
        return try await repository.sendTokens(recipientId: recipientId, amount: amount)
    }
}

// MARK: - Stake Tokens UseCase

struct StakeTokensUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute(amount: Decimal, duration: Int) async throws -> Transaction {
        return try await repository.stakeTokens(amount: amount, duration: duration)
    }
}

// MARK: - Unstake Tokens UseCase

struct UnstakeTokensUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute(amount: Decimal) async throws -> Transaction {
        return try await repository.unstakeTokens(amount: amount)
    }
}

// MARK: - Claim Rewards UseCase

struct ClaimRewardsUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> Transaction {
        return try await repository.claimRewards()
    }
}

// MARK: - Get Trust Score UseCase

struct GetTrustScoreUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> TrustScore {
        return try await repository.getTrustScore()
    }
}

// MARK: - Get Achievements UseCase

struct GetAchievementsUseCase {
    private let repository: TokenRepository
    
    init(repository: TokenRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> [Achievement] {
        return try await repository.getAchievements()
    }
}

// MARK: - Block User UseCase

struct BlockUserUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(userId: String) async throws {
        return try await repository.blockUser(userId: userId)
    }
}

// MARK: - Unblock User UseCase

struct UnblockUserUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(userId: String) async throws {
        return try await repository.unblockUser(userId: userId)
    }
}

// MARK: - Remove Contact UseCase

struct RemoveContactUseCase {
    private let repository: UserRepository
    
    init(repository: UserRepository) {
        self.repository = repository
    }
    
    func execute(userId: String) async throws {
        return try await repository.removeContact(userId: userId)
    }
}

// MARK: - Refresh Token UseCase

struct RefreshTokenUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute() async throws -> RefreshTokenResponse {
        return try await repository.refreshToken()
    }
}

// MARK: - Verify Biometric UseCase

struct VerifyBiometricUseCase {
    private let repository: AuthRepository
    
    init(repository: AuthRepository) {
        self.repository = repository
    }
    
    func execute(challenge: String) async throws -> Bool {
        return try await repository.verifyBiometric(challenge: challenge)
    }
}
