#if os(iOS)
// Features/Onboarding/FirstRun/OnboardingOptionsView.swift
// Step 2 of onboarding: mandatory Face ID enrollment + optional VIP signup.
//
// Face ID is required before "Continue" unlocks.
// VIP checkbox is independent — user can skip VIP and go straight to messaging.

import SwiftUI

public struct OnboardingOptionsView: View {
    @Bindable var coordinator: FirstRunCoordinator

    @State private var enrolledDID: String? = nil
    @State private var wantsVIP = false
    @State private var showBiometricSheet = false

    public init(coordinator: FirstRunCoordinator) {
        self.coordinator = coordinator
    }

    private var faceIDEnrolled: Bool { enrolledDID != nil }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(alignment: .leading, spacing: 0) {
                // Header
                VStack(alignment: .leading, spacing: 6) {
                    Text("Secure your account")
                        .font(.system(size: 28, weight: .semibold))
                        .tracking(-0.7)
                        .foregroundStyle(Color.echoInk)
                    Text("Face ID is required. Trusted User verification is optional and can be done later.")
                        .font(.system(size: 14))
                        .lineSpacing(3)
                        .foregroundStyle(Color.echoInk55)
                }
                .padding(.top, 28)
                .padding(.horizontal, 24)

                // Cards
                VStack(spacing: 14) {
                    faceIDCard
                    vipCard
                }
                .padding(.horizontal, 24)
                .padding(.top, 28)

                Spacer(minLength: 16)

                // CTA
                VStack(spacing: 10) {
                    Button {
                        coordinator.onboardingOptionsContinued(did: enrolledDID ?? "", wantsVIP: wantsVIP)
                    } label: {
                        HStack {
                            Text("Continue")
                                .font(.system(size: 15, weight: .semibold))
                            Spacer()
                            Text("→").font(.system(size: 18))
                        }
                        .foregroundStyle(Color.echoPaper)
                        .padding(.horizontal, 18)
                        .padding(.vertical, 15)
                        .background(
                            faceIDEnrolled ? Color.echoInk : Color.echoInk.opacity(0.3),
                            in: RoundedRectangle(cornerRadius: 14)
                        )
                    }
                    .disabled(!faceIDEnrolled)
                    .buttonStyle(SpringPressStyle())

                    Button("Back") { coordinator.back() }
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.echoInk55)
                        .padding(.vertical, 6)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 28)
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .sheet(isPresented: $showBiometricSheet) {
            BiometricEnrollmentView(
                username: coordinator.displayName,
                onComplete: { did in
                    enrolledDID = did
                    showBiometricSheet = false
                },
                onUnsupported: {
                    showBiometricSheet = false
                }
            )
        }
    }

    // MARK: - Face ID Card

    private var faceIDCard: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(faceIDEnrolled ? Color.echoTrustGreen.opacity(0.12) : Color.echoSignal.opacity(0.10))
                    .frame(width: 52, height: 52)
                Image(systemName: faceIDEnrolled ? "checkmark.seal.fill" : "faceid")
                    .font(.system(size: 24, weight: .medium))
                    .foregroundStyle(faceIDEnrolled ? Color.echoTrustGreen : Color.echoSignal)
            }

            VStack(alignment: .leading, spacing: 3) {
                HStack(spacing: 6) {
                    Text("Face ID")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                    Text("Required")
                        .font(.system(size: 10, weight: .semibold))
                        .tracking(0.8)
                        .foregroundStyle(Color.echoSignal)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.echoSignal.opacity(0.10), in: Capsule())
                }
                Text(faceIDEnrolled ? "Enrolled — your key is set" : "Your face is your only key")
                    .font(.system(size: 12))
                    .foregroundStyle(Color.echoInk55)
            }

            Spacer()

            if faceIDEnrolled {
                Image(systemName: "checkmark.circle.fill")
                    .font(.system(size: 22))
                    .foregroundStyle(Color.echoTrustGreen)
            } else {
                Button("Set Up") {
                    showBiometricSheet = true
                }
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(Color.echoSignal, in: Capsule())
            }
        }
        .padding(18)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(
                    faceIDEnrolled ? Color.echoTrustGreen.opacity(0.35) : Color.echoHair,
                    lineWidth: 1
                )
        )
    }

    // MARK: - VIP Card

    private var vipCard: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(Color.echoTrustGreen.opacity(wantsVIP ? 0.15 : 0.07))
                    .frame(width: 52, height: 52)
                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 24, weight: .medium))
                    .foregroundStyle(Color.echoTrustGreen.opacity(wantsVIP ? 1 : 0.4))
            }

            VStack(alignment: .leading, spacing: 3) {
                HStack(spacing: 6) {
                    Text("Verified User")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                    Text("Optional")
                        .font(.system(size: 10, weight: .semibold))
                        .tracking(0.8)
                        .foregroundStyle(Color.echoInk40)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.echoHair, in: Capsule())
                }
                Text("Get a trusted badge. Verify with a digital ID or government document.")
                    .font(.system(size: 12))
                    .lineSpacing(2)
                    .foregroundStyle(Color.echoInk55)
            }

            Spacer()

            Toggle("", isOn: $wantsVIP)
                .labelsHidden()
                .tint(Color.echoTrustGreen)
        }
        .padding(18)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(
                    wantsVIP ? Color.echoTrustGreen.opacity(0.35) : Color.echoHair,
                    lineWidth: 1
                )
        )
        .animation(.spring(response: 0.3, dampingFraction: 0.8), value: wantsVIP)
    }
}

#Preview {
    OnboardingOptionsView(
        coordinator: FirstRunCoordinator(
            onComplete: { _, _ in },
            onRestoreComplete: { _ in }
        )
    )
}
#endif
