// Core/Stargazer/WalletTypes.swift
// Wallet domain types for balance, staking, delegation, and vesting

import Foundation

// MARK: - Wallet Info

enum WalletKeychain {
    static let dagAddressKey = "echo.wallet.dagAddress"
}

struct WalletInfo {
    let address: String
    let publicKey: String
}

// MARK: - Balance

struct BalanceInfo {
    let total: Decimal
    let available: Decimal

    var staked: Decimal { total - available }
}

// MARK: - Token Lock Position (Staking)

struct TokenLockPosition: Identifiable, Equatable {
    let id: String
    let amount: Decimal
    let tier: String
    let lockedUntil: Date
    let vestingType: String?
    let originalAmount: Decimal
    let cliffDate: Date?
    let cliffCompleted: Bool
    let vestedAmount: Decimal
    let withdrawableAmount: Decimal
    let nextUnlockDate: Date?
    let nextUnlockAmount: Decimal
    let delegatedTo: String?

    var isFounderVesting: Bool { vestingType == "founder" }
    var isLocked: Bool { Date() < lockedUntil }
}

// MARK: - Delegation Position

struct DelegationPosition: Identifiable, Equatable {
    let id: String
    let stakeId: String
    let validatorId: String
    let amount: Decimal
    let since: Date
}

// MARK: - Staking Tier

enum StakingTier: String, CaseIterable, Identifiable {
    case bronze, silver, gold, platinum

    var id: String { rawValue }

    var displayName: String {
        rawValue.capitalized
    }

    var durationDays: Int {
        switch self {
        case .bronze: return 30
        case .silver: return 90
        case .gold: return 180
        case .platinum: return 365
        }
    }

    var apr: Double {
        switch self {
        case .bronze: return 5.0
        case .silver: return 8.0
        case .gold: return 12.0
        case .platinum: return 15.0
        }
    }

    var durationLabel: String {
        switch self {
        case .bronze: return "30 days"
        case .silver: return "90 days"
        case .gold: return "180 days"
        case .platinum: return "1 year"
        }
    }
}

// MARK: - Validator Info

struct ValidatorInfo: Identifiable, Equatable {
    let id: String
    let address: String
    let uptimePercent: Double
    let commissionPercent: Double
    let totalDelegated: Decimal
    let delegatorCount: Int
    let layer: String
    let estimatedAPR: Double
}

// MARK: - Auto-Scale Reward State (replaces daily caps per PRD v2.5.1)

/// Current auto-scaling reward rate and network activity state.
/// No per-user daily caps — every message always earns.
/// Rate = Daily Budget / Total Daily Activity Weight
struct AutoScaleRewardState: Equatable {
    let currentRate: Decimal               // Current per-message base rate
    let dailyBudget: Decimal               // Today's emission budget
    let effectiveDailyBudget: Decimal      // Including rollover from low-activity days
    let budgetUsedToday: Decimal           // Total ECHO distributed today
    let remainingToday: Decimal            // Budget remaining today
    let totalActivityToday: Double         // Tier-weighted activity count
    let lastUpdated: Date

    var budgetUtilization: Double {
        guard effectiveDailyBudget > 0 else { return 0 }
        return NSDecimalNumber(decimal: budgetUsedToday / effectiveDailyBudget).doubleValue
    }
}

/// Per-category reward earnings (no cap — just tracking what was earned today)
struct DailyRewardProgress: Equatable {
    let messaging: RewardEarningEntry
    let referrals: RewardEarningEntry
    let staking: RewardEarningEntry
    let paymentRail: RewardEarningEntry
    let autoScaleState: AutoScaleRewardState?
}

struct RewardEarningEntry: Equatable {
    let earned: Decimal
    let count: Int          // Number of rewarded actions today
}

// MARK: - Vesting State (Founders)

struct VestingState: Equatable {
    let role: String
    let totalAllocated: Decimal
    let vested: Decimal
    let locked: Decimal
    let withdrawable: Decimal
    let nextUnlockAmount: Decimal
    let nextUnlockDate: Date?
    let cliffDate: Date
    let cliffCompleted: Bool
    let vestingPercent: Double
}

// MARK: - Wallet State (Aggregate)

struct WalletState: Equatable {
    let totalBalance: Decimal
    let available: Decimal
    let staked: Decimal
    let pendingRewards: Decimal
    let locks: [TokenLockPosition]
    let delegations: [DelegationPosition]
    let dailyRewards: DailyRewardProgress?
    let vesting: VestingState?
}

// MARK: - Stargazer Error

enum StargazerError: Error, LocalizedError {
    case notInitialized
    case transactionFailed(String)
    case insufficientBalance
    case walletCreationFailed

    var errorDescription: String? {
        switch self {
        case .notInitialized: return "Wallet not initialized"
        case .transactionFailed(let msg): return "Transaction failed: \(msg)"
        case .insufficientBalance: return "Insufficient balance"
        case .walletCreationFailed: return "Failed to create wallet"
        }
    }
}
