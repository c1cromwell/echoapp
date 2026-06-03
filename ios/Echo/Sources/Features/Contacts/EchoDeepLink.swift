#if os(iOS)
import Foundation

/// Parses `echo://` URLs for invites, profile, and device link (WO-222 / WO-288).
enum EchoDeepLink {
    case invite(code: String)
    case profile(did: String, username: String?)
    case linkDevice(token: String)

    static func parse(_ url: URL) -> EchoDeepLink? {
        guard url.scheme?.lowercased() == "echo" else { return nil }
        let host = (url.host ?? "").lowercased()
        let path = url.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))

        switch host {
        case "invite":
            if let code = inviteCode(from: url, path: path) {
                return .invite(code: code)
            }
        case "profile", "user":
            if let identity = ScannedIdentityParser.parse(url.absoluteString) {
                return .profile(did: identity.did, username: identity.username)
            }
            if host == "user", !path.isEmpty, path.hasPrefix("did:key:") {
                return .profile(did: path, username: nil)
            }
        case "link-device", "linkdevice":
            if let token = queryValue(url, name: "token"), !token.isEmpty {
                return .linkDevice(token: token)
            }
            if !path.isEmpty { return .linkDevice(token: path) }
        default:
            break
        }
        return nil
    }

    private static func inviteCode(from url: URL, path: String) -> String? {
        if let code = queryValue(url, name: "code"), !code.isEmpty { return code }
        if !path.isEmpty { return path }
        return nil
    }

    private static func queryValue(_ url: URL, name: String) -> String? {
        URLComponents(url: url, resolvingAgainstBaseURL: false)?
            .queryItems?
            .first(where: { $0.name == name })?
            .value
    }

    /// Persist invite until the user reaches the authenticated shell.
    static func stashPendingInvite(_ code: String) {
        UserDefaults.standard.set(code, forKey: pendingInviteKey)
    }

    static func consumePendingInvite() -> String? {
        let code = UserDefaults.standard.string(forKey: pendingInviteKey)
        UserDefaults.standard.removeObject(forKey: pendingInviteKey)
        return code
    }

    private static let pendingInviteKey = "echo.pendingInviteCode"
}
#endif
