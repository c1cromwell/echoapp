import XCTest
import XCTest
@testable import Echo

// MARK: - Profile ViewModel Tests

final class ProfileViewModelUnitTests: XCTestCase {

    var profileViewModel: ProfileViewModel!
    var mockProfileService: MockProfileService!

    override func setUp() {
        super.setUp()
        mockProfileService = MockProfileService()
        profileViewModel = ProfileViewModel(profileService: mockProfileService)
    }

    override func tearDown() {
        profileViewModel = nil
        mockProfileService = nil
        super.tearDown()
    }

    func testInitialProfileState() {
        XCTAssertEqual(profileViewModel.profile.displayName, "")
        XCTAssertEqual(profileViewModel.profile.username, "")
        XCTAssertTrue(profileViewModel.personas.isEmpty)
        XCTAssertFalse(profileViewModel.isLoading)
        XCTAssertFalse(profileViewModel.isSaving)
        XCTAssertNil(profileViewModel.errorMessage)
        XCTAssertNil(profileViewModel.successMessage)
    }

    func testBeginEditProfile() {
        profileViewModel.profile = ProfileData(
            displayName: "Alex",
            username: "alexecho",
            bio: "Test bio",
            status: "Active",
            website: "https://echo.dev"
        )

        profileViewModel.beginEditProfile()

        XCTAssertEqual(profileViewModel.editDisplayName, "Alex")
        XCTAssertEqual(profileViewModel.editUsername, "alexecho")
        XCTAssertEqual(profileViewModel.editBio, "Test bio")
        XCTAssertEqual(profileViewModel.editStatus, "Active")
        XCTAssertEqual(profileViewModel.editWebsite, "https://echo.dev")
        XCTAssertNil(profileViewModel.isUsernameAvailable)
    }

    func testBeginEditProfileNilWebsite() {
        profileViewModel.profile = ProfileData(
            displayName: "Alex",
            username: "alexecho",
            bio: "",
            status: "",
            website: nil
        )

        profileViewModel.beginEditProfile()

        XCTAssertEqual(profileViewModel.editWebsite, "")
    }

    func testCanCreatePersona() {
        // Default trust level "Newcomer" → maxPersonas = 3
        profileViewModel.profile = ProfileData(trustLevel: "Newcomer")
        XCTAssertTrue(profileViewModel.canCreatePersona)
        XCTAssertEqual(profileViewModel.remainingPersonaSlots, 3)
    }

    func testCanCreatePersonaWithExisting() {
        profileViewModel.profile = ProfileData(trustLevel: "Trusted")
        // Trusted → maxPersonas = 7
        profileViewModel.personas = [
            makePersona(id: "1"),
            makePersona(id: "2"),
            makePersona(id: "3")
        ]

        XCTAssertTrue(profileViewModel.canCreatePersona)
        XCTAssertEqual(profileViewModel.remainingPersonaSlots, 4)
    }

    func testCanCreatePersonaAtLimit() {
        // Unverified → maxPersonas = 2
        profileViewModel.profile = ProfileData(trustLevel: "Unverified")
        profileViewModel.personas = [
            makePersona(id: "0"),
            makePersona(id: "1")
        ]

        XCTAssertFalse(profileViewModel.canCreatePersona)
        XCTAssertEqual(profileViewModel.remainingPersonaSlots, 0)
    }

    func testMaxBioLength() {
        XCTAssertEqual(ProfileViewModel.maxBioLength, 500)
    }

    func testMaxPersonasForTrustLevels() {
        // Test each trust level
        let levels: [(String, Int)] = [
            ("Unverified", 2),
            ("Newcomer", 3),
            ("Member", 5),
            ("Trusted", 7),
            ("Verified", 10),
        ]

        for (level, expected) in levels {
            profileViewModel.profile = ProfileData(trustLevel: level)
            XCTAssertEqual(profileViewModel.maxPersonas, expected, "Trust level '\(level)' should have \(expected) max personas")
        }
    }

