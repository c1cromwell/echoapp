#if os(iOS)
import Foundation
import CryptoKit

/// Configuration for API client
struct APIConfiguration {
    let baseURL: URL
    let timeout: TimeInterval
    let headers: [String: String]
    
    static let `default` = APIConfiguration(
        baseURL: EchoAPIBaseURL.resolved,
        timeout: 30,
        headers: ["Content-Type": "application/json"]
    )
}

/// HTTP request interceptor protocol
protocol RequestInterceptor {
    func intercept(_ request: inout URLRequest) async throws
}

/// Authentication interceptor that adds auth token to requests
actor AuthenticationInterceptor: RequestInterceptor {
    
    private let keychain: KeychainManager
    
    init(keychain: KeychainManager = .shared) {
        self.keychain = keychain
    }
    
    func intercept(_ request: inout URLRequest) async throws {
        if let token = try await keychain.getAuthToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
    }
}

/// Encryption interceptor for request/response bodies
actor EncryptionInterceptor: RequestInterceptor {

    private let encryption: KinnamiEncryption

    init(encryption: KinnamiEncryption = KinnamiEncryption()) {
        self.encryption = encryption
    }

    func intercept(_ request: inout URLRequest) async throws {
        // Placeholder — future WO for sealed-sender envelope encryption
    }
}

// MARK: - Passkey Signing Interceptor (WO-1)

/// Signs every authenticated request with the device's Secure Enclave P-256 key.
///
/// For each request the interceptor:
///   1. Reads the DID from Keychain.
///   2. Builds the canonical string `METHOD\nPATH\nTIMESTAMP\nhex(SHA-256(body))`.
///   3. Signs SHA-256(canonical) with `SecureEnclaveManager.sign(data:keyId:)`.
///   4. Sets `X-Sender-DID`, `X-Signature` (base64 DER), and `X-Timestamp` headers.
///
/// Binding the method, path, and timestamp into the signature prevents a captured
/// signature from being replayed on another endpoint or after the freshness
/// window (server enforces ±120s and single-use). Requests to public paths
/// (health, identity/register, crypto/server-key) are skipped.
actor PasskeySigningInterceptor: RequestInterceptor {

    private let secureEnclave: SecureEnclaveManager
    private let keychain: KeychainManager
    private static let publicPaths: Set<String> = [
        "/health", "/identity/register", "/v1/crypto/server-key",
        "/identity/devices",   // token-based device add is also public
        "/v1/auth/restore-challenge",
        "/v1/auth/restore-did",
    ]

    init(
        secureEnclave: SecureEnclaveManager = .shared,
        keychain: KeychainManager = .shared
    ) {
        self.secureEnclave = secureEnclave
        self.keychain = keychain
    }

    /// The bytes that get signed: `METHOD\npath\ntimestamp\nhex(SHA-256(body))`, exactly as the
    /// server rebuilds them (internal/api/passkey_auth.go `canonicalSigningString`).
    ///
    /// These are the message, not a hash of it. `SecureEnclaveManager.sign` hashes its input with
    /// SHA-256 as part of ECDSA, and the server hashes the same message once before verifying.
    /// Hashing here as well makes every signature invalid.
    static func signingInput(method: String, path: String, timestamp: String, body: Data) -> Data {
        let bodyHashHex = SHA256.hash(data: body).map { String(format: "%02x", $0) }.joined()
        return Data("\(method)\n\(path)\n\(timestamp)\n\(bodyHashHex)".utf8)
    }

    func intercept(_ request: inout URLRequest) async throws {
        guard let path = request.url?.path,
              !Self.publicPaths.contains(path) else { return }

        guard let did = try? await keychain.retrieve(
            key: "echo.did.current",
            as: String.self
        ) else { return }

        let method = request.httpMethod ?? "GET"
        let timestamp = String(Int(Date().timeIntervalSince1970))
        let message = Self.signingInput(method: method, path: path, timestamp: timestamp, body: request.httpBody ?? Data())

        let sigDER = try await secureEnclave.sign(data: message, keyId: "echo-identity-signing")

        request.setValue(did, forHTTPHeaderField: "X-Sender-DID")
        request.setValue(sigDER.base64EncodedString(), forHTTPHeaderField: "X-Signature")
        request.setValue(timestamp, forHTTPHeaderField: "X-Timestamp")
    }
}

