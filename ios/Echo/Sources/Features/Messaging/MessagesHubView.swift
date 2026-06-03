import SwiftUI

/// Messages hub (spec §3.2 + docs/design-previews/messagehub1.png). Layout:
/// persona switcher → "Messages" title with search / hidden-folders / new-message
/// icons → segment bar (Chats · Groups · Pinned · Channels) → trust folder chips →
/// conversation list. Search is icon-toggled (no persistent search bar); hidden
/// folders open via the folder icon (biometric-gated).
struct MessagesHubView: View {
    enum Segment: String, CaseIterable, Identifiable {
        case chats, groups, pinned, channels
        var id: String { rawValue }
        var title: String {
            switch self {
            case .chats: return "Chats"
            case .groups: return "Groups"
            case .pinned: return "Pinned"
            case .channels: return "Channels"
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
    @State private var showSearch = false
    @State private var folder: ChatFolder = .all
    @State private var segment: Segment = .chats
    @State private var showPersonaSheet = false
    @State private var showEditPins = false
    @State private var showIntegrityExplainer = false
    @State private var showCreateGroup = false
    @FocusState private var searchFocused: Bool

    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                header
                secureBar
                Color.clear.frame(height: 12) // breathing room under the secure line
                if showSearch { searchField }
                segmentBar

                ScrollView {
                    LazyVStack(spacing: 0) {
                        switch segment {
                        case .chats:   chatsContent
                        case .groups:  groupsContent
                        case .pinned:  pinnedContent
                        case .channels: placeholder(icon: "dot.radiowaves.left.and.right", text: "Broadcast channels are coming after groups.")
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
        .sheet(isPresented: $showCreateGroup) {
            NavigationStack {
                VStack(spacing: 14) {
                    Image(systemName: "person.3.fill").font(.system(size: 40)).foregroundColor(.echoInk40)
                    Text("Create a group").font(.system(size: 20, weight: .semibold)).foregroundColor(.echoInk)
                    Text("Name your group and add trusted contacts. Group conversations arrive with the social-graph phase.")
                        .font(.system(size: 14)).foregroundColor(.echoInk55)
                        .multilineTextAlignment(.center)
                }
                .padding(Spacing.xl.rawValue)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(Color.echoPaper)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Done") { showCreateGroup = false }
                    }
                }
            }
            .presentationDetents([.medium])
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
        VStack(spacing: 0) {
            // Persona switcher (tap to switch active persona)
            PersonaSwitcherHeader(active: activePersona) { showPersonaSheet = true }

            // Title row: "Messages" + search / new-message icons (see messagehub1.png)
            HStack(spacing: Spacing.sm.rawValue) {
                Text("Messages")
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(.echoInk)
                Spacer()

                Button {
                    withAnimation(.spring(response: 0.35, dampingFraction: 0.85)) {
                        showSearch.toggle()
                    }
                    if showSearch {
                        searchFocused = true
                    } else {
                        searchText = ""
                    }
                } label: {
                    headerIcon("magnifyingglass", filled: showSearch)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Search conversations")

                // Hidden folders — biometric-gated (folder icon next to search)
                Button(action: onOpenHidden) {
                    headerIcon("folder", filled: false)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Hidden folders")

                Button(action: onCompose) {
                    headerIcon("square.and.pencil", filled: true)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("New message")
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.top, 2)
            .padding(.bottom, Spacing.sm.rawValue)
        }
    }

    /// Circular header icon button. `filled` = signal-blue fill / white glyph (new
    /// message + active search); otherwise a paper-dim circle with an ink glyph.
    private func headerIcon(_ systemName: String, filled: Bool) -> some View {
        Image(systemName: systemName)
            .font(.system(size: 17, weight: .medium))
            .foregroundColor(filled ? .white : .echoInk70)
            .frame(width: 38, height: 38)
            .background(filled ? Color.echoSignal : Color.echoPaperDim)
            .clipShape(Circle())
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
            HStack(spacing: 10) {
                Image(systemName: "magnifyingglass").font(.system(size: 14)).foregroundColor(.echoInk40)
                TextField("Search conversations", text: $searchText)
                    .font(.system(size: 15))
                    .focused($searchFocused)
                    .submitLabel(.search)
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

            Button("Cancel") {
                withAnimation(.spring(response: 0.35, dampingFraction: 0.85)) {
                    showSearch = false
                }
                searchText = ""
                searchFocused = false
            }
            .font(.system(size: 15))
            .foregroundColor(.echoSignal)
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, Spacing.sm.rawValue)
        .transition(.move(edge: .top).combined(with: .opacity))
    }

    /// Tabbed bar — only the selected tab's content shows below. Selected tab is a
    /// dark pill; unselected are muted text (see docs/design-previews/messagehub1.png).
    private var segmentBar: some View {
        HStack(spacing: 4) {
            ForEach(Segment.allCases) { seg in
                let isSelected = segment == seg
                Button {
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) { segment = seg }
                } label: {
                    Text(seg.title)
                        .font(.system(size: 14, weight: isSelected ? .semibold : .medium))
                        .foregroundColor(isSelected ? .echoPaper : .echoInk55)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 8)
                        .background(isSelected ? Color.echoInk : Color.clear)
                        .clipShape(RoundedRectangle(cornerRadius: 9))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(4)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.bottom, Spacing.sm.rawValue)
    }

    @ViewBuilder
    private var chatsContent: some View {
        ChatFolderFilterView(selection: $folder, unreadCounts: folderUnreadCounts)

        if filtered.isEmpty {
            placeholder(icon: "bubble.left.and.bubble.right", text: "No conversations in this filter.")
        } else {
            ForEach(filtered) { conv in
                conversationRow(conv)
                Divider().background(Color.echoHair).padding(.leading, 76)
            }
        }
    }

    /// Pinned conversations, in pin order (Pinned segment — replaces the old strip).
    private var pinnedConversations: [StoredConversation] {
        pinnedStore.orderedIDs.compactMap { id in conversations.first { $0.id == id } }
    }

    @ViewBuilder
    private var pinnedContent: some View {
        if pinnedConversations.isEmpty {
            placeholder(icon: "pin.slash", text: "No pinned conversations.\nLong-press a chat and choose Pin to keep it here.")
        } else {
            HStack {
                Spacer()
                Button("Edit pins") { showEditPins = true }
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.echoSignal)
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.sm.rawValue)

            ForEach(pinnedConversations) { conv in
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

    /// Groups segment: a "New group" action + the groups the user belongs to.
    @ViewBuilder
    private var groupsContent: some View {
        Button {
            showCreateGroup = true
        } label: {
            HStack(spacing: 10) {
                Image(systemName: "plus.circle.fill").font(.system(size: 20))
                Text("New group").font(.system(size: 15, weight: .semibold))
                Spacer()
            }
            .foregroundColor(.echoSignal)
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
        }
        .buttonStyle(.plain)

        Divider().background(Color.echoHair).padding(.leading, Spacing.lg.rawValue)

        // Groups the user belongs to (none wired yet — backend arrives in a later phase).
        placeholder(icon: "person.3", text: "You're not in any groups yet.\nTap “New group” to start one.")
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
            trustTier: { _ in 3 },
            mutedIDs: [],
            onSelectConversation: { _ in },
            onCompose: {},
            onOpenHidden: {},
            onSwitchPersona: { _ in },
            onSelectHiddenPersona: { _ in }
        )
    }
}
#endif
