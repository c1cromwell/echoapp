// EchoDemo — Phase 1 + Design Review screen browser.
// Tap any row to preview a live screen with mock/stub data.
// No backend required — all screens use local state.
//
// Sections reflect the Claude Design Review implementation:
//   Onboarding (design-review 2-step flow first, legacy 4-step below)
//   Auth & Security
//   Recovery
//   Messaging
//   Personas & Hidden Folders
//   Identity & Privacy (new design-review screens)
//   Profile & Settings
//   Wallet & Staking
//   Governance
//   Contacts
//   Other
//   Design System (updated tokens)

import SwiftUI
import Echo

// MARK: - Root

struct DemoRootView: View {
    var body: some View {
        NavigationStack {
            List {

                // ────────────────────────────────────────────────────────
                // MARK: Onboarding — Design Review (2-step)
                // "Compress onboarding to two screens" — Claude Design
                // ────────────────────────────────────────────────────────
                Section("Onboarding — Design Review (2-step)") {
                    DemoLink("1 · Welcome (single page, no carousel)") {
                        EchoWelcomeView(onSetUp: {}, onAlreadyHaveAccount: {})
                    }
                    DemoLink("2 · Name + Face ID (fused)") {
                        NameAndKeyView(onComplete: { _, _ in }, onSkip: {})
                    }
                    DemoLink("2b · Recovery Setup") {
                        RecoverySetupView(
                            did: "did:key:zDemoDesignReview",
                            onComplete: {},
                            onSkip: {}
                        )
                    }
                    DemoLink("2c · SMS OTP Setup") {
                        SMSOTPSetupView(
                            did: "did:key:zDemoDesignReview",
                            onConfigured: {}
                        )
                    }
                    DemoLink("Full 2-step Flow") {
                        FirstRunCoordinatorView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Onboarding — Legacy (Wave 12 4-step, kept for reference)
                // ────────────────────────────────────────────────────────
                Section("Onboarding — Legacy (4-step, Wave 12)") {
                    DemoLink("Welcome Carousel (old)") {
                        WelcomeCarouselView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                    DemoLink("Username Entry (old step 1)") {
                        DisplayNameEntryView(
                            coordinator: FirstRunCoordinator(
                                onComplete: { _ in },
                                onRestoreComplete: { _ in }
                            )
                        )
                    }
                    DemoLink("Biometric Enrollment (old step 3)") {
                        BiometricEnrollmentView(
                            username: "chad",
                            onComplete: { _ in },
                            onUnsupported: {}
                        )
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Auth & Security
                // ────────────────────────────────────────────────────────
                Section("Auth & Security") {
                    DemoLink("Login — 'Look at your phone' (design review)") {
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

                // ────────────────────────────────────────────────────────
                // MARK: Recovery
                // ────────────────────────────────────────────────────────
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

                // ────────────────────────────────────────────────────────
                // MARK: Credential Enrollment
                // ────────────────────────────────────────────────────────
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

                // ────────────────────────────────────────────────────────
                // MARK: Messaging
                // ────────────────────────────────────────────────────────
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

                // ────────────────────────────────────────────────────────
                // MARK: Personas & Hidden Folders
                // ────────────────────────────────────────────────────────
                Section("Personas & Hidden Folders") {
                    DemoLink("Persona Management") {
                        PersonasManagementView()
                    }
                    DemoLink("Hidden Persona Gate — Night Surface (design review)") {
                        PersonaGateView(personaID: "demo-hidden-persona") {
                            VStack(spacing: 16) {
                                Image(systemName: "person.crop.rectangle.stack.fill")
                                    .font(.system(size: 48))
                                    .foregroundStyle(Color.echoPrimary)
                                Text("Hidden Persona Content")
                                    .font(.system(size: 18, weight: .semibold))
                                    .foregroundStyle(Color.echoPrimaryText)
                                Text("This content is only visible after Face ID.")
                                    .font(.system(size: 14))
                                    .foregroundStyle(Color.echoSecondaryText)
                                    .multilineTextAlignment(.center)
                                    .padding(.horizontal, 32)
                            }
                            .frame(maxWidth: .infinity, maxHeight: .infinity)
                            .background(Color.echoBackground.ignoresSafeArea())
                        }
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Identity & Privacy (design review — new screens)
                // ────────────────────────────────────────────────────────
                Section("Identity & Privacy — Design Review") {
                    DemoLink("Identity Card (DID visible)") {
                        ScrollView {
                            VStack(spacing: 20) {
                                IdentityCardView(
                                    username: "ada",
                                    did: "did:key:z6MkhRfL7r9NB8Y3xQpJh···aV2nf9Q2",
                                    joinDate: Calendar.current.date(byAdding: .month,
                                                                    value: -3, to: Date())
                                )
                                IdentityProtectedList(hasPhrase: true, hasSMS: false)
                            }
                            .padding(24)
                        }
                        .background(Color.echoPaper.ignoresSafeArea())
                        .preferredColorScheme(.light)
                    }
                    DemoLink("Privacy Settings — 'NEVER COLLECTED' chips") {
                        PrivacySecuritySettingsView(
                            settings: .constant(EnhancedPrivacySettings()),
                            onSave: { _ in }
                        )
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Profile & Settings
                // ────────────────────────────────────────────────────────
                Section("Profile & Settings") {
                    DemoLink("Profile Tab")             { ProfileTabView() }
                    DemoLink("Edit Profile")            { EditProfileView() }
                    DemoLink("Account Settings")        { AccountSettingsView() }
                    DemoLink("Notification Settings")   { NotificationSettingsView() }
                    DemoLink("Appearance")              { AppearanceSettingsView() }
                    DemoLink("Storage & Data")          { StorageDataView() }
                    DemoLink("Backup")                  { BackupView() }
                    DemoLink("About")                   { AboutView() }
                    DemoLink("QR Identity")             { QRIdentityView() }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Wallet & Staking
                // ────────────────────────────────────────────────────────
                Section("Wallet & Staking") {
                    DemoLink("Staking")                 { StakingView() }
                    DemoLink("Staking Detail")          { StakingDetailView() }
                    DemoLink("Validator Browser")       { ValidatorBrowserView() }
                    DemoLink("Rewards Dashboard")       { RewardsDashboardView() }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Governance
                // ────────────────────────────────────────────────────────
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

                // ────────────────────────────────────────────────────────
                // MARK: Contacts
                // ────────────────────────────────────────────────────────
                Section("Contacts") {
                    DemoLink("Contacts List")           { ContactsListView() }
                    DemoLink("Contact Detail") {
                        ContactDetailView(contactId: "demo-contact-1")
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Other
                // ────────────────────────────────────────────────────────
                Section("Other") {
                    DemoLink("Search")                  { SearchView() }
                    DemoLink("Notification Center")     { NotificationCenterView() }
                    DemoLink("Bot Management")          { BotManagementView() }
                    DemoLink("Media Gallery") {
                        MediaGalleryView(conversationId: "demo-conv-1")
                    }
                }

                // ────────────────────────────────────────────────────────
                // MARK: Design System (reflects design review tokens)
                // ────────────────────────────────────────────────────────
                Section("Design System") {
                    DemoLink("Color Palette (updated)")  { ColorPalettePreview() }
                    DemoLink("Typography + Mono")        { TypographyPreview() }
                    DemoLink("Components")               { ComponentsPreview() }
                    DemoLink("EchoRippleMark")           { EchoMarkPreview() }
                    DemoLink("Trust — One Hue, 5 Steps") { TrustScalePreview() }
                }
            }
            .listStyle(.insetGrouped)
            .navigationTitle("Echo — Design Review")
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

// Updated for design review: warm paper palette + trust single-hue
struct ColorPalettePreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 28) {
                tokenGroup("Surfaces (warm paper)") {
                    Swatch("echoPaper",      Color.echoPaper,      "#FBFBF7")
                    Swatch("echoPaperDim",   Color.echoPaperDim,   "#F2F1EA")
                    Swatch("echoPaperEdge",  Color.echoPaperEdge,  "#E6E4DA")
                    Swatch("echoNight",      Color.echoNight,      "#0E1418 (private)")
                    Swatch("echoNightHi",    Color.echoNightHi,    "#161D22")
                }
                tokenGroup("Ink") {
                    Swatch("echoInk",        Color.echoInk,        "#0B0F10")
                    Swatch("echoInk70",      Color.echoInk70,      "0B0F10 · 70%")
                    Swatch("echoInk55",      Color.echoInk55,      "0B0F10 · 55%")
                    Swatch("echoInk40",      Color.echoInk40,      "0B0F10 · 40%")
                    Swatch("echoHair",       Color.echoHair,       "0B0F10 · 10%")
                }
                tokenGroup("Accent + status") {
                    Swatch("echoSignal",     Color.echoSignal,     "#0E7AB8 (one accent)")
                    Swatch("echoTrustGreen", Color.echoTrustGreen, "#1F7A4C (one affirming hue)")
                    Swatch("echoAlert",      Color.echoAlert,      "#B5341B (less saturated)")
                }
                tokenGroup("Trust — one hue, five opacities (replaces 6-color rainbow)") {
                    Swatch("T0 Unverified", Color.echoTrustUnverified, "10%")
                    Swatch("T1 Basic",      Color.echoTrustBasic,      "25%")
                    Swatch("T2 Verified",   Color.echoTrustVerified,   "45%")
                    Swatch("T3 Trusted",    Color.echoTrustTrusted,    "70%")
                    Swatch("T4 Elite",      Color.echoTrustElite,      "100%")
                }
            }
            .padding(24)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .preferredColorScheme(.light)
        .navigationTitle("Color Palette")
    }

    @ViewBuilder
    func tokenGroup<C: View>(_ title: String, @ViewBuilder content: () -> C) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title)
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(Color.echoInk40)
                .textCase(.uppercase)
                .tracking(0.8)
            content()
        }
    }

    struct Swatch: View {
        let name: String; let color: Color; let detail: String
        init(_ n: String, _ c: Color, _ d: String = "") { name = n; color = c; detail = d }
        var body: some View {
            HStack(spacing: 12) {
                RoundedRectangle(cornerRadius: 8)
                    .fill(color)
                    .frame(width: 44, height: 44)
                    .overlay(RoundedRectangle(cornerRadius: 8)
                        .strokeBorder(Color.echoHair, lineWidth: 0.5))
                VStack(alignment: .leading, spacing: 2) {
                    Text(name)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.echoInk)
                    if !detail.isEmpty {
                        Text(detail)
                            .font(.echomono(10.5))
                            .foregroundStyle(Color.echoInk40)
                    }
                }
                Spacer()
            }
        }
    }
}

struct TypographyPreview: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 24) {
                group("Inter — UI") {
                    Text("Display 34 · Semibold · −1.0")
                        .font(.system(size: 34, weight: .semibold)).tracking(-1.0)
                        .foregroundStyle(Color.echoInk)
                    Text("Title 26 · Semibold · −0.7")
                        .font(.system(size: 26, weight: .semibold)).tracking(-0.7)
                        .foregroundStyle(Color.echoInk)
                    Text("Heading 22 · Semibold · −0.5")
                        .font(.system(size: 22, weight: .semibold)).tracking(-0.5)
                        .foregroundStyle(Color.echoInk)
                    Text("Body 15 · Regular")
                        .font(.system(size: 15))
                        .foregroundStyle(Color.echoInk70)
                    Text("Small 13 · Regular — secondary text and captions")
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoInk55)
                    Text("LABEL 11 · SEMIBOLD · TRACKING 0.8")
                        .font(.system(size: 11, weight: .semibold)).tracking(0.8)
                        .foregroundStyle(Color.echoInk40)
                }
                Divider()
                group("Geist Mono — cryptographic content") {
                    Text("WHAT ECHO STORES")
                        .font(.echomono(10))
                        .foregroundStyle(Color.echoInk40)
                    Text("did:key:z6MkhRfL7r9NB8Y3xQpJh···aV2nf9Q2")
                        .font(.echomono(12.5))
                        .foregroundStyle(Color.echoInk)
                    Text("verified · key e2:9c:1a")
                        .font(.echomono(10.5))
                        .foregroundStyle(Color.echoInk55)
                    Text("● E2EE")
                        .font(.echomono(10))
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("● SCANNING")
                        .font(.echomono(10)).tracking(1)
                        .foregroundStyle(Color.echoTrustGreen)
                    Text("1 / 2")
                        .font(.echomono(11))
                        .foregroundStyle(Color.echoInk40)
                    HStack(spacing: 12) {
                        ForEach(["1", "2", "3", "4", "5", "6"], id: \.self) { d in
                            Text(d)
                                .font(.echomono(22))
                                .foregroundStyle(Color.echoInk)
                                .frame(width: 40, height: 48)
                                .background(Color.echoPaperDim,
                                            in: RoundedRectangle(cornerRadius: 8))
                        }
                    }
                }
            }
            .padding(24)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .preferredColorScheme(.light)
        .navigationTitle("Typography")
    }

    @ViewBuilder
    func group<C: View>(_ title: String, @ViewBuilder content: () -> C) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(title)
                .font(.system(size: 11, weight: .semibold)).tracking(0.8)
                .foregroundStyle(Color.echoInk40).textCase(.uppercase)
            content()
        }
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
                    EchoButton("Disabled",         style: .primary)    {}.disabled(true)
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
                                Text("T\(tier)").font(.system(size: 11))
                                    .foregroundStyle(Color.echoSecondaryText)
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
                                .foregroundStyle(Color.echoPrimaryText)
                            Text("Surface card with rounded corners and subtle shadow.")
                                .font(.system(size: 13)).foregroundStyle(Color.echoSecondaryText)
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
                    }.frame(maxWidth: .infinity)
                }
            }
            .padding(24)
        }
        .background(Color.echoBackground.ignoresSafeArea())
        .navigationTitle("Components")
    }

    @ViewBuilder
    func section<C: View>(_ title: String, @ViewBuilder content: () -> C) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(title).font(.system(size: 11, weight: .semibold)).tracking(0.8)
                .foregroundStyle(Color.echoSecondaryText).textCase(.uppercase)
            content()
        }
        Divider()
    }
}

