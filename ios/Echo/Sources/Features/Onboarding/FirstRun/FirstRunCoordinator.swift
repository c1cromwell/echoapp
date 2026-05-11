#if os(iOS)
// Features/Onboarding/FirstRun/FirstRunCoordinator.swift
// Design review: compress 4 steps → 2.
//   1. Welcome (single page, no carousel)
//   2. Name + Face ID (fused — typing the name and enrolling the key happen together)
//   3. Recovery setup (phrase + optional SMS — skippable, labelled "1 / 2" above)
//
// The old 3-slide carousel and separate BiometricEnrollmentView are superseded
// by EchoWelcomeView and NameAndKeyView. The routes are kept for backward compat.

import SwiftUI
import Observation

@MainActor
@Observable
final class FirstRunCoordinator {
    enum Route: Hashable {
        case welcome          // EchoWelcomeView — single page, no carousel
        case nameAndKey       // NameAndKeyView  — fused name + Face ID (2-in-1)
        case recoverySetup(did: String)
        case restore
        // Legacy routes (kept for demo app back-compat)
        case displayName
        case biometricEnrollment
    }

    var path: [Route] = []
    var displayName: String = ""

    let onComplete: (String) -> Void
    let onRestoreComplete: (RestoredIdentity) -> Void

    init(
        onComplete: @escaping (String) -> Void,
        onRestoreComplete: @escaping (RestoredIdentity) -> Void
    ) {
        self.onComplete = onComplete
        self.onRestoreComplete = onRestoreComplete
    }

    func welcomeContinueTapped() {
        path.append(.nameAndKey)
    }

    // Called by NameAndKeyView on completion (fused name + Face ID).
    func nameAndKeyCompleted(username: String, did: String) {
        displayName = username
        path.append(.recoverySetup(did: did))
    }

    // Legacy — kept for demo app and existing call sites
    func displayNameSubmitted(_ name: String) {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard DisplayNameValidator.isValid(trimmed) else { return }
        displayName = trimmed
        path.append(.biometricEnrollment)
    }

    func biometricEnrollmentCompleted(did: String) {
        path.append(.recoverySetup(did: did))
    }

    func biometricUnsupported() {}

    func recoverySetupCompleted() {
        onComplete(displayName)
    }

    func recoverySetupSkipped() {
        onComplete(displayName)
    }

    func restoreTapped() {
        path.append(.restore)
    }

    func back() {
        guard !path.isEmpty else { return }
        path.removeLast()
    }
}

// MARK: - Root NavigationStack

struct FirstRunCoordinatorView: View {
    @State var coordinator: FirstRunCoordinator

    var body: some View {
        NavigationStack(path: $coordinator.path) {
            // Root: new single-page welcome (no carousel)
            EchoWelcomeView(
                onSetUp: { coordinator.welcomeContinueTapped() },
                onAlreadyHaveAccount: { coordinator.restoreTapped() }
            )
            .navigationDestination(for: FirstRunCoordinator.Route.self) { route in
                switch route {
                case .welcome:
                    EchoWelcomeView(
                        onSetUp: { coordinator.welcomeContinueTapped() },
                        onAlreadyHaveAccount: { coordinator.restoreTapped() }
                    )

                case .nameAndKey:
                    NameAndKeyView(
                        onComplete: { name, did in coordinator.nameAndKeyCompleted(username: name, did: did) },
                        onSkip: { coordinator.recoverySetupSkipped() }
                    )
                    .navigationBarBackButtonHidden(true)

                // Legacy routes kept for demo app
                case .displayName:
                    DisplayNameEntryView(coordinator: coordinator)

                case .biometricEnrollment:
                    BiometricEnrollmentView(
                        username: coordinator.displayName,
                        onComplete: { did in coordinator.biometricEnrollmentCompleted(did: did) },
                        onUnsupported: { coordinator.biometricUnsupported() }
                    )
                    .navigationBarBackButtonHidden(true)

                case .recoverySetup(let did):
                    RecoverySetupView(
                        did: did,
                        onComplete: { coordinator.recoverySetupCompleted() },
                        onSkip: { coordinator.recoverySetupSkipped() }
                    )
                    .navigationBarBackButtonHidden(true)

                case .restore:
                    RestoreFromPhraseView(
                        coordinator: RecoveryCoordinator(
                            onExportComplete: { },
                            onRestoreComplete: { coordinator.onRestoreComplete($0) },
                            onCancel: { coordinator.back() }
                        )
                    )
                }
            }
        }
        .tint(Color.Echo.primaryContainer)
    }
}

// MARK: - Display Name Validator

enum DisplayNameValidator {
    /// Trimmed length 1–32, Unicode letters + digits + space + hyphen + underscore + apostrophe.
    static func isValid(_ raw: String) -> Bool {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard (1...32).contains(trimmed.count) else { return false }
        let allowed = CharacterSet.letters
            .union(.decimalDigits)
            .union(CharacterSet(charactersIn: " -_'"))
        return trimmed.unicodeScalars.allSatisfy { allowed.contains($0) }
    }
}

#endif
