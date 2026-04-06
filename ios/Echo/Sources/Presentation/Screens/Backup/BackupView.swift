// Presentation/Screens/Backup/BackupView.swift
// Backup & security screen with recovery phrase, encrypted backup, and export options

import SwiftUI

// MARK: - Backup View

struct BackupView: View {
    @StateObject private var viewModel = BackupViewModel()

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Recovery phrase section
                GhostBorderSection(title: "RECOVERY PHRASE") {
                    HStack(spacing: 12) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundStyle(Color.Echo.warning)
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Keep this phrase secret")
                                .font(Font.Echo.bodyMedium).fontWeight(.semibold)
                            Text("Never share it with anyone")
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.outline)
                        }
                    }

                    Button {
                        viewModel.showRecoveryPhrase = true
                    } label: {
                        Text("View Recovery Phrase")
                            .font(.custom("Inter", size: 14)).fontWeight(.bold)
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Capsule().fill(LinearGradient.signature))
                    }
                    .deepGlacialShadow()

                    Text("Requires biometric authentication")
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.outline)
                }

                // Encrypted backup
                GhostBorderSection(title: "ENCRYPTED BACKUP") {
                    InfoRow(label: "Last backup", value: viewModel.lastBackupDate)
                    InfoRow(label: "Size", value: viewModel.backupSize)
                    InfoRow(label: "Location", value: "iCloud Keychain")

                    Button("Back Up Now") { Task { await viewModel.backupNow() } }
                        .font(.custom("Inter", size: 14)).fontWeight(.bold)
                        .foregroundStyle(Color.Echo.primaryContainer)
                }

                // Auto-backup settings
                GhostBorderSection(title: "AUTO-BACKUP") {
                    SettingsRow(icon: "clock", label: "Frequency", value: viewModel.backupFrequency)
                    SettingsRow(icon: "photo", label: "Include Media", value: viewModel.includeMedia ? "Yes" : "No")
                    SettingsRow(icon: "wifi", label: "WiFi Only", value: viewModel.wifiOnly ? "Yes" : "No")
                }

                // Export options
                GhostBorderSection(title: "EXPORT DATA") {
                    ExportButton(label: "Export Chat History", icon: "text.bubble")
                    ExportButton(label: "Export Contacts", icon: "person.2")
                    ExportButton(label: "Export Identity (DID)", icon: "person.text.rectangle")
                }

                // Danger zone
                VStack(spacing: 16) {
                    Text("DANGER ZONE")
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold).tracking(2)
                        .foregroundStyle(Color.Echo.error)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.leading, 28)

                    Button("Delete All Data") {
                        viewModel.showDeleteConfirmation = true
                    }
                    .font(.custom("Inter", size: 14)).fontWeight(.bold)
                    .foregroundStyle(Color.Echo.error)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(
                        RoundedRectangle(cornerRadius: 9999)
                            .stroke(Color.Echo.error.opacity(0.3), lineWidth: 1)
                    )
                    .padding(.horizontal, 20)
                }
            }
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Backup & Security")
    }
}

// MARK: - Backup ViewModel

@MainActor
class BackupViewModel: ObservableObject {
    @Published var lastBackupDate: String = "Never"
    @Published var backupSize: String = "—"
    @Published var backupFrequency: String = "Daily"
    @Published var includeMedia = true
    @Published var wifiOnly = true
    @Published var showRecoveryPhrase = false
    @Published var showDeleteConfirmation = false

    func backupNow() async {
        // TODO: Trigger encrypted backup to iCloud Keychain
    }
}

// MARK: - Info Row

struct InfoRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
            Spacer()
            Text(value)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
        }
    }
}

// MARK: - Export Button

struct ExportButton: View {
    let label: String
    let icon: String

    var body: some View {
        Button {
            // TODO: Trigger export
        } label: {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundStyle(Color.Echo.primaryContainer)
                    .frame(width: 24)
                Text(label)
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.onSurface)
                Spacer()
                Image(systemName: "arrow.down.circle")
                    .foregroundStyle(Color.Echo.outline)
            }
        }
    }
}
