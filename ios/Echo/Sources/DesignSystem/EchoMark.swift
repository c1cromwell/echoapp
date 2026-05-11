#if os(iOS)
import SwiftUI

/// The Echo concentric-ripple mark.
/// Used top-left on welcome and login screens — no competing logo stacking.
public struct EchoRippleMark: View {
    public var size: CGFloat
    public var color: Color

    public init(size: CGFloat = 24, color: Color = .echoSignal) {
        self.size = size
        self.color = color
    }

    private let ringRadii: [CGFloat] = [6, 11, 17, 24, 34]

    public var body: some View {
        Canvas { ctx, sz in
            let cx = sz.width / 2
            let cy = sz.height / 2
            let scale = sz.width / 80

            ctx.fill(
                Path(ellipseIn: CGRect(x: cx - 3.5*scale, y: cy - 3.5*scale,
                                      width: 7*scale, height: 7*scale)),
                with: .color(color)
            )
            for (i, r) in ringRadii.enumerated() {
                let sr = r * scale
                let opacity = 0.95 - Double(i) * 0.18
                var ring = Path()
                ring.addEllipse(in: CGRect(x: cx-sr, y: cy-sr, width: sr*2, height: sr*2))
                ctx.stroke(ring,
                           with: .color(color.opacity(opacity)),
                           style: StrokeStyle(lineWidth: 1.6 * scale))
            }
        }
        .frame(width: size, height: size)
    }
}

#Preview {
    HStack(spacing: 20) {
        EchoRippleMark(size: 20)
        EchoRippleMark(size: 32)
        EchoRippleMark(size: 48)
        EchoRippleMark(size: 32, color: .echoTrustGreen)
    }
    .padding()
    .background(Color.echoPaper)
    .preferredColorScheme(.light)
}
#endif
