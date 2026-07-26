import SwiftUI

/// Per-chat settings (spec §5.4). Signal conversation-settings analog: silent, disappearing
/// timer, verify (DID safety-number analog), block/report. Persists via `ConversationPreferencesStore`.
public struct ChatSettingsSheet: View {
    let contactName: String
    let conversationId: String
    @State private var prefs: ConversationPreferences
    @State private var isArchived: Bool
    @State private var allowedTimers: [DisappearingTimer] = DisappearingTimer.allCases
    @State private var restrictionReason: String?
    #if os(iOS)
    @State private var hiddenFolderId: String
    #endif

    let onChange: (ConversationPreferences) -> Void
    let onArchiveChange: (Bool) -> Void
    let onVerify: () -> Void
    let onBlock: () -> Void

    public init(
        contactName: String,
        conversationId: String,
        preferences: ConversationPreferences,
        isArchived: Bool = false,
        onChange: @escaping (ConversationPreferences) -> Void,
        onArchiveChange: @escaping (Bool) -> Void = { _ in },
        onVerify: @escaping () -> Void = {},
        onBlock: @escaping () -> Void = {}
    ) {
        self.contactName = contactName
        self.conversationId = conversationId
        self._prefs = State(initialValue: preferences)
        self._isArchived = State(initialValue: isArchived)
        #if os(iOS)
        self._hiddenFolderId = State(initialValue: HiddenFolderStore.folderId(for: conversationId))
        #endif
        self.onChange = onChange
        self.onArchiveChange = onArchiveChange
        self.onVerify = onVerify
        self.onBlock = onBlock
    }

    public var body: some View {
        VStack(spacing: 0) {
            Capsule()
                .fill(Color.echoHair)
                .frame(width: 38, height: 5)
                .padding(.top, 8)
                .padding(.bottom, 14)

            Text(contactName)
                .font(.system(size: 17, weight: .semibold))
                .foregroundColor(.echoInk)
                .padding(.bottom, 6)

            // Silent notifications
            row(icon: "bell.slash", title: "Silent notifications") {
                Toggle("", isOn: Binding(
                    get: { prefs.isMuted },
                    set: { prefs.isMuted = $0; onChange(prefs) }
                ))
                .labelsHidden()
                .tint(.echoSignal)
            }

            // Disappearing messages
            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: "timer").frame(width: 24).foregroundColor(.echoInk55)
                    Text("Disappearing messages")
                        .font(.system(size: 15))
                        .foregroundColor(.echoInk)
                    Spacer()
                }
                Picker("", selection: Binding(
                    get: { prefs.disappearing },
                    set: { prefs.disappearing = $0; onChange(prefs) }
                )) {
                    ForEach(allowedTimers, id: \.self) { t in
                        Text(t.label).tag(t)
                    }
                }
                .pickerStyle(.segmented)
                if let restrictionReason {
                    Text(restrictionReason)
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk55)
                }
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
            .overlay(Divider(), alignment: .bottom)
            #if os(iOS)
            .task {
                if let remote = await DisappearingRestrictionsAPI.fetchPolicy() {
                    allowedTimers = DisappearingRestrictionsAPI.allowedTimers(policy: remote.policy)
                    restrictionReason = remote.policy.reason
                }
            }
            #endif

            // Hide conversation
            row(icon: "eye.slash", title: "Hide conversation") {
                Toggle("", isOn: Binding(
                    get: { prefs.isHidden },
                    set: { prefs.isHidden = $0; onChange(prefs) }
                ))
                .labelsHidden()
                .tint(.echoSignal)
            }

            #if os(iOS)
            if prefs.isHidden {
                hiddenFolderPicker
            }
            #endif

            // Archive conversation (WO-198)
            row(icon: "archivebox", title: "Archive conversation") {
                Toggle("", isOn: Binding(
                    get: { isArchived },
                    set: { isArchived = $0; onArchiveChange($0) }
                ))
                .labelsHidden()
                .tint(.echoSignal)
            }

            // Verify (DID safety number — Signal Parity Wave S2)
            Button(action: onVerify) {
                row(icon: "checkmark.shield", title: "Verify safety number", showChevron: true) { EmptyView() }
            }
            .buttonStyle(.plain)

            #if os(iOS)
            // Chat wallpaper (Wave S4)
            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: "photo").frame(width: 24).foregroundColor(.echoInk55)
                    Text("Chat wallpaper")
                        .font(.system(size: 15))
                        .foregroundColor(.echoInk)
                    Spacer()
                }
                Picker("", selection: Binding(
                    get: { ChatWallpaperStore.style(for: conversationId) },
                    set: { ChatWallpaperStore.set($0, for: conversationId) }
                )) {
                    ForEach(ChatWallpaperStyle.allCases) { style in
                        Text(style.title).tag(style)
                    }
                }
                .pickerStyle(.segmented)
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
            .overlay(Divider(), alignment: .bottom)
            #endif

            // Block / report
            Button(action: onBlock) {
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: "hand.raised").frame(width: 24)
                    Text("Block & report")
                        .font(.system(size: 15))
                    Spacer()
                }
                .foregroundColor(.echoAlert)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, Spacing.md.rawValue)
            }
            .buttonStyle(.plain)

            Spacer(minLength: 0)
        }
        .background(Color.echoPaper)
    }

    #if os(iOS)
    private var hiddenFolderPicker: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: "folder.fill").frame(width: 24).foregroundColor(.echoInk55)
                Text("Hidden folder")
                    .font(.system(size: 15))
                    .foregroundColor(.echoInk)
                Spacer()
            }
            Picker("", selection: $hiddenFolderId) {
                ForEach(HiddenFolderStore.all()) { folder in
                    Text(folder.name).tag(folder.id)
                }
            }
            .pickerStyle(.menu)
            .onChange(of: hiddenFolderId) { _, newId in
                HiddenFolderStore.assign(conversationId: conversationId, folderId: newId)
            }
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, Spacing.md.rawValue)
        .overlay(Divider(), alignment: .bottom)
    }
    #endif

    @ViewBuilder
    private func row<Trailing: View>(
        icon: String,
        title: String,
        showChevron: Bool = false,
        @ViewBuilder trailing: () -> Trailing
    ) -> some View {
        HStack(spacing: Spacing.md.rawValue) {
            Image(systemName: icon).frame(width: 24).foregroundColor(.echoInk55)
            Text(title)
                .font(.system(size: 15))
                .foregroundColor(.echoInk)
            Spacer()
            trailing()
            if showChevron {
                Image(systemName: "chevron.right")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.echoInk40)
            }
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, Spacing.md.rawValue)
        .overlay(Divider(), alignment: .bottom)
    }
}

#if DEBUG
struct ChatSettingsSheet_Previews: PreviewProvider {
    static var previews: some View {
        ChatSettingsSheet(
            contactName: "Aria Rao",
            conversationId: "preview-conv",
            preferences: ConversationPreferences(isMuted: true, disappearing: .h24),
            onChange: { _ in }
        )
    }
}
#endif
