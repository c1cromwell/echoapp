import Foundation
import SwiftUI

// MARK: - Navigation Model

@MainActor
enum NavigationPath {
    // Existing
    case authentication
    case onboarding
    case main
    case settings
    case chat(conversationId: String)
    case profile(userId: String)
    case contacts
    case wallet

    // New routes (per iOS Spec v4.2.1)
    case stakingDetail
    case delegationDetail
    case governance
    case proposalDetail(id: String)
    case contactDetail(id: String)
    case voiceCall(contactId: String)
    case videoCall(contactId: String)
    case search
    case searchScoped(conversationId: String)
    case mediaGallery(conversationId: String)
    case notifications
    case qrIdentity
    case backup
    case enterpriseProfile(id: String)
    case botManagement
    case botDetail(id: String)
}

// MARK: - Coordinator Protocol

@MainActor
protocol Coordinator: AnyObject {
    var navigationPath: [NavigationPath] { get set }
    func start()
    func navigate(to path: NavigationPath)
    func popBack()
    func popToRoot()
}

// MARK: - App Coordinator

@MainActor
class AppCoordinator: Coordinator {
    @Published var navigationPath: [NavigationPath] = []
    @Published var isAuthenticated = false
    
    private let container: DIContainer
    private let authRepository: AuthRepository
    
    init(container: DIContainer = .shared) {
        self.container = container
        self.authRepository = container.resolveAuthRepository() ?? ConcreteAuthRepository()
    }
    
    func start() {
        // Check if user is authenticated
        Task {
            do {
                let keychain = KeychainManager.shared
                let token = try await keychain.getAuthToken()
                
                if token != nil {
                    isAuthenticated = true
                    navigate(to: .main)
                } else {
                    navigate(to: .authentication)
                }
            } catch {
                navigate(to: .authentication)
            }
        }
    }
    
    func navigate(to path: NavigationPath) {
        navigationPath.removeAll()
        navigationPath.append(path)
    }
    
    func popBack() {
        if !navigationPath.isEmpty {
            navigationPath.removeLast()
        }
    }
    
    func popToRoot() {
        navigationPath.removeAll()
    }
    
    func logout() {
        Task {
            try await authRepository.logout()
            isAuthenticated = false
            navigate(to: .authentication)
        }
    }
}

// MARK: - Auth Coordinator

@MainActor
class AuthCoordinator: Coordinator {
    @Published var navigationPath: [NavigationPath] = []
    
    private let container: DIContainer
    private let authRepository: AuthRepository
    var onAuthenticationComplete: (() -> Void)?
    
    init(container: DIContainer = .shared) {
        self.container = container
        self.authRepository = container.resolveAuthRepository() ?? ConcreteAuthRepository()
    }
    
    func start() {
        navigationPath = [.authentication]
    }
    
    func navigate(to path: NavigationPath) {
        navigationPath.append(path)
    }
    
    func popBack() {
        if !navigationPath.isEmpty {
            navigationPath.removeLast()
        }
    }
    
    func popToRoot() {
        navigationPath = [.authentication]
    }
    
    func completeAuthentication() {
        onAuthenticationComplete?()
    }
    
    func showOnboarding() {
        navigationPath = [.onboarding]
    }
}

// MARK: - Main Coordinator

@MainActor
class MainCoordinator: Coordinator {
    @Published var navigationPath: [NavigationPath] = []
    @Published var selectedTab: MainTab = .conversations
    
    private let container: DIContainer
    
    enum MainTab: String, CaseIterable {
        case conversations, contacts, profile, wallet
    }
    
    init(container: DIContainer = .shared) {
        self.container = container
    }
    
    func start() {
        navigationPath = [.main]
    }
    
    func navigate(to path: NavigationPath) {
        navigationPath.append(path)
    }
    
    func popBack() {
        if !navigationPath.isEmpty {
            navigationPath.removeLast()
        }
    }
    
    func popToRoot() {
        navigationPath = [.main]
    }
    
    func selectTab(_ tab: MainTab) {
        selectedTab = tab
        popToRoot()
    }
    
    func openChat(conversationId: String) {
        navigate(to: .chat(conversationId: conversationId))
    }
    
    func openProfile(userId: String) {
        navigate(to: .profile(userId: userId))
    }
    
    func openWallet() {
        navigate(to: .wallet)
    }
    
    func openSettings() {
        navigate(to: .settings)
    }
}

// MARK: - Settings Coordinator

@MainActor
class SettingsCoordinator: Coordinator {
    @Published var navigationPath: [NavigationPath] = []
    
