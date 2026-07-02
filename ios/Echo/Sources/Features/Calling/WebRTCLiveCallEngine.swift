#if os(iOS)
#if canImport(WebRTC)
import Foundation
import WebRTC

/// Real RTCPeerConnection session (M4c). Linked when the WebRTC SPM package is resolved in Xcode.
@MainActor
final class WebRTCLiveCallEngine: NSObject, WebRTCCallEngine {
    private(set) var phase: WebRTCCallSession.Phase = .idle
    let usesLiveWebRTC = true
    private(set) var localVideoTrack: Any?
    private(set) var remoteVideoTrack: Any?

    private static var factory: RTCPeerConnectionFactory = {
        RTCInitializeSSL()
        return RTCPeerConnectionFactory()
    }()

    private var peerConnection: RTCPeerConnection?
    private var localAudioTrack: RTCAudioTrack?
    private var videoSource: RTCVideoSource?
    private var capturer: RTCCameraVideoCapturer?
    private var screenCapturer: ScreenShareCapturer?
    private var callType: CallType = .voice
    private var onIceCandidate: ((Data) -> Void)?
    private var onConnected: (() -> Void)?

    private lazy var mediaConstraints: RTCMediaConstraints = {
        RTCMediaConstraints(mandatoryConstraints: nil, optionalConstraints: ["DtlsSrtpKeyAgreement": "true"])
    }()

    func configure(
        iceServers: [CallICEServer],
        callType: CallType,
        onIceCandidate: @escaping (Data) -> Void,
        onConnected: @escaping () -> Void
    ) {
        self.callType = callType
        self.onIceCandidate = onIceCandidate
        self.onConnected = onConnected
        try? WebRTCAudioSession.activate(callType: callType)
        peerConnection = makePeerConnection(iceServers: iceServers)
        attachLocalTracks()
    }

    func createOffer() async throws -> String {
        guard let peerConnection else { throw WebRTCCallSessionError.iceUnavailable }
        phase = .offering
        let offer = try await asyncOffer(on: peerConnection)
        try await asyncSetLocalDescription(offer, on: peerConnection)
        return try wrapSessionDescription(offer, callType: callType)
    }

    func applyRemoteOffer(_ sdp: String) async throws -> String {
        guard let peerConnection else { throw WebRTCCallSessionError.iceUnavailable }
        phase = .ringing
        let remote = try unwrapSessionDescription(sdp, expected: .offer)
        try await asyncSetRemoteDescription(remote, on: peerConnection)
        let answer = try await asyncAnswer(on: peerConnection)
        try await asyncSetLocalDescription(answer, on: peerConnection)
        return try wrapSessionDescription(answer, callType: callType)
    }

    func applyAnswer(_ sdp: String) async throws {
        guard let peerConnection else { throw WebRTCCallSessionError.iceUnavailable }
        let remote = try unwrapSessionDescription(sdp, expected: .answer)
        try await asyncSetRemoteDescription(remote, on: peerConnection)
        phase = .connecting
    }

    func addIceCandidate(_ data: Data) async {
        guard let peerConnection else { return }
        guard let payload = try? ICECandidatePayload.decode(from: data) else { return }
        let candidate = RTCIceCandidate(
            sdp: payload.candidate,
            sdpMLineIndex: Int32(payload.sdpMLineIndex ?? 0),
            sdpMid: payload.sdpMid
        )
        await asyncAdd(candidate, on: peerConnection)
    }

    func setMuted(_ muted: Bool) {
        localAudioTrack?.isEnabled = !muted
    }

    func setVideoEnabled(_ enabled: Bool) {
        if let track = localVideoTrack as? RTCVideoTrack {
            track.isEnabled = enabled
        }
        if enabled, capturer == nil {
            startCameraIfNeeded()
        }
    }

    func setSpeakerEnabled(_ enabled: Bool) {
        WebRTCAudioSession.setSpeakerEnabled(enabled)
    }

