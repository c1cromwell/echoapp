#if os(iOS)
import SwiftUI

/// Privacy settings for hidden chats (device-local, WO-7).
struct HiddenFolderSettingsView: View {
    @Environment(\.dismiss) private var dismiss

    @State private var autoLockChoice: AutoLockChoice = .twoMinutes
    @State private var lockOnScreenshot = true
    @State private var notificationMode: HiddenNotificationMode = .suppressed
    @State private var duressPIN = ""
    @State private var duressPINConfirm = ""
    @State private var settingsMessage: String?

    private let store = HiddenFolderSettingsStore.shared

    enum AutoLockChoice: TimeInterval, CaseIterable, Identifiable {
        case thirtySeconds = 30
        case twoMinutes = 120
        case fiveMinutes = 300

        var id: TimeInterval { rawValue }

        var label: String {
            switch self {
            case .thirtySeconds: return "30 seconds"
            case .twoMinutes: return "2 minutes"
            case .fiveMinutes: return "5 minutes"
            }
        }
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Auto-lock") {
                    Picker("Lock after", selection: $autoLockChoice) {
                        ForEach(AutoLockChoice.allCases) { choice in
                            Text(choice.label).tag(choice)
                        }
                    }
                    Toggle("Lock on screenshot", isOn: $lockOnScreenshot)
                }

                Section("Notifications") {
                    Picker("Hidden chat alerts", selection: $notificationMode) {
                        ForEach(HiddenNotificationMode.allCases, id: \.self) { mode in
                            Text(mode.label).tag(mode)
                        }
                    }
                    Text("Suppressed: no previews or badge. Redacted: \"New message\" only while unlocked.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                Section("Duress PIN") {
                    SecureField("New PIN (4–8 digits)", text: $duressPIN)
                        .keyboardType(.numberPad)
                    SecureField("Confirm PIN", text: $duressPINConfirm)
                        .keyboardType(.numberPad)
                    if store.hasDuressPIN {
                        Button("Remove duress PIN", role: .destructive) {
                            store.clearDuressPIN()
                            duressPIN = ""
                            duressPINConfirm = ""
                            settingsMessage = "Duress PIN removed."
                        }
                    }
                    Text("Under coercion, enter this PIN to open an empty hidden vault.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                if let settingsMessage {
                    Section {
                        Text(settingsMessage)
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Hidden chats")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") {
                        save()
                        dismiss()
                    }
                }
            }
            .onAppear { load() }
        }
    }

    private func load() {
        let seconds = store.autoLockSeconds
        autoLockChoice = AutoLockChoice.allCases.first { $0.rawValue == seconds } ?? .twoMinutes
        lockOnScreenshot = store.lockOnScreenshot
        notificationMode = store.notificationMode
    }

    private func save() {
        store.autoLockSeconds = autoLockChoice.rawValue
        store.lockOnScreenshot = lockOnScreenshot
        store.notificationMode = notificationMode

        let pin = duressPIN.trimmingCharacters(in: .whitespacesAndNewlines)
        let confirm = duressPINConfirm.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !pin.isEmpty || !confirm.isEmpty else { return }
        guard pin == confirm else {
            settingsMessage = "PINs do not match."
            return
        }
        do {
            try store.setDuressPIN(pin)
            duressPIN = ""
            duressPINConfirm = ""
            settingsMessage = "Duress PIN saved."
        } catch {
            settingsMessage = error.localizedDescription
        }
    }
}
#endif
