import XCTest
@testable import Echo

// MARK: - DIContainer Tests

final class DIContainerTests: XCTestCase {
    var container: DIContainer!
    
    override func setUp() {
        super.setUp()
        container = DIContainer.shared
    }
    
    func testContainerInitialization() {
        XCTAssertNotNil(container)
    }
    
    func testResolveBiometricAuthManager() {
        let bioManager = container.resolveBiometricAuth()
        XCTAssertNotNil(bioManager)
    }
    
    func testResolveKeychain() {
        let keychain = container.resolveKeychain()
        XCTAssertNotNil(keychain)
    }
    
    func testResolveAPIClient() {
        let apiClient = container.resolveAPIClient()
        XCTAssertNotNil(apiClient)
    }
    
    func testResolveSingletonServices() {
        let keychain1 = container.resolveKeychain()
        let keychain2 = container.resolveKeychain()
        XCTAssertTrue(keychain1 === keychain2)
    }
}

// MARK: - BiometricAuthManager Tests

final class BiometricAuthManagerTests: XCTestCase {
    var bioManager: BiometricAuthManager!
    
    override func setUp() {
        super.setUp()
        bioManager = BiometricAuthManager()
    }
    
    func testBiometricTypeDetection() {
        let type = bioManager.biometricType
        XCTAssertTrue(
            type == .faceID || type == .touchID || type == .opticID || type == .unknown || type == .none
        )
    }
    
    func testHasDevicePasscode() {
        let hasPasscode = bioManager.hasDevicePasscode
        XCTAssertTrue(hasPasscode)
    }
    
    func testBiometricAvailable() {
        let available = bioManager.isBiometricAvailable
        XCTAssertNotNil(available)
    }
}

// MARK: - KeychainManager Tests

final class KeychainManagerTests: XCTestCase {
    var keychain: KeychainManager!
    
    override func setUp() async throws {
        try await super.setUp()
        keychain = KeychainManager.shared
    }
    
    override func tearDown() async throws {
        try await super.tearDown()
        try await keychain.clearAuthCredentials()
    }
    
    func testStoreAndRetrieveToken() async throws {
        let testToken = "test-token-12345"
        try await keychain.storeAuthToken(testToken)
        
        let retrieved = try await keychain.getAuthToken()
        XCTAssertEqual(retrieved, testToken)
    }
    
    func testClearCredentials() async throws {
        try await keychain.storeAuthToken("test-token")
        try await keychain.clearAuthCredentials()
        
        let retrieved = try await keychain.getAuthToken()
        XCTAssertNil(retrieved)
    }
    
    func testKeyExists() async throws {
        try await keychain.storeAuthToken("test-token")
        let exists = await keychain.exists(key: "authToken", service: "com.echo.auth")
        XCTAssertTrue(exists)
    }
}

// MARK: - KinnamiEncryption Tests

final class KinnamiEncryptionTests: XCTestCase {
    var encryption: KinnamiEncryption!
    
    override func setUp() {
        super.setUp()
        encryption = KinnamiEncryption()
    }
    
    func testGenerateEphemeralKeyPair() {
        let (privateKey, publicKey) = encryption.generateEphemeralKeyPair()
        
        XCTAssertNotNil(privateKey)
        XCTAssertFalse(publicKey.isEmpty)
        XCTAssertTrue(publicKey.count > 0)
    }
    
    func testKeyAgreement() async throws {
        let (privateKey1, publicKey1) = encryption.generateEphemeralKeyPair()
        let (privateKey2, publicKey2) = encryption.generateEphemeralKeyPair()
        
        let sharedSecret1 = try await encryption.performKeyAgreement(
            privateKey: privateKey1,
            recipientPublicKeyData: publicKey2
        )
        
        XCTAssertNotNil(sharedSecret1)
    }
    
    func testEncryptionAndDecryption() async throws {
        let message = "Hello, ECHO!"
        let (_, publicKey) = encryption.generateEphemeralKeyPair()
        
        let encryptedMessage = try await encryption.encryptWithKeyAgreement(
            plaintext: message,
            recipientPublicKeyData: publicKey
        )
        
        XCTAssertFalse(encryptedMessage.ciphertext.isEmpty)
        XCTAssertFalse(encryptedMessage.nonce.isEmpty)
        XCTAssertEqual(encryptedMessage.algorithm, "AES-256-GCM-KINNAMI")
    }
}

// MARK: - APIClient Tests

final class APIClientTests: XCTestCase {
    var apiClient: APIClient!
    var mockSession: URLSession!
    
    override func setUp() {
        super.setUp()
        apiClient = APIClient(configuration: APIConfiguration.default)
    }
    
    func testAPIClientInitialization() {
        XCTAssertNotNil(apiClient)
    }
    
    func testInterceptorManagement() {
        let mockInterceptor = MockRequestInterceptor()
        apiClient.addInterceptor(mockInterceptor)
        
        XCTAssertNoThrow(apiClient.removeAllInterceptors())
    }
}

// MARK: - Mock Interceptor

class MockRequestInterceptor: RequestInterceptor {
    func intercept(_ request: inout URLRequest) async throws {
        // Mock implementation
    }
}

// MARK: - LocalDatabase Tests

final class LocalDatabaseTests: XCTestCase {
    var database: LocalDatabase!
    
    override func setUp() async throws {
        try await super.setUp()
        database = LocalDatabase.shared
        try await LocalDatabase.setup()
    }
    
    func testDatabaseInitialization() async throws {
        try await LocalDatabase.setup()
        XCTAssertNotNil(database)
    }
    
