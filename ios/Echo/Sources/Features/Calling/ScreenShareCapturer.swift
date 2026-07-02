#if os(iOS)
#if canImport(WebRTC)
import AVFoundation
import Foundation
import ReplayKit
import WebRTC

/// Feeds ReplayKit screen samples into a WebRTC video source (WO-31).
@MainActor
final class ScreenShareCapturer: NSObject {
    private let videoSource: RTCVideoSource
    private let capturer: RTCVideoCapturer
    private var isRunning = false

    init(videoSource: RTCVideoSource) {
        self.videoSource = videoSource
        self.capturer = RTCVideoCapturer()
        super.init()
    }

    func start() async throws {
        guard !isRunning else { return }
        let recorder = RPScreenRecorder.shared()
        guard recorder.isAvailable else {
            throw ScreenShareError.unavailable
        }
        try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<Void, Error>) in
            recorder.startCapture(handler: { [weak self] sample, type, error in
                if let error {
                    return
                }
                guard type == .video, let self else { return }
                self.publish(sample: sample)
            }, completionHandler: { error in
                if let error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume()
                }
            })
        }
        isRunning = true
    }

    func stop() {
        guard isRunning else { return }
        RPScreenRecorder.shared().stopCapture { _ in }
        isRunning = false
    }

    private func publish(sample: CMSampleBuffer) {
        guard let pixelBuffer = CMSampleBufferGetImageBuffer(sample) else { return }
        let time = CMSampleBufferGetPresentationTimeStamp(sample)
        let frame = RTCVideoFrame(buffer: RTCCVPixelBuffer(pixelBuffer: pixelBuffer), rotation: ._0, timeStampNs: Int64(time.seconds * 1_000_000_000))
        videoSource.capturer(capturer, didCapture: frame)
    }
}

enum ScreenShareError: LocalizedError {
    case unavailable

    var errorDescription: String? {
        switch self {
        case .unavailable: return "Screen recording is not available on this device."
        }
    }
}
#endif
#endif
