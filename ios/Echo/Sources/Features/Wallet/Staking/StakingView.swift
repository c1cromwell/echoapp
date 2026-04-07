// Features/Wallet/Staking/StakingView.swift
// Amount picker + tier selector for TokenLock staking

import SwiftUI

struct WalletStakingView: View {
    @ObservedObject var viewModel: WalletViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Amount Input
                GhostBorderCard {
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Stake Amount")
                            .font(Font.Echo.titleLarge)
                            .foregroundStyle(Color.Echo.onSurface)

                        HStack {
                            TextField("0.00", text: $viewModel.stakeAmount)
                                .font(Font.Echo.displayMedium)
                                .foregroundStyle(Color.Echo.onSurface)
                                #if os(iOS)
                                .keyboardType(.decimalPad)
                                #endif

                            Text("ECHO")
                                .font(Font.Echo.bodyLarge)
                                .foregroundStyle(Color.Echo.onSurfaceVariant)
                        }

                        if let state = viewModel.walletState {
                            Text("Available: \(formatEcho(state.available)) ECHO")
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.onSurfaceVariant)
                        }
                    }
                }

                // Tier Selection
                VStack(alignment: .leading, spacing: 12) {
                    Text("Select Tier")
                        .font(Font.Echo.titleLarge)
                        .foregroundStyle(Color.Echo.onSurface)

                    ForEach(StakingTier.allCases) { tier in
                        Button {
                            viewModel.selectedTier = tier
                        } label: {
                            HStack {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text(tier.displayName)
                                        .font(Font.Echo.bodyLarge)
                                        .foregroundStyle(Color.Echo.onSurface)
                                    Text(tier.durationLabel)
                                        .font(Font.Echo.labelMd)
                                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                                }

                                Spacer()

                                Text("\(String(format: "%.0f", tier.apr))% APR")
                                    .font(Font.Echo.bodyLarge)
                                    .foregroundStyle(Color.Echo.primaryContainer)

                                if viewModel.selectedTier == tier {
                                    Image(systemName: "checkmark.circle.fill")
                                        .foregroundStyle(Color.Echo.primaryContainer)
                                }
                            }
                            .padding(16)
                            .background(
                                viewModel.selectedTier == tier
                                    ? Color.Echo.surfaceContainerHigh
                                    : Color.Echo.surfaceContainer
                            )
                            .clipShape(RoundedRectangle(cornerRadius: 20))
                            .ghostBorder(opacity: viewModel.selectedTier == tier ? 0.25 : 0.10)
                        }
                    }
                }

                // Stake Button
                SignatureGradientButton(
                    title: viewModel.isStaking ? "Staking..." : "Stake ECHO",
                    subtitle: "\(viewModel.selectedTier.displayName) • \(viewModel.selectedTier.durationLabel)",
                    icon: "lock.shield"
                ) {
                    Task {
                        await viewModel.stakeEcho()
                        if viewModel.errorMessage == nil {
                            dismiss()
                        }
                    }
                }
                .disabled(viewModel.isStaking || viewModel.stakeAmount.isEmpty)
                .opacity(viewModel.isStaking ? 0.6 : 1.0)
            }
            .padding()
        }
        .icyBackground()
        .navigationTitle("Stake ECHO")
    }
}
