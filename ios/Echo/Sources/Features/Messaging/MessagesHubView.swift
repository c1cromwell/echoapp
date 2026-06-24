#if os(iOS)
import SwiftUI

/// Messages hub (spec §3.2 + docs/design-previews/messagehub1.png). Layout:
/// persona switcher → "Messages" title with search / hidden-folders / new-message
/// icons → segment bar (Chats · Groups · Channels) → pinned strip → conversation list.
/// Search is icon-toggled (no persistent search bar); hidden folders open via the
/// folder icon (biometric-gated).
struct MessagesHubView: View {
    enum Segment: String, Identifiable {
        case chats, groups, channels, hidden, archived
        var id: String { rawValue }
        var title: String {
            switch self {
            case .chats: return "Chats"
            case .groups: return "Groups"
            case .channels: return "Channels"
            case .hidden: return "Hidden"
            case .archived: return "Archived"
            }
        }
        static let standard: [Segment] = [.chats, .groups, .channels]
    }

    let conversations: [StoredConversation]
    let hiddenConversations: [StoredConversation]
    let personas: [PersonaSummary]
    let activePersona: PersonaSummary
    let trustTier: (String) -> Int
    let mutedIDs: Set<String>
    var hiddenUnlocked: Bool = false
    var isDuressHiddenVault: Bool = false

    let onSelectConversation: (String) -> Void
    let onCompose: () -> Void
    let onComposeHidden: () -> Void
    let onOpenHidden: () -> Void
    let onLockHidden: () -> Void
    let onOpenHiddenSettings: () -> Void = {}
    let onSwitchPersona: (PersonaSummary) -> Void
    let onSelectHiddenPersona: (PersonaSummary) -> Void
    let onToggleArchive: (String, Bool) -> Void
    let onOpenMessageSearch: (String) -> Void

    @Bindable private var pinnedStore = PinnedConversationsStore.shared
    @State private var searchText = ""
    @State private var showSearch = false
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
                Color.clear.frame(height: 12)
                if showSearch { searchField }
                if showSearch && searchText.trimmingCharacters(in: .whitespacesAndNewlines).count >= 2 {
                    Button {
                        onOpenMessageSearch(searchText.trimmingCharacters(in: .whitespacesAndNewlines))
                        withAnimation { showSearch = false }
                        searchText = ""
                        searchFocused = false
                    } label: {
                        HStack(spacing: 8) {
                            Image(systemName: "text.magnifyingglass")
                            Text("Search messages for \"\(searchText)\"")
                            Spacer()
                        }
                        .font(.system(size: 14, weight: .medium))
                        .foregroundColor(.echoSignal)
                        .padding(.horizontal, Spacing.lg.rawValue)
                        .padding(.bottom, 6)
                    }
                    .buttonStyle(.plain)
                }
                segmentBar

                ScrollView {
                    LazyVStack(spacing: 0) {
                        switch segment {
                        case .chats:    chatsContent
                        case .groups:   groupsContent
                        case .channels: placeholder(icon: "dot.radiowaves.left.and.right", text: "Channels are coming soon.")
                        case .hidden:   hiddenContent
                        case .archived: archivedContent
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
                    Text("Name your group and add the people you trust. Group chats are coming soon.")
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
        .onChange(of: hiddenUnlocked) { _, unlocked in
            if unlocked {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) { segment = .hidden }
            } else if segment == .hidden {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) { segment = .chats }
            }
        }
    }

    private var filtered: [StoredConversation] {
        let archived = ConversationArchiveStore.archivedIds()
        let unpinned = conversations.filter { !pinnedStore.isPinned($0.id) && !archived.contains($0.id) }
        return unpinned.filter { conv in
            let matchesSearch = searchText.isEmpty
                || conv.contactName.localizedCaseInsensitiveContains(searchText)
                || conv.lastMessage.localizedCaseInsensitiveContains(searchText)
            return matchesSearch
        }
        .sorted { ($0.unreadCount > 0) && ($1.unreadCount == 0) }
    }

    private var archivedConversations: [StoredConversation] {
        let archived = ConversationArchiveStore.archivedIds()
        return conversations.filter { archived.contains($0.id) }
    }

