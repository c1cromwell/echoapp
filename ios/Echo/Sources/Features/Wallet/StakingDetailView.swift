// Features/Wallet/StakingDetailView.swift
// Detailed staking position view with tier table, delegation info, and actions

import SwiftUI

struct StakingDetailView: View {
    @StateObject private var viewModel = StakingDetailViewModel()

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Current Position Card
                VStack(alignment: .leading, spacing: 16) {
                    Text("CURRENT POSITION")
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold)
                        .tracking(2)
                        .foregroundStyle(Color.Echo.outline)

                    VStack(alignment: .leading, spacing: 12) {
                        HStack {
                            Text("Staked Amount")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.outline)
                            Spacer()
                            Text("\(viewModel.stakedAmount.formatted()) ECHO")
                                .font(Font.Echo.bodyLarge)
                                .fontWeight(.bold)
                                .foregroundStyle(Color.Echo.onSurface)
                        }

                        HStack {
                            Text("Current Tier")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.outline)
                            Spacer()
                            Text(viewModel.currentTier.displayName)
                                .font(Font.Echo.bodyMedium)
                                .fontWeight(.bold)
                                .foregroundStyle(Color.Echo.primaryContainer)
                        }

                        HStack {
                            Text("APY")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.outline)
                            Spacer()
                            Text("\(String(format: "%.1f", viewModel.currentTier.apy))%")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.success)
                        }

                        HStack {
                            Text("Lock Period")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.outline)
                            Spacer()
                            Text(viewModel.lockPeriod)
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.onSurface)
                        }

                        HStack {
                            Text("Multiplier")
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.outline)
                            Spacer()
                            Text("\(String(format: "%.1f", viewModel.currentTier.multiplier))x")
                                .font(Font.Echo.bodyMedium)
                                .fontWeight(.bold)
                                .foregroundStyle(Color.Echo.primaryContainer)
                        }
                    }
                    .padding(20)
                    .background(
                        RoundedRectangle(cornerRadius: 32)
                            .fill(Color.Echo.surfaceContainerLow)
                    )
                    .ghostBorder(opacity: 0.15)
                }
                .padding(.horizontal, 20)

                // Staking Tiers Reference
                VStack(alignment: .leading, spacing: 16) {
                    Text("STAKING TIERS")
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold)
                        .tracking(2)
                        .foregroundStyle(Color.Echo.outline)
                        .padding(.leading, 8)

                    VStack(spacing: 0) {
                        // Header
                        HStack {
                            Text("Tier").font(Font.Echo.labelMd).fontWeight(.bold).frame(maxWidth: .infinity, alignment: .leading)
                            Text("Min Stake").font(Font.Echo.labelMd).fontWeight(.bold).frame(maxWidth: .infinity)
                            Text("APY").font(Font.Echo.labelMd).fontWeight(.bold).frame(maxWidth: .infinity)
                            Text("Multiplier").font(Font.Echo.labelMd).fontWeight(.bold).frame(maxWidth: .infinity, alignment: .trailing)
                        }
                        .foregroundStyle(Color.Echo.outline)
                        .padding(12)
                        .background(Color.Echo.surfaceContainerHigh)

                        ForEach(StakingTierInfo.allTiers) { tier in
                            HStack {
                                Text(tier.name)
                                    .font(Font.Echo.bodyMedium)
                                    .fontWeight(tier.name == viewModel.currentTier.displayName ? .bold : .regular)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                Text("\(tier.minimumStake)")
                                    .font(Font.Echo.bodyMedium)
                                    .frame(maxWidth: .infinity)
                                Text("\(String(format: "%.1f", tier.apy))%")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.success)
                                    .frame(maxWidth: .infinity)
                                Text("\(String(format: "%.1f", tier.multiplier))x")
                                    .font(Font.Echo.bodyMedium)
                                    .frame(maxWidth: .infinity, alignment: .trailing)
                            }
                            .foregroundStyle(Color.Echo.onSurface)
                            .padding(12)
                            .background(
                                tier.name == viewModel.currentTier.displayName
                                    ? Color.Echo.primaryContainer.opacity(0.1)
                                    : Color.clear
                            )
                        }
                    }
                    .clipShape(RoundedRectangle(cornerRadius: 20))
                    .background(
                        RoundedRectangle(cornerRadius: 20)
                            .fill(Color.Echo.surfaceContainerLow)
                    )
                    .ghostBorder(opacity: 0.15)
                }
                .padding(.horizontal, 20)

                // Delegation Card
                if let validator = viewModel.delegatedValidator {
                    VStack(alignment: .leading, spacing: 16) {
                        Text("DELEGATION")
                            .font(.custom("Inter", size: 10))
                            .fontWeight(.bold)
                            .tracking(2)
                            .foregroundStyle(Color.Echo.outline)
                            .padding(.leading, 8)

                        VStack(alignment: .leading, spacing: 12) {
                            HStack {
                                Text("Validator")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.outline)
                                Spacer()
                                Text(validator.name)
                                    .font(Font.Echo.bodyMedium)
                                    .fontWeight(.bold)
                                    .foregroundStyle(Color.Echo.onSurface)
                            }
                            HStack {
                                Text("Uptime")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.outline)
                                Spacer()
                                Text("\(String(format: "%.1f", validator.uptime))%")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.success)
                            }
                            HStack {
                                Text("Commission")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.outline)
                                Spacer()
                                Text("\(String(format: "%.1f", validator.commission))%")
                                    .font(Font.Echo.bodyMedium)
                                    .foregroundStyle(Color.Echo.onSurface)
                            }
                        }
                        .padding(20)
                        .background(
                            RoundedRectangle(cornerRadius: 32)
                                .fill(Color.Echo.surfaceContainerLow)
                        )
                        .ghostBorder(opacity: 0.15)
                    }
                    .padding(.horizontal, 20)
                }

                // Action Buttons
                VStack(spacing: 12) {
                    Button {
                        // Navigate to stake more
                    } label: {
                        Text("Stake More")
                            .font(.custom("Inter", size: 14)).fontWeight(.bold)
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 16)
                            .background(
                                RoundedRectangle(cornerRadius: 9999)
                                    .fill(LinearGradient.signature)
                            )
                    }
                    .deepGlacialShadow()

                    Button {
                        // Navigate to unstake
                    } label: {
                        Text("Unstake")
                            .font(.custom("Inter", size: 14)).fontWeight(.bold)
                            .foregroundStyle(Color.Echo.onSurface)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 16)
                            .background(
                                RoundedRectangle(cornerRadius: 9999)
                                    .fill(Color.Echo.surfaceContainerLow)
                            )
                            .ghostBorder(opacity: 0.15)
                    }
                }
                .padding(.horizontal, 20)
            }
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Staking Details")
        .task { await viewModel.loadStakingDetails() }
    }
}

