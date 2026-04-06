import XCTest
@testable import Echo

// MARK: - Design System Tests

final class ColorSystemTests: XCTestCase {
    
    func testPrimaryColorHexValue() {
        let primaryColor = Color.echoPrimary
        XCTAssertNotNil(primaryColor)
    }
    
    func testTrustColorMapping() {
        let newcomerColor = Color.trustColor(for: "newcomer")
        XCTAssertEqual(newcomerColor, Color.echoTrustNewcomer)
        
        let verifiedColor = Color.trustColor(for: "verified")
        XCTAssertEqual(verifiedColor, Color.echoTrustVerified)
        
        let invalidColor = Color.trustColor(for: "invalid")
        XCTAssertEqual(invalidColor, Color.echoGray400)
    }
    
    func testGrayScaleComplete() {
        // Verify all 9 gray levels exist
        _ = Color.echoGray50
        _ = Color.echoGray100
        _ = Color.echoGray200
        _ = Color.echoGray300
        _ = Color.echoGray400
        _ = Color.echoGray500
        _ = Color.echoGray600
        _ = Color.echoGray700
        _ = Color.echoGray900
    }
    
    func testSemanticColors() {
        _ = Color.echoSuccess
        _ = Color.echoWarning
        _ = Color.echoError
        _ = Color.echoInfo
    }
    
    func testTrustLevelColors() {
        let levels = ["newcomer", "basic", "trusted", "verified", "highlytrusted"]
        for level in levels {
            let color = Color.trustColor(for: level)
            XCTAssertNotNil(color)
        }
    }
}

// MARK: - Typography Tests

final class TypographyTests: XCTestCase {
    
    func testTypographyStyleFonts() {
        let styles: [TypographyStyle] = [
            .display, .h1, .h2, .h3, .h4, .bodyLarge, .body, .bodySmall,
            .button, .caption, .tiny, .overline
        ]
        
        for style in styles {
            XCTAssertNotNil(style.font)
            XCTAssertGreaterThan(style.lineHeight, 0)
        }
    }
    
    func testDisplayHeadingSize() {
        let style = TypographyStyle.display
        XCTAssertEqual(style.font.pointSize, 36)
    }
    
    func testBodyTextSize() {
        let style = TypographyStyle.body
        XCTAssertEqual(style.font.pointSize, 14)
    }
    
    func testLetterSpacing() {
        let displayStyle = TypographyStyle.display
        XCTAssertEqual(displayStyle.letterSpacing, -0.18)
        
        let regularStyle = TypographyStyle.body
        XCTAssertEqual(regularStyle.letterSpacing, 0)
    }
}

// MARK: - Spacing Tests

final class SpacingTests: XCTestCase {
    
    func testSpacingValues() {
        XCTAssertEqual(Spacing.xs.rawValue, 4)
        XCTAssertEqual(Spacing.sm.rawValue, 8)
        XCTAssertEqual(Spacing.md.rawValue, 12)
        XCTAssertEqual(Spacing.lg.rawValue, 16)
        XCTAssertEqual(Spacing.xl.rawValue, 20)
        XCTAssertEqual(Spacing.xxl.rawValue, 24)
        XCTAssertEqual(Spacing.xxxl.rawValue, 32)
        XCTAssertEqual(Spacing.huge.rawValue, 40)
        XCTAssertEqual(Spacing.massive.rawValue, 48)
    }
    
    func testSpacingOrder() {
        let spacings = [
            Spacing.xs.rawValue,
            Spacing.sm.rawValue,
            Spacing.md.rawValue,
            Spacing.lg.rawValue
        ]
        
        for i in 0..<(spacings.count - 1) {
            XCTAssertLessThan(spacings[i], spacings[i + 1])
        }
    }
}

// MARK: - Component Tests

final class EchoButtonComponentTests: XCTestCase {
    
    func testButtonInitialization() {
        let button = EchoButton("Test Button", style: .primary, size: .medium) {}
        XCTAssertNotNil(button)
    }
    
    func testButtonStyles() {
        let styles: [EchoButtonStyle] = [.primary, .secondary, .ghost, .destructive]
        for style in styles {
            let button = EchoButton("Test", style: style) {}
            XCTAssertNotNil(button)
        }
    }
    
