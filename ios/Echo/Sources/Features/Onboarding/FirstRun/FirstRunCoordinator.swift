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
    }

    var path: [Route] = []
    var displayName: String = ""

    /// Called when the user finishes the two-page flow.
    let onComplete: (String) -> Void
    let onRestoreTapped: () -> Void

    init(
        onComplete: @escaping (String) -> Void,
        onRestoreTapped: @escaping () -> Void
    ) {
        self.onComplete = onComplete
        self.onRestoreTapped = onRestoreTapped
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

    func restoreTapped() {
        onRestoreTapped()
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