    func testPersonaLimitsFactory() {
        let newcomer = PersonaLimits.forTrustLevel("Newcomer")
        XCTAssertEqual(newcomer.maxPersonas, 3)

        let trusted = PersonaLimits.forTrustLevel("Trusted")
        XCTAssertEqual(trusted.maxPersonas, 7)

        let verified = PersonaLimits.forTrustLevel("Verified")
        XCTAssertEqual(verified.maxPersonas, 10)
    }

    func testPersonaSwitchingState() {
        XCTAssertNil(profileViewModel.activePersonaId)
        XCTAssertNil(profileViewModel.switchContext)
        XCTAssertFalse(profileViewModel.showSwitchWarning)
    }

    func testVisibilityMatrixState() {
        XCTAssertTrue(profileViewModel.visibilityMatrix.isEmpty)
    }

    func testPersonaConversationsState() {
        XCTAssertTrue(profileViewModel.personaConversations.isEmpty)
    }

    // MARK: - Helpers

    private func makePersona(id: String, isDefault: Bool = false) -> Persona {
        Persona(
            id: id, type: .personal, name: "Test", displayName: "Test",
            bio: nil, useMainAvatar: true, visibility: .all,
            selectedContactIds: [], isDefault: isDefault,
            createdAt: Date(), updatedAt: Date()
        )
    }
}

// MARK: - Persona Model Tests

final class PersonaModelTests: XCTestCase {

    func testPersonaCategoryAllCases() {
        let types = PersonaCategory.allCases
        XCTAssertEqual(types.count, 8)
        XCTAssertTrue(types.contains(.professional))
        XCTAssertTrue(types.contains(.personal))
        XCTAssertTrue(types.contains(.family))
        XCTAssertTrue(types.contains(.gaming))
        XCTAssertTrue(types.contains(.dating))
        XCTAssertTrue(types.contains(.creative))
        XCTAssertTrue(types.contains(.anonymous))
        XCTAssertTrue(types.contains(.custom))
    }

    func testPersonaTypeAlias() {
        // PersonaType is a typealias for PersonaCategory
        let pType: PersonaType = .professional
        XCTAssertEqual(pType, PersonaCategory.professional)
    }

    func testPersonaCategoryProperties() {
        for type in PersonaCategory.allCases {
            XCTAssertFalse(type.icon.isEmpty, "\(type) icon is empty")
            XCTAssertFalse(type.emoji.isEmpty, "\(type) emoji is empty")
            XCTAssertFalse(type.label.isEmpty, "\(type) label is empty")
            XCTAssertFalse(type.id.isEmpty, "\(type) id is empty")
        }
    }

    func testPersonaCategoryLabels() {
        XCTAssertEqual(PersonaCategory.professional.label, "Pro")
        XCTAssertEqual(PersonaCategory.personal.label, "Personal")
        XCTAssertEqual(PersonaCategory.family.label, "Family")
        XCTAssertEqual(PersonaCategory.gaming.label, "Gaming")
        XCTAssertEqual(PersonaCategory.dating.label, "Dating")
        XCTAssertEqual(PersonaCategory.creative.label, "Creative")
        XCTAssertEqual(PersonaCategory.anonymous.label, "Anonymous")
        XCTAssertEqual(PersonaCategory.custom.label, "Custom")
    }

    func testPersonaCategoryIcons() {
        XCTAssertEqual(PersonaCategory.professional.icon, "briefcase.fill")
        XCTAssertEqual(PersonaCategory.personal.icon, "house.fill")
        XCTAssertEqual(PersonaCategory.family.icon, "figure.2.and.child")
        XCTAssertEqual(PersonaCategory.gaming.icon, "gamecontroller.fill")
        XCTAssertEqual(PersonaCategory.custom.icon, "sparkles")
    }

