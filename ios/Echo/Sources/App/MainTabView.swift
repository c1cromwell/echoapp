#if os(iOS)
// App/MainTabView.swift
// Main tab bar: Messages · Contacts · Rewards · Profile (see docs/design-previews/bottomfooter.png).
// Personas are managed in the Profile screen, not a tab. Wallet/staking lives under Rewards.

import SwiftUI

// MARK: - Tab bar visibility (hide on pushed chat — see docs/design-previews/phaseA-chat.html)

private struct TabBarHiddenPreferenceKey: PreferenceKey {
    static var defaultValue = false
    static func reduce(value: inout Bool, nextValue: () -> Bool) {
        value = value || nextValue()
    }
}

extension View {
    /// Hides `GlacialTabBar` while a detail screen (e.g. 1:1 chat) is on the Messages stack.
    func hidesGlacialTabBarWhenPushed(_ hidden: Bool) -> some View {
        preference(key: TabBarHiddenPreferenceKey.self, value: hidden)
    }
}

struct MainTabView: View {
    @State private var selectedTab: MainTab = .messages
    @State private var hideTabBar = false

    enum MainTab: String {
        case messages, contacts, rewards, me
    }

    var body: some View {
        ZStack(alignment: .bottom) {
            TabView(selection: $selectedTab) {
                NavigationStack {
                    MessagesTabView()
                }
                .tag(MainTab.messages)

                NavigationStack {
                    ContactsListView()
                }
                .tag(MainTab.contacts)

                NavigationStack {
                    RewardsDashboardView()
                }
                .tag(MainTab.rewards)

                NavigationStack {
                    ProfileTabView()
                }
                .tag(MainTab.me)
            }
            .labelsHidden()

            if !hideTabBar {
                GlacialTabBar(selectedTab: $selectedTab)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
            }
        }
        .animation(.easeInOut(duration: 0.2), value: hideTabBar)
        .onPreferenceChange(TabBarHiddenPreferenceKey.self) { hideTabBar = $0 }
    }
}

// MARK: - Glacial Tab Bar (Messages · Contacts · Rewards · Profile — see docs/design-previews/bottomfooter.png)

struct GlacialTabBar: View {
    @Binding var selectedTab: MainTabView.MainTab

    var body: some View {
        HStack {
            tabButton(.messages, icon: "message", label: "Messages")
            tabButton(.contacts, icon: "person.2", label: "Contacts")
            tabButton(.rewards, icon: "star", label: "Rewards")
            tabButton(.me, icon: "person", label: "Profile")
        }
        .padding(.horizontal, 8)
        .padding(.bottom, 20) // safe area
        .frame(height: 82)
        .background(
            Color.Echo.surfaceContainerLowest
                .shadow(color: Color.Echo.onSurface.opacity(0.04), radius: 8, y: -4)
        )
    }

    /// `icon` is the SF Symbol base name; the filled variant is shown when selected.
    private func tabButton(_ tab: MainTabView.MainTab, icon: String, label: String) -> some View {
        let isSelected = selectedTab == tab
        return Button {
            selectedTab = tab
        } label: {
            VStack(spacing: 4) {
                ZStack(alignment: .topTrailing) {
                    Image(systemName: isSelected ? "\(icon).fill" : icon)
                        .font(.system(size: 24, weight: isSelected ? .semibold : .regular))
                        .foregroundStyle(isSelected ? Color.Echo.primaryContainer : Color.Echo.outline)

                    // Unread badge (Messages only)
                    if tab == .messages {
                        Circle()
                            .fill(Color.Echo.error)
                            .frame(width: 8, height: 8)
                            .offset(x: 4, y: -2)
                    }
                }

                Text(label)
                    .font(.system(size: 10))
                    .fontWeight(isSelected ? .semibold : .medium)
                    .foregroundStyle(isSelected ? Color.Echo.primaryContainer : Color.Echo.outline)
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

#endif
