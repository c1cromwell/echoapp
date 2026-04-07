import Foundation

// MARK: - Authentication Endpoints

enum AuthEndpoint: APIEndpoint {
    case register
    case login
    case refreshToken
    case logout
    case verifyBiometric
    case createPasskey
    case verifyPasskey
    
    var path: String {
        switch self {
        case .register:
            return "/auth/register"
        case .login:
            return "/auth/login"
        case .refreshToken:
            return "/auth/refresh"
        case .logout:
            return "/auth/logout"
        case .verifyBiometric:
            return "/auth/verify-biometric"
        case .createPasskey:
            return "/auth/passkey/create"
        case .verifyPasskey:
            return "/auth/passkey/verify"
        }
    }
}

// MARK: - Message Endpoints

enum MessageEndpoint: APIEndpoint {
    case send
    case fetch(conversationId: String, limit: Int = 50, offset: Int = 0)
    case fetchConversations(limit: Int = 20, offset: Int = 0)
    case createConversation
    case getConversation(id: String)
    case searchMessages(query: String)
    case markAsRead(messageId: String)
    case deleteMessage(id: String)
    case editMessage(id: String)
    case addReaction(messageId: String)
    case removeReaction(messageId: String)
    
    var path: String {
        switch self {
        case .send:
            return "/messages/send"
        case .fetch(let conversationId, let limit, let offset):
            return "/messages/conversations/\(conversationId)?limit=\(limit)&offset=\(offset)"
        case .fetchConversations(let limit, let offset):
            return "/messages/conversations?limit=\(limit)&offset=\(offset)"
        case .createConversation:
            return "/messages/conversations"
        case .getConversation(let id):
            return "/messages/conversations/\(id)"
        case .searchMessages(let query):
            return "/messages/search?q=\(query)"
        case .markAsRead(let messageId):
            return "/messages/\(messageId)/read"
        case .deleteMessage(let id):
            return "/messages/\(id)"
        case .editMessage(let id):
            return "/messages/\(id)"
        case .addReaction(let messageId):
            return "/messages/\(messageId)/reactions"
        case .removeReaction(let messageId):
            return "/messages/\(messageId)/reactions"
        }
    }
}

// MARK: - User Endpoints

enum UserEndpoint: APIEndpoint {
    case getProfile
    case updateProfile
    case getUser(id: String)
    case searchUsers(query: String)
    case addContact(id: String)
    case removeContact(id: String)
    case blockUser(id: String)
    case unblockUser(id: String)
    case getContacts(limit: Int = 50, offset: Int = 0)
    case getSettings
    case updateSettings
    case uploadAvatar
    case deleteAccount
    
    var path: String {
        switch self {
        case .getProfile:
            return "/users/profile"
        case .updateProfile:
            return "/users/profile"
        case .getUser(let id):
            return "/users/\(id)"
        case .searchUsers(let query):
            return "/users/search?q=\(query)"
        case .addContact(let id):
            return "/users/contacts/\(id)"
        case .removeContact(let id):
            return "/users/contacts/\(id)"
        case .blockUser(let id):
            return "/users/block/\(id)"
        case .unblockUser(let id):
            return "/users/block/\(id)"
        case .getContacts(let limit, let offset):
            return "/users/contacts?limit=\(limit)&offset=\(offset)"
        case .getSettings:
            return "/users/settings"
        case .updateSettings:
            return "/users/settings"
        case .uploadAvatar:
            return "/users/avatar"
        case .deleteAccount:
            return "/users/account"
        }
    }
}

// MARK: - Identity Endpoints

enum IdentityEndpoint: APIEndpoint {
    case createDID
    case resolveDID(did: String)
    case updateDIDDocument
    case listDIDs
    case verifyIdentity
    case getVerifications
    case addVerification
    case revokeVerification
    case createCredential
    case shareCredential
    case verifyCredential(id: String)
    case revokeCredential(id: String)
    
