import Foundation

// MARK: - User Model

struct User: Codable, Identifiable {
    let id: String
    let email: String
    let username: String
    let avatar: String?
    let publicKey: String
    let did: String?
    let trustScore: Int?
    let isVerified: Bool
    let createdAt: Date
    let updatedAt: Date
}

// MARK: - Message Model

// DeliveryStatus is defined in Features/Evidence/DeliveryStatus.swift
// with full Comparable conformance, icons, and display labels.

struct Message: Codable, Identifiable {
    let id: String
    let conversationId: String
    let sender: User
    let content: String
    let encryptedContent: String
    let nonce: String
    let attachments: [Attachment]?
    let reactions: [Reaction]?
    let timestamp: Date
    let isRead: Bool
    let editedAt: Date?
    var deliveryStatus: DeliveryStatus?
    var evidenceEventId: String?
    var snapshotHash: String?
    var snapshotHeight: Int?

    enum CodingKeys: String, CodingKey {
        case id, conversationId, sender, content, encryptedContent, nonce
        case attachments, reactions, timestamp, isRead, editedAt
        case deliveryStatus, evidenceEventId, snapshotHash, snapshotHeight
    }
}

// MARK: - Attachment Model

struct Attachment: Codable, Identifiable {
    let id: String
    let filename: String
    let mimeType: String
    let size: Int
    let url: String
    let uploadedAt: Date
}

// MARK: - Reaction Model

struct Reaction: Codable, Identifiable {
    let id: String
    let messageId: String
    let user: User
    let emoji: String
    let timestamp: Date
}

// MARK: - Conversation Model

struct Conversation: Codable, Identifiable {
    let id: String
    let participants: [User]
    let name: String?
    let lastMessage: Message?
    let unreadCount: Int
    let createdAt: Date
    let updatedAt: Date
    
    var displayName: String {
        if let name = name {
            return name
        }
        return participants.map { $0.username }.joined(separator: ", ")
    }
}

// MARK: - Contact Model

struct Contact: Codable, Identifiable {
    let id: String
    let user: User
    let isOnline: Bool
    let lastSeen: Date?
    let addedAt: Date
}

// MARK: - DID Model (Decentralized Identifier)

struct DID: Codable, Identifiable {
    let id: String
    let did: String
    let publicKey: String
    let verificationMethod: String
    let proofType: String?
    let created: Date
    let updated: Date
}

// MARK: - Credential Model

struct Credential: Codable, Identifiable {
    let id: String
    let type: String
    let subject: String
    let issuer: String
    let credentialSubject: CredentialSubject
    let proof: CredentialProof
    let issuedAt: Date
    let expiresAt: Date?
    let isRevoked: Bool
}

struct CredentialSubject: Codable {
    let id: String
    let claims: [String: AnyCodable]
}

struct CredentialProof: Codable {
    let type: String
    let created: Date
    let verificationMethod: String
    let proofValue: String
}

// MARK: - Token Model

struct Token: Codable, Identifiable {
    let id: String
    let balance: Decimal
    let available: Decimal
    let frozen: Decimal
    let staked: Decimal
    let currency: String
    let lastUpdated: Date
}

// MARK: - Transaction Model

struct Transaction: Codable, Identifiable {
    let id: String
    let type: TransactionType
    let amount: Decimal
    let from: String?
    let to: String?
    let status: TransactionStatus
    let description: String?
    let timestamp: Date
    let hash: String?
}

enum TransactionType: String, Codable {
    case send, receive, stake, unstake, reward, penalty
}

enum TransactionStatus: String, Codable {
    case pending, completed, failed, cancelled
}

// MARK: - Trust Score Model

struct TrustScore: Codable {
    let userId: String
    let score: Int
    let level: TrustLevel
    let nextLevelThreshold: Int
    let multiplier: Double
    let components: TrustScoreComponents
    let updatedAt: Date
}

enum TrustLevel: String, Codable {
    case newcomer, basic, trusted, verified, elite
}

struct TrustScoreComponents: Codable {
    let verification: Int
    let behavior: Int
    let onChain: Int
    let penalties: Int
}

