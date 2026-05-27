#if os(iOS)
import Foundation

/// Resolves the authenticated user's DID for messaging and contacts (Wave 0).
enum CurrentUserSession {
    static func currentDID() async -> String? {
        if let did = try? await KeychainManager.shared.retrieve(key: "echo.did.current"), !did.isEmpty {
            return did
        }
        if let did = UserDefaults.standard.string(forKey: "echo.did"), !did.isEmpty {
            return did
        }
        return UserDefaults.standard.string(forKey: "echo.did.current")
    }

    static func displayName() -> String {
        UserDefaults.standard.string(forKey: "echo.displayName") ?? "You"
    }
}
#endif
