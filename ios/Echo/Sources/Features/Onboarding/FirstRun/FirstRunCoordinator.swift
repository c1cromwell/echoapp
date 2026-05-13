#if os(iOS)
// Features/Onboarding/FirstRun/FirstRunCoordinator.swift
// Onboarding flow:
//   1. EchoWelcomeView      — brand entry
//   2. DisplayNameEntryView — username only
//   3. OnboardingOptionsView — Face ID (mandatory) + VIP opt-in checkbox
//      • No VIP → RecoverySetupView → done (trustTier 0)
//      • VIP    → VIPPathView → digital ID or standard IDV → VIPSuccessView
//                             → RecoverySetupView → done (trustTier 2 or 4)

import SwiftUI
import Observation

@MainActor
@Observable
public final class FirstRunCoordinator {
    enum Route: Hashable {
        case welcome
        case displayName
        case onboardingOptions                          // Face ID card + VIP checkbox
        case vipPath(did: String)                       // choose Digital ID vs Standard IDV
        case vipStandardIDV(did: String)                // scan ID + selfie + phone
        case vipSuccess(did: String, trustTier: Int)    // trusted badge success screen
        case recoverySetup(did: String)
        case restore
        // Legacy routes kept for demo back-compat
        case nameAndKey
        case biometricEnrollment
    }

    var path: [Route] = []
    var displayName: String = ""

    let onComplete: (String, Int) -> Void       // (displayName, trustTier)
    let onRestoreComplete: (RestoredIdentity) -> Void

    public init(
        onComplete: @escaping (String, Int) -> Void,
        onRestoreComplete: @escaping (RestoredIdentity) -> Void
    ) {
        self.onComplete = onComplete
        self.onRestoreComplete = onRestoreComplete
    }

    // MARK: - Welcome

    func welcomeContinueTapped() {
        path.append(.displayName)
    }

    func restoreTapped() {
        path.append(.restore)
    }

    // MARK: - Username

    func displayNameSubmitted(_ name: String) {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard DisplayNameValidator.isValid(trimmed) else { return }
        displayName = trimmed
        path.append(.onboardingOptions)
    }

    // MARK: - Options (Face ID + VIP)

    func onboardingOptionsContinued(did: String, wantsVIP: Bool) {
        if wantsVIP {
            path.append(.vipPath(did: did))
        } else {
            path.append(.recoverySetup(did: did))
        }
    }

    // MARK: - VIP

    func vipPathCompleted(did: String, trustTier: Int) {
        path.append(.vipSuccess(did: did, trustTier: trustTier))
    }

    func vipSkipped(did: String) {
        path.append(.recoverySetup(did: did))
    }

    func vipSuccessContinued(did: String) {
        path.append(.recoverySetup(did: did))
    }

    // MARK: - Recovery

    func recoverySetupCompleted() {
        onComplete(displayName, storedTrustTier)
    }

    func recoverySetupSkipped() {
        onComplete(displayName, storedTrustTier)
    }

    // Trust tier is stashed here when VIPSuccessView completes
    private var storedTrustTier: Int = 0
    func storeTrustTier(_ tier: Int) { storedTrustTier = tier }

    // MARK: - Legacy back-compat helpers (NameAndKeyView, BiometricEnrollmentView)

    func nameAndKeyCompleted(username: String, did: String) {
        displayName = username
        path.append(.recoverySetup(did: did))
    }

    func biometricEnrollmentCompleted(did: String) {
        path.append(.recoverySetup(did: did))
    }

    func biometricUnsupported() {}

    func back() {
        guard !path.isEmpty else { return }
        path.removeLast()
    }
}

// MARK: - Root NavigationStack

public struct FirstRunCoordinatorView: View {
    @State var coordinator: FirstRunCoordinator

    public init(coordinator: FirstRunCoordinator) {
        _coordinator = State(wrappedValue: coordinator)
    }

    public var body: some View {
        NavigationStack(path: $coordinator.path) {
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

                case .displayName:
                    DisplayNameEntryView(coordinator: coordinator)

                case .onboardingOptions:
                    OnboardingOptionsView(coordinator: coordinator)
                        .navigationBarBackButtonHidden(true)

                case .vipPath(let did):
                    VIPPathView(
                        did: did,
                        onDigitalID: { trustTier in coordinator.vipPathCompleted(did: did, trustTier: trustTier) },
                        onStandardIDV: { trustTier in coordinator.vipPathCompleted(did: did, trustTier: trustTier) },
                        onSkip: { coordinator.vipSkipped(did: did) }
                    )
                    .navigationBarBackButtonHidden(true)

                case .vipStandardIDV(let did):
                    VIPStandardIDVView(did: did) { trustTier in
                        coordinator.vipPathCompleted(did: did, trustTier: trustTier)
                    }
                    .navigationBarBackButtonHidden(true)

                case .vipSuccess(let did, let trustTier):
                    VIPSuccessView(trustTier: trustTier) {
                        coordinator.storeTrustTier(trustTier)
                        coordinator.vipSuccessContinued(did: did)
                    }
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
                            onExportComplete: {},
                            onRestoreComplete: { coordinator.onRestoreComplete($0) },
                            onCancel: { coordinator.back() }
                        )
                    )

                // Legacy routes
                case .nameAndKey:
                    NameAndKeyView(
                        onComplete: { name, did in coordinator.nameAndKeyCompleted(username: name, did: did) },
                        onSkip: { coordinator.recoverySetupSkipped() }
                    )
                    .navigationBarBackButtonHidden(true)

                case .biometricEnrollment:
                    BiometricEnrollmentView(
                        username: coordinator.displayName,
                        onComplete: { did in coordinator.biometricEnrollmentCompleted(did: did) },
                        onUnsupported: { coordinator.biometricUnsupported() }
                    )
                    .navigationBarBackButtonHidden(true)
                }
            }
        }
        .tint(Color.echoSignal)
    }
}

// MARK: - Display Name Validator

enum DisplayNameValidator {
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
