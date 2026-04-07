// Features/Wallet/WalletTab.swift
// Main wallet view with balance, staking, delegation, and rewards

import SwiftUI

struct WalletTab: View {
    @StateObject private var viewModel: WalletViewModel

    init(api: WalletAPIClient) {
        _viewModel = StateObject(wrappedValue: WalletViewModel(api: api))
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    if let state = viewModel.walletState {
                        BalanceCard(state: state)

                        BalanceBreakdown(state: state)

                        if let rewards = state.dailyRewards {
                            DailyRewardsSection(rewards: rewards)
                        }

                        if let vesting = state.vesting {
                            FounderVestingSection(vesting: vesting)
                        }

                        WalletActionButtons(viewModel: viewModel)

                        StakingPositionsList(
                            locks: state.locks,
                            onUnstake: { stakeId, amount in
                                Task { await viewModel.unstake(stakeId: stakeId, amount: amount) }
                            }
                        )
                    }

                    if let error = viewModel.errorMessage {
                        Text(error)
                            .font(Font.Echo.bodySm)
                            .foregroundStyle(Color.Echo.error)
                            .padding()
                    }
                }
                .padding()
            }
            .icyBackground()
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .principal) {
                    HStack(spacing: 6) {
                        EchoLogo(size: 24)
                        Text("Wallet")
                            .font(Font.Echo.headlineSm)
                            .foregroundStyle(Color.Echo.onSurface)
                    }
                }
            }
            .refreshable {
                await viewModel.loadWallet()
            }
            .task {
                await viewModel.loadWallet()
            }
        }
    }
}

// MARK: - Balance Card

struct BalanceCard: View {
    let state: WalletState

    var body: some View {
        GhostBorderCard {
            VStack(spacing: 8) {
                Text("Total Balance")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)

                Text(formatEcho(state.totalBalance))
                    .font(Font.Echo.displayMedium)
                    .foregroundStyle(Color.Echo.onSurface)

                Text("ECHO")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .frame(maxWidth: .infinity)
        }
    }
}

// MARK: - Balance Breakdown

struct BalanceBreakdown: View {
    let state: WalletState

