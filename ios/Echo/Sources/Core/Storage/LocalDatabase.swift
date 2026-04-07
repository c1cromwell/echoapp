import Foundation
import SwiftData

/// Local database manager using SwiftData
actor LocalDatabase {
    
    // MARK: - Singleton
    static let shared = LocalDatabase()
    
    // MARK: - Properties
    
    @MainActor
    private static var modelContainer: ModelContainer?
    
    private let modelTypes: [any PersistentModel.Type] = [
        LocalUser.self,
        LocalConversation.self,
        LocalMessage.self,
        LocalContact.self,
        LocalDID.self,
        LocalCredential.self,
        LocalToken.self,
        LocalAchievement.self
    ]
    
    // MARK: - Initialization
    
    @MainActor
    static func setup() throws {
        guard modelContainer == nil else { return }
        
        let schema = Schema(LocalDatabase.shared.modelTypes)
        let config = ModelConfiguration(
            isStoredInMemoryOnly: false,
            allowsSave: true,
            cloudKitDatabase: .none
        )
        let container = try ModelContainer(
            for: schema,
            configurations: config
        )
        
        modelContainer = container
    }
    
    private init() {}
    
    @MainActor
    private static var context: ModelContext? {
        guard let container = modelContainer else { return nil }
        return ModelContext(container)
    }
    
    // MARK: - User Operations
    
    func saveUser(_ user: LocalUser) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(user)
        try context.save()
    }
    
    func getUser(id: String) async throws -> LocalUser? {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalUser>(
            predicate: #Predicate { $0.id == id }
        )
        descriptor.fetchLimit = 1
        
        return try context.fetch(descriptor).first
    }
    
    func updateUser(_ user: LocalUser) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        try context.save()
    }
    
    // MARK: - Message Operations
    
    func saveMessage(_ message: LocalMessage) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(message)
        try context.save()
    }
    
    func getMessages(conversationId: String, limit: Int = 50) async throws -> [LocalMessage] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalMessage>(
            predicate: #Predicate { $0.conversationId == conversationId }
        )
        descriptor.fetchLimit = limit
        descriptor.sortBy = [SortDescriptor(\.timestamp, order: .reverse)]
        
        return try context.fetch(descriptor)
    }
    
    func deleteMessage(id: String) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalMessage>(
            predicate: #Predicate { $0.id == id }
        )
        
        if let message = try context.fetch(descriptor).first {
            context.delete(message)
            try context.save()
        }
    }
    
    // MARK: - Conversation Operations
    
    func saveConversation(_ conversation: LocalConversation) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(conversation)
        try context.save()
    }
    
    func getConversations(limit: Int = 20) async throws -> [LocalConversation] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalConversation>()
        descriptor.fetchLimit = limit
        descriptor.sortBy = [SortDescriptor(\.updatedAt, order: .reverse)]
        
        return try context.fetch(descriptor)
    }
    
    func getConversation(id: String) async throws -> LocalConversation? {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalConversation>(
            predicate: #Predicate { $0.id == id }
        )
        descriptor.fetchLimit = 1
        
        return try context.fetch(descriptor).first
    }
    
    // MARK: - Contact Operations
    
    func saveContact(_ contact: LocalContact) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(contact)
        try context.save()
    }
    
    func getContacts(limit: Int = 100) async throws -> [LocalContact] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalContact>()
        descriptor.fetchLimit = limit
        descriptor.sortBy = [SortDescriptor(\.username, order: .forward)]
        
        return try context.fetch(descriptor)
    }
    
    func removeContact(id: String) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalContact>(
            predicate: #Predicate { $0.id == id }
        )
        
        if let contact = try context.fetch(descriptor).first {
            context.delete(contact)
            try context.save()
        }
    }
    
    // MARK: - DID Operations
    
    func saveDID(_ did: LocalDID) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(did)
        try context.save()
    }
    
    func getDIDs() async throws -> [LocalDID] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalDID>()
        descriptor.sortBy = [SortDescriptor(\.createdAt, order: .reverse)]
        
        return try context.fetch(descriptor)
    }
    
    // MARK: - Credential Operations
    
    func saveCredential(_ credential: LocalCredential) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(credential)
        try context.save()
    }
    
    func getCredentials() async throws -> [LocalCredential] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        return try context.fetch(FetchDescriptor<LocalCredential>())
    }
    
    func revokeCredential(id: String) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalCredential>(
            predicate: #Predicate { $0.id == id }
        )
        
        if let credential = try context.fetch(descriptor).first {
            context.delete(credential)
            try context.save()
        }
    }
    
    // MARK: - Token Operations
    
    func saveToken(_ token: LocalToken) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(token)
        try context.save()
    }
    
    func getToken() async throws -> LocalToken? {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalToken>()
        descriptor.fetchLimit = 1
        
        return try context.fetch(descriptor).first
    }
    
    // MARK: - Achievement Operations
    
    func saveAchievement(_ achievement: LocalAchievement) async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        context.insert(achievement)
        try context.save()
    }
    
    func getAchievements() async throws -> [LocalAchievement] {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        var descriptor = FetchDescriptor<LocalAchievement>()
        descriptor.sortBy = [SortDescriptor(\.unlockedAt, order: .reverse)]
        
        return try context.fetch(descriptor)
    }
    
    // MARK: - Cleanup
    
    func clearAllData() async throws {
        guard let context = await Self.context else {
            throw DatabaseError.notInitialized
        }
        
        try context.delete(model: LocalUser.self)
        try context.delete(model: LocalMessage.self)
        try context.delete(model: LocalConversation.self)
        try context.delete(model: LocalContact.self)
        try context.delete(model: LocalDID.self)
        try context.delete(model: LocalCredential.self)
        try context.delete(model: LocalToken.self)
        try context.delete(model: LocalAchievement.self)
        
        try context.save()
    }
}

