#if os(iOS)
import Foundation

/// Sends/receives WebRTC call signaling over the shared messaging WebSocket (M4).
@MainActor
final class CallSignalingService {
    private let signalService: ConversationSignalService
    private let iceAPI: CallICEAPIClient
    private var onIncoming: (@Sendable (CallSignalEvent) -> Void)?
    private var localDID: String = ""

    init(signalService: ConversationSignalService, iceAPI: CallICEAPIClient) {
        self.signalService = signalService
        self.iceAPI = iceAPI
        signalService.setCallSignalHandler { [weak self] event in
            Task { @MainActor in self?.onIncoming?(event) }
        }
    }

    func configure(localDID: String, onIncoming: @escaping @Sendable (CallSignalEvent) -> Void) {
        self.localDID = localDID
        self.onIncoming = onIncoming
    }

    func fetchICEServers() async throws -> [CallICEServer] {
        try await iceAPI.fetchICEServers()
    }

    func sendOffer(callId: String, to peerDID: String, callType: CallType, sdp: String) async throws {
        try await send(
            to: peerDID,
            payload: CallSignalPayload(
                callId: callId,
                action: CallSignalAction.offer.rawValue,
                callType: callType.rawValue,
                sdp: sdp,
                iceCandidate: nil
            )
        )
    }

    func sendAnswer(callId: String, to peerDID: String, sdp: String) async throws {
        try await send(
            to: peerDID,
            payload: CallSignalPayload(
                callId: callId,
                action: CallSignalAction.answer.rawValue,
                callType: nil,
                sdp: sdp,
                iceCandidate: nil
            )
        )
    }

    func sendIce(callId: String, to peerDID: String, candidate: Data) async throws {
        try await send(
            to: peerDID,
            payload: CallSignalPayload(
                callId: callId,
                action: CallSignalAction.ice.rawValue,
                callType: nil,
                sdp: nil,
                iceCandidate: candidate
            )
        )
    }

    func sendHangup(callId: String, to peerDID: String) async throws {
        try await send(
            to: peerDID,
            payload: CallSignalPayload(
                callId: callId,
                action: CallSignalAction.hangup.rawValue,
                callType: nil,
                sdp: nil,
                iceCandidate: nil
            )
        )
    }

    func sendReject(callId: String, to peerDID: String) async throws {
        try await send(
            to: peerDID,
            payload: CallSignalPayload(
                callId: callId,
                action: CallSignalAction.reject.rawValue,
                callType: nil,
                sdp: nil,
                iceCandidate: nil
            )
        )
    }

    private func send(to peerDID: String, payload: CallSignalPayload) async throws {
        let wire = try CallSignalCodec.encode(to: peerDID, payload: payload)
        try await signalService.sendRaw(wire: wire)
    }
}
#endif
