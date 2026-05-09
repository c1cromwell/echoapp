#if os(iOS)
// Features/Auth/Enrollment/EnrollmentCoordinator.swift
//
// Navigation root for the credential-first enrollment journey.
// Pushed from login screen when the user taps "Get started with a credential".

import SwiftUI
import Observation

@MainActor
@Observable
final class EnrollmentCoordinator {
    /// Path driving the NavigationStack.
    var path: [EnrollmentRoute] = []

    /// Identity bundle produced by whichever branch completed the verification step.
    /// Held transiently in memory until the tail flow consumes it; never persisted.
    var verifiedBundle: VerifiedIdentityBundle?

    /// Global enrollment progress, observed by the tail progress view.
    var stage: EnrollmentStage = .idle

    /// Called by the app coordinator when the full enrollment tail finishes.
    let onComplete: (VerifiedIdentityBundle) -> Void

    /// Called if the user bails out (back from picker).
    let onCancel: () -> Void

    init(
        onComplete: @escaping (VerifiedIdentityBundle) -> Void,
        onCancel: @escaping () -> Void
    ) {
        self.onComplete = onComplete
        self.onCancel = onCancel
    }

    // MARK: - Navigation intent

    func choose(_ method: EnrollmentMethod) {
        switch method {
        case .mobileWalletCredential:
            path.append(.walletCredential)
        case .driversLicense:
            path.append(.driversLicenseHub)
        case .phoneNumber:
            // Hand off to the existing phone-based OnboardingFlow. The app coordinator
            // observes this signal via the `onCancel` closure and swaps flows.
            path.append(.phoneHandoff)
        }
    }

    func chooseMDLSubMethod(_ sub: MDLSubMethod) {
        switch sub {
        case .appleWallet:   path.append(.mdlAppleWallet)
        case .qrEngagement:  path.append(.mdlQRScanner)
        case .nfcEngagement: path.append(.mdlNFCSession)
        case .scanAndSelfie: path.append(.mdlIDVFallback)
        }
    }

    /// Called by every branch when it has a verified bundle to hand off to the tail.
    func credentialVerified(_ bundle: VerifiedIdentityBundle) {
        self.verifiedBundle = bundle
        path.append(.tail)
    }

    func tailFinished(with bundle: VerifiedIdentityBundle) {
        self.stage = .complete
        onComplete(bundle)
    }

    func failed(_ error: EnrollmentError) {
        self.stage = .failed(error)
    }
}

enum EnrollmentRoute: Hashable {
    case walletCredential
    case driversLicenseHub
    case mdlAppleWallet
    case mdlQRScanner
    case mdlNFCSession
    case mdlIDVFallback
    case phoneHandoff
    case tail
}

// MARK: - Root Container View

struct EnrollmentCoordinatorView: View {
    @State var coordinator: EnrollmentCoordinator

    var body: some View {
        NavigationStack(path: $coordinator.path) {
            EnrollmentMethodPickerView(coordinator: coordinator)
                .navigationDestination(for: EnrollmentRoute.self) { route in
                    destination(for: route)
                }
        }
        .tint(Color.Echo.primaryContainer)
    }

    @ViewBuilder
    private func destination(for route: EnrollmentRoute) -> some View {
        switch route {
        case .walletCredential:
            WalletCredentialEnrollmentView(coordinator: coordinator)
        case .driversLicenseHub:
            DriversLicenseEnrollmentView(coordinator: coordinator)
        case .mdlAppleWallet:
            AppleWalletMDLBridgeView(coordinator: coordinator)
        case .mdlQRScanner:
            MDLQRScannerView(coordinator: coordinator)
        case .mdlNFCSession:
            MDLNFCSessionView(coordinator: coordinator)
        case .mdlIDVFallback:
            IDVFallbackView(coordinator: coordinator)
        case .phoneHandoff:
            PhoneHandoffShim(coordinator: coordinator)
        case .tail:
            EnrollmentTailView(coordinator: coordinator)
        }
    }
}

/// Tiny shim that bridges the picker into the existing phone-based OnboardingFlow.
struct PhoneHandoffShim: View {
    let coordinator: EnrollmentCoordinator
    var body: some View {
        Color.Echo.surface
            .task {
                // Bail out of the enrollment stack and let the app show
                // the existing phone-based onboarding flow.
                coordinator.onCancel()
            }
    }
}

#endif
