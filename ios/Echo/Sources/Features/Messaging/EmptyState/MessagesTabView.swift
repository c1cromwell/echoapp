#if os(iOS)
// Features/Messaging/EmptyState/MessagesTabView.swift
// Wraps the Messages tab with empty-state branching and Wave 0.1 chat navigation.

import SwiftUI

struct MessagesTabView: View {
    @Environment(AppState.self) private var appState
    @Environment(\.scenePhase) private var scenePhase
    @Bindable private var conversationStore = ConversationStore.shared
    @ObservedObject private var incomingCalls = IncomingCallPresenter.shared

    @State private var composeSheetPresented = false
    @State private var enrollmentSheetPresented = false
    @State private var recoveryPromptPresented = false
    @State private var pendingInviteCode: String?
    @State private var chatPath = SwiftUI.NavigationPath()
    @Bindable private var hiddenSession = HiddenChatsSession.shared
    @State private var showBiometricGate = false
    @State private var showHiddenSettings = false
    @State private var composeHiddenPresented = false
    @State private var showCreateGroup = false
    @State private var showMessageSearch = false
    @State private var messageSearchSeed = ""

    private var personaFilteredConversations: [StoredConversation] {
        conversationStore.conversations.filter {
            $0.personaId == appState.activePersona.id
                && !ConversationPreferencesStore.shared.isHidden($0.id)
        }
    }

