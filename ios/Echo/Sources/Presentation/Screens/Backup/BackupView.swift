#if os(iOS)
// Presentation/Screens/Backup/BackupView.swift
// Backup & security screen with recovery phrase, encrypted backup, and export options

import SwiftUI

// MARK: - Backup View

public struct BackupView: View {
    @StateObject private var viewModel = BackupViewModel()

    public init() {}

    public var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Recovery phrase section
                GhostBorderSection(title: "RECOVERY PHRASE") {
                    HStack(spacing: 12) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundStyle(Color.Echo.warning)
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Keep this phrase secret")
                                .font(Font.Echo.bodyMedium).fontWeight(.semibold)
                            Text("Never share it with anyone")
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.outline)
                        }
                    }

                    Button {
                        viewModel.showRecoveryPhrase = true
                    } label: {
                        Text("View Recovery Phrase")
                            .font(.system(size: 14)).fontWeight(.bold)
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Capsule().fill(LinearGradient.signature))
                    }
                    .deepGlacialShadow()

                    Text("Requires biometric authentication")
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.outline)
                }

                // Encrypted backup
                GhostBorderSection(title: "ENCRYPTED BACKUP") {
                    InfoRow(label: "Last backup", value: viewModel.lastBackupDate)
                    InfoRow(label: "Size", value: viewModel.backupSize)
                    InfoRow(label: "Location", value: viewModel.backupLocation)

                    if let status = viewModel.statusMessage {
                        Text(status)
                            .font(Font.Echo.labelMd)
                            .foregroundStyle(viewModel.statusIsError ? Color.Echo.error : Color.Echo.primaryContainer)
                    }

                    Button("Back Up Now") {
                        viewModel.phrasePromptMode = .backup
                        viewModel.showPhrasePrompt = true
                    }
                    .font(.system(size: 14)).fontWeight(.bold)
                    .foregroundStyle(Color.Echo.primaryContainer)
                    .disabled(viewModel.isWorking)

                    Button("Restore from Cloud") {
                        viewModel.phrasePromptMode = .restoreCloud
                        viewModel.showPhrasePrompt = true
                    }
                    .font(.system(size: 14))
                    .foregroundStyle(Color.Echo.onSurface)
                    .disabled(viewModel.isWorking)

                    Button("Restore from Local File") {
                        viewModel.phrasePromptMode = .restoreLocal
                        viewModel.showPhrasePrompt = true
                    }
                    .font(.system(size: 14))
                    .foregroundStyle(Color.Echo.onSurface)
                    .disabled(viewModel.isWorking || !viewModel.hasLocalBackup)
                }

                // Auto-backup settings
                GhostBorderSection(title: "AUTO-BACKUP") {
                    Toggle("Enable automatic backup", isOn: $viewModel.autoBackupEnabled)
                        .font(Font.Echo.bodyMedium)
                        .onChange(of: viewModel.autoBackupEnabled) { _, on in
                            viewModel.persistAutoBackupSettings(enabled: on)
                        }

                    Picker("Frequency", selection: $viewModel.selectedFrequency) {
                        ForEach(BackupScheduler.Frequency.allCases, id: \.self) { freq in
                            Text(freq.displayName).tag(freq)
                        }
                    }
                    .onChange(of: viewModel.selectedFrequency) { _, value in
                        viewModel.persistFrequency(value)
                    }

                    Toggle("Wi‑Fi only", isOn: $viewModel.wifiOnly)
                        .font(Font.Echo.bodyMedium)
                        .onChange(of: viewModel.wifiOnly) { _, value in
                            BackupScheduler.wifiOnly = value
                        }

                    SettingsRow(icon: "photo", label: "Include Media", value: viewModel.includeMedia ? "Yes" : "No")
                }

                // Export options
                GhostBorderSection(title: "EXPORT DATA") {
                    ExportButton(label: "Export Chat History", icon: "text.bubble")
                    ExportButton(label: "Export Contacts", icon: "person.2")
                    ExportButton(label: "Export Identity (DID)", icon: "person.text.rectangle")
                }

                // Danger zone
                VStack(spacing: 16) {
                    Text("DANGER ZONE")
                        .font(.system(size: 10))
                        .fontWeight(.bold).tracking(2)
                        .foregroundStyle(Color.Echo.error)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.leading, 28)

                    Button("Delete All Data") {
                        viewModel.showDeleteConfirmation = true
                    }
                    .font(.system(size: 14)).fontWeight(.bold)
                    .foregroundStyle(Color.Echo.error)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(
                        RoundedRectangle(cornerRadius: 9999)
                            .stroke(Color.Echo.error.opacity(0.3), lineWidth: 1)
                    )
                    .padding(.horizontal, 20)
                }
            }
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Backup & Security")
        .sheet(isPresented: $viewModel.showPhrasePrompt) {
            BackupPhrasePromptSheet(viewModel: viewModel)
        }
        .onAppear { viewModel.refreshMetadata() }
    }
}

