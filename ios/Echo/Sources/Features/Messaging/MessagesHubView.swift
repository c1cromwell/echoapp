import SwiftUI

/// Messages hub (spec §3.2). Vertical order: persona header → search → pinned →
/// trust folder chips → conversation list → groups/channels/hidden segments.
struct MessagesHubView: View {
    enum Segment: String, CaseIterable, Identifiable {
        case chats, groups, channels, hidden
        var id: String { rawValue }
        var title: String {
            switch self {
            case .chats: return "Chats"
            case .groups: return "Groups"
            case .channels: return "Channels"
            case .hidden: return "Hidden"
            }
        }
    }

    let conversations: [StoredConversation]
    let personas: [PersonaSummary]
    let activePersona: PersonaSummary
    let trustTier: (String) -> Int
    let mutedIDs: Set<String>

    let onSelectConversation: (String) -> Void
    let onCompose: () -> Void
    let onOpenHidden: () -> Void
    let onSwitchPersona: (PersonaSummary) -> Void
    let onSelectHiddenPersona: (PersonaSummary) -> Void

    @Bindable private var pinnedStore = PinnedConversationsStore.shared
    @State private var searchText = ""
    @State private var folder: ChatFolder = .all
    @State private var segment: Segment = .chats
    @State private var showPersonaSheet = false
    @State private var showEditPins = false
    @State private var showIntegrityExplainer = false

    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                header
                secureBar
                searchField
                segmentBar

