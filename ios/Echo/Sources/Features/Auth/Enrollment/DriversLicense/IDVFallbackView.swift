#if os(iOS)
// Features/Auth/Enrollment/DriversLicense/IDVFallbackView.swift
//
// IAL2 identity verification fallback for users without a mobile driver's
// license. Uses a third-party IDV provider (Stripe Identity in Phase 1, or
// Sumsub) via its native iOS SDK.
//
// Privacy: the IDV provider creates a direct TLS session from the ECHO app
// to its servers. ECHO's backend only receives the reference ID and a
// pass/fail + confidence score result. Raw ID images and selfie video never
// transit ECHO infrastructure. See MISSING_FEATURES_GAP_ANALYSIS.md §4
// for SDK integration steps.

import SwiftUI

@MainActor
@Observable
final class IDVFallbackViewModel {
    enum State {
        case idle
        case launching
        case capturing
        case processing
        case success(VerifiedIdentityBundle)
        case failure(EnrollmentError)
    }

    var state: State = .idle

    private let api = EnrollmentAPIClient.shared

    func start(coordinator: EnrollmentCoordinator) async {
        state = .launching
        do {
            let session = try await api.startIDVSession()

            // PRODUCTION INTEGRATION POINT — Stripe Identity:
            //
            //   import StripeIdentity
            //
            //   state = .capturing
            //   let result = try await StripeIdentity.present(
            //       verificationSessionClientSecret: session.clientSecret,
            //       brandLogo: UIImage(named: "EchoLogoMark") ?? UIImage()
            //   )
            //   guard result.status == .verified else { throw ... }
            //
            //   state = .processing
            //   let bundle = try await api.awaitIDVResult(sessionID: session.id)
            //   state = .success(bundle)
            //   coordinator.credentialVerified(bundle)

            state = .capturing
            _ = session
            throw EnrollmentError.deviceUnsupported(
                reason: "IDV SDK not wired. Link StripeIdentity via SPM per MISSING_FEATURES_GAP_ANALYSIS.md §4."
            )
        } catch let error as EnrollmentError {
            state = .failure(error)
        } catch {
            state = .failure(.transportFailed(underlying: error.localizedDescription))
        }
    }
}

// MARK: - View

public struct IDVFallbackView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = IDVFallbackViewModel()

    public init(coordinator: EnrollmentCoordinator) {
        self.coordinator = coordinator
    }

    public var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 22) {
                Spacer(minLength: 20)

                iconBadge.frame(width: 88, height: 88)

                VStack(spacing: 8) {
                    Text(titleText)
                        .font(.system(size: 22, weight: .semibold))
                        .kerning(-0.3)
                        .foregroundStyle(Color.Echo.onSurface)
                        .multilineTextAlignment(.center)
                    Text(subtitleText)
                        .font(.system(size: 14))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 24)
                }

                privacyCallout.padding(.horizontal, 20)

                content

                Spacer()
            }
        }
        .navigationTitle("Scan + selfie")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    private var iconBadge: some View {
        ZStack {
            LinearGradient(
                colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                startPoint: .topLeading, endPoint: .bottomTrailing
            )
            .clipShape(Circle())
            Image(systemName: "doc.viewfinder")
                .font(.system(size: 30, weight: .semibold))
                .foregroundStyle(.white)
        }
    }

    private var titleText: String {
        switch viewModel.state {
        case .idle, .launching:    return "Scan your license"
        case .capturing:           return "Capturing"
        case .processing:          return "Verifying"
        case .success:             return "Verified"
        case .failure:             return "Couldn't verify"
        }
    }

    private var subtitleText: String {
        switch viewModel.state {
        case .idle:
            return "Photograph the front and back of your license, then a short liveness selfie. About 2 minutes."
        case .launching:
            return "Starting a secure session with our identity partner."
        case .capturing:
            return "Follow the on-screen prompts to finish capture."
        case .processing:
            return "Our partner is confirming the document and liveness check."
        case .success:
            return "Your identity is verified. One more step to finish enrolling."
        case .failure(let error):
            return error.errorDescription ?? "Something went wrong."
        }
    }

    private var privacyCallout: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "lock.shield.fill")
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 18)
            Text("Your ID images go directly to our identity partner over TLS — ECHO never sees them. Images are deleted after verification.")
                .font(.system(size: 12))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .lineSpacing(2)
        }
        .padding(14)
        .background(Color.Echo.primaryContainer.opacity(0.06), in: RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .strokeBorder(Color.Echo.primaryContainer.opacity(0.20), lineWidth: 0.5)
        )
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .idle:
            startButton("Start verification").padding(.horizontal, 20)
        case .launching, .capturing, .processing:
            ProgressView().controlSize(.large).tint(Color.Echo.primaryContainer).padding(.top, 8)
        case .success:
            EmptyView()
        case .failure:
            VStack(spacing: 10) {
                startButton("Try again")
                Button("Pick a different method") { coordinator.path.removeLast() }
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(.horizontal, 20)
        }
    }

    private func startButton(_ label: String) -> some View {
        Button {
            Task { await viewModel.start(coordinator: coordinator) }
        } label: {
            HStack {
                Text(label)
                    .font(.system(size: 15, weight: .semibold))
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.6))
            }
            .foregroundStyle(.white)
            .padding(.horizontal, 18)
            .padding(.vertical, 16)
            .frame(maxWidth: .infinity)
            .background(
                LinearGradient(
                    colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                    startPoint: .topLeading, endPoint: .bottomTrailing
                ),
                in: RoundedRectangle(cornerRadius: 28, style: .continuous)
            )
        }
        .buttonStyle(SpringPressStyle())
    }
}
#endif
