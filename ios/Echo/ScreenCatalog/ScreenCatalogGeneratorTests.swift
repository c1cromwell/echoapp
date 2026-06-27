#if os(iOS)
import SwiftUI
import XCTest
@testable import Echo

/// Exports PNG screen catalog to `docs/screen_catalog`.
/// Invoked only via `make screen-catalog` (`-only-testing:EchoUnitTests/ScreenCatalogGeneratorTests/...`).
@MainActor
final class ScreenCatalogGeneratorTests: XCTestCase {
    override class func setUp() {
        try? ScreenCatalogRenderer.resetCatalog()
    }

    func testExportFullCatalog() throws {
        let coordinator = ScreenCatalogFixtures.firstRunCoordinator()

        try ScreenCatalogRenderer.render(
            EchoWelcomeView(onSetUp: {}, onAlreadyHaveAccount: {}),
            journey: "onboarding",
            stepId: "01-welcome",
            title: "Welcome",
            e2eRef: "E2E §6.1 step 1"
        )
        try ScreenCatalogRenderer.render(
            DisplayNameEntryView(
                coordinator: coordinator,
                availabilityClient: CatalogUsernameClient(available: true)
            ),
            journey: "onboarding",
            stepId: "02-username",
            title: "Choose username",
            e2eRef: "E2E §6.1 step 2"
        )
        try ScreenCatalogRenderer.render(
            OnboardingOptionsView(coordinator: coordinator),
            journey: "onboarding",
            stepId: "03-secure-account",
            title: "Face ID + optional VIP",
            e2eRef: "E2E §6.1 step 3"
        )

        try ScreenCatalogRenderer.render(
            StorageLockedView(onUnlock: {}),
            journey: "auth",
            stepId: "01-storage-locked",
            title: "Storage locked",
            e2eRef: "E2E §6.2 step 1"
        )
        try ScreenCatalogRenderer.render(
            BiometricLockoutView(
                lockState: .requiresPasscode(failureCount: 5),
                onUnlocked: {},
                onRecovery: {}
            ),
            journey: "auth",
            stepId: "02-soft-lockout",
            title: "Biometric soft lockout",
            e2eRef: "E2E §6.2 step 3"
        )

        try ScreenCatalogRenderer.render(
            MessagesEmptyStateView(
                displayName: "Alex",
                trustTier: 1,
                onComposeTapped: {},
                onUpgradeTrustTapped: {}
            ),
            journey: "messaging",
            stepId: "01-empty-hub",
            title: "Messages hub (empty)",
            e2eRef: "TESTFLIGHT A1"
        )
        try ScreenCatalogRenderer.render(
            CatalogChatThreadPreview(),
            journey: "messaging",
            stepId: "02-dm-thread",
            title: "Direct message thread",
            e2eRef: "TESTFLIGHT A2–A7"
        )
        try ScreenCatalogRenderer.render(
            TypingIndicatorView(label: "Jordan is typing…"),
            journey: "messaging",
            stepId: "03-typing",
            title: "Typing indicator",
            e2eRef: "TESTFLIGHT A6"
        )
        try ScreenCatalogRenderer.render(
            ReactionPickerView(onSelect: { _ in }, onDismiss: {}),
            journey: "messaging",
            stepId: "04-reactions",
            title: "Reaction picker",
            e2eRef: "TESTFLIGHT A7"
        )
        try ScreenCatalogRenderer.render(
            CatalogMessagesHubGroupsPreview(),
            journey: "messaging",
            stepId: "05-hub-groups",
            title: "Messages hub (groups)",
            e2eRef: "TESTFLIGHT groups tab"
        )
        try ScreenCatalogRenderer.render(
            CatalogGroupCreatePreview(),
            journey: "messaging",
            stepId: "06-group-create",
            title: "Create group",
            e2eRef: "GroupCreateSheet"
        )

        try ScreenCatalogRenderer.render(
            CatalogSOCKSProxyPreview(),
            journey: "privacy",
            stepId: "01-socks-proxy",
            title: "SOCKS proxy settings",
            e2eRef: "WO-SX4 / Privacy Hub"
        )
        try ScreenCatalogRenderer.render(
            CatalogPQHybridPreview(),
            journey: "privacy",
            stepId: "02-post-quantum",
            title: "Post-quantum handshake",
            e2eRef: "WO-SX2 / Privacy Hub"
        )
        try ScreenCatalogRenderer.render(
            CatalogHiddenFolderPreview(),
            journey: "privacy",
            stepId: "03-hidden-folder",
            title: "Hidden folder settings",
            e2eRef: "Phase 5 privacy"
        )
        try ScreenCatalogRenderer.render(
            CatalogPrivacyHubPreview(),
            journey: "privacy",
            stepId: "04-privacy-hub",
            title: "Privacy hub",
            e2eRef: "WO-228 / Settings"
        )

        try ScreenCatalogRenderer.render(
            CatalogInviteAcceptPreview(inviteCode: "DEMO-ECHO"),
            journey: "contacts",
            stepId: "01-invite-accept",
            title: "Accept invite sheet",
            e2eRef: "TESTFLIGHT B1"
        )
        try ScreenCatalogRenderer.render(
            CatalogUsernameSearchPreview(),
            journey: "contacts",
            stepId: "02-username-search",
            title: "Add contact by username",
            e2eRef: "TESTFLIGHT A2"
        )
    }
}
#endif
