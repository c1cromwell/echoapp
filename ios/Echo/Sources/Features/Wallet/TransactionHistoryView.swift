#if os(iOS)
import SwiftUI

struct TransactionHistoryView: View {
    @State private var selectedFilter: TransactionFilter = .all
    @State private var transactions: [WalletTransaction] = []

    enum TransactionFilter: String, CaseIterable, Identifiable {
        case all, stakes, claims, transfers
        var id: String { rawValue }
        var displayName: String { rawValue.capitalized }
    }

    var filteredTransactions: [WalletTransaction] {
        switch selectedFilter {
        case .all:       return transactions
        case .stakes:    return transactions.filter { $0.type == .stake || $0.type == .unstake }
        case .claims:    return transactions.filter { $0.type == .claim || $0.type == .reward }
        case .transfers: return transactions.filter { $0.type == .transfer }
        }
    }

    var body: some View {
        VStack(spacing: 0) {
            filterBar

            if filteredTransactions.isEmpty {
                EmptyStateView(
                    icon: "clock.arrow.circlepath",
                    title: "No transactions yet",
                    subtitle: "Your staking, claims, and transfer history will appear here."
                )
            } else {
                ScrollView {
                    LazyVStack(spacing: 0) {
                        ForEach(filteredTransactions) { tx in
                            TransactionRow(transaction: tx)
                            Divider()
                                .background(Color.echoHair)
                                .padding(.leading, 64)
                        }
                    }
                }
            }
        }
        .icyBackground()
        .navigationTitle("Transaction History")
        .navigationBarTitleDisplayMode(.inline)
    }

    private var filterBar: some View {
        HStack(spacing: 4) {
            ForEach(TransactionFilter.allCases) { filter in
                let isSelected = selectedFilter == filter
                Button {
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) {
                        selectedFilter = filter
                    }
                    HapticManager.selection()
                } label: {
                    Text(filter.displayName)
                        .font(.system(size: 13, weight: isSelected ? .semibold : .medium))
                        .foregroundColor(isSelected ? .echoPaper : .echoInk55)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 7)
                        .background(isSelected ? Color.echoInk : Color.clear)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(4)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }
}

// MARK: - Transaction Row

struct TransactionRow: View {
    let transaction: WalletTransaction

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: transaction.type.icon)
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(transaction.isPositive ? .echoTrustGreen : .echoInk55)
                .frame(width: 36, height: 36)
                .background(
                    (transaction.isPositive ? Color.echoTrustGreen : Color.echoInk)
                        .opacity(0.1)
                )
                .clipShape(Circle())

            VStack(alignment: .leading, spacing: 2) {
                Text(transaction.description)
                    .font(.system(size: 15, weight: .medium))
                    .foregroundColor(.echoInk)
                    .lineLimit(1)

                Text(transaction.date.formatted(date: .abbreviated, time: .shortened))
                    .font(.system(size: 12))
                    .foregroundColor(.echoInk40)
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 2) {
                Text("\(transaction.isPositive ? "+" : "-")\(formatEcho(transaction.amount))")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundColor(transaction.isPositive ? .echoTrustGreen : .echoInk)

                if transaction.status == .pending {
                    Text("Pending")
                        .font(.system(size: 10, weight: .medium))
                        .foregroundColor(.echoSignal)
                } else if transaction.status == .failed {
                    Text("Failed")
                        .font(.system(size: 10, weight: .medium))
                        .foregroundColor(.echoAlert)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }
}

#endif
