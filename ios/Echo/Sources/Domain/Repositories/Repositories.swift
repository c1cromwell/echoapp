import Foundation

// MARK: - Auth Repository Protocol

protocol AuthRepository {
    func register(email: String, password: String, username: String) async throws -> LoginResponse
    func login(email: String, password: String) async throws -> LoginResponse
    func refreshToken() async throws -> RefreshTokenResponse
    func logout() async throws
    func verifyBiometric(challenge: String) async throws -> Bool
    func createPasskey() async throws -> String
    func verifyPasskey(credential: String) async throws -> Bool
}

// MARK: - Auth Repository Implementation

actor ConcreteAuthRepository: AuthRepository {
    
    private let apiClient: APIClient
    private let keychain: KeychainManager
    private let secureEnclave: SecureEnclaveManager
    
    init(
        apiClient: APIClient = APIClient(),
        keychain: KeychainManager = .shared,
        secureEnclave: SecureEnclaveManager = .shared
    ) {
        self.apiClient = apiClient
        self.keychain = keychain
        self.secureEnclave = secureEnclave
    }
    
    func register(email: String, password: String, username: String) async throws -> LoginResponse {
        // Generate public key from Secure Enclave
        let publicKeyData = try secureEnclave.getPublicKey()
        let publicKeyBase64 = publicKeyData.base64EncodedString()
        
        let request = RegisterRequest(
            email: email,
            password: password,
            username: username,
            publicKey: publicKeyBase64,
            biometricEnabled: false
        )
        
        let response: LoginResponse = try await apiClient.post(
            endpoint: AuthEndpoint.register,
            body: request
        )
        
        // Store tokens
        try await keychain.storeAuthToken(response.accessToken)
        try await keychain.storeRefreshToken(response.refreshToken)
        
        return response
    }
    
    func login(email: String, password: String) async throws -> LoginResponse {
        let request = LoginRequest(email: email, password: password)
        
        let response: LoginResponse = try await apiClient.post(
            endpoint: AuthEndpoint.login,
            body: request
        )
        
        // Store tokens
        try await keychain.storeAuthToken(response.accessToken)
        try await keychain.storeRefreshToken(response.refreshToken)
        
        return response
    }
    
    func refreshToken() async throws -> RefreshTokenResponse {
        guard let refreshToken = try await keychain.getRefreshToken() else {
            throw RepositoryError.noAuthToken
        }
        
        let request = RefreshTokenRequest(refreshToken: refreshToken)
        let response: RefreshTokenResponse = try await apiClient.post(
            endpoint: AuthEndpoint.refreshToken,
            body: request
        )
        
        // Update token
        try await keychain.storeAuthToken(response.accessToken)
        
        return response
    }
    
    func logout() async throws {
        try await apiClient.post(endpoint: AuthEndpoint.logout, body: EmptyRequest())
        try await keychain.clearAuthCredentials()
    }
    
    func verifyBiometric(challenge: String) async throws -> Bool {
        let bioManager = BiometricAuthManager()
        return await bioManager.authenticate(reason: "Verify your identity")
    }
    
    func createPasskey() async throws -> String {
        // Generate passkey from Secure Enclave
        let passkeyData = try secureEnclave.createPasskey()
        return passkeyData.base64EncodedString()
    }
    
    func verifyPasskey(credential: String) async throws -> Bool {
        let request = VerifyBiometricRequest(biometricData: credential, challenge: UUID().uuidString)
        
        let response: [String: Bool] = try await apiClient.post(
            endpoint: AuthEndpoint.verifyPasskey,
            body: request
        )
        
        return response["verified"] ?? false
    }
}

// MARK: - User Repository Protocol

protocol UserRepository {
    func getProfile() async throws -> User
    func updateProfile(username: String, avatar: String?) async throws -> User
    func getUser(id: String) async throws -> User
    func searchUsers(query: String) async throws -> [User]
    func addContact(userId: String) async throws
    func removeContact(userId: String) async throws
    func getContacts() async throws -> [Contact]
    func blockUser(userId: String) async throws
    func unblockUser(userId: String) async throws
    func uploadAvatar(_ imageData: Data) async throws -> String
    func deleteAccount() async throws
}

// MARK: - User Repository Implementation

