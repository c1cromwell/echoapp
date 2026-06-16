#if os(iOS)
import AVFoundation
import Foundation

/// Configures AVAudioSession for voice/video calls (M4c).
enum WebRTCAudioSession {
    static func activate(callType: CallType) throws {
        let session = AVAudioSession.sharedInstance()
        let mode: AVAudioSession.Mode = callType == .video ? .videoChat : .voiceChat
        try session.setCategory(.playAndRecord, mode: mode, options: [.allowBluetooth, .defaultToSpeaker])
        try session.setActive(true)
    }

    static func deactivate() {
        try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
    }

    static func setSpeakerEnabled(_ enabled: Bool) {
        try? AVAudioSession.sharedInstance().overrideOutputAudioPort(enabled ? .speaker : .none)
    }
}
#endif
