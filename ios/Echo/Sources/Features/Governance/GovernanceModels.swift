// Features/Governance/GovernanceModels.swift
// Trust-tier weighted governance domain types

import Foundation

// MARK: - Voting Power

struct VotingPower: Codable, Equatable {
    let did: String
    let trustTier: Int
    let multiplier: Double
    let totalStaked: Decimal
    let weight: Decimal
    let canVote: Bool
}

// MARK: - Proposal

struct Proposal: Identifiable, Codable, Hashable {
    let id: String
    let title: String
    let description: String
    let type: ProposalType
    let threshold: ThresholdType
    let createdBy: String
    let createdAt: Date
    let endsAt: Date
    let status: ProposalStatus
    let tally: ProposalTally?
}

// MARK: - Proposal Tally

struct ProposalTally: Codable, Hashable {
    let proposalId: String
    let forWeight: Decimal
    let againstWeight: Decimal
    let abstainWeight: Decimal
    let totalWeight: Decimal
    let forPercent: Double
    let voterCount: Int
    let passed: Bool

    var againstPercent: Double {
        guard totalWeight > 0 else { return 0 }
        return Double(truncating: (againstWeight * 100 / totalWeight) as NSDecimalNumber)
    }

    var abstainPercent: Double {
        guard totalWeight > 0 else { return 0 }
        return Double(truncating: (abstainWeight * 100 / totalWeight) as NSDecimalNumber)
    }
}

// MARK: - Enums

enum ProposalType: String, Codable, Hashable, CaseIterable {
    case protocolUpgrade = "protocol_upgrade"
    case treasuryAllocation = "treasury_allocation"
    case parameterChange = "parameter_change"
    case boardElection = "board_election"

    var displayName: String {
        switch self {
        case .protocolUpgrade: return "Protocol Upgrade"
        case .treasuryAllocation: return "Treasury Allocation"
        case .parameterChange: return "Parameter Change"
        case .boardElection: return "Board Election"
        }
    }
}

enum ThresholdType: String, Codable, Hashable {
    case simpleMajority = "simple_majority"
    case supermajority67 = "supermajority_67"
    case supermajority75 = "supermajority_75"

    var displayName: String {
        switch self {
        case .simpleMajority: return "Simple Majority"
        case .supermajority67: return "Supermajority (67%)"
        case .supermajority75: return "Supermajority (75%)"
        }
    }

    var requiredPercent: Int {
        switch self {
        case .simpleMajority: return 51
        case .supermajority67: return 67
        case .supermajority75: return 75
        }
    }
}

enum ProposalStatus: String, Codable, Hashable {
    case active
    case passed
    case failed
    case executed

    var displayName: String {
        switch self {
        case .active: return "Active"
        case .passed: return "Passed"
        case .failed: return "Failed"
        case .executed: return "Executed"
        }
    }
}

enum VoteValue: String, Codable, CaseIterable {
    case `for` = "for"
    case against = "against"
    case abstain = "abstain"

    var displayName: String {
        rawValue.capitalized
    }

    var systemImage: String {
        switch self {
        case .for: return "checkmark.circle.fill"
        case .against: return "xmark.circle.fill"
        case .abstain: return "minus.circle.fill"
        }
    }
}

// MARK: - Vote Request / Result

struct VoteRequest: Codable {
    let did: String
    let proposalId: String
    let value: String
}

struct VoteResult: Codable {
    let txHash: String
    let weight: Decimal
}

// MARK: - Governance Tier Multipliers

enum GovernanceTier {
    /// Maps trust tier (1-5) to multiplier. Mirrors Scala GovernanceWeightCalculator.
    static let multipliers: [Int: Double] = [
        1: 0.0,  // Unverified: no governance
        2: 0.5,  // Newcomer
        3: 1.0,  // Member
        4: 1.5,  // Verified
        5: 2.0   // Trusted
    ]

    static func multiplier(for tier: Int) -> Double {
        multipliers[tier] ?? 0.0
    }

    static func canVote(tier: Int, totalStaked: Decimal) -> Bool {
        tier >= 2 && totalStaked > 0
    }
}