actor ConcreteUserRepository: UserRepository {
    
    private let apiClient: APIClient
    private let localStorage: LocalDatabase
    
    init(
        apiClient: APIClient = APIClient(),
        localStorage: LocalDatabase = .shared
    ) {
        self.apiClient = apiClient
        self.localStorage = localStorage
    }
    
    func getProfile() async throws -> User {
        let user: User = try await apiClient.get(endpoint: UserEndpoint.getProfile)
        return user
    }
    
    func updateProfile(username: String, avatar: String?) async throws -> User {
        struct UpdateRequest: Codable {
            let username: String
            let avatar: String?
        }
        
        let user: User = try await apiClient.put(
            endpoint: UserEndpoint.updateProfile,
            body: UpdateRequest(username: username, avatar: avatar)
        )
        
        return user
    }
    
    func getUser(id: String) async throws -> User {
        try await apiClient.get(endpoint: UserEndpoint.getUser(id: id))
    }
    
    func searchUsers(query: String) async throws -> [User] {
        struct SearchResponse: Codable {
            let users: [User]
        }
        
        let response: SearchResponse = try await apiClient.get(
            endpoint: UserEndpoint.searchUsers(query: query)
        )
        
        return response.users
    }
    
    func addContact(userId: String) async throws {
        struct AddContactRequest: Codable {}
        
        _ = try await apiClient.post(
            endpoint: UserEndpoint.addContact(id: userId),
            body: AddContactRequest()
        ) as [String: String]
    }
    
    func removeContact(userId: String) async throws {
        _ = try await apiClient.delete(endpoint: UserEndpoint.removeContact(id: userId)) as [String: String]
    }
    
    func getContacts() async throws -> [Contact] {
        struct ContactsResponse: Codable {
            let contacts: [Contact]
        }
        
        let response: ContactsResponse = try await apiClient.get(
            endpoint: UserEndpoint.getContacts()
        )
        
        return response.contacts
    }
    
    func blockUser(userId: String) async throws {
        struct BlockRequest: Codable {}
        
        _ = try await apiClient.post(
            endpoint: UserEndpoint.blockUser(id: userId),
            body: BlockRequest()
        ) as [String: String]
    }
    
    func unblockUser(userId: String) async throws {
        _ = try await apiClient.delete(endpoint: UserEndpoint.unblockUser(id: userId)) as [String: String]
    }
    
    func uploadAvatar(_ imageData: Data) async throws -> String {
        let response: [String: String] = try await apiClient.upload(
            endpoint: UserEndpoint.uploadAvatar,
            data: imageData,
            filename: "avatar.jpg",
            mimeType: "image/jpeg"
        )
        
        return response["url"] ?? ""
    }
    
    func deleteAccount() async throws {
        _ = try await apiClient.delete(endpoint: UserEndpoint.deleteAccount) as [String: String]
    }
}

// MARK: - Message Repository Protocol

protocol MessageRepository {
    func sendMessage(conversationId: String, content: String) async throws -> Message
    func fetchMessages(conversationId: String, limit: Int) async throws -> [Message]
    func fetchConversations(limit: Int) async throws -> [Conversation]
    func createConversation(participantIds: [String]) async throws -> Conversation
    func getConversation(id: String) async throws -> Conversation
    func deleteMessage(id: String) async throws
    func editMessage(id: String, content: String) async throws -> Message
    func markAsRead(messageId: String) async throws
    func addReaction(messageId: String, emoji: String) async throws
    func removeReaction(messageId: String, emoji: String) async throws
}

// MARK: - Message Repository Implementation

