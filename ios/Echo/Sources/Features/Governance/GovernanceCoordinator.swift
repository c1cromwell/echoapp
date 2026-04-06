// Features/Governance/GovernanceCoordinator.swift
// Navigation coordinator for the governance feature

import SwiftUI

struct GovernanceCoordinator: View {
    let service: GovernanceServiceProtocol
    let userDID: String

    @StateObject private var viewModel: ProposalListViewModel

    init(service: GovernanceServiceProtocol, userDID: String) {
        self.service = service
        self.userDID = userDID
        _viewModel = StateObject(wrappedValue: ProposalListViewModel(service: service, userDID: userDID))
    }

    var body: some View {
        NavigationStack {
            ProposalListView(service: service, userDID: userDID)
                .navigationDestination(for: Proposal.self) { proposal in
                    ProposalDetailView(
                        proposal: proposal,
                        votingPower: viewModel.votingPower
                    ) { voteValue in
                        Task {
                            _ = try? await viewModel.submitVote(
                                proposalId: proposal.id,
                                value: voteValue
                            )
                        }
                    }
                }
        }
    }
}
