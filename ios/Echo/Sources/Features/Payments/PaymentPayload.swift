// Features/Payments/PaymentPayload.swift
//
// The structured payload for an in-chat payment, sent over the normal E2E message path and rendered
// as a payment bubble (mirrors how media messages carry a structured payload).

import Foundation

public struct PaymentPayload: Codable, Equatable, Sendable {
    public enum Kind: String, Codable, Sendable { case request, sent }
    public enum Status: String, Codable, Sendable { case pending, paid, declined, failed }

    public let id: String
    public var kind: Kind
    public let amount: Decimal
    public let memo: String?
    public var txId: String?
    public var status: Status

    public init(id: String = UUID().uuidString, kind: Kind, amount: Decimal,
                memo: String? = nil, txId: String? = nil, status: Status = .pending) {
        self.id = id
        self.kind = kind
        self.amount = amount
        self.memo = memo
        self.txId = txId
        self.status = status
    }

    public func encoded() -> Data { (try? JSONEncoder().encode(self)) ?? Data() }
    public static func decode(_ data: Data) -> PaymentPayload? {
        try? JSONDecoder().decode(PaymentPayload.self, from: data)
    }
}

public enum PaymentAmount {
    public static let maxAmount: Decimal = 1_000_000

    public static func isValid(_ amount: Decimal, max: Decimal = maxAmount) -> Bool {
        amount > 0 && amount <= max
    }

    /// Parse user input ("12.50") into a non-negative amount, or nil if malformed.
    public static func parse(_ text: String) -> Decimal? {
        let trimmed = text.trimmingCharacters(in: .whitespaces)
        guard let value = Decimal(string: trimmed), value >= 0 else { return nil }
        return value
    }
}
