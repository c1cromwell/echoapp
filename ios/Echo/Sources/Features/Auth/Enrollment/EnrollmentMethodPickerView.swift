// Features/Auth/Enrollment/EnrollmentMethodPickerView.swift
//
// First screen of the new enrollment journey. Promoted from login screen.
// Shows three paths: mobile wallet credential (gradient, primary),
// driver's license (frosted secondary), and phone number (muted tertiary).

import SwiftUI

struct EnrollmentMethodPickerView: View {
    let coordinator: EnrollmentCoordinator

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            ScrollView {
                VStack(spacing: 0) {
                    header

                    VStack(spacing: 18) {
                        titleBlock

                        recommendedSection

                        alternativeSection
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 18)
                    .padding(.bottom, 28)
                }
            }
            .scrollBounceBehavior(.basedOnSize)
        }
        .navigationBarBackButtonHidden(true)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button {
                    coordinator.onCancel()
                } label: {
                    Image(systemName: "chevron.backward")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                        .padding(8)
                        .background(Color.Echo.primaryContainer.opacity(0.09), in: Circle())
                }
                .accessibilityLabel("Back to login")
            }
        }
    }

    // MARK: - Subviews

    private var header: some View {
        HStack(spacing: 8) {
            // Use the project's existing EchoLogo component
            EchoLogo(size: 22)
            Text("ECHO")
                .font(.system(size: 15, weight: .semibold))
                .kerning(-0.2)
                .foregroundStyle(Color.Echo.primaryContainer)
            Spacer()
            Text("Get started")
                .font(.system(size: 11, weight: .semibold))
                .tracking(1.5)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .padding(.horizontal, 18)
        .padding(.top, 12)
        .padding(.bottom, 11)
        .background(
            Color.Echo.surface.opacity(0.65)
                .background(.ultraThinMaterial)
        )
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(Color.Echo.primaryContainer.opacity(0.18))
                .frame(height: 0.5)
        }
    }

    private var titleBlock: some View {
        VStack(spacing: 6) {
            Text("Choose how to enroll")
                .font(.system(size: 22, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
            Text("Every path creates an on-device identity you fully control. No email. No password.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .multilineTextAlignment(.center)
                .lineSpacing(2)
        }
        .padding(.horizontal, 8)
        .padding(.top, 12)
    }

    private var recommendedSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            TrackedSectionDivider(label: "Recommended", tint: Color.Echo.primaryContainer)

            EnrollmentOptionCard(
                style: .primaryGradient,
                title: "Mobile wallet credential",
                subtitle: "Apple Wallet, Google Wallet, or bank-issued identity VC",
                tags: ["Instant", "OIDC4VC"],
                icon: {
                    Image(systemName: "wallet.pass.fill")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.white)
                }
            ) {
                coordinator.choose(.mobileWalletCredential)
            }

            EnrollmentOptionCard(
                style: .secondaryTinted,
                title: "Driver's license",
                subtitle: "Apple Wallet mDL, scan a QR, or tap an NFC reader",
                tags: ["Tier 4 verified", "+100 ECHO"],
                icon: {
                    Image(systemName: "person.text.rectangle.fill")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
            ) {
                coordinator.choose(.driversLicense)
            }
        }
    }

    private var alternativeSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            TrackedSectionDivider(label: "Low-friction alternative", tint: Color.Echo.onSurfaceVariant)

            EnrollmentOptionCard(
                style: .tertiaryNeutral,
                title: "Phone number",
                subtitle: "SMS verification — starts at Tier 1, upgradeable later",
                tags: [],
                icon: {
                    Image(systemName: "iphone")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                }
            ) {
                coordinator.choose(.phoneNumber)
            }
        }
    }
}
