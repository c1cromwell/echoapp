// Features/Governance/GovernanceWeightView.swift
// Displays user's trust-tier weighted governance power

import SwiftUI

struct GovernanceWeightView: View {
    let power: VotingPower

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            Text("YOUR VOTING POWER")
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(1.5)
                .foregroundStyle(Color.Echo.outline)

            // Main power display
            HStack(alignment: .firstTextBaseline, spacing: 8) {
                Text(formatWeight(power.weight))
                    .font(.custom("Inter", size: 32))
                    .fontWeight(.bold)
                    .foregroundStyle(Color.Echo.onSurface)
                    .monospacedDigit()

                Text("effective ECHO")
                    .font(.custom("Inter", size: 14))
                    .foregroundStyle(Color.Echo.outline)
            }

            // Breakdown
            VStack(alignment: .leading, spacing: 8) {
                LabeledContent("Staked ECHO") {
                    Text(formatWeight(power.totalStaked))
                        .font(.custom("Inter", size: 14).monospacedDigit())
                }

                LabeledContent("Trust Tier") {
                    HStack(spacing: 4) {
                        Text("Tier \(power.trustTier)")
                        Text("(\(power.multiplier, specifier: "%.1f")×)")
                            .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .font(.custom("Inter", size: 14))
                }

                LabeledContent("Formula") {
                    Text("\(formatWeight(power.totalStaked)) × \(power.multiplier, specifier: "%.1f")")
                        .font(.custom("Inter", size: 12).monospacedDigit())
                        .foregroundStyle(Color.Echo.outline)
                }
            }
            .font(.custom("Inter", size: 14))
            .foregroundStyle(Color.Echo.onSurfaceVariant)

            // Warning if cannot vote
            if !power.canVote {
                HStack(spacing: 8) {
                    Image(systemName: "exclamationmark.triangle")
                    Text(power.trustTier < 2
                         ? "Verify your identity (Tier 2+) to participate in governance"
                         : "Stake ECHO to participate in governance")
                }
                .font(.custom("Inter", size: 13))
                .foregroundStyle(.orange)
                .padding(12)
                .background(
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color.orange.opacity(0.08))
                )
            }
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 24)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .ghostBorder(opacity: 0.15)
    }

    private func formatWeight(_ value: Decimal) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.maximumFractionDigits = 0
        return formatter.string(from: value as NSDecimalNumber) ?? "0"
    }
}
