import Foundation

import Foundation

/// Dependency Injection Container for ECHO iOS app
/// Manages creation and lifecycle of all services and dependencies
@MainActor
final class DIContainer {
    
    // MARK: - Singleton
    static let shared = DIContainer()
    
    // MARK: - Service Instances
    private var services: [String: Any] = [:]
    private var factories: [String: () -> Any] = [:]
    
    // MARK: - Initialization
    private init() {
        registerFactories()
    }
    
    // MARK: - Registration
    
    /// Register a factory for a service type
    func registerFactory<T>(_ key: String, factory: @escaping () -> T) {
        factories[key] = factory
    }
    
    /// Register a singleton service
    func registerSingleton<T>(_ key: String, service: T) {
        services[key] = service
    }
    
    // MARK: - Resolution
    
    /// Resolve a service by key
    func resolve<T>(_ key: String) -> T? {
        if let service = services[key] as? T {
            return service
        }
        
        if let factory = factories[key] as? (() -> T) {
            let service = factory()
            services[key] = service
            return service
        }
        
        return nil
    }
    
    /// Resolve or create a service using factory
    func resolve<T>(_ key: String, factory: @escaping () -> T) -> T {
        if let service = services[key] as? T {
            return service
        }
        
        let service = factory()
        services[key] = service
        return service
    }
    
    // MARK: - Factory Registration
    
    private func registerFactories() {
        // Security Services
        registerFactory(ServiceKeys.secureEnclave) {
            SecureEnclaveManager.shared
        }
        
        registerFactory(ServiceKeys.biometricAuth) {
            BiometricAuthManager()
        }
        
        registerFactory(ServiceKeys.keychain) {
            KeychainManager.shared
        }
        
        registerFactory(ServiceKeys.kinnamiEncryption) {
            KinnamiEncryption()
        }
        
        // Networking Services
        registerFactory(ServiceKeys.apiClient) {
            APIClient(configuration: APIConfiguration.default)
        }
        
        registerFactory(ServiceKeys.webSocketClient) {
            WebSocketClient(configuration: WebSocketConfiguration.default)
        }
        
        // Storage Services
        registerFactory(ServiceKeys.localStorage) {
            LocalDatabase.shared
        }
        
        // Repository Services
        registerFactory(ServiceKeys.authRepository) { [weak self] in
            ConcreteAuthRepository(
                apiClient: self?.resolve(ServiceKeys.apiClient) ?? APIClient(configuration: .default),
                keychain: self?.resolve(ServiceKeys.keychain) ?? KeychainManager.shared,
                secureEnclave: self?.resolve(ServiceKeys.secureEnclave) ?? SecureEnclaveManager.shared
            )
        }
        
        registerFactory(ServiceKeys.userRepository) { [weak self] in
            ConcreteUserRepository(
                apiClient: self?.resolve(ServiceKeys.apiClient) ?? APIClient(configuration: .default),
                localStorage: self?.resolve(ServiceKeys.localStorage) ?? LocalDatabase.shared
            )
        }
        
        registerFactory(ServiceKeys.messageRepository) { [weak self] in
            ConcreteMessageRepository(
                apiClient: self?.resolve(ServiceKeys.apiClient) ?? APIClient(configuration: .default),
                webSocketClient: self?.resolve(ServiceKeys.webSocketClient) ?? WebSocketClient(configuration: .default),
                encryption: self?.resolve(ServiceKeys.kinnamiEncryption) ?? KinnamiEncryption(),
                localStorage: self?.resolve(ServiceKeys.localStorage) ?? LocalDatabase.shared
            )
        }
        
        registerFactory(ServiceKeys.tokenRepository) { [weak self] in
            ConcreteTokenRepository(
                apiClient: self?.resolve(ServiceKeys.apiClient) ?? APIClient(configuration: .default),
                localStorage: self?.resolve(ServiceKeys.localStorage) ?? LocalDatabase.shared
            )
        }
        
        // UseCase Services
        registerFactory(ServiceKeys.authenticateUseCase) { [weak self] in
            AuthenticateUseCase(
                repository: self?.resolve(ServiceKeys.authRepository) ?? ConcreteAuthRepository()
            )
        }
        
        registerFactory(ServiceKeys.sendMessageUseCase) { [weak self] in
            SendMessageUseCase(
                repository: self?.resolve(ServiceKeys.messageRepository) ?? ConcreteMessageRepository()
            )
        }
        
        registerFactory(ServiceKeys.createDIDUseCase) { [weak self] in
            CreateDIDUseCase(
                userRepository: self?.resolve(ServiceKeys.userRepository) ?? ConcreteUserRepository(),
                secureEnclave: self?.resolve(ServiceKeys.secureEnclave) ?? SecureEnclaveManager.shared
            )
        }
        
        registerFactory(ServiceKeys.getBalanceUseCase) { [weak self] in
            GetBalanceUseCase(
                repository: self?.resolve(ServiceKeys.tokenRepository) ?? ConcreteTokenRepository()
            )
        }
    }
}

// MARK: - Service Keys

enum ServiceKeys {
    // Security
    static let secureEnclave = "security.secureEnclave"
    static let biometricAuth = "security.biometricAuth"
    static let keychain = "security.keychain"
    static let kinnamiEncryption = "security.kinnamiEncryption"
    
    // Networking
    static let apiClient = "networking.apiClient"
    static let webSocketClient = "networking.webSocketClient"
    
    // Storage
    static let localStorage = "storage.localStorage"
    static let secureStorage = "storage.secureStorage"
    static let cacheManager = "storage.cacheManager"
    
    // Repositories
    static let authRepository = "repository.auth"
    static let userRepository = "repository.user"
    static let messageRepository = "repository.message"
    static let tokenRepository = "repository.token"
    
    // UseCases
    static let authenticateUseCase = "usecase.authenticate"
    static let registerUseCase = "usecase.register"
    static let passkeyUseCase = "usecase.passkey"
    static let sendMessageUseCase = "usecase.sendMessage"
    static let fetchMessagesUseCase = "usecase.fetchMessages"
    static let createDIDUseCase = "usecase.createDID"
    static let verifyIdentityUseCase = "usecase.verifyIdentity"
    static let getBalanceUseCase = "usecase.getBalance"
    static let sendTokensUseCase = "usecase.sendTokens"
    static let stakeTokensUseCase = "usecase.stakeTokens"
}

// MARK: - Convenience Resolvers

extension DIContainer {
    
    func resolveSecureEnclave() -> SecureEnclaveManager? {
        resolve(ServiceKeys.secureEnclave)
    }
    
    func resolveBiometricAuth() -> BiometricAuthManager? {
        resolve(ServiceKeys.biometricAuth)
    }
    
    func resolveKeychain() -> KeychainManager? {
        resolve(ServiceKeys.keychain)
    }
    
    func resolveAPIClient() -> APIClient? {
        resolve(ServiceKeys.apiClient)
    }
    
    func resolveAuthRepository() -> AuthRepository? {
        resolve(ServiceKeys.authRepository)
    }
    
    func resolveUserRepository() -> UserRepository? {
        resolve(ServiceKeys.userRepository)
    }
    
    func resolveMessageRepository() -> MessageRepository? {
        resolve(ServiceKeys.messageRepository)
    }
    
    func resolveTokenRepository() -> TokenRepository? {
        resolve(ServiceKeys.tokenRepository)
    }
}
