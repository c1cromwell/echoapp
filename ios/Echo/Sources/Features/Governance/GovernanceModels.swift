// Features/Governance/GovernanceModels.swift
// Trust-tier weighted governance domain types

import Foundation

// MARK: - Voting Power

public struct VotingPower: Codable, Equatable {
    public let did: String
    public let trustTier: Int
    public let multiplier: Double
    public let totalStaked: Decimal
    public let weight: Decimal
    public let canVote: Bool

    public init(did: String, trustTier: Int, multiplier: Double, totalStaked: Decimal, weight: Decimal, canVote: Bool) {
        self.did = did
        self.trustTier = trustTier
        self.multiplier = multiplier
        self.totalStaked = totalStaked
        self.weight = weight
        self.canVote = canVote
    }
}

// MARK: - Proposal

public struct Proposal: Identifiable, Codable, Hashable {
    public let id: String
    public let title: String
    public let description: String
    public let type: ProposalType
    public let threshold: ThresholdType
    public let createdBy: String
    public let createdAt: Date
    public let endsAt: Date
    public let status: ProposalStatus
    public let tally: ProposalTally?

    public init(id: String, title: String, description: String, type: ProposalType, threshold: ThresholdType, createdBy: String, createdAt: Date, endsAt: Date, status: ProposalStatus, tally: ProposalTally?) {
        self.id = id
        self.title = title
        self.description = description
        self.type = type
        self.threshold = threshold
        self.createdBy = createdBy
        self.createdAt = createdAt
        self.endsAt = endsAt
        self.status = status
        self.tally = tally
    }
}

// MARK: - Proposal Tally

public struct ProposalTally: Codable, Hashable {
    public let proposalId: String
    public let forWeight: Decimal
    public let againstWeight: Decimal
    public let abstainWeight: Decimal
    public let totalWeight: Decimal
    public let forPercent: Double
    public let voterCount: Int
    public let passed: Bool

    public init(proposalId: String, forWeight: Decimal, againstWeight: Decimal, abstainWeight: Decimal, totalWeight: Decimal, forPercent: Double, voterCount: Int, passed: Bool) {
        self.proposalId = proposalId
        self.forWeight = forWeight
        self.againstWeight = againstWeight
        self.abstainWeight = abstainWeight
        self.totalWeight = totalWeight
        self.forPercent = forPercent
        self.voterCount = voterCount
        self.passed = passed
    }

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

public enum ProposalType: String, Codable, Hashable, CaseIterable {
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

public enum ThresholdType: String, Codable, Hashable {
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

public enum ProposalStatus: String, Codable, Hashable {
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

public enum VoteValue: String, Codable, CaseIterable {
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
