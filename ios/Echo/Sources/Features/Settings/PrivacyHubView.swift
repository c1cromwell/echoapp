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
            }
            Section("Account") {
                NavigationLink("Data & deletion") {
                    AccountDataView()
                }
            }
        }
        .navigationTitle("Privacy")
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
