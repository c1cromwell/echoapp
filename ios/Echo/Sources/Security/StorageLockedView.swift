#if os(iOS)
// Security/StorageLockedView.swift
// WO-224: Shown when LocalDatabase.isLocked == true (storage key not yet derived).
// The user must authenticate with biometrics to re-derive the HKDF storage key.

import SwiftUI

/// Full-screen gate shown when the local SwiftData store is not yet unlocked.
/// Mirrors the lock screen pattern from Signal / WhatsApp for end-to-end key management.
public struct StorageLockedView: View {
    let onUnlock: () async -> Void

    @State private var isUnlocking = false
    @State private var errorMessage: String?

    public init(onUnlock: @escaping () async -> Void) {
        self.onUnlock = onUnlock
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 32) {
                Spacer()

                // Icon
                ZStack {
                    Circle()
                        .fill(Color.echoPrimary.opacity(0.12))
                        .frame(width: 100, height: 100)
                    Image(systemName: "lock.doc.fill")
                        .font(.system(size: 44, weight: .semibold))
                        .foregroundColor(.echoPrimary)
                }

                // Title and message
                VStack(spacing: 12) {
                    Text("Authenticate to access your messages")
                        .font(.system(size: 20, weight: .bold))
                        .foregroundColor(.echoPrimaryText)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 24)

                    Text("Your messages are protected with a key derived from your biometrics. Authenticate to unlock.")
                        .font(.system(size: 14))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 32)
                }

                // Error (if any)
                if let error = errorMessage {
                    Text(error)
                        .font(.system(size: 13))
                        .foregroundColor(.echoError)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                }

                Spacer()

                // Unlock button
                VStack(spacing: 12) {
                    EchoButton(
                        isUnlocking ? "Authenticating…" : "Unlock",
                        style: .primary
                    ) {
                        guard !isUnlocking else { return }
                        isUnlocking = true
                        errorMessage = nil
                        Task {
                            await onUnlock()
                            isUnlocking = false
                        }
                    }

                    Text("Your message content never leaves your device unencrypted.")
                        .font(.system(size: 11))
                        .foregroundColor(.echoSecondaryText)
                        .multilineTextAlignment(.center)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 48)
            }
        }
    }
}

#Preview {
    StorageLockedView(onUnlock: {})
}
#endif
