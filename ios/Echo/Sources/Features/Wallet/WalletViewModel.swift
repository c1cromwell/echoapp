#if os(iOS)
// Features/Wallet/WalletViewModel.swift
// Manages wallet state: balance, staking, delegation, rewards, vesting

import Foundation
import Combine

@MainActor
class WalletViewModel: ObservableObject {
    @Published var walletState: WalletState?
    @Published var emissionStatus: EmissionStatus?
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
            async let wallet = api.fetchWalletState()
            async let emission = api.fetchEmissionStatus()
            async let vesting = api.fetchVesting()
            var state = try await wallet
            if let founderVesting = try await vesting {
                state = WalletState(
                    totalBalance: state.totalBalance,
                    available: state.available,
                    staked: state.staked,
                    pendingRewards: state.pendingRewards,
                    locks: state.locks,
                    delegations: state.delegations,
                    dailyRewards: state.dailyRewards,
                    vesting: founderVesting
                )
            }
            walletState = state
            emissionStatus = try await emission
        } catch {
            errorMessage = userFacingWalletError(error)
        }

        isLoading = false
    }

    private func userFacingWalletError(_ error: Error) -> String {
        if let stargazer = error as? StargazerError {
            return stargazer.localizedDescription
        }
        if let apiError = error as? APIError {
            return apiError.localizedDescription
        }
        return "Could not load wallet. Pull to refresh or try again later."
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
        let vestMonths = 36  // 1/36th monthly vesting after cliff (per Tokenomics v2.0)
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
#endif
