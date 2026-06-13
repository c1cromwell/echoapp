#if os(iOS)
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
        
        // Networking Services (WO-2: cert pinning; WO-1: passkey signing)
        registerFactory(ServiceKeys.certificatePinner) {
            CertificatePinner()
        }

        registerFactory(ServiceKeys.apiClient) { [weak self] () -> APIClient in
            let pinner: CertificatePinner? = self?.resolve(ServiceKeys.certificatePinner)
            let client = APIClient(configuration: APIConfiguration.default, pinner: pinner)
            // Add passkey signing interceptor so all authenticated requests carry
            // X-Sender-DID + X-Signature headers (WO-1 iOS side).
            Task { await client.addInterceptor(PasskeySigningInterceptor()) }
            return client
        }
        
        registerFactory(ServiceKeys.webSocketClient) {
            WebSocketClient(configuration: WebSocketConfiguration.default)
        }

        // Phase 3: typing / read receipts / reactions (WO-192)
        registerFactory(ServiceKeys.conversationSignalService) {
            let apiBase = WebSocketURLBuilder.apiBaseURLFromEnvironment()
                ?? APIConfiguration.default.baseURL
            return ConversationSignalService(apiBaseURL: apiBase)
        }

        registerFactory(ServiceKeys.reactionsAPI) { [weak self] () -> ReactionsAPI in
            let client: APIClient = self?.resolve(ServiceKeys.apiClient)
                ?? APIClient(configuration: .default)
            return ReactionsAPI(apiClient: client)
        }

        registerFactory(ServiceKeys.receiptsAPI) { [weak self] () -> MessageReceiptsAPI in
            let client: APIClient = self?.resolve(ServiceKeys.apiClient)
                ?? APIClient(configuration: .default)
            return MessageReceiptsAPI(apiClient: client)
        }

        registerFactory(ServiceKeys.messageOpsAPI) { [weak self] () -> MessageOpsAPI in
            let client: APIClient = self?.resolve(ServiceKeys.apiClient)
                ?? APIClient(configuration: .default)
            return MessageOpsAPI(apiClient: client)
        }

        // WO-221: private contact discovery (OPRF + PSI)
        registerFactory(ServiceKeys.contactDiscoveryService) { [weak self] () -> ContactDiscoveryService in
            let client: APIClient = self?.resolve(ServiceKeys.apiClient)
                ?? APIClient(configuration: .default)
            let api = ContactDiscoveryAPIClient(apiClient: client)
            return ContactDiscoveryService(oprf: OPRFClientFactory.makeDefault(), api: api)
        }

        registerFactory(ServiceKeys.contactSocialAPI) { [weak self] () -> ContactSocialAPIClient in
            let client: APIClient = self?.resolve(ServiceKeys.apiClient)
                ?? APIClient(configuration: .default)
            return ContactSocialAPIClient(apiClient: client)
        }

        // WO-39 / WO-221 / WO-222 contact use cases
        registerFactory(ServiceKeys.contactDiscoveryUseCase) { [weak self] () -> ContactDiscoveryUseCase in
            let service: ContactDiscoveryService = self?.resolve(ServiceKeys.contactDiscoveryService)
                ?? ContactDiscoveryService(
                    oprf: OPRFClientFactory.makeDefault(),
                    api: ContactDiscoveryAPIClient(apiClient: APIClient(configuration: .default))
                )
            return ContactDiscoveryUseCase(service: service)
        }

        registerFactory(ServiceKeys.inviteLinkUseCase) { [weak self] () -> InviteLinkUseCase in
            let client: ContactSocialAPIClient = self?.resolve(ServiceKeys.contactSocialAPI)
                ?? ContactSocialAPIClient(apiClient: APIClient(configuration: .default))
            return InviteLinkUseCase(client: client)
        }

        registerFactory(ServiceKeys.usernameSearchUseCase) { [weak self] () -> UsernameSearchUseCase in
            let client: ContactSocialAPIClient = self?.resolve(ServiceKeys.contactSocialAPI)
                ?? ContactSocialAPIClient(apiClient: APIClient(configuration: .default))
            return UsernameSearchUseCase(client: client)
        }

        registerFactory(ServiceKeys.qrContactExchangeUseCase) {
            QRContactExchangeUseCase()
        }

        // Phase B: per-conversation preferences (mute / disappearing timer)
        registerFactory(ServiceKeys.conversationPreferences) {
            ConversationPreferencesStore()
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
    static let certificatePinner = "networking.certificatePinner"
    static let webSocketClient = "networking.webSocketClient"
    static let conversationSignalService = "networking.conversationSignalService"
    static let reactionsAPI = "networking.reactionsAPI"
    static let receiptsAPI = "networking.receiptsAPI"
    static let messageOpsAPI = "networking.messageOpsAPI"
    static let contactDiscoveryService = "services.contactDiscovery"
    static let contactSocialAPI = "services.contactSocialAPI"
    static let contactDiscoveryUseCase = "usecase.contactDiscovery"
    static let inviteLinkUseCase = "usecase.inviteLink"
    static let usernameSearchUseCase = "usecase.usernameSearch"
    static let qrContactExchangeUseCase = "usecase.qrContactExchange"
    static let conversationPreferences = "services.conversationPreferences"
    
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

    func resolveConversationSignalService() -> ConversationSignalService? {
        resolve(ServiceKeys.conversationSignalService)
    }

    func resolveReactionsAPI() -> ReactionsAPI? {
        resolve(ServiceKeys.reactionsAPI)
    }

    func resolveReceiptsAPI() -> MessageReceiptsAPI? {
        resolve(ServiceKeys.receiptsAPI)
    }

    func resolveMessageOpsAPI() -> MessageOpsAPI? {
        resolve(ServiceKeys.messageOpsAPI)
    }

    func resolveContactDiscoveryService() -> ContactDiscoveryService? {
        resolve(ServiceKeys.contactDiscoveryService)
    }

    func resolveConversationPreferences() -> ConversationPreferencesStore? {
        resolve(ServiceKeys.conversationPreferences)
    }

    func resolveContactSocialAPI() -> ContactSocialAPIClient? {
        resolve(ServiceKeys.contactSocialAPI)
    }

    func resolveContactDiscoveryUseCase() -> ContactDiscoveryUseCase? {
        resolve(ServiceKeys.contactDiscoveryUseCase)
    }

    func resolveInviteLinkUseCase() -> InviteLinkUseCase? {
        resolve(ServiceKeys.inviteLinkUseCase)
    }

    func resolveUsernameSearchUseCase() -> UsernameSearchUseCase? {
        resolve(ServiceKeys.usernameSearchUseCase)
    }

    func resolveQRContactExchangeUseCase() -> QRContactExchangeUseCase? {
        resolve(ServiceKeys.qrContactExchangeUseCase)
    }

    /// Factory for Phase 3 chat detail (WO-192) — uses registered `ConversationSignalService`.
    func makeChatDetailViewModel() -> ChatDetailViewModel {
        let service: ConversationSignalService = resolve(ServiceKeys.conversationSignalService)
            ?? ConversationSignalService(
                apiBaseURL: WebSocketURLBuilder.apiBaseURLFromEnvironment()
                    ?? APIConfiguration.default.baseURL
            )
        return ChatDetailViewModel(signalService: service)
    }
}
#endif
