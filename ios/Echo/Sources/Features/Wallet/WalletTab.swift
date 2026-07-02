#if os(iOS)
// Features/Wallet/WalletTab.swift
// Main wallet view with balance, staking, delegation, and rewards

import SwiftUI

struct WalletTab: View {
    @StateObject private var viewModel: WalletViewModel
    private let screenTitle: String

    init(api: WalletAPIClient, title: String = "Rewards") {
        self.screenTitle = title
        _viewModel = StateObject(wrappedValue: WalletViewModel(api: api))
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    if viewModel.isLoading && viewModel.walletState == nil {
                        BalanceCardSkeleton()
                        SkeletonList(count: 3) { ConversationRowSkeleton() }
                    } else if let state = viewModel.walletState {
                        BalanceCard(state: state)

                        BalanceBreakdown(state: state)

                        if let rewards = state.dailyRewards {
                            DailyRewardsSection(rewards: rewards)
                        }

                        if let emission = viewModel.emissionStatus {
                            EmissionGaugeSection(status: emission)
                        }

                        if let profile = viewModel.founderVesting {
                            FounderVestingSection(
                                profile: profile,
                                onWithdraw: { amount in
                                    guard let lockId = profile.founderLockId else { return }
                                    Task { await viewModel.withdrawVested(amount: amount, lockId: lockId) }
                                }
                            )
                        } else if let vesting = state.vesting {
                            FounderVestingSection(
                                profile: FounderVestingProfile(
                                    vesting: vesting,
                                    revocationEvents: [],
                                    explorerURL: vesting.explorerURL,
                                    founderLockId: state.locks.first(where: \.isFounderVesting)?.id
                                ),
                                onWithdraw: { amount in
                                    guard let lockId = state.locks.first(where: \.isFounderVesting)?.id else { return }
                                    Task { await viewModel.withdrawVested(amount: amount, lockId: lockId) }
                                }
                            )
                        }

                        WalletActionButtons(viewModel: viewModel)

                        StakingPositionsList(
                            locks: state.locks,
                            onUnstake: { stakeId, amount in
                                Task { await viewModel.unstake(stakeId: stakeId, amount: amount) }
                            }
                        )
                    } else {
                        WalletEmptyState {
                            Task { await viewModel.loadWallet() }
                        }
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
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .principal) {
                    HStack(spacing: 6) {
                        EchoLogo(size: 24)
                        Text(screenTitle)
                            .font(Font.Echo.headlineSm)
                            .foregroundStyle(Color.Echo.onSurface)
                    }
                }
                ToolbarItem(placement: .primaryAction) {
                    HStack(spacing: 12) {
                        NavigationLink {
                            TransactionHistoryView()
                        } label: {
                            Image(systemName: "clock.arrow.circlepath")
                                .foregroundStyle(Color.Echo.onSurface)
                        }
                        NavigationLink {
                            StakingDetailView(walletViewModel: viewModel)
                        } label: {
                            Image(systemName: "chart.bar.doc.horizontal")
                                .foregroundStyle(Color.Echo.onSurface)
                        }
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

    private func rewardRow(_ label: String, entry: RewardEarningEntry) -> some View {
        HStack {
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Spacer()
            Text(formatEcho(entry.earned))
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}

// MARK: - Emission Gauge (WO-206)

struct EmissionGaugeSection: View {
    let status: EmissionStatus

    var body: some View {
        GhostBorderCard {
            VStack(alignment: .leading, spacing: 12) {
                HStack {
                    Text("Year \(status.currentYear) Emission")
                        .font(Font.Echo.titleLarge)
                        .foregroundStyle(Color.Echo.onSurface)
                    Spacer()
                    if status.alertThresholdExceeded {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundStyle(Color.Echo.warning)
                    }
                }
                ProgressView(value: min(status.percentConsumed / 100, 1.0))
                    .tint(status.alertThresholdExceeded ? Color.Echo.warning : Color.Echo.primary)
                HStack {
                    Text("\(Int(status.percentConsumed))% of annual cap")
                        .font(Font.Echo.bodySm)
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                    Spacer()
                    Text(formatEcho(status.remainingBudget) + " left")
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                }
            }
        }
    }
}

// MARK: - Founder Vesting Section

struct FounderVestingSection: View {
    let profile: FounderVestingProfile
    let onWithdraw: (Decimal) -> Void

    @State private var showWithdrawSheet = false
    @State private var withdrawAmount = ""
    @State private var showBiometricError = false

    private var vesting: VestingState { profile.vesting }

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

                ProgressView(value: vesting.vestingProgress)
                    .tint(LinearGradient.signature)

                HStack {
                    vestingDetail("Allocated", value: formatEcho(vesting.totalAllocated))
                    Spacer()
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

                if vesting.withdrawable > 0 {
                    Button("Withdraw Vested Tokens") {
                        withdrawAmount = formatEcho(vesting.withdrawable)
                        showWithdrawSheet = true
                    }
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.primary)
                }

                if let url = profile.explorerURL ?? vesting.explorerURL {
                    Link(destination: url) {
                        Label("View on DAG Explorer", systemImage: "link")
                            .font(Font.Echo.bodySm)
                            .foregroundStyle(Color.Echo.primary)
                    }
                }

                if !profile.revocationEvents.isEmpty {
                    VStack(alignment: .leading, spacing: 6) {
                        Text("Revocation History")
                            .font(Font.Echo.labelMd)
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                        ForEach(profile.revocationEvents) { event in
                            Text("\(formatEcho(event.revokedAmount)) revoked · \(event.timestamp.prefix(10))")
                                .font(Font.Echo.bodySm)
                                .foregroundStyle(Color.Echo.warning)
                        }
                    }
                }

                Text("Withdrawals require a 14-day cooldown after submission.")
                    .font(Font.Echo.bodySm)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
        }
        .sheet(isPresented: $showWithdrawSheet) {
            NavigationStack {
                Form {
                    Section {
                        TextField("Amount", text: $withdrawAmount)
                            .keyboardType(.decimalPad)
                    } footer: {
                        Text("Maximum withdrawable: \(formatEcho(vesting.withdrawable)). Withdrawals enter a 14-day cooldown.")
                    }
                }
                .navigationTitle("Withdraw Vested ECHO")
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Cancel") { showWithdrawSheet = false }
                    }
                    ToolbarItem(placement: .confirmationAction) {
                        Button("Confirm") {
                            Task {
                                let bio = DIContainer.shared.resolveBiometricAuth() ?? BiometricAuthManager()
                                do {
                                    let ok = try await bio.authenticate(reason: "Confirm founder vesting withdrawal")
                                    guard ok else { return }
                                    if let amount = Decimal(string: withdrawAmount.trimmingCharacters(in: .whitespaces)),
                                       amount > 0, amount <= vesting.withdrawable {
                                        onWithdraw(amount)
                                        showWithdrawSheet = false
                                    }
                                } catch {
                                    showBiometricError = true
                                }
                            }
                        }
                    }
                }
            }
            .presentationDetents([.medium])
        }
        .alert("Authentication failed", isPresented: $showBiometricError) {
            Button("OK", role: .cancel) {}
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
    @State private var showClaimConfirmation = false

    var body: some View {
        HStack(spacing: 12) {
            NavigationLink {
                StakingView(viewModel: viewModel)
            } label: {
                actionButton("Stake", icon: "lock.shield")
            }

            NavigationLink {
                ValidatorBrowserView(viewModel: viewModel)
            } label: {
                actionButton("Delegate", icon: "person.3")
            }

            Button {
                showClaimConfirmation = true
            } label: {
                actionButton("Claim", icon: "gift")
            }
        }
        .confirmationDialog("Claim Rewards", isPresented: $showClaimConfirmation, titleVisibility: .visible) {
            Button("Claim All Rewards") {
                HapticManager.success()
                Task {
                    await viewModel.claimRewards(types: ["messaging", "referral", "staking", "payment_rail"])
                }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("This will transfer all pending rewards to your available balance.")
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
    @State private var unstakeTarget: TokenLockPosition?

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
                                    unstakeTarget = lock
                                }
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.primaryContainer)
                            }
                        }
                    }
                }
            }
            .confirmationDialog(
                "Unstake \(formatEcho(unstakeTarget?.amount ?? 0)) ECHO?",
                isPresented: Binding(
                    get: { unstakeTarget != nil },
                    set: { if !$0 { unstakeTarget = nil } }
                ),
                titleVisibility: .visible
            ) {
                Button("Unstake", role: .destructive) {
                    if let target = unstakeTarget {
                        HapticManager.medium()
                        onUnstake(target.id, target.amount)
                    }
                }
                Button("Cancel", role: .cancel) { unstakeTarget = nil }
            } message: {
                Text("This will begin the unstaking process. Funds may take time to become available.")
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

// MARK: - Empty State

private struct WalletEmptyState: View {
    let onRetry: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "wallet.pass")
                .font(.system(size: 44))
                .foregroundStyle(Color.Echo.onSurfaceVariant.opacity(0.5))
            Text("Your ECHO balance will appear here after provisioning completes.")
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .multilineTextAlignment(.center)
            Button("Refresh", action: onRetry)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.primaryContainer)
        }
        .frame(maxWidth: .infinity, minHeight: 200)
        .padding()
    }
}

#endif
