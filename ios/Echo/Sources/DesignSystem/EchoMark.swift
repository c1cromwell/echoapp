#if os(iOS)
import SwiftUI

/// The Echo Ripple Pulse Mark — concentric rings + asymmetric ripple arcs.
/// Matches docs/logos/echo-ripple-pulse-concepts.html.
public struct EchoRippleMark: View {
    public var size: CGFloat
    public var color: Color

    public init(size: CGFloat = 24, color: Color = .echoSignal) {
        self.size = size
        self.color = color
    }

    public var body: some View {
        Canvas { ctx, sz in
            let center = CGPoint(x: sz.width / 2, y: sz.height / 2)
            let scale = sz.width / 200

            // Core filled dot
            ctx.fill(
                Path(ellipseIn: CGRect(x: center.x - 7*scale, y: center.y - 7*scale,
                                       width: 14*scale, height: 14*scale)),
                with: .color(color)
            )

            // Concentric rings
            let rings: [(radius: CGFloat, width: CGFloat, opacity: Double)] = [
                (14, 2.0, 0.8),
                (22, 1.8, 0.6),
                (31, 1.5, 0.4),
            ]
            for ring in rings {
                var path = Path()
                let r = ring.radius * scale
                path.addEllipse(in: CGRect(x: center.x - r, y: center.y - r, width: r*2, height: r*2))
                ctx.stroke(path, with: .color(color.opacity(ring.opacity)),
                           lineWidth: ring.width * scale)
            }

            // Ripple arc pairs (top + bottom) at 3 distances
            let arcs: [(radius: CGFloat, width: CGFloat, opacity: Double, lighter: Bool)] = [
                (48, 3.0, 1.0, false),
                (68, 2.5, 0.8, true),
                (93, 2.5, 0.55, false),
            ]
            for arc in arcs {
                let arcColor = arc.lighter ? color.opacity(0.5) : color
                // Top arc
                var top = Path()
                top.addArc(center: center, radius: arc.radius * scale,
                           startAngle: .degrees(225), endAngle: .degrees(315), clockwise: false)
                ctx.stroke(top, with: .color(arcColor.opacity(arc.opacity)),
                           style: StrokeStyle(lineWidth: arc.width * scale, lineCap: .round))
                // Bottom arc
                var bottom = Path()
                bottom.addArc(center: center, radius: arc.radius * scale,
                              startAngle: .degrees(45), endAngle: .degrees(135), clockwise: false)
                ctx.stroke(bottom, with: .color(arcColor.opacity(arc.opacity)),
                           style: StrokeStyle(lineWidth: arc.width * scale, lineCap: .round))
            }

            // Side whisper arcs (left + right)
            for startAngle in [135.0, 315.0] {
                var side = Path()
                side.addArc(center: center, radius: 62 * scale,
                            startAngle: .degrees(startAngle), endAngle: .degrees(startAngle + 90),
                            clockwise: false)
                ctx.stroke(side, with: .color(color.opacity(0.3)),
                           style: StrokeStyle(lineWidth: 2.0 * scale, lineCap: .round))
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