/// REST API client with TLS certificate pinning and passkey request signing.
actor APIClient {

    // MARK: - Properties

    private let configuration: APIConfiguration
    private let session: URLSession
    private var interceptors: [RequestInterceptor] = []

    // MARK: - Initialization

    /// Creates an APIClient whose URLSession uses `pinner` as its delegate.
    /// Pass `nil` to disable certificate pinning (dev/test builds).
    ///
    /// Pass interceptors here rather than adding them afterwards: `addInterceptor` runs on the
    /// actor, so adding the request signer in a detached task let the first few requests after
    /// launch leave unsigned and be refused.
    init(configuration: APIConfiguration = .default, pinner: CertificatePinner? = nil, interceptors: [RequestInterceptor] = []) {
        self.configuration = configuration
        self.interceptors = interceptors

        let sessionConfig = URLSessionConfiguration.default
        sessionConfig.timeoutIntervalForRequest = configuration.timeout
        sessionConfig.timeoutIntervalForResource = configuration.timeout * 2
        sessionConfig.waitsForConnectivity = true
        sessionConfig.tlsMinimumSupportedProtocolVersion = .TLSv13
        TransportProxySettings.apply(to: sessionConfig)

        if let pinner = pinner {
            self.session = URLSession(configuration: sessionConfig, delegate: pinner, delegateQueue: nil)
        } else {
            self.session = URLSession(configuration: sessionConfig)
        }
    }
    
    // MARK: - Interceptor Management
    
    func addInterceptor(_ interceptor: RequestInterceptor) {
        interceptors.append(interceptor)
    }
    
    func removeAllInterceptors() {
        interceptors.removeAll()
    }
    
    // MARK: - Request Building
    
    private func buildRequest(
        endpoint: APIEndpoint,
        method: HTTPMethod,
        body: Encodable? = nil
    ) async throws -> URLRequest {
        let url = Self.requestURL(base: configuration.baseURL, path: endpoint.path)
        var request = URLRequest(url: url)
        
        request.httpMethod = method.rawValue
        
        // Add default headers
        for (key, value) in configuration.headers {
            request.setValue(value, forHTTPHeaderField: key)
        }
        
        // Add endpoint-specific headers
        for (key, value) in endpoint.headers {
            request.setValue(value, forHTTPHeaderField: key)
        }
        
        // Encode request body if provided
        if let body = body {
            let encoder = JSONEncoder()
            request.httpBody = try encoder.encode(body)
        }
        
        // Apply interceptors
        for interceptor in interceptors {
            try await interceptor.intercept(&request)
        }
        
        return request
    }
    
    // MARK: - GET Request
    
    func get<T: Decodable>(
        endpoint: APIEndpoint
    ) async throws -> T {
        let request = try await buildRequest(endpoint: endpoint, method: .get)
        return try await performRequest(request)
    }
    
    // MARK: - POST Request
    
    func post<T: Decodable, B: Encodable>(
        endpoint: APIEndpoint,
        body: B
    ) async throws -> T {
        let request = try await buildRequest(endpoint: endpoint, method: .post, body: body)
        return try await performRequest(request)
    }

    /// POST with additional per-request headers (e.g. X-Wallet-Proof for
    /// real-funds custody). Headers are applied after interceptors so they are
    /// not part of the passkey-signed canonical string.
    func post<T: Decodable, B: Encodable>(
        endpoint: APIEndpoint,
        body: B,
        headers: [String: String]
    ) async throws -> T {
        var request = try await buildRequest(endpoint: endpoint, method: .post, body: body)
        for (key, value) in headers {
            request.setValue(value, forHTTPHeaderField: key)
        }
        return try await performRequest(request)
    }

    /// POST an arbitrary JSON object (used when the body embeds a client-signed
    /// {value, proofs} that isn't statically Codable). Mirrors buildRequest's
    /// default + endpoint headers and interceptors, then applies extra headers.
    func postJSON<T: Decodable>(
        endpoint: APIEndpoint,
        json: [String: Any],
        headers: [String: String] = [:]
    ) async throws -> T {
        let url = Self.requestURL(base: configuration.baseURL, path: endpoint.path)
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.post.rawValue
        for (key, value) in configuration.headers { request.setValue(value, forHTTPHeaderField: key) }
        for (key, value) in endpoint.headers { request.setValue(value, forHTTPHeaderField: key) }
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: json)
        for interceptor in interceptors { try await interceptor.intercept(&request) }
        for (key, value) in headers { request.setValue(value, forHTTPHeaderField: key) }
        return try await performRequest(request)
    }

    // MARK: - PUT Request
    
    func put<T: Decodable, B: Encodable>(
        endpoint: APIEndpoint,
        body: B
    ) async throws -> T {
        let request = try await buildRequest(endpoint: endpoint, method: .put, body: body)
        return try await performRequest(request)
    }
    
    // MARK: - DELETE Request
    
    func delete<T: Decodable>(
        endpoint: APIEndpoint
    ) async throws -> T {
        let request = try await buildRequest(endpoint: endpoint, method: .delete)
        return try await performRequest(request)
    }

    func delete<T: Decodable>(
        endpoint: APIEndpoint,
        headers: [String: String]
    ) async throws -> T {
        var request = try await buildRequest(endpoint: endpoint, method: .delete)
        for (key, value) in headers {
            request.setValue(value, forHTTPHeaderField: key)
        }
        return try await performRequest(request)
    }
    
    // MARK: - PATCH Request
    
    func patch<T: Decodable, B: Encodable>(
        endpoint: APIEndpoint,
        body: B
    ) async throws -> T {
        let request = try await buildRequest(endpoint: endpoint, method: .patch, body: body)
        return try await performRequest(request)
    }
    
    // MARK: - Upload Request
    
    func upload<T: Decodable>(
        endpoint: APIEndpoint,
        data: Data,
        filename: String,
        mimeType: String
    ) async throws -> T {
        let url = Self.requestURL(base: configuration.baseURL, path: endpoint.path)
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.post.rawValue
        
        let boundary = UUID().uuidString
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        
        var body = Data()
        
        // Add file data
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(filename)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: \(mimeType)\r\n\r\n".data(using: .utf8)!)
        body.append(data)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)
        
        request.httpBody = body
        
        return try await performRequest(request)
    }

    // MARK: - Raw binary (media upload / chunk download)

    func postRaw<T: Decodable>(
        endpoint: APIEndpoint,
        body: Data,
        extraHeaders: [String: String] = [:]
    ) async throws -> T {
        var request = try await buildRequest(endpoint: endpoint, method: .post)
        request.httpBody = body
        request.setValue(extraHeaders["Content-Type"] ?? "application/octet-stream", forHTTPHeaderField: "Content-Type")
        for (key, value) in extraHeaders where key != "Content-Type" {
            request.setValue(value, forHTTPHeaderField: key)
        }
        for interceptor in interceptors {
            try await interceptor.intercept(&request)
        }
        return try await performRequest(request)
    }

    func putRaw(
        endpoint: APIEndpoint,
        body: Data,
        extraHeaders: [String: String] = [:]
    ) async throws {
        var request = try await buildRequest(endpoint: endpoint, method: .put)
        request.httpBody = body
        request.setValue(extraHeaders["Content-Type"] ?? "application/octet-stream", forHTTPHeaderField: "Content-Type")
        for (key, value) in extraHeaders where key != "Content-Type" {
            request.setValue(value, forHTTPHeaderField: key)
        }
        for interceptor in interceptors {
            try await interceptor.intercept(&request)
        }
        let (_, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        try validateResponse(httpResponse)
    }

    func getRaw(endpoint: APIEndpoint) async throws -> Data {
        var request = try await buildRequest(endpoint: endpoint, method: .get)
        for interceptor in interceptors {
            try await interceptor.intercept(&request)
        }
        let (data, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        try validateResponse(httpResponse)
        return data
    }
    
    // MARK: - Response Handling
    
    private func performRequest<T: Decodable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        
        try validateResponse(httpResponse)
        
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        
        return try decoder.decode(T.self, from: data)
    }
    
    private func validateResponse(_ response: HTTPURLResponse) throws {
        switch response.statusCode {
        case 200...299:
            break
        case 400:
            throw APIError.badRequest
        case 401:
            throw APIError.unauthorized
        case 403:
            throw APIError.forbidden
        case 404:
            throw APIError.notFound
        case 429:
            // WO-44: read Retry-After header (seconds) set by the backend rate limiter.
            let retryAfter = response.value(forHTTPHeaderField: "Retry-After")
                .flatMap(TimeInterval.init) ?? 60
            throw APIError.rateLimited(retryAfter: retryAfter)
        case 500...599:
            throw APIError.serverError(response.statusCode)
        default:
            throw APIError.unexpectedStatusCode(response.statusCode)
        }
    }
}

