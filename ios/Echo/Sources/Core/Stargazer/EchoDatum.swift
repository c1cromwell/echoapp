// Core/Stargazer/EchoDatum.swift
// On-chain / API integer scale: 1 ECHO = 1e8 datum (matches Go wallet.DatumPerECHO).

import Foundation

enum EchoDatum {
    static let scale: Decimal = 100_000_000

    static func fromDatum(_ datum: Int64) -> Decimal {
        Decimal(datum) / scale
    }

    static func toDatum(_ echo: Decimal) -> Int64 {
        let scaled = (echo * scale as NSDecimalNumber)
        return scaled.int64Value
    }
}
