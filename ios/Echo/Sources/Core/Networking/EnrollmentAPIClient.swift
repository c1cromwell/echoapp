#if os(iOS)
// Core/Networking/EnrollmentAPIClient.swift
//
// Typed wrapper around the /v1/enrollment/* endpoints on the Go backend.
// Replace `URLSession.shared` calls with the project's existing
// certificate-pinned client once the backend is live.

import Foundation

actor EnrollmentAPIClient {
    static let shared = EnrollmentAPIClient()

    private let baseURL = URL(string: "https://api.echo.app")!
    private let session: URLSession = .shared
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init() {
        self.decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        self.encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
    }

    // MARK: - Wallet credential (OIDC4VC)

    struct WalletSession: Decodable, Sendable {
        let id: String
        let verifierURL: URL
        let expiresAt: Date
    }

    func startWalletPresentation(claims: EnrollmentClaimsRequest) async throws -> WalletSession {
        try await post(
            path: "/v1/enrollment/vc/start",
            body: ["requested_claims": claims]
        )
    }

    func finishWalletPresentation(sessionID: String, callbackURL: URL) async throws -> VerifiedIdentityBundle {
        let bundle: WireBundle = try await post(
            path: "/v1/enrollment/vc/finish",
            body: FinishWalletRequest(
                session_id: sessionID,
                callback_url: callbackURL.absoluteString
            )
        )
        return try bundle.toDomain()
    }

    // MARK: - mDL presentation (DC API, QR/BLE, NFC/BLE)

    enum MDLTransport: String, Encodable, Sendable {
        case webDCAPI = "web_dcapi"
        case qrBLE    = "qr_ble"
        case nfcBLE   = "nfc_ble"
    }

    struct MDLStartRequest: Encodable {
        let transport: MDLTransport
        let requested_claims: EnrollmentClaimsRequest
    }

    struct MDLSession: Decodable, Sendable {
        let id: String
        let verifierURL: URL
        let expiresAt: Date
    }

    func startMDLPresentation(
        transport: MDLTransport,
        claims: EnrollmentClaimsRequest
    ) async throws -> MDLSession {
        try await post(
            path: "/v1/enrollment/mdl/start",
            body: MDLStartRequest(transport: transport, requested_claims: claims)
        )
    }

    struct MDLFinishRequest: Encodable {
        let session_id: String?
        let transport: MDLTransport
        let device_response_b64: String?
        let session_transcript_b64: String?
        let callback_url: String?
    }

    func finishMDLPresentation(
        sessionID: String?,
        callbackURL: URL
    ) async throws -> VerifiedIdentityBundle {
        let body = MDLFinishRequest(
            session_id: sessionID,
            transport: .webDCAPI,
            device_response_b64: nil,
            session_transcript_b64: nil,
            callback_url: callbackURL.absoluteString
        )
        let bundle: WireBundle = try await post(path: "/v1/enrollment/mdl/finish", body: body)
        return try bundle.toDomain()
    }

    func finishMDLPresentation(
        sessionID: String?,
        deviceResponse: Data,
        sessionTranscript: Data,
        transport: MDLTransport
    ) async throws -> VerifiedIdentityBundle {
        let body = MDLFinishRequest(
            session_id: sessionID,
            transport: transport,
            device_response_b64: deviceResponse.base64EncodedString(),
            session_transcript_b64: sessionTranscript.base64EncodedString(),
            callback_url: nil
        )
        let bundle: WireBundle = try await post(path: "/v1/enrollment/mdl/finish", body: body)
        return try bundle.toDomain()
    }

    // MARK: - IDV fallback

    struct IDVSession: Decodable, Sendable {
        let id: String
        let clientSecret: String
        let provider: String   // "stripe_identity" | "sumsub"
    }

    func startIDVSession() async throws -> IDVSession {
        try await post(path: "/v1/enrollment/idv/start", body: EmptyBody())
    }

    func awaitIDVResult(sessionID: String) async throws -> VerifiedIdentityBundle {
        let bundle: WireBundle = try await post(
            path: "/v1/enrollment/idv/await",
            body: ["session_id": sessionID]
        )
        return try bundle.toDomain()
    }

    // MARK: - Shared tail

    func registerDIDFromBundle(_ bundle: VerifiedIdentityBundle) async throws -> String {
        struct R: Decodable { let did: String }
        let r: R = try await post(
            path: "/v1/enrollment/did",
            body: ["credential_reference": bundle.credentialReferenceUUID.uuidString]
        )
        return r.did
    }

    func createWalletForDID(did: String) async throws -> String {
        struct R: Decodable { let address: String }
        let r: R = try await post(path: "/v1/enrollment/wallet", body: ["did": did])
        return r.address
    }

    func registerPasskey(did: String) async throws {
        let _: EmptyResponse = try await post(path: "/v1/enrollment/passkey", body: ["did": did])
    }

    // MARK: - Transport helpers

    private struct EmptyBody: Encodable {}
    private struct EmptyResponse: Decodable {}

    private func post<B: Encodable, R: Decodable>(path: String, body: B) async throws -> R {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try encoder.encode(body)

        do {
            let (data, response) = try await session.data(for: request)
            guard let http = response as? HTTPURLResponse else {
                throw EnrollmentError.transportFailed(underlying: "No HTTP response")
            }

            if http.statusCode >= 400 {
                if let wire = try? decoder.decode(ErrorEnvelope.self, from: data) {
                    throw EnrollmentError.backendRejected(code: wire.code, message: wire.message)
                }
                throw EnrollmentError.backendRejected(
                    code: "HTTP_\(http.statusCode)",
                    message: "Server returned status \(http.statusCode)."
                )
            }

            return try decoder.decode(R.self, from: data)
        } catch let e as EnrollmentError {
            throw e
        } catch {
            throw EnrollmentError.transportFailed(underlying: error.localizedDescription)
        }
    }

    private struct ErrorEnvelope: Decodable {
        let code: String
        let message: String
    }
}

