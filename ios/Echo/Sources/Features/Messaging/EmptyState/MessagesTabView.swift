#if os(iOS)
// Features/Messaging/EmptyState/MessagesTabView.swift
// Wraps the Messages tab with empty-state branching and Wave 0.1 chat navigation.

import SwiftUI

struct MessagesTabView: View {
    @Environment(AppState.self) private var appState
    @Bindable private var conversationStore = ConversationStore.shared

    @State private var composeSheetPresented = false
    @State private var enrollmentSheetPresented = false
    @State private var recoveryPromptPresented = false
    @State private var pendingInviteCode: String?
    @State private var chatPath = SwiftUI.NavigationPath()
    @State private var hiddenPersonaGate: PersonaSummary?

    private var personaFilteredConversations: [StoredConversation] {
        conversationStore.conversations.filter { $0.personaId == appState.activePersona.id }
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
                        personas: appState.personas,
                        activePersona: appState.activePersona,
                        trustTier: { conversationId in
                            guard let conv = conversationStore.conversation(id: conversationId) else { return 1 }
                            return ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: conv.peerDID)
                        },
                        mutedIDs: mutedConversationIDs,
                        onSelectConversation: { id in
                            if let conversation = conversationStore.conversation(id: id) {
                                chatPath.append(conversation)
                            }
                        },
                        onCompose: { composeSheetPresented = true },
                        onOpenHidden: {
                            if let hidden = appState.personas.first(where: { $0.isHidden }) {
                                hiddenPersonaGate = hidden
                            }
                        },
                        onSwitchPersona: { appState.switchPersona($0) },
                        onSelectHiddenPersona: { hiddenPersonaGate = $0 }
                    )
                }
            }
            .navigationDestination(for: StoredConversation.self) { conversation in
                ChatDestinationView(conversation: conversation)
            }
        }
        .sheet(isPresented: $composeSheetPresented) {
            NewConversationSheet { conversation in
                chatPath.append(conversation)
            }
        }
        .sheet(isPresented: $enrollmentSheetPresented) {
            EnrollmentCoordinatorView(
                coordinator: EnrollmentCoordinator(
                    onComplete: { _ in enrollmentSheetPresented = false },
                    onCancel: { enrollmentSheetPresented = false }
                )
            )
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
            if let code = appState.pendingInviteCode {
                pendingInviteCode = code
                appState.pendingInviteCode = nil
            }
        }
        .task {
            await connectSharedMessageRelay()
            await refreshContactTrustIndex()
        }
        .sheet(item: $hiddenPersonaGate) { persona in
            PersonaGateView(personaID: persona.id) {
                Color.clear
                    .onAppear {
                        appState.switchPersona(persona)
                        hiddenPersonaGate = nil
                    }
            }
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

    /// Conversations the user has muted — drives the mute badge on hub rows.
    private var mutedConversationIDs: Set<String> {
        let store = ConversationPreferencesStore.shared
        return Set(conversationStore.conversations.map(\.id).filter { store.isMuted($0) })
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
                var preview = event.text
                if let wire = event.wirePayload, wire.encrypted != nil {
                    let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
                    let crypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
                    if let decrypted = try? await crypto.decryptPayload(wire) {
                        preview = decrypted
                    } else {
                        preview = "Encrypted message"
                    }
                }
                let store = ConversationStore.shared
                let match = store.conversations.first(where: { $0.id == event.conversationId })
                    ?? store.conversations.first(where: { $0.peerDID == event.peerDID })
                guard let conversation = match else { return }
                store.appendMessagePreview(conversationId: conversation.id, preview: preview)
            }
        }

        try? await service.connect(accessToken: token)
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
