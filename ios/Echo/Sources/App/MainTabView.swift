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
    @Environment(AppState.self) private var appState
    @State private var selectedTab: MainTab = .messages
    @State private var hideTabBar = false

    enum MainTab: String {
        case messages, contacts, rewards, settings
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
                    SettingsView(
                        onSignOut: {
                            Task { await appState.signOut() }
                        },
                        onNewAccountSetup: {
                            Task { await appState.signOutForNewAccountSetup() }
                        }
                    )
                }
                .opacity(selectedTab == .settings ? 1 : 0)
                .zIndex(selectedTab == .settings ? 1 : 0)
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
        .safeAreaInset(edge: .top, spacing: 0) {
            VStack(spacing: 0) {
                NetworkStatusBanner()
                ProvisioningStatusBanner()
            }
        }
        .echoToastOverlay()
    }
}

// MARK: - Glacial Tab Bar (Messages · Contacts · Rewards · Profile — see docs/design-previews/bottomfooter.png)

struct GlacialTabBar: View {
    @Binding var selectedTab: MainTabView.MainTab
    var unreadMessages: Int = 0

    var body: some View {
        HStack(spacing: 0) {
            tabButton(.messages, icon: "ellipsis.message", label: "Messages", badge: unreadMessages)
            tabButton(.contacts, icon: "person.2", label: "Contacts")
            tabButton(.rewards, icon: "gift", label: "Rewards")
            tabButton(.settings, icon: "gearshape", label: "Settings")
        }
        .padding(.top, 10)
        .padding(.bottom, 28)
        .background(
            Color.echoPaper
                .shadow(color: Color.black.opacity(0.06), radius: 12, y: -4)
                .ignoresSafeArea(edges: .bottom)
        )
    }

    private func tabButton(_ tab: MainTabView.MainTab, icon: String, label: String, badge: Int = 0) -> some View {
        let isSelected = selectedTab == tab
        return Button {
            if selectedTab != tab {
                HapticManager.selection()
            }
            selectedTab = tab
        } label: {
            VStack(spacing: 4) {
                ZStack(alignment: .topTrailing) {
                    Image(systemName: isSelected ? "\(icon).fill" : icon)
                        .font(.system(size: 22))
                        .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)

                    if badge > 0 {
                        Text(badge > 9 ? "9+" : "\(badge)")
                            .font(.system(size: 9, weight: .bold))
                            .foregroundColor(.white)
                            .frame(minWidth: 16, minHeight: 16)
                            .background(Color.echoAlert)
                            .clipShape(Capsule())
                            .offset(x: 6, y: -4)
                    }
                }

                Text(label)
                    .font(.system(size: 10, weight: isSelected ? .semibold : .medium))
                    .foregroundStyle(isSelected ? Color.echoPrimary : Color.echoInk40)
            }
            .frame(maxWidth: .infinity)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(badge > 0 ? "\(label), \(badge) unread" : label)
    }
}

// MARK: - Placeholder for WalletAPIClient

struct WalletAPIClientStub: WalletAPIClient {
    func fetchWalletState() async throws -> WalletState {
        WalletState(
            totalBalance: 0, available: 0, staked: 0, pendingRewards: 0,
            locks: [], delegations: [], dailyRewards: nil, vesting: nil
        )
    }
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
    func fetchEmissionStatus() async throws -> EmissionStatus {
        EmissionStatus(currentYear: 1, annualCap: 0, distributedToDate: 0, remainingBudget: 0, percentConsumed: 0)
    }
    func fetchFounderVesting() async throws -> FounderVestingProfile? { nil }
}

#endif
