#if os(iOS)
import Foundation

private struct UserPrefsEnvelope: Codable, Sendable {
    let prefs: [String: String]
}

private enum UserPrefsEndpoint: APIEndpoint {
    case prefs
    var path: String { "/v3/prefs" }
}

actor LiveUserPrefsAPIClient {
    private let apiClient: APIClient
    init(apiClient: APIClient) { self.apiClient = apiClient }

    func fetch() async throws -> [String: String] {
        let env: UserPrefsEnvelope = try await apiClient.get(endpoint: UserPrefsEndpoint.prefs)
        return env.prefs
    }

    func save(_ prefs: [String: String]) async throws {
        let _: UserPrefsEnvelope = try await apiClient.put(
            endpoint: UserPrefsEndpoint.prefs, body: UserPrefsEnvelope(prefs: prefs))
    }
}

/// Roams a whitelist of on-device preferences across devices via `/v3/prefs`
/// (last-write-wins). Values are serialized as strings. Mirrors ChatFolderStore's
/// best-effort sync. UserDefaults is thread-safe, so these are nonisolated.
enum PreferenceSync {
    /// UserDefaults keys that roam across devices.
    static let syncedKeys: [String] = [
        "echo.notify.messages", "echo.notify.previews", "echo.notify.sounds",
        "echo.notify.groupMentionsOnly", "echo.notify.contactRequests", "echo.notify.rewards",
        "selected_theme",
        "echo.globalSilent", "echo.defaultDisappearing",
    ]

    private static func snapshot() -> [String: String] {
        let d = UserDefaults.standard
        var out: [String: String] = [:]
        for key in syncedKeys {
            guard let v = d.object(forKey: key) else { continue }
            if let b = v as? Bool { out[key] = b ? "true" : "false" }
            else { out[key] = "\(v)" }
        }
        return out
    }

    private static func apply(_ prefs: [String: String]) {
        let d = UserDefaults.standard
        for (key, value) in prefs where syncedKeys.contains(key) {
            switch value {
            case "true": d.set(true, forKey: key)
            case "false": d.set(false, forKey: key)
            default: d.set(value, forKey: key)
            }
        }
    }

    /// Pull the server copy and apply known keys (call on launch / hub appear).
    static func hydrate() async {
        guard let api = await DIContainer.shared.resolveAPIClient() else { return }
        guard let remote = try? await LiveUserPrefsAPIClient(apiClient: api).fetch(), !remote.isEmpty else { return }
        apply(remote)
    }

    /// Best-effort push of the current preference snapshot.
    static func push() {
        let snap = snapshot()
        Task {
            guard let api = await DIContainer.shared.resolveAPIClient() else { return }
            try? await LiveUserPrefsAPIClient(apiClient: api).save(snap)
        }
    }
}
#endif