    private(set) var isScreenSharing = false

    func startScreenShare() async throws {
        guard callType == .video else {
            isScreenSharing = true
            return
        }
        capturer?.stopCapture()
        capturer = nil
        if videoSource == nil, let peerConnection {
            let source = Self.factory.videoSource()
            videoSource = source
            let videoTrack = Self.factory.videoTrack(with: source, trackId: "echo-screenshare0")
            localVideoTrack = videoTrack
            peerConnection.add(videoTrack, streamIds: ["echo-stream0"])
        }
        guard let videoSource else { throw ScreenShareError.unavailable }
        let share = ScreenShareCapturer(videoSource: videoSource)
        screenCapturer = share
        try await share.start()
        if let track = localVideoTrack as? RTCVideoTrack {
            track.isEnabled = true
        }
        isScreenSharing = true
    }

    func stopScreenShare() {
        screenCapturer?.stop()
        screenCapturer = nil
        isScreenSharing = false
        if callType == .video {
            startCameraIfNeeded()
        }
    }

    func hangup() {
        phase = .ended
        capturer?.stopCapture()
        capturer = nil
        screenCapturer?.stop()
        screenCapturer = nil
        peerConnection?.close()
        peerConnection = nil
        localAudioTrack = nil
        localVideoTrack = nil
        remoteVideoTrack = nil
        onIceCandidate = nil
        onConnected = nil
        WebRTCAudioSession.deactivate()
    }

    private func makePeerConnection(iceServers: [CallICEServer]) -> RTCPeerConnection? {
        let config = RTCConfiguration()
        config.iceServers = iceServers.map { server in
            RTCIceServer(
                urlStrings: server.urls,
                username: server.username,
                credential: server.credential
            )
        }
        config.sdpSemantics = .unifiedPlan
        config.continualGatheringPolicy = .gatherContinually
        return Self.factory.peerConnection(with: config, constraints: mediaConstraints, delegate: self)
    }

    private func attachLocalTracks() {
        guard let peerConnection else { return }
        let audioSource = Self.factory.audioSource(with: mediaConstraints)
        let audioTrack = Self.factory.audioTrack(with: audioSource, trackId: "echo-audio0")
        localAudioTrack = audioTrack
        peerConnection.add(audioTrack, streamIds: ["echo-stream0"])

        if callType == .video {
            let source = Self.factory.videoSource()
            videoSource = source
            let videoTrack = Self.factory.videoTrack(with: source, trackId: "echo-video0")
            localVideoTrack = videoTrack
            peerConnection.add(videoTrack, streamIds: ["echo-stream0"])
            startCameraIfNeeded()
        }
    }

    private func startCameraIfNeeded() {
        guard callType == .video,
              capturer == nil,
              let videoSource,
              let device = RTCCameraVideoCapturer.captureDevices().first(where: { $0.position == .front })
                ?? RTCCameraVideoCapturer.captureDevices().first,
              let format = RTCCameraVideoCapturer.supportedFormats(for: device).last,
              let fps = format.videoSupportedFrameRateRanges.first?.maxFrameRate else { return }

        let capturer = RTCCameraVideoCapturer(delegate: videoSource)
        self.capturer = capturer
        capturer.startCapture(with: device, format: format, fps: Int(fps))
    }

    private func wrapSessionDescription(_ desc: RTCSessionDescription, callType: CallType) throws -> String {
        let envelope = WebRTCSessionDescription(
            type: desc.type == .offer ? "offer" : "answer",
            sdp: desc.sdp,
            callType: callType.rawValue
        )
        return try envelope.encoded()
    }

    private func unwrapSessionDescription(_ text: String, expected: RTCSdpType) throws -> RTCSessionDescription {
        let envelope = try WebRTCSessionDescription.decode(from: text)
        let type: RTCSdpType = envelope.type == "offer" ? .offer : .answer
        guard type == expected else { throw WebRTCCallSessionError.invalidSDP }
        return RTCSessionDescription(type: type, sdp: envelope.sdp)
    }

