#if os(iOS)
// Features/Bots/BotManagementView.swift
// Bot management screen with active bots and discovery (Stage 4 / WO-11)

import SwiftUI

// MARK: - Bot Management View

public struct BotManagementView: View {
    @StateObject private var viewModel = BotManagementViewModel()

    public init() {}

    public var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.system(size: 13))
                        .foregroundColor(.echoAlert)
                        .padding(.horizontal, 20)
                }

                if !viewModel.activeBots.isEmpty {
                    SectionLabel("MY BOTS")
                    ForEach(viewModel.activeBots) { bot in
                        BotCard(bot: bot) {
                            viewModel.selectedBot = bot
                            viewModel.showBotDetail = true
                        }
                    }
                }

                SectionLabel("DISCOVER BOTS")
                if viewModel.availableBots.isEmpty && !viewModel.isLoading {
                    Text("No bots available right now.")
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk55)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, 28)
                }
                ForEach(viewModel.availableBots) { bot in
                    DiscoverBotRow(bot: bot) {
                        viewModel.beginInstall(bot)
                    }
                }

                HStack(spacing: 8) {
                    Image(systemName: "shield.checkered")
                        .foregroundStyle(Color.Echo.outline)
                    Text("All bots run in a sandboxed environment with limited permissions.")
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.outline)
                }
                .padding(16)
                .background(
                    RoundedRectangle(cornerRadius: 16)
                        .fill(Color.Echo.surfaceContainerLow)
                )
                .padding(.horizontal, 20)
            }
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Bot Management")
        .sheet(isPresented: $viewModel.showBotDetail) {
            if let bot = viewModel.selectedBot {
                BotDetailView(bot: bot, viewModel: viewModel)
            }
        }
        .sheet(item: $viewModel.pendingInstall) { manifest in
            BotInstallPermissionsSheet(
                bot: manifest,
                onInstall: { permissions in
                    Task { await viewModel.confirmInstall(manifest: manifest, permissions: permissions) }
                },
                onCancel: { viewModel.pendingInstall = nil }
            )
        }
        .task { await viewModel.loadBots() }
        .refreshable { await viewModel.loadBots() }
    }
}

// MARK: - Section Label

struct SectionLabel: View {
    let text: String

    init(_ text: String) {
        self.text = text
    }

    var body: some View {
        Text(text)
            .font(.system(size: 10))
            .fontWeight(.bold)
            .tracking(2)
            .foregroundStyle(Color.Echo.outline)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, 28)
    }
}

// MARK: - Bot Card

struct BotCard: View {
    let bot: BotInfo
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 12) {
                RoundedRectangle(cornerRadius: 16)
                    .fill(Color.Echo.surfaceContainerHigh)
                    .frame(width: 48, height: 48)
                    .overlay(
                        Image(systemName: "cpu")
                            .font(.system(size: 20))
                            .foregroundStyle(Color.Echo.primaryContainer)
                    )

                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 6) {
                        Text(bot.name)
                            .font(Font.Echo.bodyMedium)
                            .fontWeight(.semibold)
                            .foregroundStyle(Color.Echo.onSurface)
                        Text("\(bot.trustScore)")
                            .font(.system(size: 11, weight: .bold))
                            .foregroundColor(trustColor)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(trustColor.opacity(0.12))
                            .clipShape(Capsule())
                    }
                    Text(bot.description)
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.outline)
                        .lineLimit(1)
                }

                Spacer()

                Circle()
                    .fill(bot.isActive ? Color.Echo.success : Color.Echo.outline)
                    .frame(width: 8, height: 8)
            }
            .padding(16)
            .background(
                RoundedRectangle(cornerRadius: 24)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder(opacity: 0.15)
        }
        .buttonStyle(SpringButtonStyle())
        .padding(.horizontal, 20)
    }

    private var trustColor: Color {
        if bot.trustScore >= 80 { return .echoSignal }
        if bot.trustScore >= 60 { return .echoInk70 }
        return .echoAlert
    }
}

// MARK: - Discover Bot Row

struct DiscoverBotRow: View {
    let bot: BotInfo
    let onAdd: () -> Void

    var body: some View {
        HStack(spacing: 12) {
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.Echo.surfaceContainerHigh)
                .frame(width: 48, height: 48)
                .overlay(
                    Image(systemName: "cpu")
                        .font(.system(size: 20))
                        .foregroundStyle(Color.Echo.primaryContainer)
                )

            VStack(alignment: .leading, spacing: 4) {
                Text(bot.name)
                    .font(Font.Echo.bodyMedium)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.onSurface)
                Text(bot.description)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.outline)
                    .lineLimit(2)
            }

            Spacer()

            Button("Add") { onAdd() }
                .font(.system(size: 12))
                .fontWeight(.bold)
                .foregroundStyle(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(Capsule().fill(LinearGradient.signature))
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 8)
    }
}

// MARK: - Bot Detail View

