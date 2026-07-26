#if os(iOS)
import SwiftUI

/// Routes into group or 1:1 chat depending on conversation id (M2).
struct ChatDestinationView: View {
    let conversation: StoredConversation

    @State private var chatViewModel: ChatDetailViewModel?
    @State private var groupViewModel: GroupChatViewModel?
    @State private var currentUserDID = ""

    private var isGroup: Bool {
        conversation.id.hasPrefix("group:") || conversation.peerDID.hasPrefix("grp-")
    }

    private var groupId: String? {
        if conversation.id.hasPrefix("group:") {
            return String(conversation.id.dropFirst("group:".count))
        }
        if conversation.peerDID.hasPrefix("grp-") {
            return conversation.peerDID
        }
        return nil
    }

    var body: some View {
        Group {
            if isGroup, let groupViewModel {
                GroupChatView(viewModel: groupViewModel)
            } else if let chatViewModel {
                ChatView(
                    viewModel: chatViewModel,
                    contactName: conversation.contactName,
                    conversationId: conversation.id,
                    peerDID: conversation.peerDID,
                    currentUserDID: currentUserDID,
                    onSendMessage: { text in
                        ConversationStore.shared.appendMessagePreview(
                            conversationId: conversation.id,
                            preview: text
                        )
                        if !UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage") {
                            UserDefaults.standard.set(true, forKey: "echo.hasSentFirstMessage")
                        }
                    }
                )
            } else {
                ProgressView("Opening chat…")
            }
        }
        .hidesGlacialTabBarWhenPushed(true)
        .task {
            currentUserDID = await CurrentUserSession.currentDID() ?? "did:key:unknown"
            if isGroup, let groupId {
                groupViewModel = makeGroupChatViewModel(groupId: groupId)
                if let token = try? await KeychainManager.shared.getAuthToken() {
                    await groupViewModel?.configure(accessToken: token)
                }
            } else {
                chatViewModel = DIContainer.shared.makeChatDetailViewModel()
            }
        }
    }

    private func makeGroupChatViewModel(groupId: String) -> GroupChatViewModel? {
        guard let keys = DIContainer.shared.resolveGroupKeyManager(),
              let senderKeys = DIContainer.shared.resolveGroupSenderKeys(),
              let signals = DIContainer.shared.resolveConversationSignalService(),
              let distribution = DIContainer.shared.resolveGroupKeyDistribution() else {
            return nil
        }
        return GroupChatViewModel(
            groupId: groupId,
            groupName: conversation.contactName,
            currentUserDID: currentUserDID,
            keyManager: keys,
            senderKeys: senderKeys,
            signalService: signals,
            keyDistribution: distribution
        )
    }
}
#endif
