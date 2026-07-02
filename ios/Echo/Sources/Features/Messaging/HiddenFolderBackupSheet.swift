#if os(iOS)
import SwiftUI
import UniformTypeIdentifiers

/// Hidden-folder backup phrase setup, create, and restore (WO-69).
struct HiddenFolderBackupSheet: View {
    @Environment(\.dismiss) private var dismiss

    @State private var words: [String] = Array(repeating: "", count: 24)
    @State private var challengePositions: [Int] = []
    @State private var challengeAnswers: [String: String] = [:]
    @State private var step: Step = .intro
    @State private var message: String?
    @State private var showImporter = false

    enum Step {
        case intro, enterPhrase, confirmWords, restore
    }

    var body: some View {
        NavigationStack {
            Form {
                switch step {
                case .intro:
                    Section {
                        Text("Backups are encrypted with your recovery phrase and stored only on this device (no iCloud).")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    }
                    Section("Schedule") {
                        Picker("Automatic backup", selection: Binding(
                            get: { HiddenFolderBackupScheduler.frequency },
                            set: { HiddenFolderBackupScheduler.frequency = $0 }
                        )) {
                            ForEach(HiddenFolderBackupScheduler.Frequency.allCases, id: \.self) { freq in
                                Text(freq.label).tag(freq)
                            }
                        }
                    }
                    Section {
                        Button("Set up recovery phrase") { step = .enterPhrase }
                        Button("Restore from backup file") { showImporter = true }
                    }
                case .enterPhrase:
                    Section("24-word recovery phrase") {
                        ForEach(0..<24, id: \.self) { index in
                            TextField("Word \(index + 1)", text: $words[index])
                                .textInputAutocapitalization(.never)
                                .autocorrectionDisabled()
                        }
                    }
                    Section {
                        Button("Continue") { beginConfirm() }
                    }
                case .confirmWords:
                    Section("Confirm phrase") {
                        ForEach(challengePositions, id: \.self) { position in
                            TextField("Word #\(position)", text: Binding(
                                get: { challengeAnswers[String(position)] ?? "" },
                                set: { challengeAnswers[String(position)] = $0 }
                            ))
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()
                        }
                    }
                    Section {
                        Button("Save phrase & create backup") { Task { await savePhraseAndBackup() } }
                    }
                case .restore:
                    EmptyView()
                }

                if let message {
                    Section {
                        Text(message).font(.footnote).foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Hidden backup")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
            .fileImporter(
                isPresented: $showImporter,
                allowedContentTypes: [.data],
                allowsMultipleSelection: false
            ) { result in
                guard case .success(let urls) = result, let url = urls.first else { return }
                step = .enterPhrase
                message = "Enter your recovery phrase, then we'll restore from the selected file."
                Task { await restore(from: url) }
            }
        }
    }

    private func beginConfirm() {
        let normalized = words.map { $0.lowercased().trimmingCharacters(in: .whitespaces) }
        guard let phrase = RecoveryPhrase(words: normalized) else {
            message = "Invalid recovery phrase."
            return
        }
        challengePositions = phrase.challengePositions()
        challengeAnswers = [:]
        step = .confirmWords
    }

    private func savePhraseAndBackup() async {
        let normalized = words.map { $0.lowercased().trimmingCharacters(in: .whitespaces) }
        guard let phrase = RecoveryPhrase(words: normalized) else {
            message = "Invalid recovery phrase."
            return
        }
        for position in challengePositions {
            let expected = phrase.word(at: position).lowercased()
            let answer = challengeAnswers[String(position)]?.lowercased().trimmingCharacters(in: .whitespaces) ?? ""
            guard answer == expected else {
                message = "Confirmation word #\(position) does not match."
                return
            }
        }
        do {
            _ = try await HiddenFolderBackupManager.createBackup(phrase: phrase)
            HiddenFolderBackupScheduler.markPhraseConfigured()
            message = "Backup created successfully."
            step = .intro
        } catch {
            message = error.localizedDescription
        }
    }

    private func restore(from url: URL) async {
        let normalized = words.map { $0.lowercased().trimmingCharacters(in: .whitespaces) }
        guard let phrase = RecoveryPhrase(words: normalized) else {
            message = "Enter a valid 24-word phrase first."
            return
        }
        guard url.startAccessingSecurityScopedResource() else { return }
        defer { url.stopAccessingSecurityScopedResource() }
        do {
            try await HiddenFolderBackupManager.restoreBackup(from: url, phrase: phrase)
            message = "Hidden chats restored."
            HiddenFolderBackupScheduler.markPhraseConfigured()
        } catch {
            message = error.localizedDescription
        }
    }
}
#endif
