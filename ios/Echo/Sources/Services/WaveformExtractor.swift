#if os(iOS)
import AVFoundation
import Foundation

/// Extracts playback waveform samples from audio data (M5b / WO-194).
enum WaveformExtractor {
    static func samples(from audioData: Data, targetCount: Int = 40) -> [CGFloat] {
        let tempURL = FileManager.default.temporaryDirectory
            .appendingPathComponent("wave-\(UUID().uuidString).m4a")
        defer { try? FileManager.default.removeItem(at: tempURL) }
        do {
            try audioData.write(to: tempURL)
            let file = try AVAudioFile(forReading: tempURL)
            guard let buffer = AVAudioPCMBuffer(
                pcmFormat: file.processingFormat,
                frameCapacity: AVAudioFrameCount(file.length)
            ) else { return fallback(targetCount) }
            try file.read(into: buffer)
            guard let channel = buffer.floatChannelData?.pointee else { return fallback(targetCount) }
            let frameCount = Int(buffer.frameLength)
            guard frameCount > 0 else { return fallback(targetCount) }
            let bucket = max(1, frameCount / targetCount)
            var out: [CGFloat] = []
            out.reserveCapacity(targetCount)
            var idx = 0
            while idx < frameCount && out.count < targetCount {
                let end = min(frameCount, idx + bucket)
                var peak: Float = 0
                for i in idx..<end {
                    peak = max(peak, abs(channel[i]))
                }
                out.append(CGFloat(min(1, peak * 4)))
                idx += bucket
            }
            return out.isEmpty ? fallback(targetCount) : out
        } catch {
            return fallback(targetCount)
        }
    }

    private static func fallback(_ count: Int) -> [CGFloat] {
        (0..<count).map { i in CGFloat(0.2 + 0.6 * sin(Double(i) * 0.4)) }
    }
}
#endif
