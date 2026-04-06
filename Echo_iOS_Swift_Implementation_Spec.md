# ECHO iOS Swift Implementation Spec
## Missing Screens & Design Fixes | v4.2.1 | April 2026

---

## Aligned To

| Document | Version |
|----------|---------|
| PRD | v2.5.1 |
| Data Layer Architecture | v3.4 |
| iOS Frontend Architecture | v3.2 |
| ECHO Tokenomics | v2.0 |
| DESIGN.md | v1.0 — "The Glacial Interface" |
| Figma Implementation Guide | v3.2 |
| IOS_IMPL | v4.2 |

---

## Table of Contents

1. [Audit Summary & Gap Analysis](#1-audit-summary--gap-analysis)
2. [Project Structure Updates](#2-project-structure-updates)
3. [Design System Enforcement](#3-design-system-enforcement)
4. [Tab Bar Restructure](#4-tab-bar-restructure)
5. [New Screen: Wallet Tab](#5-wallet-tab)
6. [New Screen: Voice/Video Call](#6-voicevideo-call)
7. [New Screen: Contact Detail](#7-contact-detail)
8. [New Screen: Advanced Search](#8-advanced-search)
9. [New Screen: Media Gallery](#9-media-gallery)
10. [New Screen: Notification Center](#10-notification-center)
11. [New Screen: QR Identity Share](#11-qr-identity-share)
12. [New Screen: Governance/Voting](#12-governancevoting)
13. [New Screen: Backup & Export](#13-backup--export)
14. [New Screen: Enterprise Profile](#14-enterprise-profile)
15. [New Screen: Bot Management](#15-bot-management)
16. [New Screen: Staking/Delegation Detail](#16-stakingdelegation-detail)
17. [Profile Page Expansion](#17-profile-page-expansion)
18. [Existing Screen Fixes](#18-existing-screen-fixes)
19. [Navigation & Routing](#19-navigation--routing)
20. [Data Models](#20-data-models)
21. [Build Phases](#21-build-phases)

---

## 1. Audit Summary & Gap Analysis

### Current Figma Make Codebase

The existing React/TypeScript Figma Make prototype contains 23 app screens and 15 onboarding screens. The following analysis compares these against the PRD v2.5.1 feature list and IOS_IMPL v4.2 project structure.

### Screens Present (Mapped to Swift Files)

| Figma Make Screen | iOS Swift File | Status |
|-------------------|----------------|--------|
| `intro-carousel.tsx` | `OnboardingFlow.swift` | Exists in IOS_IMPL |
| `welcome.tsx` | `WelcomePage.swift` | Exists |
| `login.tsx` | `LoginScreen.swift` | Exists — detailed Glacial spec |
| `phone-entry.tsx` | `SMSVerificationScreen.swift` | Exists |
| `otp-verify.tsx` | `SMSVerificationScreen.swift` | Exists |
| `passkey.tsx` | `PasskeyAssertionDelegate.swift` | Exists |
| `recovery.tsx` | `RecoveryPhraseView.swift` | Exists |
| `profile-setup.tsx` | `ProfileSetupView.swift` | Exists |
| `verify-credentials.tsx` | `CredentialVerificationView.swift` | Exists |
| `trust-dashboard.tsx` (onboarding) | `TrustOnboardingView.swift` | Exists |
| `messages.tsx` | `ConversationListView.swift` | Exists |
| `chat.tsx` | `ChatView.swift` | Exists |
| `group-chat.tsx` | `GroupChatView.swift` | Exists |
| `contacts.tsx` | `ContactsView.swift` | Exists |
| `groups/*.tsx` (3 screens) | `GroupViews/` | Exists |
| `channels/*.tsx` (3 screens) | `ChannelViews/` | Exists |
| `personas/*.tsx` (3 screens) | `PersonaViews/` | Exists |
| `hidden-folders/*.tsx` (3 screens) | `HiddenFolderViews/` | Exists |
| `trust.tsx` | `TrustDashboardView.swift` | Exists |
| `profile.tsx` | `ProfileView.swift` | Exists — needs expansion |
| `settings.tsx` | `SettingsView.swift` | Exists |
| `rewards.tsx` | **REPLACE** with `WalletTab.swift` | Gap |
| `device-management.tsx` | `DeviceManagementView.swift` | Exists |
| `login-audit.tsx` | `LoginAuditView.swift` | Exists |
| `analytics.tsx` | `AnalyticsView.swift` | Exists |

### Screens Missing (12 new Swift files required)

| Priority | Screen | Swift File | Feature Source |
|----------|--------|------------|----------------|
| P0 | Wallet Dashboard | `WalletTab.swift` | Tokenomics v2 §4.2 |
| P0 | Voice/Video Call | `CallView.swift` | PRD §Voice/Video Calls |
| P0 | Contact Detail | `ContactDetailView.swift` | PRD §Dynamic Trust Network |
| P1 | Advanced Search | `SearchView.swift` | PRD §Advanced Message Search |
| P1 | Media Gallery | `MediaGalleryView.swift` | PRD §Large File Sharing |
| P1 | Notification Center | `NotificationCenterView.swift` | Blueprint §Notifications |
| P1 | QR Identity Share | `QRIdentityView.swift` | PRD §Decentralized Identity |
| P2 | Governance/Voting | `GovernanceView.swift` | Tokenomics §3.3 |
| P2 | Backup & Export | `BackupView.swift` | PRD §Recovery |
| P2 | Enterprise Profile | `EnterpriseProfileView.swift` | PRD §Enterprise Profiles |
| P3 | Bot Management | `BotManagementView.swift` | PRD §Bot Framework |
| P3 | Staking Detail | `StakingDetailView.swift` | Tokenomics §4.3 |

---

## 2. Project Structure Updates

Add the following files to the existing IOS_IMPL v4.2 project structure:

```
Echo/
├── Features/
│   ├── ... (existing features unchanged)
│   │
│   ├── Wallet/                          # ★ UPDATED — was partially spec'd
│   │   ├── WalletCoordinator.swift
│   │   ├── WalletTab.swift              # ★ Replaces RewardsView
│   │   ├── WalletViewModel.swift
│   │   ├── StakingDetailView.swift      # ★ NEW
│   │   ├── StakingDetailViewModel.swift # ★ NEW
│   │   ├── DelegationView.swift         # ★ NEW
│   │   ├── Components/
│   │   │   ├── BalanceCard.swift
│   │   │   ├── BalanceBreakdown.swift
│   │   │   ├── WalletActionButton.swift
│   │   │   ├── DailyRewardsSection.swift
│   │   │   ├── FounderVestingSection.swift
│   │   │   ├── RecentActivityList.swift
│   │   │   └── StakingTierCard.swift    # ★ NEW
│   │   └── Models/
│   │       ├── WalletActivity.swift
│   │       ├── DailyRewards.swift
│   │       ├── StakingPosition.swift    # ★ NEW
│   │       └── VestingInfo.swift
│   │
│   ├── Calling/                         # ★ NEW — entire feature
│   │   ├── CallCoordinator.swift
│   │   ├── CallView.swift
│   │   ├── CallViewModel.swift
│   │   ├── IncomingCallView.swift
│   │   ├── GroupCallView.swift
│   │   ├── Components/
│   │   │   ├── CallControlsBar.swift
│   │   │   ├── CallTimerView.swift
│   │   │   ├── SelfPreviewPiP.swift
│   │   │   ├── ParticipantGrid.swift
│   │   │   └── ScreenShareBanner.swift
│   │   └── Models/
│   │       ├── CallState.swift
│   │       ├── CallParticipant.swift
│   │       └── CallType.swift
│   │
│   ├── Contacts/                        # ★ EXPANDED
│   │   ├── ... (existing files)
│   │   ├── ContactDetailView.swift      # ★ NEW
│   │   └── ContactDetailViewModel.swift # ★ NEW
│   │
│   ├── Search/                          # ★ NEW — entire feature
│   │   ├── SearchView.swift
│   │   ├── SearchViewModel.swift
│   │   ├── SearchResultItem.swift
│   │   ├── SearchFilterSheet.swift
│   │   └── Models/
│   │       ├── SearchResult.swift
│   │       └── SearchFilter.swift
│   │
│   ├── Media/                           # ★ NEW
│   │   ├── MediaGalleryView.swift
│   │   ├── MediaGalleryViewModel.swift
│   │   ├── MediaDetailView.swift
│   │   └── Models/
│   │       └── SharedMedia.swift
│   │
│   ├── Notifications/                   # ★ NEW
│   │   ├── NotificationCenterView.swift
│   │   ├── NotificationViewModel.swift
│   │   └── Models/
│   │       ├── AppNotification.swift
│   │       └── NotificationCategory.swift
│   │
│   ├── Identity/                        # ★ EXPANDED
│   │   ├── ... (existing files)
│   │   ├── QRIdentityView.swift         # ★ NEW
│   │   ├── QRScannerView.swift          # ★ NEW
│   │   └── QRIdentityViewModel.swift    # ★ NEW
│   │
│   ├── Governance/                      # ★ NEW — entire feature
│   │   ├── GovernanceView.swift
│   │   ├── GovernanceViewModel.swift
│   │   ├── ProposalDetailView.swift
│   │   ├── VoteConfirmationSheet.swift
│   │   └── Models/
│   │       ├── Proposal.swift
│   │       └── VotingPower.swift
│   │
│   ├── Backup/                          # ★ NEW
│   │   ├── BackupView.swift
│   │   ├── BackupViewModel.swift
│   │   ├── RecoveryPhraseDisplayView.swift
│   │   └── ExportDataSheet.swift
│   │
│   ├── Enterprise/                      # ★ NEW
│   │   ├── EnterpriseProfileView.swift
│   │   └── EnterpriseProfileViewModel.swift
│   │
│   └── Bots/                            # ★ NEW
│       ├── BotManagementView.swift
│       ├── BotManagementViewModel.swift
│       ├── BotDetailView.swift
│       └── Models/
│           └── Bot.swift
```

---

## 3. Design System Enforcement

All new screens MUST follow the Glacial Interface. This section provides the Swift implementation of design tokens that every screen imports.

### GlacialTheme.swift — Color Tokens

```swift
// Core/DesignSystem/GlacialTheme.swift

import SwiftUI

extension Color {
    enum Echo {
        // Primary
        static let primary = Color(hex: "006591")
        static let primaryContainer = Color(hex: "0EA5E9")  // Sky Blue
        static let deepNavy = Color(hex: "0F172A")
        static let onSurface = Color(hex: "131B2E")  // NEVER use Color.black
        
        // Surface Tiers — Light
        static let surface = Color(hex: "FAF8FF")
        static let surfaceContainerLowest = Color.white
        static let surfaceContainerLow = Color(hex: "F1F5F9")
        static let surfaceContainer = Color(hex: "E2E8F0")
        static let surfaceContainerHigh = Color(hex: "C8D6E5")
        static let surfaceContainerHighest = Color(hex: "DAE2FD")
        static let surfaceBright = Color(hex: "F8FAFC")  // hover state
        
        // Outline
        static let outline = Color(hex: "64748B")
        static let outlineVariant = Color(hex: "BEC8D2")  // ghost borders at 15%
        
        // Semantic
        static let success = Color(hex: "10B981")
        static let warning = Color(hex: "F59E0B")
        static let error = Color(hex: "EF4444")
        
        // Sky accent
        static let skyLight = Color(hex: "7DD3FC")
        static let skyGlow = Color(hex: "0EA5E9").opacity(0.25)
        static let skySubtle = Color(hex: "0EA5E9").opacity(0.08)
        static let skyRing = Color(hex: "0EA5E9").opacity(0.2)
        
        // Trust Tiers
        static let tierInnerCircle = Color(hex: "0EA5E9")
        static let tierTrusted = Color(hex: "10B981")
        static let tierVerified = Color(hex: "F59E0B")
        static let tierPeer = Color(hex: "64748B")
        static let tierBasic = Color(hex: "CBD5E1")
    }
}

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: .alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let r = Double((int >> 16) & 0xFF) / 255.0
        let g = Double((int >> 8) & 0xFF) / 255.0
        let b = Double(int & 0xFF) / 255.0
        self.init(red: r, green: g, blue: b)
    }
}
```

### Signature Gradient

```swift
extension LinearGradient {
    /// 135° Deep Navy → Sky Blue — used for primary CTAs and outbound messages
    static let signature = LinearGradient(
        colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )
}
```

### Typography Scale

```swift
extension Font {
    enum Echo {
        static let displayLG = Font.custom("Inter", size: 56).weight(.bold)    // "Hello" moments
        static let headlineSM = Font.custom("Inter", size: 24).weight(.bold)   // Screen titles
        static let titleMD = Font.custom("Inter", size: 20).weight(.semibold)  // Section headers
        static let bodyLG = Font.custom("Inter", size: 16).weight(.regular)    // Chat, body
        static let bodyMD = Font.custom("Inter", size: 14).weight(.regular)    // Secondary
        static let labelMD = Font.custom("Inter", size: 12).weight(.medium)    // Timestamps
        static let labelSM = Font.custom("Inter", size: 10).weight(.bold)      // Micro labels
    }
}
```

### View Modifiers

```swift
// GhostBorder — "glint on glass" at 15% opacity
struct GhostBorder: ViewModifier {
    var radius: CGFloat = 32
    func body(content: Content) -> some View {
        content.overlay(
            RoundedRectangle(cornerRadius: radius)
                .stroke(Color.Echo.outlineVariant.opacity(0.15), lineWidth: 1)
        )
    }
}

// GlacialShadow — tinted, never grey
struct GlacialShadow: ViewModifier {
    var radius: CGFloat = 24
    var opacity: Double = 0.04
    func body(content: Content) -> some View {
        content.shadow(
            color: Color.Echo.onSurface.opacity(opacity),
            radius: radius, x: 0, y: 4
        )
    }
}

// DeepGlacialShadow — for primary CTAs
struct DeepGlacialShadow: ViewModifier {
    func body(content: Content) -> some View {
        content.shadow(
            color: Color.Echo.deepNavy.opacity(0.15),
            radius: 32, x: 0, y: 12
        )
    }
}

// IcyBackground — blurred gradient orbs
struct IcyBackground: ViewModifier {
    func body(content: Content) -> some View {
        content.background(
            ZStack {
                Color.Echo.surface
                Circle()
                    .fill(Color.Echo.primaryContainer.opacity(0.10))
                    .frame(width: 300, height: 300)
                    .blur(radius: 120)
                    .offset(x: -100, y: -200)
                Circle()
                    .fill(Color.Echo.surfaceContainerHigh.opacity(0.10))
                    .frame(width: 250, height: 250)
                    .blur(radius: 100)
                    .offset(x: 120, y: 300)
            }
        )
    }
}

extension View {
    func ghostBorder(radius: CGFloat = 32) -> some View { modifier(GhostBorder(radius: radius)) }
    func glacialShadow(radius: CGFloat = 24) -> some View { modifier(GlacialShadow(radius: radius)) }
    func deepGlacialShadow() -> some View { modifier(DeepGlacialShadow()) }
    func icyBackground() -> some View { modifier(IcyBackground()) }
}
```

### SecureThreadIndicator

```swift
// Present on ALL authenticated screens — 2px sky-blue pulsating line
struct SecureThreadIndicator: View {
    @State private var opacity: Double = 0.6
    
    var body: some View {
        Rectangle()
            .fill(Color.Echo.primaryContainer)
            .frame(height: 2)
            .opacity(opacity)
            .onAppear {
                withAnimation(.easeInOut(duration: 2).repeatForever(autoreverses: true)) {
                    opacity = 1.0
                }
            }
    }
}
```

---

## 4. Tab Bar Restructure

### Current (Figma Make): 4 tabs
```
Messages | Contacts | Rewards | Profile
```

### Required (per Tokenomics v2 §4.2): 3 tabs
```
Messages | Wallet | Me
```

### Implementation

```swift
// App/EchoApp.swift

import SwiftUI

@main
struct EchoApp: App {
    @StateObject private var appState = AppState()
    
    var body: some Scene {
        WindowGroup {
            if appState.isAuthenticated {
                MainTabView()
                    .environmentObject(appState)
            } else {
                AuthCoordinator()
                    .environmentObject(appState)
            }
        }
    }
}

// App/MainTabView.swift

struct MainTabView: View {
    @State private var selectedTab: Tab = .messages
    
    enum Tab: String {
        case messages, wallet, me
    }
    
    var body: some View {
        ZStack(alignment: .bottom) {
            TabView(selection: $selectedTab) {
                NavigationStack {
                    ConversationListView()
                }
                .tag(Tab.messages)
                
                NavigationStack {
                    WalletTab()
                }
                .tag(Tab.wallet)
                
                NavigationStack {
                    ProfileView()
                }
                .tag(Tab.me)
            }
            .labelsHidden()
            
            // Custom tab bar (no border-top — tonal shift only)
            EchoTabBar(selectedTab: $selectedTab)
        }
    }
}

struct EchoTabBar: View {
    @Binding var selectedTab: MainTabView.Tab
    
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
            // Tonal shift instead of border-top (Glacial Interface rule)
            Color.Echo.surfaceContainerLowest
                .shadow(color: Color.Echo.onSurface.opacity(0.04), radius: 8, y: -4)
        )
    }
    
    private func tabButton(_ tab: MainTabView.Tab, icon: String, label: String) -> some View {
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
    }
}
```

**Key Changes from Current:**
- Removed `Contacts` tab (accessible via Messages header icon and Settings)
- Removed `Rewards` tab (replaced by `Wallet` with full Stargazer SDK integration)
- Renamed `Profile` to `Me` (matches architecture docs)
- Tab bar background uses tonal shift (surfaceContainerLowest) instead of `border-top`
- Uses tinted ambient shadow instead of divider line

---

## 5. Wallet Tab

**Replaces:** `rewards.tsx` (324 lines)  
**Source:** Tokenomics v2 §4.2-4.3, IOS_IMPL v4.2 §Wallet  
**Priority:** P0

### WalletTab.swift

```swift
import SwiftUI
import StargazerSDK

struct WalletTab: View {
    @StateObject private var viewModel = WalletViewModel()
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Balance Card — Signature gradient
                BalanceCard(
                    total: viewModel.totalBalance,
                    usdValue: viewModel.usdValue,
                    change24h: viewModel.change24h
                )
                
                // Balance Breakdown — Ghost border card
                BalanceBreakdown(
                    available: viewModel.available,
                    staked: viewModel.staked,
                    stakingTier: viewModel.stakingTier,
                    delegatedTo: viewModel.delegatedValidator,
                    pending: viewModel.pendingRewards,
                    onClaimAll: { Task { await viewModel.claimAll() } }
                )
                
                // Action Buttons — 4-column grid
                HStack(spacing: 12) {
                    WalletActionButton(icon: "lock.fill", label: "Stake") {
                        viewModel.showStaking = true
                    }
                    WalletActionButton(icon: "arrow.up.right", label: "Delegate") {
                        viewModel.showDelegation = true
                    }
                    WalletActionButton(icon: "arrow.left.arrow.right", label: "Swap") {
                        viewModel.showSwap = true
                    }
                    WalletActionButton(icon: "link", label: "Bridge") {
                        viewModel.showBridge = true
                    }
                }
                
                // Daily Rewards — Progress bars
                DailyRewardsSection(rewards: viewModel.dailyRewards)
                
                // Founder Vesting — conditional
                if let vesting = viewModel.founderVesting {
                    FounderVestingSection(vesting: vesting)
                }
                
                // Recent Activity
                RecentActivityList(activity: viewModel.recentActivity)
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("ECHO Wallet")
        .navigationBarTitleDisplayMode(.large)
        .task { await viewModel.loadWallet() }
        .sheet(isPresented: $viewModel.showStaking) {
            StakingDetailView()
        }
    }
}
```

### BalanceCard.swift

```swift
struct BalanceCard: View {
    let total: Decimal
    let usdValue: Decimal
    let change24h: Double
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Total Balance")
                .font(.custom("Inter", size: 12))
                .fontWeight(.medium)
                .foregroundStyle(.white.opacity(0.7))
                .tracking(0.5)
            
            Text("\(total.formatted()) ECHO")
                .font(.custom("Inter", size: 32))
                .fontWeight(.bold)
                .foregroundStyle(.white)
            
            HStack(spacing: 8) {
                Text("≈ $\(usdValue.formatted()) USD")
                    .font(.custom("Inter", size: 14))
                    .foregroundStyle(.white.opacity(0.6))
                
                HStack(spacing: 2) {
                    Image(systemName: change24h >= 0 ? "arrowtriangle.up.fill" : "arrowtriangle.down.fill")
                        .font(.system(size: 8))
                    Text("\(abs(change24h), specifier: "%.1f")%")
                        .font(.custom("Inter", size: 12))
                        .fontWeight(.semibold)
                }
                .foregroundStyle(change24h >= 0 ? Color.Echo.success : Color.Echo.error)
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(.white.opacity(0.15))
                .clipShape(Capsule())
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 32)
                .fill(LinearGradient.signature)
        )
        .deepGlacialShadow()
    }
}
```

### WalletActionButton.swift

```swift
struct WalletActionButton: View {
    let icon: String
    let label: String
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.system(size: 22))
                    .foregroundStyle(Color.Echo.primaryContainer)
                
                Text(label)
                    .font(.custom("Inter", size: 11))
                    .fontWeight(.bold)
                    .foregroundStyle(Color.Echo.outline)
                    .textCase(.uppercase)
                    .tracking(0.5)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 72)
            .background(
                RoundedRectangle(cornerRadius: 20)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder(radius: 20)
        }
        .buttonStyle(SpringButtonStyle())
    }
}

/// Press state: scale(0.95) with spring animation
struct SpringButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
            .animation(.spring(response: 0.3, dampingFraction: 0.85), value: configuration.isPressed)
    }
}
```

### WalletViewModel.swift

```swift
@MainActor
class WalletViewModel: ObservableObject {
    private let stargazer: StargazerClient
    private let backendAPI: BackendAPIClient
    
    @Published var totalBalance: Decimal = 0
    @Published var available: Decimal = 0
    @Published var staked: Decimal = 0
    @Published var stakingTier: String = "None"
    @Published var pendingRewards: Decimal = 0
    @Published var usdValue: Decimal = 0
    @Published var change24h: Double = 0.0
    @Published var delegatedValidator: ValidatorInfo?
    @Published var founderVesting: VestingInfo?
    @Published var dailyRewards: DailyRewards = .empty
    @Published var recentActivity: [WalletActivity] = []
    
    @Published var showStaking = false
    @Published var showDelegation = false
    @Published var showSwap = false
    @Published var showBridge = false
    
    func loadWallet() async {
        // 1. Query balance from Stargazer SDK
        let balance = try? await stargazer.getBalance(token: .echo)
        self.totalBalance = balance?.total ?? 0
        self.available = balance?.available ?? 0
        
        // 2. Query TokenLock positions (staking)
        let locks = try? await stargazer.getTokenLocks(token: .echo)
        self.staked = locks?.reduce(0) { $0 + $1.amount } ?? 0
        self.stakingTier = StakingTier.from(amount: self.staked).displayName
        
        // 3. Query StakeDelegation positions
        let delegations = try? await stargazer.getDelegations(token: .echo)
        self.delegatedValidator = delegations?.first?.validator
        
        // 4. Query pending rewards from backend
        let rewards = try? await backendAPI.getPendingRewards()
        self.pendingRewards = rewards?.total ?? 0
        self.dailyRewards = rewards?.daily ?? .empty
        
        // 5. USD conversion
        let price = try? await backendAPI.getEchoPrice()
        self.usdValue = self.totalBalance * (price?.usd ?? 0)
        self.change24h = price?.change24h ?? 0
        
        // 6. Check founder vesting
        self.founderVesting = try? await backendAPI.getFounderVesting()
        
        // 7. Recent activity
        self.recentActivity = (try? await backendAPI.getWalletActivity(limit: 20)) ?? []
    }
    
    func claimAll() async {
        // Submit AllowSpend transaction to claim pending rewards
        try? await stargazer.submitTransaction(.claimRewards)
        await loadWallet()
    }
}
```

---

## 6. Voice/Video Call

**Source:** PRD §Voice and Video Calls, Blueprint §Call Establishment Flow  
**Priority:** P0

### CallView.swift

```swift
import SwiftUI
import WebRTC

struct CallView: View {
    @StateObject private var viewModel: CallViewModel
    @Environment(\.dismiss) private var dismiss
    
    init(contactId: String, callType: CallType) {
        _viewModel = StateObject(wrappedValue: CallViewModel(
            contactId: contactId, callType: callType
        ))
    }
    
    var body: some View {
        ZStack {
            // Background
            if viewModel.callType == .video && viewModel.state == .active {
                // Remote video fills screen
                VideoStreamView(stream: viewModel.remoteStream)
                    .ignoresSafeArea()
                
                // Self PiP — top left, draggable
                SelfPreviewPiP(stream: viewModel.localStream)
                    .frame(width: 100, height: 140)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                    .glacialShadow()
                    .position(x: 70, y: 120)
            } else {
                // Voice call — icy background
                Color.Echo.surface.icyBackground().ignoresSafeArea()
            }
            
            VStack(spacing: 0) {
                // Encryption badge — top
                EncryptionBadge()
                    .padding(.top, 60)
                
                Spacer()
                
                // Contact info — center (voice mode)
                if viewModel.callType == .voice || viewModel.state != .active {
                    VStack(spacing: 16) {
                        TrustRingAvatar(
                            imageURL: viewModel.contactAvatar,
                            trustTier: viewModel.contactTrustTier,
                            size: 80,
                            isAnimating: viewModel.state == .connecting
                        )
                        
                        Text(viewModel.contactName)
                            .font(.Echo.headlineSM)
                            .foregroundStyle(Color.Echo.onSurface)
                        
                        TrustTierPill(tier: viewModel.contactTrustTier)
                        
                        // State label
                        Text(viewModel.stateLabel)
                            .font(.custom("Inter", size: 32))
                            .fontWeight(.light)
                            .foregroundStyle(Color.Echo.outline)
                            .monospacedDigit()
                    }
                }
                
                Spacer()
                
                // Screen sharing banner
                if viewModel.isScreenSharing {
                    ScreenShareBanner(onStop: { viewModel.stopScreenShare() })
                }
                
                // Call controls
                CallControlsBar(
                    isMuted: viewModel.isMuted,
                    isSpeaker: viewModel.isSpeaker,
                    isCameraOn: viewModel.isCameraOn,
                    callType: viewModel.callType,
                    onToggleMute: { viewModel.toggleMute() },
                    onToggleSpeaker: { viewModel.toggleSpeaker() },
                    onToggleCamera: { viewModel.toggleCamera() },
                    onShareScreen: { viewModel.startScreenShare() },
                    onFlipCamera: { viewModel.flipCamera() },
                    onEndCall: {
                        viewModel.endCall()
                        dismiss()
                    }
                )
                .padding(.bottom, 40)
            }
        }
        .task { await viewModel.startCall() }
        .statusBarHidden()
    }
}
```

### CallControlsBar.swift

```swift
struct CallControlsBar: View {
    let isMuted: Bool
    let isSpeaker: Bool
    let isCameraOn: Bool
    let callType: CallType
    let onToggleMute: () -> Void
    let onToggleSpeaker: () -> Void
    let onToggleCamera: () -> Void
    let onShareScreen: () -> Void
    let onFlipCamera: () -> Void
    let onEndCall: () -> Void
    
    var body: some View {
        VStack(spacing: 24) {
            // Secondary controls
            HStack(spacing: 24) {
                CallControlButton(
                    icon: isMuted ? "mic.slash.fill" : "mic.fill",
                    isActive: !isMuted,
                    action: onToggleMute
                )
                
                CallControlButton(
                    icon: callType == .video
                        ? (isCameraOn ? "video.fill" : "video.slash.fill")
                        : (isSpeaker ? "speaker.wave.3.fill" : "speaker.fill"),
                    isActive: callType == .video ? isCameraOn : isSpeaker,
                    action: callType == .video ? onToggleCamera : onToggleSpeaker
                )
                
                CallControlButton(
                    icon: "rectangle.on.rectangle",
                    isActive: false,
                    action: onShareScreen
                )
                
                if callType == .video {
                    CallControlButton(
                        icon: "camera.rotate",
                        isActive: false,
                        action: onFlipCamera
                    )
                }
            }
            
            // End call button — 64px red circle
            Button(action: onEndCall) {
                Image(systemName: "phone.down.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(.white)
                    .frame(width: 64, height: 64)
                    .background(Circle().fill(Color.Echo.error))
                    .shadow(color: Color.Echo.error.opacity(0.4), radius: 16)
            }
            .buttonStyle(SpringButtonStyle())
        }
    }
}

struct CallControlButton: View {
    let icon: String
    let isActive: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Image(systemName: icon)
                .font(.system(size: 22))
                .foregroundStyle(.white)
                .frame(width: 56, height: 56)
                .background(
                    Circle()
                        .fill(.ultraThinMaterial)
                        .opacity(0.6)
                )
                .ghostBorder(radius: 28)
        }
    }
}
```

### CallState.swift

```swift
enum CallType: String, Codable {
    case voice, video
}

enum CallState: String {
    case idle, ringing, connecting, active, onHold, ended
}

struct CallParticipant: Identifiable {
    let id: String
    let name: String
    let avatarURL: URL?
    let trustTier: TrustTier
    let isMuted: Bool
    let isVideoOn: Bool
}
```

---

## 7. Contact Detail

**Source:** PRD §Dynamic Trust Network  
**Priority:** P0

### ContactDetailView.swift

```swift
struct ContactDetailView: View {
    @StateObject private var viewModel: ContactDetailViewModel
    @Environment(\.dismiss) private var dismiss
    
    init(contactId: String) {
        _viewModel = StateObject(wrappedValue: ContactDetailViewModel(contactId: contactId))
    }
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Hero — large avatar with trust ring
                VStack(spacing: 12) {
                    TrustRingAvatar(
                        imageURL: viewModel.contact.avatarURL,
                        trustTier: viewModel.contact.trustTier,
                        size: 140
                    )
                    
                    Text(viewModel.contact.name)
                        .font(.custom("Inter", size: 28))
                        .fontWeight(.heavy)
                        .tracking(-0.5)
                    
                    Text(viewModel.contact.echoHandle)
                        .font(.custom("Inter", size: 13))
                        .fontWeight(.semibold)
                        .foregroundStyle(Color.Echo.primaryContainer)
                    
                    HStack(spacing: 8) {
                        TrustTierPill(tier: viewModel.contact.trustTier)
                        
                        if viewModel.contact.isOnline {
                            HStack(spacing: 4) {
                                Circle().fill(Color.Echo.success).frame(width: 6, height: 6)
                                Text("Online").font(.Echo.labelMD).foregroundStyle(Color.Echo.success)
                            }
                        }
                    }
                }
                .padding(.top, 16)
                
                // Action buttons — 4-column
                HStack(spacing: 12) {
                    ContactActionButton(icon: "message.fill", label: "Message") {
                        // Navigate to chat
                    }
                    ContactActionButton(icon: "phone.fill", label: "Voice") {
                        viewModel.showVoiceCall = true
                    }
                    ContactActionButton(icon: "video.fill", label: "Video") {
                        viewModel.showVideoCall = true
                    }
                    ContactActionButton(icon: "magnifyingglass", label: "Search") {
                        viewModel.showSearch = true
                    }
                }
                .padding(.horizontal, 20)
                
                // Trust & Identity card
                GhostBorderSection(title: "TRUST & IDENTITY") {
                    TrustRow(label: "Trust Score", value: "\(viewModel.contact.trustScore)/100")
                    TrustRow(label: "DID", value: viewModel.contact.didShort, copyable: true)
                    TrustRow(label: "Verified Since", value: viewModel.contact.verifiedDate)
                    TrustRow(label: "Mutual Groups", value: "\(viewModel.contact.mutualGroups)")
                    TrustRow(label: "Mutual Contacts", value: "\(viewModel.contact.mutualContacts)")
                }
                
                // Credentials card
                GhostBorderSection(title: "CREDENTIALS") {
                    ForEach(viewModel.contact.credentials) { cred in
                        HStack(spacing: 12) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(Color.Echo.success)
                            Text(cred.name)
                                .font(.Echo.bodyMD)
                            Spacer()
                        }
                    }
                }
                
                // Shared Media preview (horizontal scroll)
                SharedMediaPreview(
                    media: viewModel.sharedMedia,
                    onSeeAll: { viewModel.showMediaGallery = true }
                )
                
                // Privacy settings for this contact
                GhostBorderSection(title: "PRIVACY FOR THIS CONTACT") {
                    SettingsRow(icon: "bell.fill", label: "Custom Notifications",
                               value: viewModel.notificationsEnabled ? "On" : "Off")
                    SettingsRow(icon: "timer", label: "Disappearing Messages",
                               value: viewModel.disappearingEnabled ? "On" : "Off")
                    
                    Divider().opacity(0) // spacing
                    
                    Button("Block Contact") {
                        viewModel.showBlockConfirmation = true
                    }
                    .font(.custom("Inter", size: 14)).fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.error)
                    
                    Button("Report Contact") {
                        viewModel.showReportSheet = true
                    }
                    .font(.custom("Inter", size: 14)).fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.error.opacity(0.7))
                }
            }
            .padding(.bottom, 40)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationBarBackButtonHidden()
        .toolbar {
            ToolbarItem(placement: .navigationBarLeading) {
                Button { dismiss() } label: {
                    Image(systemName: "arrow.left")
                        .foregroundStyle(Color.Echo.onSurface)
                }
            }
            ToolbarItem(placement: .navigationBarTrailing) {
                Menu {
                    Button("Copy DID") { viewModel.copyDID() }
                    Button("Share Contact") { viewModel.shareContact() }
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .foregroundStyle(Color.Echo.outline)
                }
            }
        }
        .task { await viewModel.loadContact() }
    }
}
```

### Helper Components

```swift
/// Reusable section wrapper with ghost border styling
struct GhostBorderSection<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content
    
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(title)
                .font(.custom("Inter", size: 10))
                .fontWeight(.bold)
                .tracking(2)
                .foregroundStyle(Color.Echo.outline)
                .padding(.leading, 8)
            
            VStack(alignment: .leading, spacing: 12) {
                content()
            }
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 32)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder()
        }
        .padding(.horizontal, 20)
    }
}

/// Contact action button (Message, Voice, Video, Search)
struct ContactActionButton: View {
    let icon: String
    let label: String
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 20))
                    .foregroundStyle(Color.Echo.primaryContainer)
                Text(label)
                    .font(.custom("Inter", size: 10))
                    .fontWeight(.bold)
                    .textCase(.uppercase)
                    .tracking(0.5)
                    .foregroundStyle(Color.Echo.outline)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 64)
            .background(
                RoundedRectangle(cornerRadius: 20)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder(radius: 20)
        }
        .buttonStyle(SpringButtonStyle())
    }
}
```

---

## 8. Advanced Search

**Source:** PRD §Advanced Message Search and Archive System  
**Priority:** P1

### SearchView.swift

```swift
struct SearchView: View {
    @StateObject private var viewModel = SearchViewModel()
    @FocusState private var isSearchFocused: Bool
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            
            // Search bar
            HStack(spacing: 12) {
                Button("Cancel") { dismiss() }
                    .font(.Echo.bodyMD)
                    .foregroundStyle(Color.Echo.primaryContainer)
                
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundStyle(Color.Echo.outline)
                    TextField("Search messages, files, links...", text: $viewModel.query)
                        .font(.Echo.bodyLG)
                        .focused($isSearchFocused)
                    if !viewModel.query.isEmpty {
                        Button { viewModel.query = "" } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(Color.Echo.outline)
                        }
                    }
                }
                .padding(12)
                .background(
                    RoundedRectangle(cornerRadius: 16)
                        .fill(Color.Echo.surfaceContainerLow)
                )
                .ghostBorder(radius: 16)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            
            // Filter tabs
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(SearchFilter.allCases, id: \.self) { filter in
                        FilterChip(
                            label: filter.displayName,
                            isSelected: viewModel.activeFilter == filter,
                            action: { viewModel.activeFilter = filter }
                        )
                    }
                }
                .padding(.horizontal, 16)
            }
            .padding(.vertical, 8)
            
            // Advanced filters (expandable)
            if viewModel.showAdvancedFilters {
                AdvancedFilterBar(
                    selectedContact: $viewModel.filterContact,
                    selectedDateRange: $viewModel.filterDateRange,
                    selectedChat: $viewModel.filterChat
                )
                .transition(.move(edge: .top).combined(with: .opacity))
            }
            
            // Results / Recent searches
            ScrollView {
                if viewModel.query.isEmpty {
                    // Recent searches
                    RecentSearchesSection(
                        searches: viewModel.recentSearches,
                        onTap: { viewModel.query = $0 },
                        onClear: { viewModel.clearRecentSearches() }
                    )
                } else {
                    // Search results
                    LazyVStack(spacing: 0) {
                        ForEach(viewModel.results) { result in
                            SearchResultRow(result: result)
                        }
                    }
                    
                    if viewModel.results.isEmpty && !viewModel.isSearching {
                        EmptySearchState()
                    }
                }
            }
        }
        .background(Color.Echo.surface)
        .onAppear { isSearchFocused = true }
    }
}
```

### SearchFilter.swift

```swift
enum SearchFilter: String, CaseIterable {
    case all, files, photos, links, voice
    
    var displayName: String {
        switch self {
        case .all: return "All"
        case .files: return "Files"
        case .photos: return "Photos"
        case .links: return "Links"
        case .voice: return "Voice"
        }
    }
}

struct SearchResult: Identifiable {
    let id: String
    let conversationId: String
    let contactName: String
    let contactAvatar: URL?
    let matchedText: String
    let highlightRange: Range<String.Index>?
    let timestamp: Date
    let messageType: MessageType
    let attachmentName: String?
    let attachmentSize: String?
}
```

---

## 9. Media Gallery

**Source:** PRD §Large File Sharing  
**Priority:** P1

### MediaGalleryView.swift

```swift
struct MediaGalleryView: View {
    @StateObject private var viewModel: MediaGalleryViewModel
    @State private var selectedTab: MediaTab = .photos
    
    enum MediaTab: CaseIterable {
        case photos, videos, files, links
        var label: String {
            switch self {
            case .photos: return "Photos"
            case .videos: return "Videos"
            case .files: return "Files"
            case .links: return "Links"
            }
        }
    }
    
    init(conversationId: String) {
        _viewModel = StateObject(wrappedValue: MediaGalleryViewModel(conversationId: conversationId))
    }
    
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            
            // Tab selector
            HStack(spacing: 0) {
                ForEach(MediaTab.allCases, id: \.self) { tab in
                    Button {
                        withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) {
                            selectedTab = tab
                        }
                    } label: {
                        Text(tab.label)
                            .font(.custom("Inter", size: 14))
                            .fontWeight(selectedTab == tab ? .bold : .medium)
                            .foregroundStyle(selectedTab == tab ? Color.Echo.primaryContainer : Color.Echo.outline)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 12)
                    }
                }
            }
            .background(Color.Echo.surfaceContainerLow)
            
            // Content
            switch selectedTab {
            case .photos, .videos:
                // Grid layout — 4 columns, 2px gap
                let columns = Array(repeating: GridItem(.flexible(), spacing: 2), count: 4)
                ScrollView {
                    LazyVGrid(columns: columns, spacing: 2) {
                        ForEach(viewModel.mediaItems(for: selectedTab)) { item in
                            MediaThumbnail(item: item)
                                .aspectRatio(1, contentMode: .fill)
                        }
                    }
                }
                
            case .files:
                // List view
                ScrollView {
                    LazyVStack(spacing: 0) {
                        ForEach(viewModel.files) { file in
                            FileRow(file: file)
                        }
                    }
                }
                
            case .links:
                // Rich preview cards
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(viewModel.links) { link in
                            LinkPreviewCard(link: link)
                                .padding(.horizontal, 16)
                        }
                    }
                    .padding(.top, 12)
                }
            }
        }
        .background(Color.Echo.surface)
        .navigationTitle("Shared Media")
        .task { await viewModel.loadMedia() }
    }
}
```

---

## 10. Notification Center

**Source:** Blueprint §Notification Badges  
**Priority:** P1

### NotificationCenterView.swift

```swift
struct NotificationCenterView: View {
    @StateObject private var viewModel = NotificationViewModel()
    
    var body: some View {
        ScrollView {
            LazyVStack(spacing: 0) {
                ForEach(viewModel.groupedNotifications.keys.sorted().reversed(), id: \.self) { date in
                    // Section header
                    Text(viewModel.sectionTitle(for: date))
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold)
                        .tracking(2)
                        .textCase(.uppercase)
                        .foregroundStyle(Color.Echo.outline)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, 20)
                        .padding(.top, 24)
                        .padding(.bottom, 8)
                    
                    ForEach(viewModel.groupedNotifications[date] ?? []) { notification in
                        NotificationRow(notification: notification)
                            .swipeActions(edge: .trailing) {
                                Button("Delete", role: .destructive) {
                                    viewModel.delete(notification)
                                }
                                Button("Mute") {
                                    viewModel.mute(notification)
                                }
                                .tint(Color.Echo.outline)
                            }
                    }
                }
            }
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Notifications")
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button("Mark All") {
                    viewModel.markAllRead()
                }
                .font(.Echo.bodyMD)
                .foregroundStyle(Color.Echo.primaryContainer)
            }
        }
        .task { await viewModel.loadNotifications() }
    }
}

struct NotificationRow: View {
    let notification: AppNotification
    
    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            // Category icon
            Circle()
                .fill(notification.category.color.opacity(0.15))
                .frame(width: 40, height: 40)
                .overlay(
                    Image(systemName: notification.category.icon)
                        .font(.system(size: 16))
                        .foregroundStyle(notification.category.color)
                )
            
            VStack(alignment: .leading, spacing: 4) {
                Text(notification.title)
                    .font(.custom("Inter", size: 14))
                    .fontWeight(notification.isRead ? .regular : .bold)
                    .foregroundStyle(Color.Echo.onSurface)
                
                if let subtitle = notification.subtitle {
                    Text(subtitle)
                        .font(.Echo.bodyMD)
                        .foregroundStyle(Color.Echo.outline)
                        .lineLimit(1)
                }
            }
            
            Spacer()
            
            Text(notification.timeAgo)
                .font(.Echo.labelMD)
                .foregroundStyle(Color.Echo.outline)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
        .background(
            notification.isRead
                ? Color.Echo.surface
                : Color.Echo.surfaceContainerLowest
        )
        .overlay(alignment: .leading) {
            if !notification.isRead {
                Rectangle()
                    .fill(Color.Echo.primaryContainer)
                    .frame(width: 3)
            }
        }
    }
}
```

### AppNotification.swift

```swift
struct AppNotification: Identifiable {
    let id: String
    let title: String
    let subtitle: String?
    let category: NotificationCategory
    let timestamp: Date
    var isRead: Bool
    let deepLink: String?
    
    var timeAgo: String {
        RelativeDateTimeFormatter().localizedString(for: timestamp, relativeTo: .now)
    }
}

enum NotificationCategory: String, CaseIterable {
    case message, call, group, channel, trust, wallet, system
    
    var icon: String {
        switch self {
        case .message: return "message.fill"
        case .call: return "phone.fill"
        case .group: return "person.3.fill"
        case .channel: return "megaphone.fill"
        case .trust: return "shield.checkered"
        case .wallet: return "trophy.fill"
        case .system: return "gear"
        }
    }
    
    var color: Color {
        switch self {
        case .message: return Color.Echo.primaryContainer
        case .call: return Color.Echo.success
        case .group: return Color(hex: "8B5CF6")
        case .channel: return Color.Echo.warning
        case .trust: return Color.Echo.primaryContainer
        case .wallet: return Color.Echo.success
        case .system: return Color.Echo.outline
        }
    }
}
```

---

## 11. QR Identity Share

**Source:** PRD §Decentralized Identity  
**Priority:** P1

### QRIdentityView.swift

```swift
import CoreImage.CIFilterBuiltins

struct QRIdentityView: View {
    @StateObject private var viewModel = QRIdentityViewModel()
    @State private var showScanner = false
    
    var body: some View {
        VStack(spacing: 32) {
            SecureThreadIndicator()
            
            Spacer()
            
            // QR Card — frosted glass
            VStack(spacing: 20) {
                // QR Code
                if let qrImage = viewModel.qrCodeImage {
                    Image(uiImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .frame(width: 200, height: 200)
                        .clipShape(RoundedRectangle(cornerRadius: 16))
                }
                
                // Identity info
                Text(viewModel.echoHandle)
                    .font(.custom("Inter", size: 16))
                    .fontWeight(.bold)
                    .foregroundStyle(Color.Echo.primaryContainer)
                
                Text(viewModel.didShort)
                    .font(.custom("Inter", size: 12))
                    .foregroundStyle(Color.Echo.outline)
                    .lineLimit(1)
                    .truncationMode(.middle)
                
                HStack(spacing: 6) {
                    Image(systemName: "lock.fill")
                        .font(.system(size: 10))
                    Text("Trust Score: \(viewModel.trustScore)/100")
                        .font(.Echo.labelMD)
                }
                .foregroundStyle(Color.Echo.outline)
            }
            .padding(32)
            .frame(maxWidth: .infinity)
            .background(
                RoundedRectangle(cornerRadius: 32)
                    .fill(.ultraThinMaterial)
                    .opacity(0.6)
            )
            .ghostBorder()
            .padding(.horizontal, 32)
            
            Spacer()
            
            // Action buttons
            HStack(spacing: 16) {
                Button {
                    showScanner = true
                } label: {
                    Label("Scan QR", systemImage: "camera.fill")
                        .font(.custom("Inter", size: 14))
                        .fontWeight(.bold)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 9999)
                                .fill(LinearGradient.signature)
                        )
                }
                .deepGlacialShadow()
                
                Button {
                    viewModel.shareLink()
                } label: {
                    Label("Share Link", systemImage: "square.and.arrow.up")
                        .font(.custom("Inter", size: 14))
                        .fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(
                            RoundedRectangle(cornerRadius: 9999)
                                .fill(Color.Echo.surfaceContainerLow)
                        )
                        .ghostBorder(radius: 9999)
                }
            }
            .padding(.horizontal, 32)
            
            Text("Scan another user's QR code to add them\nas a trusted contact with verified\nin-person attestation.")
                .font(.Echo.bodyMD)
                .foregroundStyle(Color.Echo.outline)
                .multilineTextAlignment(.center)
                .padding(.bottom, 40)
        }
        .icyBackground()
        .navigationTitle("My Identity")
        .fullScreenCover(isPresented: $showScanner) {
            QRScannerView(onScan: { did in
                showScanner = false
                viewModel.handleScannedDID(did)
            })
        }
    }
}
```

---

## 12. Governance/Voting

**Source:** Tokenomics v2 §3.3, IOS_IMPL v4.2 §7  
**Priority:** P2

### GovernanceView.swift

```swift
struct GovernanceView: View {
    @StateObject private var viewModel = GovernanceViewModel()
    
    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Voting power card
                VotingPowerCard(power: viewModel.votingPower)
                
                // Active proposals
                SectionLabel("ACTIVE PROPOSALS")
                ForEach(viewModel.activeProposals) { proposal in
                    ProposalCard(
                        proposal: proposal,
                        userVote: viewModel.userVotes[proposal.id],
                        onVote: { value in
                            viewModel.selectedProposal = proposal
                            viewModel.selectedVoteValue = value
                            viewModel.showVoteConfirmation = true
                        }
                    )
                }
                
                // Completed proposals
                SectionLabel("COMPLETED")
                ForEach(viewModel.completedProposals) { proposal in
                    CompletedProposalCard(proposal: proposal)
                }
            }
            .padding(.horizontal, 20)
            .padding(.top, 16)
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Governance")
        .sheet(isPresented: $viewModel.showVoteConfirmation) {
            if let proposal = viewModel.selectedProposal,
               let value = viewModel.selectedVoteValue {
                VoteConfirmationSheet(
                    proposal: proposal,
                    voteValue: value,
                    power: viewModel.votingPower,
                    onConfirm: {
                        Task { await viewModel.submitVote() }
                    }
                )
            }
        }
        .task { await viewModel.loadProposals() }
    }
}
```

### ProposalCard.swift

```swift
struct ProposalCard: View {
    let proposal: Proposal
    let userVote: String?
    let onVote: (String) -> Void
    
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Header
            HStack {
                Text(proposal.id)
                    .font(.Echo.labelSM)
                    .foregroundStyle(Color.Echo.outline)
                Spacer()
                Text("Ends in \(proposal.timeRemaining)")
                    .font(.Echo.labelMD)
                    .foregroundStyle(Color.Echo.outline)
            }
            
            Text(proposal.title)
                .font(.custom("Inter", size: 16))
                .fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)
            
            // Progress bar
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.Echo.surfaceContainer)
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.Echo.primaryContainer)
                        .frame(width: geo.size.width * proposal.forPercentage)
                }
            }
            .frame(height: 8)
            
            HStack {
                Text("\(Int(proposal.forPercentage * 100))% For")
                    .font(.Echo.labelMD)
                    .foregroundStyle(Color.Echo.primaryContainer)
                Spacer()
                Text("\(proposal.voterCount) voters")
                    .font(.Echo.labelMD)
                    .foregroundStyle(Color.Echo.outline)
            }
            
            // Vote buttons or status
            if let vote = userVote {
                HStack(spacing: 6) {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundStyle(Color.Echo.success)
                    Text("You voted: \(vote.capitalized)")
                        .font(.Echo.bodyMD)
                        .foregroundStyle(Color.Echo.outline)
                }
            } else {
                HStack(spacing: 8) {
                    Button("Vote For") { onVote("for") }
                        .font(.custom("Inter", size: 13)).fontWeight(.bold)
                        .foregroundStyle(.white)
                        .padding(.horizontal, 16).padding(.vertical, 10)
                        .background(Capsule().fill(LinearGradient.signature))
                    
                    Button("Against") { onVote("against") }
                        .font(.custom("Inter", size: 13)).fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)
                        .padding(.horizontal, 16).padding(.vertical, 10)
                        .background(Capsule().fill(Color.Echo.surfaceContainer))
                        .ghostBorder(radius: 9999)
                    
                    Button("Abstain") { onVote("abstain") }
                        .font(.custom("Inter", size: 13)).fontWeight(.bold)
                        .foregroundStyle(Color.Echo.outline)
                        .padding(.horizontal, 16).padding(.vertical, 10)
                }
            }
        }
        .padding(20)
        .background(
            RoundedRectangle(cornerRadius: 32)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .ghostBorder()
    }
}
```

---

## 13. Backup & Export

**Source:** PRD §Recovery, IOS_IMPL §Persistence  
**Priority:** P2

### BackupView.swift

```swift
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
                                .font(.Echo.bodyMD).fontWeight(.semibold)
                            Text("Never share it with anyone")
                                .font(.Echo.labelMD)
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
                        .font(.Echo.labelMD)
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
```

---

## 14-16. Enterprise Profile, Bot Management, Staking Detail

These follow the same patterns established above. Key implementation notes:

### EnterpriseProfileView.swift
- Displays verified organization credentials (business registration, domain, EV TLS)
- Shows official channels and authorized representatives
- Uses GhostBorderSection for each content block
- Verification badges use checkmark.circle.fill in success green

### BotManagementView.swift
- "My Bots" section with active bot cards (name, status, last triggered)
- "Discover Bots" section with add buttons
- Bot detail sheet with configuration options
- Sandboxed execution disclaimer at bottom

### StakingDetailView.swift
- Current position card (amount, tier, APY, lock period, multiplier)
- Staking tiers reference table (Bronze through Platinum)
- Delegation card (validator info, uptime, commission)
- "Stake More" and "Unstake" action buttons (signature gradient / ghost border)

---

## 17. Profile Page Expansion

**Current:** 126 lines — too thin  
**Required additions:**

```swift
// Add to ProfileView.swift body, after the trust tier bars:

// DID Display
VStack(spacing: 4) {
    Text("echo:alice.eth")
        .font(.custom("Inter", size: 13))
        .fontWeight(.semibold)
        .foregroundStyle(Color.Echo.primaryContainer)
    
    HStack(spacing: 6) {
        Text("did:cardano:addr1q8...x2f9")
            .font(.Echo.labelMD)
            .foregroundStyle(Color.Echo.outline)
            .lineLimit(1)
            .truncationMode(.middle)
        Button { UIPasteboard.general.string = fullDID } label: {
            Image(systemName: "doc.on.doc")
                .font(.system(size: 10))
                .foregroundStyle(Color.Echo.outline)
        }
    }
}

// Wallet Summary Card
Button { /* navigate to wallet */ } label: {
    HStack {
        VStack(alignment: .leading) {
            Text("ECHO BALANCE")
                .font(.Echo.labelSM).tracking(1)
                .foregroundStyle(Color.Echo.outline)
            Text("24,830.00 ECHO")
                .font(.custom("Inter", size: 20)).fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)
        }
        Spacer()
        Text("View Wallet →")
            .font(.Echo.labelMD)
            .foregroundStyle(Color.Echo.primaryContainer)
    }
    .padding(20)
    .background(RoundedRectangle(cornerRadius: 32).fill(Color.Echo.surfaceContainerLow))
    .ghostBorder()
}

// QR Code Share Button
Button { /* navigate to QR */ } label: {
    Label("Share My Identity", systemImage: "qrcode")
        .font(.custom("Inter", size: 14)).fontWeight(.bold)
        .foregroundStyle(Color.Echo.onSurface)
        .frame(maxWidth: .infinity)
        .padding(.vertical, 14)
        .background(RoundedRectangle(cornerRadius: 9999).fill(Color.Echo.surfaceContainerLow))
        .ghostBorder(radius: 9999)
}

// Credential Badges Section
GhostBorderSection(title: "VERIFIED CREDENTIALS") {
    CredentialRow(icon: "person.text.rectangle", name: "Government ID", verified: true)
    CredentialRow(icon: "apple.logo", name: "Apple ID", verified: true)
    CredentialRow(icon: "phone.fill", name: "Phone Number", verified: true)
    CredentialRow(icon: "building.2", name: "Organization", verified: false, action: "Verify →")
}
```

---

## 18. Existing Screen Fixes

### Fix 1: Remove All 1px Solid Borders

Search all views for `Divider()`, `.border()`, and `border-top`/`border-bottom` patterns. Replace with:
- Tonal shifts (different surface tier for adjacent sections)
- Ghost borders at 15% opacity where visual separation is needed
- Negative space (spacing tokens spacing-8 or spacing-10)

### Fix 2: Replace Pure Black

Search for `Color.black`, `.black`, `#000000`. Replace with:
- `Color.Echo.onSurface` (#131B2E) for text
- `Color.Echo.deepNavy` (#0F172A) for backgrounds

### Fix 3: Add SecureThreadIndicator

Add `SecureThreadIndicator()` to the top of every authenticated screen's body. Use `.overlay(alignment: .top)` pattern.

### Fix 4: Ensure 32px Corner Radius

Audit all `cornerRadius` values. Cards, buttons, and containers should use 32px. Only exceptions: inner elements (8px), input fields (16px), pills (9999px).

### Fix 5: Spring Animations Only

Replace any `.linear` or `.easeIn` animations with:
```swift
.animation(.spring(response: 0.4, dampingFraction: 0.85), value: someValue)
```

---

## 19. Navigation & Routing

### AppCoordinator Update

```swift
// Add new navigation destinations to the router

enum AppRoute: Hashable {
    // Existing
    case chat(conversationId: String)
    case groupChat(groupId: String)
    case groupCreate, groupDiscover, groupDetail(id: String)
    case channelCreate, channelDiscover, channelDetail(id: String)
    case personaCreate, personaDetail(id: String), personaList
    case hiddenFolders, hiddenFolderDetail(id: String), hiddenFolderChat(folderId: String, chatId: String)
    case trust, settings, deviceManagement, loginAudit, analytics
    
    // NEW routes
    case wallet
    case stakingDetail
    case delegationDetail
    case governance
    case proposalDetail(id: String)
    case contactDetail(id: String)
    case voiceCall(contactId: String)
    case videoCall(contactId: String)
    case search
    case searchScoped(conversationId: String)
    case mediaGallery(conversationId: String)
    case notifications
    case qrIdentity
    case backup
    case enterpriseProfile(id: String)
    case botManagement
    case botDetail(id: String)
}
```

### Deep Link Support

```swift
// Handle deep links for notifications
func handleDeepLink(_ url: URL) {
    guard let components = URLComponents(url: url, resolvingAgainstBaseURL: false) else { return }
    
    switch components.path {
    case "/chat":
        if let id = components.queryItems?.first(where: { $0.name == "id" })?.value {
            navigate(to: .chat(conversationId: id))
        }
    case "/wallet":
        selectedTab = .wallet
    case "/governance":
        navigate(to: .governance)
    case "/contact":
        if let id = components.queryItems?.first(where: { $0.name == "id" })?.value {
            navigate(to: .contactDetail(id: id))
        }
    default: break
    }
}
```

---

## 20. Data Models

### New SwiftData Models

```swift
import SwiftData

// Wallet Activity
@Model
class WalletActivityRecord {
    var id: String
    var type: String          // "reward", "stake", "unstake", "delegate", "transfer"
    var amount: Decimal
    var direction: String     // "in", "out"
    var label: String
    var timestamp: Date
    var transactionHash: String?
    
    init(id: String, type: String, amount: Decimal, direction: String, label: String, timestamp: Date) {
        self.id = id; self.type = type; self.amount = amount
        self.direction = direction; self.label = label; self.timestamp = timestamp
    }
}

// Search Index (local, never synced)
@Model
class SearchIndexEntry {
    var messageId: String
    var conversationId: String
    var keywords: String      // preprocessed searchable text
    var timestamp: Date
    var messageType: String
    var senderName: String
    var hasAttachment: Bool
    
    init(messageId: String, conversationId: String, keywords: String, timestamp: Date,
         messageType: String, senderName: String, hasAttachment: Bool) {
        self.messageId = messageId; self.conversationId = conversationId
        self.keywords = keywords; self.timestamp = timestamp
        self.messageType = messageType; self.senderName = senderName
        self.hasAttachment = hasAttachment
    }
}

// Notification
@Model
class NotificationRecord {
    var id: String
    var title: String
    var subtitle: String?
    var category: String
    var timestamp: Date
    var isRead: Bool
    var deepLink: String?
    
    init(id: String, title: String, subtitle: String?, category: String,
         timestamp: Date, isRead: Bool, deepLink: String?) {
        self.id = id; self.title = title; self.subtitle = subtitle
        self.category = category; self.timestamp = timestamp
        self.isRead = isRead; self.deepLink = deepLink
    }
}

// Governance
struct Proposal: Identifiable, Codable {
    let id: String
    let title: String
    let description: String
    let forVotes: Int
    let againstVotes: Int
    let abstainVotes: Int
    let startDate: Date
    let endDate: Date
    let status: String  // "active", "passed", "failed"
    
    var forPercentage: Double {
        let total = Double(forVotes + againstVotes + abstainVotes)
        guard total > 0 else { return 0 }
        return Double(forVotes) / total
    }
    
    var voterCount: Int { forVotes + againstVotes + abstainVotes }
    
    var timeRemaining: String {
        let remaining = endDate.timeIntervalSinceNow
        let days = Int(remaining / 86400)
        if days > 0 { return "\(days) days" }
        let hours = Int(remaining / 3600)
        return "\(hours) hours"
    }
}

struct VotingPower: Codable {
    let weight: Int           // raw token weight
    let trustTier: Int        // 1-5
    let multiplier: Double    // tier multiplier
    var effectivePower: Int { Int(Double(weight) * multiplier) }
}

// Staking
enum StakingTier: String, CaseIterable {
    case none, bronze, silver, gold, platinum
    
    var displayName: String { rawValue.capitalized }
    
    var minimumStake: Decimal {
        switch self {
        case .none: return 0
        case .bronze: return 100
        case .silver: return 1000
        case .gold: return 5000
        case .platinum: return 25000
        }
    }
    
    var apy: Double {
        switch self {
        case .none: return 0
        case .bronze: return 8.0
        case .silver: return 10.0
        case .gold: return 12.5
        case .platinum: return 15.0
        }
    }
    
    var multiplier: Double {
        switch self {
        case .none: return 0
        case .bronze: return 1.0
        case .silver: return 1.2
        case .gold: return 1.8
        case .platinum: return 2.5
        }
    }
    
    static func from(amount: Decimal) -> StakingTier {
        if amount >= 25000 { return .platinum }
        if amount >= 5000 { return .gold }
        if amount >= 1000 { return .silver }
        if amount >= 100 { return .bronze }
        return .none
    }
}
```

---

## 21. Build Phases

### Phase 1 — Foundation (Sprint 1-2)
- [ ] Update tab bar to 3 tabs (Messages, Wallet, Me)
- [ ] Implement GlacialTheme.swift with all tokens
- [ ] Create SecureThreadIndicator and add to all views
- [ ] Fix design violations (borders, black, shadows, radii)
- [ ] Create GhostBorderSection reusable component
- [ ] Create SpringButtonStyle and all view modifiers

### Phase 2 — Core Screens (Sprint 3-5)
- [ ] Wallet Tab (WalletTab, BalanceCard, BalanceBreakdown, WalletActionButton, DailyRewardsSection, RecentActivityList)
- [ ] WalletViewModel with Stargazer SDK integration
- [ ] Contact Detail (ContactDetailView, ContactDetailViewModel)
- [ ] Voice/Video Call (CallView, CallControlsBar, CallViewModel)
- [ ] Profile page expansion (DID, credentials, wallet summary, QR)

### Phase 3 — Search & Media (Sprint 5-7)
- [ ] Advanced Search (SearchView, SearchViewModel, local index)
- [ ] Media Gallery (MediaGalleryView, grid/list layouts)
- [ ] Notification Center (NotificationCenterView, badge system)
- [ ] QR Identity Share (QRIdentityView, QRScannerView)

### Phase 4 — Governance & Advanced (Sprint 7-9)
- [ ] Governance/Voting (GovernanceView, ProposalCard, VoteConfirmationSheet)
- [ ] Staking Detail (StakingDetailView, tier table)
- [ ] Backup & Export (BackupView, encrypted backup flow)
- [ ] Enterprise Profile (EnterpriseProfileView)
- [ ] Bot Management (BotManagementView)

### Phase 5 — Polish (Sprint 9-10)
- [ ] Dark mode audit for all new screens
- [ ] Accessibility pass (VoiceOver labels, Dynamic Type)
- [ ] Animation polish (all springs at 0.85 damping)
- [ ] Performance profiling (lazy loading, image caching)
- [ ] Integration testing with backend API stubs

---

## Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| StargazerSDK | Constellation wallet, token ops, delegation | Latest |
| WebRTC | Voice/video call infrastructure | M125 |
| CryptoKit | SHA-256, key derivation, QR signing | System |
| CoreImage | QR code generation | System |
| SwiftData | Local persistence | iOS 17+ |
| AVFoundation | Camera for QR scanning | System |

---

*ECHO iOS Swift Implementation Spec v4.2.1*  
*Aligned to: PRD v2.5.1, Tokenomics v2, IOS_IMPL v4.2, DESIGN.md v1.0*  
*Generated: April 2026*