// MARK: - Staking Tier Info (for reference table)

struct StakingTierInfo: Identifiable {
    let id = UUID()
    let name: String
    let minimumStake: Int
    let apy: Double
    let multiplier: Double

    static let allTiers: [StakingTierInfo] = [
        StakingTierInfo(name: "Bronze", minimumStake: 100, apy: 8.0, multiplier: 1.0),
        StakingTierInfo(name: "Silver", minimumStake: 1000, apy: 10.0, multiplier: 1.2),
        StakingTierInfo(name: "Gold", minimumStake: 5000, apy: 12.5, multiplier: 1.8),
        StakingTierInfo(name: "Platinum", minimumStake: 25000, apy: 15.0, multiplier: 2.5),
    ]
}

// MARK: - Staking Tier Enum for Detail View

enum StakingTierLevel: String, CaseIterable {
    case none, bronze, silver, gold, platinum

    var displayName: String { rawValue.capitalized }

    var apy: Double {
        switch self {
        case .none: return 0
        case .bronze: return 8.0
        case .silver: return 10.0
        case .gold: return 12.5
        case .platinum: return 15.0
        }
    }

    var multiplier: Double {
        switch self {
        case .none: return 0
        case .bronze: return 1.0
        case .silver: return 1.2
        case .gold: return 1.8
        case .platinum: return 2.5
        }
    }
}

// MARK: - Delegated Validator

struct DelegatedValidatorInfo {
    let name: String
    let uptime: Double
    let commission: Double
}

// MARK: - Staking Detail ViewModel

@MainActor
class StakingDetailViewModel: ObservableObject {
    @Published var stakedAmount: Decimal = 0
    @Published var currentTier: StakingTierLevel = .none
    @Published var lockPeriod: String = "—"
    @Published var delegatedValidator: DelegatedValidatorInfo?

    func loadStakingDetails() async {
        // TODO: Load from staking service
    }
}
