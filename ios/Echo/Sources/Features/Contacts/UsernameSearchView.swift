#if os(iOS)
import SwiftUI

/// Username search + add-contact (Phase A, WO-221/222).
/// Reuses `ContactSocialAPIClient.searchUsername` / `.addContact`; paper/ink styling.
struct UsernameSearchView: View {
    @State private var viewModel = UsernameSearchViewModel()
    @State private var query = ""

    var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: Spacing.md.rawValue) {
                searchField

                if viewModel.isSearching {
                    HStack { Spacer(); ProgressView("Searching…"); Spacer() }
                        .padding(.top, Spacing.lg.rawValue)
                } else if let error = viewModel.errorMessage {
                    ContentUnavailableView {
                        Label("Search failed", systemImage: "exclamationmark.triangle")
                    } description: {
                        Text(error)
                    }
                } else if viewModel.hasSearched && viewModel.hits.isEmpty {
                    ContentUnavailableView {
                        Label("No match", systemImage: "person.fill.questionmark")
                    } description: {
                        Text("No one on ECHO matches “\(viewModel.lastQuery)”. Usernames are case-insensitive.")
                    }
                } else {
                    List {
                        ForEach(viewModel.hits) { hit in
                            hitRow(hit)
                                .listRowBackground(Color.clear)
                        }
                    }
                    .listStyle(.plain)
                    .scrollContentBackground(.hidden)
                }

                Spacer(minLength: 0)
            }
            .padding(Spacing.lg.rawValue)
        }
        .navigationTitle("Add by username")
        .alert("Couldn't add contact", isPresented: addErrorBinding) {
            Button("OK", role: .cancel) { viewModel.addErrorMessage = nil }
        } message: {
            Text(viewModel.addErrorMessage ?? "")
        }
    }

    private var searchField: some View {
        HStack(spacing: Spacing.sm.rawValue) {
            Image(systemName: "at")
                .font(.system(size: 14))
                .foregroundColor(.echoInk40)

            TextField("username", text: $query)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
                .submitLabel(.search)
                .onSubmit { Task { await viewModel.search(query) } }
                .onChange(of: query) { _, newValue in viewModel.queryChanged(newValue) }

            if !query.isEmpty {
                Button { query = ""; viewModel.clear() } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk40)
                }
            }
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoPaperDim)
        .cornerRadius(12)
    }

    @ViewBuilder
    private func hitRow(_ hit: UsernameSearchHit) -> some View {
        let isAdded = viewModel.addedDIDs.contains(hit.did)
        let isAdding = viewModel.addingDID == hit.did

        HStack(spacing: Spacing.md.rawValue) {
            VStack(alignment: .leading, spacing: 4) {
                Text("@\(hit.username)")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.echoInk)
                Text(hit.did)
                    .font(.system(size: 12, design: .monospaced))
                    .foregroundColor(.echoInk40)
                    .lineLimit(1)
                    .truncationMode(.middle)
            }

            Spacer()

            TrustBadge(level: viewModel.trustLevel(forTier: hit.tier), size: .small)

            if isAdded {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundColor(.echoTrustGreen)
            } else if isAdding {
                ProgressView()
            } else {
                Button("Add") { Task { await viewModel.addContact(hit) } }
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.echoSignal)
            }
        }
        .padding(.vertical, Spacing.sm.rawValue)
    }

    private var addErrorBinding: Binding<Bool> {
        Binding(
            get: { viewModel.addErrorMessage != nil },
            set: { if !$0 { viewModel.addErrorMessage = nil } }
        )
    }
}

@MainActor
@Observable
final class UsernameSearchViewModel {
    var hits: [UsernameSearchHit] = []
    var isSearching = false
    var hasSearched = false
    var lastQuery = ""
    var errorMessage: String?
    var addedDIDs: Set<String> = []
    var addingDID: String?
    var addErrorMessage: String?

    private let socialAPI: ContactSocialAPIClient?
    private var debounceTask: Task<Void, Never>?

    init(socialAPI: ContactSocialAPIClient? = nil) {
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

    func queryChanged(_ text: String) {
        debounceTask?.cancel()
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else { clear(); return }
        debounceTask = Task { [weak self] in
            try? await Task.sleep(nanoseconds: 350_000_000)
            guard !Task.isCancelled else { return }
            await self?.search(trimmed)
        }
    }

    func search(_ handle: String) async {
        let trimmed = handle.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else { return }
        guard let socialAPI else {
            errorMessage = "Sign in required to search."
            return
        }
        isSearching = true
        errorMessage = nil
        lastQuery = trimmed
        defer { isSearching = false; hasSearched = true }
        do {
            hits = try await socialAPI.searchUsername(trimmed)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func addContact(_ hit: UsernameSearchHit) async {
        guard let socialAPI else {
            addErrorMessage = "Sign in required to add contacts."
            return
        }
        guard !addedDIDs.contains(hit.did) else { return }
        addingDID = hit.did
        addErrorMessage = nil
        defer { addingDID = nil }
        do {
            _ = try await socialAPI.addContact(did: hit.did, addedVia: "username_search")
            addedDIDs.insert(hit.did)
        } catch {
            addErrorMessage = error.localizedDescription
        }
    }

    func clear() {
        debounceTask?.cancel()
        hits = []
        hasSearched = false
        errorMessage = nil
        lastQuery = ""
    }

    /// Map backend trust tier (0–4) to a `TrustBadge` level string.
    func trustLevel(forTier tier: Int?) -> String {
        switch tier ?? 0 {
        case 0:  return "Newcomer"
        case 1:  return "Basic"
        case 2:  return "Verified"
        case 3:  return "Trusted"
        default: return "HighlyTrusted"
        }
    }
}
#endif
