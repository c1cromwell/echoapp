#if os(iOS)
import SwiftUI
import AVFoundation

@MainActor
final class AudioPlayerManager: ObservableObject {
    @Published var isPlaying = false
    @Published var progress: Double = 0
    @Published var duration: TimeInterval = 0
    @Published private(set) var isLoaded = false

    private var player: AVAudioPlayer?
    private var timer: Timer?

    func load(data: Data) {
        stop()
        player = try? AVAudioPlayer(data: data)
        player?.prepareToPlay()
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
            player.play()
            isPlaying = true
            startTimer()
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
        player.currentTime = player.duration * max(0, min(fraction, 1))
        progress = fraction
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
                        let fraction = value.location.x / max(value.startLocation.x + 1, 1)
                        player.seek(to: Double(fraction))
                    }
            )

            Text(formattedDuration)
                .font(.system(size: 12, weight: .medium).monospacedDigit())
                .foregroundStyle(Color.echoInk55)
        }
    }

    private var formattedDuration: String {
        let total = Int(player.duration)
        let minutes = total / 60
        let seconds = total % 60
        return String(format: "%d:%02d", minutes, seconds)
    }
}
#endif
