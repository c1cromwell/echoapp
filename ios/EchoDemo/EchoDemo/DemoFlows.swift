// EchoDemo — interactive screen browser
// Tap any row to navigate to a live screen with mock/stub data.
// No backend required; screens use built-in state.

import SwiftUI
import Echo

// MARK: - Root

struct DemoRootView: View {
    var body: some View {
        NavigationStack {
            List {
                Section("Onboarding") {
                    DemoLink("Welcome / Splash")       { WelcomeView() }
                    DemoLink("Login Screen")           { AuthLoginView() }
                    DemoLink("Trust Intro")            { TrustIntroView() }
                    DemoLink("Display Name Entry") {
                        DisplayNameEntryView(
                            coordinator: FirstRunCoordinator(onComplete: {}, onCancel: {})
                        )
                    }
                    DemoLink("Welcome Carousel") {
                        WelcomeCarouselView(
                            coordinator: FirstRunCoordinator(onComplete: {}, onCancel: {})
                        )
                    }
                }

                Section("Credential Enrollment") {
                    DemoLink("Method Picker") {
                        EnrollmentMethodPickerView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                    DemoLink("Drivers License Hub") {
                        DriversLicenseEnrollmentView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                    DemoLink("IDV Fallback (Scan + Selfie)") {
                        IDVFallbackView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                    DemoLink("Wallet Credential") {
                        WalletCredentialEnrollmentView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                }

                Section("Messaging") {
                    DemoLink("Conversation List")      { ConversationListView() }
                    DemoLink("Chat (Alice Johnson)")   { ChatView(contactName: "Alice Johnson") }
                    DemoLink("Messages Empty State") {
                        MessagesEmptyStateView(
                            displayName: "Chad",
                            trustTier: 2,
                            onComposeTapped: {},
                            onUpgradeTrustTapped: {}
                        )
                    }
                }

                Section("Profile & Settings") {
                    DemoLink("Profile Tab")            { ProfileTabView() }
                    DemoLink("Edit Profile")           { EditProfileView() }
                    DemoLink("Personas")               { PersonasManagementView() }
                    DemoLink("Account Settings")       { AccountSettingsView() }
                    DemoLink("Privacy & Security")     { PrivacySecuritySettingsView() }
                    DemoLink("Notification Settings")  { NotificationSettingsView() }
                    DemoLink("Appearance")             { AppearanceSettingsView() }
                    DemoLink("Storage & Data")         { StorageDataView() }
                    DemoLink("About")                  { AboutView() }
                }

                Section("Wallet & Staking") {
                    DemoLink("Staking")                { StakingView() }
                    DemoLink("Staking Detail")         { StakingDetailView() }
                    DemoLink("Validator Browser")      { ValidatorBrowserView() }
                    DemoLink("Rewards Dashboard")      { RewardsDashboardView() }
                }

                Section("Governance") {
                    DemoLink("Voting Power") {
                        GovernanceWeightView(power: DemoData.votingPower)
                    }
                    DemoLink("Proposal List")          { DemoProposalListView() }
                }

                Section("Auth & Security") {
                    DemoLink("Device Management")      { DeviceManagementView() }
                    DemoLink("Account Locked") {
                        AccountLockedView(
                            reason: .tooManyAttempts,
                            retryAfter: Date().addingTimeInterval(12 * 60),
                            onRecovery: {}
                        )
                    }
                    DemoLink("Recovery Phrase Export") {
                        RecoveryCoordinatorView(
                            coordinator: RecoveryCoordinator(
                                onExportComplete: {},
                                onRestoreComplete: { _ in },
                                onCancel: {}
                            )
                        )
                    }
                }

                Section("Contacts") {
                    DemoLink("Contacts List")          { ContactsListView() }
                    DemoLink("Contact Detail") {
                        ContactDetailView(contactId: "demo-contact-1")
                    }
                }

                Section("Other Screens") {
                    DemoLink("QR Identity")            { QRIdentityView() }
                    DemoLink("Backup")                 { BackupView() }
                    DemoLink("Search")                 { SearchView() }
                    DemoLink("Notification Center")    { NotificationCenterView() }
                    DemoLink("Bot Management")         { BotManagementView() }
                    DemoLink("Media Gallery") {
                        MediaGalleryView(conversationId: "demo-conv-1")
                    }
                }

                Section("Design System") {
                    DemoLink("Colors & Typography")    { ColorTypographyPreview() }
                    DemoLink("Components Gallery")     { ComponentsPreview() }
                }
            }
            .navigationTitle("ECHO Demo")
            .navigationBarTitleDisplayMode(.large)
        }
    }
}

// MARK: - Navigation helper

struct DemoLink<Destination: View>: View {
    let title: String
    let make: () -> Destination

    init(_ title: String, @ViewBuilder make: @escaping () -> Destination) {
        self.title = title
        self.make = make
    }

    var body: some View {
        NavigationLink(title) {
            make()
                .navigationTitle(title)
                .navigationBarTitleDisplayMode(.inline)
        }
    }
}

// MARK: - Stub data

enum DemoData {
    static let votingPower = VotingPower(
        did: "did:key:zDemoUser",
        trustTier: 3,
        multiplier: 1.0,
        totalStaked: 5_000,
        weight: 5_000,
        canVote: true
    )

    static var proposals: [Proposal] {[
        Proposal(
            id: "prop-1",
            title: "Increase daily reward cap by 10%",
            description: "Raises messaging reward ceiling from 500 to 550 messages per day.",
            type: "parameter_change",
            threshold: "simple_majority",
            createdBy: "did:key:zFounder",
            createdAt: Date().addingTimeInterval(-86400 * 3),
            endsAt: Date().addingTimeInterval(86400 * 4),
            status: "active",
            tally: ProposalTally(
                proposalID: "prop-1",
                forWeight: 4_200, againstWeight: 1_100, abstainWeight: 300,
                totalWeight: 5_600, forPercent: 75.0, voterCount: 18, passed: false
            )
        ),
        Proposal(
            id: "prop-2",
            title: "Relay node operator staking requirement",
            description: "Introduces a 1,000 ECHO minimum stake for relay node operators.",
            type: "protocol_upgrade",
            threshold: "supermajority_67",
            createdBy: "did:key:zCommunity",
            createdAt: Date().addingTimeInterval(-86400 * 7),
            endsAt: Date().addingTimeInterval(86400 * 1),
            status: "active",
            tally: ProposalTally(
                proposalID: "prop-2",
                forWeight: 7_200, againstWeight: 2_100, abstainWeight: 300,
                totalWeight: 9_600, forPercent: 75.0, voterCount: 42, passed: false
            )
        ),
    ]}
}

// MARK: - Governance stub list

struct DemoProposalListView: View {
    var body: some View {
        List(DemoData.proposals) { proposal in
            NavigationLink {
                ProposalDetailView(
                    proposal: proposal,
                    votingPower: DemoData.votingPower,
                    onVote: { _ in }
                )
                .navigationTitle("Proposal")
                .navigationBarTitleDisplayMode(.inline)
            } label: {
                VStack(alignment: .leading, spacing: 6) {
                    Text(proposal.title).font(.headline)
                    HStack {
                        Text(proposal.status.uppercased())
                            .font(.caption.bold()).foregroundColor(.echoSuccess)
                        Spacer()
                        Text(proposal.endsAt, style: .relative)
                            .font(.caption).foregroundColor(.echoSecondaryText)
                    }
                }
                .padding(.vertical, 4)
            }
        }
        .navigationTitle("Proposals")
        .background(Color.echoBackground.ignoresSafeArea())
    }
}

// MARK: - Colors & typography preview

struct ColorTypographyPreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Typography").typographyStyle(.h2, color: .echoPrimaryText)
                    ForEach([
                        ("Display",    TypographyStyle.display),
                        ("Heading 1",  .h1),
                        ("Heading 2",  .h2),
                        ("Heading 3",  .h3),
                        ("Body Large", .bodyLarge),
                        ("Body",       .body),
                        ("Body Small", .bodySmall),
                        ("Caption",    .caption),
                        ("Label",      .label),
                    ], id: \.0) { name, style in
                        HStack {
                            Text(name).typographyStyle(style, color: .echoPrimaryText)
                            Spacer()
                            Text(String(format: "%.0fpt", style.fontSize))
                                .typographyStyle(.caption, color: .echoSecondaryText)
                        }
                    }
                }

                Divider()

                VStack(alignment: .leading, spacing: 8) {
                    Text("Brand Colors").typographyStyle(.h2, color: .echoPrimaryText)
                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 110))], spacing: 10) {
                        Swatch("Primary",    .echoPrimary)
                        Swatch("Secondary",  .echoSecondary)
                        Swatch("Success",    .echoSuccess)
                        Swatch("Warning",    .echoWarning)
                        Swatch("Error",      .echoError)
                        Swatch("Info",       .echoInfo)
                        Swatch("Background", .echoBackground)
                        Swatch("Card",       .echoCardBackground)
                    }
                }

                Divider()

                VStack(alignment: .leading, spacing: 8) {
                    Text("Trust Tier Palette").typographyStyle(.h2, color: .echoPrimaryText)
                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 110))], spacing: 10) {
                        Swatch("T0 Unverified", .echoTrustUnverified)
                        Swatch("T1 Basic",      .echoTrustBasic)
                        Swatch("T2 Verified",   .echoTrustVerified)
                        Swatch("T3 Trusted",    .echoTrustTrusted)
                        Swatch("T4 Premium",    .echoTrustPremium)
                        Swatch("T5 Elite",      .echoTrustElite)
                    }
                }
            }
            .padding()
        }
        .background(Color.echoBackground.ignoresSafeArea())
    }

    struct Swatch: View {
        let name: String; let color: Color
        init(_ n: String, _ c: Color) { name = n; color = c }
        var body: some View {
            VStack(spacing: 4) {
                RoundedRectangle(cornerRadius: 10)
                    .fill(color)
                    .overlay(RoundedRectangle(cornerRadius: 10)
                        .strokeBorder(.primary.opacity(0.08), lineWidth: 0.5))
                    .frame(height: 52)
                Text(name).typographyStyle(.caption, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)
            }
        }
    }
}

