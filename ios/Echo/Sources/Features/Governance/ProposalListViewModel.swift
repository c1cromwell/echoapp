// Features/Governance/ProposalListViewModel.swift
// ViewModel for the proposals list and voting power

import Foundation

@MainActor
final class ProposalListViewModel: ObservableObject {
    @Published var proposals: [Proposal] = []
    @Published var votingPower: VotingPower?
    @Published var isLoading = false
    @Published var errorMessage: String?

    private let service: GovernanceServiceProtocol
    private let userDID: String

    init(service: GovernanceServiceProtocol, userDID: String) {
        self.service = service
        self.userDID = userDID
    }

    func load() async {
        isLoading = true
        errorMessage = nil

        async let proposalsTask = service.fetchProposals()
        async let powerTask = service.fetchVotingPower(did: userDID)

        do {
            let (fetchedProposals, fetchedPower) = try await (proposalsTask, powerTask)
            proposals = fetchedProposals.sorted { $0.endsAt > $1.endsAt }
            votingPower = fetchedPower
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func submitVote(proposalId: String, value: VoteValue) async throws -> VoteResult {
        let result = try await service.submitVote(proposalId: proposalId, value: value)
        // Reload to get updated tallies
        await load()
        return result
    }
}
