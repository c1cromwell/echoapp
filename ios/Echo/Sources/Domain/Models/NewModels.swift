// Domain/Models/NewModels.swift
// New data models per iOS Spec v4.2.1 §20

import Foundation

// MARK: - Wallet Activity

struct WalletActivity: Identifiable {
    let id: String
    let type: WalletActivityType
    let amount: Decimal
    let direction: WalletDirection
    let label: String
    let timestamp: Date
    let transactionHash: String?
}

enum WalletActivityType: String, Codable {
    case reward, stake, unstake, delegate, transfer
}

enum WalletDirection: String, Codable {
    case `in`, out
}

// Note: DailyRewards, RewardCapEntry, VestingInfo, and ValidatorInfo
// are already defined in Core/Stargazer/WalletTypes.swift
// Proposal and VotingPower are defined in Features/Governance/GovernanceModels.swift

// MARK: - Search Index Entry (Local persistence)

struct SearchIndexEntry: Identifiable {
    let id: String
    let messageId: String
    let conversationId: String
    let keywords: String
    let timestamp: Date
    let messageType: String
    let senderName: String
    let hasAttachment: Bool
}

// MARK: - Notification Record

struct NotificationRecord: Identifiable {
    let id: String
    let title: String
    let subtitle: String?
    let category: String
    let timestamp: Date
    var isRead: Bool
    let deepLink: String?
}

// MARK: - Staking Tier (per Tokenomics v2)

enum StakingTierSpec: String, CaseIterable {
    case none, bronze, silver, gold, platinum

    var displayName: String { rawValue.capitalized }

    var minimumStake: Decimal {
        switch self {
        case .none: return 0
        case .bronze: return 100
        case .silver: return 1000
        case .gold: return 5000
        case .platinum: return 25000
        }
    }

    var apy: Double {
        switch self {
        case .none: return 0
        case .bronze: return 8.0
        case .silver: return 10.0
        case .gold: return 12.5
        case .platinum: return 15.0
        }
    }

    var multiplier: Double {
        switch self {
        case .none: return 0
        case .bronze: return 1.0
        case .silver: return 1.2
        case .gold: return 1.8
        case .platinum: return 2.5
        }
    }

    static func from(amount: Decimal) -> StakingTierSpec {
        if amount >= 25000 { return .platinum }
        if amount >= 5000 { return .gold }
        if amount >= 1000 { return .silver }
        if amount >= 100 { return .bronze }
        return .none
    }
}
