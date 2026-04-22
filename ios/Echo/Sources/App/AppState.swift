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

    let provisionService: SilentProvisionService

    init(provisionService: SilentProvisionService) {
        self.provisionService = provisionService
        self.displayName = UserDefaults.standard.string(forKey: "echo.displayName") ?? ""
        self.root = Self.initialRoot()
    }

    private static func initialRoot() -> Root {
        let hasCompleted = UserDefaults.standard.bool(forKey: "echo.hasCompletedFirstRun")
        guard hasCompleted else { return .firstRun }
        return .login
    }

    // MARK: - First-run intents

    func firstRunCompleted(displayName: String) {
        self.displayName = displayName
        provisionService.begin(displayName: displayName)
        UserDefaults.standard.set(true, forKey: "echo.hasCompletedFirstRun")
        UserDefaults.standard.set(displayName, forKey: "echo.displayName")
        root = .authenticated
    }

    func firstRunRestoreTapped() {
        // Placeholder — BIP-39 recovery flow (REQ-RECOVERY, pending spec)
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
                        onComplete: { name in
                            appState.firstRunCompleted(displayName: name)
                        },
                        onRestoreTapped: {
                            appState.firstRunRestoreTapped()
                        }
                    )
                )
            case .login:
                NavigationStack {
                    GlacialLoginScreen(
                        onPasskeyLogin: { appState.loginSucceeded() },
                        onSMSLogin: { _ in appState.loginSucceeded() },
                        onGetStarted: { appState.loginSucceeded() }
                    )
                }
            case .authenticated:
                MainTabView()
                    .environment(appState)
            }
        }
    }
}
