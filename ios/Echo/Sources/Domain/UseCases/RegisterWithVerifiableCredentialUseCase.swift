#if os(iOS)
import Foundation
import AuthenticationServices

/// Orchestrates OIDC4VC wallet registration (WO-100).
@MainActor
struct RegisterWithVerifiableCredentialUseCase {
    private let oidc: OIDC4VCClient
    private let enrollmentAPI: EnrollmentAPIClient
    private let walletConnector: WalletConnector

    init(
        oidc: OIDC4VCClient = OIDC4VCClient(),
        enrollmentAPI: EnrollmentAPIClient,
        walletConnector: WalletConnector = WalletConnector()
    ) {
        self.oidc = oidc
        self.enrollmentAPI = enrollmentAPI
        self.walletConnector = walletConnector
    }

    func execute(claims: EnrollmentClaimsRequest = .minimumForTier4) async throws -> VerifiedIdentityBundle {
        let session = try await enrollmentAPI.startWalletPresentation(claims: claims)
        let callback = try await walletConnector.present(
            url: session.verifierURL,
            callbackScheme: "echo-enroll"
        )
        return try await enrollmentAPI.finishWalletPresentation(
            sessionID: session.id,
            callbackURL: callback
        )
    }

    /// Direct OIDC4VC path when a presentation request is already issued.
    func submitDirect(
        vpToken: String,
        state: String,
        credentialType: String = "KYCLite"
    ) async throws -> VerifiedIdentityBundle {
        let submission = OIDC4VCClient.defaultSubmission(credentialType: credentialType)
        let result = try await oidc.submitPresentation(
            vpToken: vpToken,
            state: state,
            submission: submission
        )
        guard result.verificationResult?.isValid == true else {
            throw EnrollmentError.signatureInvalid
        }
        return try await enrollmentAPI.finishWalletPresentation(
            sessionID: state,
            callbackURL: URL(string: "echo-enroll://callback?vp_token=\(vpToken.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? vpToken)&state=\(state)")!
        )
    }
}

/// Opens ASWebAuthenticationSession for wallet / verifier UI handoff.
@MainActor
struct WalletConnector {
    func present(url: URL, callbackScheme: String) async throws -> URL {
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
                    continuation.resume(throwing: EnrollmentError.transportFailed(underlying: "Empty callback"))
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
#endif
