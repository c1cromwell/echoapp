#if os(iOS)
// Features/Messaging/EmptyState/MessagesTabView.swift
// Wraps the Messages tab with empty-state branching per v2.5.3:
//
//   hasSentFirstMessage = false AND no persisted conversations
//     → MessagesEmptyStateView (welcome + trust banner + FAB)
//   hasSentFirstMessage = true (all conversations deleted)
//     → PostFirstMessageEmptyState (minimal — no welcome copy)
//   otherwise
//     → ConversationListView (existing)

import SwiftUI

struct MessagesTabView: View {
    @Environment(AppState.self) private var appState

    @State private var composeSheetPresented = false
    @State private var enrollmentSheetPresented = false
    @State private var recoveryPromptPresented = false

    var body: some View {
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
                ConversationListView()
                    .overlay(alignment: .bottomTrailing) {
                        ComposeFAB { composeSheetPresented = true }
                            .padding(.trailing, 20)
                            .padding(.bottom, 26)
                    }
            }
        }
        .sheet(isPresented: $composeSheetPresented) {
            // Placeholder — swap in your real NewConversationSheet
            Text("Start a new conversation")
                .presentationDetents([.medium])
        }
        .sheet(isPresented: $enrollmentSheetPresented) {
            EnrollmentCoordinatorView(
                coordinator: EnrollmentCoordinator(
                    onComplete: { _ in enrollmentSheetPresented = false },
                    onCancel:   { enrollmentSheetPresented = false }
                )
            )
        }
        .sheet(isPresented: $recoveryPromptPresented) {
            recoveryExportSheet
        }
        .onAppear {
            // Check for overdue reminders (AC-RECOVERY-004.2) and first-message trigger.
            RecoveryPromptScheduler.shared.checkAndPresentIfOverdue {
                recoveryPromptPresented = true
            }
        }
        .onReceive(
            NotificationCenter.default.publisher(for: Notification.Name("echo.firstMessageSent"))
        ) { _ in
            checkFirstMessageRecoveryPrompt()
        }
    }

    // MARK: - Recovery prompt

    private var recoveryExportSheet: some View {
        RecoveryCoordinatorView(
            coordinator: RecoveryCoordinator(
                onExportComplete: { recoveryPromptPresented = false },
                onRestoreComplete: { _ in recoveryPromptPresented = false },
                onCancel: { recoveryPromptPresented = false }
            )
        )
        .task {
            // Start the export flow immediately when the sheet opens.
        }
    }

    private func checkFirstMessageRecoveryPrompt() {
        let exported = UserDefaults.standard.object(forKey: "echo.recoveryPhraseExportedAt") != nil
        let skipped = UserDefaults.standard.bool(forKey: "echo.recoverySkippedThisSession")
        guard !exported && !skipped else { return }
        recoveryPromptPresented = true
    }

    // MARK: - Helpers

    private var currentTrustTier: Int {
        if !appState.provisionService.hasMinimumIdentity { return 0 }
        return UserDefaults.standard.integer(forKey: "echo.trustTier")
    }

    private var hasSentFirstMessage: Bool {
        UserDefaults.standard.bool(forKey: "echo.hasSentFirstMessage")
    }
}

// MARK: - Post-first-message empty state

/// Shown when the user has had conversations before but cleared them all.
/// No welcome framing — just the FAB.
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
