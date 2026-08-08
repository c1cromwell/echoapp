#if os(iOS)
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
    @State private var reactorDetail: (emoji: String, reactors: [String])?
    @State private var groupsInCommonText = "No groups in common"
    @State private var pinnedMessageId: String?
    @State private var showForwardSheet = false
    @State private var forwardPreview = ""
    @State private var forwardSourceMessageId: String?
    @State private var forwardSourceConversationId: String?
    @State private var showAttachmentPicker = false
    @State private var showCreatePoll = false
    @State private var smartReplies: [SmartReplySuggestion] = []
    @State private var threadSummary: ThreadSummary?
    @State private var messageTranslations: [String: String] = [:]
    @State private var safetyAssessment: SafetyAssessment?
    @State private var safetyDismissed = false
    @State private var safetyEvaluated = false
    @State private var showPaymentSheet = false
    @State private var pendingPayRequest: PaymentPayload?
    @StateObject private var voiceRecorder = VoiceNoteRecorder()
    @State private var showVoiceCall = false
    @State private var showVideoCall = false
    @State private var showScheduleMessage = false
    @State private var scheduleSourceText = ""
    @State private var showCloudPicker = false
    @State private var showSafetyNumber = false
    @State private var showKeyChangeBanner = false
    @State private var isMessageRequestPending = false
    @State private var sendMediaViewOnce = false
    @State private var showGifPicker = false
    @State private var isSelecting = false
    @State private var selectedMessageIds: Set<String> = []
    @State private var showMultiForward = false

    let contactName: String
    let conversationId: String
    let peerDID: String
    let currentUserDID: String
    let onSendMessage: (String) -> Void

    private var canSend: Bool {
        !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

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

    private var contactInitials: String {
        let cleaned = contactName.hasPrefix("@") ? String(contactName.dropFirst()) : contactName
        let parts = cleaned.split(separator: " ").prefix(2)
        return parts.map { String($0.prefix(1)).uppercased() }.joined()
    }

    private var trustTier: Int {
        ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: peerDID)
    }

    private var trustLevelLabel: String {
        ContactTrustIndex.trustLevelLabel(tier: trustTier)
    }

    var body: some View {
        ZStack {
            ChatWallpaperStore.style(for: conversationId).background.ignoresSafeArea()

            VStack(spacing: 0) {
                chatNavigationBar

                SecureThreadIndicator()

                if let threadSummary {
                    ThreadSummaryBanner(summary: threadSummary)
                }

                if viewModel.peerScreenshotNotice {
                    HStack(spacing: 8) {
                        Image(systemName: "camera.viewfinder")
                        Text("\(contactName) may have taken a screenshot")
                            .font(.system(size: 13, weight: .medium))
                        Spacer()
                    }
                    .foregroundColor(.echoAlert)
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, 8)
                    .background(Color.echoAlert.opacity(0.08))
                }

                #if os(iOS)
                if showKeyChangeBanner {
                    HStack(spacing: 8) {
                        Image(systemName: "exclamationmark.shield")
                        Text("Safety number changed — verify \(contactName)")
                            .font(.system(size: 13, weight: .medium))
                        Spacer()
                        Button("Verify") { showSafetyNumber = true }
                            .font(.system(size: 13, weight: .semibold))
                        Button {
                            showKeyChangeBanner = false
                        } label: {
                            Image(systemName: "xmark")
                        }
                    }
                    .foregroundColor(.echoAlert)
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, 8)
                    .background(Color.echoAlert.opacity(0.08))
                }
                #endif

                if isMessageRequestPending {
                    MessageRequestBanner(
                        peerDID: peerDID,
                        displayName: contactName,
                        onAccept: {
                            MessageRequestStore.accept(peerDID: peerDID)
                            isMessageRequestPending = false
                        },
                        onDecline: {
                            MessageRequestStore.decline(peerDID: peerDID)
                            isMessageRequestPending = false
                        }
                    )
                    .padding(.vertical, 6)
                }

                ScrollViewReader { proxy in
                    if let pinned = pinnedMessage(for: pinnedMessageId) {
                        ChatPinnedMessageBanner(
                            authorLabel: pinned.isFromCurrentUser ? "Pinned · You" : "Pinned · \(contactName)",
                            preview: pinned.content,
                            onUnpin: {
                                #if os(iOS)
                                ConversationPinnedMessageStore.setPinnedMessageId(nil, conversationId: conversationId)
                                #endif
                                pinnedMessageId = nil
                            },
                            onTap: {
                                withAnimation { proxy.scrollTo(pinned.id, anchor: .center) }
                            }
                        )
                    }

                    if let assessment = safetyAssessment, !safetyDismissed {
                        ContactSafetyBanner(
                            assessment: assessment,
                            onVerify: { showChatSettings = true },
                            onBlock: {
                                safetyDismissed = true
                                Task { @MainActor in
                                    try? await DIContainer.shared.resolveBlockContactUseCase()?.block(did: peerDID)
                                }
                            },
                            onDismiss: { safetyDismissed = true }
                        )
                        .padding(.horizontal, 12)
                        .padding(.bottom, 6)
                    }

                    List {
                        if viewModel.messages.isEmpty {
                            contactProfileCard
                                .listRowSeparator(.hidden)
                                .listRowInsets(.init())
                                .listRowBackground(Color.clear)
                        }
                        ForEach(viewModel.messages) { message in
                            HStack(spacing: 8) {
                                if isSelecting {
                                    Image(systemName: selectedMessageIds.contains(message.id) ? "checkmark.circle.fill" : "circle")
                                        .font(.system(size: 20))
                                        .foregroundColor(selectedMessageIds.contains(message.id) ? .echoSignal : .echoInk40)
                                        .padding(.leading, 12)
                                }
                                messageRow(for: message)
                                    .allowsHitTesting(!isSelecting)
                            }
                            .listRowSeparator(.hidden)
                            .listRowInsets(.init())
                            .listRowBackground(Color.clear)
                            .contentShape(Rectangle())
                            .gesture(
                                TapGesture().onEnded { toggleSelect(message.id) },
                                including: isSelecting ? .gesture : .subviews
                            )
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
                    .task { evaluateSafety() }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            VStack(spacing: 0) {
                if let label = viewModel.typingStatusText {
                    TypingIndicatorView(label: label)
                        .transition(.opacity.combined(with: .move(edge: .bottom)))
                        .padding(.vertical, 4)
                        .padding(.horizontal, Spacing.md.rawValue)
                }
                if let reply = viewModel.replyingTo {
                    ChatComposerBanner(
                        mode: .reply(
                            author: reply.isFromCurrentUser ? "You" : contactName,
                            preview: MessageComposerLogic.replyPreview(
                                authorName: reply.isFromCurrentUser ? "You" : contactName,
                                content: reply.content
                            )
                        ),
                        onCancel: { viewModel.cancelComposerMode() }
                    )
                } else if viewModel.editingMessageId != nil {
                    ChatComposerBanner(mode: .edit, onCancel: {
                        viewModel.cancelComposerMode()
                        messageText = ""
                    })
                }
                SmartReplyBar(suggestions: smartReplies) { suggestion in
                    messageText = suggestion
                }
                chatComposerBar
            }
            .background(Color.echoPaperDim)
        }
        .navigationBarBackButtonHidden(true)
        .sheet(isPresented: $showChatSettings) {
            ChatSettingsSheet(
                contactName: contactName,
                conversationId: conversationId,
                preferences: ConversationPreferencesStore.shared.preferences(for: conversationId),
                isArchived: ConversationArchiveStore.isArchived(conversationId: conversationId),
                onChange: { [viewModel] newPrefs in
                    let oldPrefs = ConversationPreferencesStore.shared.preferences(for: conversationId)
                    ConversationPreferencesStore.shared.save(newPrefs, for: conversationId)
                    if newPrefs.isHidden != oldPrefs.isHidden {
                        ConversationThreadStore.migrateStorageEncryption(conversationId: conversationId)
                    }
                    if newPrefs.disappearing != oldPrefs.disappearing {
                        Task { await viewModel.setDisappearing(ttlSeconds: newPrefs.disappearing.seconds) }
                    }
                },
                onArchiveChange: { archived in
                    Task {
                        let client = DIContainer.shared.resolveAPIClient().map {
                            LiveConversationArchiveAPIClient(apiClient: $0)
                        }
                        if let client {
                            try? await client.setArchived(archived, conversationId: conversationId)
                        } else {
                            ConversationArchiveStore.setArchived(archived, conversationId: conversationId)
                        }
                    }
                },
                onVerify: {
                    showChatSettings = false
                    showSafetyNumber = true
                }
            )
            .presentationDetents([.medium, .large])
        }
        #if os(iOS)
        .sheet(isPresented: $showSafetyNumber) {
            SafetyNumberCompareView(
                contactName: contactName,
                localDID: currentUserDID,
                peerDID: peerDID,
                onDismiss: { showSafetyNumber = false }
            )
        }
        #endif
        .sheet(isPresented: $showCreatePoll) {
            CreatePollSheet { question, options in
                Task { await viewModel.createPoll(question: question, optionTexts: options) }
            }
            .presentationDetents([.medium, .large])
        }
        .sheet(isPresented: $showGifPicker) {
            GifPickerSheet { result in
                showGifPicker = false
                Task { await sendGif(result) }
            }
            .presentationDetents([.medium, .large])
        }
        .sheet(isPresented: $showPaymentSheet) {
            PaymentRequestSheet(
                onSubmit: { kind, amount, memo in
                    showPaymentSheet = false
                    handlePaymentCompose(kind: kind, amount: amount, memo: memo)
                },
                onCancel: { showPaymentSheet = false }
            )
            .presentationDetents([.medium])
        }
        .alert("Pay an unverified contact?", isPresented: Binding(
            get: { pendingPayRequest != nil },
            set: { if !$0 { pendingPayRequest = nil } }
        )) {
            Button("Cancel", role: .cancel) { pendingPayRequest = nil }
            Button("Pay anyway") {
                if let request = pendingPayRequest { pendingPayRequest = nil; Task { await finalizePay(request) } }
            }
        } message: {
            Text("This person hasn't verified their identity. Only send money to people you trust.")
        }
        .sheet(item: Binding(
            get: {
                reactorDetail.map { ReactorDetailItem(emoji: $0.emoji, reactors: $0.reactors) }
            },
            set: { reactorDetail = $0.map { ($0.emoji, $0.reactors) } }
        )) { item in
            ReactionReactorDetailSheet(
                emoji: item.emoji,
                reactors: item.reactors,
                currentUserDID: currentUserDID
            )
            .presentationDetents([.medium])
        }
        .sheet(item: $actionsTargetMessage) { message in
            MessageActionsSheet(
                messagePreview: message.content,
                isOwnMessage: message.isFromCurrentUser,
                sentWithinEditWindow: MessageComposerLogic.canEdit(
                    sentAt: message.sentAt,
                    isOwnMessage: message.isFromCurrentUser
                ),
                showTranslate: PrivacyAIConsentStore.load().translationEnabled,
                onAction: { action in
                    handleMessageAction(action, on: message)
                    actionsTargetMessage = nil
                }
            )
            .presentationDetents([.height(300)])
        }
        .task {
            let reactions: ReactionsAPI? = DIContainer.shared.resolveReactionsAPI()
            let receipts: MessageReceiptsAPI? = DIContainer.shared.resolveReceiptsAPI()
            let ops: MessageOpsAPI? = DIContainer.shared.resolveMessageOpsAPI()
            let privacy = MessagingPrivacyPreferences.merged(
                global: PrivacySettingsStore.load(),
                persona: PersonaPrivacySettingsStore.load()
            )
            ActiveChatRegistry.openConversationId = conversationId
            ConversationStore.shared.clearUnread(conversationId: conversationId)
            #if os(iOS)
            if let signal: ConversationSignalService = DIContainer.shared.resolveConversationSignalService() {
                ScreenshotAlertService.shared.configure(signalService: signal)
            }
            ScreenshotAlertService.shared.setActiveChat(conversationId: conversationId, peerDID: peerDID)
            #endif
            viewModel.configure(
                conversationId: conversationId,
                peerDID: peerDID,
                currentUserDID: currentUserDID,
                peerDisplayName: contactName,
                privacy: privacy,
                reactionsAPI: reactions,
                receiptsAPI: receipts,
                opsAPI: ops,
                onSend: onSendMessage
            )
            isMessageRequestPending = MessageRequestStore.isPending(peerDID: peerDID)
            if let token = try? await KeychainManager.shared.getAuthToken() {
                await viewModel.connect(accessToken: token)
            }
            await viewModel.reconcileReactionsOnOpen()
            await viewModel.reconcileReceiptsOnOpen()
            let peerVisible = viewModel.messages
                .filter { !$0.isFromCurrentUser }
                .map(\.id)
            await viewModel.onMessagesVisible(peerVisible)
            await loadGroupsInCommonSummary()
            #if os(iOS)
            pinnedMessageId = ConversationPinnedMessageStore.pinnedMessageId(conversationId: conversationId)
            if messageText.isEmpty {
                let draft = ComposerDraftStore.load(conversationId: conversationId)
                if !draft.isEmpty { messageText = draft }
            }
            #endif
            await refreshSmartReplies()
            await refreshThreadSummary()
            #if os(iOS)
            if SafetyNumber.didPeerKeyChange(
                conversationId: conversationId,
                localDID: currentUserDID,
                peerDID: peerDID
            ) {
                showKeyChangeBanner = true
            }
            ScreenshotAlertService.shared.startMonitoring()
            #endif
        }
        .onChange(of: viewModel.messages.count) { _, _ in
            Task {
                await refreshSmartReplies()
                await refreshThreadSummary()
            }
        }
        .sheet(isPresented: $showForwardSheet) {
            ForwardMessageSheet(
                messagePreview: forwardPreview,
                sourceMessageId: forwardSourceMessageId,
                sourceConversationId: forwardSourceConversationId,
                excludingConversationId: conversationId,
                onForwarded: {
                    forwardPreview = ""
                    forwardSourceMessageId = nil
                    forwardSourceConversationId = nil
                }
            )
        }
        .sheet(isPresented: $showMultiForward) {
            ForwardMessageSheet(
                items: selectedForwardItems,
                sourceConversationId: conversationId,
                excludingConversationId: conversationId,
                onForwarded: { exitSelection() }
            )
        }
        .sheet(isPresented: $showScheduleMessage) {
            ScheduleMessageSheet(initialText: scheduleSourceText) { fireAt in
                _ = ScheduledMessageStore.schedule(
                    conversationId: conversationId,
                    peerDID: peerDID,
                    plaintext: scheduleSourceText,
                    fireAt: fireAt
                )
            }
        }
        .sheet(isPresented: $showCloudPicker) {
            CloudStoragePickerSheet { data, mime in
                Task { await viewModel.sendMedia(data: data, mimeType: mime, mediaKind: .file) }
            }
        }
        .fullScreenCover(isPresented: $showVoiceCall) {
            CallView(peerDID: peerDID, callType: .voice, contactName: contactName)
        }
        .fullScreenCover(isPresented: $showVideoCall) {
            CallView(peerDID: peerDID, callType: .video, contactName: contactName)
        }
        .onDisappear {
            if ActiveChatRegistry.openConversationId == conversationId {
                ActiveChatRegistry.openConversationId = nil
            }
            #if os(iOS)
            ComposerDraftStore.save(conversationId: conversationId, text: messageText)
            ScreenshotAlertService.shared.setActiveChat(conversationId: nil, peerDID: nil)
            #endif
            Task { await viewModel.disconnect() }
        }
        .onTapGesture {
            if reactionTargetId != nil {
                withAnimation(.glacialPress) { reactionTargetId = nil }
            }
        }
        #if os(iOS)
        .screenCaptureGuard(
            enabled: ConversationPreferencesStore.shared.isHidden(conversationId)
        )
        #endif
    }

    private var chatNavigationBar: some View {
        HStack(spacing: 10) {
            Button { dismiss() } label: {
                Image(systemName: "chevron.left")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundColor(.echoSignal)
            }
            .buttonStyle(.plain)

            Text(contactInitials)
                .font(.system(size: 12, weight: .semibold))
                .foregroundColor(.white)
                .frame(width: 32, height: 32)
                .background(NewConversationSheet.avatarColor(for: contactName))
                .clipShape(Circle())

            HStack(spacing: 4) {
                Text(contactName)
                    .font(.system(size: 17, weight: .semibold))
                    .foregroundColor(.echoInk)
                    .lineLimit(1)
                if trustTier >= 1 {
                    TrustBadge(level: trustLevelLabel, size: .small)
                }
            }

            Spacer()

            if isSelecting {
                Button(selectedMessageIds.isEmpty ? "Forward" : "Forward (\(selectedMessageIds.count))") {
                    showMultiForward = true
                }
                .font(.system(size: 15, weight: .semibold))
                .foregroundColor(.echoSignal)
                .disabled(selectedMessageIds.isEmpty)

                Button("Cancel") { exitSelection() }
                    .font(.system(size: 15))
                    .foregroundColor(.echoInk70)
            } else {
                Button { isSelecting = true } label: {
                    Image(systemName: "checkmark.circle")
                        .font(.system(size: 18))
                        .foregroundColor(.echoInk70)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Select messages")

                Button { showVideoCall = true } label: {
                    Image(systemName: "video")
                        .font(.system(size: 18))
                        .foregroundColor(.echoInk70)
                }
                .buttonStyle(.plain)

                Button { showVoiceCall = true } label: {
                    Image(systemName: "phone")
                        .font(.system(size: 18))
                        .foregroundColor(.echoInk70)
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.horizontal, Spacing.md.rawValue)
        .padding(.vertical, 10)
    }

    private func toggleSelect(_ id: String) {
        if selectedMessageIds.contains(id) { selectedMessageIds.remove(id) }
        else { selectedMessageIds.insert(id) }
    }

    private func exitSelection() {
        isSelecting = false
        selectedMessageIds = []
    }

    /// Selected messages as forward items, in conversation order.
    private var selectedForwardItems: [ForwardMessageSheet.Item] {
        viewModel.messages
            .filter { selectedMessageIds.contains($0.id) }
            .map { ForwardMessageSheet.Item(id: $0.id, preview: $0.content) }
    }

    private var contactProfileCard: some View {
        VStack(spacing: 14) {
            Text(contactInitials)
                .font(.system(size: 28, weight: .semibold))
                .foregroundColor(.white)
                .frame(width: 80, height: 80)
                .background(NewConversationSheet.avatarColor(for: contactName))
                .clipShape(Circle())

            Button { showChatSettings = true } label: {
                HStack(spacing: 4) {
                    Text(contactName)
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.echoInk)
                    if trustTier >= 2 {
                        Image(systemName: "checkmark.seal.fill")
                            .font(.system(size: 15))
                            .foregroundColor(.echoTrustGreen)
                    }
                    Image(systemName: "chevron.right")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.echoInk40)
                }
            }
            .buttonStyle(.plain)

            if contactName.hasPrefix("@") {
                HStack(spacing: 6) {
                    Image(systemName: "at")
                        .font(.system(size: 13))
                    Text(String(contactName.dropFirst()))
                        .font(.system(size: 14))
                }
                .foregroundColor(.echoInk55)
            }

            HStack(spacing: 6) {
                Image(systemName: "person.2")
                    .font(.system(size: 13))
                Text(groupsInCommonText)
                    .font(.system(size: 14))
            }
            .foregroundColor(.echoInk55)
        }
        .padding(.vertical, 24)
        .padding(.horizontal, Spacing.lg.rawValue)
        .frame(maxWidth: .infinity)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.echoPaperDim)
        )
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.top, 16)
        .padding(.bottom, 8)
    }

    private func loadGroupsInCommonSummary() async {
        guard !peerDID.isEmpty,
              let client = DIContainer.shared.resolveAPIClient() else { return }
        let social = ContactSocialAPIClient(apiClient: client)
        guard let rel = try? await social.fetchRelationship(peerDID: peerDID) else { return }
        let groups = rel.mutual_groups ?? []
        let count = rel.mutual_groups_count ?? groups.count
        guard count > 0 else {
            groupsInCommonText = "No groups in common"
            return
        }
        let names = groups.compactMap(\.name).filter { !$0.isEmpty }
        if names.isEmpty {
            groupsInCommonText = count == 1 ? "1 group in common" : "\(count) groups in common"
        } else if names.count == 1, count == 1 {
            groupsInCommonText = names[0]
        } else {
            let head = names.prefix(2).joined(separator: ", ")
            groupsInCommonText = count > 2 ? "\(head) +\(count - 2) more" : head
        }
    }

    private func refreshSmartReplies() async {
        let bodies = viewModel.messages.suffix(12).map(\.content)
        smartReplies = await OnDeviceAIService.shared.smartReplies(
            from: Array(bodies),
            conversationId: conversationId
        )
    }

    private func refreshThreadSummary() async {
        let bodies = viewModel.messages.map(\.content)
        threadSummary = await OnDeviceAIService.shared.summarizeThread(
            messages: bodies,
            conversationId: conversationId
        )
    }

    private var preferredTranslationLanguage: String {
        Locale.current.language.languageCode?.identifier ?? "en"
    }

    private func pinnedMessage(for id: String?) -> ChatDetailMessage? {
        guard let id else { return nil }
        return viewModel.messages.first { $0.id == id }
    }

    private func handleMessageAction(_ action: MessageAction, on message: ChatDetailMessage) {
        switch action {
        case .delete:
            let deletedId = message.id
            let wasPinned = pinnedMessageId == deletedId
            Task { @MainActor in
                await viewModel.deleteMessage(deletedId)
                if wasPinned {
                    #if os(iOS)
                    ConversationPinnedMessageStore.setPinnedMessageId(nil, conversationId: conversationId)
                    #endif
                    pinnedMessageId = nil
                }
            }
        case .copy:
            #if os(iOS)
            UIPasteboard.general.string = message.content
            #endif
        case .reply:
            viewModel.beginReply(to: message)
        case .edit:
            viewModel.beginEdit(message: message)
            messageText = message.content
        case .forward:
            forwardPreview = message.content
            forwardSourceMessageId = message.id
            forwardSourceConversationId = conversationId
            showForwardSheet = true
        case .pin:
            let pinId = message.id
            Task { @MainActor in
                guard let nowPinned = await viewModel.togglePin(messageId: pinId) else { return }
                #if os(iOS)
                ConversationPinnedMessageStore.setPinnedMessageId(nowPinned ? pinId : nil, conversationId: conversationId)
                #endif
                pinnedMessageId = nowPinned ? pinId : nil
            }
        case .translate:
            Task {
                if let translated = await OnDeviceAIService.shared.translateOnDevice(
                    text: message.content,
                    to: preferredTranslationLanguage,
                    conversationId: conversationId
                ) {
                    messageTranslations[message.id] = translated
                }
            }
        case .schedule:
            scheduleSourceText = message.content
            showScheduleMessage = true
        }
    }

    /// Phase A chat composer (`docs/design-previews/phaseA-chat.html` + `previews.css` `.composer`).
    private var chatComposerBar: some View {
        VStack(spacing: 0) {
            if voiceRecorder.isRecording {
                voiceRecordingBar
            }

            if showAttachmentPicker {
                AttachmentPickerView(
                    isPresented: $showAttachmentPicker,
                    onImageSelected: { data, mime in
                        Task {
                            await viewModel.sendMedia(
                                data: data,
                                mimeType: mime,
                                mediaKind: .image,
                                viewOnce: sendMediaViewOnce
                            )
                            sendMediaViewOnce = false
                        }
                    },
                    onVideoSelected: { data, mime in
                        Task { await viewModel.sendMedia(data: data, mimeType: mime, mediaKind: .video) }
                    },
                    onFileSelected: { data, mime in
                        Task { await viewModel.sendMedia(data: data, mimeType: mime, mediaKind: .file) }
                    },
                    onCloudTapped: {
                        showCloudPicker = true
                    },
                    onVoiceNoteTapped: {
                        try? voiceRecorder.startRecording()
                    },
                    onPollTapped: {
                        showAttachmentPicker = false
                        showCreatePoll = true
                    },
                    onPaymentTapped: WalletTransferAvailability.inChatPaymentsEnabled ? {
                        showAttachmentPicker = false
                        showPaymentSheet = true
                    } : nil,
                    onGifTapped: {
                        showGifPicker = true
                    }
                )
                .transition(.move(edge: .bottom).combined(with: .opacity))
                Toggle("View once", isOn: $sendMediaViewOnce)
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.echoInk70)
                    .padding(.horizontal, Spacing.md.rawValue)
                    .padding(.bottom, 6)
            }

            HStack(spacing: Spacing.md.rawValue) {
                Button {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        showAttachmentPicker.toggle()
                    }
                } label: {
                    Image(systemName: showAttachmentPicker ? "xmark.circle.fill" : "plus.circle.fill")
                        .font(.system(size: 24))
                        .foregroundStyle(Color.echoSignal)
                }
                .accessibilityLabel("Add attachment")

                TextField("Message…", text: $messageText)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 9)
                    .background(Color.echoPaper)
                    .clipShape(Capsule())
                    .overlay(Capsule().strokeBorder(Color.echoHair, lineWidth: 1))
                    .onChange(of: messageText) { _, newValue in
                        viewModel.onInputChanged(newValue)
                        #if os(iOS)
                        ComposerDraftStore.save(conversationId: conversationId, text: newValue)
                        #endif
                    }

                Button {
                    let text = messageText
                    messageText = ""
                    #if os(iOS)
                    ComposerDraftStore.clear(conversationId: conversationId)
                    #endif
                    Task {
                        if let editId = viewModel.editingMessageId {
                            _ = await viewModel.applyEdit(messageId: editId, newText: text)
                        } else {
                            await viewModel.sendMessage(text)
                        }
                    }
                } label: {
                    Image(systemName: "paperplane.fill")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundStyle(canSend ? Color.echoSignal : Color.echoInk40)
                }
                .disabled(!canSend)
                .accessibilityLabel("Send message")
            }
            .padding(.horizontal, Spacing.md.rawValue)
            .padding(.vertical, 12)
        }
        .overlay(alignment: .top) {
            Rectangle()
                .fill(Color.echoHair)
                .frame(height: 1)
        }
    }

    private var voiceRecordingBar: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.red)
                .frame(width: 10, height: 10)

            WaveformView(
                samples: voiceRecorder.waveformSamples,
                progress: 1.0,
                accentColor: .echoSignal
            )
            .frame(height: 28)

            Text(voiceRecorderElapsed)
                .font(.system(size: 14, weight: .medium).monospacedDigit())
                .foregroundStyle(Color.echoInk)

            Spacer()

            Button {
                voiceRecorder.cancelRecording()
            } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 22))
                    .foregroundStyle(Color.echoInk40)
            }
            .buttonStyle(.plain)

            Button {
                Task { await viewModel.sendVoiceNote(from: voiceRecorder) }
            } label: {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(Color.echoSignal)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, Spacing.md.rawValue)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
    }

    private var voiceRecorderElapsed: String {
        let total = Int(voiceRecorder.elapsed)
        let minutes = total / 60
        let seconds = total % 60
        return String(format: "%d:%02d", minutes, seconds)
    }



    @ViewBuilder
    private func messageRow(for message: ChatDetailMessage) -> some View {
        let rowAlignment: HorizontalAlignment = message.isFromCurrentUser ? .trailing : .leading
        VStack(alignment: rowAlignment, spacing: 4) {
            if let quote = message.replyPreview, !quote.isEmpty {
                replyQuoteBubble(quote, isFromCurrentUser: message.isFromCurrentUser)
            }
            if message.forwardedFromMessageId != nil {
                Text("Forwarded")
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(.echoInk40)
            }
            messageBubbleContent(for: message)
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
                    },
                    onShowReactors: { emoji, reactors in
                        reactorDetail = (emoji, reactors)
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
                reactionQuickBar(for: message)
                .transition(.scale.combined(with: .opacity))
                .offset(y: -44)
            }
        }
    }
    @ViewBuilder
    private func reactionQuickBar(for message: ChatDetailMessage) -> some View {
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
    }
    /// Trust-aware scam/impersonation check, run once when the thread opens. Shows the safety
    /// banner on a risky first contact (unverified/low-trust, or name-impersonation of a contact).
    private func evaluateSafety() {
        guard !safetyEvaluated else { return }
        safetyEvaluated = true
        #if os(iOS)
        let tier = ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: peerDID)
        let isFirst = FirstContactStore.isFirstContact(peerDID)
        let assessment = ContactSafetyEvaluator().evaluate(
            peerDID: peerDID,
            peerName: contactName,
            peerTier: tier,
            isFirstContact: isFirst,
            knownContacts: KnownContactsProvider.forSafety(excludingPeerDID: peerDID)
        )
        if assessment.level != .ok { safetyAssessment = assessment }
        FirstContactStore.markSeen(peerDID)
        #endif
    }

    // MARK: - Payments

    private func handlePaymentCompose(kind: PaymentPayload.Kind, amount: Decimal, memo: String?) {
        guard WalletTransferAvailability.inChatPaymentsEnabled else { return }
        let coordinator = PaymentCoordinator(transfer: WalletTransferAvailability.makeTransfer())
        guard let request = try? coordinator.makeRequest(amount: amount, memo: memo) else { return }
        if kind == .request {
            Task { await viewModel.sendMessage(PaymentMessageEncoding.encode(request)) }
        } else {
            attemptPay(request)
        }
    }

    /// Pay-guard: paying an unverified/low-trust contact prompts an extra confirmation first.
    private func attemptPay(_ request: PaymentPayload) {
        guard WalletTransferAvailability.inChatPaymentsEnabled else { return }
        #if os(iOS)
        let tier = ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: peerDID)
        if PaymentCoordinator(transfer: WalletTransferAvailability.makeTransfer()).needsExtraConfirmation(peerTier: tier) {
            pendingPayRequest = request
            return
        }
        #endif
        Task { await finalizePay(request) }
    }

    private func finalizePay(_ request: PaymentPayload) async {
        guard WalletTransferAvailability.inChatPaymentsEnabled else { return }
        let coordinator = PaymentCoordinator(transfer: WalletTransferAvailability.makeTransfer())
        guard let sent = try? await coordinator.pay(request, toDID: peerDID) else { return }
        await viewModel.sendMessage(PaymentMessageEncoding.encode(sent))
    }

    @ViewBuilder
    private func replyQuoteBubble(_ quote: String, isFromCurrentUser: Bool) -> some View {
        HStack(spacing: 0) {
            Rectangle()
                .fill(Color.echoSignal)
                .frame(width: 3)
            Text(quote)
                .font(.system(size: 12))
                .foregroundColor(.echoInk55)
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 8))
        .frame(maxWidth: 280, alignment: isFromCurrentUser ? .trailing : .leading)
    }

    @ViewBuilder
    private func messageBubbleContent(for message: ChatDetailMessage) -> some View {
        if let payment = PaymentMessageEncoding.decode(message.content) {
            PaymentBubbleView(
                payment: payment,
                isIncoming: !message.isFromCurrentUser,
                payActionsEnabled: WalletTransferAvailability.inChatPaymentsEnabled,
                onPay: { attemptPay(payment) }
            )
        } else {
            textBubbleContent(for: message)
        }
    }

    @ViewBuilder
    private func textBubbleContent(for message: ChatDetailMessage) -> some View {
        let peerStatus: DeliveryStatus? = message.isFromCurrentUser
            ? viewModel.displayedDeliveryStatus(message.deliveryStatus)
            : nil
        if let pollId = message.pollId, let poll = viewModel.polls[pollId] {
            PollBubbleView(
                poll: poll,
                currentUserDID: currentUserDID,
                isSent: message.isFromCurrentUser,
                onVote: { optionId in
                    Task { await viewModel.votePoll(pollId: pollId, optionId: optionId) }
                },
                onClose: {
                    Task { await viewModel.closePoll(pollId: pollId) }
                }
            )
        } else if let mediaRef = message.mediaRef {
            MediaBubbleView(
                mediaRef: mediaRef,
                peerDID: peerDID,
                isSent: message.isFromCurrentUser,
                timestamp: message.timestamp,
                deliveryStatus: peerStatus
            )
        } else {
            VStack(alignment: message.isFromCurrentUser ? .trailing : .leading, spacing: 4) {
                DisappearingMessageView(
                    messageId: message.id,
                    expiresAt: message.expiresAt,
                    onExpired: { messageId in
                        Task { await viewModel.handleDisappearingExpired(messageId: messageId) }
                    }
                ) {
                    MessageBubble(
                        message: message.content,
                        isSent: message.isFromCurrentUser,
                        status: mapDeliveryStatus(viewModel.displayedDeliveryStatus(message.deliveryStatus)),
                        deliveryStatus: peerStatus,
                        timestamp: message.timestamp
                    )
                }
                if let translated = messageTranslations[message.id], translated != message.content {
                    Text(translated)
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk70)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 6)
                        .background(Color.echoPaperDim)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                        .frame(maxWidth: 280, alignment: message.isFromCurrentUser ? .trailing : .leading)
                }
                #if os(iOS)
                if MessageLinkPreview.firstURL(in: message.content) != nil {
                    ChatInlineLinkPreview(
                        text: message.content,
                        alignTrailing: message.isFromCurrentUser
                    )
                }
                #endif
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
        case .anchorVerificationFailed: return .failed
        }
    }

    private func sendGif(_ gif: GifSearchService.GifResult) async {
        guard let url = gif.previewURL else {
            await viewModel.sendMessage("GIF: \(gif.title)")
            return
        }
        do {
            let (data, response) = try await URLSession.shared.data(from: url)
            guard let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode) else {
                throw URLError(.badServerResponse)
            }
            await viewModel.sendMedia(data: data, mimeType: "image/gif", mediaKind: .image)
        } catch {
            await viewModel.sendMessage("GIF: \(gif.title)")
        }
    }
}

private struct ReactorDetailItem: Identifiable {
    let id = UUID()
    let emoji: String
    let reactors: [String]
}

private struct GifPickerSheet: View {
    let onSelect: (GifSearchService.GifResult) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var query = ""
    @State private var results: [GifSearchService.GifResult] = []

    var body: some View {
        NavigationStack {
            List(results) { gif in
                Button {
                    onSelect(gif)
                    dismiss()
                } label: {
                    HStack {
                        Image(systemName: gif.previewURL == nil ? "face.smiling" : "photo")
                            .foregroundStyle(Color.echoSignal)
                        Text(gif.title)
                            .foregroundStyle(Color.echoInk)
                    }
                }
            }
            .navigationTitle("GIFs")
            .searchable(text: $query)
            .onChange(of: query, initial: true) {
                Task { results = await GifSearchService.shared.search(query: query) }
            }
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
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
#endif
