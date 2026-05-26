#if os(iOS)
import Foundation

#if canImport(Echooprf)
import Echooprf

/// Live OPRF client backed by `mobile/echooprf` (gomobile / EchoOPRF.xcframework).
final class LiveOPRFClient: OPRFClient, @unchecked Sendable {
    private let client = EchooprfClient()

    func blind(phones: [String]) async throws -> OPRFBlindResult {
        let result = try client.blindPhones(phones)
        guard let blindResult = result else {
            throw ContactDiscoveryError.discoveryUnavailable
        }
        return OPRFBlindResult(
            sessionID: blindResult.sessionID,
            blinded: blindResult.blinded as? [String] ?? []
        )
    }

    func finalize(sessionID: String, evaluated: [String]) async throws -> [String] {
        let keys = try client.finalizePhones(sessionID, evaluated: evaluated)
        return keys as? [String] ?? []
    }
}
#endif

enum OPRFClientFactory {
    static func makeDefault() -> any OPRFClient {
        #if canImport(Echooprf)
        return LiveOPRFClient()
        #else
        return MockOPRFClient()
        #endif
    }
}
#endif