    var path: String {
        switch self {
        case .createDID:
            return "/identity/did/create"
        case .resolveDID(let did):
            return "/identity/did/\(did)"
        case .updateDIDDocument:
            return "/identity/did/update"
        case .listDIDs:
            return "/identity/did/list"
        case .verifyIdentity:
            return "/identity/verify"
        case .getVerifications:
            return "/identity/verifications"
        case .addVerification:
            return "/identity/verifications"
        case .revokeVerification:
            return "/identity/verifications/revoke"
        case .createCredential:
            return "/identity/credentials/create"
        case .shareCredential:
            return "/identity/credentials/share"
        case .verifyCredential(let id):
            return "/identity/credentials/\(id)/verify"
        case .revokeCredential(let id):
            return "/identity/credentials/\(id)/revoke"
        }
    }
}

// MARK: - Token Endpoints

enum TokenEndpoint: APIEndpoint {
    case getBalance
    case getTransactionHistory(limit: Int = 50, offset: Int = 0)
    case sendTokens
    case stakeTokens
    case unstakeTokens
    case claimRewards
    case getStakingInfo
    case getTrusScore
    case getAchievements
    case getTokenStats
    
    var path: String {
        switch self {
        case .getBalance:
            return "/tokens/balance"
        case .getTransactionHistory(let limit, let offset):
            return "/tokens/history?limit=\(limit)&offset=\(offset)"
        case .sendTokens:
            return "/tokens/send"
        case .stakeTokens:
            return "/tokens/stake"
        case .unstakeTokens:
            return "/tokens/unstake"
        case .claimRewards:
            return "/tokens/rewards/claim"
        case .getStakingInfo:
            return "/tokens/staking"
        case .getTrusScore:
            return "/tokens/trust-score"
        case .getAchievements:
            return "/tokens/achievements"
        case .getTokenStats:
            return "/tokens/stats"
        }
    }
}

// MARK: - Persona Endpoints

enum PersonaEndpoint: APIEndpoint {
    case listPersonas
    case getPersona(id: String)
    case createPersona
    case updatePersona(id: String)
    case deletePersona(id: String)
    case recoverPersona(id: String)
    case setDefaultPersona(id: String)
    case checkUsernameAvailability(username: String)
    // Per-persona settings
    case getPersonaPrivacySettings(personaId: String)
    case updatePersonaPrivacySettings(personaId: String)
    case getPersonaNotificationSettings(personaId: String)
    case updatePersonaNotificationSettings(personaId: String)
    case getPersonaFeatureSettings(personaId: String)
    case updatePersonaFeatureSettings(personaId: String)
    // Access grants
    case grantAccess(personaId: String)
    case revokeAccess(grantId: String)
    case getVisibilityMatrix
    // Persona switching
    case validateSwitch
    // Persona conversations (isolated)
    case getPersonaConversations(personaId: String)
    // Badges
    case addBadge(personaId: String)
    case removeBadge(personaId: String, badgeId: String)
    // Export
    case exportPersonaData(personaId: String)