// MARK: - Achievement Model

struct Achievement: Codable, Identifiable {
    let id: String
    let name: String
    let description: String
    let icon: String
    let category: String
    let level: Int
    let unlockedAt: Date?
    let points: Int
    
    var isUnlocked: Bool {
        unlockedAt != nil
    }
}

// MARK: - Badge Model

struct Badge: Codable, Identifiable {
    let id: String
    let name: String
    let description: String
    let icon: String
    let color: String
    let earnedAt: Date?
    
    var isEarned: Bool {
        earnedAt != nil
    }
}

// MARK: - Verification Model

struct Verification: Codable, Identifiable {
    let id: String
    let userId: String
    let type: VerificationType
    let status: VerificationStatus
    let evidence: String?
    let verifiedAt: Date?
    let expiresAt: Date?
}

enum VerificationType: String, Codable {
    case email, phone, identity, business
}

enum VerificationStatus: String, Codable {
    case pending, verified, rejected, expired
}

// MARK: - Profile Model

struct Profile: Codable {
    let user: User
    let bio: String?
    let location: String?
    let website: String?
    let coverImage: String?
    let followers: Int
    let following: Int
    let verifications: [Verification]
    let achievements: [Achievement]
    let trustScore: TrustScore?
}

// MARK: - Settings Model

struct AppSettings: Codable {
    var notifications: NotificationSettings
    var privacy: PrivacySettings
    var security: SecuritySettings
    var preferences: PreferenceSettings
}

struct NotificationSettings: Codable {
    var enabled: Bool = true
    var messages: Bool = true
    var reactions: Bool = true
    var mentions: Bool = true
    var achievements: Bool = true
    var systemNotifications: Bool = true
}

struct PrivacySettings: Codable {
    var profileVisible: Bool = true
    var showOnlineStatus: Bool = true
    var allowDirectMessages: Bool = true
    var allowGroupInvites: Bool = true
}

struct SecuritySettings: Codable {
    var biometricEnabled: Bool = false
    var twoFactorEnabled: Bool = false
    var sessionTimeout: Int = 3600
    var loginAlerts: Bool = true
}

struct PreferenceSettings: Codable {
    var theme: String = "system"
    var language: String = "en"
    var fontSize: Int = 16
    var compactMode: Bool = false
}

// MARK: - Persona Model

enum PersonaCategory: String, Codable, CaseIterable, Identifiable {
    case professional, personal, family, gaming, dating, creative, anonymous, custom

    var id: String { rawValue }

    var icon: String {
        switch self {
        case .professional: return "briefcase.fill"
        case .personal: return "house.fill"
        case .family: return "figure.2.and.child"
        case .gaming: return "gamecontroller.fill"
        case .dating: return "heart.fill"
        case .creative: return "paintpalette.fill"
        case .anonymous: return "theatermasks.fill"
        case .custom: return "sparkles"
        }
    }

    var emoji: String {
        switch self {
        case .professional: return "\u{1F454}"
        case .personal: return "\u{1F3E0}"
        case .family: return "\u{1F468}\u{200D}\u{1F469}\u{200D}\u{1F467}"
        case .gaming: return "\u{1F3AE}"
        case .dating: return "\u{1F495}"
        case .creative: return "\u{1F3A8}"
        case .anonymous: return "\u{1F3AD}"
        case .custom: return "\u{2728}"
        }
    }

    var label: String {
        switch self {
        case .professional: return "Pro"
        case .personal: return "Personal"
        case .family: return "Family"
        case .gaming: return "Gaming"
        case .dating: return "Dating"
        case .creative: return "Creative"
        case .anonymous: return "Anon"
        case .custom: return "Custom"
        }
    }
}

// Keep backward-compatible typealias
typealias PersonaType = PersonaCategory

enum PersonaVisibility: String, Codable {
    case all, selected, hidden
}