struct BotDetailView: View {
    let bot: BotInfo
    @ObservedObject var viewModel: BotManagementViewModel
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    VStack(spacing: 12) {
                        RoundedRectangle(cornerRadius: 24)
                            .fill(Color.Echo.surfaceContainerHigh)
                            .frame(width: 80, height: 80)
                            .overlay(
                                Image(systemName: "cpu")
                                    .font(.system(size: 36))
                                    .foregroundStyle(Color.Echo.primaryContainer)
                            )

                        Text(bot.name)
                            .font(.system(size: 24))
                            .fontWeight(.heavy)

                        Text(bot.description)
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.outline)
                            .multilineTextAlignment(.center)
                    }
                    .padding(.top, 16)

                    GhostBorderSection(title: "CONFIGURATION") {
                        InfoRow(label: "Status", value: bot.isActive ? "Active" : "Inactive")
                        InfoRow(label: "Trust score", value: "\(bot.trustScore)/100")
                        InfoRow(label: "Permissions", value: bot.permissions)
                    }

                    VStack(spacing: 12) {
                        Button(bot.isActive ? "Disable Bot" : "Enable Bot") {
                            Task {
                                await viewModel.setActive(botDID: bot.id, active: !bot.isActive)
                                dismiss()
                            }
                        }
                        .font(.system(size: 14)).fontWeight(.bold)
                        .foregroundStyle(bot.isActive ? Color.Echo.error : Color.Echo.primaryContainer)

                        Button("Remove Bot") {
                            Task {
                                await viewModel.removeBot(botDID: bot.id)
                                dismiss()
                            }
                        }
                        .font(.system(size: 14)).fontWeight(.bold)
                        .foregroundStyle(Color.Echo.error.opacity(0.7))
                    }
                }
                .padding(.bottom, 40)
            }
            .background(Color.Echo.surface)
            .navigationTitle("Bot Details")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button("Done") { dismiss() }
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
            }
        }
    }
}

// MARK: - Bot Model

struct BotInfo: Identifiable {
    let id: String
    let name: String
    let description: String
    let isActive: Bool
    let trustScore: Int
    let permissions: String

    static func from(manifest: BotManifestDTO) -> BotInfo {
        BotInfo(
            id: manifest.botDID,
            name: manifest.name,
            description: manifest.description,
            isActive: false,
            trustScore: manifest.trustScore,
            permissions: manifest.requiredPermissions.joined(separator: ", ")
        )
    }

    static func from(install: BotInstallationDTO, manifest: BotManifestDTO?) -> BotInfo {
        BotInfo(
            id: install.botDID,
            name: manifest?.name ?? "Installed bot",
            description: manifest?.description ?? install.botDID,
            isActive: install.active,
            trustScore: manifest?.trustScore ?? 0,
            permissions: install.grantedPermissions.joined(separator: ", ")
        )
    }
}

// MARK: - Bot Management ViewModel

@MainActor
class BotManagementViewModel: ObservableObject {
    @Published var activeBots: [BotInfo] = []
    @Published var availableBots: [BotInfo] = []
    @Published var selectedBot: BotInfo?
    @Published var showBotDetail = false
    @Published var pendingInstall: BotManifestDTO?
    @Published var isLoading = false
    @Published var errorMessage: String?

    private var catalog: [BotManifestDTO] = []

    private var api: BotAPIClient? {
        guard let client = DIContainer.shared.resolveAPIClient() else { return nil }
        return BotAPIClient(apiClient: client)
    }

    func loadBots() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        guard let api else {
            catalog = localFallbackCatalog()
            rebuildLists(installed: [])
            return
        }

        do {
            async let catalogTask = api.fetchCatalog()
            async let installedTask = api.fetchInstalled()
            catalog = try await catalogTask
            let installed = try await installedTask
            rebuildLists(installed: installed)
        } catch {
            errorMessage = error.localizedDescription
            catalog = localFallbackCatalog()
            rebuildLists(installed: [])
        }
    }

    func beginInstall(_ bot: BotInfo) {
        guard let manifest = catalog.first(where: { $0.botDID == bot.id }) else { return }
        pendingInstall = manifest
    }

    func confirmInstall(manifest: BotManifestDTO, permissions: [String]) async {
        guard let api else { return }
        pendingInstall = nil
        do {
            _ = try await api.install(botDID: manifest.botDID, permissions: permissions)
            await loadBots()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func removeBot(botDID: String) async {
        guard let api else { return }
        do {
            try await api.uninstall(botDID: botDID)
            await loadBots()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func setActive(botDID: String, active: Bool) async {
        guard let api else { return }
        do {
            _ = try await api.setActive(botDID: botDID, active: active)
            await loadBots()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func rebuildLists(installed: [BotInstallationDTO]) {
        let installedIDs = Set(installed.map(\.botDID))
        let manifestByID = Dictionary(uniqueKeysWithValues: catalog.map { ($0.botDID, $0) })
        activeBots = installed.map { BotInfo.from(install: $0, manifest: manifestByID[$0.botDID]) }
        availableBots = catalog
            .filter { !installedIDs.contains($0.botDID) }
            .map(BotInfo.from(manifest:))
    }

    private func localFallbackCatalog() -> [BotManifestDTO] {
        [
            BotManifestDTO(
                botDID: "did:key:z6MkechoReminderBot00000000000000000001",
                name: "Echo Reminder",
                description: "Schedule gentle nudges in your chats.",
                version: "0.1.0",
                requiredPermissions: ["send_message", "read_messages"],
                trustScore: 82
            ),
        ]
    }
}

#endif
