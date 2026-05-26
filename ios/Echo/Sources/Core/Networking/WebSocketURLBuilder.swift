import Foundation

/// Builds WebSocket URLs from the REST API base (`API_URL` / `ECHO_API_URL`).
enum WebSocketURLBuilder {

    /// Converts `https://host:port` → `wss://host:port/ws?token=…`.
    static func webSocketURL(apiBaseURL: URL, accessToken: String, path: String = "/ws") -> URL? {
        guard var components = URLComponents(url: apiBaseURL, resolvingAgainstBaseURL: false) else {
            return nil
        }
        switch components.scheme?.lowercased() {
        case "https":
            components.scheme = "wss"
        case "http":
            components.scheme = "ws"
        case "wss", "ws":
            break
        default:
            components.scheme = "wss"
        }
        components.path = path.hasPrefix("/") ? path : "/\(path)"
        components.queryItems = [URLQueryItem(name: "token", value: accessToken)]
        return components.url
    }

    /// Reads API host from env/build settings (same precedence as SilentProvisionService).
    static func apiBaseURLFromEnvironment() -> URL? {
        let env = ProcessInfo.processInfo.environment
        if let raw = env["ECHO_API_URL"] ?? env["API_URL"],
           let url = URL(string: raw) {
            return url
        }
        return nil
    }
}