struct Persona: Codable, Identifiable {
    let id: String
    var type: PersonaType
    var name: String
    var displayName: String
    var username: String?
    var bio: String?
    var avatarURL: String?
    var useMainAvatar: Bool
    var visibility: PersonaVisibility
    var defaultVisibility: PersonaDefaultVisibility
    var discoverability: Bool
    var selectedContactIds: [String]
    var accessGrants: [AccessGrant]
    var isDefault: Bool
    var createdAt: Date
    var updatedAt: Date
    var lastActiveAt: Date?
    var messageCount: Int
    var status: String?

    // Per-persona settings
    var privacySettings: PersonaPrivacySettings
    var notificationSettings: PersonaNotificationSettings
    var featureSettings: PersonaFeatureSettings

    // Per-persona keys (derived from master)
    var keys: PersonaKeys?

    // Per-persona badges and verification
    var badges: [PersonaBadge]
    var credentials: [PersonaCredential]

    // Deletion state
    var deletionState: PersonaDeletionState?

    var contactCount: Int {
        selectedContactIds.count
    }

    init(
        id: String,
        type: PersonaType,
        name: String,
        displayName: String,
        username: String? = nil,
        bio: String? = nil,
        avatarURL: String? = nil,
        useMainAvatar: Bool = true,
        visibility: PersonaVisibility = .all,
        defaultVisibility: PersonaDefaultVisibility = .contacts,
        discoverability: Bool = true,
        selectedContactIds: [String] = [],
        accessGrants: [AccessGrant] = [],
        isDefault: Bool = false,
        createdAt: Date = Date(),
        updatedAt: Date = Date(),
        lastActiveAt: Date? = nil,
        messageCount: Int = 0,
        status: String? = nil,
        privacySettings: PersonaPrivacySettings = .init(),
        notificationSettings: PersonaNotificationSettings = .init(),
        featureSettings: PersonaFeatureSettings = .init(),
        keys: PersonaKeys? = nil,
        badges: [PersonaBadge] = [],
        credentials: [PersonaCredential] = [],
        deletionState: PersonaDeletionState? = nil
    ) {
        self.id = id
        self.type = type
        self.name = name
        self.displayName = displayName
        self.username = username
        self.bio = bio
        self.avatarURL = avatarURL
        self.useMainAvatar = useMainAvatar
        self.visibility = visibility
        self.defaultVisibility = defaultVisibility
        self.discoverability = discoverability
        self.selectedContactIds = selectedContactIds
        self.accessGrants = accessGrants
        self.isDefault = isDefault
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.lastActiveAt = lastActiveAt
        self.messageCount = messageCount
        self.status = status
        self.privacySettings = privacySettings
        self.notificationSettings = notificationSettings
        self.featureSettings = featureSettings
        self.keys = keys
        self.badges = badges
        self.credentials = credentials
        self.deletionState = deletionState
    }
}

// MARK: - Persona Default Visibility

enum PersonaDefaultVisibility: String, Codable {
    case none, contacts, `public`
}

// MARK: - Access Grant

struct AccessGrant: Codable, Identifiable {
    let id: String
    let contactId: String
    let personaId: String
    let grantedAt: Date
    let grantedBy: String
    var permissions: AccessPermissions
    var expiresAt: Date?
    var revocable: Bool

    init(
        id: String = UUID().uuidString,
        contactId: String,
        personaId: String,
        grantedAt: Date = Date(),
        grantedBy: String = "",
        permissions: AccessPermissions = .init(),
        expiresAt: Date? = nil,
        revocable: Bool = true
    ) {
        self.id = id
        self.contactId = contactId
        self.personaId = personaId
        self.grantedAt = grantedAt
        self.grantedBy = grantedBy
        self.permissions = permissions
        self.expiresAt = expiresAt
        self.revocable = revocable
    }
}

struct AccessPermissions: Codable {
    var canView: Bool = true
    var canMessage: Bool = true
    var canCall: Bool = false
    var canSeeOtherPersonas: Bool = false
}

// MARK: - Per-Persona Privacy Settings

struct PersonaPrivacySettings: Codable {
    // Visibility controls
    var lastSeenVisibility: String = "contacts"
    var onlineStatusVisibility: String = "contacts"
    var profilePictureVisibility: String = "everyone"
    var bioVisibility: String = "everyone"
    var statusMessageVisibility: String = "contacts"

