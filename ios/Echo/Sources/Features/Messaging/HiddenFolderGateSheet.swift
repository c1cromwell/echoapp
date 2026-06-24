#if os(iOS)
import SwiftUI

/// Biometric gate with optional duress PIN entry (WO-7 §5.5).
struct HiddenFolderGateSheet: View {
    let onUnlocked: () -> Void
    let onCancel: () -> Void

    @State private var pin = ""
    @State private var pinError: String?
    @State private var showPINField = false
    @FocusState private var pinFocused: Bool

    var body: some View {
        NavigationStack {
            ZStack {
                Color.echoNight.ignoresSafeArea()

                VStack(spacing: 0) {
                    if showPINField {
                        pinEntry
                    } else {
                        PersonaGateView(personaID: "hidden-chats") {
                            Color.clear.onAppear { finishUnlock(duress: false) }
                        }
                    }
                }
            }
            .preferredColorScheme(.dark)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel", action: onCancel)
                        .foregroundStyle(Color.echoNightInk70)
                }
                if HiddenFolderSettingsStore.shared.hasDuressPIN {
                    ToolbarItem(placement: .primaryAction) {
                        Button(showPINField ? "Use Face ID" : "Use PIN") {
                            showPINField.toggle()
                            pin = ""
                            pinError = nil
                            pinFocused = showPINField
                        }
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(Color.echoNightInk70)
                    }
                }
            }
        }
    }

    private var pinEntry: some View {
        VStack(spacing: 20) {
            Spacer()
            Image(systemName: "lock.rectangle")
                .font(.system(size: 44, weight: .ultraLight))
                .foregroundStyle(Color.echoNightInk)
            Text("Enter PIN")
                .font(.system(size: 22, weight: .semibold))
                .foregroundStyle(Color.echoNightInk)
            Text("Unlock hidden chats with your PIN.")
                .font(.system(size: 13.5))
                .foregroundStyle(Color.echoNightInk70)

            SecureField("PIN", text: $pin)
                .keyboardType(.numberPad)
                .textContentType(.oneTimeCode)
                .multilineTextAlignment(.center)
                .font(.system(size: 28, weight: .medium, design: .rounded))
                .foregroundStyle(Color.echoNightInk)
                .padding()
                .background(Color.echoNightHi)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .padding(.horizontal, 48)
                .focused($pinFocused)

            if let pinError {
                Text(pinError)
                    .font(.echomono(11))
                    .foregroundStyle(Color.echoAlert)
            }

            Button("Unlock") { submitPIN() }
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(Color.echoNightInk)
                .padding(.top, 8)

            Spacer()
        }
        .onAppear { pinFocused = true }
    }

    private func submitPIN() {
        pinError = nil
        if HiddenFolderSettingsStore.shared.matchesDuressPIN(pin) {
            finishUnlock(duress: true)
            return
        }
        pinError = "Incorrect PIN."
        pin = ""
    }

    private func finishUnlock(duress: Bool) {
        HiddenChatsSession.shared.unlock(biometric: !duress, duress: duress)
        onUnlocked()
    }
}
#endif
