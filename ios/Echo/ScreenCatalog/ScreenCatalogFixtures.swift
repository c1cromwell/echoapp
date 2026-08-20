#if os(iOS)
import Foundation
import SwiftUI
@testable import Echo

enum ScreenCatalogFixtures {
    @MainActor
    static func firstRunCoordinator(displayName: String = "alex") -> FirstRunCoordinator {
        let coordinator = FirstRunCoordinator(onComplete: { _, _ in }, onRestoreComplete: { _ in })
        coordinator.displayName = displayName
        return coordinator
    }
}

struct CatalogUsernameClient: UsernameAvailabilityClient {
    let available: Bool

    func checkAvailability(username: String) async throws -> UsernameAvailabilityResult {
        UsernameAvailabilityResult(
            username: username,
            available: available,
            reason: available ? nil : "taken"
        )
    }
}

/// Composite chat thread preview for catalog (not a shipped screen).
struct CatalogChatThreadPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            catalogHeader
            ScrollView {
                VStack(spacing: 12) {
                    MessageBubble(
                        message: "Hey — did you get the invite link?",
                        isSent: false,
                        status: .delivered,
                        timestamp: "9:41 AM"
                    )
                    MessageBubble(
                        message: "Yes! End-to-end encrypted on this thread.",
                        isSent: true,
                        status: .read,
                        deliveryStatus: .read,
                        timestamp: "9:42 AM"
                    )
                    TypingIndicatorView(label: "Jordan is typing…")
                }
                .padding(.vertical, 16)
            }
            ReactionPickerView(onSelect: { _ in }, onDismiss: {})
                .padding(.bottom, 8)
            catalogComposer
        }
        .background(Color.Echo.surface)
    }

    private var catalogHeader: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.echoSignal.opacity(0.2))
                .frame(width: 36, height: 36)
                .overlay(Text("J").font(.headline).foregroundStyle(Color.echoSignal))
            VStack(alignment: .leading, spacing: 2) {
                Text("Jordan")
                    .font(.system(size: 17, weight: .semibold))
                Text("Verified · end-to-end encrypted")
                    .font(.caption)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer()
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color.Echo.surfaceContainerLow)
    }

    private var catalogComposer: some View {
        HStack(spacing: 10) {
            Image(systemName: "plus.circle.fill")
                .font(.title2)
                .foregroundStyle(Color.echoSignal)
            Text("Message")
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Spacer()
            Image(systemName: "mic.fill")
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color.Echo.surfaceContainerLow)
    }
}

// MARK: - Privacy catalog composites (Form/List do not rasterize in ImageRenderer)

struct CatalogSOCKSProxyPreview: View {
    var body: some View {
        CatalogSettingsScaffold(title: "SOCKS proxy") {
            CatalogToggleRow(title: "Route traffic through SOCKS5", isOn: true)
            CatalogFieldRow(label: "Host", value: "127.0.0.1")
            CatalogFieldRow(label: "Port", value: "9050")
            Text("Compatible with Tor (default port 9050). Reconnect chats after changing proxy settings.")
                .font(.footnote)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            CatalogPrimaryButton(title: "Save")
        }
    }
}

struct CatalogPQHybridPreview: View {
    var body: some View {
        CatalogSettingsScaffold(title: "Post-quantum") {
            CatalogToggleRow(title: "Post-quantum handshake", isOn: true)
            Text("Uses ML-KEM-768 plus P-256 for new 1:1 ratchet sessions. Hybrid only — never PQ-only.")
                .font(.footnote)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            HStack {
                Text("Device support")
                Spacer()
                Text("Available").foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            .font(.body)
        }
    }
}

struct CatalogHiddenFolderPreview: View {
    var body: some View {
        CatalogSettingsScaffold(title: "Hidden folder") {
            CatalogPickerRow(label: "Lock after", value: "2 minutes")
            CatalogToggleRow(title: "Lock on screenshot", isOn: true)
            CatalogPickerRow(label: "Notifications", value: "Suppressed")
            Text("Hidden chats stay on this device and require biometrics to open.")
                .font(.footnote)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}

struct CatalogPrivacyHubPreview: View {
    var body: some View {
        CatalogSettingsScaffold(title: "Privacy") {
            CatalogNavRow(title: "Privacy & security", subtitle: "Typing, read receipts")
            CatalogNavRow(title: "On-device AI", subtitle: "Translation & summaries")
            CatalogNavRow(title: "SOCKS proxy (Tor)", subtitle: "Optional transport")
            CatalogNavRow(title: "Post-quantum encryption", subtitle: "ML-KEM hybrid handshake")
            CatalogNavRow(title: "Data & deletion", subtitle: "Account controls")
        }
    }
}

private struct CatalogSettingsScaffold<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text(title)
                .font(.system(size: 28, weight: .bold))
                .foregroundStyle(Color.Echo.onSurface)
                .padding(.horizontal, 20)
                .padding(.top, 16)
                .padding(.bottom, 12)
            VStack(alignment: .leading, spacing: 16) {
                content()
            }
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.Echo.surfaceContainerLow)
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .padding(.horizontal, 16)
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.Echo.surface)
    }
}

private struct CatalogToggleRow: View {
    let title: String
    let isOn: Bool

    var body: some View {
        Toggle(isOn: .constant(isOn)) {
            Text(title).foregroundStyle(Color.Echo.onSurface)
        }
        .tint(Color.echoSignal)
    }
}

private struct CatalogFieldRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label).foregroundStyle(Color.Echo.onSurfaceVariant)
            Spacer()
            Text(value).foregroundStyle(Color.Echo.onSurface)
        }
    }
}

private struct CatalogPickerRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label).foregroundStyle(Color.Echo.onSurface)
            Spacer()
            Text(value).foregroundStyle(Color.echoSignal)
            Image(systemName: "chevron.right")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}

private struct CatalogNavRow: View {
    let title: String
    let subtitle: String

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(title).font(.body.weight(.medium)).foregroundStyle(Color.Echo.onSurface)
                Text(subtitle).font(.caption).foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer()
            Image(systemName: "chevron.right")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}

