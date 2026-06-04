#if os(iOS)
import SwiftUI

/// Routes into `ChatView` with DI-backed `ChatDetailViewModel` (Wave 0.1).
struct ChatDestinationView: View {
    let conversation: StoredConversation

    @State private var viewModel: ChatDetailViewModel?
    @State private var currentUserDID = ""

    var body: some View {
        Group {
            if let viewModel {
                ChatView(
                    viewModel: viewModel,
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
                ProgressView("Opening secure thread…")
            }
        }
        .hidesGlacialTabBarWhenPushed(true)
        .task {
            currentUserDID = await CurrentUserSession.currentDID() ?? "did:key:unknown"
            viewModel = DIContainer.shared.makeChatDetailViewModel()
        }
    }
}
#endif
