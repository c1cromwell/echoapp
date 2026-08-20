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
            .clipShape(RoundedRectangle(cornerRadius: Spacing.lg.rawValue, style: .continuous))
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
            .clipShape(RoundedRectangle(cornerRadius: Spacing.md.rawValue, style: .continuous))
        }
    }
}

// MARK: - Echo Score

/// Compact living score for the Rewards hub. Uses the trust-green ladder, not a second accent.
struct EchoScoreSection: View {
    let snapshot: EchoScoreSnapshot?

    private var score: EchoScoreSnapshot {
        snapshot ?? EchoScoreSnapshot.from(score: 0)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            HStack {
                Text("Echo Score")
                    .typographyStyle(.bodySmall, color: .echoInk55)
                Spacer()
                TrustBadge(level: score.level, size: .small)
            }

            HStack(alignment: .firstTextBaseline, spacing: Spacing.sm.rawValue) {
                Text("\(score.score)")
                    .font(.echomono(34, weight: .semibold))
                    .foregroundStyle(Color.trustColor(for: score.level))
                Text("of 100")
                    .typographyStyle(.caption, color: .echoInk40)
                Spacer()
                VStack(alignment: .trailing, spacing: 2) {
                    Text(String(format: "%.1f×", score.multiplier))
                        .font(.echomono(18, weight: .medium))
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("earn rate")
                        .typographyStyle(.caption, color: .echoInk40)
                }
            }

            ProgressView(value: Double(score.score), total: 100)
                .tint(Color.trustColor(for: score.level))
                .accessibilityLabel("Echo Score \(score.score) of 100")
        }
        .rewardCard()
    }
}

// MARK: - Next unlock

struct NextUnlockSection: View {
    let snapshot: EchoScoreSnapshot?

    var body: some View {
        if let unlock = snapshot?.nextUnlock {
            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                Text("Next unlock")
                    .typographyStyle(.bodySmall, color: .echoInk55)
                Text(unlock.feature)
                    .typographyStyle(.body, color: .echoPrimaryText)
                Text("Tier \(unlock.tier) · \(unlock.pointsNeeded) points to go")
                    .typographyStyle(.caption, color: .echoInk40)
            }
            .rewardCard()
        } else if snapshot != nil {
            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                Text("Next unlock")
                    .typographyStyle(.bodySmall, color: .echoInk55)
                Text("Top tier reached")
                    .typographyStyle(.body, color: .echoPrimaryText)
            }
            .rewardCard()
        }
    }
}

// MARK: - Claim

struct NextClaimSection: View {
    let quest: QuestItem?
    let isClaiming: Bool
    let onClaim: () -> Void

    var body: some View {
        if let quest {
            VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                Text("Claim")
                    .typographyStyle(.bodySmall, color: .echoInk55)
                HStack(alignment: .top, spacing: Spacing.md.rawValue) {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(quest.title)
                            .typographyStyle(.body, color: .echoPrimaryText)
                        Text(quest.description)
                            .typographyStyle(.caption, color: .echoInk55)
                            .lineLimit(2)
                    }
                    Spacer()
                    EchoButton(
                        "Claim",
                        style: .primary,
                        size: .small,
                        isLoading: isClaiming,
                        action: onClaim
                    )
                }
            }
            .rewardCard()
        }
    }
}

// MARK: - Weekly pack

struct WeeklyPackSection: View {
    let pack: WeeklyPack?
    let isOpening: Bool
    let onOpen: () -> Void

    var body: some View {
        if let pack {
            VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                HStack {
                    Text(pack.label)
                        .typographyStyle(.bodySmall, color: .echoInk55)
                    Spacer()
                    if pack.opened {
                        Text("Opened")
                            .typographyStyle(.caption, color: .echoTrustGreen)
                    }
                }
                ForEach(pack.items) { item in
                    VStack(alignment: .leading, spacing: 2) {
                        Text(item.title)
                            .typographyStyle(.body, color: .echoPrimaryText)
                        Text(item.detail)
                            .typographyStyle(.caption, color: .echoInk55)
                    }
                }
                if !pack.opened {
                    EchoButton("Open pack", style: .secondary, size: .small, isLoading: isOpening, action: onOpen)
                }
            }
            .rewardCard()
        }
    }
}

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
        .clipShape(RoundedRectangle(cornerRadius: Spacing.sm.rawValue, style: .continuous))
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

/// The value-free Rewards tab: score, streak, next unlock, claim, then leaderboard.
/// Wallet/balance stays reachable behind a link. Pings daily activity on appear.
struct RewardsHubView: View {
    let gamification: GamificationAPI
    let walletAPI: WalletAPIClient

    @StateObject private var quests: QuestViewModel
    @State private var status: GamificationStatus?
    @State private var streak: StreakInfo?
    @State private var leaderboard: LeaderboardSnapshot?
    @State private var pack: WeeklyPack?
    @State private var isOpeningPack = false
    @State private var selfDID: String = ""

    init(gamification: GamificationAPI, walletAPI: WalletAPIClient) {
        self.gamification = gamification
        self.walletAPI = walletAPI
        _quests = StateObject(wrappedValue: QuestViewModel(api: gamification))
    }

    private var echoScore: EchoScoreSnapshot? { status?.scoreSnapshot }

    var body: some View {
        ScrollView {
            VStack(spacing: Spacing.lg.rawValue) {
                NoValueDisclaimerBanner(status: status)
                EchoScoreSection(snapshot: echoScore)
                StreakSection(streak: streak)
                NextUnlockSection(snapshot: echoScore)
                NextClaimSection(
                    quest: quests.nextClaimableQuest,
                    isClaiming: quests.claimingQuestId != nil
                ) {
                    if let q = quests.nextClaimableQuest {
                        Task { await quests.claim(q) }
                    }
                }
                WeeklyPackSection(pack: pack, isOpening: isOpeningPack) {
                    Task { await openPack() }
                }
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
                    .clipShape(RoundedRectangle(cornerRadius: Spacing.lg.rawValue, style: .continuous))
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
        if let pinged = try? await gamification.pingActivity() {
            streak = pinged.streak
        } else if let s = try? await gamification.fetchStreak() {
            streak = s.streak
        }
        status = try? await gamification.fetchStatus()
        if let lb = try? await gamification.fetchLeaderboard(window: "weekly") {
            leaderboard = lb.snapshot
        }
        await quests.load()
        if let preview = try? await gamification.fetchWeeklyPack() {
            pack = preview.pack
        }
    }

    private func openPack() async {
        isOpeningPack = true
        defer { isOpeningPack = false }
        if let opened = try? await gamification.openWeeklyPack() {
            pack = opened.pack
        }
    }
}
#endif
