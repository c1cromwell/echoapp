// Features/Payments/PaymentMessageEncoding.swift
//
// Carries a PaymentPayload inside a normal text message (no message-model/wire/persistence changes):
// the payload rides in the message content behind an invisible sentinel, and the chat renders it as
// a payment bubble instead of text. Flows over the existing E2E text path unchanged.

import Foundation

public enum PaymentMessageEncoding {
    // U+2063 (invisible separator) keeps the marker from ever showing as readable text.
    private static let marker = "\u{2063}echo-pay:"

    public static func encode(_ payload: PaymentPayload) -> String {
        marker + (String(data: payload.encoded(), encoding: .utf8) ?? "")
    }

    public static func decode(_ content: String) -> PaymentPayload? {
        guard content.hasPrefix(marker) else { return nil }
        return PaymentPayload.decode(Data(content.dropFirst(marker.count).utf8))
    }

    public static func isPayment(_ content: String) -> Bool { content.hasPrefix(marker) }
}
