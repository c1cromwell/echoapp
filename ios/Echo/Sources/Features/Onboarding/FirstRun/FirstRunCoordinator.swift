#if os(iOS)
// Features/Onboarding/FirstRun/FirstRunCoordinator.swift
// Wave 12 — 4-step onboarding:
//   1. Welcome carousel
//   2. Display name (username) entry
//   3. Biometric enrollment (mandatory Face ID / Touch ID)
//   4. Recovery setup (phrase + optional SMS — skippable)

import SwiftUI
import Observation

@MainActor
@Observable
final class FirstRunCoordinator {
    enum Route: Hashable {
        case welcome
        case displayName
        case biometricEnrollment
        case recoverySetup(did: String)
        case restore
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
        path.append(.displayName)
    }

    func displayNameSubmitted(_ name: String) {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard DisplayNameValidator.isValid(trimmed) else { return }
        displayName = trimmed
        path.append(.biometricEnrollment)
    }

    func biometricEnrollmentCompleted(did: String) {
        path.append(.recoverySetup(did: did))
    }

    func biometricUnsupported() {
        // Device has no biometrics — show error and block onboarding.
        // In Phase 1, biometric is mandatory.
    }

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
            WelcomeCarouselView(coordinator: coordinator)
                .navigationDestination(for: FirstRunCoordinator.Route.self) { route in
                    switch route {
                    case .welcome:
                        WelcomeCarouselView(coordinator: coordinator)

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
