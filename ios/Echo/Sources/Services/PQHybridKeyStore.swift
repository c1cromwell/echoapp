#if os(iOS)
import CryptoKit
import Foundation
import MLKEMNativeSwift

/// Device-local hybrid KEM key material for PQ ratchet bootstrap (T1).
actor PQHybridKeyStore {
    static let shared = PQHybridKeyStore()

    private let ecKey = "echo.pq_hybrid.local.ec.priv"
    private let pqKey = "echo.pq_hybrid.local.pq.priv"

    func loadOrGenerateBundle() async throws -> HybridPublicBundleWire {
        if let ecRaw = UserDefaults.standard.data(forKey: ecKey),
           let pqRaw = UserDefaults.standard.data(forKey: pqKey) {
            let ec = try P256.KeyAgreement.PrivateKey(rawRepresentation: ecRaw)
            let pq = try MLKEMNative768.PrivateKey(representation: pqRaw)
            return try PQHybridCrypto.publicBundle(ec: ec, pq: pq)
        }
        let generated = try PQHybridCrypto.generateKeyPair()
        UserDefaults.standard.set(generated.ec.rawRepresentation, forKey: ecKey)
        UserDefaults.standard.set(generated.pq.representation, forKey: pqKey)
        return generated.bundle
    }

    func loadPrivateKeys() async throws -> (ec: P256.KeyAgreement.PrivateKey, pq: MLKEMNative768.PrivateKey)? {
        guard let ecRaw = UserDefaults.standard.data(forKey: ecKey),
              let pqRaw = UserDefaults.standard.data(forKey: pqKey) else { return nil }
        return (
            try P256.KeyAgreement.PrivateKey(rawRepresentation: ecRaw),
            try MLKEMNative768.PrivateKey(representation: pqRaw)
        )
    }
}
#endif