private struct CatalogPrimaryButton: View {
    let title: String

    var body: some View {
        Text(title)
            .font(.body.weight(.semibold))
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
            .background(Color.echoSignal)
            .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct CatalogInviteAcceptPreview: View {
    let inviteCode: String

    var body: some View {
        VStack(spacing: 20) {
            Text("Accept contact invite?")
                .font(.headline)
                .foregroundStyle(Color.Echo.onSurface)
            Text("Code: \(inviteCode)")
                .font(.caption.monospaced())
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            CatalogPrimaryButton(title: "Accept invite")
        }
        .padding(24)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.Echo.surface)
    }
}

struct CatalogMessagesHubGroupsPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Text("Messages")
                    .font(.system(size: 28, weight: .bold))
                Spacer()
            }
            .padding(.horizontal, 20)
            .padding(.top, 12)

            HStack(spacing: 4) {
                segmentChip("Chats", selected: false)
                segmentChip("Groups", selected: true)
            }
            .padding(4)
            .background(Color.Echo.surfaceContainerLow)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .padding(.horizontal, 20)
            .padding(.vertical, 12)

            VStack(spacing: 0) {
                HStack {
                    Image(systemName: "plus.circle.fill")
                        .foregroundStyle(Color.echoSignal)
                    Text("New group")
                        .font(.system(size: 15, weight: .semibold))
                    Spacer()
                }
                .padding(.horizontal, 20)
                .padding(.vertical, 14)

                hubRow(initials: "TE", name: "Team Echo", preview: "Welcome to the group")
                hubRow(initials: "PR", name: "Product", preview: "Ship checklist updated")
            }
            Spacer(minLength: 0)
        }
        .background(Color.echoPaper)
    }

    private func segmentChip(_ title: String, selected: Bool) -> some View {
        Text(title)
            .font(.system(size: 14, weight: selected ? .semibold : .medium))
            .foregroundStyle(selected ? Color.echoPaper : Color.echoInk55)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 8)
            .background(selected ? Color.echoInk : Color.clear)
            .clipShape(RoundedRectangle(cornerRadius: 9))
    }

    private func hubRow(initials: String, name: String, preview: String) -> some View {
        HStack(spacing: 12) {
            Text(initials)
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(.white)
                .frame(width: 48, height: 48)
                .background(Color.echoSignal)
                .clipShape(Circle())
            VStack(alignment: .leading, spacing: 2) {
                Text(name).font(.system(size: 16, weight: .semibold))
                Text(preview).font(.system(size: 14)).foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer()
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 12)
    }
}

struct CatalogGroupCreatePreview: View {
    var body: some View {
        CatalogSettingsScaffold(title: "New group") {
            CatalogFieldRow(label: "Name", value: "Weekend crew")
            Text("Members")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            CatalogNavRow(title: "Jordan", subtitle: "Selected")
            CatalogNavRow(title: "Sam", subtitle: "Selected")
            CatalogNavRow(title: "Riley", subtitle: "Tap to add")
            CatalogPrimaryButton(title: "Create")
        }
    }
}

struct CatalogUsernameSearchPreview: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Add by username")
                .font(.system(size: 28, weight: .bold))
                .padding(.horizontal, 20)
                .padding(.top, 16)
            HStack {
                Image(systemName: "magnifyingglass")
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                Text("@jordan")
                    .foregroundStyle(Color.Echo.onSurface)
                Spacer()
            }
            .padding(14)
            .background(Color.Echo.surfaceContainerLow)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .padding(.horizontal, 20)

            VStack(spacing: 12) {
                searchResult(name: "Jordan Lee", handle: "@jordan", tier: "Verified")
                searchResult(name: "Jordan Kim", handle: "@jordank", tier: "Basic")
            }
            .padding(.horizontal, 20)
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.Echo.surface)
    }

    private func searchResult(name: String, handle: String, tier: String) -> some View {
        HStack {
            Circle()
                .fill(Color.echoSignal.opacity(0.15))
                .frame(width: 44, height: 44)
                .overlay(Text(String(name.prefix(1))).font(.headline))
            VStack(alignment: .leading, spacing: 2) {
                Text(name).font(.body.weight(.semibold))
                Text(handle).font(.caption).foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer()
            Text(tier)
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.echoTrustGreen)
        }
        .padding(12)
        .background(Color.Echo.surfaceContainerLow)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Glacial login composites (state is keychain-driven at runtime)

private struct CatalogGlacialLoginScaffold<Content: View>: View {
    @ViewBuilder let content: () -> Content

    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()
            VStack(spacing: 0) {
                HStack(spacing: 8) {
                    EchoRippleMark(size: 22, color: .echoSignal)
                    Text("ECHO")
                        .font(.system(size: 18, weight: .bold))
                        .tracking(1)
                        .foregroundStyle(Color.echoInk)
                    Spacer()
                }
                .padding(.horizontal, 28)
                .padding(.top, 12)

                Spacer()
                content()
                Spacer()

                VStack(spacing: 16) {
                    Text("By continuing, you agree to Echo's Terms and Privacy Policy.")
                        .font(.system(size: 11))
                        .foregroundStyle(Color.echoInk40)
                        .multilineTextAlignment(.center)
                }
                .padding(.horizontal, 28)
                .padding(.bottom, 48)
            }
        }
    }
}

struct CatalogGlacialLoginNormalPreview: View {
    var body: some View {
        CatalogGlacialLoginScaffold {
            VStack(spacing: 24) {
                EchoRippleMark(size: 120, color: .echoSignal)
                Text("Private Messaging, Always")
                    .font(.system(size: 17))
                    .foregroundStyle(Color.echoInk70)
                Text("@alex")
                    .font(.echomono(11))
                    .foregroundStyle(Color.echoInk40)

                HStack(spacing: 14) {
                    Image(systemName: "faceid")
                        .font(.system(size: 28, weight: .light))
                        .foregroundStyle(Color.echoSignal)
                        .frame(width: 44, height: 44)
                        .background(Color.echoSignal.opacity(0.12))
                        .clipShape(Circle())
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Secure Login")
                            .font(.system(size: 18, weight: .semibold))
                        Text("FaceID, TouchID, or PIN")
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoInk55)
                    }
                    Spacer()
                    Image(systemName: "chevron.right")
                        .foregroundStyle(Color.echoInk40)
                }
                .padding(.horizontal, 20)
                .padding(.vertical, 16)
                .background(Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 16))
                .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
                .padding(.horizontal, 28)
            }
        }
    }
}

