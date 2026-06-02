#if os(iOS)
import SwiftUI
import CryptoKit

// Wave 12 — Onboarding Step 4: Recovery setup.
//
// Users can set up one or both of:
//   - Recovery phrase (24-word BIP-39) — primary; shown immediately
//   - SMS backup (optional) — stores H(phone) on backend; raw phone discarded
//
// Both are skippable — accessible later from Settings.

public struct RecoverySetupView: View {
    let did: String
    let onComplete: () -> Void
    let onSkip: () -> Void

    public init(did: String, onComplete: @escaping () -> Void = {}, onSkip: @escaping () -> Void = {}) {
        self.did = did
        self.onComplete = onComplete
        self.onSkip = onSkip
    }

    @State private var showPhraseSetup = false
    @State private var showSMSSetup = false
    @State private var phraseConfigured = false
    @State private var smsConfigured = false

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            VStack(spacing: 0) {
                EchoNavBar(title: "Back up your account")

                ScrollView {
                    VStack(spacing: 20) {
                        Text("Set up at least one recovery method so you can restore your account if you get a new device.")
                            .font(.system(size: 14))
                            .foregroundColor(.echoSecondaryText)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 24)
                            .padding(.top, 16)

                        // Recovery phrase card
                        RecoveryOptionCard(
                            icon: "key.horizontal",
                            title: "Recovery Phrase",
                            description: "24 words that let you restore your account on any device. Keep them offline and safe.",
                            isConfigured: phraseConfigured,
                            action: { showPhraseSetup = true }
                        )

                        // SMS backup card
                        RecoveryOptionCard(
                            icon: "message",
                            title: "SMS Backup (recommended)",
                            description: "Link your phone for recovery and to find friends on ECHO via private contact discovery. Only a hash is stored.",
                            isConfigured: smsConfigured,
                            action: { showSMSSetup = true }
                        )

                        Spacer(minLength: 32)

                        VStack(spacing: 12) {
                            EchoButton("Continue", style: .primary) {
                                UserDefaults.standard.set(true, forKey: "echoapp.recoveryConfigured")
                                onComplete()
                            }
                            .disabled(!phraseConfigured && !smsConfigured)

                            Button("Skip for now — I'll do this in Settings") {
                                onSkip()
                            }
                            .font(.system(size: 13))
                            .foregroundColor(.echoSecondaryText)
                        }
                        .padding(.horizontal, 24)
                        .padding(.bottom, 48)
                    }
                }
            }
        }
        .sheet(isPresented: $showPhraseSetup) {
            PhraseBackupSheet(onConfigured: {
                phraseConfigured = true
                showPhraseSetup = false
            })
        }
        .sheet(isPresented: $showSMSSetup) {
            SMSOTPSetupView(did: did, onConfigured: {
                smsConfigured = true
                showSMSSetup = false
            })
        }
    }
}

// MARK: - Recovery option card

private struct RecoveryOptionCard: View {
    let icon: String
    let title: String
    let description: String
    let isConfigured: Bool
    let action: () -> Void

    public var body: some View {
        Button(action: action) {
            HStack(spacing: 16) {
                ZStack {
                    Circle()
                        .fill(isConfigured ? Color(hex: 0x10B981).opacity(0.12) : Color.echoPrimary.opacity(0.10))
                        .frame(width: 48, height: 48)
                    Image(systemName: isConfigured ? "checkmark" : icon)
                        .font(.system(size: 20, weight: .medium))
                        .foregroundColor(isConfigured ? Color(hex: 0x10B981) : .echoPrimary)
                }
                VStack(alignment: .leading, spacing: 4) {
                    Text(title)
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundColor(.echoPrimaryText)
                    Text(description)
                        .font(.system(size: 13))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.leading)
                }
                Spacer()
                Image(systemName: isConfigured ? "checkmark.circle.fill" : "chevron.right")
                    .foregroundColor(isConfigured ? Color(hex: 0x10B981) : .echoSecondaryText)
            }
            .padding(16)
            .background(Color.echoCardBackground)
            .cornerRadius(12)
        }
        .padding(.horizontal, 24)
    }
}

// MARK: - Phrase backup sheet (wraps existing RecoveryPhraseDisplayView flow)

private struct PhraseBackupSheet: View {
    let onConfigured: () -> Void
    @Environment(\.dismiss) private var dismiss

    public var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Text("Your recovery phrase is 24 words. Write them down and store offline — never share them.")
                    .font(.system(size: 14))
                    .foregroundColor(.echoSecondaryText)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 24)
                    .padding(.top, 24)

                // RecoveryPhraseDisplayView handles actual phrase display + blur-on-capture.
                // In a full implementation, RecoveryCoordinator is pushed here.
                // For now, simulate phrase confirmation.
                EchoButton("I've saved my phrase", style: .primary) {
                    onConfigured()
                }
                .padding(.horizontal, 24)
            }
            .navigationTitle("Recovery Phrase")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }
}
#endif
