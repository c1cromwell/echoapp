#if os(iOS)
import SwiftUI

/// Global on-device AI consent and audit transparency (WO-CA1 / M7).
struct PrivacyAISettingsView: View {
    @State private var consent = PrivacyAIConsentStore.load()
    @State private var auditEntries: [PrivacyAIAuditEntry] = []

    var body: some View {
        List {
            Section {
                Toggle("Smart replies", isOn: $consent.smartRepliesEnabled)
                Toggle("Thread summaries", isOn: $consent.summariesEnabled)
                Toggle("On-device translation", isOn: $consent.translationEnabled)
            } header: {
                Text("In-chat AI")
            } footer: {
                Text("Echo processes these features on your device. Message text never leaves your phone for AI unless you opt into future server-assist features.")
            }

            Section {
                NavigationLink("AI activity log") {
                    PrivacyAIAuditLogView(entries: auditEntries)
                }
                if !auditEntries.isEmpty {
                    Button("Clear log", role: .destructive) {
                        PrivacyAIAuditLog.clear()
                        auditEntries = []
                    }
                }
            } footer: {
                Text("The log records which features ran and when — never message content.")
            }
        }
        .navigationTitle("On-device AI")
        .onAppear { auditEntries = PrivacyAIAuditLog.load() }
        .onDisappear { PrivacyAIConsentStore.save(consent) }
        .onChange(of: consent.smartRepliesEnabled) { _, _ in PrivacyAIConsentStore.save(consent) }
        .onChange(of: consent.summariesEnabled) { _, _ in PrivacyAIConsentStore.save(consent) }
        .onChange(of: consent.translationEnabled) { _, _ in PrivacyAIConsentStore.save(consent) }
    }
}

private struct PrivacyAIAuditLogView: View {
    let entries: [PrivacyAIAuditEntry]

    private static let formatter: DateFormatter = {
        let f = DateFormatter()
        f.dateStyle = .short
        f.timeStyle = .short
        return f
    }()

    var body: some View {
        List(entries) { entry in
            VStack(alignment: .leading, spacing: 4) {
                Text(entry.feature.label)
                    .font(.system(size: 15, weight: .medium))
                HStack(spacing: 8) {
                    Text(Self.formatter.string(from: entry.timestamp))
                    if let ref = entry.conversationRef {
                        Text("chat \(ref)")
                    }
                    if entry.onDevice {
                        Text("on-device")
                    }
                }
                .font(.system(size: 12))
                .foregroundColor(.echoInk55)
            }
            .padding(.vertical, 2)
        }
        .navigationTitle("AI activity")
        .overlay {
            if entries.isEmpty {
                ContentUnavailableView(
                    "No activity yet",
                    systemImage: "sparkles",
                    description: Text("AI features log here when used in chats.")
                )
            }
        }
    }
}
#endif
