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
    
    private func generateDID() -> String {
        return "did:echo:" + UUID().uuidString.lowercased()
    }
}

/// Identity service errors
public enum IdentityError: LocalizedError {
    case invalidPhoneHash
    case userNotFound
    case invalidVerificationLevel
    
    public var errorDescription: String? {
        switch self {
        case .invalidPhoneHash:
            return "Phone hash is required and cannot be empty"
        case .userNotFound:
            return "User not found"
        case .invalidVerificationLevel:
            return "Verification level must be between 0 and 5"
        }
    }
}
