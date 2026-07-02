#if os(iOS)
import Foundation

/// Shared peer-session contract used by the stub and live WebRTC engines (M4c).
@MainActor
protocol WebRTCCallEngine: AnyObject {
    var phase: WebRTCCallSession.Phase { get }
    var usesLiveWebRTC: Bool { get }
    var localVideoTrack: Any? { get }
    var remoteVideoTrack: Any? { get }

    func configure(
        iceServers: [CallICEServer],
        callType: CallType,
        onIceCandidate: @escaping (Data) -> Void,
        onConnected: @escaping () -> Void
    )
    func createOffer() async throws -> String
    func applyRemoteOffer(_ sdp: String) async throws -> String
    func applyAnswer(_ sdp: String) async throws
    func addIceCandidate(_ data: Data) async
    func setMuted(_ muted: Bool)
    func setVideoEnabled(_ enabled: Bool)
    func setSpeakerEnabled(_ enabled: Bool)
    func startScreenShare() async throws
    func stopScreenShare()
    var isScreenSharing: Bool { get }
    func hangup()
}

enum WebRTCCallEngineFactory {
    @MainActor
    static func make() -> WebRTCCallEngine {
        #if canImport(WebRTC)
        return WebRTCLiveCallEngine()
        #else
        return WebRTCStubCallEngine()
        #endif
    }
}
#endif
