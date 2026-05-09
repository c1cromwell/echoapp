#if os(iOS)
// Features/Auth/Enrollment/DriversLicense/MDLVerifier.swift
//
// ISO 18013-5 reader-side operations. Wraps a third-party SDK such as
// Multipaz (OpenWallet Foundation) or IOWalletProximity (pagopa). The
// ECHO codebase keeps the dependency behind a protocol so the rest of the
// enrollment code is testable and SDK-agnostic.
//
// SDK choice — production: `pagopa/iso18013-ios` (IOWalletProximity) via SPM.
// Alternative: `openwallet-foundation/multipaz` once Swift bindings stabilise.
// See MISSING_FEATURES_GAP_ANALYSIS.md §1 for integration steps.

import Foundation
#if canImport(CoreNFC)
import CoreNFC
#endif
import CryptoKit

// MARK: - Verifier Protocol

protocol MDLVerifierProtocol: Sendable {
    func retrieveOverBLE(
        engagementURI: String,
        requestedClaims: EnrollmentClaimsRequest
    ) async throws -> MDLPresentation

    func readNFCDeviceEngagement(tag: NFCISO7816Tag) async throws -> String
}

struct MDLPresentation: Sendable {
    let deviceResponse: Data
    let sessionTranscript: Data
}

// MARK: - Production Verifier

final class MDLVerifier: MDLVerifierProtocol {

    func retrieveOverBLE(
        engagementURI: String,
        requestedClaims: EnrollmentClaimsRequest
    ) async throws -> MDLPresentation {
        // PRODUCTION INTEGRATION POINT:
        //
        //   import IOWalletProximity
        //
        //   let reader = Proximity.Reader(
        //       readerAuthKey: try loadReaderAuthKey(),
        //       vicalStore: VICALStore.shared
        //   )
        //   let result = try await reader.retrieve(
        //       engagementURI: engagementURI,
        //       request: Proximity.Reader.DocumentRequest(
        //           docType: "org.iso.18013.5.1.mDL",
        //           nameSpaces: ["org.iso.18013.5.1": mapToClaimNames(requestedClaims)]
        //       ),
        //       retrievalTimeout: .seconds(15)
        //   )
        //   return MDLPresentation(
        //       deviceResponse: result.deviceResponseCBOR,
        //       sessionTranscript: result.sessionTranscriptCBOR
        //   )

        throw EnrollmentError.deviceUnsupported(
            reason: "ISO 18013-5 SDK not wired in this build. Link IOWalletProximity via SPM and replace this stub. See MISSING_FEATURES_GAP_ANALYSIS.md §1."
        )
    }

    func readNFCDeviceEngagement(tag: NFCISO7816Tag) async throws -> String {
        // PRODUCTION INTEGRATION POINT — ISO 18013-5 §8.3.2.1.2:
        //
        //   1. SELECT AID d2760000850101 (NDEF application)
        //   2. SELECT file 0001 (NDEF file)
        //   3. READ BINARY the NDEF message
        //   4. Parse NDEF → Device Engagement URI (mdoc:)
        //
        // See MISSING_FEATURES_GAP_ANALYSIS.md §1 for SDK integration steps.

        throw EnrollmentError.deviceUnsupported(
            reason: "NFC engagement parser not wired in this build. See MISSING_FEATURES_GAP_ANALYSIS.md §1."
        )
    }
}

// MARK: - Namespace / claim mapping helpers

enum MDLNamespace {
    static let iso180135 = "org.iso.18013.5.1"
    static let aamva    = "org.iso.18013.5.1.aamva"
}

func mapToClaimNames(_ request: EnrollmentClaimsRequest) -> [String: Bool] {
    var map: [String: Bool] = [:]
    if request.familyName    { map["family_name"]      = true }
    if request.givenName     { map["given_name"]       = true }
    if request.ageOver18     { map["age_over_18"]      = true }
    if request.issuingCountry { map["issuing_country"] = true }
    if request.portrait      { map["portrait"]         = true }
    return map
}
#endif
