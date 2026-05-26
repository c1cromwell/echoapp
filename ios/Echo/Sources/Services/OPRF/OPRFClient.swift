#if os(iOS)
import Foundation
import CryptoKit

/// OPRF blind/finalize for contact discovery (WO-221). Production builds link `EchoOPRF` via gomobile;
/// tests inject `MockOPRFClient`.
protocol OPRFClient: Sendable {
    func blind(phones: [String]) async throws -> OPRFBlindResult
    func finalize(sessionID: String, evaluated: [String]) async throws -> [String]
}

struct OPRFBlindResult: Sendable {
    let sessionID: String
    let blinded: [String]
}

/// In-memory mock for unit tests — returns deterministic keys from phone hash (not production safe).
struct MockOPRFClient: OPRFClient {
    private struct Session { let phones: [String] }

    private let lock = NSLock()
    private var sessions: [String: Session] = [:]

    func blind(phones: [String]) async throws -> OPRFBlindResult {
        let id = UUID().uuidString
        lock.lock()
        sessions[id] = Session(phones: phones.map { PhoneNormalizer.normalize($0) })
        lock.unlock()
        let blinded = phones.enumerated().map { idx, phone in
            Data("mock:\(PhoneNormalizer.normalize(phone)):\(idx)".utf8).base64EncodedString()
        }
        return OPRFBlindResult(sessionID: id, blinded: blinded)
    }

    func finalize(sessionID: String, evaluated: [String]) async throws -> [String] {
        lock.lock()
        let session = sessions.removeValue(forKey: sessionID)
        lock.unlock()
        guard let session else { throw ContactDiscoveryError.oprfSessionExpired }
        return session.phones.map { phone in
            let digest = SHA256Helper.hex("mock-oprf:\(phone)")
            return digest
        }
    }
}

private enum SHA256Helper {
    static func hex(_ value: String) -> String {
        let data = Data(value.utf8)
        return Data(SHA256.hash(data: data)).map { String(format: "%02x", $0) }.joined()
    }
}
#endif
