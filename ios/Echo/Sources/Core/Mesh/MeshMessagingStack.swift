// Core/Mesh/MeshMessagingStack.swift
//
// Assembles the mesh messaging path: a BLE node, the offline peer-key cache (fed by key
// announces), the transport seam, and the store-and-forward queue. OPT-IN — this does not replace
// the default WebSocket path; the mesh feature builds a stack and runs `combined(with:)` so
// messages flow over relay + mesh in parallel. Live runtime is verified on hardware
// (docs/MESH_TWO_DEVICE_TEST.md).

#if os(iOS)
import Foundation
import CryptoKit

/// Supplies this device's signed messaging-key cert for broadcast (offline verified bootstrap).
/// The dev/sim path uses `SoftwareMeshCertProvider`; the on-device provider signs with the
/// Secure-Enclave identity key (`echo-identity-signing`) — see MeshMessagingStack docs.
public protocol MeshCertProvider: Sendable {
    func currentCert() async throws -> MeshKeyCert
}

final class MeshMessagingStack {
    let mesh: BLEMeshService
    let peerCache = MeshPeerCache()
    let queue = OfflineMessageQueue()
    let transport: MeshSignalTransport
    private let certProvider: MeshCertProvider?

    init(localPeerID: Data, certProvider: MeshCertProvider? = nil) {
        self.mesh = BLEMeshService(localID: localPeerID)
        self.transport = MeshSignalTransport(node: mesh)
        self.certProvider = certProvider

        // Verified-peer key bootstrap: ingest every key-announce cert we hear.
        mesh.onKeyAnnounce = { [weak self] data, _ in
            guard let cert = MeshKeyAnnounce.decode(data) else { return }
            _ = self?.peerCache.ingest(cert)
        }
        // Store-and-forward: flush queued frames when a peer appears.
        mesh.onPeerCountChange = { [weak self] count in
            guard count > 0 else { return }
            Task { await self?.flushQueue() }
        }
    }

    func start() {
        mesh.start()
        Task { await announceCert() }
    }

    func stop() { mesh.stop() }

    /// Run the mesh alongside another transport (e.g. the WebSocket relay): online + offline in
    /// parallel. Returns the combined transport to hand to `ConversationSignalService`.
    func combined(with other: ConversationSignalTransport) -> ConversationSignalTransport {
        TransportRouter([other, transport])
    }

    private func announceCert() async {
        guard let cert = try? await certProvider?.currentCert() else { return }
        mesh.announce(MeshKeyAnnounce.encode(cert))
    }

    private func flushQueue() async {
        await queue.flush { [weak self] queued in
            // Mesh send is fire-and-forget flooding; treat as relayed. Real delivery is confirmed
            // by the recipient's ack over the normal message path (which calls markDelivered).
            self?.mesh.send(Data(queued.wire.utf8), to: MeshPacket.broadcast)
            return true
        }
    }

    /// Build a stack for the current user (peer id derived from their DID). `certProvider` is
    /// injected by the mesh feature; pass the Secure-Enclave provider on device.
    static func forCurrentUser(certProvider: MeshCertProvider? = nil) async -> MeshMessagingStack? {
        guard let did = await CurrentUserSession.currentDID(), !did.isEmpty else { return nil }
        return MeshMessagingStack(localPeerID: MeshPacket.peerID(from: Data(did.utf8)),
                                  certProvider: certProvider)
    }
}

/// Dev/sim cert provider backed by an in-memory P-256 identity key. On device, replace with a
/// provider that re-derives the same `(did ‖ kaPublicKey)` signature from the Secure-Enclave
/// identity key via `SecureEnclaveManager.sign(...)` (convert its DER signature to raw r‖s).
public struct SoftwareMeshCertProvider: MeshCertProvider {
    let did: String
    let identityKey: P256SigningKeyBox
    let kaPublicKey: Data

    public func currentCert() async throws -> MeshKeyCert {
        try MeshPeerDirectory.makeCert(did: did, identityKey: identityKey.key, kaPublicKey: kaPublicKey)
    }
}

/// Sendable wrapper so a `P256.Signing.PrivateKey` can live in a `Sendable` provider.
public struct P256SigningKeyBox: @unchecked Sendable {
    let key: P256.Signing.PrivateKey
    public init(_ key: P256.Signing.PrivateKey) { self.key = key }
}
#endif