    func testButtonSizes() {
        let sizes: [EchoButtonSize] = [.large, .medium, .small]
        for size in sizes {
            let button = EchoButton("Test", size: size) {}
            XCTAssertNotNil(button)
        }
    }
    
    func testButtonStates() {
        let loadingButton = EchoButton("Loading", isLoading: true) {}
        XCTAssertNotNil(loadingButton)
        
        let disabledButton = EchoButton("Disabled", isDisabled: true) {}
        XCTAssertNotNil(disabledButton)
    }
}

final class AvatarViewComponentTests: XCTestCase {
    
    func testAvatarSizes() {
        let sizes: [AvatarSize] = [.xs, .sm, .md, .lg, .xl, .xxl]
        for size in sizes {
            let avatar = AvatarView(initials: "JD", size: size)
            XCTAssertEqual(avatar.size, size)
            XCTAssertGreaterThan(size.dimension, 0)
        }
    }
    
    func testAvatarStatus() {
        let statuses: [AvatarStatus] = [.none, .online, .idle, .offline, .verified]
        for status in statuses {
            let avatar = AvatarView(initials: "JD", status: status)
            XCTAssertEqual(avatar.status, status)
        }
    }
    
    func testAvatarTrustRing() {
        let avatar = AvatarView(
            initials: "JD",
            size: .lg,
            showTrustRing: true,
            trustLevel: "verified"
        )
        XCTAssertTrue(avatar.showTrustRing)
        XCTAssertEqual(avatar.trustLevel, "verified")
    }
}

final class TrustScoreViewComponentTests: XCTestCase {
    
    func testScoreRangeValidation() {
        let underZero = TrustScoreView(score: -10)
        XCTAssertEqual(underZero.score, 0)
        
        let overHundred = TrustScoreView(score: 150)
        XCTAssertEqual(overHundred.score, 100)
        
        let valid = TrustScoreView(score: 75)
        XCTAssertEqual(valid.score, 75)
    }
    
    func testScorePercentage() {
        let scoreView = TrustScoreView(score: 50)
        XCTAssertEqual(scoreView.scorePercentage, 0.5)
    }
    
    func testTrustLevelMapping() {
        let levels = ["Newcomer", "Basic", "Trusted", "Verified", "Highly Trusted"]
        for level in levels {
            let view = TrustScoreView(score: 50, level: level)
            XCTAssertEqual(view.level, level)
        }
    }
}

final class TrustBadgeComponentTests: XCTestCase {
    
    func testBadgeDisplayText() {
        let badge = TrustBadge(level: "verified")
        XCTAssertEqual(badge.level, "verified")
    }
    
    func testBadgeSizes() {
        let sizes: [TrustBadge.BadgeSize] = [.small, .medium, .large]
        for size in sizes {
            let badge = TrustBadge(level: "verified", size: size)
            XCTAssertEqual(badge.size, size)
        }
    }
}

// MARK: - Screen Tests

final class OnboardingScreenTests: XCTestCase {
    
    func testWelcomeViewInitialization() {
        let welcome = WelcomeView()
        XCTAssertNotNil(welcome)
    }
    
    func testPhoneEntryValidation() {
        var phone = ""
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertFalse(isValid)
        
        phone = "5551234567"
        let isValid2 = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertTrue(isValid2)
    }
    
    func testOTPVerificationInitialization() {
        let otp = OTPVerificationView(phoneNumber: "+1 (555) 123-4567")
        XCTAssertNotNil(otp)
    }
}

final class MessagingScreenTests: XCTestCase {
    
    func testConversationListInitialization() {
        let list = ConversationListView()
        XCTAssertNotNil(list)
    }
    
    func testChatViewInitialization() {
        let chat = ChatView(contactName: "John Doe")
        XCTAssertNotNil(chat)
    }
}

final class ContactsScreenTests: XCTestCase {
    
    func testContactsListInitialization() {
        let list = ContactsListView()
        XCTAssertNotNil(list)
    }
    
    func testTrustDashboardInitialization() {
        let dashboard = TrustDashboardView()
        XCTAssertNotNil(dashboard)
    }
}

