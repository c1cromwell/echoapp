#if os(iOS)
import Foundation

/// Optional SOCKS5 proxy for REST (WO-SX4). WebSocket uses the same `URLSessionConfiguration` hook;
/// full Tor circuit isolation is out of scope for Phase 3.
enum TransportProxySettings {
    private static let enabledKey = "echo.transport.proxy.socks.enabled"
    private static let hostKey = "echo.transport.proxy.socks.host"
    private static let portKey = "echo.transport.proxy.socks.port"

    static var isEnabled: Bool {
        get { UserDefaults.standard.bool(forKey: enabledKey) }
        set { UserDefaults.standard.set(newValue, forKey: enabledKey) }
    }

    /// Default Tor SOCKS port when unset.
    static var socksHost: String {
        get {
            let stored = UserDefaults.standard.string(forKey: hostKey) ?? ""
            return stored.isEmpty ? "127.0.0.1" : stored
        }
        set { UserDefaults.standard.set(newValue, forKey: hostKey) }
    }

    static var socksPort: Int {
        get {
            let stored = UserDefaults.standard.integer(forKey: portKey)
            return stored > 0 ? stored : 9050
        }
        set { UserDefaults.standard.set(newValue, forKey: portKey) }
    }

    static func apply(to configuration: URLSessionConfiguration) {
        guard isEnabled else { return }
        configuration.connectionProxyDictionary = [
            "SOCKSEnable": 1,
            "SOCKSProxy": socksHost,
            "SOCKSPort": socksPort,
        ]
    }
}
#endif
