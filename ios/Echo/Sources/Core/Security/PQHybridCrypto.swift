#if os(iOS)
import CryptoKit
import Foundation
import MLKEMNativeSwift

/// Post-quantum hybrid KEM matching Go `internal/crypto/pqhybrid.go` (WO-SX2).
enum PQHybridCrypto {
    private static let combinerInfo = "ECHO-PQ-HYBRID-KEM-v1"

    enum Error: Swift.Error {
        case invalidKeyMaterial
        case combinerFailed
    }

    static func generateKeyPair() throws -> (
        ec: P256.KeyAgreement.PrivateKey,
        pq: MLKEMNative768.PrivateKey,
        bundle: HybridPublicBundleWire
    ) {
        let ec = P256.KeyAgreement.PrivateKey()
        let pq = try MLKEMNative768.PrivateKey.generate()
        let bundle = try publicBundle(ec: ec, pq: pq)
        return (ec, pq, bundle)
    }

    static func publicBundle(
        ec: P256.KeyAgreement.PrivateKey,
        pq: MLKEMNative768.PrivateKey
    ) throws -> HybridPublicBundleWire {
        let ecRaw = rawP256Public(ec)
        let pqRaw = pq.publicKey.rawRepresentation
        guard ecRaw.count == 64, pqRaw.count == 1184 else { throw Error.invalidKeyMaterial }
        return HybridPublicBundleWire(
            ec: ecRaw.base64EncodedString(),
            pq: pqRaw.base64EncodedString()
        )
    }

    static func encapsulate(remote: HybridPublicBundleWire) throws -> (HybridCiphertextWire, Data) {
        guard let remoteECRaw = Data(base64Encoded: remote.ec),
              let remotePQRaw = Data(base64Encoded: remote.pq),
              remoteECRaw.count == 64,
              remotePQRaw.count == 1184 else {
            throw Error.invalidKeyMaterial
        }

        let ephemeral = P256.KeyAgreement.PrivateKey()
        let ssEC = try rawECDH(local: ephemeral, peerRaw: remoteECRaw)

        let pqPub = try MLKEMNative768.PublicKey(rawRepresentation: remotePQRaw)
        let encapsulated = try pqPub.encapsulate()
        let ctPQ = encapsulated.ciphertext
        guard ctPQ.count == 1088 else { throw Error.invalidKeyMaterial }
        let ssPQ = symmetricBytes(encapsulated.sharedSecret)

        let ephemeralRaw = rawP256Public(ephemeral)
        let wire = HybridCiphertextWire(
            ephemeralEC: ephemeralRaw.base64EncodedString(),
            pq: ctPQ.base64EncodedString()
        )
        let secret = try hybridCombine(ssEC: ssEC, ssPQ: ssPQ, ephemeralEC: ephemeralRaw, ctPQ: ctPQ)
        return (wire, secret)
    }

    static func decapsulate(
        ec: P256.KeyAgreement.PrivateKey,
        pq: MLKEMNative768.PrivateKey,
        ciphertext: HybridCiphertextWire
    ) throws -> Data {
        guard let ephemeralECRaw = Data(base64Encoded: ciphertext.ephemeralEC),
              let ctPQ = Data(base64Encoded: ciphertext.pq),
              ephemeralECRaw.count == 64,
              ctPQ.count == 1088 else {
            throw Error.invalidKeyMaterial
        }

        let ssEC = try rawECDH(local: ec, peerRaw: ephemeralECRaw)
        let ssPQ = symmetricBytes(try pq.decapsulate(ctPQ))
        return try hybridCombine(ssEC: ssEC, ssPQ: ssPQ, ephemeralEC: ephemeralECRaw, ctPQ: ctPQ)
    }

    private static func hybridCombine(
        ssEC: Data,
        ssPQ: Data,
        ephemeralEC: Data,
        ctPQ: Data
    ) throws -> Data {
        var ikm = Data()
        ikm.append(ssEC)
        ikm.append(ssPQ)

        var transcript = SHA256()
        transcript.update(data: ephemeralEC)
        transcript.update(data: ctPQ)
        let digest = Data(transcript.finalize())

        var info = Data(combinerInfo.utf8)
        info.append(digest)

        let derived = SymmetricKey(data: ikm).hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data(),
            sharedInfo: info,
            outputByteCount: 32
        )
        return symmetricBytes(derived)
    }

    private static func rawECDH(local: P256.KeyAgreement.PrivateKey, peerRaw: Data) throws -> Data {
        let peer = try P256.KeyAgreement.PublicKey(rawRepresentation: normalizedP256Pub(peerRaw))
        let secret = try local.sharedSecretFromKeyAgreement(with: peer)
        return symmetricBytes(secret)
    }

    private static func rawP256Public(_ key: P256.KeyAgreement.PrivateKey) -> Data {
        let pub = key.publicKey.rawRepresentation
        if pub.count == 65, pub.first == 0x04 { return Data(pub.dropFirst()) }
        return pub
    }

    private static func normalizedP256Pub(_ raw: Data) -> Data {
        if raw.count == 64 { return raw }
        if raw.count == 65, raw.first == 0x04 { return Data(raw.dropFirst()) }
        return raw
    }

    private static func symmetricBytes(_ key: SymmetricKey) -> Data {
        key.withUnsafeBytes { Data($0) }
    }
}
#endif
