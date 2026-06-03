#if os(iOS)
import SwiftUI

/// Start a chat by @username search (WO-222 + Wave 0.1).
struct NewConversationSheet: View {
    @Environment(\.dismiss) private var dismiss

    let onConversationCreated: (StoredConversation) -> Void

    @State private var handle = ""
    @State private var results: [UsernameSearchHit] = []
    @State private var isSearching = false
    @State private var errorMessage: String?
    @State private var searchTask: Task<Void, Never>?

    var body: some View {
        NavigationStack {
            List {
                Section {
                    HStack {
                        TextField("@username", text: $handle)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()
                            .onChange(of: handle) { _, newValue in
                                scheduleSearch(newValue)
                            }
                        if isSearching {
                            ProgressView()
                        }
                    }
                } footer: {
                    Text("Search uses your authenticated session. Only public handles are returned.")
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.footnote)
                    }
                }

                Section("Results") {
                    if results.isEmpty && !handle.isEmpty && !isSearching {
                        Text("No users found")
                            .foregroundStyle(.secondary)
                    }
                    ForEach(results) { hit in
                        Button {
                            startConversation(with: hit)
                        } label: {
                            VStack(alignment: .leading, spacing: 4) {
                                Text("@\(hit.username)")
                                    .font(.headline)
                                Text(hit.did)
                                    .font(.caption2)
                                    .foregroundStyle(.secondary)
                                    .lineLimit(1)
                            }
                        }
                    }
                }
            }
            .navigationTitle("New conversation")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }

    private func scheduleSearch(_ query: String) {
        searchTask?.cancel()
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else {
            results = []
            return
        }
        searchTask = Task {
            try? await Task.sleep(nanoseconds: 350_000_000)
            guard !Task.isCancelled else { return }
            await runSearch(trimmed)
        }
    }

    private func runSearch(_ query: String) async {
        isSearching = true
        errorMessage = nil
        defer { isSearching = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = "API client unavailable"
            return
        }
        let social = ContactSocialAPIClient(apiClient: client)
        do {
            results = try await social.searchUsername(query)
        } catch {
            errorMessage = error.localizedDescription
            results = []
        }
    }

    private func startConversation(with hit: UsernameSearchHit) {
        Task { @MainActor in
            let localDID = await CurrentUserSession.currentDID() ?? ""
            let threadId = localDID.isEmpty
                ? UUID().uuidString
                : ConversationID.direct(localDID: localDID, peerDID: hit.did)
            let conversation = StoredConversation(
                id: threadId,
                contactName: "@\(hit.username)",
                peerDID: hit.did
            )
            ConversationStore.shared.upsert(conversation)
            if !UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage") {
                UserDefaults.standard.set(true, forKey: "echo.hasSentFirstMessage")
            }
            onConversationCreated(conversation)
            dismiss()
        }
    }
}
#endif
