// Core/Mesh/MeshSignalTransport.swift
//
// Bridges the BLE mesh to ECHO's existing transport seam (ConversationSignalTransport), so the
// unchanged ConversationSignalService can run over mesh exactly as it does over WebSocket.

import Foundation

/// Minimal mesh-node abstraction so the transport (and its tests) don't depend on CoreBluetooth.
/// `BLEMeshService` conforms on iOS; tests use a mock.
public protocol MeshNode: AnyObject {
    var onMessage: ((_ payload: Data, _ fromPeer: Data) -> Void)? { get set }
    func start()
    func stop()
    func send(_ payload: Data, to recipient: Data)
}

/// The wire envelope already carries E2E ciphertext + addressing (`WSEnvelope.to`), so we flood
/// each frame (recipient = broadcast) and let the app layer accept/decrypt — which is how a flood
/// mesh works and keeps `ConversationSignalService` and the crypto path unchanged.
public final class MeshSignalTransport: ConversationSignalTransport, @unchecked Sendable {
    public var onTextMessage: (@Sendable (String) -> Void)?
    private let node: MeshNode

    public init(node: MeshNode) {
        self.node = node
        node.onMessage = { [weak self] payload, _ in
            guard let text = String(data: payload, encoding: .utf8) else { return }
            self?.onTextMessage?(text)
        }
    }

    public func connect(accessToken: String) async throws { node.start() }   // offline: token unused
    public func disconnect() async { node.stop() }
    public func send(text: String) async throws {
        node.send(Data(text.utf8), to: MeshPacket.broadcast)
    }
}