    func testPersonaCategoryId() {
        XCTAssertEqual(PersonaCategory.professional.id, "professional")
        XCTAssertEqual(PersonaCategory.gaming.id, "gaming")
        XCTAssertEqual(PersonaCategory.dating.id, "dating")
        XCTAssertEqual(PersonaCategory.creative.id, "creative")
        XCTAssertEqual(PersonaCategory.anonymous.id, "anonymous")
    }

    func testPersonaContactCount() {
        let persona = Persona(
            id: "1", type: .professional, name: "Pro", displayName: "Alex",
            bio: nil, useMainAvatar: true, visibility: .selected,
            selectedContactIds: ["c1", "c2", "c3"], isDefault: false,
            createdAt: Date(), updatedAt: Date()
        )
        XCTAssertEqual(persona.contactCount, 3)
    }

    func testPersonaContactCountEmpty() {
        let persona = Persona(
            id: "1", type: .personal, name: "P", displayName: "D",
            bio: nil, useMainAvatar: true, visibility: .all,
            selectedContactIds: [], isDefault: true,
            createdAt: Date(), updatedAt: Date()
        )
        XCTAssertEqual(persona.contactCount, 0)
    }

    func testPersonaWithNewFields() {
        let persona = Persona(
            id: "1", type: .dating, name: "Dating", displayName: "NightOwl",
            username: "nightowl42",
            bio: "Looking for connections",
            useMainAvatar: false, visibility: .selected,
            defaultVisibility: .contacts,
            discoverability: true,
            selectedContactIds: ["c1"],
            isDefault: false,
            createdAt: Date(), updatedAt: Date(),
            messageCount: 42
        )

        XCTAssertEqual(persona.username, "nightowl42")
        XCTAssertEqual(persona.type, .dating)
        XCTAssertEqual(persona.messageCount, 42)
        XCTAssertTrue(persona.discoverability)
        XCTAssertEqual(persona.defaultVisibility, .contacts)
    }

    func testPersonaDefaultSettings() {
        let persona = Persona(
            id: "1", type: .professional, name: "Pro", displayName: "Alex",
            isDefault: true, createdAt: Date(), updatedAt: Date()
        )

        // Verify default per-persona settings are initialized
        XCTAssertNotNil(persona.privacySettings)
        XCTAssertNotNil(persona.notificationSettings)
        XCTAssertNotNil(persona.featureSettings)
        XCTAssertNil(persona.deletionState)
        XCTAssertTrue(persona.badges.isEmpty)
        XCTAssertTrue(persona.accessGrants.isEmpty)
    }

    func testPersonaCodable() throws {
        let persona = Persona(
            id: "test-1", type: .gaming, name: "Gaming", displayName: "NightOwl42",
            bio: "Top 500", avatarURL: nil, useMainAvatar: false, visibility: .selected,
            selectedContactIds: ["c1"], isDefault: true,
            createdAt: Date(), updatedAt: Date()
        )

        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        let data = try encoder.encode(persona)

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let decoded = try decoder.decode(Persona.self, from: data)

        XCTAssertEqual(decoded.id, persona.id)
        XCTAssertEqual(decoded.type, .gaming)
        XCTAssertEqual(decoded.name, "Gaming")
        XCTAssertEqual(decoded.displayName, "NightOwl42")
        XCTAssertEqual(decoded.bio, "Top 500")
        XCTAssertNil(decoded.avatarURL)
        XCTAssertFalse(decoded.useMainAvatar)
        XCTAssertTrue(decoded.isDefault)
        XCTAssertEqual(decoded.selectedContactIds, ["c1"])
        XCTAssertEqual(decoded.visibility, .selected)
    }

    func testPersonaVisibilityCodable() throws {
        let visibilities: [PersonaVisibility] = [.all, .selected, .hidden]
        let encoder = JSONEncoder()
        let decoder = JSONDecoder()

        for vis in visibilities {
            let data = try encoder.encode(vis)
            let decoded = try decoder.decode(PersonaVisibility.self, from: data)
            XCTAssertEqual(decoded, vis)
        }
    }

