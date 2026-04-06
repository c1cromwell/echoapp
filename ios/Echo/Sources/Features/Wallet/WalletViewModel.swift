// Features/Wallet/WalletViewModel.swift
// Manages wallet state: balance, staking, delegation, rewards, vesting

import Foundation
import Combine

@MainActor
class WalletViewModel: ObservableObject {
    @Published var walletState: WalletState?
    @Published var validators: [ValidatorInfo] = []
    @Published var isLoading = false
    @Published var errorMessage: String?

    // Staking
    @Published var stakeAmount: String = ""
    @Published var selectedTier: StakingTier = .bronze
    @Published var isStaking = false

    // Delegation
    @Published var selectedValidator: ValidatorInfo?
    @Published var isDelegating = false

    private let api: WalletAPIClient

    init(api: WalletAPIClient) {
        self.api = api
    }

    // MARK: - Load Wallet State

    func loadWallet() async {
        isLoading = true
        errorMessage = nil

        do {
            let balance = try await api.getBalance()
            let locks = try await api.getTokenLocks()
            let delegations = try await api.getDelegations()

            // Compute vesting for founder locks
            var vesting: VestingState?
            if let founderLock = locks.first(where: { $0.isFounderVesting }) {
                vesting = computeVesting(founderLock)
            }

            walletState = WalletState(
                totalBalance: balance.total,
                available: balance.available,
                staked: balance.staked,
                pendingRewards: 0, // loaded separately
                locks: locks,
                delegations: delegations,
                dailyRewards: nil, // loaded separately
                vesting: vesting
            )
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Load Validators

    func loadValidators() async {
        do {
            validators = try await api.getValidators()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    // MARK: - Stake ECHO

    func stakeEcho() async {
        guard let amount = Decimal(string: stakeAmount), amount > 0 else {
            errorMessage = "Enter a valid amount"
            return
        }

        isStaking = true
        errorMessage = nil

        do {
            _ = try await api.submitTokenLock(amount: amount, tier: selectedTier)
            stakeAmount = ""
            await loadWallet()
        } catch {
            errorMessage = error.localizedDescription
        }

        isStaking = false
    }

    // MARK: - Delegate

    func delegateToValidator(stakeId: String, validatorId: String) async {
        isDelegating = true
        errorMessage = nil

        do {
            _ = try await api.submitStakeDelegation(stakeId: stakeId, validatorId: validatorId)
            await loadWallet()
        } catch {
            errorMessage = error.localizedDescription
        }

        isDelegating = false
    }

    // MARK: - Unstake

    func unstake(stakeId: String, amount: Decimal) async {
        errorMessage = nil
        do {
            _ = try await api.submitWithdrawLock(stakeId: stakeId, amount: amount)
            await loadWallet()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    // MARK: - Claim Rewards

    func claimRewards(types: [String]) async {
        errorMessage = nil
        do {
            _ = try await api.submitRewardClaim(rewardTypes: types)
            await loadWallet()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    // MARK: - Helpers

    private func computeVesting(_ lock: TokenLockPosition) -> VestingState {
        let cliffMonths = 12
        let vestMonths = 48
        let cliffDate = lock.cliffDate ?? lock.lockedUntil.addingTimeInterval(TimeInterval(-((vestMonths - cliffMonths) * 30 * 24 * 3600)))
        let cliffCompleted = Date() > cliffDate

        var vestingPercent = 0.0
        var vestedAmount: Decimal = 0
        if cliffCompleted {
            let elapsed = Date().timeIntervalSince(cliffDate)
            let monthsElapsed = min(Int(elapsed / (30 * 24 * 3600)), vestMonths - cliffMonths)
            vestingPercent = Double(monthsElapsed) / Double(vestMonths - cliffMonths) * 100
            vestedAmount = lock.amount * Decimal(vestingPercent) / 100
        }

        return VestingState(
            role: "Founder",
            totalAllocated: lock.originalAmount,
            vested: vestedAmount,
            locked: lock.amount - vestedAmount,
            withdrawable: lock.withdrawableAmount,
            nextUnlockAmount: lock.nextUnlockAmount,
            nextUnlockDate: lock.nextUnlockDate,
            cliffDate: cliffDate,
            cliffCompleted: cliffCompleted,
            vestingPercent: vestingPercent
        )
    }
}
