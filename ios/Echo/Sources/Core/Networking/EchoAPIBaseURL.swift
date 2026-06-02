import Foundation

/// Resolves the Echo REST API base URL for TestFlight and local dev.
///
/// Precedence (matches `SilentProvisionService` / Xcode scheme docs):
/// 1. `ECHO_API_URL` environment variable
/// 2. `API_URL` environment variable
/// 3. `API_URL` in the app Info.plist (build setting substitution)
/// 4. `https://api.echo.local`
enum EchoAPIBaseURL {

    private static let fallback = URL(string: "https://api.echo.local")!

    static var resolved: URL {
        let env = ProcessInfo.processInfo.environment
        if let raw = env["ECHO_API_URL"] ?? env["API_URL"],
           let url = URL(string: raw.trimmingCharacters(in: .whitespacesAndNewlines)) {
            return url
        }
        if let raw = Bundle.main.object(forInfoDictionaryKey: "API_URL") as? String,
           !raw.isEmpty,
           let url = URL(string: raw.trimmingCharacters(in: .whitespacesAndNewlines)) {
            return url
        }
        return fallback
    }

    /// Builds an absolute URL from a path such as `/v1/auth/sms-recovery/register`.
    static func url(path: String) -> URL {
        let trimmed = path.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return resolved }
        if trimmed.hasPrefix("http://") || trimmed.hasPrefix("https://") {
            return URL(string: trimmed) ?? resolved
        }
        var components = URLComponents(url: resolved, resolvingAgainstBaseURL: false)!
        let suffix = trimmed.hasPrefix("/") ? String(trimmed.dropFirst()) : trimmed
        let basePath = components.path.hasSuffix("/") ? String(components.path.dropLast()) : components.path
        components.path = basePath.isEmpty ? "/\(suffix)" : "\(basePath)/\(suffix)"
        return components.url ?? resolved.appendingPathComponent(suffix)
    }
}
