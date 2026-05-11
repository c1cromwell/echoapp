#if os(iOS)
import SwiftUI

// Design review finding: the 3-slide carousel auto-advances past content the
// user hasn't read, and the logo lockup appears twice on screen at once.
//
// Proposed: one screen, no carousel, no dots, no timer.
// The pitch IS the layout. Three privacy facts inline as anchored content.

struct EchoWelcomeView: View {
    let onSetUp: () -> Void
    let onAlreadyHaveAccount: () -> Void

    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(alignment: .leading, spacing: 0) {
                // Mark — top-left, no centred logo stacking with the h1
                EchoRippleMark(size: 32)
                    .padding(.top, 8)

                Spacer()

                // Headline
                VStack(alignment: .leading, spacing: 20) {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("ECHO")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoInk40)

                        Text("Messages that\nbelong to you.")
                            .font(.system(size: 34, weight: .semibold))
                            .tracking(-1.0)
                            .lineSpacing(2)
                            .foregroundStyle(Color.echoInk)
                    }

                    Text("No phone number. No email. No password.\nYour face on this device is your only key.")
                        .font(.system(size: 15))
                        .lineSpacing(4)
                        .foregroundStyle(Color.echoInk70)
                }

                // Privacy facts — inline, not slides
                VStack(alignment: .leading, spacing: 12) {
                    ForEach(privacyFacts, id: \.title) { fact in
                        HStack(alignment: .firstTextBaseline, spacing: 12) {
                            Circle()
                                .fill(Color.echoSignal)
                                .frame(width: 5, height: 5)
                                .offset(y: -2)

                            VStack(alignment: .leading, spacing: 0) {
                                Text(fact.title)
                                    .font(.system(size: 13, weight: .semibold))
                                    .foregroundStyle(Color.echoInk)
                                + Text("  ·  ")
                                    .font(.system(size: 13))
                                    .foregroundStyle(Color.echoInk55)
                                + Text(fact.subtitle)
                                    .font(.system(size: 13))
                                    .foregroundStyle(Color.echoInk55)
                            }
                        }
                    }
                }
                .padding(.vertical, 18)
                .padding(.horizontal, 0)
                .overlay(alignment: .top)    { Divider().foregroundStyle(Color.echoHair) }
                .overlay(alignment: .bottom) { Divider().foregroundStyle(Color.echoHair) }
                .padding(.top, 28)

                // CTAs
                VStack(spacing: 10) {
                    Button(action: onSetUp) {
                        HStack {
                            Text("Set up Echo")
                                .font(.system(size: 15, weight: .semibold))
                                .tracking(-0.2)
                            Spacer()
                            Text("→")
                                .font(.system(size: 18))
                        }
                        .foregroundStyle(Color.echoPaper)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 15)
                        .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                    }

                    Button("I already have an account", action: onAlreadyHaveAccount)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.echoInk70)
                        .padding(.vertical, 6)
                }
                .padding(.top, 22)
            }
            .padding(.horizontal, 28)
            .padding(.vertical, 24)
        }
        .preferredColorScheme(.light)
    }

    private let privacyFacts: [(title: String, subtitle: String)] = [
        ("End-to-end encrypted", "Not even Echo can read your messages."),
        ("Zero metadata",        "No phone book uploads. No graph."),
        ("Keys on this device",  "Hardware-bound. Never leave."),
    ]
}

// MARK: - Concentric ripple mark (also defined in DesignSystem/EchoMark.swift)
// This duplicate is kept for the Preview; the canonical definition is in Components.

private struct EchoRippleMarkLocal: View {
    var size: CGFloat = 24
    var color: Color = .echoSignal

    private let ringRadii: [CGFloat] = [6, 11, 17, 24, 34]

    var body: some View {
        Canvas { ctx, sz in
            let cx = sz.width / 2
            let cy = sz.height / 2
            let scale = sz.width / 80

            // Centre dot
            ctx.fill(Path(ellipseIn: CGRect(x: cx - 3.5*scale, y: cy - 3.5*scale,
                                            width: 7*scale, height: 7*scale)),
                     with: .color(color))
            // Rings
            for (i, r) in ringRadii.enumerated() {
                let sr = r * scale
                let opacity = 0.95 - Double(i) * 0.18
                var ring = Path()
                ring.addEllipse(in: CGRect(x: cx-sr, y: cy-sr, width: sr*2, height: sr*2))
                ctx.stroke(ring, with: .color(color.opacity(opacity)), lineWidth: 1.6*scale)
            }
        }
        .frame(width: size, height: size)
    }
}

#Preview {
    EchoWelcomeView(onSetUp: {}, onAlreadyHaveAccount: {})
}
#endif