actor ConcreteMessageRepository: MessageRepository {
    
    private let apiClient: APIClient
    private let webSocketClient: WebSocketClient
    private let encryption: KinnamiEncryption
    private let localStorage: LocalDatabase
    
    init(
        apiClient: APIClient = APIClient(),
        webSocketClient: WebSocketClient = WebSocketClient(),
        encryption: KinnamiEncryption = KinnamiEncryption(),
        localStorage: LocalDatabase = .shared
    ) {
        self.apiClient = apiClient
        self.webSocketClient = webSocketClient
        self.encryption = encryption
        self.localStorage = localStorage
    }
    
    func sendMessage(conversationId: String, content: String) async throws -> Message {
        let encryptedMessage = try await encryption.encryptWithKeyAgreement(
            plaintext: content,
            recipientPublicKeyData: Data()
        )
        
        let request = SendMessageRequest(
            conversationId: conversationId,
            content: content,
            encryptedContent: encryptedMessage.ciphertext,
            nonce: encryptedMessage.nonce,
            attachments: nil
        )
        
        let message: Message = try await apiClient.post(
            endpoint: MessageEndpoint.send,
            body: request
        )
        
        return message
    }
    
    func fetchMessages(conversationId: String, limit: Int) async throws -> [Message] {
        struct MessagesResponse: Codable {
            let messages: [Message]
        }
        
        let response: MessagesResponse = try await apiClient.get(
            endpoint: MessageEndpoint.fetch(conversationId: conversationId, limit: limit)
        )
        
        return response.messages
    }
    
    func fetchConversations(limit: Int) async throws -> [Conversation] {
        struct ConversationsResponse: Codable {
            let conversations: [Conversation]
        }
        
        let response: ConversationsResponse = try await apiClient.get(
            endpoint: MessageEndpoint.fetchConversations(limit: limit)
        )
        
        return response.conversations
    }
    
    func createConversation(participantIds: [String]) async throws -> Conversation {
        struct CreateRequest: Codable {
            let participantIds: [String]
        }
        
        let conversation: Conversation = try await apiClient.post(
            endpoint: MessageEndpoint.createConversation,
            body: CreateRequest(participantIds: participantIds)
        )
        
        return conversation
    }
    
    func getConversation(id: String) async throws -> Conversation {
        try await apiClient.get(endpoint: MessageEndpoint.getConversation(id: id))
    }
    
    func deleteMessage(id: String) async throws {
        _ = try await apiClient.delete(endpoint: MessageEndpoint.deleteMessage(id: id)) as [String: String]
    }
    
    func editMessage(id: String, content: String) async throws -> Message {
        struct EditRequest: Codable {
            let content: String
        }
        
        let message: Message = try await apiClient.patch(
            endpoint: MessageEndpoint.editMessage(id: id),
            body: EditRequest(content: content)
        )
        
        return message
    }
    
    func markAsRead(messageId: String) async throws {
        struct MarkReadRequest: Codable {}
        
        _ = try await apiClient.put(
            endpoint: MessageEndpoint.markAsRead(messageId: messageId),
            body: MarkReadRequest()
        ) as [String: String]
    }
    
    func addReaction(messageId: String, emoji: String) async throws {
        struct ReactionRequest: Codable {
            let emoji: String
        }
        
        _ = try await apiClient.post(
            endpoint: MessageEndpoint.addReaction(messageId: messageId),
            body: ReactionRequest(emoji: emoji)
        ) as [String: String]
    }
    
    func removeReaction(messageId: String, emoji: String) async throws {
        _ = try await apiClient.delete(
            endpoint: MessageEndpoint.removeReaction(messageId: messageId)
        ) as [String: String]
    }
}

// MARK: - Token Repository Protocol

protocol TokenRepository {
    func getBalance() async throws -> Token
    func getTransactionHistory(limit: Int) async throws -> [Transaction]
    func sendTokens(recipientId: String, amount: Decimal) async throws -> Transaction
    func stakeTokens(amount: Decimal, duration: Int) async throws -> Transaction
    func unstakeTokens(amount: Decimal) async throws -> Transaction
    func claimRewards() async throws -> Transaction
    func getStakingInfo() async throws -> [String: AnyCodable]
    func getTrustScore() async throws -> TrustScore
    func getAchievements() async throws -> [Achievement]
}

// MARK: - Token Repository Implementation