    var body: some View {
        NavigationStack(path: $chatPath) {
            Group {
                if !hasSentFirstMessage {
                    MessagesEmptyStateView(
                        displayName: appState.displayName,
                        trustTier: currentTrustTier,
                        onComposeTapped: { composeSheetPresented = true },
                        onUpgradeTrustTapped: { enrollmentSheetPresented = true }
                    )
                    .navigationTitle("Messages")
                    .navigationBarTitleDisplayMode(.inline)
                } else {
                    MessagesHubView(
                        conversations: personaFilteredConversations,
                        hiddenConversations: hiddenConversations,
                        personas: appState.personas,
                        activePersona: appState.activePersona,
                        trustTier: { conversationId in
                            guard let conv = conversationStore.conversation(id: conversationId) else { return 1 }
                            return ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: conv.peerDID)
                        },
                        mutedIDs: mutedConversationIDs,
                        hiddenUnlocked: hiddenSession.isUnlocked,
                        isDuressHiddenVault: hiddenSession.isDuressMode,
                        onSelectConversation: { id in
                            if let conversation = conversationStore.conversation(id: id) {
                                chatPath.append(conversation)
                            }
                        },
                        onCompose: { composeSheetPresented = true },
                        onComposeHidden: { composeHiddenPresented = true },
                        onOpenHidden: { showBiometricGate = true },
                        onLockHidden: { hiddenSession.lock() },
                        onOpenHiddenSettings: { showHiddenSettings = true },
                        onSwitchPersona: { appState.switchPersona($0) },
                        onSelectHiddenPersona: { _ in showBiometricGate = true },
                        onToggleArchive: { conversationId, archived in
                            Task {
                                if let client = DIContainer.shared.resolveAPIClient().map({
                                    LiveConversationArchiveAPIClient(apiClient: $0)
                                }) {
                                    try? await client.setArchived(archived, conversationId: conversationId)
                                } else {
                                    ConversationArchiveStore.setArchived(archived, conversationId: conversationId)
                                }
                            }
                        },
                        onOpenMessageSearch: { query in
                            messageSearchSeed = query
                            showMessageSearch = true
                        },
                        onGroupCreated: { groupId, name in
                            openGroupConversation(groupId: groupId, name: name)
                        }
                    )
                }
            }
            .navigationDestination(for: StoredConversation.self) { conversation in
                ChatDestinationView(conversation: conversation)
            }
        }
        .sheet(isPresented: $composeSheetPresented) {
            NewConversationSheet(
                onConversationCreated: { conversation in
                    chatPath.append(conversation)
                },
                onHiddenConversationCreated: { conversation in
                    ConversationPreferencesStore.shared.setHidden(true, for: conversation.id)
                    chatPath.append(conversation)
                },
                onNewGroup: {
                    showCreateGroup = true
                }
            )
        }
        .sheet(isPresented: $enrollmentSheetPresented) {
            EnrollmentCoordinatorView(
                coordinator: EnrollmentCoordinator(
                    onComplete: { bundle in
                        enrollmentSheetPresented = false
                        UserDefaults.standard.set(
                            bundle.assuranceLevel.trustTier,
                            forKey: "echo.trustTier"
                        )
                    },
                    onCancel: { enrollmentSheetPresented = false }
                )
            )
        }
        .sheet(isPresented: $showMessageSearch) {
            MessageSearchSheet(initialQuery: messageSearchSeed)
        }
        .sheet(isPresented: $recoveryPromptPresented) {
            recoveryExportSheet
        }
        .sheet(isPresented: Binding(
            get: { pendingInviteCode != nil },
            set: { if !$0 { pendingInviteCode = nil } }
        )) {
            if let code = pendingInviteCode {
                AcceptInviteSheet(inviteCode: code) {
                    pendingInviteCode = nil
                }
            }
        }
        .onAppear {
            RecoveryPromptScheduler.shared.checkAndPresentIfOverdue {
                recoveryPromptPresented = true
            }
            DeviceHistorySyncBootstrap.pullIfNeeded()
            SearchIndexSyncBootstrap.pullIfNeeded()
            ContactDiscoveryScheduler.runIfDue()
            BackupScheduler.runIfDue()
            if let code = appState.pendingInviteCode {
                pendingInviteCode = code
                appState.pendingInviteCode = nil
            }
        }
        .task {
            await connectSharedMessageRelay()
            await refreshContactTrustIndex()
        }
        .onChange(of: scenePhase) { _, phase in
            switch phase {
            case .background:
                hiddenSession.noteDidEnterBackground()
            case .active:
                hiddenSession.refreshForegroundLockIfNeeded()
                Task { await connectSharedMessageRelay() }
            default:
                break
            }
        }
        .sheet(isPresented: $showBiometricGate) {
            HiddenFolderGateSheet(
                onUnlocked: { showBiometricGate = false },
                onCancel: { showBiometricGate = false }
            )
        }
        .sheet(isPresented: $showHiddenSettings) {
            HiddenFolderSettingsView()
        }
        .sheet(isPresented: $composeHiddenPresented) {
            NewConversationSheet(
                onConversationCreated: { conversation in
                    ConversationPreferencesStore.shared.setHidden(true, for: conversation.id)
                    chatPath.append(conversation)
                },
                onHiddenConversationCreated: { conversation in
                    ConversationPreferencesStore.shared.setHidden(true, for: conversation.id)
                    chatPath.append(conversation)
                },
                onNewGroup: {
                    showCreateGroup = true
                }
            )
        }
        .sheet(isPresented: $showCreateGroup) {
            GroupCreateSheet { groupId, name in
                openGroupConversation(groupId: groupId, name: name)
            }
        }
        .fullScreenCover(item: Binding(
            get: { incomingCalls.activeCall },
            set: { if $0 == nil { incomingCalls.clearAfterDismiss() } }
        )) { call in
            CallView(
                peerDID: call.peerDID,
                callType: call.callType,
                isOutgoing: false,
                incomingCallId: call.callId,
                pendingOfferSDP: call.offerSDP
            )
            .onDisappear { incomingCalls.clearAfterDismiss() }
        }
        .onReceive(
            NotificationCenter.default.publisher(for: .echoPendingInvite)
        ) { notification in
            if let code = notification.userInfo?["code"] as? String {
                pendingInviteCode = code
            }
        }
        .onReceive(
            NotificationCenter.default.publisher(for: Notification.Name("echo.firstMessageSent"))
        ) { _ in
            checkFirstMessageRecoveryPrompt()
        }
        .hidesGlacialTabBarWhenPushed(!chatPath.isEmpty)
    }

    private var recoveryExportSheet: some View {
        RecoveryCoordinatorView(
            coordinator: RecoveryCoordinator(
                onExportComplete: { recoveryPromptPresented = false },
                onRestoreComplete: { _ in recoveryPromptPresented = false },
                onCancel: { recoveryPromptPresented = false }
            )
        )
    }

    private func checkFirstMessageRecoveryPrompt() {
        let exported = UserDefaults.standard.object(forKey: "echo.recoveryPhraseExportedAt") != nil
        let skipped = UserDefaults.standard.bool(forKey: "echo.recoverySkippedThisSession")
        guard !exported && !skipped else { return }
        recoveryPromptPresented = true
    }

    private var currentTrustTier: Int {
        if !appState.provisionService.hasMinimumIdentity { return 0 }
        return UserDefaults.standard.integer(forKey: "echo.trustTier")
    }

    private var hasSentFirstMessage: Bool {
        UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage")
            || !conversationStore.conversations.isEmpty
    }

    private var hiddenConversations: [StoredConversation] {
        guard hiddenSession.isUnlocked, !hiddenSession.isDuressMode else { return [] }
        return conversationStore.conversations.filter {
            ConversationPreferencesStore.shared.isHidden($0.id)
        }
    }

    private var mutedConversationIDs: Set<String> {
        let store = ConversationPreferencesStore.shared
        return Set(conversationStore.conversations.map(\.id).filter { store.isMuted($0) })
    }

    private func openGroupConversation(groupId: String, name: String) {
        let conversation = StoredConversation(
            id: "group:\(groupId)",
            contactName: name,
            peerDID: groupId,
            personaId: appState.activePersona.id
        )
        ConversationStore.shared.upsert(conversation)
        chatPath.append(conversation)
        if !UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage") {
            UserDefaults.standard.set(true, forKey: "echo.hasSentFirstMessage")
        }
    }

    private func refreshContactTrustIndex() async {
        guard let social = DIContainer.shared.resolveContactSocialAPI() else { return }
        if let contacts = try? await social.listContacts() {
            ContactTrustIndex.shared.ingestRemoteContacts(contacts)
        }
    }

    /// Keeps one WebSocket open for inbound chat + signals (Wave 0.1 relay E2E).
    private func connectSharedMessageRelay() async {
        guard let service = DIContainer.shared.resolveConversationSignalService(),
              let token = try? await KeychainManager.shared.getAuthToken() else { return }

        service.setInboundTextHandler { event in
            Task { @MainActor in
                let resolved = await InboundTextMessageResolver.resolveBody(for: event)
                let store = ConversationStore.shared
                let match = store.conversations.first(where: { $0.id == event.conversationId })
                    ?? store.conversations.first(where: { $0.peerDID == event.peerDID })
                guard let conversation = match else { return }

                let currentDID = await CurrentUserSession.currentDID() ?? ""
                guard !resolved.senderDID.isEmpty, resolved.senderDID != currentDID else { return }

                let inbound = ChatDetailMessage(
                    id: event.messageId,
                    senderDID: resolved.senderDID.isEmpty ? event.peerDID : resolved.senderDID,
                    currentUserDID: currentDID,
                    content: resolved.body,
                    timestamp: "Now",
                    deliveryStatus: .delivered
                )
                ConversationThreadStore.appendIfNew(conversationId: conversation.id, message: inbound)
                let session = HiddenChatsSession.shared
                guard session.shouldSurfaceNotification(for: conversation.id) else { return }
                let preview = session.redactedPreviewIfNeeded(
                    for: conversation.id,
                    resolved: resolved.preview
                )
                store.appendMessagePreview(conversationId: conversation.id, preview: preview)
                if ActiveChatRegistry.openConversationId != conversation.id {
                    store.incrementUnread(conversationId: conversation.id)
                }
            }
        }

        try? await service.connect(accessToken: token)
        #if os(iOS)
        await IncomingCallPresenter.shared.configureIfNeeded()
        ScreenshotAlertService.shared.configure(signalService: service)
        ScreenshotAlertService.shared.startMonitoring()
        #endif
    }
}

// MARK: - Post-first-message empty state

struct PostFirstMessageEmptyState: View {
    let onComposeTapped: () -> Void

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()

            VStack(spacing: 12) {
                Image(systemName: "bubble.left.and.bubble.right")
                    .font(.system(size: 30, weight: .light))
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                Text("No conversations yet")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(Color.Echo.onSurface)
                Text("Tap + to start one.")
                    .font(.system(size: 13))
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }

            ComposeFAB(onTap: onComposeTapped)
                .padding(.trailing, 20)
                .padding(.bottom, 26)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomTrailing)
        }
    }
}
#endif
