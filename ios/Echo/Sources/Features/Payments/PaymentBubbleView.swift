// Features/Payments/PaymentBubbleView.swift
//
// Renders a payment message in the chat thread: a request shows a Pay button (incoming), a sent
// payment shows the amount + status. Decorative — the Pay action is handled by the chat view model
// (which runs the pay-guard before transferring).

#if os(iOS)
import SwiftUI

public struct PaymentBubbleView: View {
    let payment: PaymentPayload
    let isIncoming: Bool
    let onPay: () -> Void
    private let payActionsEnabled: Bool

    public init(
        payment: PaymentPayload,
        isIncoming: Bool,
        payActionsEnabled: Bool = true,
        onPay: @escaping () -> Void = {}
    ) {
        self.payment = payment
        self.isIncoming = isIncoming
        self.payActionsEnabled = payActionsEnabled
        self.onPay = onPay
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: payment.kind == .sent ? "checkmark.seal.fill" : "arrow.down.left.circle.fill")
                    .foregroundColor(.echoSignal)
                Text(payment.kind == .sent ? "Payment" : "Payment request")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.echoInk55)
            }
            Text("\(formatted(payment.amount)) ECHO")
                .font(.system(size: 22, weight: .bold, design: .rounded))
                .foregroundColor(.echoInk)
            if let memo = payment.memo, !memo.isEmpty {
                Text(memo).font(.system(size: 13)).foregroundColor(.echoInk70)
            }
            Text(statusText).font(.system(size: 11, weight: .medium)).foregroundColor(.echoInk40)

            if showPayButton {
                Button(action: onPay) {
                    Text("Pay \(formatted(payment.amount)) ECHO")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity).frame(height: 40)
                        .background(Color.echoSignal)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                }
            }
        }
        .padding(14)
        .frame(maxWidth: 260, alignment: .leading)
        .background(Color.echoPaperDim)
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    /// Only an incoming, still-pending request can be paid.
    private var showPayButton: Bool {
        payActionsEnabled && isIncoming && payment.kind == .request && payment.status == .pending
    }

    private var statusText: String {
        switch payment.status {
        case .pending: return payment.kind == .request ? "Awaiting payment" : "Pending"
        case .paid:    return "Paid"
        case .declined: return "Declined"
        case .failed:  return "Failed"
        }
    }

    private func formatted(_ amount: Decimal) -> String {
        let n = NSDecimalNumber(decimal: amount)
        let f = NumberFormatter()
        f.minimumFractionDigits = 0
        f.maximumFractionDigits = 4
        return f.string(from: n) ?? "\(amount)"
    }
}
#endif