    // Interaction controls
    var whoCanMessage: String = "contacts"
    var whoCanCall: String = "contacts"
    var whoCanAddToGroups: String = "contacts"
    var requireApprovalForContact: Bool = false

    // Read receipts
    var sendReadReceipts: Bool = true
    var sendTypingIndicators: Bool = true

    // Discovery
    var searchable: Bool = true
    var showInSuggestions: Bool = true
    var allowContactSharing: Bool = false

    // Cross-persona
    var allowLinkingDiscovery: Bool = false
    var showSharedTrustScore: Bool = true
    var allowCrossPersonaForward: Bool = false
}

// MARK: - Per-Persona Notification Settings

struct PersonaNotificationSettings: Codable {
    var enabled: Bool = true

    // Quiet hours per persona
    var quietHoursEnabled: Bool = false
    var quietHoursStart: String = "22:00"
    var quietHoursEnd: String = "07:00"
    var quietHoursTimezone: String = "UTC"
    var quietHoursAllowExceptions: Bool = true

    // Notification types
    var messagesMode: String = "all"
    var callsMode: String = "all"
    var groupActivityMode: String = "all"
    var contactRequests: Bool = true

    // Sound and vibration
    var soundEnabled: Bool = true
    var soundId: String = "default"
    var vibrationEnabled: Bool = true

    // Preview
    var showContent: Bool = true
    var showSender: Bool = true
    var showPersonaName: Bool = true
}

// MARK: - Per-Persona Feature Settings

struct PersonaFeatureSettings: Codable {
    var voiceCalls: Bool = true
    var videoCalls: Bool = true
    var screenSharing: Bool = false
    var voiceMessages: Bool = true
    var fileSharing: Bool = true
    var locationSharing: Bool = false
    var disappearingMessages: Bool = false
    var scheduledMessages: Bool = true
    var silentMessages: Bool = true
    var maxGroupSize: Int = 256
    var maxFileSizeMB: Int = 100
    var voiceMessageDurationSeconds: Int = 300
}

// MARK: - Persona Keys (HD-derived)

struct PersonaKeys: Codable {
    let signingKey: String
    let encryptionKey: String
    let derivationPath: String
}

// MARK: - Persona Badge

struct PersonaBadge: Codable, Identifiable {
    let id: String
    let type: PersonaBadgeType
    let issuedAt: Date
    let issuer: String
    var verifiable: Bool
    var proof: String?

    init(
        id: String = UUID().uuidString,
        type: PersonaBadgeType,
        issuedAt: Date = Date(),
        issuer: String = "",
        verifiable: Bool = false,
        proof: String? = nil
    ) {
        self.id = id
        self.type = type
        self.issuedAt = issuedAt
        self.issuer = issuer
        self.verifiable = verifiable
        self.proof = proof
    }
}

enum PersonaBadgeType: String, Codable, CaseIterable {
    // Professional
    case verifiedEmployer = "verified_employer"
    case professionalCredential = "professional_credential"
    case linkedinVerified = "linkedin_verified"
    case domainEmail = "domain_email"
    // Gaming
    case gameAchievement = "game_achievement"
    case tournamentWinner = "tournament_winner"
    case verifiedGamerTag = "verified_gamer_tag"
    // Creative
    case portfolioVerified = "portfolio_verified"
    case publishedWork = "published_work"
    // Community
    case communityModerator = "community_moderator"
    case trustedContributor = "trusted_contributor"
    case eventOrganizer = "event_organizer"
    // Dating
    case photoVerified = "photo_verified"
    case ageVerified = "age_verified"
    case locationVerified = "location_verified"
}

// MARK: - Persona Credential

struct PersonaCredential: Codable, Identifiable {
    let id: String
    let type: String
    let issuer: String
    let issuedAt: Date
    var expiresAt: Date?
    var isRevoked: Bool
    var proof: String?
}

// MARK: - Persona Deletion State

struct PersonaDeletionState: Codable {
    let deletedAt: Date
    let recoveryExpiresAt: Date?
    var archiveConversations: Bool
    var notifyContacts: Bool
    var isRecoverable: Bool

