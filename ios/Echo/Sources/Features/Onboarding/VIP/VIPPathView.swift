#if os(iOS)
// Features/Onboarding/VIP/VIPPathView.swift
// Choose how to verify identity for the Trusted User (VIP) badge.
// Two paths:
//   • Digital ID  — mDL or OIDC4VC credential  → T4 Elite
//   • Standard     — scan ID + selfie + phone   → T2 Verified

import SwiftUI

public struct VIPPathView: View {
    let did: String
    let onDigitalID: (Int) -> Void   // passes trustTier
    let onStandardIDV: (Int) -> Void
    let onSkip: () -> Void

    @State private var showEnrollment = false

    public init(
        did: String,
        onDigitalID: @escaping (Int) -> Void,
        onStandardIDV: @escaping (Int) -> Void,
        onSkip: @escaping () -> Void
    ) {
        self.did = did
        self.onDigitalID = onDigitalID
        self.onStandardIDV = onStandardIDV
        self.onSkip = onSkip
    }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(alignment: .leading, spacing: 0) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Verify your identity")
                        .font(.system(size: 28, weight: .semibold))
                        .tracking(-0.7)
                        .foregroundStyle(Color.echoInk)
                    Text("Choose how you'd like to prove who you are.")
                        .font(.system(size: 14))
                        .lineSpacing(3)
                        .foregroundStyle(Color.echoInk55)
                }
                .padding(.top, 28)
                .padding(.horizontal, 24)

                VStack(spacing: 14) {
                    // Digital ID card (primary)
                    Button {
                        showEnrollment = true
                    } label: {
                        pathCard(
                            icon: "creditcard.fill",
                            iconColor: Color.echoSignal,
                            title: "Digital ID",
                            subtitle: "Use a mobile driver's licence or digital wallet credential.",
                            badge: "T4 Trusted",
                            badgeColor: Color.echoTrustGreen,
                            isPrimary: true
                        )
                    }
                    .buttonStyle(SpringPressStyle())

                    // Standard IDV card
                    NavigationLink(value: FirstRunCoordinator.Route.vipStandardIDV(did: did)) {
                        pathCard(
                            icon: "doc.text.viewfinder",
                            iconColor: Color.echoInk55,
                            title: "Standard Verification",
                            subtitle: "Scan your licence or passport, take a selfie, and add a recovery phone number.",
                            badge: "T2 Verified",
                            badgeColor: Color.echoTrustVerified,
                            isPrimary: false
                        )
                    }
                    .buttonStyle(SpringPressStyle())
                }
                .padding(.horizontal, 24)
                .padding(.top, 28)

                Spacer(minLength: 16)

                Button("Skip for now — I'll verify later in Settings") { onSkip() }
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk55)
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding(.bottom, 32)
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .sheet(isPresented: $showEnrollment) {
            EnrollmentCoordinatorView(
                coordinator: EnrollmentCoordinator(
                    onComplete: { bundle in
                        showEnrollment = false
                        onDigitalID(bundle.assuranceLevel.trustTier)
                    },
                    onCancel: { showEnrollment = false }
                )
            )
        }
    }

    // MARK: - Card builder

    private func pathCard(
        icon: String,
        iconColor: Color,
        title: String,
        subtitle: String,
        badge: String,
        badgeColor: Color,
        isPrimary: Bool
    ) -> some View {
        HStack(alignment: .top, spacing: 16) {
            ZStack {
                RoundedRectangle(cornerRadius: 12)
                    .fill(iconColor.opacity(0.10))
                    .frame(width: 48, height: 48)
                Image(systemName: icon)
                    .font(.system(size: 22, weight: .medium))
                    .foregroundStyle(iconColor)
            }

            VStack(alignment: .leading, spacing: 5) {
                HStack(spacing: 8) {
                    Text(title)
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                    Text(badge)
                        .font(.system(size: 10, weight: .semibold))
                        .foregroundStyle(badgeColor)
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(badgeColor.opacity(0.12), in: Capsule())
                }
                Text(subtitle)
                    .font(.system(size: 13))
                    .lineSpacing(3)
                    .foregroundStyle(Color.echoInk55)
                    .multilineTextAlignment(.leading)
            }

            Spacer()
            Image(systemName: "chevron.right")
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Color.echoInk40)
                .padding(.top, 4)
        }
        .padding(18)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(
                    isPrimary ? Color.echoSignal.opacity(0.35) : Color.echoHair,
                    lineWidth: isPrimary ? 1.5 : 1
                )
        )
    }
}

#Preview {
    NavigationStack {
        VIPPathView(
            did: "did:key:zDemo",
            onDigitalID: { _ in },
            onStandardIDV: { _ in },
            onSkip: {}
        )
    }
}
#endif
