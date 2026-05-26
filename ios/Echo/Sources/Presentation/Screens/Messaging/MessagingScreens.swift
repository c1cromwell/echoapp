import SwiftUI

import SwiftUI

/// Messaging - Conversation List Screen
public struct ConversationListView: View {
    @State private var searchText = ""
    @State private var conversations: [ConversationItem] = [
        ConversationItem(id: "1", contactName: "John Doe", lastMessage: "That sounds great!", timestamp: "2:45 PM", unreadCount: 3, isOnline: true),
        ConversationItem(id: "2", contactName: "Jane Smith", lastMessage: "See you tomorrow", timestamp: "Yesterday", unreadCount: 0, isOnline: false),
        ConversationItem(id: "3", contactName: "Alice Johnson", lastMessage: "Perfect! Thanks for the help", timestamp: "Mon", unreadCount: 1, isOnline: true)
    ]
    @State private var pinnedItems: [PinnedItem] = []

    let onSelectConversation: (String) -> Void
    let onPinnedItemTap: (PinnedItem) -> Void
    let onEditPinned: () -> Void

    public init(
        onSelectConversation: @escaping (String) -> Void = { _ in },
        onPinnedItemTap: @escaping (PinnedItem) -> Void = { _ in },
        onEditPinned: @escaping () -> Void = {},
        pinnedItems: [PinnedItem] = []
    ) {
        self.onSelectConversation = onSelectConversation
        self.onPinnedItemTap = onPinnedItemTap
        self.onEditPinned = onEditPinned
        self._pinnedItems = State(initialValue: pinnedItems)
    }
    
    var filteredConversations: [ConversationItem] {
        if searchText.isEmpty {
            return conversations.sorted { ($0.unreadCount > 0) && ($1.unreadCount == 0) }
        }
        return conversations.filter { $0.contactName.localizedCaseInsensitiveContains(searchText) }
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Messages",
                    showBackButton: false
                )
                
                VStack(spacing: Spacing.lg.rawValue) {
                    // Search Bar
                    HStack {
                        Image(systemName: "magnifyingglass")
                            .font(.system(size: 14))
                            .foregroundColor(.echoGray500)
                        
                        TextField("Search conversations", text: $searchText)
                            .textFieldStyle(.roundedBorder)
                        
                        if !searchText.isEmpty {
                            Button(action: { searchText = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .font(.system(size: 14))
                                    .foregroundColor(.echoGray400)
                            }
                        }
                    }
                    .padding(Spacing.md.rawValue)
                    .background(Color.echoSurface)
                    .cornerRadius(12)

                    // Pinned Section
                    PinnedSectionView(
                        items: pinnedItems,
                        onItemTap: { item in onPinnedItemTap(item) },
                        onEditTap: onEditPinned
                    )

                    // Section Divider
                    if !pinnedItems.isEmpty {
                        SectionDivider(title: "ALL MESSAGES")
                    }

                    if filteredConversations.isEmpty {
                        VStack(spacing: Spacing.md.rawValue) {
                            Image(systemName: "bubble.left.and.bubble.right")
                                .font(.system(size: 48))
                                .foregroundColor(.echoGray400)
                            
                            Text("No conversations yet")
                                .typographyStyle(.h4, color: .echoGray600)
                            
                            Text("Start messaging with your contacts")
                                .typographyStyle(.body, color: .echoSecondaryText)
                        }
                        .frame(maxHeight: .infinity, alignment: .center)
                    } else {
                        List {
                            ForEach(filteredConversations) { conversation in
                                ConversationListItem(
                                    contactName: conversation.contactName,
                                    lastMessage: conversation.lastMessage,
                                    timestamp: conversation.timestamp,
                                    unreadCount: conversation.unreadCount,
                                    isOnline: conversation.isOnline,
                                    onTap: { onSelectConversation(conversation.id) }
                                )
                                .listRowSeparator(.hidden)
                                .listRowInsets(.init())
                                .listRowBackground(Color.clear)
                                .padding(.vertical, Spacing.xs.rawValue)
                            }
                        }
                        .listStyle(.plain)
                        .scrollContentBackground(.hidden)
                    }
                }
                .echoSpacing(.lg)
            }
        }
    }
}

struct ConversationItem: Identifiable {
    let id: String
    let contactName: String
    let lastMessage: String
    let timestamp: String
    let unreadCount: Int
    let isOnline: Bool
}

// MARK: - Chat View

public struct ChatView: View {
    @Environment(\.dismiss) var dismiss
    @Bindable var viewModel: ChatDetailViewModel
    @State private var messageText = ""
    @State private var reactionTargetId: String?

    let contactName: String
    let conversationId: String
    let peerDID: String
    let currentUserDID: String
    let onSendMessage: (String) -> Void

    init(
        viewModel: ChatDetailViewModel,
        contactName: String = "",
        conversationId: String = "",
        peerDID: String = "",
        currentUserDID: String = "",
        onSendMessage: @escaping (String) -> Void = { _ in }
    ) {
        self.viewModel = viewModel
        self.contactName = contactName
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.currentUserDID = currentUserDID
        self.onSendMessage = onSendMessage
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: contactName,
                    showBackButton: true,
                    onBackPressed: { dismiss() },
                    trailingAction: {},
                    trailingIcon: Image(systemName: "info.circle")
                )

