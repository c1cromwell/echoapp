#if os(iOS)
// App/AppState.swift
// v2.5.3 — routes .firstRun / .login / .authenticated

import SwiftUI
import Observation

@MainActor
@Observable
final class AppState {

    enum Root: Equatable {
        case firstRun
        case login
        case authenticated
    }

    var root: Root
    var selectedTab: AppTab = .messages
    var displayName: String = ""
    var trustTier: Int = 0
    /// Set when user opens echo://invite?code=… (Wave 0.4).
    var pendingInviteCode: String?

    let provisionService: SilentProvisionService

    init(provisionService: SilentProvisionService) {
        self.provisionService = provisionService
        self.displayName = UserDefaults.standard.string(forKey: "echo.displayName") ?? ""
        self.trustTier = UserDefaults.standard.integer(forKey: "echo.trustTier")
        self.root = Self.initialRoot()
    }

    private static func initialRoot() -> Root {
        let hasCompleted = UserDefaults.standard.bool(forKey: "echo.hasCompletedFirstRun")
        guard hasCompleted else { return .firstRun }
        return .login
    }

    // MARK: - First-run intents

    func firstRunCompleted(displayName: String, trustTier: Int = 0) {
        self.displayName = displayName
        self.trustTier = trustTier
        provisionService.begin(displayName: displayName)
        UserDefaults.standard.set(true, forKey: "echo.hasCompletedFirstRun")
        UserDefaults.standard.set(Date(), forKey: "echo.firstRunCompletedAt")
        UserDefaults.standard.set(displayName, forKey: "echo.displayName")
        UserDefaults.standard.set(trustTier, forKey: "echo.trustTier")
        root = .authenticated
    }

    func firstRunRestoreCompleted(_ identity: RestoredIdentity) {
        self.displayName = identity.displayName
        root = .authenticated
    }

    // MARK: - Auth intents

    func loginSucceeded() {
        root = .authenticated
    }

    func loggedOut() {
        root = .login
    }
}

enum AppTab: String { case messages, wallet, me }

// MARK: - Root View

struct EchoRootView: View {
    @State var appState: AppState

    var body: some View {
        Group {
            switch appState.root {
            case .firstRun:
                FirstRunCoordinatorView(
                    coordinator: FirstRunCoordinator(
                        onComplete: { name, tier in
                            appState.firstRunCompleted(displayName: name, trustTier: tier)
                        },
                        onRestoreComplete: { identity in
                            appState.firstRunRestoreCompleted(identity)
                        }
                    )
                )
            case .login:
                NavigationStack {
                    GlacialLoginScreen(
                        onPasskeyLogin: { appState.loginSucceeded() },
                        onSMSLogin: { _ in appState.loginSucceeded() },
                        onGetStarted: { appState.root = .firstRun }
                    )
                }
            case .authenticated:
                MainTabView()
                    .environment(appState)
                    .onOpenURL { url in
                        if url.scheme == "echo-enroll" {
                            return
                        }
                        if url.host == "invite",
                           let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
                           let code = components.queryItems?.first(where: { $0.name == "code" })?.value {
                            appState.pendingInviteCode = code
                        } else {
                            DeepLinkHandler.shared.handle(url)
                        }
                    }
            }
        }
    }
}

#endif
