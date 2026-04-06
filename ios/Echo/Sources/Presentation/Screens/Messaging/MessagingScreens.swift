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
    @State private var messageText = ""
    @State private var messages: [ChatMessage] = [
        ChatMessage(id: "1", content: "Hey! How are you?", isSent: false, status: .read, timestamp: "10:30 AM"),
        ChatMessage(id: "2", content: "I'm doing great!", isSent: true, status: .read, timestamp: "10:31 AM"),
        ChatMessage(id: "3", content: "That's awesome! Want to grab coffee?", isSent: false, status: .read, timestamp: "10:32 AM")
    ]
    
    let contactName: String
    let onSendMessage: (String) -> Void
    
    public init(contactName: String = "", onSendMessage: @escaping (String) -> Void = { _ in }) {
        self.contactName = contactName
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
                
                // Messages List
                ScrollViewReader { proxy in
                    List {
                        ForEach(messages) { message in
                            MessageBubble(
                                message: message.content,
                                isSent: message.isSent,
                                status: message.status,
                                timestamp: message.timestamp
                            )
                            .listRowSeparator(.hidden)
                            .listRowInsets(.init())
                            .listRowBackground(Color.clear)
                            .id(message.id)
                        }
                    }
                    .listStyle(.plain)
                    .scrollContentBackground(.hidden)
                    .onChange(of: messages.count) { _ in
                        if let lastMessage = messages.last {
                            withAnimation {
                                proxy.scrollTo(lastMessage.id, anchor: .bottom)
                            }
                        }
                    }
                }
                
                // Message Input
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
                        
                        Button(action: {
                            if !messageText.isEmpty {
                                let newMessage = ChatMessage(
                                    id: UUID().uuidString,
                                    content: messageText,
                                    isSent: true,
                                    status: .sent,
                                    timestamp: "Now"
                                )
                                messages.append(newMessage)
                                onSendMessage(messageText)
                                messageText = ""
                            }
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
