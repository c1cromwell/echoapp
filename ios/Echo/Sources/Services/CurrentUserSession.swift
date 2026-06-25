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

    /// True once the user has completed a government-ID + selfie/liveness verification (IDV or
    /// digital-ID). Set only by the enrollment tail — NOT by a paid VIP tier — so features that
    /// require a verified human (e.g. verified BLE mesh) should gate on this, not `trustTier()`.
    static func isIdentityVerified() -> Bool {
        UserDefaults.standard.bool(forKey: "echo.idvVerified")
    }

    /// The kind of verification completed: `"standard_idv"` or `"digital_id"` (nil if unverified).
    static func identityEvidenceType() -> String? {
        UserDefaults.standard.string(forKey: "echo.evidenceType")
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
