#if os(iOS)
// Features/Onboarding/Recovery/RecoveryCoordinator.swift

import SwiftUI
import Observation

@MainActor
@Observable
public final class RecoveryCoordinator {
    enum Route: Hashable {
        case phraseDisplay
        case phraseConfirm
        case restore
    }

    var path: [Route] = []
    var phrase: RecoveryPhrase?
    var challengePositions: [Int] = []
    var userAnswers: [Int: String] = [:]

    let onExportComplete: () -> Void
    let onRestoreComplete: (RestoredIdentity) -> Void
    let onCancel: () -> Void

    public init(
        onExportComplete: @escaping () -> Void,
        onRestoreComplete: @escaping (RestoredIdentity) -> Void,
        onCancel: @escaping () -> Void
    ) {
        self.onExportComplete = onExportComplete
        self.onRestoreComplete = onRestoreComplete
        self.onCancel = onCancel
    }

    // MARK: - Export flow

    func startExport() async throws {
        let phrase = try await StargazerBridgeForRecovery.shared.exportRecoveryPhrase()
        self.phrase = phrase
        self.challengePositions = phrase.challengePositions()
        self.userAnswers = [:]
        path.append(.phraseDisplay)
    }

    func phraseDisplayContinued() {
        path.append(.phraseConfirm)
    }

    func confirmationSucceeded() {
        UserDefaults.standard.set(Date(), forKey: "echo.recoveryPhraseExportedAt")
        RecoveryPromptScheduler.shared.cancelAllPendingReminders()
        phrase = nil
        onExportComplete()
    }

    func confirmationFailed() {
        userAnswers = [:]
        // Return to phrase display so the user can re-read the words.
        path.removeAll()
        path.append(.phraseDisplay)
    }

    // MARK: - Restore flow

    func startRestore() {
        path.append(.restore)
    }
}

// MARK: - Root NavigationStack

public struct RecoveryCoordinatorView: View {
    @State var coordinator: RecoveryCoordinator

    public init(coordinator: RecoveryCoordinator) {
        _coordinator = State(wrappedValue: coordinator)
    }

    public var body: some View {
        NavigationStack(path: $coordinator.path) {
            // RecoveryCoordinatorView is always entered via a sheet; the root
            // is either the phrase display (export flow) or the restore form.
            Color.clear
                .navigationDestination(for: RecoveryCoordinator.Route.self) { route in
                    switch route {
                    case .phraseDisplay:
                        RecoveryPhraseDisplayView(coordinator: coordinator)
                    case .phraseConfirm:
                        RecoveryPhraseConfirmView(coordinator: coordinator)
                    case .restore:
                        RestoreFromPhraseView(coordinator: coordinator)
                    }
                }
        }
        .tint(Color.Echo.primaryContainer)
    }
}

// MARK: - Restored identity

public struct RestoredIdentity: Sendable {
    let did: String
    let walletAddress: String
    let displayName: String
}

// MARK: - Stargazer bridge extension for recovery

/// Adapts the existing StargazerBridge actor for recovery-specific operations.
/// In production this delegates to the Stargazer SDK's BIP-39 export/import APIs.
actor StargazerBridgeForRecovery {
    static let shared = StargazerBridgeForRecovery()

    private init() {}

    /// Returns the 24-word BIP-39 recovery phrase for the current wallet.
    /// The phrase is retrieved from Keychain where it was stored during provisioning
    /// (kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly).
    func exportRecoveryPhrase() async throws -> RecoveryPhrase {
        // Production: read BIP-39 seed from Keychain, convert to mnemonic words.
        // Stub returns a deterministic test phrase so the UI can be exercised in Simulator.
        let stubWords = [
            "abandon", "ability", "able", "about", "above", "absent",
            "absorb", "abstract", "absurd", "abuse", "access", "accident",
            "account", "accuse", "achieve", "acid", "acoustic", "acquire",
            "across", "act", "action", "actor", "actress", "actual"
        ]
        guard let phrase = RecoveryPhrase(words: stubWords) else {
            throw RecoveryError.walletDerivationFailed
        }
        return phrase
    }

    /// Restores the Constellation wallet from a BIP-39 phrase and returns the wallet address.
    func restoreWallet(from phrase: RecoveryPhrase) async throws -> (address: String, publicKey: String) {
        // Production: derive HD wallet from phrase entropy via Stargazer SDK.
        guard phrase.entropy != nil else {
            throw RecoveryError.walletDerivationFailed
        }
        return (address: "DAG_restored_placeholder", publicKey: "se_pubkey_placeholder")
    }
}

#endif
