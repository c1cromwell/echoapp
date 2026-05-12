// Features/Governance/GovernanceWeightView.swift
// Displays user's trust-tier weighted governance power

import SwiftUI

public struct GovernanceWeightView: View {
    public let power: VotingPower

    public init(power: VotingPower) {
        self.power = power
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            Text("YOUR VOTING POWER")
                .font(.system(size: 10))
                .fontWeight(.bold)
                .tracking(1.5)
                .foregroundStyle(Color.Echo.outline)

            // Main power display
            HStack(alignment: .firstTextBaseline, spacing: 8) {
                Text(formatWeight(power.weight))
                    .font(.system(size: 32))
                    .fontWeight(.bold)
                    .foregroundStyle(Color.Echo.onSurface)
                    .monospacedDigit()

                Text("effective ECHO")
                    .font(.system(size: 14))
                    .foregroundStyle(Color.Echo.outline)
            }

            // Breakdown
            VStack(alignment: .leading, spacing: 8) {
                LabeledContent("Staked ECHO") {
                    Text(formatWeight(power.totalStaked))
                        .font(.system(size: 14).monospacedDigit())
                }

                LabeledContent("Trust Tier") {
                    HStack(spacing: 4) {
                        Text("Tier \(power.trustTier)")
                        Text("(\(power.multiplier, specifier: "%.1f")×)")
                            .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .font(.system(size: 14))
                }

                LabeledContent("Formula") {
                    Text("\(formatWeight(power.totalStaked)) × \(power.multiplier, specifier: "%.1f")")
                        .font(.system(size: 12).monospacedDigit())
                        .foregroundStyle(Color.Echo.outline)
                }
            }
            .font(.system(size: 14))
            .foregroundStyle(Color.Echo.onSurfaceVariant)

            // Warning if cannot vote
            if !power.canVote {
                HStack(spacing: 8) {
                    Image(systemName: "exclamationmark.triangle")
                    Text(power.trustTier < 2
                         ? "Verify your identity (Tier 2+) to participate in governance"
                         : "Stake ECHO to participate in governance")
                }
                .font(.system(size: 13))
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