final class ProfileScreenTests: XCTestCase {

    func testProfileTabViewInitialization() {
        let profile = ProfileTabView()
        XCTAssertNotNil(profile)
    }

    func testProfileTabViewWithData() {
        let profileData = ProfileData(
            displayName: "Alex Echo",
            username: "alexecho",
            bio: "Test bio",
            status: "Online",
            trustScore: 72,
            trustLevel: "Trusted",
            isVerified: true,
            messagesSent: 100,
            contactsCount: 42,
            echoRewards: 142.5
        )
        let view = ProfileTabView(profile: profileData)
        XCTAssertNotNil(view)
    }

    func testProfileTabViewWithPersonas() {
        let personas = [
            Persona(
                id: "1", type: .professional, name: "Pro", displayName: "Alex",
                bio: "Work", useMainAvatar: true, visibility: .all,
                selectedContactIds: [], isDefault: true,
                createdAt: Date(), updatedAt: Date()
            )
        ]
        let view = ProfileTabView(personas: personas)
        XCTAssertNotNil(view)
    }

    func testSettingsViewInitialization() {
        let settings = SettingsView()
        XCTAssertNotNil(settings)
    }

    func testEditProfileViewInitialization() {
        let view = EditProfileView(
            displayName: .constant("Alex"),
            username: .constant("alexecho"),
            bio: .constant("Test"),
            status: .constant("Active"),
            website: .constant("https://echo.dev")
        )
        XCTAssertNotNil(view)
    }

    func testPersonasManagementViewEmpty() {
        let view = PersonasManagementView(personas: [])
        XCTAssertNotNil(view)
    }

    func testPersonasManagementViewWithPersonas() {
        let personas = [
            Persona(
                id: "1", type: .professional, name: "Pro", displayName: "Alex",
                bio: nil, useMainAvatar: true, visibility: .all,
                selectedContactIds: [], isDefault: true,
                createdAt: Date(), updatedAt: Date()
            ),
            Persona(
                id: "2", type: .personal, name: "Personal", displayName: "Al",
                bio: "Vibing", useMainAvatar: false, visibility: .selected,
                selectedContactIds: ["c1", "c2"], isDefault: false,
                createdAt: Date(), updatedAt: Date()
            )
        ]
        let view = PersonasManagementView(personas: personas, maxPersonas: 7)
        XCTAssertNotNil(view)
    }

    func testCreateEditPersonaViewNew() {
        let view = CreateEditPersonaView()
        XCTAssertNotNil(view)
    }

    func testCreateEditPersonaViewEditing() {
        let persona = Persona(
            id: "1", type: .gaming, name: "Gaming", displayName: "NightOwl42",
            bio: "Competitive gamer", useMainAvatar: false, visibility: .selected,
            selectedContactIds: ["c1"], isDefault: false,
            createdAt: Date(), updatedAt: Date()
        )
        let view = CreateEditPersonaView(persona: persona)
        XCTAssertNotNil(view)
    }

    func testAccountSettingsViewInitialization() {
        let view = AccountSettingsView()
        XCTAssertNotNil(view)
    }

    func testAccountSettingsViewWithData() {
        let account = AccountInfo(
            phone: "+15551234567",
            email: "test@echo.dev",
            did: "did:cardano:abc123xyz",
            passkeyCount: 2,
            twoFactorEnabled: true,
            activeSessionCount: 3,
            recoveryPhraseSetUp: true,
            trustedRecoveryContactCount: 2
        )
        let view = AccountSettingsView(account: account)
        XCTAssertNotNil(view)
    }

    func testPrivacySecuritySettingsViewInitialization() {
        let view = PrivacySecuritySettingsView(settings: .constant(EnhancedPrivacySettings()))
        XCTAssertNotNil(view)
    }

    func testNotificationSettingsViewInitialization() {
        let view = NotificationSettingsView(settings: .constant(EnhancedNotificationSettings()))
        XCTAssertNotNil(view)
    }

    func testAppearanceSettingsViewInitialization() {
        let view = AppearanceSettingsView(settings: .constant(AppearanceSettings()))
        XCTAssertNotNil(view)
    }

