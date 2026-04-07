import SwiftUI
import Combine

/// Auth ViewModel - Manages authentication flow
@Observable
 class AuthViewModel: NSObject {
    // MARK: - Published State
    
    @ObservationIgnored @Published var isLoading = false
    @ObservationIgnored @Published var errorMessage: String?
    @ObservationIgnored @Published var authState: AuthState = .welcome
    @ObservationIgnored @Published var phone = ""
    @ObservationIgnored @Published var otp = ""
    @ObservationIgnored @Published var isAuthenticated = false
    
    // MARK: - Dependencies
    
    private let authService: AuthServiceProtocol
    private let keychainManager: TokenKeychainProtocol
    private var cancellables = Set<AnyCancellable>()
    
     enum AuthState {
        case welcome
        case phoneEntry
        case otpVerification
        case passkeySetup
        case profileSetup
        case complete
    }
    
     init(
        authService: AuthServiceProtocol,
        keychainManager: TokenKeychainProtocol
    ) {
        self.authService = authService
        self.keychainManager = keychainManager
        super.init()
    }
    
    // MARK: - Public Methods
    
     func requestOTP(phone: String) {
        isLoading = true
        errorMessage = nil
        
        Task {
            do {
                let response = try await authService.requestOTP(phone: phone)
                await MainActor.run {
                    self.phone = phone
                    self.authState = .otpVerification
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
    
     func verifyOTP(_ code: String) {
        isLoading = true
        errorMessage = nil
        
        Task {
            do {
                let response = try await authService.verifyOTP(phone: phone, code: code)
                try keychainManager.saveToken(response.token)
                await MainActor.run {
                    self.otp = code
                    self.authState = .passkeySetup
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
    
     func setupPasskey() {
        isLoading = true
        
        Task {
            do {
                try await authService.registerPasskey()
                await MainActor.run {
                    self.authState = .profileSetup
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
    
     func signOut() {
        keychainManager.clearAll()
        isAuthenticated = false
        authState = .welcome
        phone = ""
        otp = ""
    }
}

// MARK: - Protocol Definition

 protocol AuthServiceProtocol {
    func requestOTP(phone: String) async throws -> OTPResponse
    func verifyOTP(phone: String, code: String) async throws -> AuthResponse
    func registerPasskey() async throws
    func authenticateWithPasskey() async throws -> AuthResponse
    func refreshToken() async throws -> String
}

 struct OTPResponse: Codable {
     let expiresIn: Int
     let phone: String
}

 struct AuthResponse: Codable {
     let token: String
     let refreshToken: String
     let did: String
     let user: UserProfile
}

 struct UserProfile: Codable {
     let id: String
     let phone: String
     let displayName: String?
     let username: String?
     let avatarURL: String?
}

// MARK: - Profile ViewModel

 protocol ProfileServiceProtocol {
    func fetchProfile() async throws -> ProfileData
    func updateProfile(_ profile: ProfileData) async throws -> ProfileData
    func checkUsernameAvailability(_ username: String) async throws -> Bool
    func uploadAvatar(data: Data) async throws -> String
    func fetchPersonas() async throws -> [Persona]
    func createPersona(_ persona: Persona) async throws -> Persona
    func updatePersona(_ persona: Persona) async throws -> Persona
    func deletePersona(id: String, options: PersonaDeletionOptions) async throws
    func recoverPersona(id: String) async throws -> Persona
    func setDefaultPersona(id: String) async throws
    func fetchSettings() async throws -> ProfileSettings
    func updateNotificationSettings(_ settings: EnhancedNotificationSettings) async throws
    func updatePrivacySettings(_ settings: EnhancedPrivacySettings) async throws
    func updateAppearanceSettings(_ settings: AppearanceSettings) async throws
    func updatePersonaPrivacySettings(personaId: String, settings: PersonaPrivacySettings) async throws
    func updatePersonaNotificationSettings(personaId: String, settings: PersonaNotificationSettings) async throws
    func updatePersonaFeatureSettings(personaId: String, settings: PersonaFeatureSettings) async throws
    func fetchStorageInfo() async throws -> StorageInfo
    func clearCache() async throws
    func backUpNow() async throws
    func deleteAccount() async throws
    // Access grants
    func grantAccess(personaId: String, contactId: String, permissions: AccessPermissions) async throws -> AccessGrant
    func revokeAccess(grantId: String) async throws
    func fetchVisibilityMatrix() async throws -> [VisibilityMatrixEntry]
    // Persona switching
    func validatePersonaSwitch(from: String?, to: String, contactId: String) async throws -> PersonaSwitchContext
    // Persona conversations
    func fetchPersonaConversations(personaId: String) async throws -> [PersonaConversation]
    // Badges
    func addPersonaBadge(personaId: String, badge: PersonaBadge) async throws -> PersonaBadge
    func removePersonaBadge(personaId: String, badgeId: String) async throws
    // Export
    func exportPersonaData(personaId: String) async throws -> URL
}

 struct ProfileData {
     var displayName: String
     var username: String
     var bio: String
     var status: String
     var avatarURL: String?
     var website: String?
     var links: [String]
     var trustScore: Int
     var trustLevel: String
     var isVerified: Bool
     var messagesSent: Int
     var contactsCount: Int
     var echoRewards: Double

     init(
        displayName: String = "",
        username: String = "",
        bio: String = "",
        status: String = "",
        avatarURL: String? = nil,
        website: String? = nil,
        links: [String] = [],
        trustScore: Int = 0,
        trustLevel: String = "Newcomer",
        isVerified: Bool = false,
        messagesSent: Int = 0,
        contactsCount: Int = 0,
        echoRewards: Double = 0
    ) {
        self.displayName = displayName
        self.username = username
        self.bio = bio
        self.status = status
        self.avatarURL = avatarURL
        self.website = website
        self.links = links
        self.trustScore = trustScore
        self.trustLevel = trustLevel
        self.isVerified = isVerified
        self.messagesSent = messagesSent
        self.contactsCount = contactsCount
        self.echoRewards = echoRewards
    }
}

 struct ProfileSettings {
     var notifications: EnhancedNotificationSettings
     var privacy: EnhancedPrivacySettings
     var appearance: AppearanceSettings
     var account: AccountInfo

     init(
        notifications: EnhancedNotificationSettings = .init(),
        privacy: EnhancedPrivacySettings = .init(),
        appearance: AppearanceSettings = .init(),
        account: AccountInfo = .init()
    ) {
        self.notifications = notifications
        self.privacy = privacy
        self.appearance = appearance
        self.account = account
    }
}

@Observable
 class ProfileViewModel: NSObject {
    // MARK: - Profile State
    @ObservationIgnored @Published var profile = ProfileData()
    @ObservationIgnored @Published var personas: [Persona] = []
    @ObservationIgnored @Published var settings = ProfileSettings()
    @ObservationIgnored @Published var storageInfo = StorageInfo()
    @ObservationIgnored @Published var isLoading = false
    @ObservationIgnored @Published var isSaving = false
    @ObservationIgnored @Published var errorMessage: String?
    @ObservationIgnored @Published var successMessage: String?

    // MARK: - Edit Profile State
    @ObservationIgnored @Published var editDisplayName = ""
    @ObservationIgnored @Published var editUsername = ""
    @ObservationIgnored @Published var editBio = ""
    @ObservationIgnored @Published var editStatus = ""
    @ObservationIgnored @Published var editWebsite = ""
    @ObservationIgnored @Published var isUsernameAvailable: Bool?
    @ObservationIgnored @Published var isCheckingUsername = false

    // MARK: - Persona Edit State
    @ObservationIgnored @Published var editingPersona: Persona?

    // MARK: - Persona Switching State
    @ObservationIgnored @Published var activePersonaId: String?
    @ObservationIgnored @Published var switchContext: PersonaSwitchContext?
    @ObservationIgnored @Published var showSwitchWarning = false

    // MARK: - Visibility Matrix State
    @ObservationIgnored @Published var visibilityMatrix: [VisibilityMatrixEntry] = []

    // MARK: - Persona Conversations
    @ObservationIgnored @Published var personaConversations: [String: [PersonaConversation]] = [:]

    private let profileService: ProfileServiceProtocol
    private var cancellables = Set<AnyCancellable>()

    static let maxBioLength = 500

     init(profileService: ProfileServiceProtocol) {
        self.profileService = profileService
        super.init()
    }

    // MARK: - Trust-Level Persona Limits

     var personaLimits: PersonaLimits {
        PersonaLimits.forTrustLevel(profile.trustLevel)
    }

     var maxPersonas: Int {
        personaLimits.maxPersonas
    }

    // MARK: - Profile Methods

     func fetchProfile() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                let data = try await profileService.fetchProfile()
                await MainActor.run {
                    self.profile = data
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }

     func beginEditProfile() {
        editDisplayName = profile.displayName
        editUsername = profile.username
        editBio = profile.bio
        editStatus = profile.status
        editWebsite = profile.website ?? ""
        isUsernameAvailable = nil
    }

     func checkUsernameAvailability() {
        let username = editUsername
        guard !username.isEmpty, username != profile.username else {
            isUsernameAvailable = nil
            return
        }
        isCheckingUsername = true

        Task {
            do {
                let available = try await profileService.checkUsernameAvailability(username)
                await MainActor.run {
                    self.isUsernameAvailable = available
                    self.isCheckingUsername = false
                }
            } catch {
                await MainActor.run {
                    self.isUsernameAvailable = nil
                    self.isCheckingUsername = false
                }
            }
        }
    }

     func saveProfile() {
        isSaving = true
        errorMessage = nil

        var updated = profile
        updated.displayName = editDisplayName
        updated.username = editUsername
        updated.bio = editBio
        updated.status = editStatus
        updated.website = editWebsite.isEmpty ? nil : editWebsite

        Task {
            do {
                let saved = try await profileService.updateProfile(updated)
                await MainActor.run {
                    self.profile = saved
                    self.isSaving = false
                    self.successMessage = "Profile updated"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isSaving = false
                }
            }
        }
    }

    // MARK: - Persona Methods

     func fetchPersonas() {
        Task {
            do {
                let result = try await profileService.fetchPersonas()
                await MainActor.run {
                    self.personas = result
                    if self.activePersonaId == nil, let defaultPersona = result.first(where: { $0.isDefault }) {
                        self.activePersonaId = defaultPersona.id
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     var canCreatePersona: Bool {
        personas.count < maxPersonas
    }

     var remainingPersonaSlots: Int {
        max(0, maxPersonas - personas.count)
    }

     var activePersona: Persona? {
        guard let id = activePersonaId else { return nil }
        return personas.first { $0.id == id }
    }

     func createPersona(_ persona: Persona) {
        guard canCreatePersona else {
            errorMessage = "You've reached the maximum number of personas for your trust level."
            return
        }

        // Check custom category permission
        if persona.type == .custom && !personaLimits.allowCustomCategories {
            errorMessage = "Custom categories require a higher trust level."
            return
        }

        isSaving = true

        Task {
            do {
                let created = try await profileService.createPersona(persona)
                await MainActor.run {
                    self.personas.append(created)
                    self.isSaving = false
                    self.successMessage = "Persona created"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isSaving = false
                }
            }
        }
    }

     func updatePersona(_ persona: Persona) {
        isSaving = true

        Task {
            do {
                let updated = try await profileService.updatePersona(persona)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == updated.id }) {
                        self.personas[index] = updated
                    }
                    self.isSaving = false
                    self.successMessage = "Persona updated"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isSaving = false
                }
            }
        }
    }

     func deletePersona(id: String, options: PersonaDeletionOptions = PersonaDeletionOptions()) {
        guard personas.count > 1 else {
            errorMessage = "You must have at least one persona."
            return
        }

        Task {
            do {
                try await profileService.deletePersona(id: id, options: options)
                await MainActor.run {
                    if options.recoveryPeriodDays > 0 {
                        // Soft delete - mark as deleted but keep in list
                        if let index = self.personas.firstIndex(where: { $0.id == id }) {
                            self.personas[index].deletionState = PersonaDeletionState(
                                deletedAt: Date(),
                                recoveryExpiresAt: Calendar.current.date(byAdding: .day, value: options.recoveryPeriodDays, to: Date()),
                                archiveConversations: options.archiveConversations,
                                notifyContacts: options.notifyContacts,
                                isRecoverable: true
                            )
                        }
                    } else {
                        self.personas.removeAll { $0.id == id }
                    }
                    // Adjust active persona if needed
                    if self.activePersonaId == id {
                        self.activePersonaId = self.personas.first(where: { $0.isDefault })?.id ?? self.personas.first?.id
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func recoverPersona(id: String) {
        Task {
            do {
                let recovered = try await profileService.recoverPersona(id: id)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == id }) {
                        self.personas[index] = recovered
                    }
                    self.successMessage = "Persona restored. You may need to re-grant access to contacts."
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func setDefaultPersona(id: String) {
        Task {
            do {
                try await profileService.setDefaultPersona(id: id)
                await MainActor.run {
                    for i in self.personas.indices {
                        self.personas[i].isDefault = (self.personas[i].id == id)
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Persona Switching

     func switchPersona(to personaId: String, forContact contactId: String) {
        Task {
            do {
                let context = try await profileService.validatePersonaSwitch(
                    from: activePersonaId,
                    to: personaId,
                    contactId: contactId
                )
                await MainActor.run {
                    if context.requiresConfirmation {
                        self.switchContext = context
                        self.showSwitchWarning = true
                    } else {
                        self.activePersonaId = personaId
                        self.updatePersonaLastActive(personaId)
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func confirmPersonaSwitch() {
        guard let context = switchContext else { return }
        activePersonaId = context.toPersonaId
        updatePersonaLastActive(context.toPersonaId)
        showSwitchWarning = false
        switchContext = nil
    }

     func cancelPersonaSwitch() {
        showSwitchWarning = false
        switchContext = nil
    }

    private func updatePersonaLastActive(_ personaId: String) {
        if let index = personas.firstIndex(where: { $0.id == personaId }) {
            personas[index].lastActiveAt = Date()
        }
    }

    // MARK: - Visibility Matrix

     func fetchVisibilityMatrix() {
        Task {
            do {
                let matrix = try await profileService.fetchVisibilityMatrix()
                await MainActor.run {
                    self.visibilityMatrix = matrix
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Access Grants

     func grantAccess(personaId: String, contactId: String, permissions: AccessPermissions = .init()) {
        Task {
            do {
                let grant = try await profileService.grantAccess(personaId: personaId, contactId: contactId, permissions: permissions)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].accessGrants.append(grant)
                        if !self.personas[index].selectedContactIds.contains(contactId) {
                            self.personas[index].selectedContactIds.append(contactId)
                        }
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func revokeAccess(personaId: String, grantId: String) {
        Task {
            do {
                try await profileService.revokeAccess(grantId: grantId)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].accessGrants.removeAll { $0.id == grantId }
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Persona Conversations

     func fetchPersonaConversations(personaId: String) {
        Task {
            do {
                let convos = try await profileService.fetchPersonaConversations(personaId: personaId)
                await MainActor.run {
                    self.personaConversations[personaId] = convos
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Per-Persona Settings

     func updatePersonaPrivacySettings(personaId: String, settings: PersonaPrivacySettings) {
        Task {
            do {
                try await profileService.updatePersonaPrivacySettings(personaId: personaId, settings: settings)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].privacySettings = settings
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func updatePersonaNotificationSettings(personaId: String, settings: PersonaNotificationSettings) {
        Task {
            do {
                try await profileService.updatePersonaNotificationSettings(personaId: personaId, settings: settings)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].notificationSettings = settings
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func updatePersonaFeatureSettings(personaId: String, settings: PersonaFeatureSettings) {
        Task {
            do {
                try await profileService.updatePersonaFeatureSettings(personaId: personaId, settings: settings)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].featureSettings = settings
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Badges

     func addBadge(personaId: String, badge: PersonaBadge) {
        Task {
            do {
                let added = try await profileService.addPersonaBadge(personaId: personaId, badge: badge)
                await MainActor.run {
                    if let index = self.personas.firstIndex(where: { $0.id == personaId }) {
                        self.personas[index].badges.append(added)
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Export

     func exportPersonaData(personaId: String) {
        isLoading = true
        Task {
            do {
                let _ = try await profileService.exportPersonaData(personaId: personaId)
                await MainActor.run {
                    self.isLoading = false
                    self.successMessage = "Persona data exported"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }

    // MARK: - Settings Methods

     func fetchSettings() {
        Task {
            do {
                let result = try await profileService.fetchSettings()
                await MainActor.run {
                    self.settings = result
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func updateNotificationSettings(_ settings: EnhancedNotificationSettings) {
        Task {
            do {
                try await profileService.updateNotificationSettings(settings)
                await MainActor.run {
                    self.settings.notifications = settings
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func updatePrivacySettings(_ settings: EnhancedPrivacySettings) {
        Task {
            do {
                try await profileService.updatePrivacySettings(settings)
                await MainActor.run {
                    self.settings.privacy = settings
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func updateAppearanceSettings(_ settings: AppearanceSettings, themeManager: ThemeManager? = nil) {
        Task {
            do {
                try await profileService.updateAppearanceSettings(settings)
                await MainActor.run {
                    self.settings.appearance = settings
                    // Sync theme selection with ThemeManager
                    if let themeManager = themeManager,
                       let themeType = ThemeType(rawValue: settings.accentColor) {
                        themeManager.setTheme(themeType)
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func fetchStorageInfo() {
        Task {
            do {
                let info = try await profileService.fetchStorageInfo()
                await MainActor.run {
                    self.storageInfo = info
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

     func clearCache() {
        isLoading = true
        Task {
            do {
                try await profileService.clearCache()
                await MainActor.run {
                    self.storageInfo.cacheBytes = 0
                    self.isLoading = false
                    self.successMessage = "Cache cleared"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }

     func backUpNow() {
        isLoading = true
        Task {
            do {
                try await profileService.backUpNow()
                await MainActor.run {
                    self.storageInfo.lastBackupDate = Date()
                    self.isLoading = false
                    self.successMessage = "Backup complete"
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }

     func deleteAccount() {
        isLoading = true
        Task {
            do {
                try await profileService.deleteAccount()
                await MainActor.run {
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
}

// MARK: - Messaging ViewModel

@Observable
 class MessagingViewModel: NSObject {
    @ObservationIgnored @Published var conversations: [ConversationModel] = []
    @ObservationIgnored @Published var selectedConversation: ConversationModel?
    @ObservationIgnored @Published var messages: [MessageModel] = []
    @ObservationIgnored @Published var isLoading = false
    @ObservationIgnored @Published var errorMessage: String?
    
    private let messagingService: MessagingServiceProtocol
    private var cancellables = Set<AnyCancellable>()
    
     init(messagingService: MessagingServiceProtocol) {
        self.messagingService = messagingService
        super.init()
    }
    
     func fetchConversations() {
        isLoading = true
        
        Task {
            do {
                let conversations = try await messagingService.fetchConversations()
                await MainActor.run {
                    self.conversations = conversations
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
    
     func selectConversation(_ conversation: ConversationModel) {
        selectedConversation = conversation
        fetchMessages(for: conversation.id)
    }
    
     func fetchMessages(for conversationId: String) {
        Task {
            do {
                let messages = try await messagingService.fetchMessages(conversationId: conversationId)
                await MainActor.run {
                    self.messages = messages
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }
    
     func sendMessage(_ content: String, to conversationId: String) {
        Task {
            do {
                try await messagingService.sendMessage(content, to: conversationId)
                await MainActor.run {
                    self.fetchMessages(for: conversationId)
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }
}

 protocol MessagingServiceProtocol {
    func fetchConversations() async throws -> [ConversationModel]
    func fetchMessages(conversationId: String) async throws -> [MessageModel]
    func sendMessage(_ content: String, to conversationId: String) async throws
    func markAsRead(conversationId: String) async throws
}

 struct ConversationModel: Identifiable {
     let id: String
     let participantId: String
     let participantName: String
     var lastMessage: String?
     var unreadCount: Int
     var updatedAt: Date
}

 struct MessageModel: Identifiable {
     let id: String
     let conversationId: String
     let senderId: String
     let content: String
     let status: MessageStatus
     let createdAt: Date
}

// MARK: - Trust ViewModel

@Observable
 class TrustViewModel: NSObject {
    @ObservationIgnored @Published var trustScore = 0
    @ObservationIgnored @Published var trustLevel = "Newcomer"
    @ObservationIgnored @Published var breakdown: TrustBreakdown?
    @ObservationIgnored @Published var isLoading = false
    @ObservationIgnored @Published var errorMessage: String?
    
    private let trustService: TrustServiceProtocol
    
     init(trustService: TrustServiceProtocol) {
        self.trustService = trustService
        super.init()
    }
    
     func fetchTrustScore(userId: String) {
        isLoading = true
        
        Task {
            do {
                let result = try await trustService.fetchTrustScore(userId: userId)
                await MainActor.run {
                    self.trustScore = result.score
                    self.trustLevel = result.level
                    self.breakdown = result.breakdown
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
    
     func submitVerification(documents: [URL], selfie: URL) {
        isLoading = true
        
        Task {
            do {
                try await trustService.submitVerification(documents: documents, selfie: selfie)
                await MainActor.run {
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
}

 protocol TrustServiceProtocol {
    func fetchTrustScore(userId: String) async throws -> TrustScoreResult
    func submitVerification(documents: [URL], selfie: URL) async throws
    func updateTrustCircle(contactId: String, tier: String) async throws
}

 struct TrustScoreResult {
     let score: Int
     let level: String
     let breakdown: TrustBreakdown
}

 struct TrustBreakdown {
     let identity: Int
     let behavior: Int
     let network: Int
     let activity: Int
}

// MARK: - Rewards ViewModel

@Observable
 class RewardsViewModel: NSObject {
    @ObservationIgnored @Published var tokenBalance = 0.0
    @ObservationIgnored @Published var activities: [RewardActivityModel] = []
    @ObservationIgnored @Published var isLoading = false
    @ObservationIgnored @Published var errorMessage: String?
    
    private let rewardsService: RewardsServiceProtocol
    
     init(rewardsService: RewardsServiceProtocol) {
        self.rewardsService = rewardsService
        super.init()
    }
    
     func fetchBalance() {
        Task {
            do {
                let balance = try await rewardsService.fetchBalance()
                await MainActor.run {
                    self.tokenBalance = balance
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }
    
     func fetchActivity() {
        Task {
            do {
                let activities = try await rewardsService.fetchActivity()
                await MainActor.run {
                    self.activities = activities
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }
    
     func claimRewards() {
        isLoading = true
        
        Task {
            do {
                let newBalance = try await rewardsService.claimRewards()
                await MainActor.run {
                    self.tokenBalance = newBalance
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.isLoading = false
                }
            }
        }
    }
}

 protocol RewardsServiceProtocol {
    func fetchBalance() async throws -> Double
    func fetchActivity() async throws -> [RewardActivityModel]
    func stakeTokens(amount: Double, period: Int) async throws
    func claimRewards() async throws -> Double
}

 struct RewardActivityModel: Identifiable {
     let id: String
     let type: String
     let amount: Double
     let description: String
     let date: Date
}