struct CatalogGlacialLoginNoAccountPreview: View {
    var body: some View {
        CatalogGlacialLoginScaffold {
            VStack(alignment: .leading, spacing: 12) {
                Text("Welcome to Echo.")
                    .font(.system(size: 32, weight: .semibold))
                    .tracking(-0.8)
                Text("No password. No email. Your face is your key.")
                    .font(.system(size: 14))
                    .foregroundStyle(Color.echoInk55)
                HStack {
                    Text("Create account")
                        .font(.system(size: 15, weight: .semibold))
                    Spacer()
                    Text("→").font(.system(size: 18))
                }
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, 18)
                .padding(.vertical, 15)
                .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                .padding(.top, 8)
            }
            .padding(.horizontal, 28)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
}

struct CatalogGlacialLoginSoftLockPreview: View {
    var body: some View {
        CatalogGlacialLoginScaffold {
            VStack(alignment: .leading, spacing: 0) {
                Text("@alex")
                    .font(.echomono(11))
                    .foregroundStyle(Color.echoInk40)
                    .padding(.bottom, 8)
                Text("Too many\nattempts.")
                    .font(.system(size: 32, weight: .semibold))
                    .tracking(-0.8)
                    .lineSpacing(2)
                Text("For your security, Face ID is locked. Use your device passcode to continue.")
                    .font(.system(size: 13))
                    .foregroundStyle(Color.echoInk55)
                    .padding(.top, 10)
                HStack {
                    Image(systemName: "lock.fill")
                    Text("Use device passcode")
                        .font(.system(size: 15, weight: .semibold))
                    Spacer()
                    Text("→")
                }
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, 18)
                .padding(.vertical, 14)
                .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                .padding(.top, 24)
            }
            .padding(.horizontal, 28)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
}

// MARK: - Main tab composites (List/NavigationStack do not rasterize cleanly)

struct CatalogContactsListPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            catalogNavBar(title: "Contacts", trailingIcon: "plus")
            VStack(spacing: 16) {
                catalogSearchField(placeholder: "Search contacts")
                catalogSegmentedPicker
                VStack(spacing: 0) {
                    catalogContactRow(name: "Jordan Lee", handle: "@jordan", tier: "Verified", favorite: true)
                    catalogContactRow(name: "Sam Rivera", handle: "@sam", tier: "Trusted", favorite: false)
                    catalogContactRow(name: "Riley Chen", handle: "@riley", tier: "Basic", favorite: false)
                }
            }
            .padding(16)
            Spacer(minLength: 0)
            CatalogGlacialTabBarPreview(selected: .contacts)
        }
        .background(Color.echoPaper)
    }

    private var catalogSegmentedPicker: some View {
        HStack(spacing: 4) {
            ForEach(["All", "Inner Circle", "Trusted"], id: \.self) { label in
                Text(label)
                    .font(.system(size: 12, weight: label == "All" ? .semibold : .medium))
                    .foregroundStyle(label == "All" ? Color.echoPaper : Color.echoInk55)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 8)
                    .background(label == "All" ? Color.echoInk : Color.clear)
                    .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
        .padding(4)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}

struct CatalogMessagesHubPopulatedPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            catalogNavBar(title: "Messages")
            VStack(spacing: 16) {
                catalogSearchField(placeholder: "Search conversations")
                VStack(spacing: 0) {
                    catalogHubRow(name: "Jordan", preview: "See you at 3?", time: "9:41 AM", unread: 2, online: true)
                    catalogHubRow(name: "Sam", preview: "Invite link worked!", time: "Yesterday", unread: 0, online: false)
                    catalogHubRow(name: "Team Echo", preview: "Ship checklist updated", time: "Mon", unread: 0, online: false)
                }
            }
            .padding(16)
            Spacer(minLength: 0)
            CatalogGlacialTabBarPreview(selected: .messages)
        }
        .background(Color.echoPaper)
    }
}

struct CatalogRewardsPreview: View {
    private let walletState = WalletState(
        totalBalance: 12_450.75,
        available: 8_200.50,
        staked: 4_250.25,
        pendingRewards: 125.00,
        locks: [],
        delegations: [],
        dailyRewards: nil,
        vesting: nil
    )

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 6) {
                EchoLogo(size: 24)
                Text("Rewards")
                    .font(Font.Echo.headlineSm)
                Spacer()
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 16)

            ScrollView {
                VStack(spacing: 20) {
                    EchoScoreSection(snapshot: EchoScoreSnapshot.from(score: 34))
                    StreakSection(streak: StreakInfo(
                        did: "did:key:catalog",
                        currentDays: 8,
                        longestDays: 8,
                        lastActive: nil,
                        multiplier: 1.10,
                        milestone: "Week One"
                    ))
                    NextUnlockSection(snapshot: EchoScoreSnapshot.from(score: 34))
                    WeeklyPackSection(
                        pack: WeeklyPack(
                            weekKey: "weekly:2026-W34",
                            label: "Week One pack",
                            opened: false,
                            items: [
                                WeeklyPackItem(kind: "activity", title: "Week recorded", detail: "Your activity this week is counted toward standing."),
                                WeeklyPackItem(kind: "badge", title: "Week One", detail: "Seven-day streak standing."),
                            ],
                            openedAt: nil
                        ),
                        isOpening: false,
                        onOpen: {}
                    )
                    RewardHubNavRow(title: "Invite with @username")
                    BalanceCard(state: walletState)
                }
                .padding(20)
            }
            CatalogGlacialTabBarPreview(selected: .rewards)
        }
        .background(Color.echoPaper)
    }
}

