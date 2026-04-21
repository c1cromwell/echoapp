// Features/Auth/Enrollment/Wallet/WalletCredentialEnrollmentView.swift
//
// Mobile wallet credential (OIDC4VC / OID4VP) enrollment branch.
// Uses ASWebAuthenticationSession to invoke the W3C Digital Credentials API
// via a verifier page hosted on the ECHO backend. The backend generates the
// request manifest, signs it with the verifier reader key, and validates the
// returned mdoc or SD-JWT VP before issuing an enrollment token.

import SwiftUI
import AuthenticationServices

@MainActor
@Observable
final class WalletCredentialEnrollmentViewModel {
    enum State {
        case idle
        case generatingRequest
        case awaitingWallet
        case verifying
        case success(VerifiedIdentityBundle)
        case failure(EnrollmentError)
    }

    var state: State = .idle

    private let api = EnrollmentAPIClient.shared

    func start(coordinator: EnrollmentCoordinator) async {
        state = .generatingRequest

        do {
            let session = try await api.startWalletPresentation(
                claims: .minimumForTier4
            )

            state = .awaitingWallet
            let callback = try await openAuthSession(
                url: session.verifierURL,
                callbackScheme: "echo-enroll"
            )

            state = .verifying
            let bundle = try await api.finishWalletPresentation(
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

struct WalletCredentialEnrollmentView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = WalletCredentialEnrollmentViewModel()

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 24) {
                Spacer(minLength: 20)

                iconBadge
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

                if case .idle = viewModel.state {
                    startButton
                        .padding(.horizontal, 20)
                        .padding(.top, 16)
                } else if case .failure(let error) = viewModel.state {
                    errorPanel(error)
                        .padding(.horizontal, 20)
                        .padding(.top, 16)
                } else {
                    ProgressView()
                        .controlSize(.large)
                        .tint(Color.Echo.primaryContainer)
                        .padding(.top, 24)
                }

                Spacer()
            }
        }
        .navigationTitle("Wallet credential")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    // MARK: - Subviews

    private var iconBadge: some View {
        ZStack {
            LinearGradient(
                colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .clipShape(Circle())

            Image(systemName: "wallet.pass.fill")
                .font(.system(size: 32, weight: .semibold))
                .foregroundStyle(.white)
        }
    }

    private var titleText: String {
        switch viewModel.state {
        case .idle:                return "Present a credential"
        case .generatingRequest:   return "Preparing request"
        case .awaitingWallet:      return "Open your wallet"
        case .verifying:           return "Verifying issuer"
        case .success:             return "Verified"
        case .failure:             return "Couldn't verify"
        }
    }

    private var subtitleText: String {
        switch viewModel.state {
        case .idle:
            return "ECHO will ask your wallet for your name, country, and age-over-18 — nothing else. You can review before sharing."
        case .generatingRequest, .awaitingWallet:
            return "Pick the credential you want to share from the system sheet."
        case .verifying:
            return "Checking the issuer's signature and revocation status."
        case .success:
            return "Your credential is verified. One more step to finish enrolling."
        case .failure(let error):
            return error.errorDescription ?? "Something went wrong."
        }
    }

    private var startButton: some View {
        Button {
            Task { await viewModel.start(coordinator: coordinator) }
        } label: {
            HStack {
                Image(systemName: "wallet.pass")
                    .font(.system(size: 15, weight: .semibold))
                Text("Open wallet")
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

    @ViewBuilder
    private func errorPanel(_ error: EnrollmentError) -> some View {
        VStack(spacing: 12) {
            Button {
                Task { await viewModel.start(coordinator: coordinator) }
            } label: {
                Text("Try again")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.white)
                    .padding(.vertical, 14)
                    .frame(maxWidth: .infinity)
                    .background(Color.Echo.primaryContainer, in: RoundedRectangle(cornerRadius: 24))
            }
            Button {
                coordinator.path.removeLast()
            } label: {
                Text("Pick a different method")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
        }
    }
}