    func testPersonaCategoryRawValues() {
        XCTAssertEqual(PersonaCategory.professional.rawValue, "professional")
        XCTAssertEqual(PersonaCategory.personal.rawValue, "personal")
        XCTAssertEqual(PersonaCategory.family.rawValue, "family")
        XCTAssertEqual(PersonaCategory.gaming.rawValue, "gaming")
        XCTAssertEqual(PersonaCategory.dating.rawValue, "dating")
        XCTAssertEqual(PersonaCategory.creative.rawValue, "creative")
        XCTAssertEqual(PersonaCategory.anonymous.rawValue, "anonymous")
        XCTAssertEqual(PersonaCategory.custom.rawValue, "custom")
    }

    func testPersonaVisibilityRawValues() {
        XCTAssertEqual(PersonaVisibility.all.rawValue, "all")
        XCTAssertEqual(PersonaVisibility.selected.rawValue, "selected")
        XCTAssertEqual(PersonaVisibility.hidden.rawValue, "hidden")
    }
}

// MARK: - Per-Persona Settings Model Tests

final class PerPersonaSettingsTests: XCTestCase {

    func testPersonaPrivacySettingsDefaults() {
        let settings = PersonaPrivacySettings()
        XCTAssertTrue(settings.sendReadReceipts)
        XCTAssertTrue(settings.sendTypingIndicators)
        XCTAssertTrue(settings.searchable)
    }

    func testPersonaNotificationSettingsDefaults() {
        let settings = PersonaNotificationSettings()
        XCTAssertNotNil(settings)
    }

    func testPersonaFeatureSettingsDefaults() {
        let settings = PersonaFeatureSettings()
        XCTAssertNotNil(settings)
    }

    func testAccessGrantCreation() {
        let grant = AccessGrant(
            contactId: "contact-1",
            personaId: "persona-1",
            grantedAt: Date(),
            grantedBy: "persona-1",
            permissions: AccessPermissions(canView: true, canMessage: true, canCall: false, canSeeOtherPersonas: false)
        )

        XCTAssertEqual(grant.contactId, "contact-1")
        XCTAssertEqual(grant.personaId, "persona-1")
        XCTAssertTrue(grant.permissions.canView)
        XCTAssertTrue(grant.permissions.canMessage)
        XCTAssertFalse(grant.permissions.canCall)
        XCTAssertFalse(grant.permissions.canSeeOtherPersonas)
        XCTAssertTrue(grant.revocable)
    }

    func testPersonaBadgeCreation() {
        let badge = PersonaBadge(
            type: .professionalCertified,
            issuedAt: Date(),
            issuer: "EchoVerify",
            verifiable: true,
            proof: "proof-hash"
        )

        XCTAssertEqual(badge.type, .professionalCertified)
        XCTAssertEqual(badge.issuer, "EchoVerify")
        XCTAssertTrue(badge.verifiable)
        XCTAssertFalse(badge.id.isEmpty)
    }

    func testPersonaBadgeTypeAllCases() {
        let types = PersonaBadgeType.allCases
        XCTAssertEqual(types.count, 15)
        XCTAssertTrue(types.contains(.identityVerified))
        XCTAssertTrue(types.contains(.professionalCertified))
        XCTAssertTrue(types.contains(.gamerRank))
        XCTAssertTrue(types.contains(.creativePortfolio))
        XCTAssertTrue(types.contains(.communityModerator))
    }

    func testPersonaDeletionState() {
        let state = PersonaDeletionState(
            deletedAt: Date(),
            recoveryExpiresAt: Date().addingTimeInterval(30 * 24 * 60 * 60),
            archiveConversations: true,
            notifyContacts: false,
            isRecoverable: true
        )

        XCTAssertTrue(state.isRecoverable)
        XCTAssertTrue(state.archiveConversations)
        XCTAssertFalse(state.notifyContacts)
        XCTAssertTrue(state.canRecover)
    }