    private let container: DIContainer
    var onLogout: (() -> Void)?
    
    init(container: DIContainer = .shared) {
        self.container = container
    }
    
    func start() {
        navigationPath = [.settings]
    }
    
    func navigate(to path: NavigationPath) {
        navigationPath.append(path)
    }
    
    func popBack() {
        if !navigationPath.isEmpty {
            navigationPath.removeLast()
        }
    }
    
    func popToRoot() {
        navigationPath = [.settings]
    }
    
    func logout() {
        Task {
            let authRepository = container.resolveAuthRepository() ?? ConcreteAuthRepository()
            try await authRepository.logout()
            onLogout?()
        }
    }
}

// MARK: - Chat Coordinator

@MainActor
class ChatCoordinator: Coordinator {
    @Published var navigationPath: [NavigationPath] = []
    
    let conversationId: String
    private let container: DIContainer
    
    init(conversationId: String, container: DIContainer = .shared) {
        self.conversationId = conversationId
        self.container = container
    }
    
    func start() {
        navigationPath = [.chat(conversationId: conversationId)]
    }
    
    func navigate(to path: NavigationPath) {
        navigationPath.append(path)
    }
    
    func popBack() {
        if !navigationPath.isEmpty {
            navigationPath.removeLast()
        }
    }
    
    func popToRoot() {
        navigationPath = [.chat(conversationId: conversationId)]
    }
}

// MARK: - Navigation Helper

@MainActor
class NavigationHelper {
    static let shared = NavigationHelper()
    
    @Published var currentPath: NavigationPath?
    @Published var isNavigating = false
    
    func navigate(to path: NavigationPath, completion: @escaping () -> Void = {}) {
        withAnimation(.easeInOut(duration: 0.3)) {
            isNavigating = true
            currentPath = path
        }
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
            self.isNavigating = false
            completion()
        }
    }
    
    func dismiss() {
        withAnimation(.easeInOut(duration: 0.3)) {
            currentPath = nil
        }
    }
}

// MARK: - Route Manager

@MainActor
class RouteManager: ObservableObject {
    @Published var navigationStack: [NavigationPath] = []
    
    func push(_ route: NavigationPath) {
        navigationStack.append(route)
    }
    
    func pop() {
        if !navigationStack.isEmpty {
            navigationStack.removeLast()
        }
    }
    
    func popToRoot() {
        navigationStack.removeAll()
    }
    
    func replace(_ route: NavigationPath) {
        navigationStack.removeAll()
        navigationStack.append(route)
    }
}

// MARK: - Deep Link Handler

@MainActor
class DeepLinkHandler {
    static let shared = DeepLinkHandler()
    
    private let appCoordinator: AppCoordinator
    
    init(appCoordinator: AppCoordinator = AppCoordinator()) {
        self.appCoordinator = appCoordinator
    }
    
    func handle(_ url: URL) {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true) else {
            return
        }
        
        // Example: echo://chat/123abc
        if let host = components.host {
            switch host {
            case "chat":
                if let conversationId = components.path.dropFirst().split(separator: "/").first {
                    appCoordinator.navigate(to: .chat(conversationId: String(conversationId)))
                }
            case "profile":
                if let userId = components.path.dropFirst().split(separator: "/").first {
                    appCoordinator.navigate(to: .profile(userId: String(userId)))
                }
            case "wallet":
                appCoordinator.navigate(to: .wallet)
            case "governance":
                appCoordinator.navigate(to: .governance)
            case "contact":
                if let contactId = components.path.dropFirst().split(separator: "/").first {
                    appCoordinator.navigate(to: .contactDetail(id: String(contactId)))
                }
            case "notifications":
                appCoordinator.navigate(to: .notifications)
            case "backup":
                appCoordinator.navigate(to: .backup)
            case "qr":
                appCoordinator.navigate(to: .qrIdentity)
            case "search":
                appCoordinator.navigate(to: .search)
            default:
                break
            }
        }
    }
}

// MARK: - Navigation Stack

@MainActor
class NavigationStack: ObservableObject {
    @Published var stack: [NavigationPath] = []
    
    func push(_ path: NavigationPath) {
        stack.append(path)
    }
    
    func pop() -> NavigationPath? {
        return stack.popLast()
    }
    
    func popAll() {
        stack.removeAll()
    }
    
    func replace(_ path: NavigationPath) {
        stack.removeLast()
        stack.append(path)
    }
    
    var current: NavigationPath? {
        stack.last
    }
    
    var canPop: Bool {
        !stack.isEmpty
    }
}
