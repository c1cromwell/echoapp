import Foundation

/// `@username` is the public invite. Opaque `code=` links still parse for older shares.
enum InviteHandle {
    /// Strip whitespace and a leading `@`. Case is preserved for username lookup.
    static func normalize(_ raw: String) -> String {
        var handle = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        while handle.hasPrefix("@") {
            handle.removeFirst()
            handle = handle.trimmingCharacters(in: .whitespacesAndNewlines)
        }
        return handle
    }

    static func display(_ raw: String) -> String {
        let handle = normalize(raw)
        return handle.isEmpty ? "" : "@\(handle)"
    }

    /// Token stashed in `pendingInviteCode` so accept UI can tell handle from opaque code.
    static func pendingToken(username: String) -> String {
        display(username)
    }

    static func isUsernameToken(_ token: String) -> Bool {
        let trimmed = token.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.hasPrefix("@") && !normalize(trimmed).isEmpty
    }

    static func shareURL(username: String) -> URL? {
        let handle = normalize(username)
        guard !handle.isEmpty else { return nil }
        var components = URLComponents()
        components.scheme = "echo"
        components.host = "invite"
        components.queryItems = [URLQueryItem(name: "u", value: handle)]
        return components.url
    }
}

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
            if let handle = inviteHandle(from: url, path: path) {
                return .invite(code: InviteHandle.pendingToken(username: handle))
            }
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

    private static func inviteHandle(from url: URL, path: String) -> String? {
        if let raw = queryValue(url, name: "u") {
            let handle = InviteHandle.normalize(raw)
            if !handle.isEmpty { return handle }
        }
        if path.hasPrefix("@") {
            let handle = InviteHandle.normalize(path)
            if !handle.isEmpty { return handle }
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