    func testPersonaDeletionStateExpired() {
        let state = PersonaDeletionState(
            deletedAt: Date().addingTimeInterval(-60 * 24 * 60 * 60),
            recoveryExpiresAt: Date().addingTimeInterval(-30 * 24 * 60 * 60),
            archiveConversations: true,
            notifyContacts: false,
            isRecoverable: true
        )

        XCTAssertFalse(state.canRecover)
    }

    func testPersonaDeletionOptions() {
        let options = PersonaDeletionOptions(
            archiveConversations: true,
            notifyContacts: true,
            recoveryPeriodDays: 30,
            exportBeforeDelete: true
        )

        XCTAssertTrue(options.archiveConversations)
        XCTAssertTrue(options.notifyContacts)
        XCTAssertEqual(options.recoveryPeriodDays, 30)
        XCTAssertTrue(options.exportBeforeDelete)
    }

    func testPersonaSwitchContext() {
        let context = PersonaSwitchContext(
            fromPersonaId: "persona-1",
            toPersonaId: "persona-2",
            contactId: "contact-1",
            contactKnowsLink: false,
            requiresConfirmation: true,
            warningMessage: "Contact may discover identity link"
        )

        XCTAssertFalse(context.contactKnowsLink)
        XCTAssertTrue(context.requiresConfirmation)
        XCTAssertNotNil(context.warningMessage)
    }

    func testVisibilityMatrixEntry() {
        let entry = VisibilityMatrixEntry(
            contactId: "contact-1",
            contactName: "John",
            personaVisibility: ["persona-1": true, "persona-2": false]
        )

        XCTAssertEqual(entry.contactId, "contact-1")
        XCTAssertEqual(entry.contactName, "John")
        XCTAssertEqual(entry.personaVisibility["persona-1"], true)
        XCTAssertEqual(entry.personaVisibility["persona-2"], false)
    }
}

// MARK: - Enhanced Settings Model Tests

final class EnhancedSettingsModelTests: XCTestCase {

    func testNotificationSettingsDefaults() {
        let settings = EnhancedNotificationSettings()
        XCTAssertTrue(settings.messageNotifications)
        XCTAssertEqual(settings.showPreviews, "always")
        XCTAssertEqual(settings.messageSound, "Echo Default")
        XCTAssertTrue(settings.groupNotifications)
        XCTAssertFalse(settings.mentionsOnly)
        XCTAssertTrue(settings.callNotifications)
        XCTAssertEqual(settings.ringSound, "Reflection")
        XCTAssertTrue(settings.contactRequests)
        XCTAssertTrue(settings.trustScoreChanges)
        XCTAssertTrue(settings.rewardNotifications)
        XCTAssertFalse(settings.quietHoursEnabled)
        XCTAssertEqual(settings.quietHoursFrom, "22:00")
        XCTAssertEqual(settings.quietHoursTo, "07:00")
        XCTAssertTrue(settings.allowInnerCircleCalls)
    }

    func testPrivacySettingsDefaults() {
        let settings = EnhancedPrivacySettings()
        XCTAssertEqual(settings.findByUsername, "everyone")
        XCTAssertEqual(settings.showOnlineStatus, "contacts")
        XCTAssertEqual(settings.showTrustScore, "everyone")
        XCTAssertEqual(settings.whoCanMessage, "contacts")
        XCTAssertTrue(settings.readReceipts)
        XCTAssertTrue(settings.typingIndicators)
        XCTAssertEqual(settings.whoCanCall, "trusted")
        XCTAssertFalse(settings.anchorMessagesByDefault)
        XCTAssertTrue(settings.showVerificationBadges)
        XCTAssertEqual(settings.screenLockTimeout, "immediately")
        XCTAssertTrue(settings.hideMessagePreviews)
        XCTAssertTrue(settings.screenshotNotifications)
    }