// MARK: - HTTP Methods

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case patch = "PATCH"
    case delete = "DELETE"
    case head = "HEAD"
    case options = "OPTIONS"
}

// MARK: - API Endpoint Protocol

protocol APIEndpoint {
    var path: String { get }
    var headers: [String: String] { get }
}

extension APIEndpoint {
    var headers: [String: String] {
        [:]
    }
}

extension APIClient {
    /// The URL for an endpoint path that may carry a query string (`/v3/contacts/search?handle=bob`).
    ///
    /// `appendingPathComponent` treats the whole string as one path segment and percent-encodes
    /// the `?`, so the server is asked for a route that does not exist. The path and the query
    /// have to be attached separately.
    static func requestURL(base: URL, path: String) -> URL {
        // Some callers percent-encode an identifier before putting it in the path (a DID's ":"
        // becomes "%3A"). Appending that as-is encodes the "%" a second time and the server is
        // asked about an identifier that does not exist. Decode first; appending encodes once.
        func append(_ rawPath: String) -> URL {
            base.appendingPathComponent(rawPath.removingPercentEncoding ?? rawPath)
        }
        guard let mark = path.firstIndex(of: "?") else {
            return append(path)
        }
        let pathPart = String(path[..<mark])
        let query = String(path[path.index(after: mark)...])
        let url = append(pathPart)
        guard var components = URLComponents(url: url, resolvingAgainstBaseURL: false) else { return url }
        // Callers may already have percent-encoded values; keep those, encode anything else.
        if query.removingPercentEncoding != nil, query.range(of: "%") != nil {
            components.percentEncodedQuery = query
        } else {
            components.query = query
        }
        return components.url ?? url
    }
}