actor ConcreteTokenRepository: TokenRepository {
    
    private let apiClient: APIClient
    private let localStorage: LocalDatabase
    
    init(
        apiClient: APIClient = APIClient(),
        localStorage: LocalDatabase = .shared
    ) {
        self.apiClient = apiClient
        self.localStorage = localStorage
    }
    
    func getBalance() async throws -> Token {
        let response: TokenBalanceResponse = try await apiClient.get(
            endpoint: TokenEndpoint.getBalance
        )
        
        let token = Token(
            id: "ECHO-balance",
            balance: Decimal(string: response.balance.description) ?? 0,
            available: Decimal(string: response.available.description) ?? 0,
            frozen: Decimal(string: response.frozen.description) ?? 0,
            staked: 0,
            currency: response.currency,
            lastUpdated: Date()
        )
        
        try await localStorage.saveToken(token)
        return token
    }
    
    func getTransactionHistory(limit: Int) async throws -> [Transaction] {
        struct HistoryResponse: Codable {
            let transactions: [TransactionResponse]
        }
        
        let response: HistoryResponse = try await apiClient.get(
            endpoint: TokenEndpoint.getTransactionHistory(limit: limit)
        )
        
        return response.transactions.map { resp in
            Transaction(
                id: resp.id,
                type: TransactionType(rawValue: resp.type) ?? .send,
                amount: Decimal(string: resp.amount.description) ?? 0,
                from: resp.from,
                to: resp.to,
                status: TransactionStatus(rawValue: resp.status) ?? .pending,
                description: resp.description,
                timestamp: resp.timestamp,
                hash: nil
            )
        }
    }
    
    func sendTokens(recipientId: String, amount: Decimal) async throws -> Transaction {
        let request = SendTokensRequest(recipientId: recipientId, amount: amount, message: nil)
        
        let response: TransactionResponse = try await apiClient.post(
            endpoint: TokenEndpoint.sendTokens,
            body: request
        )
        
        return Transaction(
            id: response.id,
            type: .send,
            amount: amount,
            from: nil,
            to: recipientId,
            status: .completed,
            description: response.description,
            timestamp: response.timestamp,
            hash: nil
        )
    }
    
    func stakeTokens(amount: Decimal, duration: Int) async throws -> Transaction {
        let request = StakeTokensRequest(amount: amount, duration: duration, validator: nil)
        
        let response: TransactionResponse = try await apiClient.post(
            endpoint: TokenEndpoint.stakeTokens,
            body: request
        )
        
        return Transaction(
            id: response.id,
            type: .stake,
            amount: amount,
            from: nil,
            to: nil,
            status: .completed,
            description: response.description,
            timestamp: response.timestamp,
            hash: nil
        )
    }
    
    func unstakeTokens(amount: Decimal) async throws -> Transaction {
        struct UnstakeRequest: Codable {
            let amount: Decimal
        }
        
        let response: TransactionResponse = try await apiClient.post(
            endpoint: TokenEndpoint.unstakeTokens,
            body: UnstakeRequest(amount: amount)
        )
        
        return Transaction(
            id: response.id,
            type: .unstake,
            amount: amount,
            from: nil,
            to: nil,
            status: .completed,
            description: response.description,
            timestamp: response.timestamp,
            hash: nil
        )
    }
    
    func claimRewards() async throws -> Transaction {
        struct ClaimRequest: Codable {}
        
        let response: TransactionResponse = try await apiClient.post(
            endpoint: TokenEndpoint.claimRewards,
            body: ClaimRequest()
        )
        
        return Transaction(
            id: response.id,
            type: .reward,
            amount: Decimal(string: response.amount.description) ?? 0,
            from: nil,
            to: nil,
            status: .completed,
            description: response.description,
            timestamp: response.timestamp,
            hash: nil
        )
    }
    
    func getStakingInfo() async throws -> [String: AnyCodable] {
        return try await apiClient.get(endpoint: TokenEndpoint.getStakingInfo)
    }
    
    func getTrustScore() async throws -> TrustScore {
        let response: TrustScoreResponse = try await apiClient.get(
            endpoint: TokenEndpoint.getTrusScore
        )
        
        let trustScore = TrustScore(
            userId: "current-user",
            score: response.score,
            level: TrustLevel(rawValue: response.level) ?? .newcomer,
            nextLevelThreshold: response.nextLevelThreshold,
            multiplier: response.multiplier,
            components: response.components,
            updatedAt: Date()
        )
        
        return trustScore
    }
    
    func getAchievements() async throws -> [Achievement] {
        struct AchievementsResponse: Codable {
            let achievements: [AchievementResponse]
        }
        
        let response: AchievementsResponse = try await apiClient.get(
            endpoint: TokenEndpoint.getAchievements
        )
        
        return response.achievements.map { resp in
            Achievement(
                id: resp.id,
                name: resp.name,
                description: resp.description,
                icon: resp.icon,
                category: resp.category,
                level: 1,
                unlockedAt: resp.unlockedAt,
                points: 0
            )
        }
    }
}

// MARK: - Helper Models

struct EmptyRequest: Codable {}

// MARK: - Repository Error

enum RepositoryError: LocalizedError {
    case noAuthToken
    case invalidResponse
    case networkError(String)
    case unknown
    
    var errorDescription: String? {
        switch self {
        case .noAuthToken:
            return "No authentication token available"
        case .invalidResponse:
            return "Invalid response from server"
        case .networkError(let message):
            return "Network error: \(message)"
        case .unknown:
            return "Unknown error occurred"
        }
    }
}