    var path: String {
        switch self {
        case .listPersonas:
            return "/users/personas"
        case .getPersona(let id):
            return "/users/personas/\(id)"
        case .createPersona:
            return "/users/personas"
        case .updatePersona(let id):
            return "/users/personas/\(id)"
        case .deletePersona(let id):
            return "/users/personas/\(id)"
        case .recoverPersona(let id):
            return "/users/personas/\(id)/recover"
        case .setDefaultPersona(let id):
            return "/users/personas/\(id)/default"
        case .checkUsernameAvailability(let username):
            return "/users/check-username?username=\(username)"
        case .getPersonaPrivacySettings(let id):
            return "/users/personas/\(id)/settings/privacy"
        case .updatePersonaPrivacySettings(let id):
            return "/users/personas/\(id)/settings/privacy"
        case .getPersonaNotificationSettings(let id):
            return "/users/personas/\(id)/settings/notifications"
        case .updatePersonaNotificationSettings(let id):
            return "/users/personas/\(id)/settings/notifications"
        case .getPersonaFeatureSettings(let id):
            return "/users/personas/\(id)/settings/features"
        case .updatePersonaFeatureSettings(let id):
            return "/users/personas/\(id)/settings/features"
        case .grantAccess(let id):
            return "/users/personas/\(id)/access"
        case .revokeAccess(let grantId):
            return "/users/personas/access/\(grantId)"
        case .getVisibilityMatrix:
            return "/users/personas/visibility-matrix"
        case .validateSwitch:
            return "/users/personas/validate-switch"
        case .getPersonaConversations(let id):
            return "/users/personas/\(id)/conversations"
        case .addBadge(let id):
            return "/users/personas/\(id)/badges"
        case .removeBadge(let personaId, let badgeId):
            return "/users/personas/\(personaId)/badges/\(badgeId)"
        case .exportPersonaData(let id):
            return "/users/personas/\(id)/export"
        }
    }
}

// MARK: - WebSocket Endpoints

enum WebSocketEndpoint: APIEndpoint {
    case messages
    case notifications
    case presence
    case typing
    
    var path: String {
        switch self {
        case .messages:
            return "/ws/messages"
        case .notifications:
            return "/ws/notifications"
        case .presence:
            return "/ws/presence"
        case .typing:
            return "/ws/typing"
        }
    }
}

// MARK: - Request/Response Models

struct RegisterRequest: Codable {
    let email: String
    let password: String
    let username: String
    let publicKey: String
    let biometricEnabled: Bool
}

struct LoginRequest: Codable {
    let email: String
    let password: String
}

struct LoginResponse: Codable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: Int
    let user: UserResponse
}

struct RefreshTokenRequest: Codable {
    let refreshToken: String
}

struct RefreshTokenResponse: Codable {
    let accessToken: String
    let expiresIn: Int
}

struct UserResponse: Codable {
    let id: String
    let email: String
    let username: String
    let avatar: String?
    let publicKey: String
    let createdAt: Date
}

struct SendMessageRequest: Codable {
    let conversationId: String
    let content: String
    let encryptedContent: String
    let nonce: String
    let attachments: [AttachmentRequest]?
}

struct AttachmentRequest: Codable {
    let filename: String
    let mimeType: String
    let size: Int
    let url: String
}

struct MessageResponse: Codable {
    let id: String
    let conversationId: String
    let senderId: String
    let content: String
    let encryptedContent: String
    let nonce: String
    let attachments: [AttachmentResponse]?
    let reactions: [ReactionResponse]?
    let timestamp: Date
    let isRead: Bool
}

struct AttachmentResponse: Codable {
    let id: String
    let filename: String
    let mimeType: String
    let size: Int
    let url: String
}

struct ReactionResponse: Codable {
    let id: String
    let messageId: String
    let userId: String
    let emoji: String
    let timestamp: Date
}

struct ConversationResponse: Codable {
    let id: String
    let participants: [ParticipantResponse]
    let name: String?
    let lastMessage: MessageResponse?
    let unreadCount: Int
    let createdAt: Date
    let updatedAt: Date
}

struct ParticipantResponse: Codable {
    let userId: String
    let username: String
    let avatar: String?
    let isOnline: Bool
}

struct CreateDIDRequest: Codable {
    let publicKey: String
    let verificationMethod: String
}

struct DIDResponse: Codable {
    let did: String
    let didDocument: DIDDocument
    let createdAt: Date
}

struct DIDDocument: Codable {
    let id: String
    let publicKey: [PublicKeyEntry]
    let authentication: [String]
    let assertionMethod: [String]
}

struct PublicKeyEntry: Codable {
    let id: String
    let type: String
    let controller: String
    let publicKeyPem: String
}

struct TokenBalanceResponse: Codable {
    let balance: Decimal
    let frozen: Decimal
    let available: Decimal
    let currency: String
}

