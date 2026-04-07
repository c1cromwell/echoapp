import SwiftUI

/// Rewards Dashboard Screen
public struct RewardsDashboardView: View {
    @State private var tokenBalance = 1250.50
    @State private var activities: [RewardActivity] = [
        RewardActivity(id: "1", type: "messaging", amount: 50, description: "Sent 10 messages", date: "Today"),
        RewardActivity(id: "2", type: "referral", amount: 100, description: "Referred John Doe", date: "Yesterday"),
        RewardActivity(id: "3", type: "transaction", amount: 75, description: "Completed transaction", date: "2 days ago")
    ]
    
    let onStaking: () -> Void
    let onReferrals: () -> Void
    let onAchievements: () -> Void
    
    public init(
        onStaking: @escaping () -> Void = {},
        onReferrals: @escaping () -> Void = {},
        onAchievements: @escaping () -> Void = {}
    ) {
        self.onStaking = onStaking
        self.onReferrals = onReferrals
        self.onAchievements = onAchievements
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Rewards",
                    showBackButton: false
                )
                
                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Token Balance Card
                        VStack(spacing: Spacing.lg.rawValue) {
                            Text("ECHO Token Balance")
                                .typographyStyle(.body, color: .echoSecondaryText)
                            
                            Text("\(String(format: "%.2f", tokenBalance))")
                                .typographyStyle(.display, color: .echoPrimary)
                                .font(.system(size: 48, weight: .bold))
                            
                            Text("$\(String(format: "%.2f", tokenBalance * 0.50))")
                                .typographyStyle(.body, color: .echoGray500)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(Spacing.xl.rawValue)
                        .background(
                            LinearGradient(
                                gradient: Gradient(colors: [
                                    Color.echoPrimary.opacity(0.1),
                                    Color.echoPrimary.opacity(0.05)
                                ]),
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .cornerRadius(16)
                        
                        // Action Buttons
                        HStack(spacing: Spacing.md.rawValue) {
                            EchoButton(
                                "Stake",
                                style: .secondary,
                                size: .medium,
                                icon: Image(systemName: "chart.line.uptrend.xyaxis"),
                                action: onStaking
                            )
                            
                            EchoButton(
                                "Refer",
                                style: .secondary,
                                size: .medium,
                                icon: Image(systemName: "person.badge.plus"),
                                action: onReferrals
                            )
                            
                            EchoButton(
                                "Badges",
                                style: .secondary,
                                size: .medium,
                                icon: Image(systemName: "star.fill"),
                                action: onAchievements
                            )
                        }
                        
                        // Recent Activity
                        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                            Text("Recent Activity")
                                .typographyStyle(.h4, color: .echoPrimaryText)
                            
                            VStack(spacing: Spacing.md.rawValue) {
                                ForEach(activities) { activity in
                                    ActivityRow(activity: activity)
                                }
                            }
                        }
                        
                        Spacer()
                    }
                    .echoSpacing(.lg)
                }
            }
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

// MARK: - Staking View

public struct StakingView: View {
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

// MARK: - Referral View

public struct ReferralView: View {
    @Environment(\.dismiss) var dismiss
    @State private var copied = false
    
    let referralCode = "ECHO2024"
    let referralCount = 12
    let earnings = 600.0
    
    public init() {}
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Referrals",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )
                
                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Referral Code
                        VStack(spacing: Spacing.md.rawValue) {
                            Text("Your Referral Code")
                                .typographyStyle(.caption, color: .echoGray500)
                            
                            HStack(spacing: Spacing.md.rawValue) {
                                Text(referralCode)
                                    .typographyStyle(.h4, color: .echoPrimary)
                                    .monospaced()
                                
                                Button(action: {
                                    #if os(iOS)
                                    UIPasteboard.general.string = referralCode
                                    #elseif os(macOS)
                                    NSPasteboard.general.setString(referralCode, forType: .string)
                                    #endif
                                    copied = true
                                    DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                                        copied = false
                                    }
                                }) {
                                    Image(systemName: copied ? "checkmark" : "doc.on.doc")
                                        .font(.system(size: 16, weight: .semibold))
                                        .foregroundColor(.echoPrimary)
                                }
                            }
                        }
                        .frame(maxWidth: .infinity)
                        .padding(Spacing.lg.rawValue)
                        .background(Color.echoSurface)
                        .cornerRadius(12)
                        
                        // Stats
                        HStack(spacing: Spacing.md.rawValue) {
                            StatCard(label: "Referrals", value: "\(referralCount)")
                            StatCard(label: "Earnings", value: "\(String(format: "%.0f", earnings))")
                        }
                        
                        // Share Button
                        EchoButton(
                            "Share Referral Link",
                            style: .primary,
                            size: .large,
                            icon: Image(systemName: "square.and.arrow.up"),
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
