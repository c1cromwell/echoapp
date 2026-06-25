// Core/Mesh/MeshGateView.swift
//
// Gates the VERIFIED mesh lane behind a completed IDV + selfie/liveness (echo.idvVerified) —
// deliberately NOT a paid VIP tier (which only raises trustTier). Unverified users see a prompt
// to verify. The anonymous lane (free, limited) is a separate entry and is not gated here.

#if os(iOS)
import SwiftUI

public struct MeshGateView<Content: View>: View {
    private let onVerify: () -> Void
    @ViewBuilder private let content: () -> Content
    @State private var verified = CurrentUserSession.isIdentityVerified()

    public init(onVerify: @escaping () -> Void = {}, @ViewBuilder content: @escaping () -> Content) {
        self.onVerify = onVerify
        self.content = content
    }

    public var body: some View {
        Group {
            if verified {
                content()
            } else {
                locked
            }
        }
        .onAppear { verified = CurrentUserSession.isIdentityVerified() }
    }

    private var locked: some View {
        VStack(spacing: 16) {
            Image(systemName: "dot.radiowaves.left.and.right")
                .font(.system(size: 40))
                .foregroundColor(.echoSignal)
            Text("Verify to use offline mesh")
                .font(.system(size: 20, weight: .semibold))
                .foregroundColor(.echoInk)
            Text("Mesh lets you message people nearby with no internet — directly over Bluetooth. "
                 + "It's for verified people only, so finish a quick ID + selfie check to turn it on.")
                .font(.system(size: 14))
                .foregroundColor(.echoInk55)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 24)
            Button(action: onVerify) {
                Text("Verify identity")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: 52)
                    .background(Color.echoSignal)
                    .clipShape(RoundedRectangle(cornerRadius: 14))
            }
            .padding(.horizontal, 24)
            .padding(.top, 4)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.echoPaper)
    }
}
#endif