struct TransactionResponse: Codable {
    let id: String
    let type: String
    let amount: Decimal
    let from: String?
    let to: String?
    let status: String
    let timestamp: Date
    let description: String?
}

struct TrustScoreResponse: Codable {
    let score: Int
    let level: String
    let nextLevelThreshold: Int
    let multiplier: Double
    let components: TrustScoreComponents
}

struct AchievementResponse: Codable {
    let id: String
    let name: String
    let description: String
    let icon: String
    let unlockedAt: Date?
    let category: String
    let level: Int
}

// MARK: - Persona Request/Response Models

struct CreatePersonaRequest: Codable {
    let type: String
    let name: String
    let displayName: String
    let bio: String?
    let useMainAvatar: Bool
    let visibility: String
    let selectedContactIds: [String]
}

struct UpdatePersonaRequest: Codable {
    let type: String?
    let name: String?
    let displayName: String?
    let bio: String?
    let useMainAvatar: Bool?
    let visibility: String?
    let selectedContactIds: [String]?
}

struct PersonaResponse: Codable {
    let id: String
    let type: String
    let name: String
    let displayName: String
    let bio: String?
    let avatarURL: String?
    let useMainAvatar: Bool
    let visibility: String
    let selectedContactIds: [String]
    let isDefault: Bool
    let createdAt: Date
    let updatedAt: Date
}

struct UsernameAvailabilityResponse: Codable {
    let username: String
    let available: Bool
}

struct UpdateProfileRequest: Codable {
    let displayName: String?
    let username: String?
    let bio: String?
    let status: String?
    let website: String?
}

struct ProfileResponse: Codable {
    let displayName: String
    let username: String
    let bio: String
    let status: String
    let avatarURL: String?
    let website: String?
    let links: [String]
    let trustScore: Int
    let trustLevel: String
    let isVerified: Bool
    let messagesSent: Int
    let contactsCount: Int
    let echoRewards: Double
}

struct NotificationSettingsResponse: Codable {
    let messageNotifications: Bool
    let showPreviews: String
    let messageSound: String
    let groupNotifications: Bool
    let mentionsOnly: Bool
    let callNotifications: Bool
    let ringSound: String
    let contactRequests: Bool
    let trustScoreChanges: Bool
    let rewardNotifications: Bool
    let quietHoursEnabled: Bool
    let quietHoursFrom: String
    let quietHoursTo: String
    let allowInnerCircleCalls: Bool
}

struct PrivacySettingsResponse: Codable {
    let findByUsername: String
    let showOnlineStatus: String
    let showTrustScore: String
    let whoCanMessage: String
    let readReceipts: Bool
    let typingIndicators: Bool
    let whoCanCall: String
    let anchorMessagesByDefault: Bool
    let showVerificationBadges: Bool
    let screenLockTimeout: String
    let hideMessagePreviews: Bool
    let screenshotNotifications: Bool
}

struct AppearanceSettingsResponse: Codable {
    let theme: String
    let accentColor: String
    let chatWallpaper: String
    let messageCorners: String
    let fontSize: String
    let appIcon: String
}

struct StorageInfoResponse: Codable {
    let totalUsedBytes: Int64
    let totalCapacityBytes: Int64
    let photosVideosBytes: Int64
    let documentsBytes: Int64
    let voiceMessagesBytes: Int64
    let otherBytes: Int64
    let cacheBytes: Int64
    let autoDownload: String
    let mediaQuality: String
    let keepMedia: String
    let useLessDataForCalls: Bool
    let autoBackup: String
    let includeMediaInBackup: Bool
    let lastBackupDate: Date?
}

struct AccountInfoResponse: Codable {
    let phone: String
    let email: String?
    let did: String?
    let passkeyCount: Int
    let twoFactorEnabled: Bool
    let activeSessionCount: Int
    let recoveryPhraseSetUp: Bool
    let trustedRecoveryContactCount: Int
}