// MARK: - Backup ViewModel

@MainActor
class BackupViewModel: ObservableObject {
    enum PhrasePromptMode {
        case backup
        case restoreCloud
        case restoreLocal
    }

    @Published var lastBackupDate: String = "Never"
    @Published var backupSize: String = "—"
    @Published var backupLocation: String = "Local + Cloud"
    @Published var includeMedia = false
    @Published var wifiOnly = true
    @Published var autoBackupEnabled = false
    @Published var selectedFrequency: BackupScheduler.Frequency = .manual
    @Published var hasLocalBackup = false
    @Published var showRecoveryPhrase = false
    @Published var showDeleteConfirmation = false
    @Published var showPhrasePrompt = false
    @Published var phrasePromptMode: PhrasePromptMode = .backup
    @Published var phraseText = ""
    @Published var isWorking = false
    @Published var statusMessage: String?
    @Published var statusIsError = false

    func refreshMetadata() {
        #if os(iOS)
        guard let service = DIContainer.shared.resolveMessageBackup() else { return }
        hasLocalBackup = service.hasLocalBackup
        autoBackupEnabled = BackupScheduler.autoBackupEnabled
        selectedFrequency = BackupScheduler.frequency
        wifiOnly = BackupScheduler.wifiOnly
        if let date = service.lastBackupDate {
            let formatter = RelativeDateTimeFormatter()
            lastBackupDate = formatter.localizedString(for: date, relativeTo: Date())
            backupSize = ByteCountFormatter.string(fromByteCount: Int64(service.lastBackupByteCount), countStyle: .file)
        } else {
            lastBackupDate = "Never"
            backupSize = "—"
        }
        #endif
    }

    func persistFrequency(_ frequency: BackupScheduler.Frequency) {
        BackupScheduler.frequency = frequency
    }

    func persistAutoBackupSettings(enabled: Bool) {
        BackupScheduler.autoBackupEnabled = enabled
    }

    func submitPhrase() async {
        #if os(iOS)
        let words = phraseText
            .lowercased()
            .components(separatedBy: .whitespacesAndNewlines)
            .filter { !$0.isEmpty }
        guard let phrase = RecoveryPhrase(words: words) else {
            statusMessage = "Enter all 24 recovery words."
            statusIsError = true
            return
        }
        guard let service = DIContainer.shared.resolveMessageBackup() else { return }

        isWorking = true
        statusMessage = nil
        defer {
            isWorking = false
            showPhrasePrompt = false
            phraseText = ""
        }

        do {
            switch phrasePromptMode {
            case .backup:
                try await service.uploadCloudBackup(phrase: phrase)
                try await service.createLocalBackup(phrase: phrase)
                if BackupScheduler.autoBackupEnabled {
                    try await BackupSessionKeyStore.save(from: phrase)
                }
                statusMessage = "Backup saved locally and uploaded to cloud."
                statusIsError = false
            case .restoreCloud:
                let count = try await service.restoreCloudBackup(phrase: phrase)
                statusMessage = "Restored \(count) conversations from cloud backup."
                statusIsError = false
            case .restoreLocal:
                let count = try await service.restoreLocalBackup(phrase: phrase)
                statusMessage = "Restored \(count) conversations from local backup."
                statusIsError = false
            }
            refreshMetadata()
        } catch {
            statusMessage = error.localizedDescription
            statusIsError = true
        }
        #endif
    }
}

// MARK: - Phrase prompt

private struct BackupPhrasePromptSheet: View {
    @ObservedObject var viewModel: BackupViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 16) {
                Text(viewModel.phrasePromptMode == .backup
                     ? "Enter your 24-word recovery phrase to encrypt this backup."
                     : viewModel.phrasePromptMode == .restoreCloud
                        ? "Enter your recovery phrase to decrypt and restore from cloud."
                        : "Enter your recovery phrase to restore from the local backup file.")
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.outline)

                TextEditor(text: $viewModel.phraseText)
                    .font(.system(.caption, design: .monospaced))
                    .frame(minHeight: 160)
                    .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color.Echo.outline.opacity(0.3)))

                if viewModel.isWorking {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                } else {
                    Button("Continue") {
                        Task { await viewModel.submitPhrase() }
                    }
                    .buttonStyle(.borderedProminent)
                    .frame(maxWidth: .infinity)
                }
            }
            .padding()
            .navigationTitle("Recovery Phrase")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }
}

// MARK: - Info Row

struct InfoRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
            Spacer()
            Text(value)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
        }
    }
}

// MARK: - Export Button

struct ExportButton: View {
    let label: String
    let icon: String

    var body: some View {
        Button {
            // TODO: Trigger export
        } label: {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundStyle(Color.Echo.primaryContainer)
                    .frame(width: 24)
                Text(label)
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.onSurface)
                Spacer()
                Image(systemName: "arrow.down.circle")
                    .foregroundStyle(Color.Echo.outline)
            }
        }
    }
}
#endif
