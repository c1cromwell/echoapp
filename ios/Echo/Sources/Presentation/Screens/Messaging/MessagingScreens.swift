import SwiftUI

/// Messaging - Conversation List Screen
struct ConversationListView: View {
    @State private var searchText = ""
    let conversations: [StoredConversation]
    @State private var pinnedItems: [PinnedItem] = []

    let onSelectConversation: (String) -> Void
    let onPinnedItemTap: (PinnedItem) -> Void
    let onEditPinned: () -> Void

    init(
        conversations: [StoredConversation] = [],
        onSelectConversation: @escaping (String) -> Void = { _ in },
        onPinnedItemTap: @escaping (PinnedItem) -> Void = { _ in },
        onEditPinned: @escaping () -> Void = {},
        pinnedItems: [PinnedItem] = []
    ) {
        self.conversations = conversations
        self.onSelectConversation = onSelectConversation
        self.onPinnedItemTap = onPinnedItemTap
        self.onEditPinned = onEditPinned
        self._pinnedItems = State(initialValue: pinnedItems)
    }
    
    var filteredConversations: [StoredConversation] {
        if searchText.isEmpty {
            return conversations.sorted { ($0.unreadCount > 0) && ($1.unreadCount == 0) }
        }
        return conversations.filter { $0.contactName.localizedCaseInsensitiveContains(searchText) }
    }
    
    var body: some View {
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
                            .foregroundColor(.echoInk40)

                        TextField("Search conversations", text: $searchText)
                            .textFieldStyle(.roundedBorder)

                        if !searchText.isEmpty {
                            Button(action: { searchText = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .font(.system(size: 14))
                                    .foregroundColor(.echoInk40)
                            }
                        }
                    }
                    .padding(Spacing.md.rawValue)
                    .background(Color.echoPaperDim)
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
                                .foregroundColor(.echoInk40)

                            Text("No conversations yet")
                                .typographyStyle(.h4, color: .echoInk70)

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

// StoredConversation lives in Services/ConversationStore.swift (Wave 0.1).

// MARK: - Chat View

struct ChatView: View {
    @Environment(\.dismiss) var dismiss
    @Bindable var viewModel: ChatDetailViewModel
    @State private var messageText = ""
    @State private var reactionTargetId: String?
    @State private var showChatSettings = false
    @State private var actionsTargetMessage: ChatDetailMessage?

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

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: contactName,
                    showBackButton: true,
                    onBackPressed: { dismiss() },
                    trailingAction: { showChatSettings = true },
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
                                    HStack(spacing: 6) {
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

                                        // "More" → full Telegram-style action grid (§5.3)
                                        Button {
                                            withAnimation(.glacialPress) { reactionTargetId = nil }
                                            actionsTargetMessage = message
                                        } label: {
                                            Image(systemName: "ellipsis")
                                                .font(.system(size: 17, weight: .semibold))
                                                .foregroundColor(.echoInk70)
                                                .frame(width: 40, height: 40)
                                                .background(Color.echoPaper)
                                                .clipShape(Circle())
                                                .overlay(Circle().stroke(Color.echoHair, lineWidth: 1))
                                        }
                                        .buttonStyle(.plain)
                                        .accessibilityLabel("More actions")
                                    }
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
                            messageText = ""
                            Task { await viewModel.sendMessage(text) }
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
        .sheet(isPresented: $showChatSettings) {
            ChatSettingsSheet(
                contactName: contactName,
                preferences: ConversationPreferencesStore.shared.preferences(for: conversationId),
                onChange: { ConversationPreferencesStore.shared.save($0, for: conversationId) }
            )
            .presentationDetents([.medium, .large])
        }
        .sheet(item: $actionsTargetMessage) { message in
            MessageActionsSheet(
                messagePreview: message.content,
                isOwnMessage: message.isFromCurrentUser,
                sentWithinEditWindow: message.isFromCurrentUser,
                onAction: { action in
                    handleMessageAction(action, on: message)
                    actionsTargetMessage = nil
                }
            )
            .presentationDetents([.height(300)])
        }
        .task {
            let reactions: ReactionsAPI? = DIContainer.shared.resolveReactionsAPI()
            let privacy = MessagingPrivacyPreferences.merged(
                global: PrivacySettingsStore.load(),
                persona: PersonaPrivacySettings()
            )
            viewModel.configure(
                conversationId: conversationId,
                peerDID: peerDID,
                currentUserDID: currentUserDID,
                privacy: privacy,
                reactionsAPI: reactions,
                onSend: onSendMessage
            )
            if let token = try? await KeychainManager.shared.getAuthToken() {
                await viewModel.connect(accessToken: token)
            }
        }
        .onDisappear {
            Task { await viewModel.disconnect() }
        }
        .onTapGesture {
            if reactionTargetId != nil {
                withAnimation(.glacialPress) { reactionTargetId = nil }
            }
        }
    }

    /// Handle a message action. Copy is handled inside the sheet (UIPasteboard);
    /// Delete removes locally; Reply/Forward/Pin/Edit are wired to ViewModel hooks
    /// in a later pass.
    private func handleMessageAction(_ action: MessageAction, on message: ChatDetailMessage) {
        switch action {
        case .delete:
            viewModel.messages.removeAll { $0.id == message.id }
        case .copy, .reply, .forward, .pin, .edit:
            break
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
