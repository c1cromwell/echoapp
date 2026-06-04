#if os(iOS)
import SwiftUI

struct NewConversationSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable private var conversationStore = ConversationStore.shared

    let onConversationCreated: (StoredConversation) -> Void
    var onHiddenConversationCreated: ((StoredConversation) -> Void)? = nil
    var onNewGroup: (() -> Void)? = nil

    @State private var searchText = ""
    @State private var mode: Mode = .contacts
    @State private var usernameResults: [UsernameSearchHit] = []
    @State private var isSearching = false
    @State private var errorMessage: String?
    @State private var searchTask: Task<Void, Never>?
    @State private var hiddenMode = false
    @FocusState private var searchFocused: Bool

    private enum Mode: Equatable {
        case contacts, username, phone
    }

    private var allContacts: [StoredConversation] {
        conversationStore.conversations.filter {
            !ConversationPreferencesStore.shared.isHidden($0.id)
        }
    }

    private var filteredContacts: [StoredConversation] {
        let base = allContacts
        if searchText.isEmpty { return base }
        return base.filter {
            $0.contactName.localizedCaseInsensitiveContains(searchText)
        }
    }

    private var groupedContacts: [(String, [StoredConversation])] {
        let sorted = filteredContacts.sorted {
            $0.contactName.localizedCompare($1.contactName) == .orderedAscending
        }
        let grouped = Dictionary(grouping: sorted) { conv in
            let first = conv.contactName.prefix(1).uppercased()
            return first.rangeOfCharacter(from: .letters) != nil ? first : "#"
        }
        return grouped.sorted { $0.key < $1.key }
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                if hiddenMode {
                    hiddenModeHeader
                }

                switch mode {
                case .contacts:
                    actionRows
                    contactsList
                case .username:
                    usernameSearchContent
                case .phone:
                    phoneSearchContent
                }
            }
            .background(Color.echoPaper)
            .safeAreaInset(edge: .bottom) {
                searchBar
            }
            .navigationTitle(hiddenMode ? "New Hidden Chat" : "New Message")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    if mode != .contacts || hiddenMode {
                        Button {
                            if hiddenMode && mode == .contacts {
                                hiddenMode = false
                            } else {
                                withAnimation { mode = .contacts }
                                searchText = ""
                                usernameResults = []
                                errorMessage = nil
                            }
                        } label: {
                            Image(systemName: "chevron.left")
                                .font(.system(size: 15, weight: .semibold))
                                .foregroundColor(.echoInk)
                        }
                    }
                }
                ToolbarItem(placement: .primaryAction) {
                    Button { dismiss() } label: {
                        Image(systemName: "xmark")
                            .font(.system(size: 13, weight: .bold))
                            .foregroundColor(.echoInk55)
                            .frame(width: 30, height: 30)
                            .background(Color.echoPaperDim)
                            .clipShape(Circle())
                    }
                }
            }
        }
    }

    // MARK: - Hidden Mode Banner

    private var hiddenModeHeader: some View {
        HStack(spacing: 8) {
            Image(systemName: "eye.slash.fill")
                .font(.system(size: 13))
                .foregroundColor(.echoSignal)
            Text("This chat will be hidden and require biometric unlock")
                .font(.system(size: 13))
                .foregroundColor(.echoInk70)
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, 8)
        .frame(maxWidth: .infinity)
        .background(Color.echoSignal.opacity(0.08))
    }

    // MARK: - Action Rows

    private var actionRows: some View {
        VStack(spacing: 0) {
            actionRow(icon: "person.2", title: "New Group") {
                dismiss()
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                    onNewGroup?()
                }
            }
            if onHiddenConversationCreated != nil {
                actionRow(icon: "eye.slash", title: "New Hidden Chat") {
                    withAnimation { hiddenMode = true }
                }
            }
            actionRow(icon: "at", title: "Find by Username") {
                withAnimation { mode = .username }
                searchFocused = true
            }
            actionRow(icon: "number", title: "Find by Phone Number") {
                withAnimation { mode = .phone }
                searchFocused = true
            }
        }
        .padding(.top, 4)
    }

    private func actionRow(icon: String, title: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack(spacing: 14) {
                Image(systemName: icon)
                    .font(.system(size: 17, weight: .medium))
                    .foregroundColor(.echoInk70)
                    .frame(width: 36, height: 36)
                    .background(Color.echoPaperDim)
                    .clipShape(Circle())
                Text(title)
                    .font(.system(size: 16, weight: .medium))
                    .foregroundColor(.echoInk)
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.echoInk40)
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 12)
        }
        .buttonStyle(.plain)
    }

    // MARK: - Contacts List

    private var contactsList: some View {
        ScrollView {
            LazyVStack(spacing: 0, pinnedViews: .sectionHeaders) {
                if groupedContacts.isEmpty && !searchText.isEmpty {
                    emptyContactsState
                } else {
                    ForEach(groupedContacts, id: \.0) { letter, contacts in
                        Section {
                            ForEach(contacts) { conv in
                                contactRow(conv)
                            }
                        } header: {
                            sectionHeader(letter)
                        }
                    }
                }
            }
        }
        .overlay(alignment: .trailing) {
            alphabetScrubber
        }
    }

    private var emptyContactsState: some View {
        VStack(spacing: 12) {
            Image(systemName: "person.crop.circle.badge.questionmark")
                .font(.system(size: 40))
                .foregroundColor(.echoInk40)
            Text("No contacts match \"\(searchText)\"")
                .font(.system(size: 14))
                .foregroundColor(.echoInk55)
        }
        .frame(maxWidth: .infinity)
        .padding(.top, 60)
    }

    private func contactRow(_ conv: StoredConversation) -> some View {
        Button {
            if hiddenMode, let onHidden = onHiddenConversationCreated {
                onHidden(conv)
            } else {
                onConversationCreated(conv)
            }
            dismiss()
        } label: {
            HStack(spacing: 12) {
                contactAvatar(conv.contactName)
                Text(conv.contactName)
                    .font(.system(size: 16, weight: .medium))
                    .foregroundColor(.echoInk)
                trustSeal(for: conv)
                Spacer()
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 10)
        }
        .buttonStyle(.plain)
    }

    private func contactAvatar(_ name: String) -> some View {
        let initials = name.split(separator: " ").prefix(2)
            .map { String($0.prefix(1)).uppercased() }.joined()
        return Text(initials.isEmpty ? "?" : initials)
            .font(.system(size: 14, weight: .semibold))
            .foregroundColor(.white)
            .frame(width: 40, height: 40)
            .background(Self.avatarColor(for: name))
            .clipShape(Circle())
    }

    @ViewBuilder
    private func trustSeal(for conv: StoredConversation) -> some View {
        let tier = ContactTrustIndex.shared.tier(
            conversationId: conv.id, peerDID: conv.peerDID
        )
        if tier >= 2 {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: 14))
                .foregroundColor(.echoTrustGreen)
        }
    }

    private func sectionHeader(_ letter: String) -> some View {
        HStack {
            Text(letter)
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.echoInk55)
                .padding(.leading, Spacing.lg.rawValue)
            Spacer()
        }
        .padding(.vertical, 6)
        .background(Color.echoPaper)
    }

    private var alphabetScrubber: some View {
        VStack(spacing: 1) {
            ForEach(Array("ABCDEFGHIJKLMNOPQRSTUVWXYZ#"), id: \.self) { char in
                Text(String(char))
                    .font(.system(size: 10, weight: .medium))
                    .foregroundColor(.echoInk40)
            }
        }
        .padding(.trailing, 4)
    }

    // MARK: - Username Search

    private var usernameSearchContent: some View {
        VStack(spacing: 0) {
            if let errorMessage {
                Text(errorMessage)
                    .font(.system(size: 13))
                    .foregroundColor(.echoAlert)
                    .padding()
            }

            if !usernameResults.isEmpty {
                ScrollView {
                    LazyVStack(spacing: 0) {
                        ForEach(usernameResults) { hit in
                            usernameRow(hit)
                        }
                    }
                }
            } else {
                Spacer()
                if isSearching {
                    ProgressView()
                } else if !searchText.isEmpty {
                    VStack(spacing: 12) {
                        Image(systemName: "person.crop.circle.badge.questionmark")
                            .font(.system(size: 40))
                            .foregroundColor(.echoInk40)
                        Text("No users found")
                            .font(.system(size: 14))
                            .foregroundColor(.echoInk55)
                    }
                } else {
                    VStack(spacing: 12) {
                        Image(systemName: "at.circle")
                            .font(.system(size: 40))
                            .foregroundColor(.echoInk40)
                        Text("Type a username to search")
                            .font(.system(size: 14))
                            .foregroundColor(.echoInk55)
                    }
                }
                Spacer()
            }
        }
    }

    private func usernameRow(_ hit: UsernameSearchHit) -> some View {
        Button {
            startConversation(with: hit)
        } label: {
            HStack(spacing: 12) {
                contactAvatar("@\(hit.username)")
                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 4) {
                        Text("@\(hit.username)")
                            .font(.system(size: 16, weight: .medium))
                            .foregroundColor(.echoInk)
                        if let tier = hit.tier, tier >= 2 {
                            Image(systemName: "checkmark.seal.fill")
                                .font(.system(size: 14))
                                .foregroundColor(.echoTrustGreen)
                        }
                    }
                    Text(String(hit.did.prefix(28)) + "…")
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk40)
                }
                Spacer()
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, 10)
        }
        .buttonStyle(.plain)
    }

    // MARK: - Phone Search

    private var phoneSearchContent: some View {
        VStack(spacing: 16) {
            Spacer()
            Image(systemName: "phone.circle")
                .font(.system(size: 40))
                .foregroundColor(.echoInk40)
            Text("Enter a phone number to find contacts")
                .font(.system(size: 14))
                .foregroundColor(.echoInk55)
                .multilineTextAlignment(.center)
            Text("Phone search arrives in a future update.")
                .font(.system(size: 13))
                .foregroundColor(.echoInk40)
                .multilineTextAlignment(.center)
            Spacer()
        }
        .padding(.horizontal, Spacing.xl.rawValue)
    }

    // MARK: - Search Bar

    private var searchBar: some View {
        HStack(spacing: 10) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 14))
                .foregroundColor(.echoInk40)
            TextField(
                mode == .phone ? "Phone number" :
                    mode == .username ? "Username" :
                    "Name, username, or number",
                text: $searchText
            )
            .font(.system(size: 15))
            .textInputAutocapitalization(.never)
            .autocorrectionDisabled()
            .focused($searchFocused)
            .submitLabel(.search)
            .onChange(of: searchText) { _, newValue in
                if mode == .username {
                    scheduleUsernameSearch(newValue)
                }
            }
            if !searchText.isEmpty {
                Button { searchText = "" } label: {
                    Image(systemName: "xmark.circle.fill")
                        .foregroundColor(.echoInk40)
                }
                .buttonStyle(.plain)
            }
        }
        .padding(12)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, 8)
        .background(Color.echoPaper)
    }

    // MARK: - Search Logic

    private func scheduleUsernameSearch(_ query: String) {
        searchTask?.cancel()
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else {
            usernameResults = []
            return
        }
        searchTask = Task {
            try? await Task.sleep(nanoseconds: 350_000_000)
            guard !Task.isCancelled else { return }
            await runUsernameSearch(trimmed)
        }
    }

    private func runUsernameSearch(_ query: String) async {
        isSearching = true
        errorMessage = nil
        defer { isSearching = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = "API client unavailable"
            return
        }
        let social = ContactSocialAPIClient(apiClient: client)
        do {
            usernameResults = try await social.searchUsername(query)
        } catch {
            errorMessage = error.localizedDescription
            usernameResults = []
        }
    }

    private func startConversation(with hit: UsernameSearchHit) {
        Task { @MainActor in
            if let client = DIContainer.shared.resolveAPIClient() {
                let social = ContactSocialAPIClient(apiClient: client)
                _ = try? await social.addContact(did: hit.did, addedVia: "new_conversation")
            }
            guard let conversation = await ContactThreadHelper.upsertDirectThread(
                peerDID: hit.did,
                displayName: "@\(hit.username)"
            ) else { return }
            if !UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage") {
                UserDefaults.standard.set(true, forKey: "echo.hasSentFirstMessage")
            }
            if hiddenMode, let onHidden = onHiddenConversationCreated {
                onHidden(conversation)
            } else {
                onConversationCreated(conversation)
            }
            dismiss()
        }
    }

    static func avatarColor(for name: String) -> Color {
        let colors: [Color] = [
            .echoSignal, .echoTrustGreen, .orange,
            .purple, .pink, .cyan, .indigo
        ]
        var hash: UInt64 = 5381
        for char in name.unicodeScalars {
            hash = ((hash &<< 5) &+ hash) &+ UInt64(char.value)
        }
        return colors[Int(hash % UInt64(colors.count))]
    }
}
#endif
