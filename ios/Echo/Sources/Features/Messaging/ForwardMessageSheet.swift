#if os(iOS)
import SwiftUI

/// Pick a conversation to forward message text into (local thread + WS send).
struct ForwardMessageSheet: View {
    let messagePreview: String
    let sourceMessageId: String?
    let sourceConversationId: String?
    let excludingConversationId: String
    let onForwarded: () -> Void

    @Environment(\.dismiss) private var dismiss
    @Bindable private var conversationStore = ConversationStore.shared
    @State private var errorMessage: String?

    private var destinations: [StoredConversation] {
        conversationStore.conversations.filter {
            $0.id != excludingConversationId
                && !ConversationPreferencesStore.shared.isHidden($0.id)
        }
    }

    var body: some View {
        NavigationStack {
            Group {
                if destinations.isEmpty {
                    ContentUnavailableView(
                        "No other chats",
                        systemImage: "bubble.left.and.bubble.right",
                        description: Text("Start a conversation first, then forward.")
                    )
                } else {
                    List(destinations) { conversation in
                        Button {
                            Task {
                                await MessageForwarder.forward(
                                    content: messagePreview,
                                    to: conversation,
                                    fromMessageId: sourceMessageId,
                                    fromConversationId: sourceConversationId
                                )
                                onForwarded()
                                dismiss()
                            }
                        } label: {
                            HStack(spacing: 12) {
                                Text(String(conversation.contactName.prefix(1)).uppercased())
                                    .font(.system(size: 14, weight: .semibold))
                                    .foregroundColor(.white)
                                    .frame(width: 36, height: 36)
                                    .background(NewConversationSheet.avatarColor(for: conversation.contactName))
                                    .clipShape(Circle())
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(conversation.contactName)
                                        .font(.system(size: 16, weight: .medium))
                                        .foregroundColor(.echoInk)
                                    Text(ContactThreadHelper.truncatedDID(conversation.peerDID))
                                        .font(.system(size: 12))
                                        .foregroundColor(.echoInk55)
                                }
                                Spacer()
                                Image(systemName: "arrowshape.turn.up.right")
                                    .foregroundColor(.echoSignal)
                            }
                        }
                        .buttonStyle(.plain)
                    }
                    .listStyle(.plain)
                }
            }
            .navigationTitle("Forward to")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
            if let errorMessage {
                Text(errorMessage)
                    .font(.footnote)
                    .foregroundColor(.echoAlert)
                    .padding()
            }
        }
    }
}
#endif
