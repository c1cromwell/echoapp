#if os(iOS)
import Foundation

/// Opt-in data sovereignty settings (WO-248 stub).
enum DataSovereigntySettingsStore {
    private static let key = "echo.datasov.optedIn"

    static var isOptedIn: Bool {
        UserDefaults.standard.bool(forKey: key)
    }

    static func setOptedIn(_ value: Bool) {
        UserDefaults.standard.set(value, forKey: key)
        Task { await syncRemote(optedIn: value) }
    }

    private static func syncRemote(optedIn: Bool) async {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: APIConfiguration.default.baseURL.absoluteString + "/v3/datasov/settings") else {
            return
        }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONEncoder().encode(["opted_in": optedIn])
        _ = try? await URLSession.shared.data(for: request)
    }
}
#endif
