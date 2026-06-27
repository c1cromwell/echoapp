#if os(iOS)
import XCTest
@testable import Echo

final class TransportProxySettingsTests: XCTestCase {
    func testAppliesSocksDictionaryWhenEnabled() {
        TransportProxySettings.isEnabled = true
        TransportProxySettings.socksHost = "127.0.0.1"
        TransportProxySettings.socksPort = 9050

        let config = URLSessionConfiguration.ephemeral
        TransportProxySettings.apply(to: config)

        let proxy = config.connectionProxyDictionary
        XCTAssertEqual(proxy?["SOCKSEnable"] as? Int, 1)
        XCTAssertEqual(proxy?["SOCKSProxy"] as? String, "127.0.0.1")
        XCTAssertEqual(proxy?["SOCKSPort"] as? Int, 9050)

        TransportProxySettings.isEnabled = false
        let cleared = URLSessionConfiguration.ephemeral
        TransportProxySettings.apply(to: cleared)
        XCTAssertNil(cleared.connectionProxyDictionary)
    }

    func testDefaultsTorPort() {
        TransportProxySettings.socksHost = ""
        TransportProxySettings.socksPort = 0
        XCTAssertEqual(TransportProxySettings.socksHost, "127.0.0.1")
        XCTAssertEqual(TransportProxySettings.socksPort, 9050)
    }
}
#endif
