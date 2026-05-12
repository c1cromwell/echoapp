#if os(iOS)
// Features/Auth/Enrollment/DriversLicense/DriversLicenseEnrollmentView.swift
//
// Hub screen that offers four ways to present a mobile driver's license:
//   1. Apple Wallet mDL (W3C DC API) — primary
//   2. QR device engagement (ISO 18013-5 §8.2)
//   3. NFC device engagement (ISO 18013-5 §8.3)
//   4. IDV fallback — scan + selfie (Stripe Identity / Sumsub)
//
// The capability probe at init-time disables paths the device can't support
// (no NFC hardware, camera denied, no mDL provisioned in Wallet).

import SwiftUI
#if canImport(CoreNFC)
import CoreNFC
#endif
import AVFoundation

@MainActor
@Observable
final class DriversLicenseEnrollmentViewModel {
    struct Capability {
        var appleWalletAvailable: Bool
        var nfcAvailable: Bool
        var cameraAvailable: Bool
    }

    var capability: Capability = .init(
        appleWalletAvailable: true,    // assume until DC API probe fails
        nfcAvailable: NFCTagReaderSession.readingAvailable,
        cameraAvailable: true          // verified on tap to avoid prompt on load
    )

    func probeCamera() async {
        switch AVCaptureDevice.authorizationStatus(for: .video) {
        case .notDetermined:
            capability.cameraAvailable = await AVCaptureDevice.requestAccess(for: .video)
        case .denied, .restricted:
            capability.cameraAvailable = false
        case .authorized:
            capability.cameraAvailable = true
        @unknown default:
            capability.cameraAvailable = false
        }
    }
}

// MARK: - View

public struct DriversLicenseEnrollmentView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = DriversLicenseEnrollmentViewModel()

    public init(coordinator: EnrollmentCoordinator) {
        self.coordinator = coordinator
    }

    public var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            ScrollView {
                VStack(spacing: 18) {
                    titleBlock

                    primaryPathsSection

                    fallbackSection

                    rewardCallout
                }
                .padding(.horizontal, 16)
                .padding(.top, 12)
                .padding(.bottom, 28)
            }
            .scrollBounceBehavior(.basedOnSize)
        }
        .navigationTitle("Driver's license")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    // MARK: - Sections

    private var titleBlock: some View {
        VStack(spacing: 6) {
            Text("Present your mobile license")
                .font(.system(size: 20, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
                .multilineTextAlignment(.center)
            Text("ECHO never sees your license number or address — only the fields you approve on the consent sheet.")
                .font(.system(size: 12.5))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .multilineTextAlignment(.center)
                .lineSpacing(2)
                .padding(.horizontal, 16)
        }
        .padding(.top, 12)
    }

    private var primaryPathsSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            TrackedSectionDivider(label: "Fastest", tint: Color.Echo.primaryContainer)

            // 1. Apple Wallet mDL — W3C DC API
            EnrollmentOptionCard(
                style: .primaryGradient,
                title: "Apple Wallet",
                subtitle: "Share your mobile ID from Wallet using Face ID.",
                tags: ["ISO 18013-7", "Selective disclosure"],
                badge: viewModel.capability.appleWalletAvailable ? nil : "Unavailable",
                icon: {
                    Image(systemName: "applelogo")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(.white)
                }
            ) {
                guard viewModel.capability.appleWalletAvailable else { return }
                coordinator.chooseMDLSubMethod(.appleWallet)
            }
            .opacity(viewModel.capability.appleWalletAvailable ? 1.0 : 0.5)

            // 2. QR engagement
            EnrollmentOptionCard(
                style: .secondaryTinted,
                title: "Scan a QR code",
                subtitle: "Scan the QR shown by an issuer or reader terminal.",
                tags: ["ISO 18013-5 §8.2"],
                icon: {
                    Image(systemName: "qrcode.viewfinder")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
            ) {
                Task {
                    await viewModel.probeCamera()
                    if viewModel.capability.cameraAvailable {
                        coordinator.chooseMDLSubMethod(.qrEngagement)
                    }
                }
            }

            // 3. NFC engagement
            EnrollmentOptionCard(
                style: .secondaryTinted,
                title: "Tap to an NFC reader",
                subtitle: "Hold your phone near a reader to begin device engagement.",
                tags: ["ISO 18013-5 §8.3"],
                badge: viewModel.capability.nfcAvailable ? nil : "Not supported",
                icon: {
                    Image(systemName: "wave.3.right")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
            ) {
                guard viewModel.capability.nfcAvailable else { return }
                coordinator.chooseMDLSubMethod(.nfcEngagement)
            }
            .opacity(viewModel.capability.nfcAvailable ? 1.0 : 0.5)
        }
    }

    private var fallbackSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            TrackedSectionDivider(label: "No mobile license yet", tint: Color.Echo.onSurfaceVariant)

            EnrollmentOptionCard(
                style: .tertiaryNeutral,
                title: "Scan a physical license",
                subtitle: "Photograph the front and back, then a quick liveness selfie.",
                tags: [],
                icon: {
                    Image(systemName: "doc.viewfinder")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                }
            ) {
                coordinator.chooseMDLSubMethod(.scanAndSelfie)
            }
        }
    }

    private var rewardCallout: some View {
        HStack(spacing: 10) {
            Image(systemName: "gift.fill")
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
            Text("Any successful path rewards you with **100 ECHO** and Tier 4 Verified status.")
                .font(.system(size: 12))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .lineSpacing(2)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(
            Color.Echo.primaryContainer.opacity(0.06),
            in: RoundedRectangle(cornerRadius: 14)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 14)
                .strokeBorder(Color.Echo.primaryContainer.opacity(0.20), lineWidth: 0.5)
        )
        .padding(.top, 8)
    }
}
#endif
