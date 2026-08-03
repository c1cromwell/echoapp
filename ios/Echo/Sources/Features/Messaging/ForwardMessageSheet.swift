#if os(iOS)
import SwiftUI

/// Pick one or more conversations (including Saved Messages) to forward into.
struct ForwardMessageSheet: View {
    let messagePreview: String
    let sourceMessageId: String?
    let sourceConversationId: String?
    let excludingConversationId: String
    let onForwarded: () -> Void

    @Environment(\.dismiss) private var dismiss
    @Bindable private var conversationStore = ConversationStore.shared
    @State private var selectedIds: Set<String> = []
    @State private var isForwarding = false
    @State private var errorMessage: String?
    @State private var savedConversation: StoredConversation?
    @State private var localDID: String = ""

    private var destinations: [StoredConversation] {
        conversationStore.conversations.filter {
            $0.id != excludingConversationId
                && !ConversationPreferencesStore.shared.isHidden($0.id)
                && !SavedMessagesStore.isSavedMessages($0, localDID: localDID)
        }
    }

    var body: some View {
        NavigationStack {
            Group {
                if destinations.isEmpty && savedConversation == nil {
                    ContentUnavailableView(
                        "No other chats",
                        systemImage: "bubble.left.and.bubble.right",
                        description: Text("Start a conversation first, then forward.")
                    )
                } else {
                    List {
                        if let saved = savedConversation, saved.id != excludingConversationId {
                            destinationRow(saved, isSaved: true)
                        }
                        ForEach(destinations) { conversation in
                            destinationRow(conversation, isSaved: false)
                        }
                    }
                    .listStyle(.plain)
                }
            }
            .navigationTitle("Forward to")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                        .disabled(isForwarding)
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(selectedIds.isEmpty ? "Forward" : "Forward (\(selectedIds.count))") {
                        Task { await forwardSelected() }
                    }
                    .disabled(selectedIds.isEmpty || isForwarding)
                }
            }
            .safeAreaInset(edge: .bottom) {
                if let errorMessage {
                    Text(errorMessage)
                        .font(.footnote)
                        .foregroundColor(.echoAlert)
                        .padding()
                        .frame(maxWidth: .infinity)
                        .background(Color.echoPaper)
                }
            }
            .task {
                localDID = await CurrentUserSession.currentDID() ?? ""
                savedConversation = await SavedMessagesStore.ensureConversation()
            }
        }
    }

    @ViewBuilder
    private func destinationRow(_ conversation: StoredConversation, isSaved: Bool) -> some View {
        let selected = selectedIds.contains(conversation.id)
        Button {
            if selected {
                selectedIds.remove(conversation.id)
            } else {
                selectedIds.insert(conversation.id)
            }
        } label: {
            HStack(spacing: 12) {
                Image(systemName: selected ? "checkmark.circle.fill" : "circle")
                    .foregroundColor(selected ? .echoSignal : .echoInk40)
                    .font(.system(size: 22))

                ZStack {
                    Circle()
                        .fill(isSaved ? Color.echoSignal : NewConversationSheet.avatarColor(for: conversation.contactName))
                        .frame(width: 36, height: 36)
                    if isSaved {
                        Image(systemName: "bookmark.fill")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundColor(.white)
                    } else {
                        Text(String(conversation.contactName.prefix(1)).uppercased())
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundColor(.white)
                    }
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text(conversation.contactName)
                        .font(.system(size: 16, weight: .medium))
                        .foregroundColor(.echoInk)
                    Text(isSaved ? "Notes and forwards" : ContactThreadHelper.truncatedDID(conversation.peerDID))
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk55)
                }
                Spacer()
            }
        }
        .buttonStyle(.plain)
    }

    private func forwardSelected() async {
        isForwarding = true
        errorMessage = nil
        defer { isForwarding = false }

        let targets = conversationStore.conversations.filter { selectedIds.contains($0.id) }
        guard !targets.isEmpty else { return }

        for conversation in targets {
            await MessageForwarder.forward(
                content: messagePreview,
                to: conversation,
                fromMessageId: sourceMessageId,
                fromConversationId: sourceConversationId
            )
        }
        onForwarded()
        dismiss()
    }
}
#endif