    var canRecover: Bool {
        guard isRecoverable, let expires = recoveryExpiresAt else { return false }
        return Date() < expires
    }
}

// MARK: - Persona Deletion Options

struct PersonaDeletionOptions {
    var archiveConversations: Bool = true
    var notifyContacts: Bool = false
    var recoveryPeriodDays: Int = 30
    var exportBeforeDelete: Bool = false
}

// MARK: - Conversation Isolation

struct PersonaConversation: Codable, Identifiable {
    let id: String
    let personaId: String
    let contactId: String
    let contactPersonaId: String?
    var lastMessageAt: Date?
    var unreadCount: Int
    var messageCount: Int
}

// MARK: - Persona Switch Context

struct PersonaSwitchContext {
    let fromPersonaId: String?
    let toPersonaId: String
    let contactId: String
    var contactKnowsLink: Bool
    var requiresConfirmation: Bool
    var warningMessage: String?
}

// MARK: - Trust-Level Persona Limits

struct PersonaLimits {
    let maxPersonas: Int
    let allowCustomCategories: Bool
    let maxBadgeSlots: Int

    static func forTrustLevel(_ level: String) -> PersonaLimits {
        switch level.lowercased() {
        case "unverified":
            return PersonaLimits(maxPersonas: 2, allowCustomCategories: false, maxBadgeSlots: 1)
        case "newcomer":
            return PersonaLimits(maxPersonas: 3, allowCustomCategories: false, maxBadgeSlots: 2)
        case "member", "basic":
            return PersonaLimits(maxPersonas: 5, allowCustomCategories: true, maxBadgeSlots: 3)
        case "trusted":
            return PersonaLimits(maxPersonas: 7, allowCustomCategories: true, maxBadgeSlots: 5)
        case "verified", "elite":
            return PersonaLimits(maxPersonas: 10, allowCustomCategories: true, maxBadgeSlots: .max)
        default:
            return PersonaLimits(maxPersonas: 2, allowCustomCategories: false, maxBadgeSlots: 1)
        }
    }
}

// MARK: - Visibility Matrix Entry

struct VisibilityMatrixEntry: Identifiable {
    let id: String
    let contactId: String
    let contactName: String
    var personaVisibility: [String: Bool]
}

// MARK: - Enhanced Settings Models

struct EnhancedNotificationSettings: Codable {
    var messageNotifications: Bool = true
    var showPreviews: String = "always"
    var messageSound: String = "Echo Default"
    var groupNotifications: Bool = true
    var mentionsOnly: Bool = false
    var callNotifications: Bool = true
    var ringSound: String = "Reflection"
    var contactRequests: Bool = true
    var trustScoreChanges: Bool = true
    var rewardNotifications: Bool = true
    var quietHoursEnabled: Bool = false
    var quietHoursFrom: String = "22:00"
    var quietHoursTo: String = "07:00"
    var allowInnerCircleCalls: Bool = true
}

struct EnhancedPrivacySettings: Codable {
    var findByUsername: String = "everyone"
    var showOnlineStatus: String = "contacts"
    var showTrustScore: String = "everyone"
    var whoCanMessage: String = "contacts"
    var readReceipts: Bool = true
    var typingIndicators: Bool = true
    var whoCanCall: String = "trusted"
    var anchorMessagesByDefault: Bool = false
    var showVerificationBadges: Bool = true
    var screenLockTimeout: String = "immediately"
    var hideMessagePreviews: Bool = true
    var screenshotNotifications: Bool = true
}

struct AppearanceSettings: Codable {
    var theme: String = "light"
    var accentColor: String = "indigo"
    var chatWallpaper: String = "default"
    var messageCorners: String = "rounded"
    var fontSize: String = "medium"
    var appIcon: String = "default"
}

