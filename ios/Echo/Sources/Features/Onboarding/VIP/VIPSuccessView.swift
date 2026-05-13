#if os(iOS)
// Features/Onboarding/VIP/VIPSuccessView.swift
// Celebration screen shown after completing VIP identity verification.
// T2 → "Verified User" · T4 → "Trusted User"

import SwiftUI

public struct VIPSuccessView: View {
    let trustTier: Int
    let onContinue: () -> Void

    public init(trustTier: Int, onContinue: @escaping () -> Void) {
        self.trustTier = trustTier
        self.onContinue = onContinue
    }

    private var evidenceType: String {
        trustTier >= 4 ? "digital_id" : "standard_idv"
    }

    private var isElite: Bool { trustTier >= 4 }

    private var headline: String {
        isElite ? "You're a Trusted User" : "You're a Verified User"
    }

    private var subtext: String {
        isElite
            ? "Your identity has been confirmed with a digital credential. The full green badge will appear on your profile and messages."
            : "Your identity has been confirmed. A verified checkmark will appear on your profile and messages."
    }

    private var tierLabel: String { isElite ? "Trusted" : "Verified" }
    private var tierColor: Color  { isElite ? Color.echoTrustElite : Color.echoTrustVerified }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                // Mark
                EchoRippleMark(size: 64, color: Color.echoTrustGreen)
                    .padding(.bottom, 28)

                // Badge
                ZStack {
                    Circle()
                        .fill(tierColor.opacity(0.12))
                        .frame(width: 96, height: 96)
                    Image(systemName: "checkmark.seal.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(tierColor)
                }
                .padding(.bottom, 28)

                // Copy
                VStack(spacing: 10) {
                    Text(headline)
                        .font(.system(size: 26, weight: .semibold))
                        .tracking(-0.7)
                        .foregroundStyle(Color.echoInk)
                        .multilineTextAlignment(.center)
                    Text(subtext)
                        .font(.system(size: 14))
                        .lineSpacing(4)
                        .foregroundStyle(Color.echoInk55)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                }

                // Tier badge
                HStack(spacing: 6) {
                    Image(systemName: "checkmark.seal.fill")
                        .font(.system(size: 13))
                    Text(tierLabel)
                        .font(.system(size: 13, weight: .semibold))
                }
                .foregroundStyle(tierColor)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(tierColor.opacity(0.10), in: Capsule())
                .padding(.top, 18)

                Spacer()

                // CTA
                Button(action: onContinue) {
                    HStack {
                        Text("Start messaging")
                            .font(.system(size: 15, weight: .semibold))
                        Spacer()
                        Text("→").font(.system(size: 18))
                    }
                    .foregroundStyle(Color.echoPaper)
                    .padding(.horizontal, 18)
                    .padding(.vertical, 15)
                    .background(Color.echoInk, in: RoundedRectangle(cornerRadius: 14))
                }
                .buttonStyle(SpringPressStyle())
                .padding(.horizontal, 24)
                .padding(.bottom, 32)
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .task {
            // Persist trust tier to backend (best-effort, does not block the UX)
            guard let did = UserDefaults.standard.string(forKey: "echo.did.current") else { return }
            try? await RealProvisionAPI().updateTrustTier(did: did, trustTier: trustTier, evidenceType: evidenceType)
        }
    }
}

#Preview("T2 Verified") {
    VIPSuccessView(trustTier: 2) {}
}
#Preview("T4 Elite") {
    VIPSuccessView(trustTier: 4) {}
}
#endif
