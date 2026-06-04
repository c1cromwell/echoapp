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
            ZStack {
                MessagesTabView()
                    .opacity(selectedTab == .messages ? 1 : 0)
                    .zIndex(selectedTab == .messages ? 1 : 0)

                NavigationStack {
                    ContactsListView()
                }
                .opacity(selectedTab == .contacts ? 1 : 0)
                .zIndex(selectedTab == .contacts ? 1 : 0)

                NavigationStack {
                    RewardsDashboardView()
                }
                .opacity(selectedTab == .rewards ? 1 : 0)
                .zIndex(selectedTab == .rewards ? 1 : 0)

                NavigationStack {
                    ProfileTabView()
                }
                .opacity(selectedTab == .me ? 1 : 0)
                .zIndex(selectedTab == .me ? 1 : 0)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .safeAreaPadding(.bottom, hideTabBar ? 0 : 66)

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
        HStack(spacing: 0) {
            tabButton(.messages, icon: "ellipsis.message", label: "Messages")
            tabButton(.contacts, icon: "person.2", label: "Contacts")
            tabButton(.rewards, icon: "star", label: "Rewards")
            tabButton(.me, icon: "person.crop.circle", label: "Profile")
        }
        .padding(.top, 10)
        .padding(.bottom, 28)
        .background(
            Color.echoPaper
                .shadow(color: Color.black.opacity(0.06), radius: 12, y: -4)
                .ignoresSafeArea(edges: .bottom)
        )
    }

    private func tabButton(_ tab: MainTabView.MainTab, icon: String, label: String) -> some View {
        let isSelected = selectedTab == tab
        return Button {
            selectedTab = tab
        } label: {
            VStack(spacing: 4) {
                ZStack(alignment: .topTrailing) {
                    Image(systemName: isSelected ? "\(icon).fill" : icon)
                        .font(.system(size: 22))
                        .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)

                    if tab == .messages {
                        Circle()
                            .fill(Color.red)
                            .frame(width: 7, height: 7)
                            .offset(x: 4, y: -2)
                    }
                }

                Text(label)
                    .font(.system(size: 10, weight: isSelected ? .semibold : .medium))
                    .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)
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