    var body: some View {
        GhostBorderCard {
            VStack(spacing: 12) {
                breakdownRow("Available", amount: state.available, color: Color.Echo.primaryContainer)
                breakdownRow("Staked", amount: state.staked, color: Color.Echo.secondary)
                if state.pendingRewards > 0 {
                    breakdownRow("Pending Rewards", amount: state.pendingRewards, color: Color.Echo.skyBlue)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private func breakdownRow(_ label: String, amount: Decimal, color: Color) -> some View {
        HStack {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Spacer()
            Text(formatEcho(amount))
                .font(Font.Echo.bodyLarge)
                .foregroundStyle(Color.Echo.onSurface)
        }
    }
}

// MARK: - Daily Rewards Section

struct DailyRewardsSection: View {
    let rewards: DailyRewardProgress

    var body: some View {
        GhostBorderCard {
            VStack(alignment: .leading, spacing: 12) {
                Text("Daily Rewards")
                    .font(Font.Echo.titleLarge)
                    .foregroundStyle(Color.Echo.onSurface)

                rewardRow("Messaging", entry: rewards.messaging)
                rewardRow("Referrals", entry: rewards.referrals)
                rewardRow("Staking", entry: rewards.staking)
                rewardRow("Payment Rail", entry: rewards.paymentRail)
            }
        }
    }

    private func rewardRow(_ label: String, entry: RewardCapEntry) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(label)
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                Spacer()
                Text("\(formatEcho(entry.earned)) / \(formatEcho(entry.cap))")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            ProgressView(value: entry.progress)
                .tint(Color.Echo.primaryContainer)
        }
    }
}

// MARK: - Founder Vesting Section

struct FounderVestingSection: View {
    let vesting: VestingState

    var body: some View {
        GhostBorderCard {
            VStack(alignment: .leading, spacing: 12) {
                HStack {
                    Text("Founder Vesting")
                        .font(Font.Echo.titleLarge)
                        .foregroundStyle(Color.Echo.onSurface)
                    Spacer()
                    Text(vesting.role)
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.primaryContainer)
                }

                ProgressView(value: vesting.vestingPercent / 100)
                    .tint(LinearGradient.signature)

                HStack {
                    vestingDetail("Vested", value: formatEcho(vesting.vested))
                    Spacer()
                    vestingDetail("Locked", value: formatEcho(vesting.locked))
                    Spacer()
                    vestingDetail("Withdrawable", value: formatEcho(vesting.withdrawable))
                }

                if !vesting.cliffCompleted {
                    Text("Cliff: \(vesting.cliffDate.formatted(date: .abbreviated, time: .omitted))")
                        .font(Font.Echo.bodySm)
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                } else if let nextDate = vesting.nextUnlockDate {
                    Text("Next unlock: \(formatEcho(vesting.nextUnlockAmount)) on \(nextDate.formatted(date: .abbreviated, time: .omitted))")
                        .font(Font.Echo.bodySm)
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                }
            }
        }
    }

    private func vestingDetail(_ label: String, value: String) -> some View {
        VStack(spacing: 2) {
            Text(value)
                .font(Font.Echo.bodyLarge)
                .foregroundStyle(Color.Echo.onSurface)
            Text(label)
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}

// MARK: - Wallet Action Buttons

struct WalletActionButtons: View {
    @ObservedObject var viewModel: WalletViewModel

    var body: some View {
        HStack(spacing: 12) {
            NavigationLink {
                WalletStakingView(viewModel: viewModel)
            } label: {
                actionButton("Stake", icon: "lock.shield")
            }

            NavigationLink {
                ValidatorBrowserView(viewModel: viewModel)
            } label: {
                actionButton("Delegate", icon: "person.3")
            }

            Button {
                Task {
                    await viewModel.claimRewards(types: ["messaging", "referral", "staking", "payment_rail"])
                }
            } label: {
                actionButton("Claim", icon: "gift")
            }
        }
    }

    private func actionButton(_ title: String, icon: String) -> some View {
        VStack(spacing: 6) {
            Image(systemName: icon)
                .font(.system(size: 20))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 48, height: 48)
                .background(Color.Echo.surfaceContainerLow)
                .clipShape(Circle())
                .ghostBorder(opacity: 0.10)

            Text(title)
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .frame(maxWidth: .infinity)
    }
}

// MARK: - Staking Positions List

struct StakingPositionsList: View {
    let locks: [TokenLockPosition]
    let onUnstake: (String, Decimal) -> Void

    var body: some View {
        if !locks.isEmpty {
            VStack(alignment: .leading, spacing: 12) {
                Text("Active Positions")
                    .font(Font.Echo.titleLarge)
                    .foregroundStyle(Color.Echo.onSurface)

                ForEach(locks) { lock in
                    GhostBorderCard {
                        HStack {
                            VStack(alignment: .leading, spacing: 4) {
                                Text("\(formatEcho(lock.amount)) ECHO")
                                    .font(Font.Echo.bodyLarge)
                                    .foregroundStyle(Color.Echo.onSurface)
                                Text("\(lock.tier.capitalized) • Until \(lock.lockedUntil.formatted(date: .abbreviated, time: .omitted))")
                                    .font(Font.Echo.labelMd)
                                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                            }
                            Spacer()
                            if !lock.isLocked {
                                Button("Unstake") {
                                    onUnstake(lock.id, lock.amount)
                                }
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.primaryContainer)
                            }
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Formatting

func formatEcho(_ amount: Decimal) -> String {
    let formatter = NumberFormatter()
    formatter.numberStyle = .decimal
    formatter.minimumFractionDigits = 2
    formatter.maximumFractionDigits = 2
    return formatter.string(from: amount as NSDecimalNumber) ?? "0.00"
}
