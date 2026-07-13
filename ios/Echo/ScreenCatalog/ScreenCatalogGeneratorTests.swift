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
            CatalogGlacialLoginNormalPreview(),
            journey: "auth",
            stepId: "03-glacial-login-normal",
            title: "Glacial login (saved user)",
            e2eRef: "GlacialLoginScreen · normal"
        )
        try ScreenCatalogRenderer.render(
            CatalogGlacialLoginNoAccountPreview(),
            journey: "auth",
            stepId: "04-glacial-login-no-account",
            title: "Glacial login (no account)",
            e2eRef: "GlacialLoginScreen · noAccount"
        )
        try ScreenCatalogRenderer.render(
            CatalogGlacialLoginSoftLockPreview(),
            journey: "auth",
            stepId: "05-glacial-login-soft-lock",
            title: "Glacial login (soft lock)",
            e2eRef: "GlacialLoginScreen · softLocked"
        )

        try ScreenCatalogRenderer.render(
            CatalogContactsListPreview(),
            journey: "tabs",
            stepId: "01-contacts",
            title: "Contacts tab",
            e2eRef: "MainTabView · Contacts"
        )
        try ScreenCatalogRenderer.render(
            CatalogMessagesHubPopulatedPreview(),
            journey: "tabs",
            stepId: "02-messages-hub",
            title: "Messages hub (populated)",
            e2eRef: "MainTabView · Messages"
        )
        try ScreenCatalogRenderer.render(
            CatalogRewardsPreview(),
            journey: "tabs",
            stepId: "03-rewards",
            title: "Rewards tab",
            e2eRef: "MainTabView · Rewards / WalletTab"
        )
        try ScreenCatalogRenderer.render(
            CatalogSettingsPreview(),
            journey: "tabs",
            stepId: "04-settings",
            title: "Settings tab",
            e2eRef: "MainTabView · Settings"
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
            CatalogFullChatPreview(),
            journey: "messaging",
            stepId: "07-full-chat",
            title: "Full DM chat (ChatView chrome)",
            e2eRef: "ChatView composite"
        )
        try ScreenCatalogRenderer.render(
            CatalogGroupChatPreview(),
            journey: "messaging",
            stepId: "08-group-chat",
            title: "Group chat thread",
            e2eRef: "GroupChatView composite"
        )
        try ScreenCatalogRenderer.render(
            ChatSettingsSheet(
                contactName: "Jordan",
                conversationId: "catalog-conv",
                preferences: ConversationPreferences(isMuted: false, disappearing: .h24),
                onChange: { _ in },
                onArchiveChange: { _ in }
            ),
            journey: "messaging",
            stepId: "09-chat-settings",
            title: "Chat settings sheet",
            e2eRef: "ChatSettingsSheet"
        )
        try ScreenCatalogRenderer.render(
            MessageActionsSheet(
                messagePreview: "Yes! End-to-end encrypted on this thread.",
                isOwnMessage: true,
                sentWithinEditWindow: true,
                showTranslate: true,
                onAction: { _ in }
            ),
            journey: "messaging",
            stepId: "10-message-actions",
            title: "Message actions sheet",
            e2eRef: "MessageActionsSheet"
        )

        try ScreenCatalogRenderer.render(
            CatalogVIPPathPreview(),
            journey: "vip",
            stepId: "01-path",
            title: "VIP verification path",
            e2eRef: "VIPPathView"
        )
        try ScreenCatalogRenderer.render(
            CatalogVIPStandardIDVPreview(),
            journey: "vip",
            stepId: "02-standard-idv",
            title: "Standard IDV — scan ID",
            e2eRef: "VIPStandardIDVView"
        )
        try ScreenCatalogRenderer.render(
            VIPSuccessView(trustTier: 2, onContinue: {}),
            journey: "vip",
            stepId: "03-success-verified",
            title: "VIP success (T2 Verified)",
            e2eRef: "VIPSuccessView"
        )
        try ScreenCatalogRenderer.render(
            VIPSuccessView(trustTier: 4, onContinue: {}),
            journey: "vip",
            stepId: "04-success-trusted",
            title: "VIP success (T4 Trusted)",
            e2eRef: "VIPSuccessView"
        )
        try ScreenCatalogRenderer.render(
            CatalogVIPSubscriptionPreview(),
            journey: "vip",
            stepId: "05-subscription",
            title: "ECHO VIP subscription",
            e2eRef: "VIPSubscriptionView"
        )

        try ScreenCatalogRenderer.render(
            CatalogLinkDeviceQRPreview(),
            journey: "device",
            stepId: "01-link-qr",
            title: "Link new device — QR",
            e2eRef: "LinkNewDeviceQRView"
        )
        try ScreenCatalogRenderer.render(
            CatalogLinkDeviceScanPreview(),
            journey: "device",
            stepId: "02-link-scan",
            title: "Link device — scan",
            e2eRef: "LinkDeviceScanView"
        )
        try ScreenCatalogRenderer.render(
            CatalogLinkDeviceSuccessPreview(),
            journey: "device",
            stepId: "03-link-success",
            title: "Device linked",
            e2eRef: "LinkDeviceScanView · success"
        )

        try ScreenCatalogRenderer.render(
            CatalogCallConnectingPreview(),
            journey: "calls",
            stepId: "01-connecting",
            title: "Voice call — connecting",
            e2eRef: "CallView · outgoing"
        )
        try ScreenCatalogRenderer.render(
            CatalogCallActivePreview(),
            journey: "calls",
            stepId: "02-active",
            title: "Voice call — active",
            e2eRef: "CallView · active"
        )
        try ScreenCatalogRenderer.render(
            CatalogCallIncomingPreview(),
            journey: "calls",
            stepId: "03-incoming",
            title: "Incoming voice call",
            e2eRef: "CallView · incoming"
        )

        try ScreenCatalogRenderer.render(
            CatalogStakingPreview(),
            journey: "wallet",
            stepId: "01-staking",
            title: "Stake ECHO",
            e2eRef: "StakingView"
        )

        try ScreenCatalogRenderer.render(
            CatalogProposalListPreview(),
            journey: "governance",
            stepId: "01-proposals",
            title: "Governance proposals",
            e2eRef: "ProposalListView"
        )
        try ScreenCatalogRenderer.render(
            CatalogProposalDetailPreview(),
            journey: "governance",
            stepId: "02-proposal-detail",
            title: "Proposal detail",
            e2eRef: "ProposalDetailView composite"
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

        try ScreenCatalogRenderer.render(
            IdentityCardView(
                displayName: "Alex Cromwell",
                username: "alex",
                did: "did:key:z6MkhaXgBZDv7Doxv5c5nPgYt3KuoAihXhWbcXdKq7t6C",
                trustTier: 2
            ),
            journey: "profile",
            stepId: "01-identity-card",
            title: "Identity card",
            e2eRef: "Profile · IdentityCardView"
        )

        try ScreenCatalogRenderer.render(
            BackupView(),
            journey: "utility",
            stepId: "01-backup",
            title: "Backup & recovery",
            e2eRef: "BackupView"
        )
        try ScreenCatalogRenderer.render(
            CatalogSearchPreview(),
            journey: "utility",
            stepId: "02-search",
            title: "Message search",
            e2eRef: "SearchView"
        )
        try ScreenCatalogRenderer.render(
            CatalogNotificationsPreview(),
            journey: "utility",
            stepId: "03-notifications",
            title: "Notification center",
            e2eRef: "NotificationCenterView"
        )
    }
}
#endif
