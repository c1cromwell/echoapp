// Features/Governance/ProposalListView.swift
// Active proposals list with tally progress and voting power card

import SwiftUI

struct ProposalListView: View {
    @StateObject private var viewModel: ProposalListViewModel

    init(service: GovernanceServiceProtocol, userDID: String) {
        _viewModel = StateObject(wrappedValue: ProposalListViewModel(service: service, userDID: userDID))
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                // Voting power card
                if let power = viewModel.votingPower {
                    GovernanceWeightView(power: power)
                }

                // Error state
                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.custom("Inter", size: 14))
                        .foregroundStyle(Color.Echo.error)
                        .padding()
                }

                // Proposals
                if viewModel.proposals.isEmpty && !viewModel.isLoading {
                    Text("No active proposals")
                        .font(.custom("Inter", size: 16))
                        .foregroundStyle(Color.Echo.outline)
                        .padding(.top, 40)
                } else {
                    ForEach(viewModel.proposals) { proposal in
                        NavigationLink(value: proposal) {
                            ProposalCard(proposal: proposal)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
        }
        .navigationTitle("Governance")
        .task { await viewModel.load() }
        .refreshable { await viewModel.load() }
    }
}

// MARK: - Proposal Card

struct ProposalCard: View {
    let proposal: Proposal

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Type badge
            Text(proposal.type.displayName.uppercased())
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(1.0)
                .foregroundStyle(Color.Echo.primaryContainer)

            // Title
            Text(proposal.title)
                .font(.custom("Inter", size: 18))
                .fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)
                .multilineTextAlignment(.leading)

            // Tally progress bar
            if let tally = proposal.tally, tally.totalWeight > 0 {
                TallyProgressBar(tally: tally)

                HStack {
                    Text("\(Int(tally.forPercent))% For")
                        .font(.custom("Inter", size: 12))
                        .foregroundStyle(Color.Echo.primaryContainer)

                    Spacer()

                    Text("\(tally.voterCount) voter\(tally.voterCount == 1 ? "" : "s")")
                        .font(.custom("Inter", size: 12))
                        .foregroundStyle(Color.Echo.outline)
                }
            }

            // Time remaining
            HStack(spacing: 4) {
                Image(systemName: "clock")
                    .font(.system(size: 12))
                Text(proposal.endsAt, style: .relative)
                Text("remaining")
            }
            .font(.custom("Inter", size: 12))
            .foregroundStyle(Color.Echo.outline)

            // Status badge for non-active
            if proposal.status != .active {
                StatusBadge(status: proposal.status)
            }
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 24)
                .fill(Color.Echo.surfaceContainerLowest)
        )
        .ghostBorder(opacity: 0.12)
        .glacialShadow(radius: 16, opacity: 0.03)
    }
}

// MARK: - Tally Progress Bar

struct TallyProgressBar: View {
    let tally: ProposalTally

    var body: some View {
        GeometryReader { geo in
            HStack(spacing: 2) {
                let total = Double(truncating: tally.totalWeight as NSDecimalNumber)
                let forFrac = total > 0 ? Double(truncating: tally.forWeight as NSDecimalNumber) / total : 0
                let againstFrac = total > 0 ? Double(truncating: tally.againstWeight as NSDecimalNumber) / total : 0
                let abstainFrac = total > 0 ? Double(truncating: tally.abstainWeight as NSDecimalNumber) / total : 0

                if forFrac > 0 {
                    Capsule()
                        .fill(Color.Echo.primaryContainer)
                        .frame(width: geo.size.width * forFrac)
                }
                if againstFrac > 0 {
                    Capsule()
                        .fill(Color.Echo.error.opacity(0.6))
                        .frame(width: geo.size.width * againstFrac)
                }
                if abstainFrac > 0 {
                    Capsule()
                        .fill(Color.Echo.outline.opacity(0.3))
                        .frame(width: geo.size.width * abstainFrac)
                }
            }
        }
        .frame(height: 6)
        .clipShape(Capsule())
    }
}

// MARK: - Status Badge

struct StatusBadge: View {
    let status: ProposalStatus

    private var color: Color {
        switch status {
        case .active: return Color.Echo.primaryContainer
        case .passed: return .green
        case .failed: return Color.Echo.error
        case .executed: return Color.Echo.secondary
        }
    }

    var body: some View {
        Text(status.displayName)
            .font(.custom("Inter", size: 11))
            .fontWeight(.semibold)
            .foregroundStyle(color)
            .padding(.horizontal, 10)
            .padding(.vertical, 4)
            .background(
                Capsule().fill(color.opacity(0.12))
            )
    }
}