struct CatalogSettingsPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            catalogNavBar(title: "Settings")
            ScrollView {
                VStack(spacing: 16) {
                    VStack(spacing: 0) {
                        catalogSettingsRow(icon: "star.circle.fill", title: "ECHO VIP", value: "Free tier")
                        catalogDivider
                        catalogSettingsRow(icon: "person.badge.plus", title: "Add contact by username")
                        catalogDivider
                        catalogSettingsRow(icon: "person.fill", title: "Account")
                        catalogDivider
                        catalogSettingsRow(icon: "lock.fill", title: "Privacy & Security")
                        catalogDivider
                        catalogSettingsRow(icon: "bell.fill", title: "Notifications")
                        catalogDivider
                        catalogSettingsRow(icon: "paintbrush.fill", title: "Appearance")
                        catalogDivider
                        catalogSettingsRow(icon: "internaldrive.fill", title: "Storage & Data")
                        catalogDivider
                        catalogSettingsRow(icon: "questionmark.circle.fill", title: "Help & Support")
                        catalogDivider
                        catalogSettingsRow(icon: "info.circle.fill", title: "About Echo")
                    }
                    .background(Color.echoPaperDim)
                    .clipShape(RoundedRectangle(cornerRadius: 12))

                    Text("Sign Out")
                        .font(.body.weight(.semibold))
                        .foregroundStyle(Color.echoAlert)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .background(Color.echoPaperDim)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(16)
            }
            CatalogGlacialTabBarPreview(selected: .settings)
        }
        .background(Color.echoPaper)
    }
}

struct CatalogSearchPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            HStack(spacing: 12) {
                Text("Cancel")
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.primaryContainer)
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundStyle(Color.Echo.outline)
                    Text("invite link")
                        .font(Font.Echo.bodyLarge)
                    Spacer()
                }
                .padding(12)
                .background(RoundedRectangle(cornerRadius: 16).fill(Color.Echo.surfaceContainerLow))
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(["All", "Messages", "Files", "Links"], id: \.self) { filter in
                        Text(filter)
                            .font(.system(size: 13, weight: filter == "All" ? .semibold : .medium))
                            .padding(.horizontal, 14)
                            .padding(.vertical, 8)
                            .background(filter == "All" ? Color.echoInk : Color.echoPaperDim)
                            .foregroundStyle(filter == "All" ? Color.echoPaper : Color.echoInk55)
                            .clipShape(Capsule())
                    }
                }
                .padding(.horizontal, 16)
            }
            .padding(.bottom, 12)

            VStack(spacing: 12) {
                catalogSearchResult(title: "Jordan Lee", snippet: "…did you get the invite link?", time: "9:41 AM")
                catalogSearchResult(title: "Sam Rivera", snippet: "Invite link worked!", time: "Yesterday")
            }
            .padding(.horizontal, 16)
            Spacer(minLength: 0)
        }
        .background(Color.Echo.surface)
    }
}

struct CatalogNotificationsPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            HStack {
                Text("Notifications")
                    .font(.system(size: 28, weight: .bold))
                Spacer()
                Text("Mark All")
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 12)

            VStack(alignment: .leading, spacing: 0) {
                Text("TODAY")
                    .font(.system(size: 10, weight: .bold))
                    .tracking(2)
                    .foregroundStyle(Color.Echo.outline)
                    .padding(.horizontal, 20)
                    .padding(.top, 16)
                    .padding(.bottom, 8)
                catalogNotificationRow(title: "New message from Jordan", body: "See you at 3?", unread: true)
                catalogNotificationRow(title: "Trust tier upgraded", body: "You're now Verified (T2)", unread: true)
                Text("YESTERDAY")
                    .font(.system(size: 10, weight: .bold))
                    .tracking(2)
                    .foregroundStyle(Color.Echo.outline)
                    .padding(.horizontal, 20)
                    .padding(.top, 16)
                    .padding(.bottom, 8)
                catalogNotificationRow(title: "Contact request accepted", body: "Sam joined your network", unread: false)
            }
            Spacer(minLength: 0)
        }
        .background(Color.Echo.surface)
    }
}

enum CatalogTabSelection {
    case messages, contacts, rewards, settings
}

struct CatalogGlacialTabBarPreview: View {
    let selected: CatalogTabSelection

    var body: some View {
        HStack(spacing: 0) {
            tabItem(.messages, icon: "ellipsis.message", label: "Messages")
            tabItem(.contacts, icon: "person.2", label: "Contacts")
            tabItem(.rewards, icon: "gift", label: "Rewards")
            tabItem(.settings, icon: "gearshape", label: "Settings")
        }
        .padding(.top, 10)
        .padding(.bottom, 28)
        .background(Color.echoPaper.shadow(color: Color.black.opacity(0.06), radius: 12, y: -4))
    }

    private func tabItem(_ tab: CatalogTabSelection, icon: String, label: String) -> some View {
        let isSelected = selected == tab
        return VStack(spacing: 4) {
            Image(systemName: isSelected ? "\(icon).fill" : icon)
                .font(.system(size: 22))
                .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)
            Text(label)
                .font(.system(size: 10, weight: isSelected ? .semibold : .medium))
                .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)
        }
        .frame(maxWidth: .infinity)
    }
}

// MARK: - Shared catalog chrome helpers

private func catalogNavBar(title: String, trailingIcon: String? = nil) -> some View {
    HStack {
        Text(title)
            .font(.system(size: 28, weight: .bold))
        Spacer()
        if let trailingIcon {
            Image(systemName: trailingIcon)
                .font(.system(size: 18, weight: .semibold))
                .foregroundStyle(Color.echoSignal)
        }
    }
    .padding(.horizontal, 20)
    .padding(.vertical, 12)
}

private func catalogSearchField(placeholder: String) -> some View {
    HStack {
        Image(systemName: "magnifyingglass")
            .foregroundStyle(Color.echoInk40)
        Text(placeholder)
            .foregroundStyle(Color.echoInk40)
        Spacer()
    }
    .padding(12)
    .background(Color.echoPaperDim)
    .clipShape(RoundedRectangle(cornerRadius: 12))
}

