#if os(iOS)
// Features/Personas/PersonaGateView.swift
// Biometric re-auth gate for hidden personas, folders, and chats.
//
// Supports three step-up methods via StepUpAuthManager:
//   .faceID         — SE key sign (default)
//   .passkey        — same SE key, different label
//   .devicePasscode — LAContext .deviceOwnerAuthentication
//
// The preferred method is read from StepUpAuthManager.shared.preferredMethod.
// A "Try another method" sheet is shown after the second failure.
//
// Security: auto-locks after 2 minutes in the background.

import SwiftUI

public struct PersonaGateView<Content: View>: View {
    let personaID: String
    @ViewBuilder let protectedContent: () -> Content

    @State private var isUnlocked = false
    @State private var isUnlocking = false
    @State private var errorMessage: String?
    @State private var unlockTime: Date?
    @State private var failureCount = 0
    @State private var showMethodPicker = false
    @Environment(\.scenePhase) private var scenePhase

    private let lockAfterBackground: TimeInterval = 120

    public init(personaID: String, @ViewBuilder protectedContent: @escaping () -> Content) {
        self.personaID = personaID
        self.protectedContent = protectedContent
    }

    public var body: some View {
        Group {
            if isUnlocked {
                protectedContent()
                    .onChange(of: scenePhase) { _, newPhase in
                        if newPhase == .active,
                           let t = unlockTime,
                           Date().timeIntervalSince(t) > lockAfterBackground {
                            isUnlocked = false
                            unlockTime = nil
                            failureCount = 0
                        }
                    }
            } else {
                gateScreen
            }
        }
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .background {
                unlockTime = isUnlocked ? Date() : nil
            }
        }
        .sheet(isPresented: $showMethodPicker) {
            methodPickerSheet
        }
    }

    // MARK: - Gate screen

    private var preferredMethod: StepUpMethod {
        StepUpAuthManager.shared.preferredMethod
    }

    private var gateScreen: some View {
        ZStack {
            Color.echoNight.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                // Method icon in double-ring
                ZStack {
                    Circle()
                        .stroke(Color.echoNightHair, lineWidth: 1)
                        .frame(width: 136, height: 136)
                    Circle()
                        .stroke(Color.echoNightHair.opacity(0.5), lineWidth: 1)
                        .frame(width: 120, height: 120)
                    Circle()
                        .fill(Color.echoNightHi)
                        .frame(width: 120, height: 120)
                    Image(systemName: preferredMethod.systemIcon)
                        .font(.system(size: 48, weight: .ultraLight))
                        .foregroundStyle(Color.echoNightInk)
                        .scaleEffect(isUnlocking ? 1.06 : 1.0)
                        .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true),
                                   value: isUnlocking)
                }
                .padding(.bottom, 28)

                Text("Verify to continue.")
                    .font(.system(size: 22, weight: .semibold))
                    .tracking(-0.5)
                    .foregroundStyle(Color.echoNightInk)

                Text("This area is protected. It locks again after two minutes in the background.")
                    .font(.system(size: 13.5))
                    .lineSpacing(4)
                    .foregroundStyle(Color.echoNightInk70)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 40)
                    .padding(.top, 10)

                // Status tag
                Group {
                    if let error = errorMessage {
                        Text(error)
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoAlert)
                    } else {
                        Text(isUnlocking ? "● scanning" : "● awaiting \(preferredMethod.displayName)")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoNightInk40)
                    }
                }
                .padding(.top, 36)

                Spacer()

                // Action buttons
                VStack(spacing: 4) {
                    HStack(spacing: 16) {
                        Button("Cancel") { /* dismiss handled by parent */ }
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoNightInk70)
                            .padding(8)

                        Button("Try again") { Task { await unlock() } }
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.echoNightInk)
                            .padding(8)
                            .disabled(isUnlocking)
                    }

                    // Show alternative method button after 2nd failure
                    if failureCount >= 2 {
                        Button("Try another method") { showMethodPicker = true }
                            .font(.system(size: 12))
                            .foregroundStyle(Color.echoNightInk40)
                            .padding(.top, 4)
                    }
                }
                .padding(.bottom, 48)
            }
        }
        .preferredColorScheme(.dark)
        .onAppear {
            Task {
                try? await Task.sleep(nanoseconds: 300_000_000)
                await unlock()
            }
        }
    }

    // MARK: - Method picker sheet

    private var methodPickerSheet: some View {
        VStack(spacing: 0) {
            Capsule()
                .fill(Color.echoHair)
                .frame(width: 36, height: 5)
                .padding(.top, 12)
                .padding(.bottom, 20)

            Text("Choose verification method")
                .font(.system(size: 16, weight: .semibold))
                .foregroundStyle(Color.echoInk)
                .padding(.bottom, 16)

            ForEach(StepUpMethod.allCases, id: \.rawValue) { method in
                if method.isAvailable {
                    Button {
                        showMethodPicker = false
                        Task {
                            try? await Task.sleep(nanoseconds: 200_000_000)
                            await unlock(override: method)
                        }
                    } label: {
                        HStack(spacing: 14) {
                            Image(systemName: method.systemIcon)
                                .font(.system(size: 20))
                                .foregroundStyle(Color.echoSignal)
                                .frame(width: 32)
                            Text(method.displayName)
                                .font(.system(size: 15))
                                .foregroundStyle(Color.echoInk)
                            Spacer()
                            if method == preferredMethod {
                                Text("Default")
                                    .font(.system(size: 11, weight: .medium))
                                    .foregroundStyle(Color.echoInk40)
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 3)
                                    .background(Color.echoHair, in: Capsule())
                            }
                        }
                        .padding(.horizontal, 24)
                        .padding(.vertical, 14)
                    }
                    Divider().padding(.leading, 70)
                }
            }

            Spacer()
        }
        .presentationDetents([.fraction(0.42)])
        .presentationDragIndicator(.hidden)
    }

    // MARK: - Unlock

    private func unlock(override method: StepUpMethod? = nil) async {
        guard !isUnlocking else { return }
        isUnlocking = true
        errorMessage = nil

        do {
            try await StepUpAuthManager.shared.authenticate(
                reason: "Access hidden content in Echo",
                override: method
            )
            isUnlocked = true
            unlockTime = Date()
            failureCount = 0
        } catch {
            failureCount += 1
            errorMessage = failureCount >= 2
                ? "Verification failed. Try another method below."
                : "Verification failed. Please try again."
        }

        isUnlocking = false
    }
}

// MARK: - Persona extension

extension Persona {
    var requiresGate: Bool { visibility == .hidden }
}
#endif