                SecureThreadIndicator()

                ScrollViewReader { proxy in
                    List {
                        ForEach(viewModel.messages) { message in
                            VStack(
                                alignment: message.isFromCurrentUser ? .trailing : .leading,
                                spacing: 4
                            ) {
                                MessageBubble(
                                    message: message.content,
                                    isSent: message.isFromCurrentUser,
                                    status: mapDeliveryStatus(message.deliveryStatus),
                                    timestamp: message.timestamp
                                )
                                .onAppear {
                                    if !message.isFromCurrentUser {
                                        Task {
                                            await viewModel.onMessageAppeared(
                                                messageId: message.id,
                                                senderDID: message.senderDID
                                            )
                                        }
                                    }
                                }
                                .onLongPressGesture {
                                    withAnimation(.glacialPress) {
                                        reactionTargetId = reactionTargetId == message.id ? nil : message.id
                                    }
                                }

                                if !message.reactions.isEmpty {
                                    ReactionChipsView(
                                        reactions: message.reactions,
                                        currentUserDID: viewModel.currentUserDID,
                                        onTap: { emoji in
                                            Task {
                                                await viewModel.toggleReaction(
                                                    messageId: message.id,
                                                    emoji: emoji
                                                )
                                            }
                                        }
                                    )
                                }
                            }
                            .listRowSeparator(.hidden)
                            .listRowInsets(.init())
                            .listRowBackground(Color.clear)
                            .id(message.id)
                            .overlay(alignment: message.isFromCurrentUser ? .bottomTrailing : .bottomLeading) {
                                if reactionTargetId == message.id {
                                    ReactionPickerView(
                                        onSelect: { emoji in
                                            Task {
                                                await viewModel.toggleReaction(
                                                    messageId: message.id,
                                                    emoji: emoji
                                                )
                                            }
                                            withAnimation(.glacialPress) { reactionTargetId = nil }
                                        },
                                        onDismiss: {
                                            withAnimation(.glacialPress) { reactionTargetId = nil }
                                        }
                                    )
                                    .transition(.scale.combined(with: .opacity))
                                    .offset(y: -44)
                                }
                            }
                        }
                    }
                    .listStyle(.plain)
                    .scrollContentBackground(.hidden)
                    .onChange(of: viewModel.messages.count) {
                        if let last = viewModel.messages.last {
                            withAnimation {
                                proxy.scrollTo(last.id, anchor: .bottom)
                            }
                        }
                    }
                }

                if viewModel.peerIsTyping {
                    TypingIndicatorView()
                        .transition(.opacity.combined(with: .move(edge: .bottom)))
                        .padding(.vertical, 4)
                }

                VStack(spacing: Spacing.sm.rawValue) {
                    HStack(spacing: Spacing.md.rawValue) {
                        Button(action: {}) {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 24))
                                .foregroundColor(.echoPrimary)
                        }
                        .accessibility(label: Text("Add attachment"))

                        TextField("Message...", text: $messageText)
                            .textFieldStyle(.roundedBorder)
                            .onChange(of: messageText) { _, newValue in
                                viewModel.onInputChanged(newValue)
                            }

                        Button(action: {
                            guard !messageText.isEmpty else { return }
                            let text = messageText
                            let newMessage = ChatDetailMessage(
                                id: UUID().uuidString,
                                senderDID: currentUserDID,
                                currentUserDID: currentUserDID,
                                content: text,
                                timestamp: "Now",
                                deliveryStatus: .sending
                            )
                            viewModel.messages.append(newMessage)
                            onSendMessage(text)
                            messageText = ""
                            Task { await viewModel.onSendTapped() }
                        }) {
                            Image(systemName: "paperplane.fill")
                                .font(.system(size: 18))
                                .foregroundColor(.echoPrimary)
                        }
                        .accessibility(label: Text("Send message"))
                    }
                    .padding(Spacing.md.rawValue)
                    .background(Color.echoSurface)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
        .task {
            let reactions: ReactionsAPI? = DIContainer.shared.resolveReactionsAPI()
            viewModel.configure(
                conversationId: conversationId,
                peerDID: peerDID,
                currentUserDID: currentUserDID,
                reactionsAPI: reactions
            )
            if let token = try? await KeychainManager.shared.getAuthToken() {
                await viewModel.connect(accessToken: token)
            }
        }
        .onTapGesture {
            if reactionTargetId != nil {
                withAnimation(.glacialPress) { reactionTargetId = nil }
            }
        }
    }

    private func mapDeliveryStatus(_ status: DeliveryStatus?) -> MessageStatus {
        guard let status else { return .sent }
        switch status {
        case .sending:   return .sending
        case .sent:      return .sent
        case .delivered:  return .delivered
        case .read:      return .read
        case .failed:    return .failed
        case .anchored:  return .anchored
        case .verified:  return .verified
        }
    }
}

struct ChatMessage: Identifiable {
    let id: String
    let content: String
    let isSent: Bool
    let status: MessageStatus
    let timestamp: String
}

// MARK: - Preview

#if DEBUG
struct MessagingScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            ConversationListView()
        }
    }
}
#endif
