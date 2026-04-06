// Features/Wallet/Delegation/ValidatorBrowserView.swift
// List validators with metrics for delegation

import SwiftUI

struct ValidatorBrowserView: View {
    @ObservedObject var viewModel: WalletViewModel

    var body: some View {
        ScrollView {
            VStack(spacing: 12) {
                if viewModel.validators.isEmpty {
                    VStack(spacing: 16) {
                        Image(systemName: "person.3")
                            .font(.system(size: 48))
                            .foregroundStyle(Color.Echo.onSurfaceVariant.opacity(0.4))
                        Text("Loading validators...")
                            .font(Font.Echo.bodyLarge)
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.top, 80)
                } else {
                    ForEach(viewModel.validators) { validator in
                        ValidatorRow(
                            validator: validator,
                            isSelected: viewModel.selectedValidator?.id == validator.id
                        ) {
                            viewModel.selectedValidator = validator
                        }
                    }
                }
            }
            .padding()
        }
        .icyBackground()
        .navigationTitle("Validators")
        .task {
            await viewModel.loadValidators()
        }
    }
}

// MARK: - Validator Row

struct ValidatorRow: View {
    let validator: ValidatorInfo
    let isSelected: Bool
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            GhostBorderCard {
                VStack(alignment: .leading, spacing: 8) {
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(validator.id.prefix(12) + "...")
                                .font(Font.Echo.bodyLarge)
                                .foregroundStyle(Color.Echo.onSurface)
                            Text(validator.layer.replacingOccurrences(of: "_", with: " ").capitalized)
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.onSurfaceVariant)
                        }
                        Spacer()
                        if isSelected {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(Color.Echo.primaryContainer)
                        }
                    }

                    HStack(spacing: 16) {
                        metricView("Uptime", value: String(format: "%.1f%%", validator.uptimePercent))
                        metricView("Commission", value: String(format: "%.1f%%", validator.commissionPercent))
                        metricView("APR", value: String(format: "%.1f%%", validator.estimatedAPR))
                        metricView("Delegators", value: "\(validator.delegatorCount)")
                    }
                }
            }
        }
        .buttonStyle(.plain)
    }

    private func metricView(_ label: String, value: String) -> some View {
        VStack(spacing: 2) {
            Text(value)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
            Text(label)
                .font(Font.Echo.labelSm)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }
}
