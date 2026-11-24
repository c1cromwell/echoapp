import Foundation

/// Builds WebSocket URLs from the REST API base (`API_URL` / `ECHO_API_URL`).
enum WebSocketURLBuilder {

    /// Converts `https://host:port` → `wss://host:port/ws`. The access token is NOT part of the
    /// URL: it goes in the `Authorization` header (`authorizationHeader`). A token in a URL ends up
    /// in the logs of every proxy and load balancer the request passes through.
    static func webSocketURL(apiBaseURL: URL, path: String = "/ws") -> URL? {
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
        components.queryItems = nil
        return components.url
    }

    /// The header value that authenticates the socket.
    static func authorizationHeader(accessToken: String) -> String {
        "Bearer \(accessToken)"
    }

    /// Reads API host from env/build settings (same precedence as `EchoAPIBaseURL`).
    static func apiBaseURLFromEnvironment() -> URL? {
        let resolved = EchoAPIBaseURL.resolved
        let fallback = URL(string: "https://api.echo.local")!
        return resolved == fallback ? nil : resolved
    }
}