// MARK: - Database Models

@Model
final class LocalUser {
    @Attribute(.unique) var id: String
    var email: String
    var username: String
    var avatar: String?
    var publicKey: String
    var createdAt: Date
    var updatedAt: Date
    
    init(
        id: String,
        email: String,
        username: String,
        avatar: String?,
        publicKey: String,
        createdAt: Date = Date(),
        updatedAt: Date = Date()
    ) {
        self.id = id
        self.email = email
        self.username = username
        self.avatar = avatar
        self.publicKey = publicKey
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}

@Model
final class LocalConversation {
    @Attribute(.unique) var id: String
    var name: String?
    var participantIds: [String] = []
    var lastMessage: String?
    var unreadCount: Int = 0
    var createdAt: Date
    var updatedAt: Date
    
    init(
        id: String,
        name: String?,
        participantIds: [String],
        lastMessage: String?,
        unreadCount: Int = 0,
        createdAt: Date = Date(),
        updatedAt: Date = Date()
    ) {
        self.id = id
        self.name = name
        self.participantIds = participantIds
        self.lastMessage = lastMessage
        self.unreadCount = unreadCount
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}

@Model
final class LocalMessage {
    @Attribute(.unique) var id: String
    var conversationId: String
    var senderId: String
    var content: String
    var encryptedContent: String
    var nonce: String
    var timestamp: Date
    var isRead: Bool = false
    
    init(
        id: String,
        conversationId: String,
        senderId: String,
        content: String,
        encryptedContent: String,
        nonce: String,
        timestamp: Date = Date(),
        isRead: Bool = false
    ) {
        self.id = id
        self.conversationId = conversationId
        self.senderId = senderId
        self.content = content
        self.encryptedContent = encryptedContent
        self.nonce = nonce
        self.timestamp = timestamp
        self.isRead = isRead
    }
}

@Model
final class LocalContact {
    @Attribute(.unique) var id: String
    var username: String
    var avatar: String?
    var isOnline: Bool = false
    var addedAt: Date
    
    init(
        id: String,
        username: String,
        avatar: String?,
        isOnline: Bool = false,
        addedAt: Date = Date()
    ) {
        self.id = id
        self.username = username
        self.avatar = avatar
        self.isOnline = isOnline
        self.addedAt = addedAt
    }
}

@Model
final class LocalDID {
    @Attribute(.unique) var did: String
    var publicKey: String
    var verificationMethod: String
    var createdAt: Date
    
    init(
        did: String,
        publicKey: String,
        verificationMethod: String,
        createdAt: Date = Date()
    ) {
        self.did = did
        self.publicKey = publicKey
        self.verificationMethod = verificationMethod
        self.createdAt = createdAt
    }
}

@Model
final class LocalCredential {
    @Attribute(.unique) var id: String
    var type: String
    var subject: String
    var issuer: String
    var issuedAt: Date
    var expiresAt: Date?
    var credentialData: Data
    
    init(
        id: String,
        type: String,
        subject: String,
        issuer: String,
        issuedAt: Date = Date(),
        expiresAt: Date?,
        credentialData: Data
    ) {
        self.id = id
        self.type = type
        self.subject = subject
        self.issuer = issuer
        self.issuedAt = issuedAt
        self.expiresAt = expiresAt
        self.credentialData = credentialData
    }
}

@Model
final class LocalToken {
    @Attribute(.unique) var id: String
    var balance: Decimal
    var frozen: Decimal
    var available: Decimal
    var staked: Decimal = 0
    var currency: String
    var lastUpdated: Date
    
    init(
        id: String,
        balance: Decimal,
        frozen: Decimal,
        available: Decimal,
        staked: Decimal = 0,
        currency: String = "ECHO",
        lastUpdated: Date = Date()
    ) {
        self.id = id
        self.balance = balance
        self.frozen = frozen
        self.available = available
        self.staked = staked
        self.currency = currency
        self.lastUpdated = lastUpdated
    }
}

@Model
final class LocalAchievement {
    @Attribute(.unique) var id: String
    var name: String
    var achievementDescription: String
    var icon: String
    var category: String
    var level: Int
    var unlockedAt: Date?
    
    init(
        id: String,
        name: String,
        achievementDescription: String,
        icon: String,
        category: String,
        level: Int,
        unlockedAt: Date?
    ) {
        self.id = id
        self.name = name
        self.achievementDescription = achievementDescription
        self.icon = icon
        self.category = category
        self.level = level
        self.unlockedAt = unlockedAt
    }
}

// MARK: - Database Errors

enum DatabaseError: LocalizedError {
    case notInitialized
    case saveFailed
    case fetchFailed
    case deleteFailed
    case invalidData
    
    var errorDescription: String? {
        switch self {
        case .notInitialized:
            return "Database has not been initialized"
        case .saveFailed:
            return "Failed to save data to database"
        case .fetchFailed:
            return "Failed to fetch data from database"
        case .deleteFailed:
            return "Failed to delete data from database"
        case .invalidData:
            return "Invalid data format"
        }
    }
}
