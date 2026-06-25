// Core/Mesh/MeshPeerDirectory.swift
//
// Offline, server-less binding of an identity (did:key) to its messaging key-agreement key —
// the crux that lets *verified* messaging run on the mesh when IdentityResolveClient is
// unreachable. A holder signs (did ‖ kaPublicKey) with its identity signing key; verifiers
// re-derive the did:key from the included signing public key (so the DID can't be forged) and
// check the ECDSA signature. No network required.

import Foundation
import CryptoKit

public struct MeshKeyCert: Codable, Sendable, Equatable {
    public let did: String
    public let signingPublicKey: Data   // P-256 X9.63 (65B) identity signing key
    public let kaPublicKey: Data        // P-256 X9.63 (65B) messaging key-agreement key
    public let signature: Data          // ECDSA-P256 over (did ‖ kaPublicKey), raw r‖s

    public init(did: String, signingPublicKey: Data, kaPublicKey: Data, signature: Data) {
        self.did = did
        self.signingPublicKey = signingPublicKey
        self.kaPublicKey = kaPublicKey
        self.signature = signature
    }
}

public enum MeshPeerDirectory {
    static func signedPayload(did: String, kaPublicKey: Data) -> Data {
        Data("echo-mesh-keycert-v1".utf8) + Data(did.utf8) + kaPublicKey
    }

    /// Build a cert signing with a P-256 identity key (sim/dev path; on device the app injects a
    /// Secure-Enclave signer that produces the same `(did ‖ kaPublicKey)` ECDSA signature).
    public static func makeCert(
        did: String,
        identityKey: P256.Signing.PrivateKey,
        kaPublicKey: Data
    ) throws -> MeshKeyCert {
        let sig = try identityKey.signature(for: signedPayload(did: did, kaPublicKey: kaPublicKey))
        return MeshKeyCert(
            did: did,
            signingPublicKey: identityKey.publicKey.x963Representation,
            kaPublicKey: kaPublicKey,
            signature: sig.rawRepresentation
        )
    }

    /// Verify a cert fully offline. Returns the bound `(did, kaPublicKey)` iff the DID is bound to
    /// the signing key (did:key re-derivation matches) AND the signature is valid; else nil.
    public static func verify(_ cert: MeshKeyCert) -> (did: String, kaPublicKey: Data)? {
        guard let signingKey = try? P256.Signing.PublicKey(x963Representation: cert.signingPublicKey),
              let derivedDID = try? DidKeyDeriver.deriveFromPublicKeyBytes(cert.signingPublicKey),
              derivedDID == cert.did,
              let sig = try? P256.Signing.ECDSASignature(rawRepresentation: cert.signature),
              signingKey.isValidSignature(sig, for: signedPayload(did: cert.did, kaPublicKey: cert.kaPublicKey))
        else { return nil }
        return (cert.did, cert.kaPublicKey)
    }
}

/// In-memory cache of peers whose key-cert we've verified over the mesh. Bounded; feeds the
/// verified-lane crypto path (peer DID → messaging KA public key) without any server.
public final class MeshPeerCache {
    private var byDID: [String: Data] = [:]   // did -> kaPublicKey
    private var order: [String] = []
    private let cap: Int
    public init(capacity: Int = 256) { cap = capacity }

    /// Verify and store a received cert. Returns the bound did on success.
    @discardableResult
    public func ingest(_ cert: MeshKeyCert) -> String? {
        guard let (did, ka) = MeshPeerDirectory.verify(cert) else { return nil }
        if byDID[did] == nil { order.append(did) }
        byDID[did] = ka
        while order.count > cap, let oldest = order.first {
            order.removeFirst(); byDID[oldest] = nil
        }
        return did
    }

    public func kaPublicKey(forDID did: String) -> Data? { byDID[did] }
    public var count: Int { byDID.count }
}
