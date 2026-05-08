// Features/Auth/Enrollment/Shared/EnrollmentTailView.swift
//
// Shared tail of every enrollment path. Once a branch produces a
// VerifiedIdentityBundle, this view animates the user through:
//   1. Secure Enclave P-256 key generation (biometric gate)
//   2. did:key registration with Echo API (POST /identity/register)
//   3. Constellation wallet creation via Stargazer SDK
//   4. WebAuthn passkey registration
// and then hands control back to AppState via `onComplete`.

import SwiftUI

@MainActor
@Observable
final class EnrollmentTailViewModel {
    var currentStep: Step = .creatingIdentity
    var didComplete = false
    var error: EnrollmentError?

    enum Step: Int, CaseIterable {
        case creatingIdentity   // Secure Enclave key + did:key
        case creatingWallet     // Stargazer
        case registeringPasskey // WebAuthn
        case done

        var title: String {
            switch self {
            case .creatingIdentity:   return "Creating your on-device identity"
            case .creatingWallet:     return "Setting up your ECHO wallet"
            case .registeringPasskey: return "Registering a passkey"
            case .done:               return "All set"
            }
        }
        var subtitle: String {
            switch self {
            case .creatingIdentity:
                return "Registering your did:key with Echo and linking it to your verified credential."
            case .creatingWallet:
                return "Generating your Constellation wallet via Stargazer SDK. Your keys never leave this device."
            case .registeringPasskey:
                return "Bind your account to Face ID so future sign-ins take one tap."
            case .done:
                return "Welcome to ECHO."
            }
        }
    }

    private let api = EnrollmentAPIClient.shared

    func run(bundle: VerifiedIdentityBundle, coordinator: EnrollmentCoordinator) async {
        do {
            coordinator.stage = .creatingDID
            currentStep = .creatingIdentity
            let did = try await api.registerDIDFromBundle(bundle)

            coordinator.stage = .creatingWallet
            currentStep = .creatingWallet
            let walletAddress = try await api.createWalletForDID(did: did)

            coordinator.stage = .registeringPasskey
            currentStep = .registeringPasskey
            try await api.registerPasskey(did: did)

            currentStep = .done
            didComplete = true

            // Short beat so the user sees the success state before transitioning.
            try? await Task.sleep(for: .milliseconds(700))

            _ = walletAddress
            coordinator.tailFinished(with: bundle)
        } catch let enrollError as EnrollmentError {
            self.error = enrollError
            coordinator.failed(enrollError)
        } catch {
            let mapped = EnrollmentError.transportFailed(underlying: error.localizedDescription)
            self.error = mapped
            coordinator.failed(mapped)
        }
    }
}

// MARK: - View

struct EnrollmentTailView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = EnrollmentTailViewModel()

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 0) {
                Spacer(minLength: 24)

                trustBadge
                    .frame(width: 100, height: 100)

                VStack(spacing: 8) {
                    Text(viewModel.currentStep.title)
                        .font(.system(size: 22, weight: .semibold))
                        .kerning(-0.3)
                        .foregroundStyle(Color.Echo.onSurface)
                        .multilineTextAlignment(.center)
                    Text(viewModel.currentStep.subtitle)
                        .font(.system(size: 14))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 28)
                }
                .padding(.top, 24)

                progressSteps
                    .padding(.horizontal, 24)
                    .padding(.top, 32)

                if let error = viewModel.error {
                    errorPanel(error).padding(.horizontal, 24).padding(.top, 24)
                }

                Spacer()
            }
        }
        .navigationTitle("Finishing up")
        .navigationBarTitleDisplayMode(.inline)
        .navigationBarBackButtonHidden(true)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .task {
            guard let bundle = coordinator.verifiedBundle else { return }
            await viewModel.run(bundle: bundle, coordinator: coordinator)
        }
    }

    private var trustBadge: some View {
        ZStack {
            LinearGradient(
                colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                startPoint: .topLeading, endPoint: .bottomTrailing
            )
            .clipShape(Circle())

            Image(systemName: viewModel.didComplete ? "checkmark.seal.fill" : "shield.lefthalf.filled")
                .font(.system(size: 36, weight: .semibold))
                .foregroundStyle(.white)
                .contentTransition(.symbolEffect(.replace))
        }
    }

    private var progressSteps: some View {
        VStack(spacing: 12) {
            ForEach(EnrollmentTailViewModel.Step.allCases, id: \.rawValue) { step in
                if step != .done {
                    stepRow(for: step)
                }
            }
        }
    }

    private func stepRow(for step: EnrollmentTailViewModel.Step) -> some View {
        let isDone = step.rawValue < viewModel.currentStep.rawValue
        let isCurrent = step == viewModel.currentStep && !viewModel.didComplete
        return HStack(spacing: 14) {
            ZStack {
                Circle().fill(
                    isDone ? Color.Echo.primaryContainer
                           : Color.Echo.primaryContainer.opacity(0.1)
                )
                if isDone {
                    Image(systemName: "checkmark")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundStyle(.white)
                } else if isCurrent {
                    ProgressView().controlSize(.mini).tint(Color.Echo.primaryContainer)
                }
            }
            .frame(width: 24, height: 24)

            Text(step.title)
                .font(.system(size: 14, weight: isCurrent ? .semibold : .regular))
                .foregroundStyle(isDone || isCurrent ? Color.Echo.onSurface : Color.Echo.onSurfaceVariant)

            Spacer()
        }
    }

    private func errorPanel(_ error: EnrollmentError) -> some View {
        VStack(spacing: 12) {
            Text(error.errorDescription ?? "Something went wrong")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurface)
                .multilineTextAlignment(.center)
            Button("Back to the picker") {
                coordinator.path = []
            }
            .font(.system(size: 14, weight: .semibold))
            .foregroundStyle(Color.Echo.primaryContainer)
        }
        .padding(16)
        .background(Color.red.opacity(0.06), in: RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .strokeBorder(Color.red.opacity(0.25), lineWidth: 0.5)
        )
    }
}