    func testStorageDataViewInitialization() {
        let view = StorageDataView(storageInfo: .constant(StorageInfo()))
        XCTAssertNotNil(view)
    }

    func testAboutViewInitialization() {
        let view = AboutView()
        XCTAssertNotNil(view)
    }

    func testAboutViewWithVersion() {
        let view = AboutView(appVersion: "2.0.0")
        XCTAssertNotNil(view)
    }

    // MARK: - New Persona Feature Screens

    func testPersonaSwitcherViewInitialization() {
        let personas = [
            Persona(id: "1", type: .professional, name: "Pro", displayName: "Alex",
                    isDefault: true, createdAt: Date(), updatedAt: Date()),
            Persona(id: "2", type: .personal, name: "Personal", displayName: "Al",
                    isDefault: false, createdAt: Date(), updatedAt: Date())
        ]
        let view = PersonaSwitcherView(personas: personas, activePersonaId: .constant("1"))
        XCTAssertNotNil(view)
    }

    func testVisibilityMatrixViewInitialization() {
        let view = VisibilityMatrixView(entries: [])
        XCTAssertNotNil(view)
    }

    func testVisibilityMatrixViewWithEntries() {
        let entries = [
            VisibilityMatrixEntry(
                contactId: "c1", contactName: "John",
                personaVisibility: ["p1": true, "p2": false]
            )
        ]
        let view = VisibilityMatrixView(entries: entries)
        XCTAssertNotNil(view)
    }

    func testPersonaPrivacySettingsViewInitialization() {
        let view = PersonaPrivacySettingsView(settings: .constant(PersonaPrivacySettings()))
        XCTAssertNotNil(view)
    }

    func testPersonaNotificationSettingsViewInitialization() {
        let view = PersonaNotificationSettingsView(settings: .constant(PersonaNotificationSettings()))
        XCTAssertNotNil(view)
    }

    func testPersonaFeatureSettingsViewInitialization() {
        let view = PersonaFeatureSettingsView(settings: .constant(PersonaFeatureSettings()))
        XCTAssertNotNil(view)
    }

    func testPersonaDeletionViewInitialization() {
        let persona = Persona(
            id: "1", type: .professional, name: "Pro", displayName: "Alex",
            isDefault: false, createdAt: Date(), updatedAt: Date(),
            messageCount: 42
        )
        let view = PersonaDeletionView(persona: persona)
        XCTAssertNotNil(view)
    }

    func testPersonaBadgesViewInitialization() {
        let persona = Persona(
            id: "1", type: .professional, name: "Pro", displayName: "Alex",
            isDefault: true, createdAt: Date(), updatedAt: Date(),
            badges: [
                PersonaBadge(type: .identityVerified, issuer: "EchoVerify", verifiable: true)
            ]
        )
        let view = PersonaBadgesView(persona: persona)
        XCTAssertNotNil(view)
    }

    func testPersonaRecoveryBannerInitialization() {
        let persona = Persona(
            id: "1", type: .professional, name: "Pro", displayName: "Alex",
            isDefault: false, createdAt: Date(), updatedAt: Date(),
            deletionState: PersonaDeletionState(
                deletedAt: Date(),
                recoveryExpiresAt: Date().addingTimeInterval(30 * 24 * 60 * 60),
                archiveConversations: true,
                notifyContacts: false,
                isRecoverable: true
            )
        )
        let view = PersonaRecoveryBanner(persona: persona)
        XCTAssertNotNil(view)
    }

    func testPersonaSwitchWarningViewInitialization() {
        let context = PersonaSwitchContext(
            fromPersonaId: "p1",
            toPersonaId: "p2",
            contactId: "c1",
            contactKnowsLink: false,
            requiresConfirmation: true,
            warningMessage: "This may reveal your identity"
        )
        let view = PersonaSwitchWarningView(context: context)
        XCTAssertNotNil(view)
    }
}

// MARK: - Corner Radius & Screen Padding Tests

final class CornerRadiusTests: XCTestCase {

