#if os(iOS)
import Foundation

/// Phone normalization aligned with `internal/services/contacts/oprf.go`.
enum PhoneNormalizer {
    static func normalize(_ raw: String) -> String {
        raw.replacingOccurrences(of: " ", with: "")
            .replacingOccurrences(of: "-", with: "")
            .replacingOccurrences(of: "(", with: "")
            .replacingOccurrences(of: ")", with: "")
    }

    /// Best-effort E.164 from CNContact phone digits.
    static func e164(from raw: String, defaultCountryCode: String = "+1") -> String {
        let digits = raw.filter(\.isNumber)
        if raw.hasPrefix("+") {
            return "+" + digits
        }
        if digits.count == 10, defaultCountryCode == "+1" {
            return "+1" + digits
        }
        if digits.hasPrefix("1"), digits.count == 11 {
            return "+" + digits
        }
        return defaultCountryCode + digits
    }
}
#endif
