// EchoDemo — interactive Phase 1 screen browser.
// Tap any row to preview a live screen with mock/stub data.
// No backend connection required — all screens use local state.
//
// Sections:
//   Onboarding (Wave 12 flow)
//   Auth & Security
//   Recovery
//   Messaging
//   Personas & Hidden Folders
//   Profile & Settings
//   Wallet & Staking
//   Governance
//   Contacts
//   Other
//   Design System

import SwiftUI
import Echo

// MARK: - Root

struct DemoRootView: View {
    var body: some View {
        NavigationStack {
            List {

                // --------------------------------------------------------
                // MARK: Onboarding (4-step Wave 12 flow)
                // --------------------------------------------------------
                Section("Onboarding — Step by Step") {
                    DemoLink("1 · Welcome Carousel") {
                        WelcomeCarouselView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                    DemoLink("2 · Choose Username") {
                        DisplayNameEntryView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                    DemoLink("3 · Biometric Enrollment") {
                        BiometricEnrollmentView(
                            username: "chad",
                            onComplete: { _ in },
                            onUnsupported: {}
                        )
                    }
                    DemoLink("4 · Recovery Setup") {
                        RecoverySetupView(
                            did: "did:key:zDemoPhase1",
                            onComplete: {},
                            onSkip: {}
                        )
                    }
                    DemoLink("4b · SMS OTP Setup") {
                        SMSOTPSetupView(
                            did: "did:key:zDemoPhase1",
                            onConfigured: {}
                        )
                    }
                    DemoLink("Full Onboarding Flow") {
                        FirstRunCoordinatorView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                }

                // --------------------------------------------------------
                // MARK: Auth & Security
                // --------------------------------------------------------
                Section("Auth & Security") {
                    DemoLink("Login — Biometric Auto-trigger") {
                        GlacialLoginScreen(
                            onPasskeyLogin: {},
                            onSMSLogin: { _ in },
                            onGetStarted: {}
                        )
                    }
                    DemoLink("Storage Locked (WO-224)") {
                        StorageLockedView(onUnlock: {})
                    }
                    DemoLink("Biometric Lockout — Soft (5 fails)") {
                        BiometricLockoutView(
                            lockState: .requiresPasscode(failureCount: 5)
                        )
                    }
                    DemoLink("Biometric Lockout — Hard (10 fails)") {
                        BiometricLockoutView(
                            lockState: .hardLocked(until: Date().addingTimeInterval(12 * 60))
                        )
                    }
                    DemoLink("Passkey Setup (WO-136)") {
                        PasskeySetupView(onCompletion: {})
                    }
                    DemoLink("Device Management") {
                        DeviceManagementView()
                    }
                    DemoLink("Account Locked") {
                        AccountLockedView(
                            reason: .tooManyAttempts,
                            retryAfter: Date().addingTimeInterval(12 * 60),
                            onRecovery: {}
                        )
                    }
                    DemoLink("Trust Intro") {
                        TrustIntroView()
                    }
                }

                // --------------------------------------------------------
                // MARK: Recovery
                // --------------------------------------------------------
                Section("Recovery") {
                    DemoLink("Recovery Phrase Export") {
                        RecoveryCoordinatorView(
                            coordinator: RecoveryCoordinator(
                                onExportComplete: {},
                                onRestoreComplete: { _ in },
                                onCancel: {}
                            )
                        )
                    }
                    DemoLink("Restore from Phrase") {
                        RestoreFromPhraseView(
                            coordinator: RecoveryCoordinator(
                                onExportComplete: {},
                                onRestoreComplete: { _ in },
                                onCancel: {}
                            )
                        )
                    }
                }

                // --------------------------------------------------------
                // MARK: Credential Enrollment
                // --------------------------------------------------------
                Section("Credential Enrollment") {
                    DemoLink("Method Picker") {
                        EnrollmentMethodPickerView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                    DemoLink("Driver's License") {
                        DriversLicenseEnrollmentView(
                            coordinator: EnrollmentCoordinator(onComplete: { _ in }, onCancel: {})
                        )
                    }
                    DemoLink("IDV Fallback") {
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

                // --------------------------------------------------------
                // MARK: Messaging
                // --------------------------------------------------------
                Section("Messaging") {
                    DemoLink("Conversation List") {
                        ConversationListView()
                    }
                    DemoLink("Chat (Alice Johnson)") {
                        ChatView(contactName: "Alice Johnson")
                    }
                    DemoLink("Messages Empty State") {
                        MessagesEmptyStateView(
                            displayName: "Chad",
                            trustTier: 2,
                            onComposeTapped: {},
                            onUpgradeTrustTapped: {}
                        )
                    }
                }

                // --------------------------------------------------------
                // MARK: Personas & Hidden Folders (Wave 12)
                // --------------------------------------------------------
                Section("Personas & Hidden Folders") {
                    DemoLink("Persona Management") {
                        PersonasManagementView()
                    }
                    DemoLink("Hidden Persona Gate — Locked") {
                        PersonaGateView(personaID: "demo-hidden-persona") {
                            VStack(spacing: 16) {
                                Image(systemName: "person.crop.rectangle.stack.fill")
                                    .font(.system(size: 48))
                                    .foregroundColor(.echoPrimary)
                                Text("Hidden Persona Content")
                                    .font(.system(size: 18, weight: .semibold))
                                    .foregroundColor(.echoPrimaryText)
                                Text("This content is only visible after Face ID verification.")
                                    .font(.system(size: 14))
                                    .foregroundColor(.echoSecondaryText)
                                    .multilineTextAlignment(.center)
                                    .padding(.horizontal, 32)
                            }
                            .frame(maxWidth: .infinity, maxHeight: .infinity)
                            .background(Color.echoBackground.ignoresSafeArea())
                        }
                    }
                }

                // --------------------------------------------------------
                // MARK: Profile & Settings
                // --------------------------------------------------------
                Section("Profile & Settings") {
                    DemoLink("Profile Tab")             { ProfileTabView() }
                    DemoLink("Edit Profile")            { EditProfileView() }
                    DemoLink("Account Settings")        { AccountSettingsView() }
                    DemoLink("Privacy & Security")      { PrivacySecuritySettingsView() }
                    DemoLink("Notification Settings")   { NotificationSettingsView() }
                    DemoLink("Appearance")              { AppearanceSettingsView() }
                    DemoLink("Storage & Data")          { StorageDataView() }
                    DemoLink("Backup")                  { BackupView() }
                    DemoLink("About")                   { AboutView() }
                    DemoLink("QR Identity")             { QRIdentityView() }
                }

                // --------------------------------------------------------
                // MARK: Wallet & Staking
                // --------------------------------------------------------
                Section("Wallet & Staking") {
                    DemoLink("Staking")                 { StakingView() }
                    DemoLink("Staking Detail")          { StakingDetailView() }
                    DemoLink("Validator Browser")       { ValidatorBrowserView() }
                    DemoLink("Rewards Dashboard")       { RewardsDashboardView() }
                }

                // --------------------------------------------------------
                // MARK: Governance
                // --------------------------------------------------------
                Section("Governance") {
                    DemoLink("Voting Power") {
                        GovernanceWeightView(power: DemoData.votingPower)
                    }
                    DemoLink("Proposal List")           { DemoProposalListView() }
                    DemoLink("Vote Confirmation") {
                        VoteConfirmationView(
                            proposal: DemoData.proposal,
                            vote: "for",
                            onConfirm: {},
                            onCancel: {}
                        )
                    }
                }

                // --------------------------------------------------------
                // MARK: Contacts
                // --------------------------------------------------------
                Section("Contacts") {
                    DemoLink("Contacts List")           { ContactsListView() }
                    DemoLink("Contact Detail") {
                        ContactDetailView(contactId: "demo-contact-1")
                    }
                }

                // --------------------------------------------------------
                // MARK: Other Screens
                // --------------------------------------------------------
                Section("Other") {
                    DemoLink("Search")                  { SearchView() }
                    DemoLink("Notification Center")     { NotificationCenterView() }
                    DemoLink("Bot Management")          { BotManagementView() }
                    DemoLink("Media Gallery") {
                        MediaGalleryView(conversationId: "demo-conv-1")
                    }
                }

                // --------------------------------------------------------
                // MARK: Design System
                // --------------------------------------------------------
                Section("Design System") {
                    DemoLink("Color Palette")           { ColorPalettePreview() }
                    DemoLink("Typography")              { TypographyPreview() }
                    DemoLink("Components")              { ComponentsPreview() }
                }
            }
            .listStyle(.insetGrouped)
            .navigationTitle("Echo Phase 1")
            .navigationBarTitleDisplayMode(.large)
        }
    }
}

// MARK: - DemoLink helper

struct DemoLink<Destination: View>: View {
    let title: String
    let destination: () -> Destination

    init(_ title: String, @ViewBuilder destination: @escaping () -> Destination) {
        self.title = title
        self.destination = destination
    }

    var body: some View {
        NavigationLink(title, destination: destination)
    }
}

// MARK: - Demo data

enum DemoData {
    static let votingPower = GovernanceVotingPower(
        totalPower: 1_250,
        stakedTokens: 500,
        trustMultiplier: 2.5,
        tier: 3
    )

    static let proposal = GovernanceProposal(
        id: "prop-001",
        title: "Increase relay fee rebate to 15%",
        description: "Proposal to increase the relay fee rebate from 10% to 15% to incentivize node operators.",
        status: .active,
        votesFor: 12_400,
        votesAgainst: 3_200,
        abstentions: 600,
        deadline: Date().addingTimeInterval(3 * 24 * 60 * 60)
    )
}

// MARK: - Design System Previews

struct ColorPalettePreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                group("Brand") {
                    Swatch("echoPrimary",       .echoPrimary)
                    Swatch("echoBackground",    .echoBackground)
                    Swatch("echoCardBg",        .echoCardBackground)
                    Swatch("echoPrimaryText",   .echoPrimaryText)
                    Swatch("echoSecondaryText", .echoSecondaryText)
                    Swatch("echoError",         .echoError)
                }

                group("Trust Tiers") {
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
        .navigationTitle("Color Palette")
    }

    @ViewBuilder
    func group<C: View>(_ title: String, @ViewBuilder content: () -> C) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title)
                .font(.system(size: 13, weight: .semibold))
                .foregroundColor(.echoSecondaryText)
                .textCase(.uppercase)
                .tracking(1)
            content()
        }
    }

    struct Swatch: View {
        let name: String
        let color: Color
        init(_ n: String, _ c: Color) { name = n; color = c }
        var body: some View {
            HStack(spacing: 12) {
                RoundedRectangle(cornerRadius: 8)
                    .fill(color)
                    .frame(width: 44, height: 44)
                    .overlay(RoundedRectangle(cornerRadius: 8)
                        .strokeBorder(.primary.opacity(0.08), lineWidth: 0.5))
                Text(name)
                    .font(.system(size: 13))
                    .foregroundColor(.echoPrimaryText)
                Spacer()
            }
        }
    }
}

struct TypographyPreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                Text("Display Large — 56pt Bold").font(Font.Echo.displayLarge)
                Text("Display Medium — 45pt Bold").font(Font.Echo.displayMedium)
                Text("Headline Sm — 24pt Bold").font(Font.Echo.headlineSm)
                Text("Title Large — 20pt Bold").font(Font.Echo.titleLarge)
                Text("Body Large — 16pt Medium").font(Font.Echo.bodyLarge)
                Text("Body Medium — 14pt Medium").font(Font.Echo.bodyMedium)
                Divider()
                Text("System 22 Semibold — screen titles")
                    .font(.system(size: 22, weight: .semibold))
                Text("System 15 Regular — body copy with comfortable reading line length")
                    .font(.system(size: 15))
                Text("System 13 — secondary captions and metadata labels")
                    .font(.system(size: 13))
                    .foregroundColor(.echoSecondaryText)
                Text("System 11 Semibold · TRACKING · LABELS")
                    .font(.system(size: 11, weight: .semibold))
                    .tracking(1.4)
                    .foregroundColor(.echoSecondaryText)
            }
            .foregroundColor(.echoPrimaryText)
            .padding()
        }
        .background(Color.echoBackground.ignoresSafeArea())
        .navigationTitle("Typography")
    }
}

