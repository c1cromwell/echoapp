#if os(iOS)
// App/AppState.swift
// v2.5.3 — routes .firstRun / .login / .authenticated

import SwiftUI
import Observation
import UIKit

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
    var activePersona: PersonaSummary = PersonaSummary(id: "default", name: "You", initials: "")

    let provisionService: SilentProvisionService

    init(provisionService: SilentProvisionService) {
        self.provisionService = provisionService
        self.displayName = UserDefaults.standard.string(forKey: "echo.displayName") ?? ""
        self.trustTier = UserDefaults.standard.integer(forKey: "echo.trustTier")
        self.root = Self.initialRoot()

        self.personas = PersonaSessionStore.personas(displayName: displayName)
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
        // Deterministic entry point for XCUITest runs (see EchoUITests).
        if UITestSupport.isActive {
            return UITestSupport.startAuthenticated ? .authenticated : .firstRun
        }
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
        // Name the primary persona after the chosen username (not "Me").
        applyPrimaryName(displayName)
        root = .authenticated
    }

    func firstRunRestoreCompleted(_ identity: RestoredIdentity) {
        self.displayName = identity.displayName
        UserDefaults.standard.set(identity.displayName, forKey: "echo.displayName")
        applyPrimaryName(identity.displayName)
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

    // MARK: - Persona management (Persona Settings)

    /// Personas the user can manage (excludes the biometric-gated hidden vault persona).
    var manageablePersonas: [PersonaSummary] {
        personas.filter { $0.id != PersonaSessionStore.hiddenPersonaId }
    }

    /// Renames the primary persona and the profile display name. Used by
    /// "edit username" (local display name only — the account handle is unchanged).
    func renamePrimaryPersona(_ name: String) {
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty else { return }
        displayName = clean
        UserDefaults.standard.set(clean, forKey: "echo.displayName")
        applyPrimaryName(clean)
    }

    func addPersona(name: String, colorHex: UInt32, trustLevel: String = "Trusted") {
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty else { return }
        let persona = PersonaSummary(
            id: PersonaSessionStore.newPersonaId(),
            name: clean,
            initials: PersonaSessionStore.initials(from: clean),
            trustLevel: trustLevel,
            colorHex: colorHex
        )
        personas = PersonaSessionStore.add(persona, to: personas)
    }

    func updatePersona(id: String, name: String, colorHex: UInt32, trustLevel: String) {
        guard let existing = personas.first(where: { $0.id == id }) else { return }
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty else { return }
        // The primary persona keeps its display name in sync with the profile.
        if id == PersonaSessionStore.primaryPersonaId {
            renamePrimaryPersona(clean)
            return
        }
        let updated = PersonaSummary(
            id: existing.id,
            name: clean,
            initials: PersonaSessionStore.initials(from: clean),
            trustLevel: trustLevel,
            isHidden: existing.isHidden,
            colorHex: colorHex
        )
        personas = PersonaSessionStore.update(updated, in: personas)
        if activePersona.id == id { activePersona = updated }
    }

    func deletePersona(id: String) {
        guard !PersonaSessionStore.isSystemPersona(id) else { return }
        personas = PersonaSessionStore.delete(id: id, from: personas)
        // If the active persona was removed, fall back to primary.
        if activePersona.id == id {
            let primary = personas.first { $0.id == PersonaSessionStore.primaryPersonaId } ?? personas.first!
            switchPersona(primary)
        }
    }

    /// Renames the primary persona in the persisted list and refreshes live state.
    private func applyPrimaryName(_ name: String) {
        personas = PersonaSessionStore.renamePrimary(name: name, in: PersonaSessionStore.personas(displayName: name))
        activePersona = PersonaSessionStore.resolveActive(in: personas)
    }
}

enum AppTab: String { case messages, wallet, me }

// MARK: - UI test support

/// Deterministic launch behaviour for automated QA/QE (XCUITest). Driven only by
/// launch arguments that the `EchoUITests` bundle passes, so it stays inert in
/// normal use. See `scripts/qa-record.sh` and `EchoUXFlowUITests`.
enum UITestSupport {
    /// True when launched by the UI-test runner (`-uiTestMode`).
    static var isActive: Bool {
        ProcessInfo.processInfo.arguments.contains("-uiTestMode")
    }

    /// `-uiTestAuthenticated` lands directly on the main tabs; otherwise a
    /// UI-test launch starts at first-run onboarding.
    static var startAuthenticated: Bool {
        ProcessInfo.processInfo.arguments.contains("-uiTestAuthenticated")
    }

    /// Call as early as possible during app launch. Disables animations so
    /// XCUITest queries settle immediately instead of racing transitions.
    @MainActor static func applyIfNeeded() {
        guard isActive else { return }
        UIView.setAnimationsEnabled(false)
    }
}

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