// EchoRippleMark — new canonical mark (replaces centred logo stacking)
struct EchoMarkPreview: View {
    var body: some View {
        ScrollView {
            VStack(spacing: 40) {
                // Light backgrounds
                VStack(spacing: 16) {
                    Text("On paper (default)").font(.echomono(10))
                        .foregroundStyle(Color.echoInk40)
                    HStack(spacing: 28) {
                        ForEach([16, 22, 32, 48, 64] as [CGFloat], id: \.self) { sz in
                            VStack(spacing: 8) {
                                EchoRippleMark(size: sz)
                                Text("\(Int(sz))pt").font(.echomono(9))
                                    .foregroundStyle(Color.echoInk40)
                            }
                        }
                    }
                }
                .padding(24)
                .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 16))

                // Night surface
                VStack(spacing: 16) {
                    Text("On night (private surfaces)").font(.echomono(10))
                        .foregroundStyle(Color.echoNightInk40)
                    HStack(spacing: 28) {
                        ForEach([22, 32, 48] as [CGFloat], id: \.self) { sz in
                            VStack(spacing: 8) {
                                EchoRippleMark(size: sz, color: .echoNightInk)
                                Text("\(Int(sz))pt").font(.echomono(9))
                                    .foregroundStyle(Color.echoNightInk40)
                            }
                        }
                    }
                }
                .padding(24)
                .background(Color.echoNight, in: RoundedRectangle(cornerRadius: 16))

                // Trust variant
                VStack(spacing: 16) {
                    Text("Trust green variant").font(.echomono(10))
                        .foregroundStyle(Color.echoInk40)
                    EchoRippleMark(size: 48, color: .echoTrustGreen)
                }
                .padding(24)
                .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 16))
            }
            .padding(24)
        }
        .background(Color.echoPaperDim.ignoresSafeArea())
        .preferredColorScheme(.light)
        .navigationTitle("EchoRippleMark")
    }
}