    private func markConnectedIfNeeded() {
        guard phase != .connected else { return }
        phase = .connected
        onConnected?()
    }

    private func asyncOffer(on peerConnection: RTCPeerConnection) async throws -> RTCSessionDescription {
        try await withCheckedThrowingContinuation { continuation in
            peerConnection.offer(for: mediaConstraints) { sdp, error in
                if let error {
                    continuation.resume(throwing: error)
                } else if let sdp {
                    continuation.resume(returning: sdp)
                } else {
                    continuation.resume(throwing: WebRTCCallSessionError.iceUnavailable)
                }
            }
        }
    }

    private func asyncAnswer(on peerConnection: RTCPeerConnection) async throws -> RTCSessionDescription {
        try await withCheckedThrowingContinuation { continuation in
            peerConnection.answer(for: mediaConstraints) { sdp, error in
                if let error {
                    continuation.resume(throwing: error)
                } else if let sdp {
                    continuation.resume(returning: sdp)
                } else {
                    continuation.resume(throwing: WebRTCCallSessionError.iceUnavailable)
                }
            }
        }
    }

    private func asyncSetLocalDescription(
        _ description: RTCSessionDescription,
        on peerConnection: RTCPeerConnection
    ) async throws {
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            peerConnection.setLocalDescription(description) { error in
                if let error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume()
                }
            }
        }
    }

    private func asyncSetRemoteDescription(
        _ description: RTCSessionDescription,
        on peerConnection: RTCPeerConnection
    ) async throws {
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            peerConnection.setRemoteDescription(description) { error in
                if let error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume()
                }
            }
        }
    }

    private func asyncAdd(_ candidate: RTCIceCandidate, on peerConnection: RTCPeerConnection) async {
        await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
            peerConnection.add(candidate) { _ in
                continuation.resume()
            }
        }
    }
}

extension WebRTCLiveCallEngine: RTCPeerConnectionDelegate {
    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didChange stateChanged: RTCSignalingState) {}

    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didAdd stream: RTCMediaStream) {
        Task { @MainActor in
            if let track = stream.videoTracks.first {
                self.remoteVideoTrack = track
            }
        }
    }

    nonisolated func peerConnection(
        _ peerConnection: RTCPeerConnection,
        didStartReceivingOn transceiver: RTCRtpTransceiver
    ) {
        Task { @MainActor in
            if let track = transceiver.receiver.track as? RTCVideoTrack {
                self.remoteVideoTrack = track
            }
        }
    }

    nonisolated func peerConnection(
        _ peerConnection: RTCPeerConnection,
        didGenerate candidate: RTCIceCandidate
    ) {
        Task { @MainActor in
            let payload = ICECandidatePayload(
                candidate: candidate.sdp,
                sdpMid: candidate.sdpMid,
                sdpMLineIndex: Int(candidate.sdpMLineIndex)
            )
            if let data = try? payload.encoded() {
                self.onIceCandidate?(data)
            }
        }
    }

    nonisolated func peerConnection(
        _ peerConnection: RTCPeerConnection,
        didChange newState: RTCIceConnectionState
    ) {
        Task { @MainActor in
            switch newState {
            case .connected, .completed:
                self.markConnectedIfNeeded()
            case .failed, .disconnected, .closed:
                if self.phase != .ended {
                    self.phase = .ended
                }
            default:
                break
            }
        }
    }

    nonisolated func peerConnectionShouldNegotiate(_ peerConnection: RTCPeerConnection) {}
    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didRemove stream: RTCMediaStream) {}
    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didChange newState: RTCIceGatheringState) {}
    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didRemove candidates: [RTCIceCandidate]) {}
    nonisolated func peerConnection(_ peerConnection: RTCPeerConnection, didOpen dataChannel: RTCDataChannel) {}
}
#endif
#endif
