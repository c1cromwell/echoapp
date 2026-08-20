#if os(iOS)
import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

/// Rewards Dashboard Screen — live ECHO balance, staking, and claims.
public struct RewardsDashboardView: View {
    public init() {}

    public var body: some View {
        // Wave R0: the Rewards tab is the value-free gamification hub. The full
        // wallet/balance remains reachable from inside the hub. Falls back to the
        // wallet directly if gamification DI is unavailable.
        if let gamification = DIContainer.shared.resolveGamificationAPI() {
            RewardsHubView(gamification: gamification, walletAPI: resolvedWalletAPI())
        } else {
            WalletTab(api: resolvedWalletAPI())
        }
    }

    @MainActor
    private func resolvedWalletAPI() -> WalletAPIClient {
        DIContainer.shared.resolveWalletAPI() ?? WalletAPIClientStub()
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

// MARK: - Staking View

// Renamed to avoid conflict with Wallet/Staking/StakingView (which takes a WalletViewModel).
public struct RewardsStakingView: View {
    @Environment(\.dismiss) var dismiss
    @State private var stakeAmount = ""
    @State private var stakingPeriod = "30"
    @State private var estimatedRewards = 0.0
    
    let currentStake = 500.0
    
    public init() {}
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Staking",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )
                
                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Current Stake
                        VStack(spacing: Spacing.md.rawValue) {
                            Text("Current Stake")
                                .typographyStyle(.caption, color: .echoGray500)
                            
                            Text("\(String(format: "%.2f", currentStake)) ECHO")
                                .typographyStyle(.h3, color: .echoPrimary)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(Spacing.lg.rawValue)
                        .background(Color.echoSurface)
                        .cornerRadius(12)
                        
                        // Staking Form
                        VStack(spacing: Spacing.lg.rawValue) {
                            EchoTextField(
                                label: "Amount to Stake",
                                placeholder: "0.00",
                                text: $stakeAmount
                            )
                            
                            VStack(spacing: Spacing.md.rawValue) {
                                HStack {
                                    Text("Staking Period")
                                        .typographyStyle(.caption, color: .echoGray500)
                                    
                                    Spacer()
                                    
                                    Picker("Period", selection: $stakingPeriod) {
                                        Text("30 days - 5% APY").tag("30")
                                        Text("90 days - 8% APY").tag("90")
                                        Text("180 days - 12% APY").tag("180")
                                        Text("365 days - 15% APY").tag("365")
                                    }
                                    .pickerStyle(.automatic)
                                }
                            }
                            
                            // APY Info
                            VStack(spacing: Spacing.xs.rawValue) {
                                HStack {
                                    Text("Estimated APY")
                                        .typographyStyle(.body, color: .echoSecondaryText)
                                    Spacer()
                                    Text("5-15%")
                                        .typographyStyle(.body, color: .echoSuccess)
                                }
                            }
                            .padding(Spacing.md.rawValue)
                            .background(Color.echoSurface)
                            .cornerRadius(12)
                        }
                        
                        EchoButton(
                            "Stake ECHO",
                            style: .primary,
                            size: .large,
                            action: {}
                        )
                        
                        Spacer()
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
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
