#if os(iOS)
// Features/Onboarding/VIP/VIPStandardIDVView.swift
// Standard identity verification flow for VIP badge (T2 Verified).
// Three steps chained in sequence:
//   1. Scan licence or passport + selfie  (IDVFallbackView)
//   2. Setup recovery phone number        (SMSOTPSetupView)
//   3. Both done → onComplete(trustTier: 2)

import SwiftUI

public struct VIPStandardIDVView: View {
    let did: String
    let onComplete: (Int) -> Void   // trustTier

    @State private var step: Step = .idv
    @State private var idvCoordinator = EnrollmentCoordinator(
        onComplete: { _ in }, onCancel: {}
    )

    public init(did: String, onComplete: @escaping (Int) -> Void) {
        self.did = did
        self.onComplete = onComplete
    }

    enum Step { case idv, phone, done }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                progressBar
                    .padding(.horizontal, 24)
                    .padding(.top, 20)
                    .padding(.bottom, 8)

                switch step {
                case .idv:
                    IDVFallbackView(coordinator: idvCoordinator)
                        .onAppear {
                            idvCoordinator = EnrollmentCoordinator(
                                onComplete: { _ in
                                    withAnimation { step = .phone }
                                },
                                onCancel: {}
                            )
                        }
                case .phone:
                    SMSOTPSetupView(did: did) {
                        withAnimation { step = .done }
                        onComplete(2)
                    }
                case .done:
                    Color.clear
                }
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    private var progressBar: some View {
        HStack(spacing: 6) {
            ForEach(0..<2) { i in
                let active = (i == 0 && step == .idv) || (i == 1 && step == .phone)
                let done   = (i == 0 && step != .idv)
                Capsule()
                    .fill(done ? Color.echoTrustGreen : (active ? Color.echoSignal : Color.echoHair))
                    .frame(height: 4)
                    .animation(.spring(response: 0.4, dampingFraction: 0.85), value: step)
            }
        }
    }
}

#Preview {
    VIPStandardIDVView(did: "did:key:zDemo") { _ in }
}
#endif
