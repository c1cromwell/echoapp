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

enum OPRFRuntimeMode: Sendable {
    case live
    case mockDev
}

enum OPRFClientFactory {
    static var runtimeMode: OPRFRuntimeMode {
        #if canImport(Echooprf)
        .live
        #else
        .mockDev
        #endif
    }

    /// Short label for Privacy → Contact discovery UI (WO-221).
    static var modeFootnote: String {
        switch runtimeMode {
        case .live:
            return "Private scan uses live OPRF (EchoOPRF). Matches require the same framework build on both devices."
        case .mockDev:
            return "Dev build uses a mock OPRF — scans won't match real users until you embed EchoOPRF.xcframework (`make echooprf-ios`)."
        }
    }

    static func makeDefault() -> any OPRFClient {
        #if canImport(Echooprf)
        return LiveOPRFClient()
        #else
        return MockOPRFClient()
        #endif
    }
}
#endif
