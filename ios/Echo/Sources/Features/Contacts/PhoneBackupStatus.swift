#if os(iOS)
import Foundation

/// Local + keychain state for SMS phone backup (WO-12 / PSI discoverability).
enum PhoneBackupStatus {
    private static let keychainKey = "echo.sms.phone_hash"
    private static let displayKey = "echo.display.masked_phone"

    static var hasBackup: Bool {
        UserDefaults.standard.bool(forKey: "echo.phone.backup.configured")
    }

    static var displayLabel: String {
        if let masked = UserDefaults.standard.string(forKey: displayKey), !masked.isEmpty {
            return masked
        }
        return hasBackup ? "Phone backup on file" : "Not set"
    }

    static func markConfigured(maskedPhone: String) {
        UserDefaults.standard.set(true, forKey: "echo.phone.backup.configured")
        if !maskedPhone.isEmpty {
            UserDefaults.standard.set(maskedPhone, forKey: displayKey)
        }
    }

    static func maskedPhone(fromE164 e164: String) -> String {
        let digits = e164.filter(\.isNumber)
        guard digits.count >= 4 else { return "Phone verified" }
        let last4 = String(digits.suffix(4))
        if e164.hasPrefix("+1"), digits.count >= 11 {
            return "+1 ••••••\(last4)"
        }
        return "••••\(last4)"
    }
}
#endif
