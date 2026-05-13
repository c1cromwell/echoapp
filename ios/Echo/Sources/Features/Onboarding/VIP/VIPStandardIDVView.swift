#if os(iOS)
// Features/Onboarding/VIP/VIPStandardIDVView.swift
// Standard identity verification flow for VIP badge (T2 Verified).
// Two steps chained in sequence:
//   1. Document scan (Phase 1: photo capture stub — Stripe Identity connects in Phase 2)
//   2. Recovery phone number (SMSOTPSetupView — real backend call)
// On both complete → onComplete(trustTier: 2)

import SwiftUI
import AVFoundation

public struct VIPStandardIDVView: View {
    let did: String
    let onComplete: (Int) -> Void

    @State private var step: Step = .scan
    @State private var scanPhase: ScanPhase = .idle

    public init(did: String, onComplete: @escaping (Int) -> Void) {
        self.did = did
        self.onComplete = onComplete
    }

    enum Step { case scan, phone }
    enum ScanPhase { case idle, capturing, processing, done }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                progressBar
                    .padding(.horizontal, 24)
                    .padding(.top, 20)

                switch step {
                case .scan:
                    documentScanStep
                case .phone:
                    SMSOTPSetupView(did: did) {
                        onComplete(2)
                    }
                }
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    // MARK: - Progress bar

    private var progressBar: some View {
        HStack(spacing: 6) {
            ForEach(0..<2) { i in
                let active = (i == 0 && step == .scan) || (i == 1 && step == .phone)
                let done   = (i == 0 && step != .scan)
                Capsule()
                    .fill(done ? Color.echoTrustGreen : (active ? Color.echoSignal : Color.echoHair))
                    .frame(height: 4)
                    .animation(.spring(response: 0.4, dampingFraction: 0.85), value: step)
            }
        }
        .padding(.bottom, 8)
    }

    // MARK: - Document Scan Step

    private var documentScanStep: some View {
        VStack(spacing: 0) {
            VStack(alignment: .leading, spacing: 6) {
                Text("Scan your ID")
                    .font(.system(size: 26, weight: .semibold))
                    .tracking(-0.6)
                    .foregroundStyle(Color.echoInk)
                Text("Take a photo of your driver's licence or passport. Your document image is processed securely and never stored on our servers.")
                    .font(.system(size: 13))
                    .lineSpacing(3)
                    .foregroundStyle(Color.echoInk55)
            }
            .padding(.horizontal, 24)
            .padding(.top, 20)

            // Document frame
            documentFrame
                .padding(.horizontal, 24)
                .padding(.top, 24)

            // Instructions
            VStack(alignment: .leading, spacing: 10) {
                instructionRow(icon: "sun.max", text: "Good lighting — avoid glare or shadows")
                instructionRow(icon: "doc.plaintext", text: "All four corners visible, text readable")
                instructionRow(icon: "hand.raised.slash", text: "Remove from wallet sleeve or holder")
            }
            .padding(.horizontal, 24)
            .padding(.top, 20)

            Spacer(minLength: 12)

            // CTA
            VStack(spacing: 10) {
                switch scanPhase {
                case .idle:
                    Button {
                        simulateCapture()
                    } label: {
                        HStack {
                            Image(systemName: "camera.fill")
                                .font(.system(size: 15))
                            Text("Capture Document")
                                .font(.system(size: 15, weight: .semibold))
                        }
                        .foregroundStyle(Color.echoPaper)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 15)
                        .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                    }
                    .buttonStyle(SpringPressStyle())

                case .capturing:
                    HStack(spacing: 10) {
                        ProgressView().tint(Color.echoSignal)
                        Text("Capturing…").font(.system(size: 15)).foregroundStyle(Color.echoInk55)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 15)

                case .processing:
                    HStack(spacing: 10) {
                        ProgressView().tint(Color.echoSignal)
                        Text("Verifying document…").font(.system(size: 15)).foregroundStyle(Color.echoInk55)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 15)

                case .done:
                    HStack(spacing: 8) {
                        Image(systemName: "checkmark.circle.fill").foregroundStyle(Color.echoTrustGreen)
                        Text("Document captured").font(.system(size: 15, weight: .semibold)).foregroundStyle(Color.echoTrustGreen)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 15)
                }
            }
            .padding(.horizontal, 24)
            .padding(.bottom, 28)
        }
    }

    private var documentFrame: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 12)
                .fill(Color.echoPaperDim)
                .frame(height: 200)

            if scanPhase == .done {
                VStack(spacing: 8) {
                    Image(systemName: "checkmark.seal.fill")
                        .font(.system(size: 40))
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("Document verified")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.echoTrustGreen)
                }
            } else {
                VStack(spacing: 8) {
                    Image(systemName: "creditcard.viewfinder")
                        .font(.system(size: 40))
                        .foregroundStyle(Color.echoInk40)
                    Text("Position your ID here")
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoInk40)
                }
            }

            // Corner guides
            if scanPhase == .idle {
                CornerGuides()
            }
        }
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(
                    scanPhase == .done ? Color.echoTrustGreen.opacity(0.5) : Color.echoHair,
                    lineWidth: scanPhase == .done ? 1.5 : 1
                )
        )
    }

    private func instructionRow(icon: String, text: String) -> some View {
        HStack(spacing: 10) {
            Image(systemName: icon)
                .font(.system(size: 13))
                .foregroundStyle(Color.echoSignal)
                .frame(width: 20)
            Text(text)
                .font(.system(size: 13))
                .foregroundStyle(Color.echoInk55)
        }
    }

    // MARK: - Scan simulation (Phase 1)

    private func simulateCapture() {
        withAnimation { scanPhase = .capturing }
        Task {
            try? await Task.sleep(nanoseconds: 800_000_000)
            withAnimation { scanPhase = .processing }
            try? await Task.sleep(nanoseconds: 1_200_000_000)
            withAnimation { scanPhase = .done }
            try? await Task.sleep(nanoseconds: 700_000_000)
            withAnimation { step = .phone }
        }
    }
}

