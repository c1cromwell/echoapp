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

    var personas: [PersonaSummary] = []
    var activePersona: PersonaSummary = PersonaSummary(id: "default", name: "Me", initials: "ME")

    let provisionService: SilentProvisionService

    init(provisionService: SilentProvisionService) {
        self.provisionService = provisionService
        self.displayName = UserDefaults.standard.string(forKey: "echo.displayName") ?? ""
        self.trustTier = UserDefaults.standard.integer(forKey: "echo.trustTier")
        self.root = Self.initialRoot()

        self.personas = PersonaSessionStore.defaultPersonas(displayName: displayName)
        self.activePersona = PersonaSessionStore.resolveActive(in: personas)
    }

    private static func initials(from name: String) -> String {
        let parts = name.split(separator: " ")
        if parts.count >= 2 {
            return String(parts[0].prefix(1) + parts[1].prefix(1)).uppercased()
        }
        return String(name.prefix(2)).uppercased()
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
        // Mirror the username into the Keychain so the returning-user login
        // (GlacialLoginScreen) recognises the account and shows Face ID — not
        // the "Create account" screen.
        Task { try? await KeychainManager.shared.store(key: "echo.username.current", value: displayName) }
        root = .authenticated
    }

    func firstRunRestoreCompleted(_ identity: RestoredIdentity) {
        self.displayName = identity.displayName
        root = .authenticated
    }

    // MARK: - Auth intents

    func loginSucceeded() {
        if let code = EchoDeepLink.consumePendingInvite() {
            pendingInviteCode = code
        }
        root = .authenticated
    }

    /// Sign out from Settings: clear tokens/WS and show login (Face ID not auto-fired once).
    func signOut() async {
        await SessionSignOut.perform()
        selectedTab = .messages
        root = .login
    }

    /// Full reset for onboarding QA — also clears `echo.hasCompletedFirstRun`.
    func signOutForNewAccountSetup() async {
        await SessionSignOut.performIncludingOnboardingReset()
        selectedTab = .messages
        root = .firstRun
    }

    func loggedOut() {
        Task { await signOut() }
    }

    func switchPersona(_ persona: PersonaSummary) {
        activePersona = persona
        PersonaSessionStore.setActivePersonaId(persona.id)
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
                .onOpenURL { url in
                    if case .invite(let code) = EchoDeepLink.parse(url) {
                        EchoDeepLink.stashPendingInvite(code)
                    }
                }
            case .authenticated:
                MainTabView()
                    .environment(appState)
                    .onOpenURL { url in
                        if url.scheme == "echo-enroll" {
                            return
                        }
                        switch EchoDeepLink.parse(url) {
                        case .invite(let code):
                            appState.pendingInviteCode = code
                        case .linkDevice:
                            DeepLinkHandler.shared.handle(url)
                        case .profile, .none:
                            DeepLinkHandler.shared.handle(url)
                        }
                    }
            }
        }
        .echoTranslationHost()
    }
}

#endif
