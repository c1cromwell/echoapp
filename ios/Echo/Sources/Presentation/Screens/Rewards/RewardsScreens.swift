#if os(iOS)
import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

/// Rewards tab: the points hub.
public struct RewardsDashboardView: View {
    public init() {}

    public var body: some View {
        // The Rewards tab is the points hub. There is no wallet or balance behind it.
        if let gamification = DIContainer.shared.resolveGamificationAPI() {
            RewardsHubView(gamification: gamification)
        } else {
            ContentUnavailableView("Rewards are unavailable", systemImage: "gift",
                                   description: Text("Try again in a moment."))
        }
    }
}

struct RewardActivity: Identifiable {
    let id: String
    let type: String
    let amount: Double
    let description: String
    let date: String
}

struct ActivityRow: View {
    let activity: RewardActivity
    
    var icon: String {
        switch activity.type {
        case "messaging":
            return "message.fill"
        case "referral":
            return "person.badge.plus"
        case "transaction":
            return "creditcard.fill"
        default:
            return "star.fill"
        }
    }
    
    var body: some View {
        HStack(spacing: Spacing.md.rawValue) {
            Image(systemName: icon)
                .font(.system(size: 18, weight: .semibold))
                .foregroundColor(.echoPrimary)
                .frame(width: 40)
            
            VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                Text(activity.description)
                    .typographyStyle(.body, color: .echoPrimaryText)
                
                Text(activity.date)
                    .typographyStyle(.caption, color: .echoGray500)
            }
            
            Spacer()
            
            Text("+\(String(format: "%.0f", activity.amount))")
                .typographyStyle(.body, color: .echoSuccess)
                .fontWeight(.semibold)
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
    }
}

// MARK: - Invite (username is the invite)

public struct ReferralView: View {
    @State private var username = ""
    @State private var copied = false

    public init() {}

    private var handle: String { InviteHandle.display(username) }
    private var shareURL: URL? { InviteHandle.shareURL(username: username) }

    public var body: some View {
        ScrollView {
            VStack(spacing: Spacing.xl.rawValue) {
                Text("Your @username is how people find you. There is no separate invite code.")
                    .typographyStyle(.body, color: .echoInk55)
                    .frame(maxWidth: .infinity, alignment: .leading)

                VStack(spacing: Spacing.md.rawValue) {
                    Text("Your username")
                        .typographyStyle(.caption, color: .echoInk55)

                    if handle.isEmpty {
                        Text("Set a username in Profile to invite people.")
                            .typographyStyle(.body, color: .echoInk55)
                            .multilineTextAlignment(.center)
                    } else {
                        HStack(spacing: Spacing.md.rawValue) {
                            Text(handle)
                                .typographyStyle(.h3, color: .echoSignal)

                            Button(action: copyHandle) {
                                Image(systemName: copied ? "checkmark" : "doc.on.doc")
                                    .font(.body.weight(.semibold))
                                    .foregroundStyle(Color.echoSignal)
                            }
                            .accessibilityLabel(copied ? "Copied" : "Copy username")
                        }
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(Spacing.lg.rawValue)
                .background(Color.echoSurface)
                .clipShape(RoundedRectangle(cornerRadius: Spacing.lg.rawValue, style: .continuous))
                .glacialShadow()

                EchoButton(
                    "Share invite",
                    style: .primary,
                    size: .large,
                    isDisabled: shareURL == nil,
                    icon: Image(systemName: "square.and.arrow.up"),
                    action: shareInvite
                )
            }
            .padding(Spacing.lg.rawValue)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .navigationTitle("Invite")
        .navigationBarTitleDisplayMode(.inline)
        .task { username = await CurrentUserSession.currentUsername() }
    }

    private func copyHandle() {
        guard !handle.isEmpty else { return }
        #if os(iOS)
        UIPasteboard.general.string = handle
        HapticManager.light()
        ToastManager.shared.show("Username copied", style: .copied)
        #elseif os(macOS)
        NSPasteboard.general.setString(handle, forType: .string)
        #endif
        copied = true
        DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
            copied = false
        }
    }

    private func shareInvite() {
        guard let url = shareURL else { return }
        #if os(iOS)
        guard let scene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let root = scene.windows.first(where: { $0.isKeyWindow })?.rootViewController else { return }
        var presenter = root
        while let presented = presenter.presentedViewController {
            presenter = presented
        }
        let text = "Add me on Echo: \(handle)"
        let activity = UIActivityViewController(activityItems: [text, url], applicationActivities: nil)
        presenter.present(activity, animated: true)
        #endif
    }
}

// MARK: - Preview

#if DEBUG
struct RewardsScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            RewardsDashboardView()
        }
    }
}
#endif
#endif // os(iOS)
