#if os(iOS)
import SwiftUI

#if canImport(WebRTC)
import WebRTC

/// Renders a remote or local `RTCVideoTrack` inside SwiftUI (M4c).
struct WebRTCVideoView: UIViewRepresentable {
    let track: Any?

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeUIView(context: Context) -> RTCMTLVideoView {
        let view = RTCMTLVideoView(frame: .zero)
        view.videoContentMode = .scaleAspectFill
        context.coordinator.bind(track: track as? RTCVideoTrack, to: view)
        return view
    }

    func updateUIView(_ uiView: RTCMTLVideoView, context: Context) {
        context.coordinator.bind(track: track as? RTCVideoTrack, to: uiView)
    }

    final class Coordinator {
        private weak var boundView: RTCMTLVideoView?
        private var boundTrack: RTCVideoTrack?

        func bind(track: RTCVideoTrack?, to view: RTCMTLVideoView) {
            if boundView === view, boundTrack === track { return }
            boundTrack?.remove(view)
            boundView = view
            boundTrack = track
            track?.add(view)
        }
    }
}
#else
struct WebRTCVideoView: View {
    let track: Any?

    var body: some View {
        Color.Echo.deepNavy
    }
}
#endif
#endif
