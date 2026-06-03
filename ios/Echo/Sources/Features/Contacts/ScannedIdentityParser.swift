#if os(iOS)
import Foundation

/// Parses QR payloads: raw `did:key` or `echo://profile?did=…&u=…`.
struct ScannedIdentity: Equatable, Sendable {
    let did: String
    let username: String?
}

enum ScannedIdentityParser {
    static func parse(_ raw: String) -> ScannedIdentity? {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        if trimmed.hasPrefix("did:key:") {
            return ScannedIdentity(did: trimmed, username: nil)
        }

        if let url = URL(string: trimmed),
           url.scheme?.lowercased() == "echo",
           let host = url.host?.lowercased(),
           host == "profile" || host == "user" {
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
            let did = components?.queryItems?.first(where: { $0.name == "did" })?.value
            let username = components?.queryItems?.first(where: { $0.name == "u" })?.value
            if let did, did.hasPrefix("did:key:") {
                return ScannedIdentity(did: did, username: username)
            }
        }

        return nil
    }
}
#endif