    func testCornerRadiusValues() {
        XCTAssertEqual(CornerRadius.sm, 4)
        XCTAssertEqual(CornerRadius.md, 8)
        XCTAssertEqual(CornerRadius.lg, 12)
        XCTAssertEqual(CornerRadius.xl, 16)
        XCTAssertEqual(CornerRadius.xxl, 24)
        XCTAssertEqual(CornerRadius.full, 9999)
    }

    func testCornerRadiusOrdering() {
        let radii: [CGFloat] = [CornerRadius.sm, CornerRadius.md, CornerRadius.lg, CornerRadius.xl, CornerRadius.xxl]
        for i in 0..<(radii.count - 1) {
            XCTAssertLessThan(radii[i], radii[i + 1])
        }
    }
}

final class ScreenPaddingTests: XCTestCase {

    func testScreenPaddingValues() {
        XCTAssertEqual(ScreenPadding.horizontal, 24)
        XCTAssertEqual(ScreenPadding.vertical, 24)
        XCTAssertEqual(ScreenPadding.cardPadding, 16)
    }
}

// MARK: - Theme System Tests

final class ThemeTypeTests: XCTestCase {

    func testThemeTypeCases() {
        let types = ThemeType.allCases
        XCTAssertEqual(types.count, 3)
        XCTAssertTrue(types.contains(.electricViolet))
        XCTAssertTrue(types.contains(.tealWarm))
        XCTAssertTrue(types.contains(.icyMinimal))
    }

    func testThemeTypeRawValues() {
        XCTAssertEqual(ThemeType.electricViolet.rawValue, "electric_violet")
        XCTAssertEqual(ThemeType.tealWarm.rawValue, "teal_warm")
        XCTAssertEqual(ThemeType.icyMinimal.rawValue, "icy_minimal")
    }

    func testThemeTypeNames() {
        XCTAssertEqual(ThemeType.electricViolet.name, "Electric Violet")
        XCTAssertEqual(ThemeType.tealWarm.name, "Teal + Warm Neutral")
        XCTAssertEqual(ThemeType.icyMinimal.name, "Icy Minimal")
    }

    func testThemeTypeDescriptions() {
        XCTAssertEqual(ThemeType.electricViolet.description, "Bold & vibrant")
        XCTAssertEqual(ThemeType.tealWarm.description, "Modern & professional")
        XCTAssertEqual(ThemeType.icyMinimal.description, "Fresh & techy")
    }

    func testThemeTypeIdentifiable() {
        for type in ThemeType.allCases {
            XCTAssertEqual(type.id, type.rawValue)
        }
    }

    func testThemeTypePaletteReturnsDistinctPalettes() {
        let ev = ThemeType.electricViolet.palette
        let tw = ThemeType.tealWarm.palette
        let im = ThemeType.icyMinimal.palette
        // Each palette should have different primary colors
        XCTAssertNotEqual(ev.primary, tw.primary)
        XCTAssertNotEqual(tw.primary, im.primary)
        XCTAssertNotEqual(ev.primary, im.primary)
    }
}

final class ThemeTests: XCTestCase {

    func testThemeInitialization() {
        let theme = Theme(type: .electricViolet)
        XCTAssertEqual(theme.type, .electricViolet)
        XCTAssertNotNil(theme.colors)
    }

    func testThemeColorsMatchPalette() {
        for type in ThemeType.allCases {
            let theme = Theme(type: type)
            let palette = type.palette
            XCTAssertEqual(theme.colors.primary, palette.primary)
            XCTAssertEqual(theme.colors.error, palette.error)
            XCTAssertEqual(theme.colors.background, palette.background)
        }
    }
}

final class ThemeManagerTests: XCTestCase {

    func testThemeManagerDefaultsToElectricViolet() {
        // Clear any saved theme
        UserDefaults.standard.removeObject(forKey: "selected_theme")
        let manager = ThemeManager()
        XCTAssertEqual(manager.currentTheme.type, .electricViolet)
    }

    func testThemeManagerSetTheme() {
        let manager = ThemeManager()
        manager.setTheme(.tealWarm)
        XCTAssertEqual(manager.currentTheme.type, .tealWarm)

        manager.setTheme(.icyMinimal)
        XCTAssertEqual(manager.currentTheme.type, .icyMinimal)

        manager.setTheme(.electricViolet)
        XCTAssertEqual(manager.currentTheme.type, .electricViolet)
    }
}

