// Features/Payments/PaymentRequestSheet.swift
//
// Compose an in-chat payment: enter an amount + optional memo, then Request or Send. Validation
// reuses PaymentAmount; the actual send/transfer is handled by the caller (PaymentCoordinator).

#if os(iOS)
import SwiftUI

public struct PaymentRequestSheet: View {
    let onSubmit: (_ kind: PaymentPayload.Kind, _ amount: Decimal, _ memo: String?) -> Void
    let onCancel: () -> Void

    @State private var amountText = ""
    @State private var memo = ""

    public init(onSubmit: @escaping (PaymentPayload.Kind, Decimal, String?) -> Void,
                onCancel: @escaping () -> Void = {}) {
        self.onSubmit = onSubmit
        self.onCancel = onCancel
    }

    private var amount: Decimal? {
        guard let value = PaymentAmount.parse(amountText), PaymentAmount.isValid(value) else { return nil }
        return value
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                VStack(spacing: 6) {
                    Text("Amount (ECHO)").font(.system(size: 12, weight: .semibold)).foregroundColor(.echoInk55)
                    TextField("0.00", text: $amountText)
                        .keyboardType(.decimalPad)
                        .multilineTextAlignment(.center)
                        .font(.system(size: 34, weight: .bold, design: .rounded))
                        .foregroundColor(.echoInk)
                }
                .padding(.top, 12)

                TextField("What's it for? (optional)", text: $memo)
                    .font(.system(size: 15))
                    .padding(12)
                    .background(Color.echoPaperDim)
                    .clipShape(RoundedRectangle(cornerRadius: 10))

                VStack(spacing: 10) {
                    Button { submit(.sent) } label: { primaryButton("Send", filled: true) }
                        .disabled(amount == nil)
                    Button { submit(.request) } label: { primaryButton("Request", filled: false) }
                        .disabled(amount == nil)
                }
                Spacer()
            }
            .padding(20)
            .background(Color.echoPaper)
            .navigationTitle("Payment")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) { Button("Cancel", action: onCancel) }
            }
        }
    }

    private func submit(_ kind: PaymentPayload.Kind) {
        guard let amount else { return }
        onSubmit(kind, amount, memo.isEmpty ? nil : memo)
    }

    @ViewBuilder private func primaryButton(_ title: String, filled: Bool) -> some View {
        Text(title)
            .font(.system(size: 16, weight: .semibold))
            .foregroundColor(filled ? .white : .echoSignal)
            .frame(maxWidth: .infinity).frame(height: 50)
            .background(filled ? Color.echoSignal : Color.clear)
            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.echoSignal.opacity(filled ? 0 : 0.6), lineWidth: 1))
            .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
#endif
