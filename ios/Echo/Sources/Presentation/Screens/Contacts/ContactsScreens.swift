#if os(iOS)
import SwiftUI
#if os(iOS)
import Contacts
#endif

/// Contacts - Contacts List Screen
public struct ContactsListView: View {
    @State private var searchText = ""
    @State private var selectedFilter = "All"
    @State private var showFavoritesOnly = false
    @State private var showAddContact = false
    @State private var showUsernameSearch = false
    #if os(iOS)
    @State private var viewModel = ContactsListViewModel()
    @State private var selectedContact: ContactModel?
    @State private var chatThread: StoredConversation?
    @State private var importedContacts: [ContactModel] = []
    #else
    @State private var contacts: [ContactModel] = []
    #endif

    let onSelectContact: (String) -> Void

    public init(onSelectContact: @escaping (String) -> Void = { _ in }) {
        self.onSelectContact = onSelectContact
    }

    private var displayedContacts: [ContactModel] {
        #if os(iOS)
        if !viewModel.contacts.isEmpty {
            return viewModel.contacts
        }
        return importedContacts
        #else
        return contacts
        #endif
    }

    var filteredContacts: [ContactModel] {
        var filtered = displayedContacts

        if !searchText.isEmpty {
            filtered = filtered.filter {
                $0.name.localizedCaseInsensitiveContains(searchText)
                    || $0.username.localizedCaseInsensitiveContains(searchText)
            }
        }

        if selectedFilter != "All" {
            filtered = filtered.filter { $0.trustLevel == selectedFilter }
        }

        if showFavoritesOnly {
            filtered = filtered.filter { ContactFavoritesStore.isFavorite(did: $0.id) }
        }

        return filtered.sorted { lhs, rhs in
            let lf = ContactFavoritesStore.isFavorite(did: lhs.id)
            let rf = ContactFavoritesStore.isFavorite(did: rhs.id)
            if lf != rf { return lf && !rf }
            return lhs.name < rhs.name
        }
    }

    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Contacts",
                    showBackButton: false,
                    trailingAction: { showAddContact = true },
                    trailingIcon: Image(systemName: "plus")
                )

                VStack(spacing: Spacing.lg.rawValue) {
                    HStack {
                        Image(systemName: "magnifyingglass")
                            .font(.system(size: 14))
                            .foregroundColor(.echoGray500)

                        TextField("Search contacts", text: $searchText)
                            .textFieldStyle(.roundedBorder)

                        if !searchText.isEmpty {
                            Button(action: { searchText = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .font(.system(size: 14))
                                    .foregroundColor(.echoGray400)
                            }
                        }
                    }
                    .padding(Spacing.md.rawValue)
                    .background(Color.echoSurface)
                    .cornerRadius(12)

                    Toggle("Favorites only", isOn: $showFavoritesOnly)
                        .padding(.horizontal, Spacing.md.rawValue)

                    Picker("Filter", selection: $selectedFilter) {
                        Text("All").tag("All")
                        Text("Inner Circle").tag("Inner Circle")
                        Text("Trusted").tag("Trusted")
                        Text("Acquaintance").tag("Acquaintance")
                    }
                    .pickerStyle(.segmented)
                    .padding(.horizontal, Spacing.md.rawValue)

                    #if os(iOS)
                    if viewModel.isLoading && viewModel.contacts.isEmpty {
                        ProgressView("Loading contacts…")
                            .frame(maxHeight: .infinity, alignment: .center)
                    } else if displayedContacts.isEmpty {
                        contactsEmptyState
                    } else if filteredContacts.isEmpty {
                        contactsEmptyState
                    } else {
                        contactsList
                    }
                    #else
                    if filteredContacts.isEmpty {
                        contactsEmptyState
                    } else {
                        contactsList
                    }
                    #endif
                }
                .echoSpacing(.lg)
            }
        }
        #if os(iOS)
        .sheet(isPresented: $showAddContact) {
            NewContactSheet {
                Task { await viewModel.refresh() }
            }
        }
        .task { await viewModel.refresh() }
        .refreshable { await viewModel.refresh() }
        .navigationDestination(item: $selectedContact) { contact in
            ContactDetailView(contactId: contact.id, displayName: contact.name) {
                Task { await openChat(with: contact) }
            }
        }
        .navigationDestination(item: $chatThread) { conversation in
            ChatDestinationView(conversation: conversation)
        }
        .navigationDestination(isPresented: $showUsernameSearch) {
            UsernameSearchView()
        }
        #endif
    }

    #if os(iOS)
    private func openChat(with contact: ContactModel) async {
        selectedContact = nil
        onSelectContact(contact.id)
        if let thread = await ContactThreadHelper.upsertDirectThread(
            peerDID: contact.id,
            displayName: contact.name
        ) {
            chatThread = thread
        }
    }

    private func importPhoneContacts() async {
        let store = CNContactStore()
        do {
            let granted = try await store.requestAccess(for: .contacts)
            guard granted else { return }
            let keys = [CNContactGivenNameKey, CNContactFamilyNameKey, CNContactPhoneNumbersKey] as [CNKeyDescriptor]
            let request = CNContactFetchRequest(keysToFetch: keys)
            var imported: [ContactModel] = []
            try store.enumerateContacts(with: request) { contact, _ in
                let name = [contact.givenName, contact.familyName]
                    .filter { !$0.isEmpty }
                    .joined(separator: " ")
                guard !name.isEmpty else { return }
                let phone = contact.phoneNumbers.first?.value.stringValue ?? ""
                imported.append(ContactModel(id: UUID().uuidString, name: name, username: phone, trustLevel: "Acquaintance"))
            }
            importedContacts = imported
            await viewModel.refresh()
        } catch {
            // Permission denied or fetch failed — no action needed
        }
    }
    #endif

    private var contactsEmptyState: some View {
        VStack(spacing: 20) {
            Image(systemName: "person.crop.circle.badge.plus")
                .font(.system(size: 52))
                .foregroundColor(.echoInk40)

            Text("No contacts yet")
                .font(.system(size: 18, weight: .semibold))
                .foregroundColor(.echoInk)

            Text("Import from your phone or add a new contact manually.")
                .font(.system(size: 14))
                .foregroundColor(.echoInk55)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 40)

            #if os(iOS)
            VStack(spacing: 12) {
                Button {
                    Task { await importPhoneContacts() }
                } label: {
                    HStack(spacing: 8) {
                        Image(systemName: "person.2.fill")
                            .font(.system(size: 14))
                        Text("Import from Phone")
                            .font(.system(size: 15, weight: .medium))
                    }
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 13)
                    .background(Color.echoSignal)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Button {
                    showAddContact = true
                } label: {
                    HStack(spacing: 8) {
                        Image(systemName: "plus")
                            .font(.system(size: 14, weight: .semibold))
                        Text("Add Contact")
                            .font(.system(size: 15, weight: .medium))
                    }
                    .foregroundColor(.echoSignal)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 13)
                    .background(Color.echoPaperDim)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Button {
                    showUsernameSearch = true
                } label: {
                    HStack(spacing: 8) {
                        Image(systemName: "at")
                            .font(.system(size: 14, weight: .semibold))
                        Text("Add by @username")
                            .font(.system(size: 15, weight: .medium))
                    }
                    .foregroundColor(.echoSignal)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 13)
                    .background(Color.echoPaperDim)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
            .padding(.horizontal, 40)
            .padding(.top, 8)
            #endif
        }
        .frame(maxHeight: .infinity, alignment: .center)
    }

    private var contactsList: some View {
        List {
            #if os(iOS)
            Button {
                showUsernameSearch = true
            } label: {
                HStack(spacing: 10) {
                    Image(systemName: "person.badge.plus")
                        .foregroundColor(.echoSignal)
                    Text("Add contact by @username")
                        .font(.system(size: 15, weight: .medium))
                        .foregroundColor(.echoSignal)
                    Spacer()
                    Image(systemName: "chevron.right")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.echoInk40)
                }
            }
            .listRowBackground(Color.echoSurface)
            #endif

            ForEach(filteredContacts) { contact in
                ContactListItem(
                    name: contact.name,
                    username: contact.username,
                    trustLevel: contact.trustLevel,
                    isFavorite: ContactFavoritesStore.isFavorite(did: contact.id),
                    onTap: { selectedContact = contact }
                )
                .listRowSeparator(.hidden)
                .listRowInsets(.init())
                .listRowBackground(Color.clear)
                .padding(.vertical, Spacing.xs.rawValue)
                .contextMenu {
                    Button {
                        ContactFavoritesStore.toggle(did: contact.id)
                    } label: {
                        Label(
                            ContactFavoritesStore.isFavorite(did: contact.id) ? "Remove favorite" : "Add favorite",
                            systemImage: "star"
                        )
                    }
                    Button {
                        Task { await openChat(with: contact) }
                    } label: {
                        Label("Message", systemImage: "message")
                    }
                    Button {
                        selectedContact = contact
                    } label: {
                        Label("View profile", systemImage: "person.crop.circle")
                    }
                }
                .swipeActions(edge: .leading) {
                    Button {
                        ContactFavoritesStore.toggle(did: contact.id)
                    } label: {
                        Label("Favorite", systemImage: "star.fill")
                    }
                    .tint(.yellow)
                }
                .swipeActions(edge: .trailing) {
                    Button {
                        Task { await openChat(with: contact) }
                    } label: {
                        Label("Message", systemImage: "message.fill")
                    }
                    .tint(.blue)
                }
            }
        }
        .listStyle(.plain)
        .scrollContentBackground(.hidden)
    }
}

struct ContactModel: Identifiable, Hashable, Sendable {
    let id: String
    let name: String
    let username: String
    let trustLevel: String
}

#if DEBUG
struct ContactsAndTrustScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            ContactsListView()
        }
    }
}
#endif
#endif
