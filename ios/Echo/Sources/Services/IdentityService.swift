import Foundation

/// Identity service for authentication and user management
public class IdentityService: BaseEchoService {
    private var users: [String: UserIdentity] = [:]
    
    public override init(name: String = "identity-service", version: String = "v1") {
        super.init(name: name, version: version)
    }
    
    /// User identity model
    public struct UserIdentity: Identifiable {
        public let id: String
        public let did: String
        public let phoneHash: String
        public var verificationLevel: Int = 0
        public var status: String = "active"
        public var metagraphAddr: String? = nil
    }

    /// Response from `POST /identity/register` (Echo API, Phase 1 / ADR-0001).
    public struct IdentityRegisterHTTPResponse: Decodable {
        public let did: String
        public let publicKeyHex: String
        public let registeredAt: String
        public let existing: Bool

        enum CodingKeys: String, CodingKey {
            case did
            case publicKeyHex = "public_key_hex"
            case registeredAt = "registered_at"
            case existing
        }
    }
    
    // MARK: - User Management
    
    /// Register a new user
    public func registerUser(phoneHash: String) async throws -> UserIdentity {
        guard !phoneHash.isEmpty else {
            throw IdentityError.invalidPhoneHash
        }
        
        let userID = generateUserID()
        let user = UserIdentity(
            id: userID,
            did: generateDID(),
            phoneHash: phoneHash
        )
        
        users[userID] = user
        return user
    }

    /// Registers a derived `did:key` and P-256 public key with the Echo backend (`POST /identity/register`).
    /// - Parameters:
    ///   - baseURL: API root without trailing slash, e.g. `https://api.example.com` or `http://localhost:8000`
    ///   - did: Canonical `did:key:z…` from the same public key material
    ///   - publicKeyHex: Lowercase hex, SEC1 uncompressed P-256 (`04` + 64 hex bytes)
    public func registerDidKeyOnEchoAPI(baseURL: URL, did: String, publicKeyHex: String) async throws -> IdentityRegisterHTTPResponse {
        let url = baseURL.appendingPathComponent("identity/register")
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let body: [String: String] = [
            "did": did,
            "public_key_hex": publicKeyHex,
        ]
        req.httpBody = try JSONSerialization.data(withJSONObject: body)

        let (data, resp) = try await URLSession.shared.data(for: req)
        guard let http = resp as? HTTPURLResponse else {
            throw IdentityError.httpError("no HTTP response")
        }
        guard (200...201).contains(http.statusCode) else {
            let msg = String(data: data, encoding: .utf8) ?? ""
            throw IdentityError.httpError("HTTP \(http.statusCode): \(msg)")
        }
        let dec = JSONDecoder()
        return try dec.decode(IdentityRegisterHTTPResponse.self, from: data)
    }
    
    /// Get user by ID
    public func getUser(_ userID: String) async throws -> UserIdentity {
        guard let user = users[userID] else {
            throw IdentityError.userNotFound
        }
        return user
    }
    
    /// Update verification level
    public func updateVerificationLevel(_ userID: String, level: Int) async throws {
        guard var user = users[userID] else {
            throw IdentityError.userNotFound
        }
        
        guard level >= 0 && level <= 5 else {
            throw IdentityError.invalidVerificationLevel
        }
        
        user.verificationLevel = level
        users[userID] = user
    }
    
    /// Update metagraph address
    public func updateMetagraphAddress(_ userID: String, address: String) async throws {
        guard var user = users[userID] else {
            throw IdentityError.userNotFound
        }
        
        user.metagraphAddr = address
        users[userID] = user
    }
    
    // MARK: - Helper Methods
    
    private func generateUserID() -> String {
        return "user-" + UUID().uuidString.prefix(12).lowercased()
    }
    
    /// Local-only placeholder DID for in-memory `registerUser` (not a W3C did:key).
    private func generateDID() -> String {
        return "did:echo:" + UUID().uuidString.lowercased()
    }
}

/// Identity service errors
public enum IdentityError: LocalizedError {
    case invalidPhoneHash
    case userNotFound
    case invalidVerificationLevel
    case httpError(String)
    
    public var errorDescription: String? {
        switch self {
        case .invalidPhoneHash:
            return "Phone hash is required and cannot be empty"
        case .userNotFound:
            return "User not found"
        case .invalidVerificationLevel:
            return "Verification level must be between 0 and 5"
        case .httpError(let s):
            return s
        }
    }
}
