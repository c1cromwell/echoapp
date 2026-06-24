#if os(iOS)
import AVFoundation
import Foundation

/// WO-194 voice note capture (AAC in .m4a; see VoiceNoteCodec).
@MainActor
final class VoiceNoteRecorder: NSObject, ObservableObject {
    @Published private(set) var isRecording = false
    @Published private(set) var elapsed: TimeInterval = 0
    @Published private(set) var waveformSamples: [CGFloat] = []

    private var recorder: AVAudioRecorder?
    private var timer: Timer?
    private var outputURL: URL?
    private let maxDuration: TimeInterval = 300

    func startRecording() throws {
        let session = AVAudioSession.sharedInstance()
        try session.setCategory(.playAndRecord, mode: .default, options: [.defaultToSpeaker])
        try session.setActive(true)

        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("voice-\(UUID().uuidString).m4a")
        let recorder = try AVAudioRecorder(url: url, settings: VoiceNoteCodec.recorderSettings)
        recorder.delegate = self
        recorder.isMeteringEnabled = true
        guard recorder.record() else {
            throw VoiceNoteRecorderError.startFailed
        }
        self.recorder = recorder
        self.outputURL = url
        isRecording = true
        elapsed = 0
        waveformSamples = []
        timer = Timer.scheduledTimer(withTimeInterval: 0.25, repeats: true) { [weak self] _ in
            Task { @MainActor in
                guard let self, self.isRecording else { return }
                self.elapsed += 0.25
                self.recorder?.updateMeters()
                let power = self.recorder?.averagePower(forChannel: 0) ?? -160
                let normalized = CGFloat(max(0, min(1, (power + 50) / 50)))
                self.waveformSamples.append(normalized)
                if self.elapsed >= self.maxDuration {
                    _ = try? self.stopRecording()
                }
            }
        }
    }

    func stopRecording() throws -> Data {
        timer?.invalidate()
        timer = nil
        recorder?.stop()
        isRecording = false
        guard let url = outputURL else { throw VoiceNoteRecorderError.noRecording }
        defer {
            try? FileManager.default.removeItem(at: url)
            outputURL = nil
            recorder = nil
        }
        return try Data(contentsOf: url)
    }

    func cancelRecording() {
        timer?.invalidate()
        timer = nil
        recorder?.stop()
        isRecording = false
        if let url = outputURL {
            try? FileManager.default.removeItem(at: url)
        }
        outputURL = nil
        recorder = nil
        elapsed = 0
    }
}

extension VoiceNoteRecorder: AVAudioRecorderDelegate {}

enum VoiceNoteRecorderError: LocalizedError {
    case startFailed
    case noRecording

    var errorDescription: String? {
        switch self {
        case .startFailed: return "Could not start voice recording."
        case .noRecording: return "No voice recording available."
        }
    }
}
#endif
