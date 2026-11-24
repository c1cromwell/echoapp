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

/// Stands in when the real OPRF library is not linked. It refuses to blind anything, so no
/// contact data can leave the device.
struct UnavailableOPRFClient: OPRFClient {
    func blind(phones: [String]) async throws -> OPRFBlindResult {
        throw ContactDiscoveryError.discoveryUnavailable
    }

    func finalize(sessionID: String, evaluated: [String]) async throws -> [String] {
        throw ContactDiscoveryError.discoveryUnavailable
    }
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
            return "Contact discovery is off in this build: the private-matching library (EchoOPRF.xcframework, `make echooprf-ios`) is not embedded, and contacts are never uploaded without it."
        }
    }

    /// The client the app uses. Without the real OPRF library there is no private way to match
    /// contacts, so discovery is switched off: it must never fall back to the mock, whose
    /// "blinded" values are the phone numbers themselves, readable by the server.
    static func makeDefault() -> any OPRFClient {
        #if canImport(Echooprf)
        return LiveOPRFClient()
        #else
        return UnavailableOPRFClient()
        #endif
    }
}
#endif