// MARK: - Concrete request/response types

private struct FinishWalletRequest: Encodable {
    let session_id: String
    let callback_url: String
}

// MARK: - Wire shape for VerifiedIdentityBundle

private struct WireBundle: Decodable {
    let credential_reference_uuid: String
    let issuer_did: String
    let credential_type: String
    let issued_at: Date
    let revocation_index: UInt32?
    let assurance_level: String                  // "ial1" | "ial2" | "ial3"
    let disclosed_claims: [String: String]
    let evidence: Evidence

    struct Evidence: Decodable {
        let kind: String                          // "mdoc" | "openid4vp" | "idv"
        let device_response_b64: String?
        let session_transcript_b64: String?
        let vp_token: String?
        let presentation_submission_b64: String?
        let provider_reference_id: String?
        let confidence_score: Double?
    }

    func toDomain() throws -> VerifiedIdentityBundle {
        guard let uuid = UUID(uuidString: credential_reference_uuid) else {
            throw EnrollmentError.backendRejected(code: "BAD_UUID", message: "Malformed credential reference.")
        }
        guard let ial = AssuranceLevel(rawValue: assurance_level) else {
            throw EnrollmentError.backendRejected(code: "BAD_IAL", message: "Unknown assurance level.")
        }

        let evidenceDomain: PresentationEvidence
        switch evidence.kind {
        case "mdoc":
            guard
                let dr = evidence.device_response_b64.flatMap({ Data(base64Encoded: $0) }),
                let st = evidence.session_transcript_b64.flatMap({ Data(base64Encoded: $0) })
            else {
                throw EnrollmentError.backendRejected(code: "BAD_EVIDENCE", message: "Missing mdoc evidence.")
            }
            evidenceDomain = .mdoc(deviceResponse: dr, sessionTranscript: st)
        case "openid4vp":
            guard
                let token = evidence.vp_token,
                let submission = evidence.presentation_submission_b64.flatMap({ Data(base64Encoded: $0) })
            else {
                throw EnrollmentError.backendRejected(code: "BAD_EVIDENCE", message: "Missing OID4VP evidence.")
            }
            evidenceDomain = .openid4vp(vpToken: token, presentationSubmission: submission)
        case "idv":
            guard
                let ref = evidence.provider_reference_id,
                let score = evidence.confidence_score
            else {
                throw EnrollmentError.backendRejected(code: "BAD_EVIDENCE", message: "Missing IDV evidence.")
            }
            evidenceDomain = .idv(providerReferenceID: ref, confidenceScore: score)
        default:
            throw EnrollmentError.backendRejected(code: "UNKNOWN_EVIDENCE", message: "Unknown evidence kind.")
        }

        return VerifiedIdentityBundle(
            credentialReferenceUUID: uuid,
            issuerDID: issuer_did,
            credentialType: credential_type,
            issuedAt: issued_at,
            revocationIndex: revocation_index,
            assuranceLevel: ial,
            disclosedClaims: disclosed_claims,
            evidence: evidenceDomain
        )
    }
}
#endif
