#if os(iOS)
import Foundation

/// Stub WebRTC session (M4a). Swaps for WebRTC.framework + CallKit in Xcode E2E.
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

    func createOffer(callType: CallType) async throws -> String {
        phase = .offering
        let sdp = "echo-stub-offer-\(callType.rawValue)-\(UUID().uuidString)"
        localSDP = sdp
        return sdp
    }

    func applyRemoteOffer(_ sdp: String, callType: CallType) async throws -> String {
        phase = .ringing
        remoteSDP = sdp
        let answer = "echo-stub-answer-\(callType.rawValue)-\(UUID().uuidString)"
        localSDP = answer
        return answer
    }

    func applyAnswer(_ sdp: String) async {
        remoteSDP = sdp
        phase = .connecting
        try? await Task.sleep(nanoseconds: 500_000_000)
        phase = .connected
    }

    func addIceCandidate(_ data: Data) async {
        _ = data
    }

    func hangup() {
        phase = .ended
    }
}
#endif
