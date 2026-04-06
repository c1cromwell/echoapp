// Features/Governance/GovernanceService.swift
// API service for governance operations

import Foundation

// MARK: - Protocol

protocol GovernanceServiceProtocol {
    func fetchProposals() async throws -> [Proposal]
    func fetchVotingPower(did: String) async throws -> VotingPower
    func submitVote(proposalId: String, value: VoteValue) async throws -> VoteResult
    func createProposal(title: String, description: String, type: ProposalType, threshold: ThresholdType, endsAt: Date) async throws -> Proposal
    func fetchTally(proposalId: String) async throws -> ProposalTally
}

// MARK: - Live Service

actor GovernanceService: GovernanceServiceProtocol {
    private let apiClient: GovernanceAPIClient

    init(apiClient: GovernanceAPIClient) {
        self.apiClient = apiClient
    }

    func fetchProposals() async throws -> [Proposal] {
        try await apiClient.getActiveProposals()
    }

    func fetchVotingPower(did: String) async throws -> VotingPower {
        try await apiClient.getVotingPower(did: did)
    }

    func submitVote(proposalId: String, value: VoteValue) async throws -> VoteResult {
        try await apiClient.postVote(proposalId: proposalId, value: value.rawValue)
    }

    func createProposal(title: String, description: String, type: ProposalType, threshold: ThresholdType, endsAt: Date) async throws -> Proposal {
        try await apiClient.postProposal(title: title, description: description, type: type.rawValue, threshold: threshold.rawValue, endsAt: endsAt)
    }

    func fetchTally(proposalId: String) async throws -> ProposalTally {
        try await apiClient.getTally(proposalId: proposalId)
    }
}

// MARK: - API Client Protocol

protocol GovernanceAPIClient {
    func getActiveProposals() async throws -> [Proposal]
    func getVotingPower(did: String) async throws -> VotingPower
    func postVote(proposalId: String, value: String) async throws -> VoteResult
    func postProposal(title: String, description: String, type: String, threshold: String, endsAt: Date) async throws -> Proposal
    func getTally(proposalId: String) async throws -> ProposalTally
}

// MARK: - Mock API Client (DEBUG)

#if DEBUG
final class MockGovernanceAPIClient: GovernanceAPIClient {
    var proposals: [Proposal] = []
    var votingPower: VotingPower?
    var voteResult: VoteResult?
    var tally: ProposalTally?
    var shouldFail = false

    func getActiveProposals() async throws -> [Proposal] {
        if shouldFail { throw GovernanceError.networkError }
        return proposals
    }

    func getVotingPower(did: String) async throws -> VotingPower {
        if shouldFail { throw GovernanceError.networkError }
        guard let power = votingPower else { throw GovernanceError.noVotingPower }
        return power
    }

    func postVote(proposalId: String, value: String) async throws -> VoteResult {
        if shouldFail { throw GovernanceError.networkError }
        guard let result = voteResult else { throw GovernanceError.voteFailed }
        return result
    }

    func postProposal(title: String, description: String, type: String, threshold: String, endsAt: Date) async throws -> Proposal {
        if shouldFail { throw GovernanceError.networkError }
        return Proposal(
            id: "mock_\(Date().timeIntervalSince1970)",
            title: title,
            description: description,
            type: ProposalType(rawValue: type) ?? .parameterChange,
            threshold: ThresholdType(rawValue: threshold) ?? .simpleMajority,
            createdBy: "mock_did",
            createdAt: Date(),
            endsAt: endsAt,
            status: .active,
            tally: nil
        )
    }

    func getTally(proposalId: String) async throws -> ProposalTally {
        if shouldFail { throw GovernanceError.networkError }
        guard let t = tally else { throw GovernanceError.proposalNotFound }
        return t
    }
}
#endif

// MARK: - Errors

enum GovernanceError: LocalizedError {
    case networkError
    case noVotingPower
    case cannotVote
    case proposalNotFound
    case alreadyVoted
    case voteFailed
    case proposalExpired

    var errorDescription: String? {
        switch self {
        case .networkError: return "Unable to connect to governance service"
        case .noVotingPower: return "Could not determine voting power"
        case .cannotVote: return "You do not meet the requirements to vote"
        case .proposalNotFound: return "Proposal not found"
        case .alreadyVoted: return "You have already voted on this proposal"
        case .voteFailed: return "Vote submission failed"
        case .proposalExpired: return "This proposal's voting period has ended"
        }
    }
}
