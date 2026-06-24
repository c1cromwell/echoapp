#if os(iOS)
import AVFoundation
import Foundation

/// Voice-note encoding preferences (M5 / WO-194). iOS records AAC in an MPEG-4 container;
/// Opus would require a third-party encoder — wire format stays `audio/mp4` until linked.
enum VoiceNoteCodec {
    static let wireMimeType = "audio/mp4"
    static let wireCodecLabel = "aac"

    /// Voice-optimized mono AAC (smaller than default 44.1 kHz music settings).
    static var recorderSettings: [String: Any] {
        [
            AVFormatIDKey: Int(kAudioFormatMPEG4AAC),
            AVSampleRateKey: 24_000,
            AVNumberOfChannelsKey: 1,
            AVEncoderBitRateKey: 32_000,
            AVEncoderAudioQualityKey: AVAudioQuality.medium.rawValue,
        ]
    }

    static func normalizedWaveform(_ samples: [CGFloat], targetCount: Int = 40) -> [Float] {
        guard !samples.isEmpty else { return [] }
        if samples.count == targetCount {
            return samples.map { Float(min(1, max(0, $0))) }
        }
        var out = [Float]()
        out.reserveCapacity(targetCount)
        let step = Double(samples.count) / Double(targetCount)
        for i in 0..<targetCount {
            let idx = min(Int(Double(i) * step), samples.count - 1)
            out.append(Float(min(1, max(0, samples[idx]))))
        }
        return out
    }
}
#endif
