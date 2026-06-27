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
#endif
