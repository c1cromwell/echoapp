#if os(iOS)
import Foundation

/// Facade over stub or live WebRTC peer engines (M4b/M4c).
@MainActor
final class WebRTCCallSession {
    enum Phase: Equatable {
        case idle
        case offering
        case ringing
        case connecting
        case connected
        case ended
    }

    private let engine: WebRTCCallEngine
    private var onConnected: (() -> Void)?

    init(engine: WebRTCCallEngine? = nil) {
        self.engine = engine ?? WebRTCCallEngineFactory.make()
    }

    var phase: Phase { engine.phase }
    var usesLiveWebRTC: Bool { engine.usesLiveWebRTC }
    var localVideoTrack: Any? { engine.localVideoTrack }
    var remoteVideoTrack: Any? { engine.remoteVideoTrack }

    func configure(
        iceServers: [CallICEServer],
        callType: CallType,
        onIceCandidate: @escaping (Data) -> Void,
        onConnected: @escaping () -> Void
    ) {
        self.onConnected = onConnected
        engine.configure(
            iceServers: iceServers,
            callType: callType,
            onIceCandidate: onIceCandidate,
            onConnected: { [weak self] in
                self?.onConnected?()
            }
        )
    }

    func createOffer(callType: CallType) async throws -> String {
        try await engine.createOffer()
    }

    func applyRemoteOffer(_ sdp: String, callType: CallType) async throws -> String {
        try await engine.applyRemoteOffer(sdp)
    }

    func applyAnswer(_ sdp: String) async throws {
        try await engine.applyAnswer(sdp)
    }

    func addIceCandidate(_ data: Data) async {
        await engine.addIceCandidate(data)
    }

    func setMuted(_ muted: Bool) { engine.setMuted(muted) }
    func setVideoEnabled(_ enabled: Bool) { engine.setVideoEnabled(enabled) }
    func setSpeakerEnabled(_ enabled: Bool) { engine.setSpeakerEnabled(enabled) }

    func startScreenShare() async throws { try await engine.startScreenShare() }
    func stopScreenShare() { engine.stopScreenShare() }
    var isScreenSharing: Bool { engine.isScreenSharing }

    func hangup() {
        engine.hangup()
    }
}
#endif