// MARK: - Corner guides overlay

private struct CornerGuides: View {
    var body: some View {
        GeometryReader { geo in
            let w = geo.size.width, h = geo.size.height
            let len: CGFloat = 20, thick: CGFloat = 2.5
            let corner: CGFloat = 12
            let color = Color.echoSignal.opacity(0.6)
            ZStack {
                // Top-left
                Path { p in
                    p.move(to: CGPoint(x: corner, y: thick/2))
                    p.addLine(to: CGPoint(x: corner + len, y: thick/2))
                }.stroke(color, lineWidth: thick)
                Path { p in
                    p.move(to: CGPoint(x: thick/2, y: corner))
                    p.addLine(to: CGPoint(x: thick/2, y: corner + len))
                }.stroke(color, lineWidth: thick)
                // Top-right
                Path { p in
                    p.move(to: CGPoint(x: w - corner, y: thick/2))
                    p.addLine(to: CGPoint(x: w - corner - len, y: thick/2))
                }.stroke(color, lineWidth: thick)
                Path { p in
                    p.move(to: CGPoint(x: w - thick/2, y: corner))
                    p.addLine(to: CGPoint(x: w - thick/2, y: corner + len))
                }.stroke(color, lineWidth: thick)
                // Bottom-left
                Path { p in
                    p.move(to: CGPoint(x: corner, y: h - thick/2))
                    p.addLine(to: CGPoint(x: corner + len, y: h - thick/2))
                }.stroke(color, lineWidth: thick)
                Path { p in
                    p.move(to: CGPoint(x: thick/2, y: h - corner))
                    p.addLine(to: CGPoint(x: thick/2, y: h - corner - len))
                }.stroke(color, lineWidth: thick)
                // Bottom-right
                Path { p in
                    p.move(to: CGPoint(x: w - corner, y: h - thick/2))
                    p.addLine(to: CGPoint(x: w - corner - len, y: h - thick/2))
                }.stroke(color, lineWidth: thick)
                Path { p in
                    p.move(to: CGPoint(x: w - thick/2, y: h - corner))
                    p.addLine(to: CGPoint(x: w - thick/2, y: h - corner - len))
                }.stroke(color, lineWidth: thick)
            }
        }
    }
}

#Preview {
    VIPStandardIDVView(did: "did:key:zDemo") { _ in }
}
#endif