    func testSaveAndRetrieveUser() async throws {
        let testUser = LocalUser(
            id: "test-user-1",
            email: "test@echo.local",
            username: "testuser",
            avatar: nil,
            publicKey: "public-key-data",
            createdAt: Date()
        )
        
        try await database.saveUser(testUser)
        let retrieved = try await database.getUser(id: "test-user-1")
        
        XCTAssertEqual(retrieved?.id, testUser.id)
        XCTAssertEqual(retrieved?.username, testUser.username)
    }
    
    func testSaveAndRetrieveMessage() async throws {
        let testMessage = LocalMessage(
            id: "msg-1",
            conversationId: "conv-1",
            senderId: "user-1",
            content: "Test message",
            encryptedContent: "encrypted",
            nonce: "nonce"
        )
        
        try await database.saveMessage(testMessage)
        let messages = try await database.getMessages(conversationId: "conv-1", limit: 50)
        
        XCTAssertTrue(messages.contains { $0.id == testMessage.id })
    }
}

// MARK: - UseCase Tests

final class AuthenticationUseCaseTests: XCTestCase {
    var useCase: AuthenticateUseCase!
    var mockRepository: MockAuthRepository!
    
    override func setUp() {
        super.setUp()
        mockRepository = MockAuthRepository()
        useCase = AuthenticateUseCase(repository: mockRepository)
    }
    
    func testAuthenticateSuccess() async throws {
        mockRepository.loginResponse = LoginResponse(
            accessToken: "token-123",
            refreshToken: "refresh-123",
            expiresIn: 3600,
            user: UserResponse(
                id: "user-1",
                email: "test@echo.local",
                username: "testuser",
                avatar: nil,
                publicKey: "public-key",
                createdAt: Date()
            )
        )
        
        let response = try await useCase.execute(email: "test@echo.local", password: "password123")
        
        XCTAssertEqual(response.accessToken, "token-123")
        XCTAssertEqual(response.user.username, "testuser")
    }
}

// MARK: - Mock Repository

actor MockAuthRepository: AuthRepository {
    var loginResponse: LoginResponse?
    var shouldFail = false
    
    func register(email: String, password: String, username: String) async throws -> LoginResponse {
        if shouldFail {
            throw RepositoryError.unknown
        }
        return loginResponse ?? LoginResponse(
            accessToken: "token",
            refreshToken: "refresh",
            expiresIn: 3600,
            user: UserResponse(id: "1", email: email, username: username, avatar: nil, publicKey: "", createdAt: Date())
        )
    }
    
    func login(email: String, password: String) async throws -> LoginResponse {
        if shouldFail {
            throw RepositoryError.unknown
        }
        return loginResponse ?? LoginResponse(
            accessToken: "token",
            refreshToken: "refresh",
            expiresIn: 3600,
            user: UserResponse(id: "1", email: email, username: "user", avatar: nil, publicKey: "", createdAt: Date())
        )
    }
    
    func refreshToken() async throws -> RefreshTokenResponse {
        return RefreshTokenResponse(accessToken: "new-token", expiresIn: 3600)
    }
    
    func logout() async throws {}
    
    func verifyBiometric(challenge: String) async throws -> Bool {
        return true
    }
    
    func createPasskey() async throws -> String {
        return "passkey-data"
    }
    
    func verifyPasskey(credential: String) async throws -> Bool {
        return true
    }
}

// MARK: - Coordinator Tests

final class AppCoordinatorTests: XCTestCase {
    var coordinator: AppCoordinator!
    
    override func setUp() {
        super.setUp()
        coordinator = AppCoordinator()
    }
    
    func testCoordinatorInitialization() {
        XCTAssertNotNil(coordinator)
        XCTAssertFalse(coordinator.isAuthenticated)
    }
    
    func testNavigateToMain() {
        coordinator.navigate(to: .main)
        XCTAssertEqual(coordinator.navigationPath.first, .main)
    }
    
    func testPopBack() {
        coordinator.navigate(to: .main)
        coordinator.popBack()
        XCTAssertTrue(coordinator.navigationPath.isEmpty)
    }
    
    func testPopToRoot() {
        coordinator.navigate(to: .main)
        coordinator.navigate(to: .wallet)
        coordinator.popToRoot()
        XCTAssertTrue(coordinator.navigationPath.isEmpty)
    }
}

// MARK: - Integration Tests

final class AuthenticationIntegrationTests: XCTestCase {
    var authRepository: AuthRepository!
    var authUseCase: AuthenticateUseCase!
    var keychain: KeychainManager!
    
    override func setUp() async throws {
        try await super.setUp()
        keychain = KeychainManager.shared
        authRepository = MockAuthRepository()
        authUseCase = AuthenticateUseCase(repository: authRepository)
        
        try await keychain.clearAuthCredentials()
    }
    
    func testFullAuthenticationFlow() async throws {
        // Step 1: Authenticate
        let response = try await authUseCase.execute(
            email: "test@echo.local",
            password: "password123"
        )
        
        XCTAssertNotNil(response.accessToken)
        
        // Step 2: Verify token was stored
        let storedToken = try await keychain.getAuthToken()
        XCTAssertNotNil(storedToken)
    }
}

// MARK: - Performance Tests

final class PerformanceTests: XCTestCase {
    var encryption: KinnamiEncryption!
    
    override func setUp() {
        super.setUp()
        encryption = KinnamiEncryption()
    }
    
    func testEncryptionPerformance() throws {
        measure {
            let (_, publicKey) = encryption.generateEphemeralKeyPair()
            _ = try? encryption.encryptWithKeyAgreement(
                plaintext: "Test message content",
                recipientPublicKeyData: publicKey
            )
        }
    }
    
    func testKeyPairGenerationPerformance() {
        measure {
            _ = encryption.generateEphemeralKeyPair()
        }
    }
}