// MARK: - Components gallery

struct ComponentsPreview: View {
    @State private var text = ""
    @State private var otpCode = ""

    private let trustLevels = ["Unverified","Standard","Phone","Active","Credential","Identity"]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 28) {
                section("Buttons") {
                    EchoButton("Primary Action",  style: .primary)    {}
                    EchoButton("Secondary",       style: .secondary)  {}
                    EchoButton("Destructive",     style: .destructive) {}
                    EchoButton("Ghost",           style: .ghost)      {}
                }

                section("Text Fields") {
                    EchoTextField(placeholder: "Username or DID", text: $text)
                    EchoTextField(placeholder: "Password", text: $text, isSecure: true)
                }

                section("Trust Badges") {
                    HStack(spacing: 12) {
                        ForEach(0..<6) { tier in
                            VStack(spacing: 4) {
                                TrustBadge(level: trustLevels[safe: tier] ?? "T\(tier)", size: .large)
                                Text("T\(tier)").typographyStyle(.caption, color: .echoSecondaryText)
                            }
                        }
                    }
                }

                section("OTP Input") {
                    OTPInputView(code: $otpCode, length: 6)
                }

                section("Card") {
                    EchoCard {
                        VStack(alignment: .leading, spacing: 6) {
                            Text("EchoCard").typographyStyle(.bodyLarge, color: .echoPrimaryText)
                            Text("Surface card with rounded corners and subtle shadow.")
                                .typographyStyle(.caption, color: .echoSecondaryText)
                        }
                        .padding()
                    }
                }

                section("Message Bubbles") {
                    MessageBubble(message: "Hey! Outgoing message.", isSent: true,
                                  status: .delivered, timestamp: "2:45 PM")
                    MessageBubble(message: "Incoming reply that wraps nicely.", isSent: false,
                                  status: .read, timestamp: "2:46 PM")
                }

                section("Trust Score") {
                    HStack(spacing: 20) {
                        TrustScoreView(score: 45, level: "Newcomer")
                        TrustScoreView(score: 72, level: "Trusted")
                        TrustScoreView(score: 95, level: "Elite")
                    }
                    .frame(maxWidth: .infinity)
                }
            }
            .padding()
        }
        .background(Color.echoBackground.ignoresSafeArea())
    }

    @ViewBuilder
    func section<Content: View>(_ title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title).typographyStyle(.h3, color: .echoPrimaryText)
            content()
        }
        Divider()
    }
}

// MARK: - Helpers

extension Array {
    subscript(safe index: Int) -> Element? {
        indices.contains(index) ? self[index] : nil
    }
}
