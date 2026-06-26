// Features/Payments/WalletTransfer.swift
//
// Injectable P2P transfer for in-chat payments. The Stargazer wallet has no transfer op today
// (only staking/locking/rewards), so the real send is deferred: a production impl needs a backend
// transfer endpoint + DID→wallet-address resolution + on-chain signing (separate, security-reviewed
// effort). `MockWalletTransfer` lets the full UX + flow ship and be tested without moving funds.

import Foundation

public enum WalletTransferError: Error, Equatable {
    case invalidAmount
    case notImplemented
}

public protocol WalletTransfer: Sendable {
    /// Transfer `amount` (ECHO token) to the recipient DID. Returns a transaction id.
    func transfer(amount: Decimal, toDID: String, memo: String?) async throws -> String
}

/// No funds move — returns a synthetic tx id. Swap for a real transfer when the backend lands.
public struct MockWalletTransfer: WalletTransfer {
    public init() {}
    public func transfer(amount: Decimal, toDID: String, memo: String?) async throws -> String {
        guard PaymentAmount.isValid(amount) else { throw WalletTransferError.invalidAmount }
        return "mocktx-" + UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(12)
    }
}
