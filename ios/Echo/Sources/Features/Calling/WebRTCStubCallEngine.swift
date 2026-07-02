#if os(iOS)
import Foundation

/// Fallback peer session when WebRTC.framework is not linked (CI / headless SPM).
@MainActor
final class WebRTCStubCallEngine: WebRTCCallEngine {
    private(set) var phase: WebRTCCallSession.Phase = .idle
    let usesLiveWebRTC = false
    var localVideoTrack: Any? { nil }
    var remoteVideoTrack: Any? { nil }

    private var iceServers: [CallICEServer] = []
    private var callType: CallType = .voice
    private var iceUfrag = UUID().uuidString.prefix(8).lowercased()
    private var onIceCandidate: ((Data) -> Void)?
    private var onConnected: (() -> Void)?

    func configure(
        iceServers: [CallICEServer],
        callType: CallType,
        onIceCandidate: @escaping (Data) -> Void,
        onConnected: @escaping () -> Void
    ) {
        self.iceServers = iceServers
        self.callType = callType
        self.onIceCandidate = onIceCandidate
        self.onConnected = onConnected
    }

    func createOffer() async throws -> String {
        phase = .offering
        let desc = WebRTCSessionDescription.offer(callType: callType, iceUfrag: String(iceUfrag))
        let encoded = try desc.encoded()
        await emitStubCandidates()
        return encoded
    }

    func applyRemoteOffer(_ sdp: String) async throws -> String {
        phase = .ringing
        let offer = try WebRTCSessionDescription.decode(from: sdp)
        let answer = WebRTCSessionDescription.answer(for: offer)
        let encoded = try answer.encoded()
        await emitStubCandidates()
        return encoded
    }

    func applyAnswer(_ sdp: String) async throws {
        _ = try WebRTCSessionDescription.decode(from: sdp)
        phase = .connecting
        try await Task.sleep(nanoseconds: 300_000_000)
        phase = .connected
        onConnected?()
    }

    func addIceCandidate(_ data: Data) async {
        _ = try? ICECandidatePayload.decode(from: data)
        if phase == .connecting {
            phase = .connected
            onConnected?()
        }
    }

    func setMuted(_ muted: Bool) {}
    func setVideoEnabled(_ enabled: Bool) {}
    func setSpeakerEnabled(_ enabled: Bool) {}

    private(set) var isScreenSharing = false

    func startScreenShare() async throws { isScreenSharing = true }
    func stopScreenShare() { isScreenSharing = false }

    func hangup() {
        phase = .ended
        onIceCandidate = nil
        onConnected = nil
    }

    private func emitStubCandidates() async {
        guard let onIceCandidate else { return }
        let stun = iceServers.first?.urls.first ?? "stun:stun.l.google.com:19302"
        let host = ICECandidatePayload(
            candidate: "candidate:1 1 UDP 2130706431 192.168.0.1 54321 typ host",
            sdpMid: "0",
            sdpMLineIndex: 0
        )
        if let data = try? host.encoded() {
            onIceCandidate(data)
        }
        _ = stun
        try? await Task.sleep(nanoseconds: 50_000_000)
        let relay = ICECandidatePayload(
            candidate: "candidate:2 1 UDP 1694498815 10.0.0.1 54322 typ relay",
            sdpMid: "0",
            sdpMLineIndex: 0
        )
        if let data = try? relay.encoded() {
            onIceCandidate(data)
        }
    }
}
#endif
