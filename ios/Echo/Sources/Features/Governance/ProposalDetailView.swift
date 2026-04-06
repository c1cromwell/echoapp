// Features/Governance/ProposalDetailView.swift
// Full proposal view with description, tally breakdown, and voting interface

import SwiftUI

struct ProposalDetailView: View {
    let proposal: Proposal
    let votingPower: VotingPower?
    let onVote: (VoteValue) -> Void

    @State private var selectedVote: VoteValue?
    @State private var showConfirmation = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                // Header
                VStack(alignment: .leading, spacing: 8) {
                    Text(proposal.type.displayName.uppercased())
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold)
                        .tracking(1.5)
                        .foregroundStyle(Color.Echo.primaryContainer)

                    Text(proposal.title)
                        .font(.custom("Inter", size: 24))
                        .fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)

                    HStack(spacing: 12) {
                        StatusBadge(status: proposal.status)

                        Text("Requires: \(proposal.threshold.displayName)")
                            .font(.custom("Inter", size: 12))
                            .foregroundStyle(Color.Echo.outline)
                    }
                }

                // Description
                Text(proposal.description)
                    .font(.custom("Inter", size: 15))
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                    .lineSpacing(4)

                // Tally breakdown
                if let tally = proposal.tally {
                    TallyBreakdown(tally: tally, threshold: proposal.threshold)
                }

                // Voting section
                if proposal.status == .active {
                    VotingSection(
                        canVote: votingPower?.canVote ?? false,
                        selectedVote: $selectedVote,
                        showConfirmation: $showConfirmation
                    )
                }

                // Metadata
                MetadataSection(proposal: proposal)
            }
            .padding(20)
        }
        .navigationTitle("Proposal")
        .navigationBarTitleDisplayMode(.inline)
        .sheet(isPresented: $showConfirmation) {
            if let vote = selectedVote, let power = votingPower {
                VoteConfirmationView(
                    proposal: proposal,
                    voteValue: vote,
                    votingPower: power
                ) {
                    onVote(vote)
                    showConfirmation = false
                }
            }
        }
    }
}

// MARK: - Tally Breakdown

private struct TallyBreakdown: View {
    let tally: ProposalTally
    let threshold: ThresholdType

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("VOTE TALLY")
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(1.5)
                .foregroundStyle(Color.Echo.outline)

            TallyProgressBar(tally: tally)

            HStack(spacing: 24) {
                TallyItem(label: "For", weight: tally.forWeight, percent: tally.forPercent, color: Color.Echo.primaryContainer)
                TallyItem(label: "Against", weight: tally.againstWeight, percent: tally.againstPercent, color: Color.Echo.error)
                TallyItem(label: "Abstain", weight: tally.abstainWeight, percent: tally.abstainPercent, color: Color.Echo.outline)
            }

            // Threshold indicator
            HStack {
                Text("Threshold: \(threshold.requiredPercent)%")
                Spacer()
                Text("\(tally.voterCount) voter\(tally.voterCount == 1 ? "" : "s")")
            }
            .font(.custom("Inter", size: 12))
            .foregroundStyle(Color.Echo.outline)
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 20)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .ghostBorder(opacity: 0.12)
    }
}

private struct TallyItem: View {
    let label: String
    let weight: Decimal
    let percent: Double
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)

            Text(label)
                .font(.custom("Inter", size: 12))
                .foregroundStyle(Color.Echo.outline)

            Text("\(Int(percent))%")
                .font(.custom("Inter", size: 18))
                .fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)
                .monospacedDigit()
        }
    }
}

// MARK: - Voting Section

private struct VotingSection: View {
    let canVote: Bool
    @Binding var selectedVote: VoteValue?
    @Binding var showConfirmation: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("CAST YOUR VOTE")
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(1.5)
                .foregroundStyle(Color.Echo.outline)

            if canVote {
                HStack(spacing: 12) {
                    ForEach(VoteValue.allCases, id: \.self) { value in
                        VoteButton(value: value, isSelected: selectedVote == value) {
                            selectedVote = value
                            showConfirmation = true
                        }
                    }
                }
            } else {
                Text("You must be Trust Tier 2+ with staked ECHO to vote")
                    .font(.custom("Inter", size: 14))
                    .foregroundStyle(Color.Echo.outline)
                    .padding(16)
                    .frame(maxWidth: .infinity)
                    .background(
                        RoundedRectangle(cornerRadius: 16)
                            .fill(Color.Echo.surfaceContainerLow)
                    )
            }
        }
    }
}

private struct VoteButton: View {
    let value: VoteValue
    let isSelected: Bool
    let action: () -> Void

    private var color: Color {
        switch value {
        case .for: return Color.Echo.primaryContainer
        case .against: return Color.Echo.error
        case .abstain: return Color.Echo.outline
        }
    }

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: value.systemImage)
                    .font(.system(size: 24))
                Text(value.displayName)
                    .font(.custom("Inter", size: 12))
                    .fontWeight(.semibold)
            }
            .foregroundStyle(isSelected ? .white : color)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(isSelected ? color : color.opacity(0.08))
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Metadata

private struct MetadataSection: View {
    let proposal: Proposal

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("DETAILS")
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(1.5)
                .foregroundStyle(Color.Echo.outline)

            LabeledContent("Created by") {
                Text(String(proposal.createdBy.prefix(20)) + "...")
                    .font(.custom("Inter", size: 13).monospacedDigit())
            }

            LabeledContent("Created") {
                Text(proposal.createdAt, style: .date)
            }

            LabeledContent("Ends") {
                Text(proposal.endsAt, style: .date)
            }
        }
        .font(.custom("Inter", size: 13))
        .foregroundStyle(Color.Echo.onSurfaceVariant)
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 20)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .ghostBorder(opacity: 0.12)
    }
}
