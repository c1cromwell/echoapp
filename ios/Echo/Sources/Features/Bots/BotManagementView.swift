// Features/Bots/BotManagementView.swift
// Bot management screen with active bots and discovery

import SwiftUI

// MARK: - Bot Management View

struct BotManagementView: View {
    @StateObject private var viewModel = BotManagementViewModel()

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // My Bots section
                if !viewModel.activeBots.isEmpty {
                    SectionLabel("MY BOTS")
                    ForEach(viewModel.activeBots) { bot in
                        BotCard(bot: bot) {
                            viewModel.selectedBot = bot
                            viewModel.showBotDetail = true
                        }
                    }
                }

                // Discover Bots section
                SectionLabel("DISCOVER BOTS")
                ForEach(viewModel.availableBots) { bot in
                    DiscoverBotRow(bot: bot) {
                        Task { await viewModel.addBot(bot) }
                    }
                }

                // Sandbox disclaimer
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
                BotDetailView(bot: bot)
            }
        }
        .task { await viewModel.loadBots() }
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
            .font(.custom("Inter", size: 10))
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
                // Bot icon
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
                        .lineLimit(1)
                }

                Spacer()

                // Status indicator
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
                .font(.custom("Inter", size: 12))
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
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Bot header
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
                            .font(.custom("Inter", size: 24))
                            .fontWeight(.heavy)

                        Text(bot.description)
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.outline)
                            .multilineTextAlignment(.center)
                    }
                    .padding(.top, 16)

                    // Configuration
                    GhostBorderSection(title: "CONFIGURATION") {
                        InfoRow(label: "Status", value: bot.isActive ? "Active" : "Inactive")
                        InfoRow(label: "Last Triggered", value: bot.lastTriggered ?? "Never")
                        InfoRow(label: "Permissions", value: bot.permissions)
                    }

                    // Actions
                    VStack(spacing: 12) {
                        Button(bot.isActive ? "Disable Bot" : "Enable Bot") {
                            // TODO: Toggle bot
                        }
                        .font(.custom("Inter", size: 14)).fontWeight(.bold)
                        .foregroundStyle(bot.isActive ? Color.Echo.error : Color.Echo.primaryContainer)

                        Button("Remove Bot") {
                            // TODO: Remove bot
                        }
                        .font(.custom("Inter", size: 14)).fontWeight(.bold)
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
    let lastTriggered: String?
    let permissions: String
}

// MARK: - Bot Management ViewModel

@MainActor
class BotManagementViewModel: ObservableObject {
    @Published var activeBots: [BotInfo] = []
    @Published var availableBots: [BotInfo] = []
    @Published var selectedBot: BotInfo?
    @Published var showBotDetail = false

    func loadBots() async {
        // TODO: Load from bot service
    }

    func addBot(_ bot: BotInfo) async {
        // TODO: Add bot to active list
    }
}