    private var header: some View {
        VStack(spacing: 0) {
            PersonaSwitcherHeader(active: activePersona) { showPersonaSheet = true }

            HStack(spacing: Spacing.sm.rawValue) {
                Text("Messages")
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(.echoInk)
                    .onLongPressGesture(minimumDuration: 2) {
                        onOpenHidden()
                    }
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

                Button(action: onOpenHidden) {
                    headerIcon("folder", filled: segment == .hidden)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Hidden chats")

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

    private func headerIcon(_ systemName: String, filled: Bool) -> some View {
        Image(systemName: systemName)
            .font(.system(size: 17, weight: .medium))
            .foregroundColor(filled ? .white : .echoInk70)
            .frame(width: 38, height: 38)
            .background(filled ? Color.echoSignal : Color.echoPaperDim)
            .clipShape(Circle())
    }

    private var secureBar: some View {
        Group {
            if PrivacySettingsStore.showsEncryptionIndicator {
                Button {
                    showIntegrityExplainer = true
                } label: {
                    SecureThreadIndicatorBar()
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Secure thread — learn about integrity")
            }
        }
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

    private var segmentBar: some View {
        HStack(spacing: 4) {
            ForEach(Segment.standard) { seg in
                let isSelected = segment == seg
                Button {
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) { segment = seg }
                    onLockHidden()
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
        if !pinnedConversations.isEmpty {
            pinnedStrip
        }

        Spacer().frame(height: 6)

        if filtered.isEmpty {
            placeholder(icon: "bubble.left.and.bubble.right", text: "No conversations yet.")
        } else {
            ForEach(filtered) { conv in
                conversationRow(conv)
                Divider().background(Color.echoHair).padding(.horizontal, 18)
            }
            if !archivedConversations.isEmpty {
                Button {
                    withAnimation { segment = .archived }
                } label: {
                    HStack {
                        Image(systemName: "archivebox")
                        Text("Archived (\(archivedConversations.count))")
                        Spacer()
                    }
                    .font(.system(size: 14, weight: .medium))
                    .foregroundColor(.echoInk55)
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, 12)
                }
                .buttonStyle(.plain)
            }
        }
    }

    private var pinnedStrip: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("PINNED")
                .font(.system(size: 10, weight: .semibold))
                .tracking(1)
                .foregroundColor(.echoInk40)
                .padding(.leading, Spacing.lg.rawValue)
                .padding(.top, 10)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 14) {
                    ForEach(pinnedConversations) { conv in
                        Button { onSelectConversation(conv.id) } label: {
                            VStack(spacing: 4) {
                                ZStack(alignment: .bottomTrailing) {
                                    pinnedAvatar(for: conv)
                                    if conv.unreadCount > 0 {
                                        Circle()
                                            .fill(Color.echoSignal)
                                            .frame(width: 10, height: 10)
                                            .overlay(Circle().stroke(Color.echoPaper, lineWidth: 1.5))
                                    }
                                }
                                Text(conv.contactName.split(separator: " ").first.map(String.init) ?? conv.contactName)
                                    .font(.system(size: 10))
                                    .foregroundColor(.echoInk55)
                                    .lineLimit(1)
                            }
                            .frame(width: 52)
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, Spacing.lg.rawValue)
            }

            Divider().background(Color.echoHair)
                .padding(.horizontal, 18)
                .padding(.top, 6)
        }
    }

    private func pinnedAvatar(for conv: StoredConversation) -> some View {
        let initials = conv.contactName.split(separator: " ").prefix(2).map { String($0.first ?? " ") }.joined()
        return Text(initials)
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.white)
            .frame(width: 40, height: 40)
            .background(NewConversationSheet.avatarColor(for: conv.contactName))
            .clipShape(Circle())
    }

    private var pinnedConversations: [StoredConversation] {
        pinnedStore.orderedIDs.compactMap { id in conversations.first { $0.id == id } }
    }

    private func trustBadgeCircle(_ color: Color) -> some View {
        Image(systemName: "checkmark")
            .font(.system(size: 7, weight: .bold))
            .foregroundColor(.white)
            .frame(width: 14, height: 14)
            .background(color)
            .clipShape(Circle())
    }

    @ViewBuilder
    private func trustOverlay(tier: Int) -> some View {
        if tier >= 3 {
            HStack(spacing: 2) {
                trustBadgeCircle(.echoTrustGreen)
                trustBadgeCircle(.echoSignal)
            }
        } else if tier >= 2 {
            trustBadgeCircle(.echoTrustGreen)
        } else {
            Text("unverified")
                .font(.system(size: 8, weight: .semibold))
                .foregroundColor(.echoAlert)
        }
    }

    @ViewBuilder
    private func conversationRow(_ conv: StoredConversation) -> some View {
        let tier = trustTier(conv.id)
        let isArchived = ConversationArchiveStore.isArchived(conversationId: conv.id)
        let initials = conv.contactName.split(separator: " ").prefix(2).map { String($0.first ?? " ") }.joined()
        Button { onSelectConversation(conv.id) } label: {
            HStack(spacing: 12) {
                ZStack(alignment: .bottomTrailing) {
                    Text(initials)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.white)
                        .frame(width: 48, height: 48)
                        .background(NewConversationSheet.avatarColor(for: conv.contactName))
                        .clipShape(Circle())
                    if conv.isOnline {
                        Circle()
                            .fill(Color.echoTrustGreen)
                            .frame(width: 12, height: 12)
                            .overlay(Circle().stroke(Color.echoPaper, lineWidth: 2))
                            .offset(x: 1, y: 1)
                    }
                }

                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Text(conv.contactName)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundColor(.echoInk)
                            .lineLimit(1)
                        trustOverlay(tier: tier)
                    }
                    Text(conv.lastMessage)
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk55)
                        .lineLimit(1)
                }

                Spacer(minLength: 0)

                VStack(alignment: .trailing, spacing: 6) {
                    Text(conv.timestamp)
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk40)
                    if conv.unreadCount > 0 {
                        Text("\(min(conv.unreadCount, 99))")
                            .font(.system(size: 12, weight: .bold))
                            .foregroundColor(.white)
                            .padding(.horizontal, 6)
                            .frame(minWidth: 20, minHeight: 20)
                            .background(Color.echoSignal)
                            .clipShape(Capsule())
                    } else if mutedIDs.contains(conv.id) {
                        Image(systemName: "bell.slash.fill")
                            .font(.system(size: 13))
                            .foregroundColor(.echoInk40)
                    }
                }
            }
            .padding(.horizontal, 18)
            .padding(.vertical, 12)
        }
        .buttonStyle(.plain)
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
            if segment == .archived || isArchived {
                Button {
                    onToggleArchive(conv.id, false)
                } label: {
                    Label("Unarchive", systemImage: "tray.and.arrow.up")
                }
            } else {
                Button {
                    onToggleArchive(conv.id, true)
                } label: {
                    Label("Archive", systemImage: "archivebox")
                }
            }
        }
    }

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

        placeholder(icon: "person.3", text: "You\u{2019}re not in any groups yet.\nTap \u{201C}New group\u{201D} to start one.")
    }

    @ViewBuilder
    private var hiddenContent: some View {
        if hiddenUnlocked {
            hiddenUnlockedHeader
        }

        Button(action: onComposeHidden) {
            HStack(spacing: 10) {
                Image(systemName: "plus.circle.fill").font(.system(size: 20))
                Text("New hidden chat").font(.system(size: 15, weight: .semibold))
                Spacer()
            }
            .foregroundColor(.echoSignal)
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
        }
        .buttonStyle(.plain)

        Divider().background(Color.echoHair).padding(.leading, Spacing.lg.rawValue)

        if hiddenConversations.isEmpty {
            let emptyText = isDuressHiddenVault
                ? "No hidden chats."
                : "No hidden chats yet.\nUse chat settings to hide a conversation, or tap \"New hidden chat\" above."
            placeholder(icon: "eye.slash", text: emptyText)
        } else {
            ForEach(hiddenConversations) { conv in
                conversationRow(conv)
                Divider().background(Color.echoHair).padding(.horizontal, 18)
            }
        }
    }

    @ViewBuilder
    private var archivedContent: some View {
        if archivedConversations.isEmpty {
            placeholder(icon: "archivebox", text: "No archived conversations.")
        } else {
            ForEach(archivedConversations) { conv in
                conversationRow(conv)
                Divider().background(Color.echoHair).padding(.horizontal, 18)
            }
        }
    }

    private var hiddenUnlockedHeader: some View {
        HStack {
            Text("Vault unlocked")
                .font(.echomono(11))
                .foregroundColor(.echoInk55)
            Spacer()
            Button(action: onOpenHiddenSettings) {
                Image(systemName: "gearshape")
                    .font(.system(size: 15, weight: .medium))
                    .foregroundColor(.echoInk70)
            }
            .buttonStyle(.plain)
            .accessibilityLabel("Hidden chat settings")
            Button(action: onLockHidden) {
                Label("Lock", systemImage: "lock.fill")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.echoSignal)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, Spacing.sm.rawValue)
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
                StoredConversation(contactName: "Aria Rao", peerDID: "did:key:a", lastMessage: "Signed receipt \u{2713}", timestamp: "9:32", unreadCount: 3, isOnline: true),
                StoredConversation(contactName: "Kai Mercer", peerDID: "did:key:k", lastMessage: "Typing\u{2026}", timestamp: "8:58"),
            ],
            hiddenConversations: [],
            personas: [PersonaSummary(id: "default", name: "Aria (public)", initials: "AR")],
            activePersona: PersonaSummary(id: "default", name: "Aria (public)", initials: "AR"),
            trustTier: { _ in 3 },
            mutedIDs: [],
            onSelectConversation: { _ in },
            onCompose: {},
            onComposeHidden: {},
            onOpenHidden: {},
            onLockHidden: {},
            onSwitchPersona: { _ in },
            onSelectHiddenPersona: { _ in },
            onToggleArchive: { _, _ in },
            onOpenMessageSearch: { _ in }
        )
    }
}
#endif
#endif
