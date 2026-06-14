#if os(iOS)
import Foundation

/// WebRTC peer session (M4b). Uses structured SDP JSON + ICE trickle; swap body when WebRTC.framework is linked.
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

    private(set) var phase: Phase = .idle
    private(set) var localSDP: String?
    private(set) var remoteSDP: String?

    private var iceServers: [CallICEServer] = []
    private var iceUfrag = UUID().uuidString.prefix(8).lowercased()
    private var onIceCandidate: ((Data) -> Void)?

    func configure(iceServers: [CallICEServer], onIceCandidate: @escaping (Data) -> Void) {
        self.iceServers = iceServers
        self.onIceCandidate = onIceCandidate
    }

    func createOffer(callType: CallType) async throws -> String {
        phase = .offering
        let desc = WebRTCSessionDescription.offer(callType: callType, iceUfrag: String(iceUfrag))
        let encoded = try desc.encoded()
        localSDP = encoded
        await emitStubCandidates()
        return encoded
    }

    func applyRemoteOffer(_ sdp: String, callType: CallType) async throws -> String {
        phase = .ringing
        remoteSDP = sdp
        let offer = try WebRTCSessionDescription.decode(from: sdp)
        let answer = WebRTCSessionDescription.answer(for: offer)
        let encoded = try answer.encoded()
        localSDP = encoded
        await emitStubCandidates()
        return encoded
    }

    func applyAnswer(_ sdp: String) async {
        remoteSDP = sdp
        phase = .connecting
        try? await Task.sleep(nanoseconds: 300_000_000)
        phase = .connected
    }

    func addIceCandidate(_ data: Data) async {
        _ = try? ICECandidatePayload.decode(from: data)
        if phase == .connecting {
            phase = .connected
        }
    }

    func hangup() {
        phase = .ended
        onIceCandidate = nil
    }

    var configuredIceServerCount: Int { iceServers.count }

    private func emitStubCandidates() async {
        guard let onIceCandidate else { return }
        let stun = iceServers.first?.urls.first ?? "stun:stun.l.google.com:19302"
        let payload = ICECandidatePayload(
            candidate: "candidate:1 1 UDP 2130706431 192.168.0.1 54321 typ host",
            sdpMid: "0",
            sdpMLineIndex: 0
        )
        if let data = try? payload.encoded() {
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
