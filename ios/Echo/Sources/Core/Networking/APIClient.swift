import Foundation

/// Configuration for API client
struct APIConfiguration {
    let baseURL: URL
    let timeout: TimeInterval
    let headers: [String: String]
    
    static let `default` = APIConfiguration(
        baseURL: URL(string: "https://api.echo.local")!,
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
        // Encryption can be applied here if needed
        // For now, this is a placeholder for future E2E request encryption
    }
}

/// REST API client with request/response handling
actor APIClient {
    
    // MARK: - Properties
    
    private let configuration: APIConfiguration
    private let session: URLSession
    private var interceptors: [RequestInterceptor] = []
    
    // MARK: - Initialization
    
    init(configuration: APIConfiguration = .default) {
        self.configuration = configuration
        
        let sessionConfig = URLSessionConfiguration.default
        sessionConfig.timeoutIntervalForRequest = configuration.timeout
        sessionConfig.timeoutIntervalForResource = configuration.timeout * 2
        sessionConfig.waitsForConnectivity = true
        sessionConfig.tlsMinimumSupportedProtocolVersion = .TLSv13
        
        self.session = URLSession(configuration: sessionConfig)
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
        let url = configuration.baseURL.appendingPathComponent(endpoint.path)
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
        let url = configuration.baseURL.appendingPathComponent(endpoint.path)
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

// MARK: - API Errors

enum APIError: LocalizedError {
    case invalidResponse
    case badRequest
    case unauthorized
    case forbidden
    case notFound
    case serverError(Int)
    case unexpectedStatusCode(Int)
    case networkError(URLError)
    case decodingError(DecodingError)
    case encodingError(EncodingError)
    
    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "Invalid response from server"
        case .badRequest:
            return "Bad request (400)"
        case .unauthorized:
            return "Unauthorized (401) - Please authenticate"
        case .forbidden:
            return "Forbidden (403)"
        case .notFound:
            return "Resource not found (404)"
        case .serverError(let code):
            return "Server error (\(code))"
        case .unexpectedStatusCode(let code):
            return "Unexpected status code (\(code))"
        case .networkError(let error):
            return "Network error: \(error.localizedDescription)"
        case .decodingError:
            return "Failed to decode response"
        case .encodingError:
            return "Failed to encode request"
        }
    }
}
