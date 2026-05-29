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

    static func currentUsername() async -> String {
        if let name = try? await KeychainManager.shared.retrieve(key: "echo.username.current"), !name.isEmpty {
            return name
        }
        if let name = UserDefaults.standard.string(forKey: "echo.displayName"), !name.isEmpty {
            return name
        }
        return displayName()
    }

    static func displayName() -> String {
        UserDefaults.standard.string(forKey: "echo.displayName") ?? "You"
    }

    static func trustTier() -> Int {
        UserDefaults.standard.integer(forKey: "echo.trustTier")
    }

    /// Deep link payload for identity sharing (B.2 growth loop).
    static func identityShareURL(did: String, username: String) -> URL? {
        var components = URLComponents()
        components.scheme = "echo"
        components.host = "profile"
        components.queryItems = [
            URLQueryItem(name: "did", value: did),
            URLQueryItem(name: "u", value: username),
        ]
        return components.url
    }
}
#endif