private func catalogContactRow(name: String, handle: String, tier: String, favorite: Bool) -> some View {
    HStack(spacing: 12) {
        Circle()
            .fill(Color.echoSignal.opacity(0.15))
            .frame(width: 44, height: 44)
            .overlay(Text(String(name.prefix(1))).font(.headline))
        VStack(alignment: .leading, spacing: 2) {
            HStack(spacing: 4) {
                Text(name).font(.body.weight(.semibold))
                if favorite {
                    Image(systemName: "star.fill")
                        .font(.caption2)
                        .foregroundStyle(Color.echoWarning)
                }
            }
            Text(handle).font(.caption).foregroundStyle(Color.echoInk55)
        }
        Spacer()
        Text(tier)
            .font(.caption.weight(.semibold))
            .foregroundStyle(Color.echoTrustGreen)
    }
    .padding(.vertical, 12)
}

private func catalogHubRow(name: String, preview: String, time: String, unread: Int, online: Bool) -> some View {
    HStack(spacing: 12) {
        ZStack(alignment: .bottomTrailing) {
            Circle()
                .fill(Color.echoSignal.opacity(0.2))
                .frame(width: 48, height: 48)
                .overlay(Text(String(name.prefix(1))).font(.headline))
            if online {
                Circle()
                    .fill(Color.echoTrustGreen)
                    .frame(width: 12, height: 12)
                    .overlay(Circle().stroke(Color.echoPaper, lineWidth: 2))
            }
        }
        VStack(alignment: .leading, spacing: 2) {
            Text(name).font(.system(size: 16, weight: .semibold))
            Text(preview).font(.system(size: 14)).foregroundStyle(Color.echoInk55).lineLimit(1)
        }
        Spacer()
        VStack(alignment: .trailing, spacing: 4) {
            Text(time).font(.caption).foregroundStyle(Color.echoInk40)
            if unread > 0 {
                Text("\(unread)")
                    .font(.system(size: 11, weight: .bold))
                    .foregroundStyle(.white)
                    .frame(width: 20, height: 20)
                    .background(Color.echoSignal)
                    .clipShape(Circle())
            }
        }
    }
    .padding(.vertical, 12)
}

private var catalogDivider: some View {
    Rectangle()
        .fill(Color.echoHair)
        .frame(height: 1)
        .padding(.leading, 52)
}

private func catalogSettingsRow(icon: String, title: String, value: String? = nil) -> some View {
    HStack(spacing: 12) {
        Image(systemName: icon)
            .foregroundStyle(Color.echoSignal)
            .frame(width: 24)
        Text(title)
            .font(.body)
        Spacer()
        if let value {
            Text(value)
                .font(.caption)
                .foregroundStyle(Color.echoInk55)
        }
        Image(systemName: "chevron.right")
            .font(.caption.weight(.semibold))
            .foregroundStyle(Color.echoInk40)
    }
    .padding(.horizontal, 16)
    .padding(.vertical, 14)
}

private func catalogSearchResult(title: String, snippet: String, time: String) -> some View {
    VStack(alignment: .leading, spacing: 4) {
        HStack {
            Text(title).font(.body.weight(.semibold))
            Spacer()
            Text(time).font(.caption).foregroundStyle(Color.echoInk40)
        }
        Text(snippet).font(.subheadline).foregroundStyle(Color.echoInk55)
    }
    .padding(14)
    .background(Color.echoPaperDim)
    .clipShape(RoundedRectangle(cornerRadius: 12))
}

private func catalogNotificationRow(title: String, body: String, unread: Bool) -> some View {
    HStack(alignment: .top, spacing: 12) {
        Circle()
            .fill(unread ? Color.echoSignal : Color.echoInk10)
            .frame(width: 8, height: 8)
            .padding(.top, 6)
        VStack(alignment: .leading, spacing: 2) {
            Text(title)
                .font(.body.weight(unread ? .semibold : .regular))
            Text(body)
                .font(.subheadline)
                .foregroundStyle(Color.echoInk55)
        }
        Spacer()
    }
    .padding(.horizontal, 20)
    .padding(.vertical, 10)
}

// MARK: - Extended catalog (messaging, VIP, device, calls, wallet, governance)

enum CatalogGovernanceMocks {
    static var sampleTally: ProposalTally {
        ProposalTally(
            proposalId: "prop-001",
            forWeight: 72_500,
            againstWeight: 18_200,
            abstainWeight: 9_300,
            totalWeight: 100_000,
            forPercent: 72.5,
            voterCount: 184,
            passed: false
        )
    }

    static var sampleProposal: Proposal {
        Proposal(
            id: "prop-001",
            title: "Increase community rewards allocation",
            description: "Raise the annual community rewards cap by 5% to support early messaging incentives through Q4.",
            type: .treasuryAllocation,
            threshold: .supermajority67,
            createdBy: "did:key:z6Mk…",
            createdAt: Date(),
            endsAt: Date().addingTimeInterval(864_000),
            status: .active,
            tally: sampleTally
        )
    }

    static var votingPower: VotingPower {
        VotingPower(
            did: "did:key:z6Mk…",
            trustTier: 2,
            multiplier: 1.5,
            totalStaked: 4_250,
            weight: 6_375,
            canVote: true
        )
    }
}

struct CatalogFullChatPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 10) {
                Image(systemName: "chevron.left")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundStyle(Color.echoSignal)
                Text("J")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(.white)
                    .frame(width: 32, height: 32)
                    .background(Color.echoSignal)
                    .clipShape(Circle())
                HStack(spacing: 4) {
                    Text("Jordan")
                        .font(.system(size: 17, weight: .semibold))
                    TrustBadge(level: "Verified", size: .small)
                }
                Spacer()
                Image(systemName: "phone.fill")
                    .foregroundStyle(Color.echoSignal)
                Image(systemName: "video.fill")
                    .foregroundStyle(Color.echoSignal)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(Color.echoPaperDim)

            SecureThreadIndicator()

