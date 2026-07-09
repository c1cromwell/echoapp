#if os(iOS)
import SwiftUI

// Presentational pieces for the value-free Rewards gamification hub (Wave R0).
// Pure views over model data — no networking — so they compose and preview cleanly.
// All styling uses Echo design tokens per the echo-ios-design skill.

// MARK: - Card container

private extension View {
    /// Standard Glacial reward card surface built from tokens.
    func rewardCard() -> some View {
        self
            .padding(Spacing.lg.rawValue)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.echoSurface)
            .clipShape(RoundedRectangle(cornerRadius: Spacing.lg, style: .continuous))
            .glacialShadow()
    }
}

// MARK: - No-value disclaimer

/// Persistent, non-alarming banner stating the launch compliance posture. Rendered
/// whenever the token is not redeemable (interim custody). Driven by GamificationStatus.
struct NoValueDisclaimerBanner: View {
    let status: GamificationStatus?

    private var message: String {
        status?.disclaimer ?? "ECHO earned in-app has no cash value and can't be transferred or redeemed."
    }

    var body: some View {
        if status?.redeemable != true {
            HStack(alignment: .top, spacing: Spacing.sm.rawValue) {
                Image(systemName: "info.circle")
                    .foregroundStyle(Color.echoInfo)
                Text(message)
                    .typographyStyle(.caption, color: .echoInk55)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .padding(Spacing.md.rawValue)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.echoPaperDim)
            .clipShape(RoundedRectangle(cornerRadius: Spacing.md, style: .continuous))
        }
    }
}

// MARK: - Streak

/// Daily-activity streak: current run, milestone, and the multiplier it grants.
struct StreakSection: View {
    let streak: StreakInfo?

    private var days: Int { streak?.currentDays ?? 0 }
    private var multiplier: Double { streak?.multiplier ?? 1.0 }
    private var milestone: String? {
        guard let m = streak?.milestone, !m.isEmpty else { return nil }
        return m
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            Text("Daily streak")
                .typographyStyle(.bodySmall, color: .echoInk55)

            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(days > 0 ? Color.echoSignal : Color.echoInk20)

                VStack(alignment: .leading, spacing: 2) {
                    HStack(alignment: .firstTextBaseline, spacing: 4) {
                        Text("\(days)")
                            .font(.echomono(30, weight: .semibold))
                            .foregroundStyle(Color.echoPrimaryText)
                        Text(days == 1 ? "day" : "days")
                            .typographyStyle(.body, color: .echoInk55)
                    }
                    if let milestone {
                        Text(milestone)
                            .typographyStyle(.caption, color: .echoSignal)
                    }
                }

                Spacer()

                VStack(alignment: .trailing, spacing: 2) {
                    Text(String(format: "%.2f×", multiplier))
                        .font(.echomono(18, weight: .medium))
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("multiplier")
                        .typographyStyle(.caption, color: .echoInk40)
                }
            }
        }
        .rewardCard()
    }
}

// MARK: - Leaderboard

/// Ranked usage leaderboard. Highlights the current user's row; ranks by ECHO earned
/// in the window (display-only). `selfDID` marks the viewer.
struct LeaderboardSection: View {
    let snapshot: LeaderboardSnapshot?
    let selfDID: String

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            HStack {
                Text("Leaderboard")
                    .typographyStyle(.bodySmall, color: .echoInk55)
                Spacer()
                if let window = snapshot?.window {
                    Text(window.capitalized)
                        .typographyStyle(.caption, color: .echoInk40)
                }
            }

            let entries = snapshot?.entries ?? []
            if entries.isEmpty {
                Text("Start messaging and completing quests to climb the ranks.")
                    .typographyStyle(.caption, color: .echoInk40)
                    .padding(.vertical, Spacing.sm.rawValue)
            } else {
                VStack(spacing: Spacing.xs.rawValue) {
                    ForEach(entries) { entry in
                        leaderboardRow(entry, isSelf: entry.did == selfDID)
                    }
                }
            }
        }
        .rewardCard()
    }

    private func leaderboardRow(_ entry: LeaderboardEntry, isSelf: Bool) -> some View {
        HStack(spacing: Spacing.md.rawValue) {
            Text("#\(entry.rank)")
                .font(.echomono(14, weight: .semibold))
                .foregroundStyle(entry.rank <= 3 ? Color.echoSignal : Color.echoInk40)
                .frame(width: 36, alignment: .leading)

            Text(shortDID(entry.did))
                .typographyStyle(.body, color: isSelf ? .echoPrimaryText : .echoInk70)
                .lineLimit(1)
                .truncationMode(.middle)

            Spacer()

            Text(formattedScore(entry.scoreAmount))
                .font(.echomono(14, weight: .regular))
                .foregroundStyle(Color.echoInk55)
        }
        .padding(.vertical, Spacing.sm.rawValue)
        .padding(.horizontal, Spacing.sm.rawValue)
        .background(isSelf ? Color.echoSignalDim.opacity(0.14) : Color.clear)
        .clipShape(RoundedRectangle(cornerRadius: Spacing.sm, style: .continuous))
    }

    private func shortDID(_ did: String) -> String {
        guard did.count > 20 else { return did }
        return String(did.prefix(12)) + "…" + String(did.suffix(6))
    }

    private func formattedScore(_ amount: Decimal) -> String {
        let n = NSDecimalNumber(decimal: amount)
        return String(format: "%.0f ECHO", n.doubleValue)
    }
}

// MARK: - Rewards gamification hub (Wave R0)

/// The value-free Rewards tab: disclaimer + streak + leaderboard, with the full
/// wallet/balance reachable behind a link (nothing removed — demoted, per the launch
/// posture). Pings daily activity on appear.
struct RewardsHubView: View {
    let gamification: GamificationAPI
    let walletAPI: WalletAPIClient

    @State private var status: GamificationStatus?
    @State private var streak: StreakInfo?
    @State private var leaderboard: LeaderboardSnapshot?
    @State private var selfDID: String = ""

    var body: some View {
        ScrollView {
            VStack(spacing: Spacing.lg.rawValue) {
                NoValueDisclaimerBanner(status: status)
                StreakSection(streak: streak)
                LeaderboardSection(snapshot: leaderboard, selfDID: selfDID)

                NavigationLink {
                    WalletTab(api: walletAPI)
                } label: {
                    HStack {
                        Text("Wallet & balance")
                            .typographyStyle(.body, color: .echoPrimaryText)
                        Spacer()
                        Image(systemName: "chevron.right")
                            .foregroundStyle(Color.echoInk40)
                    }
                    .padding(Spacing.lg.rawValue)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.echoSurface)
                    .clipShape(RoundedRectangle(cornerRadius: Spacing.lg, style: .continuous))
                    .glacialShadow()
                }
                .buttonStyle(.plain)
            }
            .padding(Spacing.lg.rawValue)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .navigationTitle("Rewards")
        .navigationBarTitleDisplayMode(.inline)
        .task { await load() }
    }

    private func load() async {
        selfDID = await CurrentUserSession.currentDID() ?? ""
        // Daily-activity signal → streak.
        if let pinged = try? await gamification.pingActivity() {
            streak = pinged.streak
        } else if let s = try? await gamification.fetchStreak() {
            streak = s.streak
        }
        status = try? await gamification.fetchStatus()
        if let lb = try? await gamification.fetchLeaderboard(window: "weekly") {
            leaderboard = lb.snapshot
        }
    }
}
#endif
