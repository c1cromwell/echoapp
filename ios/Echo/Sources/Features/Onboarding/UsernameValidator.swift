#if os(iOS)
import Foundation

/// Handle rules aligned with `internal/api/username_handlers.go` (WO-14 / WO-292).
enum UsernameValidator {
    private static let pattern = #"^[a-zA-Z0-9_]{3,30}$"#

    static func isValid(_ raw: String) -> Bool {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.range(of: pattern, options: .regularExpression) != nil else {
            return false
        }
        return true
    }
}
#endif