    func testAppearanceSettingsDefaults() {
        let settings = AppearanceSettings()
        XCTAssertEqual(settings.theme, "light")
        XCTAssertEqual(settings.accentColor, "indigo")
        XCTAssertEqual(settings.chatWallpaper, "default")
        XCTAssertEqual(settings.messageCorners, "rounded")
        XCTAssertEqual(settings.fontSize, "medium")
        XCTAssertEqual(settings.appIcon, "default")
    }

    func testStorageInfoDefaults() {
        let info = StorageInfo()
        XCTAssertEqual(info.totalUsedBytes, 0)
        XCTAssertEqual(info.totalCapacityBytes, 0)
        XCTAssertEqual(info.cacheBytes, 0)
        XCTAssertEqual(info.autoDownload, "wifi")
        XCTAssertEqual(info.mediaQuality, "standard")
        XCTAssertEqual(info.keepMedia, "forever")
        XCTAssertFalse(info.useLessDataForCalls)
        XCTAssertEqual(info.autoBackup, "daily")
        XCTAssertTrue(info.includeMediaInBackup)
        XCTAssertNil(info.lastBackupDate)
    }

    func testAccountInfoDefaults() {
        let info = AccountInfo()
        XCTAssertEqual(info.phone, "")
        XCTAssertNil(info.email)
        XCTAssertNil(info.did)
        XCTAssertEqual(info.passkeyCount, 0)
        XCTAssertFalse(info.twoFactorEnabled)
        XCTAssertEqual(info.activeSessionCount, 0)
        XCTAssertFalse(info.recoveryPhraseSetUp)
        XCTAssertEqual(info.trustedRecoveryContactCount, 0)
    }

    func testNotificationSettingsCodable() throws {
        var settings = EnhancedNotificationSettings()
        settings.quietHoursEnabled = true
        settings.messageSound = "Custom"
        settings.mentionsOnly = true

        let data = try JSONEncoder().encode(settings)
        let decoded = try JSONDecoder().decode(EnhancedNotificationSettings.self, from: data)

        XCTAssertTrue(decoded.quietHoursEnabled)
        XCTAssertEqual(decoded.messageSound, "Custom")
        XCTAssertTrue(decoded.mentionsOnly)
    }

    func testPrivacySettingsCodable() throws {
        var settings = EnhancedPrivacySettings()
        settings.readReceipts = false
        settings.whoCanCall = "everyone"
        settings.anchorMessagesByDefault = true

        let data = try JSONEncoder().encode(settings)
        let decoded = try JSONDecoder().decode(EnhancedPrivacySettings.self, from: data)

        XCTAssertFalse(decoded.readReceipts)
        XCTAssertEqual(decoded.whoCanCall, "everyone")
        XCTAssertTrue(decoded.anchorMessagesByDefault)
    }

    func testAppearanceSettingsCodable() throws {
        var settings = AppearanceSettings()
        settings.theme = "dark"
        settings.accentColor = "purple"
        settings.appIcon = "gold"

        let data = try JSONEncoder().encode(settings)
        let decoded = try JSONDecoder().decode(AppearanceSettings.self, from: data)

        XCTAssertEqual(decoded.theme, "dark")
        XCTAssertEqual(decoded.accentColor, "purple")
        XCTAssertEqual(decoded.appIcon, "gold")
    }

    func testStorageInfoCodable() throws {
        var info = StorageInfo()
        info.totalUsedBytes = 2576980378
        info.totalCapacityBytes = 5368709120
        info.cacheBytes = 163577856
        info.lastBackupDate = Date()

        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        let data = try encoder.encode(info)

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        let decoded = try decoder.decode(StorageInfo.self, from: data)

        XCTAssertEqual(decoded.totalUsedBytes, 2576980378)
        XCTAssertEqual(decoded.totalCapacityBytes, 5368709120)
        XCTAssertNotNil(decoded.lastBackupDate)
    }

