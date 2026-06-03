#if os(iOS)
import SwiftUI

struct ContactDiscoveryView: View {
    @State private var viewModel = ContactDiscoveryViewModel()
    @State private var threadToOpen: StoredConversation?

    var body: some View {
        List {
            Section {
                Label(
                    OPRFClientFactory.runtimeMode == .live ? "Live OPRF" : "Mock OPRF (dev)",
                    systemImage: OPRFClientFactory.runtimeMode == .live ? "lock.shield" : "hammer"
                )
                .font(.subheadline)
                Text(OPRFClientFactory.modeFootnote)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            if let synced = viewModel.lastSyncedAt {
                Section {
                    Text("Last scan \(synced.formatted(date: .abbreviated, time: .shortened))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            if viewModel.isLoading {
                HStack {
                    Spacer()
                    ProgressView("Scanning contacts…")
                    Spacer()
                }
            } else if let error = viewModel.errorMessage {
                ContentUnavailableView {
                    Label("Discovery failed", systemImage: "exclamationmark.triangle")
                } description: {
                    Text(error)
                } actions: {
                    Button("Try again") { Task { await viewModel.sync() } }
                }
            } else if viewModel.contacts.isEmpty {
                ContentUnavailableView {
                    Label("No matches yet", systemImage: "person.2.slash")
                } description: {
                    Text("Contacts on ECHO will appear here after a private scan. Enable discovery in Settings → Privacy and add SMS backup during onboarding.")
                } actions: {
                    Button("Scan contacts") { Task { await viewModel.sync() } }
                }
            } else {
                ForEach(viewModel.contacts) { contact in
                    discoveredRow(contact)
                }
            }
        }
        .navigationTitle("Contacts on ECHO")
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button {
                    Task { await viewModel.sync() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }
                .disabled(viewModel.isLoading)
            }
        }
        .task {
            if viewModel.contacts.isEmpty { await viewModel.sync() }
        }
        .alert("Couldn't add contact", isPresented: addErrorBinding) {
            Button("OK", role: .cancel) { viewModel.addErrorMessage = nil }
        } message: {
            Text(viewModel.addErrorMessage ?? "")
        }
        .navigationDestination(item: $threadToOpen) { conversation in
            ChatDestinationView(conversation: conversation)
        }
    }

    private var addErrorBinding: Binding<Bool> {
        Binding(
            get: { viewModel.addErrorMessage != nil },
            set: { if !$0 { viewModel.addErrorMessage = nil } }
        )
    }

    @ViewBuilder
    private func discoveredRow(_ contact: DiscoveredContact) -> some View {
        let isAdded = viewModel.addedDIDs.contains(contact.did)
        let isAdding = viewModel.addingDID == contact.did

        HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 4) {
                Text(contact.displayName)
                    .font(.headline)
                Text(ContactThreadHelper.truncatedDID(contact.did))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
            Spacer()
            if isAdded {
                Button("Message") {
                    Task {
                        if let thread = await viewModel.threadForAdded(contact) {
                            threadToOpen = thread
                        }
                    }
                }
                .buttonStyle(.bordered)
                .controlSize(.small)
            } else if isAdding {
                ProgressView()
            } else {
                Button("Add") {
                    Task { await viewModel.addContact(contact) }
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
            }
        }
        .padding(.vertical, 4)
    }
}

@MainActor
@Observable
final class ContactDiscoveryViewModel {
    var contacts: [DiscoveredContact] = []
    var isLoading = false
    var errorMessage: String?
    var lastSyncedAt: Date?
    var addedDIDs: Set<String> = []
    var addingDID: String?
    var addErrorMessage: String?

    private let service: ContactDiscoveryService
    private let socialAPI: ContactSocialAPIClient?

    init(service: ContactDiscoveryService? = nil, socialAPI: ContactSocialAPIClient? = nil) {
        if let service {
            self.service = service
        } else if let resolved: ContactDiscoveryService = DIContainer.shared.resolveContactDiscoveryService() {
            self.service = resolved
        } else if let client = DIContainer.shared.resolveAPIClient() {
            let api = ContactDiscoveryAPIClient(apiClient: client)
            self.service = ContactDiscoveryService(oprf: OPRFClientFactory.makeDefault(), api: api)
        } else {
            let config = APIConfiguration.default
            let api = ContactDiscoveryAPIClient(apiClient: APIClient(configuration: config))
            self.service = ContactDiscoveryService(oprf: OPRFClientFactory.makeDefault(), api: api)
        }

        if let socialAPI {
            self.socialAPI = socialAPI
        } else if let resolved = DIContainer.shared.resolveContactSocialAPI() {
            self.socialAPI = resolved
        } else if let client = DIContainer.shared.resolveAPIClient() {
            self.socialAPI = ContactSocialAPIClient(apiClient: client)
        } else {
            self.socialAPI = nil
        }
    }

    func sync() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            contacts = try await service.discoverFromDeviceContacts()
            lastSyncedAt = Date()
        } catch ContactDiscoveryError.noMatches {
            contacts = []
            lastSyncedAt = Date()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func addContact(_ contact: DiscoveredContact) async {
        guard let socialAPI else {
            addErrorMessage = "Sign in required to add contacts."
            return
        }
        guard !addedDIDs.contains(contact.did) else { return }

        addingDID = contact.did
        addErrorMessage = nil
        defer { addingDID = nil }

        do {
            _ = try await socialAPI.addContact(did: contact.did, addedVia: "psi_discovery")
            addedDIDs.insert(contact.did)
            _ = await ContactThreadHelper.upsertDirectThread(
                peerDID: contact.did,
                displayName: contact.displayName
            )
        } catch {
            addErrorMessage = error.localizedDescription
        }
    }

    func threadForAdded(_ contact: DiscoveredContact) async -> StoredConversation? {
        await ContactThreadHelper.upsertDirectThread(
            peerDID: contact.did,
            displayName: contact.displayName
        )
    }
}
#endif