struct StorageInfo: Codable {
    var totalUsedBytes: Int64 = 0
    var totalCapacityBytes: Int64 = 0
    var photosVideosBytes: Int64 = 0
    var documentsBytes: Int64 = 0
    var voiceMessagesBytes: Int64 = 0
    var otherBytes: Int64 = 0
    var cacheBytes: Int64 = 0
    var autoDownload: String = "wifi"
    var mediaQuality: String = "standard"
    var keepMedia: String = "forever"
    var useLessDataForCalls: Bool = false
    var autoBackup: String = "daily"
    var includeMediaInBackup: Bool = true
    var lastBackupDate: Date?
}

struct AccountInfo: Codable {
    var phone: String = ""
    var email: String?
    var did: String?
    var passkeyCount: Int = 0
    var twoFactorEnabled: Bool = false
    var activeSessionCount: Int = 0
    var recoveryPhraseSetUp: Bool = false
    var trustedRecoveryContactCount: Int = 0
}

// MARK: - Web of Trust Model

struct WebOfTrustAttestation: Codable, Identifiable {
    let id: String
    let attester: String
    let attestee: String
    let type: AttestationType
    let confidence: Int
    let points: Int
    let message: String?
    let createdAt: Date
    let expiresAt: Date
}

enum AttestationType: String, Codable {
    case vouch, endorse, verify
}

// MARK: - AnyCodable (for flexible JSON)

enum AnyCodable: Codable {
    case null
    case bool(Bool)
    case int(Int)
    case double(Double)
    case string(String)
    case array([AnyCodable])
    case object([String: AnyCodable])
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if container.decodeNil() {
            self = .null
        } else if let bool = try? container.decode(Bool.self) {
            self = .bool(bool)
        } else if let int = try? container.decode(Int.self) {
            self = .int(int)
        } else if let double = try? container.decode(Double.self) {
            self = .double(double)
        } else if let string = try? container.decode(String.self) {
            self = .string(string)
        } else if let array = try? container.decode([AnyCodable].self) {
            self = .array(array)
        } else if let object = try? container.decode([String: AnyCodable].self) {
            self = .object(object)
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode AnyCodable")
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch self {
        case .null:
            try container.encodeNil()
        case .bool(let value):
            try container.encode(value)
        case .int(let value):
            try container.encode(value)
        case .double(let value):
            try container.encode(value)
        case .string(let value):
            try container.encode(value)
        case .array(let value):
            try container.encode(value)
        case .object(let value):
            try container.encode(value)
        }
    }
}

// MARK: - Request Models

struct VerifyBiometricRequest: Codable {
    let biometricData: String
    let challenge: String
}

struct CreateCredentialRequest: Codable {
    let type: String
    let subject: String
    let claims: [String: AnyCodable]
    let expiresIn: Int?
}

struct ShareCredentialRequest: Codable {
    let credentialId: String
    let recipientDid: String
    let verificationRequired: Bool
}

struct SendTokensRequest: Codable {
    let recipientId: String
    let amount: Decimal
    let message: String?
}

struct StakeTokensRequest: Codable {
    let amount: Decimal
    let duration: Int
    let validator: String?
}

// MARK: - Error Models

struct ErrorResponse: Codable {
    let code: String
    let message: String
    let details: [String: String]?
    let timestamp: Date
}

// MARK: - Pagination

struct PaginatedResponse<T: Codable>: Codable {
    let data: [T]
    let pagination: Pagination
}

struct Pagination: Codable {
    let current: Int
    let total: Int
    let limit: Int
    let offset: Int
}

// MARK: - Pinned Item Model

public struct PinnedItem: Identifiable {
    public let id: String
    public let type: PinnedItemType
    public let name: String
    public let avatar: String?
    public let initials: String
    public let gradientIndex: Int
    public var isOnline: Bool
    public var unreadCount: Int

    public enum PinnedItemType {
        case contact
        case group
    }

    public init(
        id: String,
        type: PinnedItemType,
        name: String,
        avatar: String? = nil,
        initials: String,
        gradientIndex: Int,
        isOnline: Bool = false,
        unreadCount: Int = 0
    ) {
        self.id = id
        self.type = type
        self.name = name
        self.avatar = avatar
        self.initials = initials
        self.gradientIndex = gradientIndex
        self.isOnline = isOnline
        self.unreadCount = unreadCount
    }
}
