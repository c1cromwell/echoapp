// Features/Payments/PaymentCoordinator.swift
//
// Pure coordination for in-chat payments: build a request, run the trust pay-guard, and execute a
// transfer producing a `.sent` payment. UI-free so it's fully unit-testable.

import Foundation

public struct PaymentCoordinator {
    private let transfer: WalletTransfer
    private let evaluator: ContactSafetyEvaluator

    public init(transfer: WalletTransfer, evaluator: ContactSafetyEvaluator = ContactSafetyEvaluator()) {
        self.transfer = transfer
        self.evaluator = evaluator
    }

    /// Compose a payment request (amount + memo). Throws on an invalid amount.
    public func makeRequest(amount: Decimal, memo: String?) throws -> PaymentPayload {
        guard PaymentAmount.isValid(amount) else { throw WalletTransferError.invalidAmount }
        return PaymentPayload(kind: .request, amount: amount, memo: memo, status: .pending)
    }

    /// Pay-guard: paying an unverified/low-trust recipient needs an extra confirmation step.
    public func needsExtraConfirmation(peerTier: Int) -> Bool {
        evaluator.requiresExtraConfirmationToPay(peerTier: peerTier)
    }

    /// Execute payment for a request, returning the `.sent` payment to broadcast back. Callers must
    /// have already passed the pay-guard (extra confirm) when `needsExtraConfirmation` is true.
    public func pay(_ request: PaymentPayload, toDID: String) async throws -> PaymentPayload {
        guard PaymentAmount.isValid(request.amount) else { throw WalletTransferError.invalidAmount }
        let txId = try await transfer.transfer(amount: request.amount, toDID: toDID, memo: request.memo)
        return PaymentPayload(id: request.id, kind: .sent, amount: request.amount,
                              memo: request.memo, txId: txId, status: .paid)
    }
}
