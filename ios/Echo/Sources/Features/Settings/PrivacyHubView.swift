#if os(iOS)
import SwiftUI

/// Single entry for privacy controls (Wave 0.5 / WO-228 consolidation shell).
struct PrivacyHubView: View {
    @State private var privacySettings = EnhancedPrivacySettings()
    @State private var personaPrivacySettings = PersonaPrivacySettings()

    var body: some View {
        List {
            Section("Messaging privacy") {
                NavigationLink("Privacy & security") {
                    PrivacySecuritySettingsView(settings: $privacySettings)
                }
                NavigationLink("Persona privacy") {
                    PersonaPrivacySettingsView(settings: $personaPrivacySettings)
                }
            }
            Section("Discovery") {
                NavigationLink("Contact discovery (PSI)") {
                    ContactDiscoveryView()
                }
                NavigationLink("Phone discoverability") {
                    ContactDiscoverySettingsView()
                }
            }
            Section("Account") {
                NavigationLink("Data & deletion") {
                    AccountDataView()
                }
            }
        }
        .navigationTitle("Privacy")
        .onAppear { privacySettings = PrivacySettingsStore.load() }
        .onDisappear { PrivacySettingsStore.save(privacySettings) }
    }
}

/// PSI discoverability toggle wired to `/v3/contacts/discovery-settings` (WO-220 / WO-228).
struct ContactDiscoverySettingsView: View {
    @State private var discoverable = false
    @State private var tierDefault = false
    @State private var trustTier = 0
    @State private var isLoading = true
    @State private var errorMessage: String?

    private var api: ContactDiscoveryAPIClient? {
        guard let client = DIContainer.shared.resolveAPIClient() else { return nil }
        return ContactDiscoveryAPIClient(apiClient: client)
    }

    var body: some View {
        List {
            if let errorMessage {
                Text(errorMessage).foregroundStyle(.red).font(.footnote)
            }
            Section {
                Toggle("Allow others to find me by phone", isOn: $discoverable)
                    .disabled(isLoading || api == nil)
                    .onChange(of: discoverable) { _, newValue in
                        Task { await persistOptIn(newValue) }
                    }
            } footer: {
                Text(footerText)
            }
            Section("Remove from discovery") {
                Button("Remove my phone from discovery", role: .destructive) {
                    discoverable = false
                    Task { await persistOptIn(false) }
                }
                .disabled(isLoading || api == nil)
            }
        }
        .navigationTitle("Phone discovery")
        .task { await loadSettings() }
    }

    private var footerText: String {
        if trustTier >= 3 && tierDefault {
            return "Tier \(trustTier) accounts are discoverable by default. You can opt out anytime."
        }
        return "Tier \(trustTier) accounts are hidden by default. Opt in to let contacts find you via PSI."
    }

    private func loadSettings() async {
        guard let api else {
            errorMessage = "Sign in required to manage discovery settings."
            isLoading = false
            return
        }
        isLoading = true
        defer { isLoading = false }
        do {
            let settings = try await api.fetchDiscoverySettings()
            discoverable = settings.phone_discoverable ?? false
            tierDefault = settings.tier_default_discoverable ?? false
            trustTier = settings.trust_tier ?? CurrentUserSession.trustTier()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func persistOptIn(_ enabled: Bool) async {
        guard let api else { return }
        do {
            let settings = try await api.updateDiscoveryOptIn(enabled)
            discoverable = settings.phone_discoverable ?? enabled
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

/// Account deletion and export — links existing profile actions (WO-228).
private struct AccountDataView: View {
    @State private var showDeleteConfirm = false

    var body: some View {
        List {
            Section {
                Text("Your messages stay E2E encrypted. Account deletion revokes credentials on the Identity Metagraph; your did:key string remains mathematically valid.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            Section {
                Button(role: .destructive) {
                    showDeleteConfirm = true
                } label: {
                    Text("Delete account…")
                }
            }
        }
        .navigationTitle("Data & deletion")
        .alert("Delete account?", isPresented: $showDeleteConfirm) {
            Button("Cancel", role: .cancel) {}
            Button("Delete", role: .destructive) {
                Task {
                    _ = try? await DIContainer.shared.resolveUserRepository()?.deleteAccount()
                }
            }
        } message: {
            Text("This cannot be undone from the device alone. Confirm you have exported your recovery phrase.")
        }
    }
}
#endif