    func testAccountInfoCodable() throws {
        var info = AccountInfo()
        info.phone = "+15551234567"
        info.email = "test@echo.dev"
        info.did = "did:cardano:abc123"
        info.passkeyCount = 2
        info.twoFactorEnabled = true

        let data = try JSONEncoder().encode(info)
        let decoded = try JSONDecoder().decode(AccountInfo.self, from: data)

        XCTAssertEqual(decoded.phone, "+15551234567")
        XCTAssertEqual(decoded.email, "test@echo.dev")
        XCTAssertEqual(decoded.did, "did:cardano:abc123")
        XCTAssertEqual(decoded.passkeyCount, 2)
        XCTAssertTrue(decoded.twoFactorEnabled)
    }
}

// MARK: - ProfileData Tests

final class ProfileDataTests: XCTestCase {

    func testProfileDataDefaults() {
        let data = ProfileData()
        XCTAssertEqual(data.displayName, "")
        XCTAssertEqual(data.username, "")
        XCTAssertEqual(data.bio, "")
        XCTAssertEqual(data.status, "")
        XCTAssertNil(data.avatarURL)
        XCTAssertNil(data.website)
        XCTAssertTrue(data.links.isEmpty)
        XCTAssertEqual(data.trustScore, 0)
        XCTAssertEqual(data.trustLevel, "Newcomer")
        XCTAssertFalse(data.isVerified)
        XCTAssertEqual(data.messagesSent, 0)
        XCTAssertEqual(data.contactsCount, 0)
        XCTAssertEqual(data.echoRewards, 0)
    }

    func testProfileDataCustomValues() {
        let data = ProfileData(
            displayName: "Alex Echo",
            username: "alexecho",
            bio: "Product designer",
            status: "Shipping",
            avatarURL: "https://example.com/avatar.jpg",
            website: "https://echo.dev",
            links: ["https://github.com/alexecho"],
            trustScore: 72,
            trustLevel: "Trusted",
            isVerified: true,
            messagesSent: 247,
            contactsCount: 42,
            echoRewards: 142.5
        )

        XCTAssertEqual(data.displayName, "Alex Echo")
        XCTAssertEqual(data.username, "alexecho")
        XCTAssertEqual(data.bio, "Product designer")
        XCTAssertEqual(data.status, "Shipping")
        XCTAssertEqual(data.avatarURL, "https://example.com/avatar.jpg")
        XCTAssertEqual(data.website, "https://echo.dev")
        XCTAssertEqual(data.links.count, 1)
        XCTAssertEqual(data.trustScore, 72)
        XCTAssertEqual(data.trustLevel, "Trusted")
        XCTAssertTrue(data.isVerified)
        XCTAssertEqual(data.messagesSent, 247)
        XCTAssertEqual(data.contactsCount, 42)
        XCTAssertEqual(data.echoRewards, 142.5)
    }
}

// MARK: - ProfileSettings Tests

final class ProfileSettingsTests: XCTestCase {

    func testProfileSettingsDefaults() {
        let settings = ProfileSettings()
        XCTAssertTrue(settings.notifications.messageNotifications)
        XCTAssertEqual(settings.privacy.findByUsername, "everyone")
        XCTAssertEqual(settings.appearance.theme, "light")
        XCTAssertEqual(settings.account.phone, "")
    }

    func testProfileSettingsCustomValues() {
        var notifications = EnhancedNotificationSettings()
        notifications.quietHoursEnabled = true

        var privacy = EnhancedPrivacySettings()
        privacy.readReceipts = false

        var appearance = AppearanceSettings()
        appearance.theme = "dark"

        var account = AccountInfo()
        account.phone = "+15551234567"

        let settings = ProfileSettings(
            notifications: notifications,
            privacy: privacy,
            appearance: appearance,
            account: account
        )

        XCTAssertTrue(settings.notifications.quietHoursEnabled)
        XCTAssertFalse(settings.privacy.readReceipts)
        XCTAssertEqual(settings.appearance.theme, "dark")
        XCTAssertEqual(settings.account.phone, "+15551234567")
    }
}

