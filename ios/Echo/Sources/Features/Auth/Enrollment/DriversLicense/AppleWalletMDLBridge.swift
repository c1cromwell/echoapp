#if os(iOS)
// Features/Auth/Enrollment/DriversLicense/AppleWalletMDLBridge.swift
//
// Apple Wallet mDL presentation branch.
//
// iOS 26 does not expose a native API for third-party apps to request a mobile
// driver's license from Apple Wallet. The supported path is the W3C Digital
// Credentials API (DC API), surfaced through Safari / WKWebView since Safari 26.
//
// We open an ASWebAuthenticationSession pointing at a verifier page hosted on
// the ECHO backend (`verifier.echo.app`). That page calls
// `navigator.credentials.get({ digital: { providers: [{ protocol: 'org-iso-mdoc', ... }] } })`.
// The system surfaces an Apple Wallet consent sheet; the encrypted mdoc
// response is posted back to the backend, which then redirects to a callback
// URL handled via the `echo-enroll://` custom scheme.

import SwiftUI
import AuthenticationServices

@MainActor
@Observable
final class AppleWalletMDLViewModel {
    enum State {
        case idle
        case launching
        case awaitingConsent
        case verifying
        case success(VerifiedIdentityBundle)
        case failure(EnrollmentError)
    }

    var state: State = .idle

    private let api = EnrollmentAPIClient.shared

    func start(coordinator: EnrollmentCoordinator) async {
        state = .launching

        do {
            let session = try await api.startMDLPresentation(
                transport: .webDCAPI,
                claims: .minimumForTier4
            )

            state = .awaitingConsent
            let callback = try await openAuthSession(
                url: session.verifierURL,
                callbackScheme: "echo-enroll"
            )

            state = .verifying
            let bundle = try await api.finishMDLPresentation(
                sessionID: session.id,
                callbackURL: callback
            )

            state = .success(bundle)
            coordinator.credentialVerified(bundle)
        } catch let error as EnrollmentError {
            state = .failure(error)
        } catch {
            state = .failure(.transportFailed(underlying: error.localizedDescription))
        }
    }

    private func openAuthSession(url: URL, callbackScheme: String) async throws -> URL {
        try await withCheckedThrowingContinuation { continuation in
            let session = ASWebAuthenticationSession(
                url: url,
                callbackURLScheme: callbackScheme
            ) { callbackURL, error in
                if let error {
                    if case ASWebAuthenticationSessionError.canceledLogin = error {
                        continuation.resume(throwing: EnrollmentError.userCancelled)
                    } else {
                        continuation.resume(throwing: EnrollmentError.transportFailed(
                            underlying: error.localizedDescription
                        ))
                    }
                    return
                }
                guard let callbackURL else {
                    continuation.resume(throwing: EnrollmentError.transportFailed(
                        underlying: "Empty callback"
                    ))
                    return
                }
                continuation.resume(returning: callbackURL)
            }
            session.prefersEphemeralWebBrowserSession = true
            session.presentationContextProvider = AuthPresentationContextProvider.shared
            session.start()
        }
    }
}

// MARK: - View

struct AppleWalletMDLBridgeView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = AppleWalletMDLViewModel()

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 22) {
                Spacer(minLength: 20)

                ZStack {
                    LinearGradient(
                        colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                        startPoint: .topLeading, endPoint: .bottomTrailing
                    )
                    .clipShape(Circle())
                    Image(systemName: "applelogo")
                        .font(.system(size: 30, weight: .semibold))
                        .foregroundStyle(.white)
                }
                .frame(width: 88, height: 88)

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

                claimsPreview

                content

                Spacer()
            }
        }
        .navigationTitle("Apple Wallet")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    private var titleText: String {
        switch viewModel.state {
        case .idle, .launching:   return "Share your mobile ID"
        case .awaitingConsent:    return "Confirm in Wallet"
        case .verifying:          return "Verifying issuer"
        case .success:            return "Verified"
        case .failure:            return "Couldn't complete"
        }
    }

    private var subtitleText: String {
        switch viewModel.state {
        case .idle, .launching:
            return "Apple Wallet will ask you to approve exactly what's shared. We request only these fields."
        case .awaitingConsent:
            return "Your system consent sheet is open. Approve with Face ID to continue."
        case .verifying:
            return "Checking signature against the DMV's IACA root certificate."
        case .success:
            return "Your mDL is verified. One more step to finish enrolling."
        case .failure(let error):
            return error.errorDescription ?? "Something went wrong."
        }
    }

    private var claimsPreview: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("ECHO will request")
                .font(.system(size: 11, weight: .semibold))
                .tracking(1.5)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            VStack(alignment: .leading, spacing: 6) {
                claimRow("person.fill", "Given name and family name")
                claimRow("checkmark.shield.fill", "Age over 18")
                claimRow("flag.fill", "Issuing country")
            }
            .padding(12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.Echo.surfaceContainerLowest, in: RoundedRectangle(cornerRadius: 16))
            .overlay(
                RoundedRectangle(cornerRadius: 16)
                    .strokeBorder(Color.Echo.primaryContainer.opacity(0.20), lineWidth: 0.5)
            )
        }
        .padding(.horizontal, 20)
    }

    private func claimRow(_ systemImage: String, _ text: String) -> some View {
        HStack(spacing: 9) {
            Image(systemName: systemImage)
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 16)
            Text(text)
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurface)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .idle:
            actionButton("Continue with Apple Wallet")
                .padding(.horizontal, 20)
        case .launching, .awaitingConsent, .verifying:
            ProgressView().controlSize(.large).tint(Color.Echo.primaryContainer).padding(.top, 8)
        case .success:
            EmptyView()
        case .failure:
            VStack(spacing: 10) {
                actionButton("Try again")
                Button("Pick a different method") {
                    coordinator.path.removeLast()
                }
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(.horizontal, 20)
        }
    }

    private func actionButton(_ label: String) -> some View {
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
