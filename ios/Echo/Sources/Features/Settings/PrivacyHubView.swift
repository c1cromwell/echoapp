#if os(iOS)
import SwiftUI

/// Single entry for privacy controls (Wave 0.5 / WO-228 consolidation shell).
struct PrivacyHubView: View {
    @State private var privacySettings = EnhancedPrivacySettings()
    @State private var personaPrivacySettings = PersonaPrivacySettings()
    @State private var showPhoneBackup = false
    @State private var sessionDID = ""

    var body: some View {
        List {
            Section {
                HStack {
                    Text("Mobile number")
                    Spacer()
                    Text(PhoneBackupStatus.displayLabel)
                        .foregroundStyle(.secondary)
                }
                Button(PhoneBackupStatus.hasBackup ? "Update phone number" : "Add phone number") {
                    showPhoneBackup = true
                }
            } header: {
                Text("Phone backup")
            } footer: {
                Text("We store only a hash for recovery and optional PSI discovery. Raw numbers are not kept on the server.")
            }

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
        .task {
            privacySettings = PrivacySettingsStore.load()
            personaPrivacySettings = PersonaPrivacySettingsStore.load()
            sessionDID = await CurrentUserSession.currentDID() ?? ""
        }
        .onDisappear {
            PrivacySettingsStore.save(privacySettings)
            PersonaPrivacySettingsStore.save(personaPrivacySettings)
        }
        .sheet(isPresented: $showPhoneBackup) {
            if !sessionDID.isEmpty {
                SMSOTPSetupView(did: sessionDID) {
                    showPhoneBackup = false
                }
            }
        }
    }
}

/// PSI discoverability toggle wired to `/v3/contacts/discovery-settings` (WO-220 / WO-228).
struct ContactDiscoverySettingsView: View {
    @State private var discoverable = false
    @State private var tierDefault = false
    @State private var trustTier = 0
    @State private var isLoading = true
    @State private var errorMessage: String?
    @State private var syncCadence = ContactDiscoverySyncPreferences.cadence

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
            Section {
                Picker("Automatic scan", selection: $syncCadence) {
                    ForEach(ContactDiscoverySyncCadence.allCases, id: \.self) { cadence in
                        Text(cadence.label).tag(cadence)
                    }
                }
                .onChange(of: syncCadence) { _, newValue in
                    ContactDiscoverySyncPreferences.cadence = newValue
                }
            } header: {
                Text("Contact scan")
            } footer: {
                Text("Manual only runs when you tap Scan in Privacy → Contact discovery (PSI). Weekly and monthly are stored locally until background sync ships.")
            }
        }
        .navigationTitle("Phone discovery")
        .task {
            syncCadence = ContactDiscoverySyncPreferences.cadence
            await loadSettings()
        }
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
    @State private var isDeleting = false
    @State private var deleteError: String?

    var body: some View {
        List {
            Section {
                Text("Your messages stay E2E encrypted. Account deletion revokes server sessions and refresh tokens; durable Identity Metagraph revocation ships in a later wave.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            if let deleteError {
                Section {
                    Text(deleteError)
                        .font(.footnote)
                        .foregroundStyle(.red)
                }
            }
            Section {
                Button(role: .destructive) {
                    showDeleteConfirm = true
                } label: {
                    if isDeleting {
                        HStack {
                            ProgressView()
                            Text("Deleting…")
                        }
                    } else {
                        Text("Delete account…")
                    }
                }
                .disabled(isDeleting)
            }
        }
        .navigationTitle("Data & deletion")
        .alert("Delete account?", isPresented: $showDeleteConfirm) {
            Button("Cancel", role: .cancel) {}
            Button("Delete", role: .destructive) {
                Task { await deleteAccount() }
            }
        } message: {
            Text("This cannot be undone from the device alone. Confirm you have exported your recovery phrase.")
        }
    }

    private func deleteAccount() async {
        guard let repository = DIContainer.shared.resolveUserRepository() else {
            deleteError = "Sign in required to delete your account."
            return
        }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await repository.deleteAccount()
            await SessionSignOut.performIncludingOnboardingReset()
        } catch {
            deleteError = error.localizedDescription
        }
    }
}
#endif