// MARK: - API Errors

enum APIError: LocalizedError {
    case invalidResponse
    case badRequest
    case unauthorized
    case forbidden
    case notFound
    /// WO-44: 429 Too Many Requests; retryAfter is the value of the Retry-After header.
    case rateLimited(retryAfter: TimeInterval)
    case serverError(Int)
    case unexpectedStatusCode(Int)
    case networkError(URLError)
    case decodingError(DecodingError)
    case encodingError(EncodingError)

    /// Shown to people as-is wherever a screen surfaces `error.localizedDescription`,
    /// so it says what happened and what to do — no HTTP jargon. Status codes for
    /// logs and support live in `diagnosticDescription`.
    var errorDescription: String? {
        switch self {
        case .invalidResponse, .decodingError, .encodingError:
            return "Something went wrong talking to Echo. Please try again."
        case .badRequest:
            return "Echo couldn't process that request. Please try again."
        case .unauthorized:
            return "Your session has expired. Sign in again to continue."
        case .forbidden:
            return "You don't have permission to do that."
        case .notFound:
            return "We couldn't find that. It may have been removed."
        case .rateLimited(let retryAfter):
            let seconds = Int(retryAfter.rounded(.up))
            return seconds > 0
                ? "Echo is busy. Try again in \(seconds) second\(seconds == 1 ? "" : "s")."
                : "Echo is busy. Try again in a moment."
        case .serverError, .unexpectedStatusCode:
            return "Echo is having trouble right now. Please try again in a moment."
        case .networkError(let error):
            switch error.code {
            case .notConnectedToInternet, .networkConnectionLost, .dataNotAllowed:
                return "You're offline. Check your connection and try again."
            case .timedOut:
                return "The connection timed out. Please try again."
            default:
                return "Couldn't reach Echo. Check your connection and try again."
            }
        }
    }

    /// Technical detail (status codes, underlying error) for logs — never for UI copy.
    var diagnosticDescription: String {
        switch self {
        case .invalidResponse: return "invalid response"
        case .badRequest: return "HTTP 400"
        case .unauthorized: return "HTTP 401"
        case .forbidden: return "HTTP 403"
        case .notFound: return "HTTP 404"
        case .rateLimited(let retryAfter): return "HTTP 429 (retry after \(Int(retryAfter))s)"
        case .serverError(let code): return "HTTP \(code)"
        case .unexpectedStatusCode(let code): return "unexpected HTTP \(code)"
        case .networkError(let error): return "network error \(error.code.rawValue): \(error.localizedDescription)"
        case .decodingError(let error): return "decoding error: \(error)"
        case .encodingError(let error): return "encoding error: \(error)"
        }
    }
}
#endif
