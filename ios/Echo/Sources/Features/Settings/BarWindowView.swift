import SwiftUI

/// Echo Bar. One column, composer pinned to the bottom. Not a tab.
struct BarWindowView: View {
    @State private var session: BarSession
    @State private var draft = ""
    @FocusState private var composerFocused: Bool
    #if os(iOS)
    @State private var chat: ChatDetailViewModel?
    #endif

    init(session: BarSession = BarSession()) {
        _session = State(initialValue: session)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            header
            transcript
            composer
        }
        .padding(Spacing.lg.rawValue)
        .frame(maxWidth: 420, maxHeight: .infinity, alignment: .top)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color.echoPaper)
        .task { await reloadAllowed() }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    Text(session.selected?.name ?? "Echo")
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                    Text(headerDetail)
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoInk55)
                }
                Spacer()
                if session.selected != nil {
                    Button("Switch") { closeThread() }
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.echoSignal)
                    Button("Revoke") { revokeSelected() }
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.echoAlert)
                }
            }
            if let notice = session.notice {
                Text(notice)
                    .font(.system(size: 13))
                    .foregroundStyle(Color.echoInk70)
            }
        }
    }

    private var headerDetail: String {
        if let person = session.selected {
            return "Allowed · \(person.name)"
        }
        return "People who can reach you here"
    }

    private var transcript: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                if let person = session.selected {
                    threadLines(for: person)
                } else {
                    ForEach(session.allowed) { person in
                        personRow(person, actionTitle: nil) {
                            Task { await open(person) }
                        }
                    }
                    if !session.candidates.isEmpty {
                        Text("From this search")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.echoInk55)
                            .padding(.top, Spacing.sm.rawValue)
                        ForEach(session.candidates) { person in
                            personRow(person, actionTitle: "Allow") {
                                Task { await allow(person) }
                            }
                        }
                    }
                    #if os(iOS)
                    Button("Add from contacts") {
                        Task { await addFromContacts() }
                    }
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(Color.echoSignal)
                    .padding(.top, Spacing.sm.rawValue)
                    #endif
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    private func personRow(_ person: BarPerson, actionTitle: String?, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack {
                Text(person.name.isEmpty ? person.did : person.name)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(Color.echoInk)
                Spacer()
                if person.allowed {
                    Text("Allowed")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(Color.echoTrustGreen)
                } else if let actionTitle {
                    Text(actionTitle)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(Color.echoSignal)
                }
            }
            .padding(.vertical, Spacing.sm.rawValue)
        }
        .buttonStyle(.plain)
    }

    private var composer: some View {
        HStack(spacing: Spacing.sm.rawValue) {
            TextField(composerPlaceholder, text: $draft, axis: .vertical)
                .focused($composerFocused)
                .lineLimit(1...4)
                .textFieldStyle(.plain)
                .font(.system(size: 16))
                .foregroundStyle(Color.echoInk)
                .onSubmit { submit() }
            Button("Send") { submit() }
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.echoPaper)
                .padding(.horizontal, Spacing.md.rawValue)
                .padding(.vertical, Spacing.sm.rawValue)
                .background(Color.echoSignal)
                .clipShape(Capsule())
        }
        .padding(.horizontal, Spacing.md.rawValue)
        .padding(.vertical, Spacing.sm.rawValue)
        .background(Color.echoPaperDim)
        .clipShape(Capsule())
        .overlay(Capsule().stroke(Color.echoHair, lineWidth: 1))
    }

    private var composerPlaceholder: String {
        if let name = session.selected?.name, !name.isEmpty {
            return "Message \(name)"
        }
        return "Name, @username, or email"
    }

    @ViewBuilder
    private func threadLines(for person: BarPerson) -> some View {
        #if os(iOS)
        if let chat, chat.peerDID == person.did {
            if chat.messages.isEmpty {
                Text("No messages yet.")
                    .font(.system(size: 15))
                    .foregroundStyle(Color.echoInk55)
            }
            ForEach(chat.messages) { message in
                lineBubble(message.content, mine: message.isFromCurrentUser)
            }
        } else {
            Text("Opening the thread.")
                .font(.system(size: 15))
                .foregroundStyle(Color.echoInk55)
        }
        #else
        ForEach(session.lines[person.did] ?? []) { line in
            lineBubble(line.text, mine: line.mine)
        }
        #endif
    }

    private func lineBubble(_ text: String, mine: Bool) -> some View {
        HStack {
            if mine { Spacer(minLength: Spacing.xxl.rawValue) }
            Text(text)
                .font(.system(size: 15))
                .foregroundStyle(mine ? Color.echoPaper : Color.echoInk)
                .padding(.horizontal, Spacing.md.rawValue)
                .padding(.vertical, Spacing.sm.rawValue)
                .background(mine ? Color.echoSignal : Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 12))
            if !mine { Spacer(minLength: Spacing.xxl.rawValue) }
        }
    }

    private func submit() {
        if let person = session.selected {
            let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !text.isEmpty else { return }
            draft = ""
            #if os(iOS)
            Task { await deliver(text, to: person) }
            #else
            _ = session.send(text)
            #endif
            return
        }
        switch BarSession.interpret(draft) {
        case .empty:
            break
        case .emailNotLive:
            session.notice = "Email lookup is not live yet. Use an @username."
            draft = ""
        case .username(let handle):
            draft = ""
            Task { await lookup(handle) }
        }
    }

    private func revokeSelected() {
        guard let did = session.selected?.did else { return }
        session.revoke(did: did)
        Task { await revokeOnServer(did) }
    }

    private func allow(_ person: BarPerson) async {
        #if os(iOS)
        do {
            _ = try await barClient()?.allowBar(did: person.did)
        } catch {
            session.notice = "Allow did not reach Echo."
            return
        }
        #endif
        session.allow(person)
    }

    private func lookup(_ handle: String) async {
        #if os(iOS)
        guard let client = barClient() else {
            session.notice = "Sign in to find @\(handle)."
            return
        }
        do {
            let hits = try await client.searchUsername(handle)
            if hits.isEmpty {
                session.notice = "No one uses @\(handle)."
                return
            }
            session.noteDiscovery(hits.map {
                BarPerson(did: $0.did, name: "@\($0.username)", allowed: false)
            })
            session.notice = nil
        } catch {
            session.notice = "Could not look up @\(handle)."
        }
        #else
        session.notice = "Look up @\(handle) after this Mac signs in with the device key."
        #endif
    }

    private func open(_ person: BarPerson) async {
        session.select(did: person.did)
        #if os(iOS)
        await bindChat(to: person)
        #endif
    }

    private func closeThread() {
        session.showList()
        #if os(iOS)
        let current = chat
        chat = nil
        Task { await current?.disconnect() }
        #endif
    }

    private func deliver(_ text: String, to person: BarPerson) async {
        #if os(iOS)
        if chat?.peerDID != person.did {
            await bindChat(to: person)
        }
        guard let chat else { return }
        _ = await chat.sendMessage(text)
        if let message = chat.errorMessage, !message.isEmpty {
            session.notice = "Saved in this thread. Echo did not deliver it yet."
        }
        #endif
    }

    #if os(iOS)
    private func bindChat(to person: BarPerson) async {
        let me = await CurrentUserSession.currentDID() ?? ""
        guard !me.isEmpty else {
            session.notice = "Sign in on this device before messaging."
            session.showList()
            return
        }
        if let current = chat, current.peerDID != person.did {
            await current.disconnect()
        }
        let service = DIContainer.shared.resolveConversationSignalService()
            ?? ConversationSignalService(apiBaseURL: EchoAPIBaseURL.resolved)
        let vm = ChatDetailViewModel(signalService: service)
        vm.configure(
            conversationId: BarThread.id(a: me, b: person.did),
            peerDID: person.did,
            currentUserDID: me,
            peerDisplayName: person.name
        )
        chat = vm
        await MessageRelaySession.connect()
        if let token = try? await KeychainManager.shared.getAuthToken(), !token.isEmpty {
            await vm.connect(accessToken: token)
        }
    }
    #endif

    #if os(iOS)
    private func addFromContacts() async {
        guard let discovery = DIContainer.shared.resolveContactDiscoveryUseCase() else {
            session.notice = "Contact discovery is not available."
            return
        }
        do {
            let matches = try await discovery.discoverContacts()
            session.noteDiscovery(matches.map {
                BarPerson(did: $0.did, name: $0.displayName, allowed: false)
            })
            if matches.isEmpty {
                session.notice = "No contacts on Echo yet. Allow is a separate step."
            }
        } catch {
            session.notice = "Could not read contacts."
        }
    }

    private func revokeOnServer(_ did: String) async {
        try? await barClient()?.revokeBar(did: did)
    }

    private func reloadAllowed() async {
        guard let client = barClient() else { return }
        guard let rows = try? await client.listBar() else { return }
        for row in rows {
            guard let did = row.contactDid, !did.isEmpty else { continue }
            let label = did.split(separator: ":").last.map(String.init) ?? did
            let name = label.count > 12 ? "…" + label.suffix(8) : label
            session.allow(BarPerson(did: did, name: name, allowed: true))
        }
    }

    private func barClient() -> ContactSocialAPIClient? {
        let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
        return ContactSocialAPIClient(apiClient: client)
    }
    #else
    private func reloadAllowed() async {}
    private func revokeOnServer(_ did: String) async {}
    #endif
}
