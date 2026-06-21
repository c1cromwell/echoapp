// Features/Auth/Enrollment/EnrollmentModels.swift
//
// Shared types for the credential-first enrollment journey.
// Added in v2.5.2: surfaces Mobile Wallet Credential, Driver's License,
// and Phone Number paths from the login screen.
//
// Target: iPhone 17 simulator (iOS 26), Swift 6.0, strict concurrency.

import Foundation

// MARK: - Enrollment Method

/// A path the user can take to enroll in ECHO from the login screen.
/// Ordering of the enum cases reflects promotion order in the picker.
public enum EnrollmentMethod: String, CaseIterable, Identifiable, Sendable {
    case mobileWalletCredential   // OIDC4VC / OID4VP presentation
    case driversLicense           // Apple Wallet mDL, QR, NFC, or IDV fallback
    case phoneNumber              // Existing SMS + DID flow

    public var id: String { rawValue }
}

// MARK: - Driver's License Sub-Method

/// How the user wants to present their driver's license for enrollment.
/// Order matters: `.appleWallet` is tried first when available, then QR, then NFC,
/// and `.scanAndSelfie` is the IDV fallback.
public enum MDLSubMethod: String, CaseIterable, Identifiable, Sendable {
    case appleWallet      // W3C DC API via ASWebAuthenticationSession
    case qrEngagement     // ISO 18013-5 §8.2 QR device engagement
    case nfcEngagement    // ISO 18013-5 §8.3 NFC device engagement
    case scanAndSelfie    // Stripe Identity / Sumsub IAL2 fallback

    public var id: String { rawValue }
}

// MARK: - Requested Claims

/// Minimum claims ECHO requests during enrollment. Bound to the verifier's
/// request manifest and enforced on the backend. Selective disclosure means
/// the wallet releases only these fields, per ISO 18013-5 §7.1.
public struct EnrollmentClaimsRequest: Codable, Sendable {
    let familyName: Bool
    let givenName: Bool
    let ageOver18: Bool
    let issuingCountry: Bool
    let portrait: Bool      // for profile photo — optional

    static let minimumForTier4 = EnrollmentClaimsRequest(
        familyName: true,
        givenName: true,
        ageOver18: true,
        issuingCountry: true,
        portrait: false
    )
}

// MARK: - Verified Identity Bundle

/// The on-device outcome of a successful credential presentation.
/// Never persisted in plaintext outside the Secure Enclave-gated CredentialCache.
public struct VerifiedIdentityBundle: Sendable {
    let credentialReferenceUUID: UUID      // opaque ref stored on-chain
    let issuerDID: String
    let credentialType: String             // e.g. "org.iso.18013.5.1.mDL"
    let issuedAt: Date
    let revocationIndex: UInt32?
    let assuranceLevel: AssuranceLevel
    let disclosedClaims: [String: String]  // name → value; on-device only
    let evidence: PresentationEvidence
}

public enum AssuranceLevel: String, Codable, Sendable {
    case ial1, ial2, ial3

    /// Maps NIST 800-63-3 IAL to ECHO trust tier (1–5).
    var trustTier: Int {
        switch self {
        case .ial1: return 2
        case .ial2: return 4
        case .ial3: return 5
        }
    }
}

/// Protocol-specific proof captured during presentation. Used server-side to
/// re-verify the issuer signature and session transcript.
public enum PresentationEvidence: Sendable {
    case mdoc(deviceResponse: Data, sessionTranscript: Data)  // ISO 18013-5
    case openid4vp(vpToken: String, presentationSubmission: Data)
    case idv(providerReferenceID: String, confidenceScore: Double)
}

// MARK: - Enrollment Errors

public enum EnrollmentError: LocalizedError, Sendable, Equatable {
    case userCancelled
    case deviceUnsupported(reason: String)
    case issuerNotInTrustRegistry(issuerDID: String)
    case credentialRevoked
    case signatureInvalid
    case claimsMissing(required: [String])
    case transportFailed(underlying: String)
    case backendRejected(code: String, message: String)
    case nfcUnavailable
    case cameraUnavailable

    public var errorDescription: String? {
        switch self {
        case .userCancelled:
            return "You cancelled the enrollment."
        case .deviceUnsupported(let reason):
            return "This device can't complete enrollment that way. \(reason)"
        case .issuerNotInTrustRegistry(let did):
            return "The credential was issued by \(did), which isn't in ECHO's trust registry yet."
        case .credentialRevoked:
            return "This credential has been revoked by its issuer."
        case .signatureInvalid:
            return "The credential's signature didn't verify."
        case .claimsMissing(let required):
            return "The credential is missing required claims: \(required.joined(separator: ", "))."
        case .transportFailed(let underlying):
            return "Connection to the credential wallet failed: \(underlying)"
        case .backendRejected(let code, let message):
            switch code {
            case "OIDC4VC_DISABLED":
                return "Wallet enrollment isn't enabled on this server. For local testing, set OIDC4VC_ENABLED=true and restart the API."
            case "SESSION_EXPIRED":
                return "This wallet session expired. Tap Try again to start a new request."
            case "VERIFICATION_FAILED":
                return message.isEmpty ? "The wallet credential didn't verify. Try a different credential or method." : message
            case "PRESENTATION_REQUEST_FAILED":
                return "Couldn't start the wallet request. Check that the API is running and OIDC4VC is configured."
            default:
                return message
            }
        case .nfcUnavailable:
            return "NFC isn't available on this device."
        case .cameraUnavailable:
            return "Camera access is required to scan QR codes."
        }
    }
}

// MARK: - Enrollment Progress

/// Drives the shared tail flow (DID → wallet → passkey) progress UI.
public enum EnrollmentStage: Sendable, Equatable {
    case idle
    case verifyingCredential
    case creatingDID
    case creatingWallet
    case registeringPasskey
    case complete
    case failed(EnrollmentError)
}