// Trust scale: one hue, five intensities
struct TrustScalePreview: View {
    private let steps: [(String, Color, String)] = [
        ("T0 · none",      .echoTrustUnverified, "10%  — not verified"),
        ("T1 · seen",      .echoTrustBasic,      "25%  — profile visible"),
        ("T2 · verified",  .echoTrustVerified,   "45%  — identity confirmed"),
        ("T3 · trusted",   .echoTrustTrusted,    "70%  — community trusted"),
        ("T4 · primary",   .echoTrustElite,      "100% — full trust"),
    ]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("One green. Five opacities.")
                        .font(.system(size: 22, weight: .semibold)).tracking(-0.5)
                        .foregroundStyle(Color.echoInk)
                    Text("Replaces the 6-color tier rainbow (grey/blue/green/purple/amber/pink)\nthat read as a loyalty program.")
                        .font(.system(size: 13)).lineSpacing(3)
                        .foregroundStyle(Color.echoInk55)
                }

                ForEach(steps, id: \.0) { label, color, detail in
                    HStack(spacing: 16) {
                        RoundedRectangle(cornerRadius: 10)
                            .fill(color)
                            .overlay(RoundedRectangle(cornerRadius: 10)
                                .strokeBorder(Color.echoTrustGreen.opacity(0.25), lineWidth: 1))
                            .frame(width: 52, height: 52)
                            .overlay(
                                Text(label.prefix(2))
                                    .font(.echomono(12))
                                    .foregroundStyle(color == .echoTrustElite
                                                     ? .white : Color.echoTrustGreenDim)
                            )

                        VStack(alignment: .leading, spacing: 3) {
                            Text(label)
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundStyle(Color.echoInk)
                            Text(detail)
                                .font(.echomono(11))
                                .foregroundStyle(Color.echoInk40)
                        }
                    }
                }

                Text("Avatar ring tint, header checkmark, and message header all use the same green — no legend required.")
                    .font(.system(size: 12)).lineSpacing(3)
                    .foregroundStyle(Color.echoInk55)
                    .padding(.top, 8)
            }
            .padding(24)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .preferredColorScheme(.light)
        .navigationTitle("Trust Scale")
    }
}

// MARK: - Helpers

extension Array {
    subscript(safe index: Int) -> Element? {
        indices.contains(index) ? self[index] : nil
    }
}