final class ColorPaletteTests: XCTestCase {

    func testElectricVioletPaletteColors() {
        let palette = ElectricVioletPalette()
        XCTAssertNotNil(palette.primary)
        XCTAssertNotNil(palette.accent)
        XCTAssertNotNil(palette.secondary)
        XCTAssertNotNil(palette.success)
        XCTAssertNotNil(palette.warning)
        XCTAssertNotNil(palette.error)
        XCTAssertNotNil(palette.background)
        XCTAssertNotNil(palette.textPrimary)
        XCTAssertNotNil(palette.border)
    }

    func testTealWarmPaletteColors() {
        let palette = TealWarmPalette()
        XCTAssertNotNil(palette.primary)
        XCTAssertNotNil(palette.accent)
        XCTAssertNotNil(palette.background)
        XCTAssertNotNil(palette.textPrimary)
    }

    func testIcyMinimalPaletteColors() {
        let palette = IcyMinimalPalette()
        XCTAssertNotNil(palette.primary)
        XCTAssertNotNil(palette.accent)
        XCTAssertNotNil(palette.background)
        XCTAssertNotNil(palette.textPrimary)
    }

    func testAllPalettesHaveGradients() {
        let palettes: [ColorPalette] = [ElectricVioletPalette(), TealWarmPalette(), IcyMinimalPalette()]
        for palette in palettes {
            XCTAssertNotNil(palette.primaryGradient)
            XCTAssertNotNil(palette.secondaryGradient)
            XCTAssertNotNil(palette.warmGradient)
        }
    }
}

// MARK: - Pinned Section Component Tests

final class PinnedItemModelTests: XCTestCase {

    func testPinnedItemContactCreation() {
        let item = PinnedItem(
            id: "1", type: .contact, name: "Alice", initials: "AL", gradientIndex: 0
        )
        XCTAssertEqual(item.id, "1")
        XCTAssertEqual(item.name, "Alice")
        XCTAssertEqual(item.initials, "AL")
        XCTAssertFalse(item.isOnline)
        XCTAssertEqual(item.unreadCount, 0)
    }

    func testPinnedItemGroupCreation() {
        let item = PinnedItem(
            id: "2", type: .group, name: "Design Team", initials: "DT", gradientIndex: 1
        )
        XCTAssertEqual(item.type, .group)
    }

    func testPinnedItemWithOnlineAndUnread() {
        let item = PinnedItem(
            id: "3", type: .contact, name: "Bob", initials: "BW",
            gradientIndex: 2, isOnline: true, unreadCount: 5
        )
        XCTAssertTrue(item.isOnline)
        XCTAssertEqual(item.unreadCount, 5)
    }
}

final class OnlineIndicatorTests: XCTestCase {

    func testOnlineIndicatorInitialization() {
        let indicator = OnlineIndicator()
        XCTAssertNotNil(indicator)
    }
}

final class UnreadBadgeTests: XCTestCase {

    func testUnreadBadgeInitialization() {
        let badge = UnreadBadge(count: 3)
        XCTAssertEqual(badge.count, 3)
    }

    func testUnreadBadgeLargeCount() {
        let badge = UnreadBadge(count: 150)
        XCTAssertEqual(badge.count, 150)
    }
}

final class ContactAvatarViewTests: XCTestCase {

    func testContactAvatarViewInitialization() {
        let view = ContactAvatarView(
            initials: "JD",
            gradient: [.blue, .purple],
            size: 56
        )
        XCTAssertEqual(view.initials, "JD")
        XCTAssertEqual(view.size, 56)
    }
}

final class GroupAvatarViewTests: XCTestCase {

    func testGroupAvatarViewInitialization() {
        let members = [
            GroupMember(initials: "A", color: .blue),
            GroupMember(initials: "B", color: .green)
        ]
        let view = GroupAvatarView(members: members, size: 56)
        XCTAssertEqual(view.members.count, 2)
        XCTAssertEqual(view.size, 56)
    }
}

final class PinnedItemCardTests: XCTestCase {

