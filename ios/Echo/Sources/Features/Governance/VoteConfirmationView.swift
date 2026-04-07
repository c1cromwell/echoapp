// Features/Governance/VoteConfirmationView.swift
// Review weight and submit vote — on-chain and irreversible

import SwiftUI

struct VoteConfirmationView: View {
    let proposal: Proposal
    let voteValue: VoteValue
    let votingPower: VotingPower
    let onConfirm: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var isSubmitting = false

    private var voteColor: Color {
        switch voteValue {
        case .for: return Color.Echo.primaryContainer
        case .against: return Color.Echo.error
        case .abstain: return Color.Echo.outline
        }
    }

    var body: some View {
        VStack(spacing: 24) {
            // Vote icon
            Image(systemName: voteValue.systemImage)
                .font(.system(size: 48))
                .foregroundStyle(voteColor)
                .padding(.top, 32)

            // Title
            Text("Confirm Your Vote")
                .font(.custom("Inter", size: 24))
                .fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)

            // Summary card
            VStack(alignment: .leading, spacing: 12) {
                SummaryRow(label: "Proposal", value: proposal.title)
                SummaryRow(label: "Your Vote", value: voteValue.displayName, valueColor: voteColor)
                SummaryRow(label: "Your Weight", value: formatWeight(votingPower.weight))
                SummaryRow(label: "Trust Tier", value: "Tier \(votingPower.trustTier) (\(String(format: "%.1f", votingPower.multiplier))×)")
            }
            .padding(20)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .padding(.horizontal, 20)

            // Warning
            Text("This vote will be recorded on-chain and cannot be changed.")
                .font(.custom("Inter", size: 13))
                .foregroundStyle(Color.Echo.outline)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)

            Spacer()

            // Submit button
            Button(action: {
                isSubmitting = true
                onConfirm()
            }) {
                HStack {
                    if isSubmitting {
                        ProgressView()
                            .tint(.white)
                    }
                    Text("Submit Vote")
                        .font(.custom("Inter", size: 18))
                        .fontWeight(.bold)
                }
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 18)
                .background(
                    Capsule()
                        .fill(LinearGradient.signature)
                )
                .deepGlacialShadow()
            }
            .disabled(isSubmitting)
            .padding(.horizontal, 20)

            // Cancel
            Button("Cancel") {
                dismiss()
            }
            .font(.custom("Inter", size: 16))
            .foregroundStyle(Color.Echo.outline)
            .padding(.bottom, 32)
        }
    }

    private func formatWeight(_ value: Decimal) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.maximumFractionDigits = 0
        return formatter.string(from: value as NSDecimalNumber) ?? "0"
    }
}

// MARK: - Summary Row

private struct SummaryRow: View {
    let label: String
    let value: String
    var valueColor: Color = Color.Echo.onSurface

    var body: some View {
        HStack {
            Text(label)
                .font(.custom("Inter", size: 14))
                .foregroundStyle(Color.Echo.outline)
            Spacer()
            Text(value)
                .font(.custom("Inter", size: 14))
                .fontWeight(.semibold)
                .foregroundStyle(valueColor)
        }
    }
}
