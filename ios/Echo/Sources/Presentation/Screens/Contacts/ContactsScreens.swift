import SwiftUI

/// Contacts - Contacts List Screen
public struct ContactsListView: View {
    @State private var searchText = ""
    @State private var selectedFilter = "All"
    @State private var showInviteSheet = false
    #if os(iOS)
    @State private var viewModel = ContactsListViewModel()
    @State private var chatThread: StoredConversation?
    #else
    @State private var contacts: [ContactModel] = []
    #endif

    let onSelectContact: (String) -> Void
    
    public init(onSelectContact: @escaping (String) -> Void = { _ in }) {
        self.onSelectContact = onSelectContact
    }
    
    var filteredContacts: [ContactModel] {
        #if os(iOS)
        var filtered = viewModel.contacts
        #else
        var filtered = contacts
        #endif
        
        if !searchText.isEmpty {
            filtered = filtered.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        
        if selectedFilter != "All" {
            filtered = filtered.filter { $0.trustLevel == selectedFilter }
        }
        
        return filtered.sorted { $0.name < $1.name }
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Contacts",
                    showBackButton: false,
                    trailingAction: { showInviteSheet = true },
                    trailingIcon: Image(systemName: "person.badge.plus")
                )

                NavigationLink {
                    UsernameSearchView()
                } label: {
                    HStack {
                        Image(systemName: "at")
                        Text("Search by @username")
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, Spacing.sm.rawValue)
                }

                NavigationLink {
                    ContactDiscoveryView()
                } label: {
                    HStack {
                        Image(systemName: "person.2.circle")
                        Text("Find contacts on ECHO")
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, Spacing.sm.rawValue)
                }
                
                VStack(spacing: Spacing.lg.rawValue) {
                    // Search Bar
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
                    
                    // Filter Picker
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
                    } else if let error = viewModel.errorMessage, viewModel.contacts.isEmpty {
                        ContentUnavailableView {
                            Label("Couldn't load contacts", systemImage: "exclamationmark.triangle")
                        } description: {
                            Text(error)
                        } actions: {
                            Button("Retry") { Task { await viewModel.refresh() } }
                        }
                        .frame(maxHeight: .infinity, alignment: .center)
                    } else
                    #endif
                    if filteredContacts.isEmpty {
                        VStack(spacing: Spacing.md.rawValue) {
                            Image(systemName: "person.slash")
                                .font(.system(size: 48))
                                .foregroundColor(.echoGray400)
                            
                            Text("No contacts found")
                                .typographyStyle(.h4, color: .echoGray600)
                            
                            Text("Find people on ECHO or add by @username")
                                .typographyStyle(.body, color: .echoSecondaryText)
                        }
                        .frame(maxHeight: .infinity, alignment: .center)
                    } else {
                        List {
                            ForEach(filteredContacts) { contact in
                                ContactListItem(
                                    name: contact.name,
                                    username: contact.username,
                                    trustLevel: contact.trustLevel,
                                    onTap: {
                                        onSelectContact(contact.id)
                                        #if os(iOS)
                                        Task {
                                            if let thread = await ContactThreadHelper.upsertDirectThread(
                                                peerDID: contact.id,
                                                displayName: contact.name
                                            ) {
                                                chatThread = thread
                                            }
                                        }
                                        #endif
                                    }
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
        .sheet(isPresented: $showInviteSheet) {
            InviteLinkSheet()
        }
        #if os(iOS)
        .task { await viewModel.refresh() }
        .refreshable { await viewModel.refresh() }
        .navigationDestination(item: $chatThread) { conversation in
            ChatDestinationView(conversation: conversation)
        }
        #endif
    }
}

struct ContactModel: Identifiable {
    let id: String
    let name: String
    let username: String
    let trustLevel: String
}

// MARK: - Trust Dashboard Screen
// `TrustDashboardView` now lives in Features/Trust/TrustDashboardView.swift —
// the paper/ink T0–T4 trust-tier ladder + verification sources + VIP CTA (Phase B).

// MARK: - Preview

#if DEBUG
struct ContactsAndTrustScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            ContactsListView()
        }
    }
}
#endif
