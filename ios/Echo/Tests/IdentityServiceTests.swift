import XCTest
@testable import Echo

final class IdentityServiceTests: XCTestCase {
    
    var service: IdentityService!
    
    override func setUp() {
        super.setUp()
        service = IdentityService()
    }
    
    @MainActor
    func testRegisterUser() async throws {
        let user = try await service.registerUser(phoneHash: "hash123")
        
        XCTAssertFalse(user.id.isEmpty)
        XCTAssertEqual(user.phoneHash, "hash123")
        XCTAssertEqual(user.verificationLevel, 0)
        XCTAssertEqual(user.status, "active")
    }
    
    @MainActor
    func testRegisterUserEmptyPhone() async throws {
        do {
            _ = try await service.registerUser(phoneHash: "")
            XCTFail("Should throw invalidPhoneHash error")
        } catch IdentityError.invalidPhoneHash {
            // Expected
        }
    }
    
    @MainActor
    func testGetUser() async throws {
        let registered = try await service.registerUser(phoneHash: "hash123")
        let retrieved = try await service.getUser(registered.id)
        
        XCTAssertEqual(retrieved.id, registered.id)
        XCTAssertEqual(retrieved.phoneHash, "hash123")
    }
    
    @MainActor
    func testGetUserNotFound() async throws {
        do {
            _ = try await service.getUser("nonexistent")
            XCTFail("Should throw userNotFound error")
        } catch IdentityError.userNotFound {
            // Expected
        }
    }
    
    @MainActor
    func testUpdateVerificationLevel() async throws {
        let user = try await service.registerUser(phoneHash: "hash123")
        
        try await service.updateVerificationLevel(user.id, level: 3)
        
        let updated = try await service.getUser(user.id)
        XCTAssertEqual(updated.verificationLevel, 3)
    }
    
    @MainActor
    func testUpdateVerificationInvalidLevel() async throws {
        let user = try await service.registerUser(phoneHash: "hash123")
        
        do {
            try await service.updateVerificationLevel(user.id, level: 10)
            XCTFail("Should throw invalidVerificationLevel error")
        } catch IdentityError.invalidVerificationLevel {
            // Expected
        }
    }
    
    @MainActor
    func testUpdateMetagraphAddress() async throws {
        let user = try await service.registerUser(phoneHash: "hash123")
        
        try await service.updateMetagraphAddress(user.id, address: "mgraph-addr-123")
        
        let updated = try await service.getUser(user.id)
        XCTAssertEqual(updated.metagraphAddr, "mgraph-addr-123")
    }
    
    @MainActor
    func testMultipleUsers() async throws {
        let user1 = try await service.registerUser(phoneHash: "hash1")
        let user2 = try await service.registerUser(phoneHash: "hash2")
        
        XCTAssertNotEqual(user1.id, user2.id)
        
        let retrieved1 = try await service.getUser(user1.id)
        XCTAssertEqual(retrieved1.phoneHash, "hash1")
        
        let retrieved2 = try await service.getUser(user2.id)
        XCTAssertEqual(retrieved2.phoneHash, "hash2")
    }
}