    func testPinnedItemCardInitialization() {
        let item = PinnedItem(
            id: "1", type: .contact, name: "Alice", initials: "AL", gradientIndex: 0
        )
        var tapped = false
        let card = PinnedItemCard(item: item) { tapped = true }
        XCTAssertNotNil(card)
    }
}

final class PinnedSectionViewTests: XCTestCase {

    func testPinnedSectionViewEmpty() {
        let view = PinnedSectionView(items: [], onItemTap: { _ in }, onEditTap: {})
        XCTAssertNotNil(view)
    }

    func testPinnedSectionViewWithItems() {
        let items = [
            PinnedItem(id: "1", type: .contact, name: "Alice", initials: "AL", gradientIndex: 0, isOnline: true, unreadCount: 3),
            PinnedItem(id: "2", type: .contact, name: "Bob", initials: "BW", gradientIndex: 1)
        ]
        let view = PinnedSectionView(items: items, onItemTap: { _ in }, onEditTap: {})
        XCTAssertNotNil(view)
    }

    func testPinnedSectionViewMaxItems() {
        let view = PinnedSectionView(items: [], maxItems: 5, onItemTap: { _ in }, onEditTap: {})
        XCTAssertEqual(view.maxItems, 5)
    }
}

// MARK: - ScaleButtonStyle & SectionDivider Tests

final class ScaleButtonStyleTests: XCTestCase {

    func testScaleButtonStyleInitialization() {
        let style = ScaleButtonStyle()
        XCTAssertNotNil(style)
    }
}

final class SectionDividerTests: XCTestCase {

    func testSectionDividerInitialization() {
        let divider = SectionDivider(title: "ALL MESSAGES")
        XCTAssertEqual(divider.title, "ALL MESSAGES")
    }
}

final class RewardsScreenTests: XCTestCase {
    
    func testRewardsDashboardInitialization() {
        let rewards = RewardsDashboardView()
        XCTAssertNotNil(rewards)
    }
    
    func testStakingViewInitialization() {
        let staking = StakingView()
        XCTAssertNotNil(staking)
    }
    
    func testReferralViewInitialization() {
        let referral = ReferralView()
        XCTAssertNotNil(referral)
    }
}

// MARK: - ViewModel Tests

final class AuthViewModelUnitTests: XCTestCase {
    
    var authViewModel: AuthViewModel!
    var mockAuthService: MockAuthService!
    var mockKeychainManager: MockKeychainManager!
    
    override func setUp() {
        super.setUp()
        mockAuthService = MockAuthService()
        mockKeychainManager = MockKeychainManager()
        authViewModel = AuthViewModel(
            authService: mockAuthService,
            keychainManager: mockKeychainManager
        )
    }
    
    override func tearDown() {
        authViewModel = nil
        mockAuthService = nil
        mockKeychainManager = nil
        super.tearDown()
    }
    
    func testInitialAuthState() {
        XCTAssertEqual(authViewModel.authState, .welcome)
        XCTAssertFalse(authViewModel.isAuthenticated)
        XCTAssertFalse(authViewModel.isLoading)
        XCTAssertNil(authViewModel.errorMessage)
    }
    
    func testSignOut() {
        authViewModel.isAuthenticated = true
        authViewModel.signOut()
        
        XCTAssertFalse(authViewModel.isAuthenticated)
        XCTAssertEqual(authViewModel.authState, .welcome)
        XCTAssertEqual(authViewModel.phone, "")
    }
}

final class MessagingViewModelUnitTests: XCTestCase {
    
    var messagingViewModel: MessagingViewModel!
    var mockMessagingService: MockMessagingService!
    
    override func setUp() {
        super.setUp()
        mockMessagingService = MockMessagingService()
        messagingViewModel = MessagingViewModel(messagingService: mockMessagingService)
    }
    
    func testInitialMessagingState() {
        XCTAssertTrue(messagingViewModel.conversations.isEmpty)
        XCTAssertNil(messagingViewModel.selectedConversation)
        XCTAssertTrue(messagingViewModel.messages.isEmpty)
        XCTAssertFalse(messagingViewModel.isLoading)
    }
}

