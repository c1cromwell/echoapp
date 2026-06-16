#if os(iOS)
import SwiftUI

#if canImport(WebRTC)
import WebRTC

/// Renders a remote or local `RTCVideoTrack` inside SwiftUI (M4c).
struct WebRTCVideoView: UIViewRepresentable {
    let track: RTCVideoTrack?

    func makeUIView(context: Context) -> RTCMTLVideoView {
        let view = RTCMTLVideoView(frame: .zero)
        view.videoContentMode = .scaleAspectFill
        view.track = track
        return view
    }

    func updateUIView(_ uiView: RTCMTLVideoView, context: Context) {
        uiView.track = track
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
