// Core/DesignSystem/Components/EchoLogo.swift
// Concentric ripple circle logo — represents encrypted signal propagation

import SwiftUI

struct EchoLogo: View {
    let size: CGFloat
    
    init(size: CGFloat = 32) {
        self.size = size
    }
    
    var body: some View {
        Canvas { context, canvasSize in
            let center = CGPoint(x: canvasSize.width / 2, y: canvasSize.height / 2)
            let scale = size / 200 // Original viewBox is 200x200
            
            // Core dot
            let corePath = Path(ellipseIn: CGRect(
                x: center.x - 7 * scale, y: center.y - 7 * scale,
                width: 14 * scale, height: 14 * scale
            ))
            context.fill(corePath, with: .color(Color.Echo.primaryContainer))
            
            // Concentric rings (decreasing opacity)
            let rings: [(radius: CGFloat, width: CGFloat, opacity: Double)] = [
                (14, 2.0, 0.8),
                (22, 1.8, 0.6),
                (31, 1.5, 0.4),
            ]
            
            for ring in rings {
                var ringPath = Path()
                ringPath.addEllipse(in: CGRect(
                    x: center.x - ring.radius * scale,
                    y: center.y - ring.radius * scale,
                    width: ring.radius * 2 * scale,
                    height: ring.radius * 2 * scale
                ))
                context.stroke(
                    ringPath,
                    with: .color(Color.Echo.primaryContainer.opacity(ring.opacity)),
                    lineWidth: ring.width * scale
                )
            }
            
            // Signal arcs (top and bottom pairs at 3 distances)
            let arcSets: [(radius: CGFloat, width: CGFloat, opacity: Double, color: Color)] = [
                (48, 3.0, 1.0, Color.Echo.primaryContainer),
                (68, 2.5, 0.8, Color.Echo.skyLight),
                (93, 2.5, 0.55, Color.Echo.primaryContainer),
            ]
            
            for arc in arcSets {
                // Top arc
                var topArc = Path()
                topArc.addArc(
                    center: center,
                    radius: arc.radius * scale,
                    startAngle: .degrees(225),
                    endAngle: .degrees(315),
                    clockwise: false
                )
                context.stroke(
                    topArc,
                    with: .color(arc.color.opacity(arc.opacity)),
                    style: StrokeStyle(lineWidth: arc.width * scale, lineCap: .round)
                )
                
                // Bottom arc
                var bottomArc = Path()
                bottomArc.addArc(
                    center: center,
                    radius: arc.radius * scale,
                    startAngle: .degrees(45),
                    endAngle: .degrees(135),
                    clockwise: false
                )
                context.stroke(
                    bottomArc,
                    with: .color(arc.color.opacity(arc.opacity)),
                    style: StrokeStyle(lineWidth: arc.width * scale, lineCap: .round)
                )
            }
            
            // Side arcs (left and right, subtle)
            for startAngle in [135.0, 315.0] {
                var sideArc = Path()
                sideArc.addArc(
                    center: center,
                    radius: 62 * scale,
                    startAngle: .degrees(startAngle),
                    endAngle: .degrees(startAngle + 90),
                    clockwise: false
                )
                context.stroke(
                    sideArc,
                    with: .color(Color.Echo.primaryContainer.opacity(0.3)),
                    style: StrokeStyle(lineWidth: 2.0 * scale, lineCap: .round)
                )
            }
        }
        .frame(width: size, height: size)
    }
}

#Preview {
    HStack(spacing: 8) {
        EchoLogo(size: 32)
        Text("ECHO")
            .font(.custom("Inter", size: 24))
            .fontWeight(.bold)
            .foregroundStyle(Color.Echo.primaryContainer)
    }
    .padding()
}
