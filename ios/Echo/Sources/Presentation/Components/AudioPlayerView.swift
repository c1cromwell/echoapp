#if os(iOS)
import SwiftUI
import AVFoundation

@MainActor
final class AudioPlayerManager: ObservableObject {
    @Published var isPlaying = false
    @Published var progress: Double = 0
    @Published var duration: TimeInterval = 0
    @Published var playbackRate: Float = 1.0
    @Published private(set) var isLoaded = false

    private var player: AVAudioPlayer?
    private var timer: Timer?

    func load(data: Data) {
        stop()
        player = try? AVAudioPlayer(data: data)
        player?.prepareToPlay()
        player?.enableRate = true
        player?.rate = playbackRate
        duration = player?.duration ?? 0
        isLoaded = player != nil
    }

    func togglePlayback() {
        guard let player else { return }
        if player.isPlaying {
            player.pause()
            isPlaying = false
            timer?.invalidate()
        } else {
            try? AVAudioSession.sharedInstance().setCategory(.playback)
            try? AVAudioSession.sharedInstance().setActive(true)
            player.rate = playbackRate
            player.play()
            isPlaying = true
            startTimer()
        }
    }

    func cyclePlaybackRate() {
        switch playbackRate {
        case 0.75: playbackRate = 1.0
        case 1.0: playbackRate = 1.25
        case 1.25: playbackRate = 1.5
        default: playbackRate = 0.75
        }
        player?.enableRate = true
        if isPlaying {
            player?.rate = playbackRate
        }
    }

    func stop() {
        timer?.invalidate()
        timer = nil
        player?.stop()
        player = nil
        isPlaying = false
        progress = 0
    }

    func seek(to fraction: Double) {
        guard let player else { return }
        let clamped = max(0, min(fraction, 1))
        player.currentTime = player.duration * clamped
        progress = clamped
    }

    private func startTimer() {
        timer?.invalidate()
        timer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { [weak self] _ in
            Task { @MainActor in
                guard let self, let player = self.player else { return }
                if !player.isPlaying {
                    self.isPlaying = false
                    self.timer?.invalidate()
                    return
                }
                self.progress = player.duration > 0 ? player.currentTime / player.duration : 0
            }
        }
    }
}

struct AudioPlayerView: View {
    let waveformSamples: [CGFloat]
    @StateObject private var player = AudioPlayerManager()
    let audioData: Data?

    var body: some View {
        HStack(spacing: 10) {
            Button {
                if !player.isLoaded, let audioData {
                    player.load(data: audioData)
                }
                player.togglePlayback()
            } label: {
                Image(systemName: player.isPlaying ? "pause.circle.fill" : "play.circle.fill")
                    .font(.system(size: 32))
                    .foregroundStyle(Color.echoSignal)
            }
            .buttonStyle(.plain)

            GeometryReader { geo in
                WaveformView(
                    samples: waveformSamples,
                    progress: player.progress,
                    accentColor: .echoSignal,
                    inactiveColor: .echoInk40.opacity(0.5)
                )
                .frame(height: 30)
                .contentShape(Rectangle())
                .gesture(
                    DragGesture(minimumDistance: 0)
                        .onChanged { value in
                            let width = max(geo.size.width, 1)
                            let fraction = Double(value.location.x / width)
                            player.seek(to: fraction)
                        }
                )
            }
            .frame(height: 30)

            Button {
                player.cyclePlaybackRate()
            } label: {
                Text(rateLabel)
                    .font(.system(size: 11, weight: .semibold).monospacedDigit())
                    .foregroundStyle(Color.echoInk55)
                    .frame(minWidth: 34)
            }
            .buttonStyle(.plain)

            Text(formattedTime)
                .font(.system(size: 12, weight: .medium).monospacedDigit())
                .foregroundStyle(Color.echoInk55)
        }
    }

    private var rateLabel: String {
        switch player.playbackRate {
        case 0.75: return "0.75×"
        case 1.25: return "1.25×"
        case 1.5: return "1.5×"
        default: return "1×"
        }
    }

    private var formattedTime: String {
        let current = Int(player.duration * player.progress)
        let total = Int(player.duration)
        return "\(formatSeconds(current))/\(formatSeconds(total))"
    }

    private func formatSeconds(_ total: Int) -> String {
        let minutes = total / 60
        let seconds = total % 60
        return String(format: "%d:%02d", minutes, seconds)
    }
}
#endif
