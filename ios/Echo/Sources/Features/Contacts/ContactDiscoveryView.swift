#if os(iOS)
import SwiftUI

struct ContactDiscoveryView: View {
    @State private var viewModel = ContactDiscoveryViewModel()

    var body: some View {
        List {
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
                    Text("Contacts on ECHO will appear here after a private scan.")
                } actions: {
                    Button("Scan contacts") { Task { await viewModel.sync() } }
                }
            } else {
                ForEach(viewModel.contacts) { contact in
                    VStack(alignment: .leading, spacing: 4) {
                        Text(contact.displayName)
                            .font(.headline)
                        Text(contact.did)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }
                    .padding(.vertical, 4)
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
    }
}

@MainActor
@Observable
final class ContactDiscoveryViewModel {
    var contacts: [DiscoveredContact] = []
    var isLoading = false
    var errorMessage: String?

    private let service: ContactDiscoveryService

    init(service: ContactDiscoveryService? = nil) {
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
    }

    func sync() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            contacts = try await service.discoverFromDeviceContacts()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
