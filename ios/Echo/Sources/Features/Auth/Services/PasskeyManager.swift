import AuthenticationServices
import AuthenticationServices
import CryptoKit
#if canImport(UIKit)
import UIKit
#endif

// MARK: - Protocol

protocol PasskeyManagerProtocol {
    func createPasskey(
        challenge: Data,
        userId: Data,
        userName: String
    ) async throws -> PasskeyRegistrationResult

    func authenticateWithPasskey(
        challenge: Data
    ) async throws -> PasskeyAssertionResult
}

// MARK: - Result Types

struct PasskeyRegistrationResult {
    let credentialId: Data
    let rawId: Data
    let clientDataJSON: Data
    let attestationObject: Data
}

struct PasskeyAssertionResult {
    let credentialId: Data
    let rawId: Data
    let clientDataJSON: Data
    let authenticatorData: Data
    let signature: Data
    let userHandle: Data
}

// MARK: - Implementation

final class PasskeyManager: NSObject, PasskeyManagerProtocol,
    ASAuthorizationControllerDelegate,
    ASAuthorizationControllerPresentationContextProviding
{
    private let relyingPartyIdentifier = "echo.app"
    private var authContinuation: CheckedContinuation<ASAuthorization, Error>?

    // MARK: - Registration

    func createPasskey(
        challenge: Data,
        userId: Data,
        userName: String
    ) async throws -> PasskeyRegistrationResult {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )

        let request = provider.createCredentialRegistrationRequest(
            challenge: challenge,
            name: userName,
            userID: userId
        )
        request.attestationPreference = .direct

        let authorization = try await performAuthorizationRequest(request)

        guard let credential = authorization.credential
            as? ASAuthorizationPlatformPublicKeyCredentialRegistration
        else {
            throw AuthError.passkeyCreationFailed
        }

        return PasskeyRegistrationResult(
            credentialId: credential.credentialID,
            rawId: credential.credentialID,
            clientDataJSON: credential.rawClientDataJSON,
            attestationObject: credential.rawAttestationObject ?? Data()
        )
    }

    // MARK: - Assertion (Login)

    func authenticateWithPasskey(
        challenge: Data
    ) async throws -> PasskeyAssertionResult {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )

        let request = provider.createCredentialAssertionRequest(
            challenge: challenge
        )

        let authorization = try await performAuthorizationRequest(request)

        guard let credential = authorization.credential
            as? ASAuthorizationPlatformPublicKeyCredentialAssertion
        else {
            throw AuthError.passkeyAssertionFailed
        }

        return PasskeyAssertionResult(
            credentialId: credential.credentialID,
            rawId: credential.credentialID,
            clientDataJSON: credential.rawClientDataJSON,
            authenticatorData: credential.rawAuthenticatorData,
            signature: credential.signature,
            userHandle: credential.userID
        )
    }

    // MARK: - Authorization Execution

    private func performAuthorizationRequest(
        _ request: ASAuthorizationRequest
    ) async throws -> ASAuthorization {
        try await withCheckedThrowingContinuation { continuation in
            self.authContinuation = continuation
            let controller = ASAuthorizationController(authorizationRequests: [request])
            controller.delegate = self
            controller.presentationContextProvider = self
            controller.performRequests()
        }
    }

    // MARK: - ASAuthorizationControllerDelegate

    func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithAuthorization authorization: ASAuthorization
    ) {
        authContinuation?.resume(returning: authorization)
        authContinuation = nil
    }

    func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithError error: Error
    ) {
        authContinuation?.resume(throwing: error)
        authContinuation = nil
    }

    // MARK: - Presentation Context

    func presentationAnchor(
        for controller: ASAuthorizationController
    ) -> ASPresentationAnchor {
        #if os(iOS)
        guard let scene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let window = scene.windows.first
        else {
            return ASPresentationAnchor()
        }
        return window
        #else
        return ASPresentationAnchor()
        #endif
    }
}

// MARK: - Mock for Testing

#if DEBUG
final class MockPasskeyManager: PasskeyManagerProtocol {
    var createResult: PasskeyRegistrationResult?
    var assertResult: PasskeyAssertionResult?
    var errorToThrow: Error?

    func createPasskey(
        challenge: Data,
        userId: Data,
        userName: String
    ) async throws -> PasskeyRegistrationResult {
        if let error = errorToThrow { throw error }
        return createResult ?? PasskeyRegistrationResult(
            credentialId: Data("mock-cred-id".utf8),
            rawId: Data("mock-raw-id".utf8),
            clientDataJSON: Data("{}".utf8),
            attestationObject: Data()
        )
    }

    func authenticateWithPasskey(
        challenge: Data
    ) async throws -> PasskeyAssertionResult {
        if let error = errorToThrow { throw error }
        return assertResult ?? PasskeyAssertionResult(
            credentialId: Data("mock-cred-id".utf8),
            rawId: Data("mock-raw-id".utf8),
            clientDataJSON: Data("{}".utf8),
            authenticatorData: Data("mock-auth-data".utf8),
            signature: Data("mock-sig".utf8),
            userHandle: Data("mock-user".utf8)
        )
    }
}
#endif