            ScrollView {
                VStack(spacing: 12) {
                    MessageBubble(
                        message: "Hey — did you get the invite link?",
                        isSent: false,
                        status: .delivered,
                        timestamp: "9:41 AM"
                    )
                    MessageBubble(
                        message: "Yes! End-to-end encrypted on this thread.",
                        isSent: true,
                        status: .read,
                        deliveryStatus: .read,
                        timestamp: "9:42 AM"
                    )
                    MessageBubble(
                        message: "Perfect. I'll send the doc after standup.",
                        isSent: false,
                        status: .delivered,
                        timestamp: "9:43 AM"
                    )
                    TypingIndicatorView(label: "Jordan is typing…")
                }
                .padding(.vertical, 16)
            }

            HStack(spacing: 10) {
                Image(systemName: "plus.circle.fill")
                    .font(.title2)
                    .foregroundStyle(Color.echoSignal)
                Text("Message")
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                Spacer()
                Image(systemName: "mic.fill")
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(Color.echoPaperDim)
        }
        .background(Color.Echo.surface)
    }
}

struct CatalogGroupChatPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Text("Team Echo")
                    .font(.system(size: 17, weight: .semibold))
                Spacer()
                Image(systemName: "person.3")
                    .foregroundStyle(Color.echoSignal)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(Color.echoPaperDim)

            ScrollView {
                VStack(spacing: 12) {
                    catalogGroupBubble("Welcome to the group!", outgoing: false, author: "Jordan")
                    catalogGroupBubble("Ship checklist updated in the doc.", outgoing: true, author: nil)
                    catalogGroupBubble("I'll review tonight.", outgoing: false, author: "Sam")
                }
                .padding(16)
            }

            HStack {
                Image(systemName: "plus.circle.fill")
                    .foregroundStyle(Color.echoSignal)
                TextField("Message", text: .constant(""))
                    .textFieldStyle(.roundedBorder)
                Text("Send")
                    .foregroundStyle(Color.echoSignal)
            }
            .padding()
        }
        .background(Color.echoPaper)
    }
}

private func catalogGroupBubble(_ text: String, outgoing: Bool, author: String?) -> some View {
    HStack {
        if outgoing { Spacer(minLength: 40) }
        VStack(alignment: outgoing ? .trailing : .leading, spacing: 4) {
            if let author {
                Text(author)
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(Color.echoInk55)
            }
            Text(text)
                .padding(12)
                .background(outgoing ? Color.echoSignal.opacity(0.15) : Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        if !outgoing { Spacer(minLength: 40) }
    }
}

struct CatalogVIPPathPreview: View {
    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()
            VStack(alignment: .leading, spacing: 0) {
                SecureThreadIndicator()
                VStack(alignment: .leading, spacing: 6) {
                    Text("Verify your identity")
                        .font(.system(size: 28, weight: .semibold))
                    Text("Choose how you'd like to prove who you are.")
                        .font(.system(size: 14))
                        .foregroundStyle(Color.echoInk55)
                }
                .padding(.top, 28)
                .padding(.horizontal, 24)

                VStack(spacing: 14) {
                    catalogVIPPathCard(
                        icon: "creditcard.fill",
                        title: "Digital ID",
                        subtitle: "Use a mobile driver's licence or digital wallet credential.",
                        badge: "T4 Trusted",
                        primary: true
                    )
                    catalogVIPPathCard(
                        icon: "doc.text.viewfinder",
                        title: "Standard Verification",
                        subtitle: "Scan your licence or passport, take a selfie, and add a recovery phone number.",
                        badge: "T2 Verified",
                        primary: false
                    )
                }
                .padding(.horizontal, 24)
                .padding(.top, 28)

                Spacer(minLength: 16)

                Text("Skip for now — I'll verify later in Settings")
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk55)
                    .frame(maxWidth: .infinity)
                    .padding(.bottom, 32)
            }
        }
    }
}

private func catalogVIPPathCard(
    icon: String,
    title: String,
    subtitle: String,
    badge: String,
    primary: Bool
) -> some View {
    HStack(alignment: .top, spacing: 14) {
        Image(systemName: icon)
            .font(.system(size: 22))
            .foregroundStyle(primary ? Color.echoSignal : Color.echoInk55)
            .frame(width: 32)
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.system(size: 17, weight: .semibold))
            Text(subtitle)
                .font(.system(size: 13))
                .foregroundStyle(Color.echoInk55)
            Text(badge)
                .font(.system(size: 11, weight: .bold))
                .foregroundStyle(Color.echoTrustGreen)
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color.echoTrustGreen.opacity(0.12))
                .clipShape(Capsule())
        }
        Spacer()
        Image(systemName: "chevron.right")
            .foregroundStyle(Color.echoInk40)
    }
    .padding(16)
    .background(primary ? Color.echoPaperDim : Color.echoPaper)
    .clipShape(RoundedRectangle(cornerRadius: 16))
    .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
}

struct CatalogVIPStandardIDVPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            HStack(spacing: 6) {
                Capsule().fill(Color.echoSignal).frame(height: 4)
                Capsule().fill(Color.echoHair).frame(height: 4)
            }
            .padding(.horizontal, 24)
            .padding(.top, 20)

            VStack(alignment: .leading, spacing: 6) {
                Text("Scan your ID")
                    .font(.system(size: 26, weight: .semibold))
                Text("Take a photo of your driver's licence or passport.")
                    .font(.system(size: 13))
                    .foregroundStyle(Color.echoInk55)
            }
            .padding(.horizontal, 24)
            .padding(.top, 20)

            RoundedRectangle(cornerRadius: 16)
                .stroke(Color.echoSignal, style: StrokeStyle(lineWidth: 2, dash: [8]))
                .frame(height: 200)
                .overlay(
                    Image(systemName: "doc.text.viewfinder")
                        .font(.system(size: 48))
                        .foregroundStyle(Color.echoInk40)
                )
                .padding(.horizontal, 24)
                .padding(.top, 24)

            Spacer()

            Text("Capture document")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 15)
                .background(Color.echoInk)
                .clipShape(RoundedRectangle(cornerRadius: 14))
                .padding(.horizontal, 24)
                .padding(.bottom, 32)
        }
        .background(Color.echoPaper)
    }
}

struct CatalogVIPSubscriptionPreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                HStack(spacing: 10) {
                    Image(systemName: "star.circle.fill")
                        .font(.system(size: 32))
                        .foregroundStyle(Color.echoTrustGreen)
                    VStack(alignment: .leading, spacing: 2) {
                        Text("ECHO VIP")
                            .font(.system(size: 22, weight: .semibold))
                        Text("Get premium features")
                            .font(.system(size: 14))
                            .foregroundStyle(Color.echoInk55)
                    }
                }
                .padding(16)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 14))

                VStack(alignment: .leading, spacing: 12) {
                    Text("WHAT'S INCLUDED")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.echoInk55)
                    catalogVIPBenefitRow("2× API rate limits")
                    catalogVIPBenefitRow("Premium themes & app icon")
                    catalogVIPBenefitRow("VIP trust badge on profile")
                }

                VStack(spacing: 8) {
                    Text("$9.99 / month")
                        .font(.system(size: 28, weight: .semibold))
                    Text("Subscribe with Apple")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .background(Color.echoSignal)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(16)
                .background(Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 14))
            }
            .padding(20)
        }
        .background(Color.echoPaper)
    }
}

private func catalogVIPBenefitRow(_ text: String) -> some View {
    HStack(spacing: 12) {
        Image(systemName: "checkmark.circle.fill")
            .foregroundStyle(Color.echoSignal)
        Text(text)
            .font(.system(size: 15))
    }
}

struct CatalogLinkDeviceQRPreview: View {
    var body: some View {
        VStack(spacing: 20) {
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.white)
                .frame(width: 220, height: 220)
                .overlay(
                    Image(systemName: "qrcode")
                        .font(.system(size: 120))
                        .foregroundStyle(Color.echoInk)
                )
            Text("Scan with your new iPhone or iPad")
                .font(.headline)
            Text("Expires in 5 minutes. Message history syncs after the new device links.")
                .font(.footnote)
                .foregroundStyle(Color.echoInk55)
                .multilineTextAlignment(.center)
                .padding(.horizontal)
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.echoPaper)
    }
}

struct CatalogLinkDeviceScanPreview: View {
    var body: some View {
        VStack(spacing: 16) {
            Text("Scan the QR on your signed-in device, or paste the link token.")
                .font(.subheadline)
                .foregroundStyle(Color.echoInk55)
                .multilineTextAlignment(.center)
                .padding(.horizontal)

            RoundedRectangle(cornerRadius: 12)
                .fill(Color.echoInk.opacity(0.85))
                .frame(height: 240)
                .overlay(
                    RoundedRectangle(cornerRadius: 8)
                        .stroke(Color.echoSignal, lineWidth: 2)
                        .frame(width: 180, height: 180)
                )

            HStack {
                Text("echo-link-token-…")
                    .font(.caption.monospaced())
                    .foregroundStyle(Color.echoInk55)
                Spacer()
            }
            .padding(12)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 8))
            .padding(.horizontal)

            Text("Link this device")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(Color.echoSignal)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .padding(.horizontal)
        }
        .padding(.top, 24)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.echoPaper)
    }
}

struct CatalogLinkDeviceSuccessPreview: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 48))
                .foregroundStyle(Color.echoTrustGreen)
            Text("Device linked")
                .font(.headline)
            Text("Sign in with Face ID on this device. Your message history will download after unlock.")
                .multilineTextAlignment(.center)
                .foregroundStyle(Color.echoInk55)
                .padding(.horizontal)
            Text("Continue")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
                .padding(.horizontal, 32)
                .padding(.vertical, 12)
                .background(Color.echoSignal)
                .clipShape(Capsule())
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.echoPaper)
    }
}

struct CatalogCallConnectingPreview: View {
    var body: some View {
        catalogCallScaffold(stateLabel: "Calling…", showControls: false)
    }
}

struct CatalogCallActivePreview: View {
    var body: some View {
        catalogCallScaffold(stateLabel: "00:42", showControls: true)
    }
}

struct CatalogCallIncomingPreview: View {
    var body: some View {
        VStack(spacing: 0) {
            EncryptionBadge()
                .padding(.top, 60)
            Spacer()
            VStack(spacing: 16) {
                Circle()
                    .fill(Color.echoSignal.opacity(0.15))
                    .frame(width: 80, height: 80)
                    .overlay(Text("J").font(.title.weight(.semibold)))
                Text("Jordan")
                    .font(Font.Echo.headlineSm)
                Text("Incoming voice call")
                    .font(.system(size: 32, weight: .light))
                    .foregroundStyle(Color.Echo.outline)
            }
            Spacer()
            HStack(spacing: 48) {
                VStack(spacing: 8) {
                    Image(systemName: "phone.down.fill")
                        .foregroundStyle(.white)
                        .frame(width: 64, height: 64)
                        .background(Circle().fill(Color.Echo.error))
                    Text("Decline")
                        .font(.caption)
                }
                VStack(spacing: 8) {
                    Image(systemName: "phone.fill")
                        .foregroundStyle(.white)
                        .frame(width: 64, height: 64)
                        .background(Circle().fill(Color.echoTrustGreen))
                    Text("Accept")
                        .font(.caption)
                }
            }
            .padding(.bottom, 48)
        }
        .background(Color.Echo.surface.icyBackground())
    }
}

private func catalogCallScaffold(stateLabel: String, showControls: Bool) -> some View {
    VStack(spacing: 0) {
        EncryptionBadge()
            .padding(.top, 60)
        Spacer()
        VStack(spacing: 16) {
            Circle()
                .fill(Color.echoSignal.opacity(0.15))
                .frame(width: 80, height: 80)
                .overlay(Text("J").font(.title.weight(.semibold)))
            Text("Jordan")
                .font(Font.Echo.headlineSm)
            Text(stateLabel)
                .font(.system(size: 32, weight: .light))
                .foregroundStyle(Color.Echo.outline)
                .monospacedDigit()
        }
        Spacer()
        if showControls {
            VStack(spacing: 24) {
                HStack(spacing: 24) {
                    catalogCallControlIcon("mic.fill", active: true)
                    catalogCallControlIcon("speaker.wave.3.fill", active: true)
                    catalogCallControlIcon("rectangle.on.rectangle", active: false)
                }
                Image(systemName: "phone.down.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(.white)
                    .frame(width: 64, height: 64)
                    .background(Circle().fill(Color.Echo.error))
            }
            .padding(.bottom, 40)
        }
    }
    .background(Color.Echo.surface.icyBackground())
}

private func catalogCallControlIcon(_ icon: String, active: Bool) -> some View {
    Image(systemName: icon)
        .font(.system(size: 22))
        .foregroundStyle(active ? Color.echoInk : Color.echoInk55)
        .frame(width: 56, height: 56)
        .background(Circle().fill(Color.echoPaperDim))
}

struct CatalogStakingPreview: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 24) {
            Text("Stake ECHO")
                .font(.system(size: 28, weight: .bold))
                .padding(.horizontal, 20)
                .padding(.top, 16)

