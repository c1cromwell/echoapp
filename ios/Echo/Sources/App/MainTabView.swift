// App/MainTabView.swift
// Main tab view with 3 tabs: Messages, Wallet, Me
// Per Tokenomics v2 §4.2 — restructured from 5 tabs

import SwiftUI

struct MainTabView: View {
    @State private var selectedTab: MainTab = .messages

    enum MainTab: String {
        case messages, wallet, me
    }

    var body: some View {
        ZStack(alignment: .bottom) {
            TabView(selection: $selectedTab) {
                NavigationStack {
                    ConversationListView()
                }
                .tag(MainTab.messages)

                NavigationStack {
                    WalletTab(api: WalletAPIClientStub())
                }
                .tag(MainTab.wallet)

                NavigationStack {
                    ProfileTabView()
                }
                .tag(MainTab.me)
            }
            .labelsHidden()

            // Custom tab bar (tonal shift, no border-top)
            GlacialTabBar(selectedTab: $selectedTab)
        }
    }
}

// MARK: - Glacial Tab Bar (3-tab design)

struct GlacialTabBar: View {
    @Binding var selectedTab: MainTabView.MainTab

    var body: some View {
        HStack {
            tabButton(.messages, icon: "message.fill", label: "Messages")
            tabButton(.wallet, icon: "wallet.pass.fill", label: "Wallet")
            tabButton(.me, icon: "person.fill", label: "Me")
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 20) // safe area
        .frame(height: 82)
        .background(
            Color.Echo.surfaceContainerLowest
                .shadow(color: Color.Echo.onSurface.opacity(0.04), radius: 8, y: -4)
        )
    }

    private func tabButton(_ tab: MainTabView.MainTab, icon: String, label: String) -> some View {
        Button {
            selectedTab = tab
        } label: {
            VStack(spacing: 4) {
                ZStack(alignment: .topTrailing) {
                    Image(systemName: icon)
                        .font(.system(size: 24, weight: selectedTab == tab ? .semibold : .regular))
                        .foregroundStyle(selectedTab == tab ? Color.Echo.primaryContainer : Color.Echo.outline)

                    // Unread badge (Messages only)
                    if tab == .messages {
                        Circle()
                            .fill(Color.Echo.error)
                            .frame(width: 8, height: 8)
                            .offset(x: 4, y: -2)
                    }
                }

                Text(label)
                    .font(.custom("Inter", size: 10))
                    .fontWeight(selectedTab == tab ? .semibold : .medium)
                    .foregroundStyle(selectedTab == tab ? Color.Echo.primaryContainer : Color.Echo.outline)
            }
            .frame(maxWidth: .infinity)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(label)
    }
}

// MARK: - Placeholder for WalletAPIClient

struct WalletAPIClientStub: WalletAPIClient {
    func createWallet() async throws -> WalletInfo { WalletInfo(address: "", publicKey: "") }
    func importWallet(mnemonic: String) async throws -> WalletInfo { WalletInfo(address: "", publicKey: "") }
    func getBalance() async throws -> BalanceInfo { BalanceInfo(total: 0, available: 0) }
    func getTokenLocks() async throws -> [TokenLockPosition] { [] }
    func getDelegations() async throws -> [DelegationPosition] { [] }
    func getValidators() async throws -> [ValidatorInfo] { [] }
    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String { "" }
    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String { "" }
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String { "" }
    func submitRewardClaim(rewardTypes: [String]) async throws -> String { "" }
}