// MARK: - Mock Profile Service

class MockProfileService: ProfileServiceProtocol {
    var profileToReturn = ProfileData(
        displayName: "Alex Echo",
        username: "alexecho",
        bio: "Test bio",
        trustScore: 72,
        trustLevel: "Trusted",
        isVerified: true,
        messagesSent: 247,
        contactsCount: 42,
        echoRewards: 142.5
    )

    var personasToReturn: [Persona] = []
    var settingsToReturn = ProfileSettings()
    var storageInfoToReturn = StorageInfo()
    var usernameAvailable = true
    var shouldFail = false

    func fetchProfile() async throws -> ProfileData {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return profileToReturn
    }

    func updateProfile(_ profile: ProfileData) async throws -> ProfileData {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return profile
    }

    func checkUsernameAvailability(_ username: String) async throws -> Bool {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return usernameAvailable
    }

    func uploadAvatar(data: Data) async throws -> String {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return "https://example.com/avatar.jpg"
    }

    func fetchPersonas() async throws -> [Persona] {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return personasToReturn
    }

    func createPersona(_ persona: Persona) async throws -> Persona {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return persona
    }

    func updatePersona(_ persona: Persona) async throws -> Persona {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return persona
    }

    func deletePersona(id: String, options: PersonaDeletionOptions) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func recoverPersona(id: String) async throws -> Persona {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return Persona(id: id, type: .personal, name: "Recovered", displayName: "Recovered",
                       isDefault: false, createdAt: Date(), updatedAt: Date())
    }

    func setDefaultPersona(id: String) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func fetchSettings() async throws -> ProfileSettings {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return settingsToReturn
    }

    func updateNotificationSettings(_ settings: EnhancedNotificationSettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func updatePrivacySettings(_ settings: EnhancedPrivacySettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func updateAppearanceSettings(_ settings: AppearanceSettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func updatePersonaPrivacySettings(personaId: String, settings: PersonaPrivacySettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func updatePersonaNotificationSettings(personaId: String, settings: PersonaNotificationSettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func updatePersonaFeatureSettings(personaId: String, settings: PersonaFeatureSettings) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func fetchStorageInfo() async throws -> StorageInfo {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return storageInfoToReturn
    }

    func clearCache() async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func backUpNow() async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func deleteAccount() async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func grantAccess(personaId: String, contactId: String, permissions: AccessPermissions) async throws -> AccessGrant {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return AccessGrant(contactId: contactId, personaId: personaId, grantedAt: Date(),
                           grantedBy: personaId, permissions: permissions)
    }

    func revokeAccess(grantId: String) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func fetchVisibilityMatrix() async throws -> [VisibilityMatrixEntry] {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return []
    }

    func validatePersonaSwitch(from: String?, to: String, contactId: String) async throws -> PersonaSwitchContext {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return PersonaSwitchContext(fromPersonaId: from, toPersonaId: to, contactId: contactId,
                                   contactKnowsLink: true, requiresConfirmation: false)
    }

    func fetchPersonaConversations(personaId: String) async throws -> [PersonaConversation] {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return []
    }

    func addPersonaBadge(personaId: String, badge: PersonaBadge) async throws -> PersonaBadge {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return badge
    }

    func removePersonaBadge(personaId: String, badgeId: String) async throws {
        if shouldFail { throw NSError(domain: "test", code: 1) }
    }

    func exportPersonaData(personaId: String) async throws -> URL {
        if shouldFail { throw NSError(domain: "test", code: 1) }
        return URL(fileURLWithPath: "/tmp/export.json")
    }
}