struct ComponentsPreview: View {
    @State private var text = ""
    @State private var otpCode = ""

    private let trustLevels = ["Unverified", "Standard", "Phone", "Active", "Credential", "Identity"]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 28) {
                section("Buttons") {
                    EchoButton("Primary Action",   style: .primary)    {}
                    EchoButton("Secondary",        style: .secondary)  {}
                    EchoButton("Destructive",      style: .destructive) {}
                    EchoButton("Ghost",            style: .ghost)      {}
                    EchoButton("Disabled",         style: .primary)    {}
                        .disabled(true)
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
                                Text("T\(tier)").font(.system(size: 11)).foregroundColor(.echoSecondaryText)
                            }
                        }
                    }
                }

                section("OTP Input") {
                    OTPInputView(code: $otpCode, length: 6)
                }

                section("Secure Thread Indicator") {
                    SecureThreadIndicator()
                }

                section("Card") {
                    EchoCard {
                        VStack(alignment: .leading, spacing: 6) {
                            Text("EchoCard").font(.system(size: 15, weight: .semibold))
                                .foregroundColor(.echoPrimaryText)
                            Text("Surface card with rounded corners and subtle shadow.")
                                .font(.system(size: 13)).foregroundColor(.echoSecondaryText)
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
        .navigationTitle("Components")
    }

    @ViewBuilder
    func section<C: View>(_ title: String, @ViewBuilder content: () -> C) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title).font(.system(size: 13, weight: .semibold))
                .foregroundColor(.echoSecondaryText).textCase(.uppercase).tracking(1)
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