final class TrustViewModelUnitTests: XCTestCase {
    
    var trustViewModel: TrustViewModel!
    var mockTrustService: MockTrustService!
    
    override func setUp() {
        super.setUp()
        mockTrustService = MockTrustService()
        trustViewModel = TrustViewModel(trustService: mockTrustService)
    }
    
    func testInitialTrustState() {
        XCTAssertEqual(trustViewModel.trustScore, 0)
        XCTAssertEqual(trustViewModel.trustLevel, "Newcomer")
        XCTAssertNil(trustViewModel.breakdown)
        XCTAssertFalse(trustViewModel.isLoading)
    }
}

final class RewardsViewModelUnitTests: XCTestCase {
    
    var rewardsViewModel: RewardsViewModel!
    var mockRewardsService: MockRewardsService!
    
    override func setUp() {
        super.setUp()
        mockRewardsService = MockRewardsService()
        rewardsViewModel = RewardsViewModel(rewardsService: mockRewardsService)
    }
    
    func testInitialRewardsState() {
        XCTAssertEqual(rewardsViewModel.tokenBalance, 0)
        XCTAssertTrue(rewardsViewModel.activities.isEmpty)
        XCTAssertFalse(rewardsViewModel.isLoading)
    }
}

// MARK: - Mock Services for Testing

class MockAuthService: AuthServiceProtocol {
    func requestOTP(phone: String) async throws -> OTPResponse {
        OTPResponse(expiresIn: 300, phone: phone)
    }
    
    func verifyOTP(phone: String, code: String) async throws -> AuthResponse {
        AuthResponse(
            token: "mock_token",
            refreshToken: "mock_refresh",
            did: "did:example:123",
            user: UserProfile(id: "1", phone: phone, displayName: nil, username: nil, avatarURL: nil)
        )
    }
    
    func registerPasskey() async throws {}
    
    func authenticateWithPasskey() async throws -> AuthResponse {
        AuthResponse(
            token: "mock_token",
            refreshToken: "mock_refresh",
            did: "did:example:123",
            user: UserProfile(id: "1", phone: "+1234567890", displayName: nil, username: nil, avatarURL: nil)
        )
    }
    
    func refreshToken() async throws -> String {
        "new_mock_token"
    }
}

class MockKeychainManager: KeychainManagerProtocol {
    private var storage: [String: String] = [:]
    
    func saveToken(_ token: String) throws {
        storage["token"] = token
    }
    
    func retrieveToken() -> String? {
        storage["token"]
    }
    
    func clearAll() {
        storage.removeAll()
    }
}

class MockMessagingService: MessagingServiceProtocol {
    func fetchConversations() async throws -> [ConversationModel] {
        [
            ConversationModel(
                id: "1",
                participantId: "user1",
                participantName: "John Doe",
                lastMessage: "Hello!",
                unreadCount: 2,
                updatedAt: Date()
            )
        ]
    }
    
    func fetchMessages(conversationId: String) async throws -> [MessageModel] {
        [
            MessageModel(
                id: "msg1",
                conversationId: conversationId,
                senderId: "user1",
                content: "Hi there!",
                status: .read,
                createdAt: Date()
            )
        ]
    }
    
    func sendMessage(_ content: String, to conversationId: String) async throws {}
    
    func markAsRead(conversationId: String) async throws {}
}

class MockTrustService: TrustServiceProtocol {
    func fetchTrustScore(userId: String) async throws -> TrustScoreResult {
        TrustScoreResult(
            score: 65,
            level: "Verified",
            breakdown: TrustBreakdown(identity: 25, behavior: 18, network: 15, activity: 7)
        )
    }
    
    func submitVerification(documents: [URL], selfie: URL) async throws {}
    
    func updateTrustCircle(contactId: String, tier: String) async throws {}
}

class MockRewardsService: RewardsServiceProtocol {
    func fetchBalance() async throws -> Double {
        1250.50
    }
    
    func fetchActivity() async throws -> [RewardActivityModel] {
        []
    }
    
    func stakeTokens(amount: Double, period: Int) async throws {}
    
    func claimRewards() async throws -> Double {
        1500.00
    }
}