// MARK: - Enhanced Persona Request/Response Models

struct DeletePersonaRequest: Codable {
    let archiveConversations: Bool
    let notifyContacts: Bool
    let recoveryPeriodDays: Int
    let exportBeforeDelete: Bool
}

struct PersonaDeletionResponse: Codable {
    let deleted: Bool
    let personaId: String
    let recoverable: Bool
    let recoveryExpiresAt: Date?
}

struct PersonaRecoveryResponse: Codable {
    let recovered: Bool
    let persona: PersonaResponse
}

struct GrantAccessRequest: Codable {
    let contactId: String
    let canView: Bool
    let canMessage: Bool
    let canCall: Bool
    let canSeeOtherPersonas: Bool
    let expiresAt: Date?
}

struct AccessGrantResponse: Codable {
    let id: String
    let contactId: String
    let personaId: String
    let grantedAt: Date
    let grantedBy: String
    let canView: Bool
    let canMessage: Bool
    let canCall: Bool
    let canSeeOtherPersonas: Bool
    let expiresAt: Date?
    let revocable: Bool
}

struct VisibilityMatrixResponse: Codable {
    let entries: [VisibilityMatrixEntryResponse]
}

struct VisibilityMatrixEntryResponse: Codable {
    let contactId: String
    let contactName: String
    let personaVisibility: [String: Bool]
}

struct ValidateSwitchRequest: Codable {
    let fromPersonaId: String?
    let toPersonaId: String
    let contactId: String
}

struct ValidateSwitchResponse: Codable {
    let fromPersonaId: String?
    let toPersonaId: String
    let contactId: String
    let contactKnowsLink: Bool
    let requiresConfirmation: Bool
    let warningMessage: String?
}

struct PersonaConversationResponse: Codable {
    let id: String
    let personaId: String
    let contactId: String
    let contactPersonaId: String?
    let lastMessageAt: Date?
    let unreadCount: Int
    let messageCount: Int
}

struct PersonaPrivacySettingsResponse: Codable {
    let lastSeenVisibility: String
    let onlineStatusVisibility: String
    let profilePictureVisibility: String
    let bioVisibility: String
    let statusMessageVisibility: String
    let whoCanMessage: String
    let whoCanCall: String
    let whoCanAddToGroups: String
    let requireApprovalForContact: Bool
    let sendReadReceipts: Bool
    let sendTypingIndicators: Bool
    let searchable: Bool
    let showInSuggestions: Bool
    let allowContactSharing: Bool
    let allowLinkingDiscovery: Bool
    let showSharedTrustScore: Bool
    let allowCrossPersonaForward: Bool
}

struct PersonaNotificationSettingsResponse: Codable {
    let enabled: Bool
    let quietHoursEnabled: Bool
    let quietHoursStart: String
    let quietHoursEnd: String
    let quietHoursTimezone: String
    let quietHoursAllowExceptions: Bool
    let messagesMode: String
    let callsMode: String
    let groupActivityMode: String
    let contactRequests: Bool
    let soundEnabled: Bool
    let soundId: String
    let vibrationEnabled: Bool
    let showContent: Bool
    let showSender: Bool
    let showPersonaName: Bool
}

struct PersonaFeatureSettingsResponse: Codable {
    let voiceCalls: Bool
    let videoCalls: Bool
    let screenSharing: Bool
    let voiceMessages: Bool
    let fileSharing: Bool
    let locationSharing: Bool
    let disappearingMessages: Bool
    let scheduledMessages: Bool
    let silentMessages: Bool
    let maxGroupSize: Int
    let maxFileSizeMB: Int
    let voiceMessageDurationSeconds: Int
}

struct PersonaBadgeResponse: Codable {
    let id: String
    let type: String
    let issuedAt: Date
    let issuer: String
    let verifiable: Bool
    let proof: String?
}

struct AddBadgeRequest: Codable {
    let type: String
    let issuer: String
    let verifiable: Bool
    let proof: String?
}