                ScrollView {
                    LazyVStack(spacing: 0) {
                        switch segment {
                        case .chats:   chatsContent
                        case .groups:  placeholder(icon: "person.3", text: "Groups arrive with the social-graph phase.")
                        case .channels: placeholder(icon: "dot.radiowaves.left.and.right", text: "Broadcast channels are coming after groups.")
                        case .hidden:  hiddenSegment
                        }
                    }
                }
            }
        }
        .sheet(isPresented: $showPersonaSheet) {
            PersonaSwitcherSheet(
                personas: personas,
                activeID: activePersona.id,
                onSelect: { showPersonaSheet = false; onSwitchPersona($0) },
                onSelectHidden: { showPersonaSheet = false; onSelectHiddenPersona($0) }
            )
        }
        .sheet(isPresented: $showEditPins) {
            EditPinnedSheet(conversations: conversations)
        }
        .sheet(isPresented: $showIntegrityExplainer) {
            NavigationStack {
                IntegrityExplainerView()
                    .toolbar {
                        ToolbarItem(placement: .cancellationAction) {
                            Button("Done") { showIntegrityExplainer = false }
                        }
                    }
            }
        }
    }

    private var pinnedItems: [PinnedItem] {
        MessagesHubSupport.pinnedItems(
            conversations: conversations,
            orderedPinIDs: pinnedStore.orderedIDs
        )
    }

    private var folderUnreadCounts: [ChatFolder: Int] {
        MessagesHubSupport.folderUnreadCounts(
            conversations: conversations,
            trustTier: { trustTier($0) }
        )
    }

    private var filtered: [StoredConversation] {
        let unpinned = conversations.filter { !pinnedStore.isPinned($0.id) }
        return unpinned.filter { conv in
            (searchText.isEmpty || conv.contactName.localizedCaseInsensitiveContains(searchText))
                && folder.includes(tier: trustTier(conv.id))
        }
        .sorted { ($0.unreadCount > 0) && ($1.unreadCount == 0) }
    }

    private var header: some View {
        HStack(spacing: 0) {
            PersonaSwitcherHeader(active: activePersona) { showPersonaSheet = true }
            Button(action: onCompose) {
                Image(systemName: "square.and.pencil")
                    .font(.system(size: 19))
                    .foregroundColor(.echoSignal)
                    .padding(.trailing, Spacing.lg.rawValue)
            }
            .buttonStyle(.plain)
        }
    }

    private var secureBar: some View {
        Button {
            showIntegrityExplainer = true
        } label: {
            SecureThreadIndicatorBar()
        }
        .buttonStyle(.plain)
        .accessibilityLabel("Secure thread — learn about integrity")
    }

    private var searchField: some View {
        HStack(spacing: 10) {
            Image(systemName: "magnifyingglass").font(.system(size: 14)).foregroundColor(.echoInk40)
            TextField("Search conversations", text: $searchText)
                .font(.system(size: 15))
            if !searchText.isEmpty {
                Button { searchText = "" } label: {
                    Image(systemName: "xmark.circle.fill").foregroundColor(.echoInk40)
                }
                .buttonStyle(.plain)
            }
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, Spacing.sm.rawValue)
    }

    private var segmentBar: some View {
        Picker("", selection: $segment) {
            ForEach(Segment.allCases) { seg in Text(seg.title).tag(seg) }
        }
        .pickerStyle(.segmented)
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.bottom, Spacing.sm.rawValue)
    }

    @ViewBuilder
    private var chatsContent: some View {
        if !pinnedItems.isEmpty {
            PinnedSectionView(
                items: pinnedItems,
                maxItems: PinnedConversationsStore.maxPins,
                onItemTap: { onSelectConversation($0.id) },
                onEditTap: { showEditPins = true }
            )
            .padding(.vertical, Spacing.sm.rawValue)
        }

        ChatFolderFilterView(selection: $folder, unreadCounts: folderUnreadCounts)

        if filtered.isEmpty && pinnedItems.isEmpty {
            placeholder(icon: "bubble.left.and.bubble.right", text: "No conversations in this filter.")
        } else {
            ForEach(filtered) { conv in
                conversationRow(conv)
                Divider().background(Color.echoHair).padding(.leading, 76)
            }
        }
    }

    @ViewBuilder
    private func conversationRow(_ conv: StoredConversation) -> some View {
        HStack(spacing: 0) {
            ConversationListItem(
                contactName: conv.contactName,
                lastMessage: conv.lastMessage,
                timestamp: conv.timestamp,
                unreadCount: conv.unreadCount,
                isOnline: conv.isOnline,
                onTap: { onSelectConversation(conv.id) }
            )
            if mutedIDs.contains(conv.id) {
                Image(systemName: "bell.slash.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.echoInk40)
                    .padding(.trailing, Spacing.lg.rawValue)
            }
        }
        .contextMenu {
            if pinnedStore.isPinned(conv.id) {
                Button {
                    pinnedStore.unpin(conv.id)
                } label: {
                    Label("Unpin", systemImage: "pin.slash")
                }
            } else {
                Button {
                    pinnedStore.pin(conv.id)
                } label: {
                    Label("Pin", systemImage: "pin")
                }
                .disabled(pinnedStore.orderedIDs.count >= PinnedConversationsStore.maxPins)
            }
            Button {
                onSelectConversation(conv.id)
            } label: {
                Label("Open chat", systemImage: "bubble.left")
            }
        }
    }

    private var hiddenSegment: some View {
        Button(action: onOpenHidden) {
            VStack(spacing: 12) {
                Image(systemName: "eye.slash.fill").font(.system(size: 34)).foregroundColor(.echoInk40)
                Text("Hidden folders")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.echoInk70)
                Text("Unlock with Face ID to reveal private conversations.")
                    .font(.system(size: 13))
                    .foregroundColor(.echoInk40)
                    .multilineTextAlignment(.center)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 60)
            .padding(.horizontal, Spacing.xl.rawValue)
        }
        .buttonStyle(.plain)
    }

    private func placeholder(icon: String, text: String) -> some View {
        VStack(spacing: 12) {
            Image(systemName: icon).font(.system(size: 40)).foregroundColor(.echoInk40)
            Text(text)
                .font(.system(size: 14))
                .foregroundColor(.echoInk55)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
        .padding(.horizontal, Spacing.xl.rawValue)
    }
}

private struct SecureThreadIndicatorBar: View {
    var body: some View {
        Rectangle()
            .fill(Color.echoSignal.opacity(0.85))
            .frame(height: 2)
    }
}

#if DEBUG
struct MessagesHubView_Previews: PreviewProvider {
    static var previews: some View {
        MessagesHubView(
            conversations: [
                StoredConversation(contactName: "Aria Rao", peerDID: "did:key:a", lastMessage: "Signed receipt ✓", timestamp: "9:32", unreadCount: 3, isOnline: true),
                StoredConversation(contactName: "Kai Mercer", peerDID: "did:key:k", lastMessage: "Typing…", timestamp: "8:58"),
            ],
            personas: [PersonaSummary(id: "default", name: "Aria (public)", initials: "AR")],
            activePersona: PersonaSummary(id: "default", name: "Aria (public)", initials: "AR"),
            trustTier: { _ in 3 }
        )
    }
}
#endif