            VStack(alignment: .leading, spacing: 12) {
                Text("Stake Amount")
                    .font(Font.Echo.titleLarge)
                HStack {
                    Text("1,000.00")
                        .font(Font.Echo.displayMedium)
                    Text("ECHO")
                        .font(Font.Echo.bodyLarge)
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                }
                Text("Available: 8,200.50 ECHO")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 20))
            .padding(.horizontal, 20)

            VStack(alignment: .leading, spacing: 12) {
                Text("Select Tier")
                    .font(Font.Echo.titleLarge)
                catalogStakingTierRow("30 days", apr: "8", selected: false)
                catalogStakingTierRow("90 days", apr: "12", selected: true)
                catalogStakingTierRow("180 days", apr: "18", selected: false)
            }
            .padding(.horizontal, 20)

            Text("Stake ECHO")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(Color.echoSignal)
                .clipShape(RoundedRectangle(cornerRadius: 14))
                .padding(.horizontal, 20)

            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.echoPaper)
    }
}

private func catalogStakingTierRow(_ duration: String, apr: String, selected: Bool) -> some View {
    HStack {
        VStack(alignment: .leading, spacing: 4) {
            Text(duration)
                .font(Font.Echo.bodyLarge)
            Text("Lock period")
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        Spacer()
        Text("\(apr)% APR")
            .foregroundStyle(Color.Echo.primaryContainer)
        if selected {
            Image(systemName: "checkmark.circle.fill")
                .foregroundStyle(Color.Echo.primaryContainer)
        }
    }
    .padding(16)
    .background(selected ? Color.Echo.surfaceContainerHigh : Color.Echo.surfaceContainer)
    .clipShape(RoundedRectangle(cornerRadius: 20))
}

struct CatalogProposalListPreview: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Governance")
                .font(.system(size: 28, weight: .bold))
                .padding(.horizontal, 20)
                .padding(.top, 16)

            VStack(alignment: .leading, spacing: 8) {
                Text("YOUR VOTING POWER")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundStyle(Color.echoInk55)
                Text("6,375")
                    .font(.system(size: 32, weight: .semibold))
                Text("Trust tier T2 · 1.5× multiplier")
                    .font(.subheadline)
                    .foregroundStyle(Color.echoInk55)
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .padding(.horizontal, 20)

            catalogProposalCard(
                type: "TREASURY ALLOCATION",
                title: "Increase community rewards allocation",
                forPercent: 72,
                againstPercent: 18
            )
            .padding(.horizontal, 20)

            catalogProposalCard(
                type: "PARAMETER CHANGE",
                title: "Adjust PQ handshake default",
                forPercent: 54,
                againstPercent: 31
            )
            .padding(.horizontal, 20)

            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.echoPaper)
    }
}

struct CatalogProposalDetailPreview: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 24) {
            Text("Proposal")
                .font(.system(size: 28, weight: .bold))
                .padding(.horizontal, 20)
                .padding(.top, 16)

            VStack(alignment: .leading, spacing: 12) {
                Text("TREASURY ALLOCATION")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundStyle(Color.Echo.primaryContainer)
                Text("Increase community rewards allocation")
                    .font(.system(size: 22, weight: .bold))
                Text("Raise the annual community rewards cap by 5% to support early messaging incentives through Q4.")
                    .font(.system(size: 15))
                    .foregroundStyle(Color.echoInk55)
                    .lineSpacing(4)

                HStack(spacing: 0) {
                    Rectangle().fill(Color.echoTrustGreen).frame(width: 220, height: 8)
                    Rectangle().fill(Color.echoAlert.opacity(0.7)).frame(width: 55, height: 8)
                }
                .clipShape(Capsule())

                HStack {
                    Text("72% For").font(.caption).foregroundStyle(Color.echoTrustGreen)
                    Spacer()
                    Text("18% Against").font(.caption).foregroundStyle(Color.echoAlert)
                }

                HStack(spacing: 12) {
                    Text("Vote For")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(Color.echoTrustGreen)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    Text("Vote Against")
                        .font(.system(size: 15, weight: .semibold))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(Color.echoPaperDim)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
            .padding(20)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .padding(.horizontal, 20)

            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Color.echoPaper)
    }
}

private func catalogProposalCard(type: String, title: String, forPercent: Int, againstPercent: Int) -> some View {
    VStack(alignment: .leading, spacing: 12) {
        Text(type)
            .font(.system(size: 10, weight: .bold))
            .foregroundStyle(Color.Echo.primaryContainer)
        Text(title)
            .font(.system(size: 18, weight: .bold))
        HStack(spacing: 0) {
            Rectangle()
                .fill(Color.echoTrustGreen)
                .frame(width: CGFloat(forPercent) * 2.5, height: 8)
            Rectangle()
                .fill(Color.echoAlert.opacity(0.7))
                .frame(width: CGFloat(againstPercent) * 2.5, height: 8)
        }
        .clipShape(Capsule())
        HStack {
            Text("\(forPercent)% For")
                .font(.caption)
                .foregroundStyle(Color.echoTrustGreen)
            Spacer()
            Text("\(againstPercent)% Against")
                .font(.caption)
                .foregroundStyle(Color.echoAlert)
        }
    }
    .padding(16)
    .background(Color.echoPaperDim)
    .clipShape(RoundedRectangle(cornerRadius: 16))
}
#endif
