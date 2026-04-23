// Features/Onboarding/FirstRun/FirstRunCoordinator.swift
// Session-style cold-start coordinator.
// Routes: 1. Welcome carousel  2. Display name entry  3. Silent provisioning (background)

import SwiftUI
import Observation

@MainActor
@Observable
final class FirstRunCoordinator {
    enum Route: Hashable {
        case welcome
        case displayName
        case restore
    }

    var path: [Route] = []
    var displayName: String = ""

    /// Called when the user finishes the two-page first-run flow.
    let onComplete: (String) -> Void
    /// Called when account restore completes successfully.
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
        onComplete(trimmed)
    }

    // Navigates within the first-run NavigationStack to the restore flow (AC-MSG-003.8).
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
