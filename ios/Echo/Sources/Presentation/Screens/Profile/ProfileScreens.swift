import SwiftUI

import SwiftUI

// MARK: - Screen 1: Profile Tab (Enhanced per spec)

 struct ProfileTabView: View {
    @State private var showEditProfile = false
    @State private var showPersonas = false
    @State private var showSettings = false

    let profile: ProfileData
    let personas: [Persona]
    let maxPersonas: Int
    let onEditProfile: () -> Void
    let onPersonaTapped: (Persona) -> Void
    let onCreatePersona: () -> Void
    let onQRCode: () -> Void
    let onInviteFriends: () -> Void
    let onRewards: () -> Void
    let onAccountSettings: () -> Void
    let onPrivacySettings: () -> Void
    let onNotificationSettings: () -> Void
    let onAppearanceSettings: () -> Void
    let onStorageSettings: () -> Void
    let onHelpSupport: () -> Void
    let onAbout: () -> Void

    init(
        profile: ProfileData = ProfileData(),
        personas: [Persona] = [],
        maxPersonas: Int = 10,
        onEditProfile: @escaping () -> Void = {},
        onPersonaTapped: @escaping (Persona) -> Void = { _ in },
        onCreatePersona: @escaping () -> Void = {},
        onQRCode: @escaping () -> Void = {},
        onInviteFriends: @escaping () -> Void = {},
        onRewards: @escaping () -> Void = {},
        onAccountSettings: @escaping () -> Void = {},
        onPrivacySettings: @escaping () -> Void = {},
        onNotificationSettings: @escaping () -> Void = {},
        onAppearanceSettings: @escaping () -> Void = {},
        onStorageSettings: @escaping () -> Void = {},
        onHelpSupport: @escaping () -> Void = {},
        onAbout: @escaping () -> Void = {}
    ) {
        self.profile = profile
        self.personas = personas
        self.maxPersonas = maxPersonas
        self.onEditProfile = onEditProfile
        self.onPersonaTapped = onPersonaTapped
        self.onCreatePersona = onCreatePersona
        self.onQRCode = onQRCode
        self.onInviteFriends = onInviteFriends
        self.onRewards = onRewards
        self.onAccountSettings = onAccountSettings
        self.onPrivacySettings = onPrivacySettings
        self.onNotificationSettings = onNotificationSettings
        self.onAppearanceSettings = onAppearanceSettings
        self.onStorageSettings = onStorageSettings
        self.onHelpSupport = onHelpSupport
        self.onAbout = onAbout
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "My Profile",
                    showBackButton: false
                )

                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Profile Header
                        profileHeader

                        // Personas Section
                        personasSection

                        // Quick Actions
                        quickActionsSection

                        // Settings
                        settingsSection
                    }
                    .echoSpacing(.lg)
                }
            }
        }
    }

    // MARK: - Profile Header

    private var profileHeader: some View {
        VStack(spacing: Spacing.md.rawValue) {
            AvatarView(
                initials: initials,
                size: .xxl,
                showTrustRing: true,
                trustLevel: profile.trustLevel
            )

            HStack(spacing: Spacing.xs.rawValue) {
                Text(profile.displayName)
                    .typographyStyle(.h2, color: .echoPrimaryText)

                if profile.isVerified {
                    Image(systemName: "checkmark.seal.fill")
                        .foregroundColor(.echoPrimary)
                        .font(.system(size: 18))
                }
            }

            Text("@\(profile.username)")
                .typographyStyle(.body, color: .echoSecondaryText)

            TrustBadge(level: profile.trustLevel, size: .medium)

            EchoButton(
                "Edit Profile",
                style: .secondary,
                size: .medium,
                action: onEditProfile
            )
        }
    }

    // MARK: - Personas Section

    private var personasSection: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            Text("MY PERSONAS")
                .typographyStyle(.caption, color: .echoGray500)
                .textCase(.uppercase)

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: Spacing.md.rawValue) {
                    ForEach(personas) { persona in
                        PersonaCard(persona: persona) {
                            onPersonaTapped(persona)
                        }
                    }

                    if personas.count < maxPersonas {
                        AddPersonaCard(action: onCreatePersona)
                    }
                }
            }
        }
    }

    // MARK: - Quick Actions

    private var quickActionsSection: some View {
        VStack(spacing: Spacing.sm.rawValue) {
            QuickActionRow(
                icon: "qrcode",
                title: "My QR Code",
                action: onQRCode
            )

            QuickActionRow(
                icon: "person.2.fill",
                title: "Invite Friends",
                action: onInviteFriends
            )

            QuickActionRow(
                icon: "gift.fill",
                title: "Echo Rewards",
                subtitle: String(format: "%.1f ECHO available", profile.echoRewards),
                action: onRewards
            )
        }
    }

    // MARK: - Settings Section

    private var settingsSection: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            Text("SETTINGS")
                .typographyStyle(.caption, color: .echoGray500)
                .textCase(.uppercase)

            VStack(spacing: 0) {
                SettingsNavRow(icon: "person.fill", title: "Account", action: onAccountSettings)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "lock.fill", title: "Privacy & Security", action: onPrivacySettings)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "bell.fill", title: "Notifications", action: onNotificationSettings)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "paintbrush.fill", title: "Appearance", action: onAppearanceSettings)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "internaldrive.fill", title: "Storage & Data", action: onStorageSettings)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "questionmark.circle.fill", title: "Help & Support", action: onHelpSupport)
                Divider().padding(.leading, 52)
                SettingsNavRow(icon: "info.circle.fill", title: "About Echo", action: onAbout)
            }
            .background(Color.echoSurface)
            .cornerRadius(12)
        }
    }

    private var initials: String {
        profile.displayName.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

// MARK: - Screen 2: Edit Profile

 struct EditProfileView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var displayName: String
    @Binding var username: String
    @Binding var bio: String
    @Binding var status: String
    @Binding var website: String
    let isUsernameAvailable: Bool?
    let isCheckingUsername: Bool
    let isSaving: Bool
    let onCheckUsername: () -> Void
    let onSave: () -> Void
    let onChangePhoto: () -> Void

     init(
        displayName: Binding<String>,
        username: Binding<String>,
        bio: Binding<String>,
        status: Binding<String>,
        website: Binding<String>,
        isUsernameAvailable: Bool? = nil,
        isCheckingUsername: Bool = false,
        isSaving: Bool = false,
        onCheckUsername: @escaping () -> Void = {},
        onSave: @escaping () -> Void = {},
        onChangePhoto: @escaping () -> Void = {}
    ) {
        self._displayName = displayName
        self._username = username
        self._bio = bio
        self._status = status
        self._website = website
        self.isUsernameAvailable = isUsernameAvailable
        self.isCheckingUsername = isCheckingUsername
        self.isSaving = isSaving
        self.onCheckUsername = onCheckUsername
        self.onSave = onSave
        self.onChangePhoto = onChangePhoto
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                // Nav Bar
                HStack {
                    Button("Cancel") { dismiss() }
                        .foregroundColor(.echoPrimary)

                    Spacer()

                    Text("Edit Profile")
                        .typographyStyle(.h4, color: .echoPrimaryText)

                    Spacer()

                    Button("Save") { onSave() }
                        .foregroundColor(.echoPrimary)
                        .disabled(isSaving)
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, Spacing.md.rawValue)

                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Avatar
                        VStack(spacing: Spacing.sm.rawValue) {
                            AvatarView(initials: avatarInitials, size: .xxl)

                            Button("Change Photo", action: onChangePhoto)
                                .typographyStyle(.body, color: .echoPrimary)
                        }

                        // Form Fields
                        VStack(spacing: Spacing.lg.rawValue) {
                            ProfileTextField(label: "Display Name", text: $displayName)

                            VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                                ProfileTextField(label: "Username", text: $username, prefix: "@")
                                    .onChange(of: username) { _, _ in onCheckUsername() }

                                if isCheckingUsername {
                                    Text("Checking...")
                                        .typographyStyle(.caption, color: .echoGray400)
                                } else if let available = isUsernameAvailable {
                                    Text(available ? "Available" : "Taken")
                                        .typographyStyle(.caption, color: available ? .echoSuccess : .echoError)
                                }
                            }

                            VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                                ProfileTextField(label: "Bio", text: $bio, isMultiline: true)

                                Text("\(bio.count)/\(ProfileViewModel.maxBioLength)")
                                    .typographyStyle(.caption, color: .echoGray400)
                                    .frame(maxWidth: .infinity, alignment: .trailing)
                            }

                            ProfileTextField(label: "Status", text: $status)
                        }

                        // Links Section
                        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                            Text("LINKS")
                                .typographyStyle(.caption, color: .echoGray500)
                                .textCase(.uppercase)

                            ProfileTextField(label: "Website", text: $website)
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }

    private var avatarInitials: String {
        displayName.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

// MARK: - Screen 3: Personas Management

 struct PersonasManagementView: View {
    @Environment(\.dismiss) var dismiss
    let personas: [Persona]
    let maxPersonas: Int
    let onEditPersona: (Persona) -> Void
    let onDeletePersona: (String) -> Void
    let onSetDefault: (String) -> Void
    let onCreatePersona: () -> Void

     init(
        personas: [Persona] = [],
        maxPersonas: Int = 10,
        onEditPersona: @escaping (Persona) -> Void = { _ in },
        onDeletePersona: @escaping (String) -> Void = { _ in },
        onSetDefault: @escaping (String) -> Void = { _ in },
        onCreatePersona: @escaping () -> Void = {}
    ) {
        self.personas = personas
        self.maxPersonas = maxPersonas
        self.onEditPersona = onEditPersona
        self.onDeletePersona = onDeletePersona
        self.onSetDefault = onSetDefault
        self.onCreatePersona = onCreatePersona
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "My Personas",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Description
                        Text("Personas let you show different sides of yourself to different contacts.")
                            .typographyStyle(.body, color: .echoSecondaryText)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, Spacing.lg.rawValue)

                        if personas.isEmpty {
                            // Empty State
                            emptyState
                        } else {
                            // Active Personas
                            VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                                Text("ACTIVE PERSONAS (\(personas.count)/\(maxPersonas))")
                                    .typographyStyle(.caption, color: .echoGray500)
                                    .textCase(.uppercase)

                                ForEach(personas) { persona in
                                    PersonaDetailCard(
                                        persona: persona,
                                        onEdit: { onEditPersona(persona) },
                                        onDelete: { onDeletePersona(persona.id) },
                                        onSetDefault: { onSetDefault(persona.id) }
                                    )
                                }
                            }

                            // Create Button
                            if personas.count < maxPersonas {
                                Button(action: onCreatePersona) {
                                    HStack {
                                        Image(systemName: "plus")
                                        Text("Create New Persona")
                                    }
                                    .typographyStyle(.body, color: .echoPrimary)
                                    .frame(maxWidth: .infinity)
                                    .padding(Spacing.lg.rawValue)
                                    .background(Color.echoSurface)
                                    .cornerRadius(12)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 12)
                                            .stroke(Color.echoGray200, lineWidth: 1)
                                    )
                                }

                                let remaining = maxPersonas - personas.count
                                Text("You can create up to \(maxPersonas) personas. \(remaining) remaining.")
                                    .typographyStyle(.caption, color: .echoGray400)
                                    .multilineTextAlignment(.center)
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }

    private var emptyState: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Image(systemName: "theatermasks.fill")
                .font(.system(size: 48))
                .foregroundColor(.echoGray300)

            Text("Express different sides of yourself")
                .typographyStyle(.h3, color: .echoPrimaryText)

            Text("Create personas to show different profiles to different people")
                .typographyStyle(.body, color: .echoSecondaryText)
                .multilineTextAlignment(.center)

            EchoButton(
                "Create Your First Persona",
                style: .primary,
                size: .large,
                action: onCreatePersona
            )
        }
        .padding(Spacing.xl.rawValue)
    }
}

// MARK: - Screen 4: Create/Edit Persona

 struct CreateEditPersonaView: View {
    @Environment(\.dismiss) var dismiss
    @State var selectedType: PersonaType
    @State var personaName: String
    @State var displayName: String
    @State var bio: String
    @State var useMainAvatar: Bool
    @State var visibility: PersonaVisibility
    @State var selectedContactIds: [String]

    let isEditing: Bool
    let isSaving: Bool
    let onSave: (PersonaType, String, String, String, Bool, PersonaVisibility, [String]) -> Void

     init(
        persona: Persona? = nil,
        isSaving: Bool = false,
        onSave: @escaping (PersonaType, String, String, String, Bool, PersonaVisibility, [String]) -> Void = { _, _, _, _, _, _, _ in }
    ) {
        self.isEditing = persona != nil
        self.isSaving = isSaving
        self.onSave = onSave
        _selectedType = State(initialValue: persona?.type ?? .professional)
        _personaName = State(initialValue: persona?.name ?? "")
        _displayName = State(initialValue: persona?.displayName ?? "")
        _bio = State(initialValue: persona?.bio ?? "")
        _useMainAvatar = State(initialValue: persona?.useMainAvatar ?? true)
        _visibility = State(initialValue: persona?.visibility ?? .all)
        _selectedContactIds = State(initialValue: persona?.selectedContactIds ?? [])
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                // Nav Bar
                HStack {
                    Button("Cancel") { dismiss() }
                        .foregroundColor(.echoPrimary)

                    Spacer()

                    Text(isEditing ? "Edit Persona" : "New Persona")
                        .typographyStyle(.h4, color: .echoPrimaryText)

                    Spacer()

                    Button("Save") {
                        onSave(selectedType, personaName, displayName, bio, useMainAvatar, visibility, selectedContactIds)
                    }
                    .foregroundColor(.echoPrimary)
                    .disabled(isSaving || personaName.isEmpty || displayName.isEmpty)
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, Spacing.md.rawValue)

                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Persona Type Selection
                        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                            Text("PERSONA TYPE")
                                .typographyStyle(.caption, color: .echoGray500)
                                .textCase(.uppercase)

                            LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 4), spacing: Spacing.md.rawValue) {
                                ForEach(PersonaType.allCases) { type in
                                    PersonaTypeSelector(
                                        type: type,
                                        isSelected: selectedType == type
                                    ) {
                                        selectedType = type
                                    }
                                }
                            }
                        }

                        // Form Fields
                        VStack(spacing: Spacing.lg.rawValue) {
                            ProfileTextField(label: "Persona Name", text: $personaName)

                            VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                                ProfileTextField(label: "Display Name", text: $displayName)
                                Text("This is what contacts see")
                                    .typographyStyle(.caption, color: .echoGray400)
                            }

                            // Avatar Choice
                            VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                                Text("Avatar")
                                    .typographyStyle(.body, color: .echoGray500)

                                VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                    RadioRow(
                                        title: "Use main avatar",
                                        isSelected: useMainAvatar
                                    ) { useMainAvatar = true }

                                    RadioRow(
                                        title: "Use different avatar",
                                        isSelected: !useMainAvatar
                                    ) { useMainAvatar = false }
                                }
                            }

                            ProfileTextField(label: "Bio", text: $bio, isMultiline: true)
                        }

                        // Visibility
                        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                            Text("VISIBILITY")
                                .typographyStyle(.caption, color: .echoGray500)
                                .textCase(.uppercase)

                            Text("Who can see this persona?")
                                .typographyStyle(.body, color: .echoPrimaryText)

                            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                RadioRow(
                                    title: "All contacts",
                                    isSelected: visibility == .all
                                ) { visibility = .all }

                                RadioRow(
                                    title: "Selected contacts only",
                                    isSelected: visibility == .selected
                                ) { visibility = .selected }

                                RadioRow(
                                    title: "No one (hidden)",
                                    isSelected: visibility == .hidden
                                ) { visibility = .hidden }
                            }

                            if visibility == .selected {
                                Button(action: {}) {
                                    HStack {
                                        Text("Select Contacts")
                                        Spacer()
                                        Text("\(selectedContactIds.count) selected")
                                            .foregroundColor(.echoGray400)
                                        Image(systemName: "chevron.right")
                                            .foregroundColor(.echoGray400)
                                    }
                                    .padding(Spacing.lg.rawValue)
                                    .background(Color.echoSurface)
                                    .cornerRadius(12)
                                }
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Screen 5: Account Settings

 struct AccountSettingsView: View {
    @Environment(\.dismiss) var dismiss
    let account: AccountInfo
    let onDeleteAccount: () -> Void
    @State private var showDeleteAlert = false

     init(
        account: AccountInfo = AccountInfo(),
        onDeleteAccount: @escaping () -> Void = {}
    ) {
        self.account = account
        self.onDeleteAccount = onDeleteAccount
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Account",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Identity
                        SettingsSectionView(title: "Identity") {
                            SettingsListItem(
                                icon: Image(systemName: "phone.fill"),
                                title: "Phone Number",
                                value: maskedPhone
                            )

                            SettingsListItem(
                                icon: Image(systemName: "envelope.fill"),
                                title: "Email (optional)",
                                value: account.email ?? "Not set"
                            )

                            if let did = account.did {
                                HStack {
                                    Image(systemName: "link")
                                        .foregroundColor(.echoGray500)
                                    VStack(alignment: .leading, spacing: 2) {
                                        Text("DID")
                                            .typographyStyle(.caption, color: .echoGray500)
                                        Text(truncatedDID(did))
                                            .typographyStyle(.body, color: .echoPrimaryText)
                                            .lineLimit(1)
                                    }
                                    Spacer()
                                    Button(action: {
                                        #if os(iOS)
                                        UIPasteboard.general.string = did
                                        #elseif os(macOS)
                                        NSPasteboard.general.setString(did, forType: .string)
                                        #endif
                                    }) {
                                        Image(systemName: "doc.on.doc")
                                            .foregroundColor(.echoGray400)
                                    }
                                }
                                .padding(Spacing.lg.rawValue)
                            }
                        }

                        // Security
                        SettingsSectionView(title: "Security") {
                            SettingsListItem(
                                icon: Image(systemName: "key.fill"),
                                title: "Passkeys",
                                value: "\(account.passkeyCount) passkeys registered"
                            )

                            SettingsListItem(
                                icon: Image(systemName: "lock.shield.fill"),
                                title: "Two-Factor Authentication",
                                value: account.twoFactorEnabled ? "Enabled" : "Disabled"
                            )

                            SettingsListItem(
                                icon: Image(systemName: "desktopcomputer"),
                                title: "Active Sessions",
                                value: "\(account.activeSessionCount) devices"
                            )
                        }

                        // Recovery
                        SettingsSectionView(title: "Recovery") {
                            SettingsListItem(
                                icon: Image(systemName: "doc.text.fill"),
                                title: "Recovery Phrase",
                                value: account.recoveryPhraseSetUp ? "Set up" : "Not set up"
                            )

                            SettingsListItem(
                                icon: Image(systemName: "person.2.fill"),
                                title: "Trusted Recovery Contacts",
                                value: "\(account.trustedRecoveryContactCount) contacts set up"
                            )
                        }

                        // Danger Zone
                        SettingsSectionView(title: "Danger Zone") {
                            Button(action: { showDeleteAlert = true }) {
                                HStack {
                                    Image(systemName: "trash.fill")
                                        .foregroundColor(Color(hex: 0xDC2626))
                                    Text("Delete Account")
                                        .typographyStyle(.body, color: Color(hex: 0xDC2626))
                                    Spacer()
                                }
                                .padding(Spacing.lg.rawValue)
                                .background(Color(hex: 0xFEE2E2))
                                .cornerRadius(12)
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
        .alert("Delete Account", isPresented: $showDeleteAlert) {
            Button("Cancel", role: .cancel) {}
            Button("Delete", role: .destructive) { onDeleteAccount() }
        } message: {
            Text("This action is permanent and cannot be undone. All your data will be lost.")
        }
    }

    private var maskedPhone: String {
        guard !account.phone.isEmpty else { return "Not set" }
        let last4 = String(account.phone.suffix(4))
        return "+1 \u{2022}\u{2022}\u{2022}\u{2022}\u{2022}\u{2022}\u{2022}\(last4)"
    }

    private func truncatedDID(_ did: String) -> String {
        guard did.count > 20 else { return did }
        return "\(did.prefix(16))...\(did.suffix(3))"
    }
}

// MARK: - Screen 6: Privacy & Security Settings

 struct PrivacySecuritySettingsView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var settings: EnhancedPrivacySettings
    let onSave: (EnhancedPrivacySettings) -> Void

     init(
        settings: Binding<EnhancedPrivacySettings>,
        onSave: @escaping (EnhancedPrivacySettings) -> Void = { _ in }
    ) {
        self._settings = settings
        self.onSave = onSave
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Privacy & Security",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Profile Visibility
                        SettingsSectionView(title: "Profile Visibility") {
                            SettingsListItem(
                                icon: Image(systemName: "magnifyingglass"),
                                title: "Who can find me by username",
                                value: settings.findByUsername.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "circle.fill"),
                                title: "Who can see my online status",
                                value: settings.showOnlineStatus.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "shield.fill"),
                                title: "Who can see my trust score",
                                value: settings.showTrustScore.capitalized
                            )
                        }

                        // Messaging
                        SettingsSectionView(title: "Messaging") {
                            SettingsListItem(
                                icon: Image(systemName: "message.fill"),
                                title: "Who can message me",
                                value: settings.whoCanMessage.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "eye.fill"),
                                title: "Read receipts",
                                hasToggle: true,
                                toggleValue: $settings.readReceipts
                            )

                            SettingsListItem(
                                icon: Image(systemName: "text.cursor"),
                                title: "Typing indicators",
                                hasToggle: true,
                                toggleValue: $settings.typingIndicators
                            )
                        }

                        // Calls
                        SettingsSectionView(title: "Calls") {
                            SettingsListItem(
                                icon: Image(systemName: "phone.fill"),
                                title: "Who can call me",
                                value: settings.whoCanCall.capitalized
                            )
                        }

                        // Blockchain
                        SettingsSectionView(title: "Blockchain") {
                            SettingsListItem(
                                icon: Image(systemName: "link"),
                                title: "Anchor messages by default",
                                hasToggle: true,
                                toggleValue: $settings.anchorMessagesByDefault
                            )

                            SettingsListItem(
                                icon: Image(systemName: "checkmark.seal.fill"),
                                title: "Show verification badges",
                                hasToggle: true,
                                toggleValue: $settings.showVerificationBadges
                            )
                        }

                        // Security
                        SettingsSectionView(title: "Security") {
                            SettingsListItem(
                                icon: Image(systemName: "lock.fill"),
                                title: "Screen lock for app",
                                value: settings.screenLockTimeout.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "eye.slash.fill"),
                                title: "Hide message previews",
                                hasToggle: true,
                                toggleValue: $settings.hideMessagePreviews
                            )

                            SettingsListItem(
                                icon: Image(systemName: "camera.viewfinder"),
                                title: "Screenshot notifications",
                                hasToggle: true,
                                toggleValue: $settings.screenshotNotifications
                            )
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Screen 7: Notification Settings

 struct NotificationSettingsView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var settings: EnhancedNotificationSettings
    let onSave: (EnhancedNotificationSettings) -> Void

     init(
        settings: Binding<EnhancedNotificationSettings>,
        onSave: @escaping (EnhancedNotificationSettings) -> Void = { _ in }
    ) {
        self._settings = settings
        self.onSave = onSave
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Notifications",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Messages
                        SettingsSectionView(title: "Messages") {
                            SettingsListItem(
                                icon: Image(systemName: "message.fill"),
                                title: "Message notifications",
                                hasToggle: true,
                                toggleValue: $settings.messageNotifications
                            )

                            SettingsListItem(
                                icon: Image(systemName: "eye.fill"),
                                title: "Show previews",
                                value: settings.showPreviews.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "speaker.wave.2.fill"),
                                title: "Sound",
                                value: settings.messageSound
                            )
                        }

                        // Groups
                        SettingsSectionView(title: "Groups") {
                            SettingsListItem(
                                icon: Image(systemName: "person.3.fill"),
                                title: "Group notifications",
                                hasToggle: true,
                                toggleValue: $settings.groupNotifications
                            )

                            SettingsListItem(
                                icon: Image(systemName: "at"),
                                title: "@Mentions only",
                                hasToggle: true,
                                toggleValue: $settings.mentionsOnly
                            )
                        }

                        // Calls
                        SettingsSectionView(title: "Calls") {
                            SettingsListItem(
                                icon: Image(systemName: "phone.fill"),
                                title: "Call notifications",
                                hasToggle: true,
                                toggleValue: $settings.callNotifications
                            )

                            SettingsListItem(
                                icon: Image(systemName: "music.note"),
                                title: "Ring sound",
                                value: settings.ringSound
                            )
                        }

                        // Other
                        SettingsSectionView(title: "Other") {
                            SettingsListItem(
                                icon: Image(systemName: "person.badge.plus"),
                                title: "Contact requests",
                                hasToggle: true,
                                toggleValue: $settings.contactRequests
                            )

                            SettingsListItem(
                                icon: Image(systemName: "shield.fill"),
                                title: "Trust score changes",
                                hasToggle: true,
                                toggleValue: $settings.trustScoreChanges
                            )

                            SettingsListItem(
                                icon: Image(systemName: "gift.fill"),
                                title: "Reward notifications",
                                hasToggle: true,
                                toggleValue: $settings.rewardNotifications
                            )
                        }

                        // Quiet Hours
                        SettingsSectionView(title: "Quiet Hours") {
                            SettingsListItem(
                                icon: Image(systemName: "moon.fill"),
                                title: "Enable quiet hours",
                                hasToggle: true,
                                toggleValue: $settings.quietHoursEnabled
                            )

                            if settings.quietHoursEnabled {
                                SettingsListItem(
                                    icon: Image(systemName: "clock.fill"),
                                    title: "From",
                                    value: settings.quietHoursFrom
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "clock.fill"),
                                    title: "To",
                                    value: settings.quietHoursTo
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "phone.arrow.up.right"),
                                    title: "Allow calls from Inner Circle",
                                    hasToggle: true,
                                    toggleValue: $settings.allowInnerCircleCalls
                                )
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Screen 8: Appearance Settings

 struct AppearanceSettingsView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var settings: AppearanceSettings
    let onSave: (AppearanceSettings) -> Void

     init(
        settings: Binding<AppearanceSettings>,
        onSave: @escaping (AppearanceSettings) -> Void = { _ in }
    ) {
        self._settings = settings
        self.onSave = onSave
    }

    private let accentColors = [
        ("Indigo", "indigo", Color.echoPrimary),
        ("Purple", "purple", Color(hex: 0x8B5CF6)),
        ("Blue", "blue", Color(hex: 0x3B82F6)),
        ("Green", "green", Color(hex: 0x10B981)),
        ("Orange", "orange", Color(hex: 0xF59E0B))
    ]

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Appearance",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Theme
                        SettingsSectionView(title: "Theme") {
                            HStack(spacing: Spacing.md.rawValue) {
                                ThemeOption(
                                    icon: "sun.max.fill",
                                    label: "Light",
                                    isSelected: settings.theme == "light"
                                ) { settings.theme = "light" }

                                ThemeOption(
                                    icon: "moon.fill",
                                    label: "Dark",
                                    isSelected: settings.theme == "dark"
                                ) { settings.theme = "dark" }

                                ThemeOption(
                                    icon: "iphone",
                                    label: "System",
                                    isSelected: settings.theme == "system"
                                ) { settings.theme = "system" }
                            }
                        }

                        // Accent Color
                        SettingsSectionView(title: "Accent Color") {
                            ForEach(accentColors, id: \.1) { name, key, color in
                                Button(action: { settings.accentColor = key }) {
                                    HStack(spacing: Spacing.md.rawValue) {
                                        Circle()
                                            .fill(color)
                                            .frame(width: 24, height: 24)

                                        Text(name + (key == "indigo" ? " (default)" : ""))
                                            .typographyStyle(.body, color: .echoPrimaryText)

                                        Spacer()

                                        if settings.accentColor == key {
                                            Image(systemName: "checkmark")
                                                .foregroundColor(.echoPrimary)
                                        }
                                    }
                                    .padding(.vertical, Spacing.sm.rawValue)
                                }
                            }
                        }

                        // Chat
                        SettingsSectionView(title: "Chat") {
                            SettingsListItem(
                                icon: Image(systemName: "photo.fill"),
                                title: "Chat wallpaper",
                                value: settings.chatWallpaper.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "bubble.left.fill"),
                                title: "Message corners",
                                value: settings.messageCorners.capitalized
                            )

                            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                Text("Font size")
                                    .typographyStyle(.body, color: .echoPrimaryText)
                                    .padding(.horizontal, Spacing.lg.rawValue)

                                Picker("Font size", selection: $settings.fontSize) {
                                    Text("Small").tag("small")
                                    Text("Medium").tag("medium")
                                    Text("Large").tag("large")
                                }
                                .pickerStyle(.segmented)
                                .padding(.horizontal, Spacing.lg.rawValue)
                            }
                        }

                        // App Icon
                        SettingsSectionView(title: "App Icon") {
                            HStack(spacing: Spacing.md.rawValue) {
                                AppIconOption(color: .echoPrimary, label: "Default", isSelected: settings.appIcon == "default") {
                                    settings.appIcon = "default"
                                }
                                AppIconOption(color: .echoGray900, label: "Dark", isSelected: settings.appIcon == "dark") {
                                    settings.appIcon = "dark"
                                }
                                AppIconOption(color: .echoGray200, label: "Light", isSelected: settings.appIcon == "light") {
                                    settings.appIcon = "light"
                                }
                                AppIconOption(color: .echoWarning, label: "Gold", isPremium: true, isSelected: settings.appIcon == "gold") {
                                    settings.appIcon = "gold"
                                }
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Screen 9: Storage & Data

 struct StorageDataView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var storageInfo: StorageInfo
    let onClearCache: () -> Void
    let onBackUp: () -> Void
    @State private var showClearCacheAlert = false

     init(
        storageInfo: Binding<StorageInfo>,
        onClearCache: @escaping () -> Void = {},
        onBackUp: @escaping () -> Void = {}
    ) {
        self._storageInfo = storageInfo
        self.onClearCache = onClearCache
        self.onBackUp = onBackUp
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Storage & Data",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Storage Used
                        SettingsSectionView(title: "Storage Used") {
                            VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                                StorageBar(
                                    used: storageInfo.totalUsedBytes,
                                    total: storageInfo.totalCapacityBytes
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "photo.fill"),
                                    title: "Photos & Videos",
                                    value: formatBytes(storageInfo.photosVideosBytes)
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "doc.fill"),
                                    title: "Documents",
                                    value: formatBytes(storageInfo.documentsBytes)
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "waveform"),
                                    title: "Voice Messages",
                                    value: formatBytes(storageInfo.voiceMessagesBytes)
                                )

                                SettingsListItem(
                                    icon: Image(systemName: "ellipsis.circle.fill"),
                                    title: "Other",
                                    value: formatBytes(storageInfo.otherBytes)
                                )
                            }
                        }

                        // Manage Storage
                        SettingsSectionView(title: "Manage Storage") {
                            SettingsListItem(
                                icon: Image(systemName: "arrow.down.circle.fill"),
                                title: "Auto-download media",
                                value: storageInfo.autoDownload.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "sparkles"),
                                title: "Media quality",
                                value: storageInfo.mediaQuality.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "clock.fill"),
                                title: "Keep media",
                                value: storageInfo.keepMedia.capitalized
                            )

                            Button(action: { showClearCacheAlert = true }) {
                                HStack {
                                    Image(systemName: "trash.fill")
                                        .foregroundColor(.echoError)
                                    VStack(alignment: .leading) {
                                        Text("Clear Cache")
                                            .typographyStyle(.body, color: .echoError)
                                        Text(formatBytes(storageInfo.cacheBytes))
                                            .typographyStyle(.caption, color: .echoGray400)
                                    }
                                    Spacer()
                                }
                                .padding(Spacing.lg.rawValue)
                            }
                        }

                        // Network
                        SettingsSectionView(title: "Network") {
                            SettingsListItem(
                                icon: Image(systemName: "antenna.radiowaves.left.and.right"),
                                title: "Use less data for calls",
                                hasToggle: true,
                                toggleValue: $storageInfo.useLessDataForCalls
                            )

                            SettingsListItem(
                                icon: Image(systemName: "network"),
                                title: "Proxy settings"
                            )
                        }

                        // Backup
                        SettingsSectionView(title: "Backup") {
                            Button(action: onBackUp) {
                                HStack {
                                    Image(systemName: "icloud.fill")
                                        .foregroundColor(.echoPrimary)
                                    VStack(alignment: .leading) {
                                        Text("Back Up Now")
                                            .typographyStyle(.body, color: .echoPrimary)
                                        if let date = storageInfo.lastBackupDate {
                                            Text("Last: \(date.formatted())")
                                                .typographyStyle(.caption, color: .echoGray400)
                                        }
                                    }
                                    Spacer()
                                }
                                .padding(Spacing.lg.rawValue)
                            }

                            SettingsListItem(
                                icon: Image(systemName: "arrow.clockwise"),
                                title: "Auto backup",
                                value: storageInfo.autoBackup.capitalized
                            )

                            SettingsListItem(
                                icon: Image(systemName: "photo.on.rectangle"),
                                title: "Include media",
                                hasToggle: true,
                                toggleValue: $storageInfo.includeMediaInBackup
                            )
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
        .alert("Clear Cache", isPresented: $showClearCacheAlert) {
            Button("Cancel", role: .cancel) {}
            Button("Clear", role: .destructive) { onClearCache() }
        } message: {
            Text("This will free up \(formatBytes(storageInfo.cacheBytes)) of space.")
        }
    }

    private func formatBytes(_ bytes: Int64) -> String {
        let formatter = ByteCountFormatter()
        formatter.allowedUnits = [.useMB, .useGB]
        formatter.countStyle = .file
        return formatter.string(fromByteCount: bytes)
    }
}

// MARK: - Screen 10: About Echo

 struct AboutView: View {
    @Environment(\.dismiss) var dismiss

    let appVersion: String
    let onHelpCenter: () -> Void
    let onContactSupport: () -> Void
    let onReportProblem: () -> Void
    let onTerms: () -> Void
    let onPrivacyPolicy: () -> Void
    let onOpenSource: () -> Void

     init(
        appVersion: String = "1.0.0",
        onHelpCenter: @escaping () -> Void = {},
        onContactSupport: @escaping () -> Void = {},
        onReportProblem: @escaping () -> Void = {},
        onTerms: @escaping () -> Void = {},
        onPrivacyPolicy: @escaping () -> Void = {},
        onOpenSource: @escaping () -> Void = {}
    ) {
        self.appVersion = appVersion
        self.onHelpCenter = onHelpCenter
        self.onContactSupport = onContactSupport
        self.onReportProblem = onReportProblem
        self.onTerms = onTerms
        self.onPrivacyPolicy = onPrivacyPolicy
        self.onOpenSource = onOpenSource
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "About",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Logo & Version
                        VStack(spacing: Spacing.sm.rawValue) {
                            Image(systemName: "waveform.circle.fill")
                                .font(.system(size: 64))
                                .foregroundColor(.echoPrimary)

                            Text("Echo")
                                .typographyStyle(.h1, color: .echoPrimaryText)

                            Text("Version \(appVersion)")
                                .typographyStyle(.body, color: .echoSecondaryText)
                        }
                        .padding(.vertical, Spacing.xl.rawValue)

                        // Support
                        SettingsSectionView(title: "Support") {
                            SettingsNavRow(icon: "questionmark.circle.fill", title: "Help Center", action: onHelpCenter)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "headphones", title: "Contact Support", action: onContactSupport)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "exclamationmark.bubble.fill", title: "Report a Problem", action: onReportProblem)
                        }

                        // Legal
                        SettingsSectionView(title: "Legal") {
                            SettingsNavRow(icon: "doc.text.fill", title: "Terms of Service", action: onTerms)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "hand.raised.fill", title: "Privacy Policy", action: onPrivacyPolicy)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "chevron.left.forwardslash.chevron.right", title: "Open Source Licenses", action: onOpenSource)
                        }

                        // Developer
                        SettingsSectionView(title: "Developer") {
                            SettingsNavRow(icon: "doc.plaintext.fill", title: "API Documentation", action: {})
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "cpu", title: "Create a Bot", action: {})
                        }

                        // Footer
                        VStack(spacing: Spacing.sm.rawValue) {
                            Text("Made with \u{1F49C} by the Echo Team")
                                .typographyStyle(.caption, color: .echoGray400)

                            HStack(spacing: Spacing.md.rawValue) {
                                Label("Powered by Cardano", systemImage: "link")
                                    .typographyStyle(.caption, color: .echoGray400)
                            }
                        }
                        .padding(.vertical, Spacing.xl.rawValue)
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Persona Switcher View

 struct PersonaSwitcherView: View {
    let personas: [Persona]
    let activePersonaId: String?
    let contactName: String
    let onSelectPersona: (String) -> Void
    let onCreatePersona: () -> Void

     init(
        personas: [Persona] = [],
        activePersonaId: String? = nil,
        contactName: String = "",
        onSelectPersona: @escaping (String) -> Void = { _ in },
        onCreatePersona: @escaping () -> Void = {}
    ) {
        self.personas = personas
        self.activePersonaId = activePersonaId
        self.contactName = contactName
        self.onSelectPersona = onSelectPersona
        self.onCreatePersona = onCreatePersona
    }

     var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Message as:")
                .typographyStyle(.caption, color: .echoGray500)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.top, Spacing.md.rawValue)
                .padding(.bottom, Spacing.sm.rawValue)

            ForEach(personas.filter { $0.deletionState == nil }) { persona in
                let isActive = persona.id == activePersonaId
                let contactCanSee = persona.visibility == .all ||
                    persona.selectedContactIds.contains(where: { _ in true })

                Button(action: { onSelectPersona(persona.id) }) {
                    HStack(spacing: Spacing.md.rawValue) {
                        Image(systemName: isActive ? "largecircle.fill.circle" : "circle")
                            .foregroundColor(isActive ? .echoPrimary : .echoGray400)

                        Text(persona.type.emoji)
                            .font(.system(size: 20))

                        VStack(alignment: .leading, spacing: 2) {
                            HStack(spacing: Spacing.xs.rawValue) {
                                Text(persona.displayName)
                                    .typographyStyle(.body, color: .echoPrimaryText)

                                if isActive {
                                    Text("Current")
                                        .typographyStyle(.tiny, color: .echoPrimary)
                                        .padding(.horizontal, 6)
                                        .padding(.vertical, 2)
                                        .background(Color.echoPrimary.opacity(0.1))
                                        .cornerRadius(4)
                                }
                            }

                            if contactCanSee {
                                Text("\(contactName) knows this persona")
                                    .typographyStyle(.caption, color: .echoGray400)
                            } else {
                                Text("\(contactName) doesn't know this persona")
                                    .typographyStyle(.caption, color: .echoWarning)
                            }
                        }

                        Spacer()
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, Spacing.md.rawValue)
                }

                if persona.id != personas.last?.id {
                    Divider().padding(.leading, 60)
                }
            }

            Divider()

            Button(action: onCreatePersona) {
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: "plus")
                        .foregroundColor(.echoPrimary)
                    Text("Create New Persona")
                        .typographyStyle(.body, color: .echoPrimary)
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, Spacing.md.rawValue)
            }
        }
        .background(Color.echoSurface)
        .cornerRadius(16)
        .shadow(color: Color.Echo.onSurface.opacity(0.1), radius: 8, y: 4)
    }
}

// MARK: - Persona Switch Warning View

 struct PersonaSwitchWarningView: View {
    let context: PersonaSwitchContext
    let onConfirm: () -> Void
    let onCancel: () -> Void

     var body: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 36))
                .foregroundColor(.echoWarning)

            Text("Switch Persona?")
                .typographyStyle(.h3, color: .echoPrimaryText)

            if let warning = context.warningMessage {
                Text(warning)
                    .typographyStyle(.body, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)
            }

            if !context.contactKnowsLink {
                HStack(spacing: Spacing.sm.rawValue) {
                    Image(systemName: "eye.slash.fill")
                        .foregroundColor(.echoWarning)
                    Text("This contact does not know these personas are linked. Switching may reveal your identity.")
                        .typographyStyle(.caption, color: .echoWarning)
                }
                .padding(Spacing.md.rawValue)
                .background(Color.echoWarning.opacity(0.1))
                .cornerRadius(8)
            }

            HStack(spacing: Spacing.md.rawValue) {
                EchoButton("Cancel", style: .secondary, size: .medium, action: onCancel)
                EchoButton("Continue (New Thread)", style: .primary, size: .medium, action: onConfirm)
            }
        }
        .padding(Spacing.xl.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(16)
    }
}

// MARK: - Visibility Matrix View

 struct VisibilityMatrixView: View {
    @Environment(\.dismiss) var dismiss
    let personas: [Persona]
    let matrixEntries: [VisibilityMatrixEntry]
    let onToggleVisibility: (String, String, Bool) -> Void

     init(
        personas: [Persona] = [],
        matrixEntries: [VisibilityMatrixEntry] = [],
        onToggleVisibility: @escaping (String, String, Bool) -> Void = { _, _, _ in }
    ) {
        self.personas = personas
        self.matrixEntries = matrixEntries
        self.onToggleVisibility = onToggleVisibility
    }

    private var activePersonas: [Persona] {
        personas.filter { $0.deletionState == nil }
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Persona Visibility",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        Text("Who can see each persona:")
                            .typographyStyle(.body, color: .echoSecondaryText)

                        // Matrix Header
                        ScrollView(.horizontal, showsIndicators: false) {
                            VStack(spacing: 0) {
                                // Header row
                                HStack(spacing: 0) {
                                    Text("Contact")
                                        .typographyStyle(.caption, color: .echoGray500)
                                        .frame(width: 120, alignment: .leading)

                                    ForEach(activePersonas) { persona in
                                        VStack(spacing: 2) {
                                            Text(persona.type.emoji)
                                                .font(.system(size: 16))
                                            Text(persona.type.label)
                                                .typographyStyle(.tiny, color: .echoGray500)
                                                .lineLimit(1)
                                        }
                                        .frame(width: 72)
                                    }
                                }
                                .padding(.vertical, Spacing.sm.rawValue)
                                .background(Color.echoSurface)

                                Divider()

                                // Contact rows
                                ForEach(matrixEntries) { entry in
                                    HStack(spacing: 0) {
                                        Text(entry.contactName)
                                            .typographyStyle(.body, color: .echoPrimaryText)
                                            .frame(width: 120, alignment: .leading)
                                            .lineLimit(1)

                                        ForEach(activePersonas) { persona in
                                            let isVisible = entry.personaVisibility[persona.id] ?? false
                                            Button(action: {
                                                onToggleVisibility(entry.contactId, persona.id, !isVisible)
                                            }) {
                                                Image(systemName: isVisible ? "checkmark.circle.fill" : "circle")
                                                    .foregroundColor(isVisible ? .echoPrimary : .echoGray300)
                                                    .font(.system(size: 20))
                                            }
                                            .frame(width: 72)
                                        }
                                    }
                                    .padding(.vertical, Spacing.sm.rawValue)

                                    Divider()
                                }
                            }
                            .padding(.horizontal, Spacing.lg.rawValue)
                        }

                        if matrixEntries.isEmpty {
                            VStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "person.3.fill")
                                    .font(.system(size: 36))
                                    .foregroundColor(.echoGray300)
                                Text("No contacts yet")
                                    .typographyStyle(.body, color: .echoSecondaryText)
                                Text("Add contacts to manage persona visibility")
                                    .typographyStyle(.caption, color: .echoGray400)
                            }
                            .padding(Spacing.xl.rawValue)
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Per-Persona Privacy Settings View

 struct PersonaPrivacySettingsView: View {
    @Environment(\.dismiss) var dismiss
    let personaName: String
    @Binding var settings: PersonaPrivacySettings
    let onSave: (PersonaPrivacySettings) -> Void

     init(
        personaName: String = "",
        settings: Binding<PersonaPrivacySettings>,
        onSave: @escaping (PersonaPrivacySettings) -> Void = { _ in }
    ) {
        self.personaName = personaName
        self._settings = settings
        self.onSave = onSave
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Privacy: \(personaName)",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        SettingsSectionView(title: "Visibility") {
                            SettingsListItem(icon: Image(systemName: "clock.fill"), title: "Last seen", value: settings.lastSeenVisibility.capitalized)
                            SettingsListItem(icon: Image(systemName: "circle.fill"), title: "Online status", value: settings.onlineStatusVisibility.capitalized)
                            SettingsListItem(icon: Image(systemName: "person.crop.circle.fill"), title: "Profile picture", value: settings.profilePictureVisibility.capitalized)
                            SettingsListItem(icon: Image(systemName: "text.alignleft"), title: "Bio", value: settings.bioVisibility.capitalized)
                        }

                        SettingsSectionView(title: "Interactions") {
                            SettingsListItem(icon: Image(systemName: "message.fill"), title: "Who can message", value: settings.whoCanMessage.capitalized)
                            SettingsListItem(icon: Image(systemName: "phone.fill"), title: "Who can call", value: settings.whoCanCall.capitalized)
                            SettingsListItem(icon: Image(systemName: "person.3.fill"), title: "Who can add to groups", value: settings.whoCanAddToGroups.capitalized)
                            SettingsListItem(icon: Image(systemName: "checkmark.circle.fill"), title: "Require approval", hasToggle: true, toggleValue: $settings.requireApprovalForContact)
                        }

                        SettingsSectionView(title: "Read Receipts") {
                            SettingsListItem(icon: Image(systemName: "eye.fill"), title: "Send read receipts", hasToggle: true, toggleValue: $settings.sendReadReceipts)
                            SettingsListItem(icon: Image(systemName: "text.cursor"), title: "Typing indicators", hasToggle: true, toggleValue: $settings.sendTypingIndicators)
                        }

                        SettingsSectionView(title: "Discovery") {
                            SettingsListItem(icon: Image(systemName: "magnifyingglass"), title: "Searchable", hasToggle: true, toggleValue: $settings.searchable)
                            SettingsListItem(icon: Image(systemName: "lightbulb.fill"), title: "Show in suggestions", hasToggle: true, toggleValue: $settings.showInSuggestions)
                            SettingsListItem(icon: Image(systemName: "square.and.arrow.up"), title: "Allow contact sharing", hasToggle: true, toggleValue: $settings.allowContactSharing)
                        }

                        SettingsSectionView(title: "Cross-Persona Privacy") {
                            SettingsListItem(icon: Image(systemName: "link"), title: "Allow persona linking discovery", hasToggle: true, toggleValue: $settings.allowLinkingDiscovery)

                            if !settings.allowLinkingDiscovery {
                                HStack(spacing: Spacing.sm.rawValue) {
                                    Image(systemName: "info.circle.fill")
                                        .foregroundColor(.echoPrimary)
                                    Text("When OFF, contacts who know you as \"\(personaName)\" cannot see or discover your other personas.")
                                        .typographyStyle(.caption, color: .echoGray400)
                                }
                                .padding(Spacing.md.rawValue)
                            }

                            SettingsListItem(icon: Image(systemName: "shield.fill"), title: "Show shared trust score", hasToggle: true, toggleValue: $settings.showSharedTrustScore)
                            SettingsListItem(icon: Image(systemName: "arrowshape.turn.up.forward.fill"), title: "Cross-persona forwarding", hasToggle: true, toggleValue: $settings.allowCrossPersonaForward)
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Per-Persona Notification Settings View

 struct PersonaNotificationSettingsView: View {
    @Environment(\.dismiss) var dismiss
    let personaName: String
    @Binding var settings: PersonaNotificationSettings
    let onSave: (PersonaNotificationSettings) -> Void

     init(
        personaName: String = "",
        settings: Binding<PersonaNotificationSettings>,
        onSave: @escaping (PersonaNotificationSettings) -> Void = { _ in }
    ) {
        self.personaName = personaName
        self._settings = settings
        self.onSave = onSave
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Notifications: \(personaName)",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        SettingsSectionView(title: "General") {
                            SettingsListItem(icon: Image(systemName: "bell.fill"), title: "Notifications enabled", hasToggle: true, toggleValue: $settings.enabled)
                        }

                        if settings.enabled {
                            SettingsSectionView(title: "Notification Types") {
                                SettingsListItem(icon: Image(systemName: "message.fill"), title: "Messages", value: settings.messagesMode.capitalized)
                                SettingsListItem(icon: Image(systemName: "phone.fill"), title: "Calls", value: settings.callsMode.capitalized)
                                SettingsListItem(icon: Image(systemName: "person.3.fill"), title: "Group activity", value: settings.groupActivityMode.capitalized)
                                SettingsListItem(icon: Image(systemName: "person.badge.plus"), title: "Contact requests", hasToggle: true, toggleValue: $settings.contactRequests)
                            }

                            SettingsSectionView(title: "Sound") {
                                SettingsListItem(icon: Image(systemName: "speaker.wave.2.fill"), title: "Sound enabled", hasToggle: true, toggleValue: $settings.soundEnabled)
                                SettingsListItem(icon: Image(systemName: "iphone.radiowaves.left.and.right"), title: "Vibration", hasToggle: true, toggleValue: $settings.vibrationEnabled)
                            }

                            SettingsSectionView(title: "Preview") {
                                SettingsListItem(icon: Image(systemName: "text.bubble.fill"), title: "Show content", hasToggle: true, toggleValue: $settings.showContent)
                                SettingsListItem(icon: Image(systemName: "person.fill"), title: "Show sender", hasToggle: true, toggleValue: $settings.showSender)
                                SettingsListItem(icon: Image(systemName: "theatermasks.fill"), title: "Show persona name", hasToggle: true, toggleValue: $settings.showPersonaName)
                            }

                            SettingsSectionView(title: "Quiet Hours") {
                                SettingsListItem(icon: Image(systemName: "moon.fill"), title: "Enable quiet hours", hasToggle: true, toggleValue: $settings.quietHoursEnabled)

                                if settings.quietHoursEnabled {
                                    SettingsListItem(icon: Image(systemName: "clock.fill"), title: "From", value: settings.quietHoursStart)
                                    SettingsListItem(icon: Image(systemName: "clock.fill"), title: "To", value: settings.quietHoursEnd)
                                    SettingsListItem(icon: Image(systemName: "exclamationmark.circle.fill"), title: "Allow exceptions", hasToggle: true, toggleValue: $settings.quietHoursAllowExceptions)
                                }
                            }
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Per-Persona Feature Settings View

 struct PersonaFeatureSettingsView: View {
    @Environment(\.dismiss) var dismiss
    let personaName: String
    @Binding var settings: PersonaFeatureSettings
    let onSave: (PersonaFeatureSettings) -> Void

     init(
        personaName: String = "",
        settings: Binding<PersonaFeatureSettings>,
        onSave: @escaping (PersonaFeatureSettings) -> Void = { _ in }
    ) {
        self.personaName = personaName
        self._settings = settings
        self.onSave = onSave
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Features: \(personaName)",
                    showBackButton: true,
                    onBackPressed: {
                        onSave(settings)
                        dismiss()
                    }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        SettingsSectionView(title: "Communication") {
                            SettingsListItem(icon: Image(systemName: "phone.fill"), title: "Voice calls", hasToggle: true, toggleValue: $settings.voiceCalls)
                            SettingsListItem(icon: Image(systemName: "video.fill"), title: "Video calls", hasToggle: true, toggleValue: $settings.videoCalls)
                            SettingsListItem(icon: Image(systemName: "rectangle.on.rectangle"), title: "Screen sharing", hasToggle: true, toggleValue: $settings.screenSharing)
                            SettingsListItem(icon: Image(systemName: "waveform"), title: "Voice messages", hasToggle: true, toggleValue: $settings.voiceMessages)
                        }

                        SettingsSectionView(title: "Sharing") {
                            SettingsListItem(icon: Image(systemName: "paperclip"), title: "File sharing", hasToggle: true, toggleValue: $settings.fileSharing)
                            SettingsListItem(icon: Image(systemName: "location.fill"), title: "Location sharing", hasToggle: true, toggleValue: $settings.locationSharing)
                        }

                        SettingsSectionView(title: "Messages") {
                            SettingsListItem(icon: Image(systemName: "timer"), title: "Disappearing messages", hasToggle: true, toggleValue: $settings.disappearingMessages)
                            SettingsListItem(icon: Image(systemName: "calendar.badge.clock"), title: "Scheduled messages", hasToggle: true, toggleValue: $settings.scheduledMessages)
                            SettingsListItem(icon: Image(systemName: "bell.slash.fill"), title: "Silent messages", hasToggle: true, toggleValue: $settings.silentMessages)
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Enhanced Persona Deletion View

 struct PersonaDeletionView: View {
    @Environment(\.dismiss) var dismiss
    let persona: Persona
    @State private var archiveConversations = true
    @State private var notifyContacts = false
    @State private var keepRecoverable = true
    @State private var exportFirst = false
    @State private var confirmationText = ""
    let onDelete: (PersonaDeletionOptions) -> Void
    let onExport: () -> Void

     init(
        persona: Persona,
        onDelete: @escaping (PersonaDeletionOptions) -> Void = { _ in },
        onExport: @escaping () -> Void = {}
    ) {
        self.persona = persona
        self.onDelete = onDelete
        self.onExport = onExport
    }

    private var confirmationRequired: String {
        "DELETE \(persona.displayName.uppercased())"
    }

    private var canDelete: Bool {
        confirmationText == confirmationRequired
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Delete Persona",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Warning Header
                        VStack(spacing: Spacing.md.rawValue) {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .font(.system(size: 36))
                                .foregroundColor(Color(hex: 0xDC2626))

                            Text("Delete \(persona.displayName)?")
                                .typographyStyle(.h3, color: .echoPrimaryText)
                        }

                        // Impact Summary
                        SettingsSectionView(title: "What happens when you delete") {
                            HStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "person.crop.circle.badge.minus")
                                    .foregroundColor(.echoGray500)
                                Text("Profile \"\(persona.displayName)\" will be removed")
                                    .typographyStyle(.body, color: .echoPrimaryText)
                                Spacer()
                            }
                            .padding(Spacing.md.rawValue)

                            HStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "person.2.slash")
                                    .foregroundColor(.echoGray500)
                                Text("\(persona.contactCount) contacts will no longer be able to reach you")
                                    .typographyStyle(.body, color: .echoPrimaryText)
                                Spacer()
                            }
                            .padding(Spacing.md.rawValue)

                            HStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "bubble.left.and.bubble.right.fill")
                                    .foregroundColor(.echoGray500)
                                Text("\(persona.messageCount) messages will be archived or deleted")
                                    .typographyStyle(.body, color: .echoPrimaryText)
                                Spacer()
                            }
                            .padding(Spacing.md.rawValue)

                            HStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "rosette")
                                    .foregroundColor(.echoGray500)
                                Text("\(persona.badges.count) badges will be revoked")
                                    .typographyStyle(.body, color: .echoPrimaryText)
                                Spacer()
                            }
                            .padding(Spacing.md.rawValue)
                        }

                        // Options
                        SettingsSectionView(title: "Message History") {
                            RadioRow(title: "Archive conversations (keep locally)", isSelected: archiveConversations) {
                                archiveConversations = true
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                            RadioRow(title: "Delete all conversations", isSelected: !archiveConversations) {
                                archiveConversations = false
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                        }

                        SettingsSectionView(title: "Notify Contacts") {
                            RadioRow(title: "Send \"account deleted\" notification", isSelected: notifyContacts) {
                                notifyContacts = true
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                            RadioRow(title: "Disappear silently", isSelected: !notifyContacts) {
                                notifyContacts = false
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                        }

                        SettingsSectionView(title: "Recovery") {
                            RadioRow(title: "Keep recoverable for 30 days", isSelected: keepRecoverable) {
                                keepRecoverable = true
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                            RadioRow(title: "Delete immediately and permanently", isSelected: !keepRecoverable) {
                                keepRecoverable = false
                            }
                            .padding(.horizontal, Spacing.md.rawValue)
                        }

                        // Export button
                        Button(action: onExport) {
                            HStack {
                                Image(systemName: "square.and.arrow.up")
                                    .foregroundColor(.echoPrimary)
                                Text("Export persona data before deletion")
                                    .typographyStyle(.body, color: .echoPrimary)
                            }
                            .frame(maxWidth: .infinity)
                            .padding(Spacing.lg.rawValue)
                            .background(Color.echoSurface)
                            .cornerRadius(12)
                        }

                        // Confirmation
                        VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                            Text("Type \"\(confirmationRequired)\" to confirm:")
                                .typographyStyle(.body, color: .echoPrimaryText)

                            TextField("", text: $confirmationText)
                                .typographyStyle(.body, color: .echoPrimaryText)
                                .padding(Spacing.md.rawValue)
                                .background(Color.echoSurface)
                                .cornerRadius(8)
                                .overlay(
                                    RoundedRectangle(cornerRadius: 8)
                                        .stroke(Color.echoGray200, lineWidth: 1)
                                )
                        }

                        HStack(spacing: Spacing.md.rawValue) {
                            EchoButton("Cancel", style: .secondary, size: .large) {
                                dismiss()
                            }

                            Button {
                                let options = PersonaDeletionOptions(
                                    archiveConversations: archiveConversations,
                                    notifyContacts: notifyContacts,
                                    recoveryPeriodDays: keepRecoverable ? 30 : 0,
                                    exportBeforeDelete: exportFirst
                                )
                                onDelete(options)
                                dismiss()
                            } label: {
                                Text("Delete Persona")
                                    .typographyStyle(.body, color: .white)
                                    .frame(maxWidth: .infinity)
                                    .padding(Spacing.md.rawValue)
                                    .background(canDelete ? Color(hex: 0xDC2626) : Color.echoGray300)
                                    .cornerRadius(12)
                            }
                            .disabled(!canDelete)
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Persona Badges View

 struct PersonaBadgesView: View {
    @Environment(\.dismiss) var dismiss
    let persona: Persona
    let onAddBadge: () -> Void
    let onRemoveBadge: (String) -> Void

     init(
        persona: Persona,
        onAddBadge: @escaping () -> Void = {},
        onRemoveBadge: @escaping (String) -> Void = { _ in }
    ) {
        self.persona = persona
        self.onAddBadge = onAddBadge
        self.onRemoveBadge = onRemoveBadge
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Badges: \(persona.displayName)",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        // Trust score inherited
                        SettingsSectionView(title: "Inherited Verification") {
                            HStack(spacing: Spacing.md.rawValue) {
                                Image(systemName: "checkmark.seal.fill")
                                    .foregroundColor(.echoPrimary)
                                VStack(alignment: .leading, spacing: 2) {
                                    Text("Master Trust Score")
                                        .typographyStyle(.body, color: .echoPrimaryText)
                                    Text("Inherited from your verified identity")
                                        .typographyStyle(.caption, color: .echoGray400)
                                }
                                Spacer()
                            }
                            .padding(Spacing.lg.rawValue)
                        }

                        // Persona-specific badges
                        SettingsSectionView(title: "Persona Badges (\(persona.badges.count))") {
                            if persona.badges.isEmpty {
                                VStack(spacing: Spacing.md.rawValue) {
                                    Image(systemName: "rosette")
                                        .font(.system(size: 36))
                                        .foregroundColor(.echoGray300)
                                    Text("No badges yet")
                                        .typographyStyle(.body, color: .echoSecondaryText)
                                    Text("Earn badges through verification and achievements")
                                        .typographyStyle(.caption, color: .echoGray400)
                                        .multilineTextAlignment(.center)
                                }
                                .padding(Spacing.xl.rawValue)
                            } else {
                                ForEach(persona.badges) { badge in
                                    HStack(spacing: Spacing.md.rawValue) {
                                        Image(systemName: badgeIcon(badge.type))
                                            .foregroundColor(.echoPrimary)
                                            .frame(width: 24)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(badgeLabel(badge.type))
                                                .typographyStyle(.body, color: .echoPrimaryText)
                                            Text("Issued: \(badge.issuedAt.formatted(date: .abbreviated, time: .omitted))")
                                                .typographyStyle(.caption, color: .echoGray400)
                                        }
                                        Spacer()
                                        if badge.verifiable {
                                            Image(systemName: "checkmark.circle.fill")
                                                .foregroundColor(.echoSuccess)
                                        }
                                    }
                                    .padding(Spacing.lg.rawValue)
                                }
                            }
                        }

                        // Available verifications
                        SettingsSectionView(title: "Available Verifications") {
                            badgeVerificationRow(icon: "building.2.fill", title: "Verify Employer", subtitle: "Link your work email")
                            Divider().padding(.leading, 52)
                            badgeVerificationRow(icon: "link", title: "Link LinkedIn", subtitle: "Connect professional profile")
                            Divider().padding(.leading, 52)
                            badgeVerificationRow(icon: "gamecontroller.fill", title: "Link Gaming Account", subtitle: "Connect Steam, Xbox, PS")
                            Divider().padding(.leading, 52)
                            badgeVerificationRow(icon: "camera.fill", title: "Photo Verification", subtitle: "Verify your photos match")
                        }
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
    }

    private func badgeVerificationRow(icon: String, title: String, subtitle: String) -> some View {
        Button(action: onAddBadge) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: icon)
                    .foregroundColor(.echoGray500)
                    .frame(width: 24)
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .typographyStyle(.body, color: .echoPrimaryText)
                    Text(subtitle)
                        .typographyStyle(.caption, color: .echoGray400)
                }
                Spacer()
                Image(systemName: "chevron.right")
                    .foregroundColor(.echoGray400)
                    .font(.system(size: 14))
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .frame(height: 56)
        }
    }

    private func badgeIcon(_ type: PersonaBadgeType) -> String {
        switch type {
        case .verifiedEmployer, .domainEmail: return "building.2.fill"
        case .professionalCredential: return "doc.text.fill"
        case .linkedinVerified: return "link"
        case .gameAchievement, .tournamentWinner: return "trophy.fill"
        case .verifiedGamerTag: return "gamecontroller.fill"
        case .portfolioVerified, .publishedWork: return "paintpalette.fill"
        case .communityModerator: return "shield.fill"
        case .trustedContributor: return "hand.thumbsup.fill"
        case .eventOrganizer: return "calendar.badge.plus"
        case .photoVerified: return "camera.fill"
        case .ageVerified: return "person.text.rectangle"
        case .locationVerified: return "location.fill"
        }
    }

    private func badgeLabel(_ type: PersonaBadgeType) -> String {
        switch type {
        case .verifiedEmployer: return "Verified Employer"
        case .professionalCredential: return "Professional Credential"
        case .linkedinVerified: return "LinkedIn Verified"
        case .domainEmail: return "Domain Email Verified"
        case .gameAchievement: return "Game Achievement"
        case .tournamentWinner: return "Tournament Winner"
        case .verifiedGamerTag: return "Verified Gamer Tag"
        case .portfolioVerified: return "Portfolio Verified"
        case .publishedWork: return "Published Work"
        case .communityModerator: return "Community Moderator"
        case .trustedContributor: return "Trusted Contributor"
        case .eventOrganizer: return "Event Organizer"
        case .photoVerified: return "Photo Verified"
        case .ageVerified: return "Age Verified"
        case .locationVerified: return "Location Verified"
        }
    }
}

// MARK: - Persona Recovery Banner

 struct PersonaRecoveryBanner: View {
    let persona: Persona
    let onRecover: () -> Void

     var body: some View {
        if let deletion = persona.deletionState, deletion.canRecover {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundColor(.echoWarning)

                VStack(alignment: .leading, spacing: 2) {
                    Text("Scheduled for deletion")
                        .typographyStyle(.body, color: .echoPrimaryText)
                    if let expires = deletion.recoveryExpiresAt {
                        let days = Calendar.current.dateComponents([.day], from: Date(), to: expires).day ?? 0
                        Text("\(days) days remaining to recover")
                            .typographyStyle(.caption, color: .echoGray400)
                    }
                }

                Spacer()

                Button("Recover") {
                    onRecover()
                }
                .typographyStyle(.body, color: .echoPrimary)
            }
            .padding(Spacing.lg.rawValue)
            .background(Color.echoWarning.opacity(0.1))
            .cornerRadius(12)
        }
    }
}

// MARK: - Refactored Settings Hub (navigates to sub-screens)

 struct SettingsView: View {
    @Environment(\.dismiss) var dismiss
    @State private var showSignOutAlert = false

    let onAccountSettings: () -> Void
    let onPrivacySettings: () -> Void
    let onNotificationSettings: () -> Void
    let onAppearanceSettings: () -> Void
    let onStorageSettings: () -> Void
    let onHelpSupport: () -> Void
    let onAbout: () -> Void
    let onSignOut: () -> Void

     init(
        onAccountSettings: @escaping () -> Void = {},
        onPrivacySettings: @escaping () -> Void = {},
        onNotificationSettings: @escaping () -> Void = {},
        onAppearanceSettings: @escaping () -> Void = {},
        onStorageSettings: @escaping () -> Void = {},
        onHelpSupport: @escaping () -> Void = {},
        onAbout: @escaping () -> Void = {},
        onSignOut: @escaping () -> Void = {}
    ) {
        self.onAccountSettings = onAccountSettings
        self.onPrivacySettings = onPrivacySettings
        self.onNotificationSettings = onNotificationSettings
        self.onAppearanceSettings = onAppearanceSettings
        self.onStorageSettings = onStorageSettings
        self.onHelpSupport = onHelpSupport
        self.onAbout = onAbout
        self.onSignOut = onSignOut
    }

     var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(
                    title: "Settings",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        VStack(spacing: 0) {
                            SettingsNavRow(icon: "person.fill", title: "Account", action: onAccountSettings)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "lock.fill", title: "Privacy & Security", action: onPrivacySettings)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "bell.fill", title: "Notifications", action: onNotificationSettings)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "paintbrush.fill", title: "Appearance", action: onAppearanceSettings)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "internaldrive.fill", title: "Storage & Data", action: onStorageSettings)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "questionmark.circle.fill", title: "Help & Support", action: onHelpSupport)
                            Divider().padding(.leading, 52)
                            SettingsNavRow(icon: "info.circle.fill", title: "About Echo", action: onAbout)
                        }
                        .background(Color.echoSurface)
                        .cornerRadius(12)

                        // Sign Out
                        VStack(spacing: Spacing.md.rawValue) {
                            EchoButton(
                                "Sign Out",
                                style: .destructive,
                                size: .large,
                                action: { showSignOutAlert = true }
                            )
                        }
                        .padding(.top, Spacing.lg.rawValue)
                    }
                    .echoSpacing(.lg)
                }
            }
        }
        .navigationBarBackButtonHidden(true)
        .alert("Sign Out", isPresented: $showSignOutAlert) {
            Button("Cancel", role: .cancel) {}
            Button("Sign Out", role: .destructive) { onSignOut() }
        } message: {
            Text("Are you sure you want to sign out?")
        }
    }
}

// MARK: - Reusable Components

struct SettingsSectionView<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    init(title: String, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.content = content
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            Text(title)
                .typographyStyle(.caption, color: .echoGray500)
                .textCase(.uppercase)

            VStack(spacing: 0) {
                content()
            }
            .background(Color.echoSurface)
            .cornerRadius(12)
        }
    }
}

struct StatCard: View {
    let label: String
    let value: String

    var body: some View {
        VStack(spacing: Spacing.xs.rawValue) {
            Text(value)
                .typographyStyle(.h4, color: .echoPrimary)

            Text(label)
                .typographyStyle(.caption, color: .echoGray500)
        }
        .frame(maxWidth: .infinity)
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
    }
}

struct PersonaCard: View {
    let persona: Persona
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: Spacing.xs.rawValue) {
                Text(persona.type.emoji)
                    .font(.system(size: 32))

                Text(persona.type.label)
                    .typographyStyle(.caption, color: .echoPrimaryText)
                    .lineLimit(1)
            }
            .frame(width: 72, height: 88)
            .background(Color.echoSurface)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(
                        persona.isDefault ? Color.echoPrimary : Color.echoGray200,
                        lineWidth: persona.isDefault ? 2 : 1
                    )
            )
        }
    }
}

struct AddPersonaCard: View {
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: Spacing.xs.rawValue) {
                Image(systemName: "plus")
                    .font(.system(size: 24, weight: .medium))
                    .foregroundColor(.echoGray400)

                Text("Add")
                    .typographyStyle(.caption, color: .echoGray400)
            }
            .frame(width: 72, height: 88)
            .background(Color.echoSurface)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(style: StrokeStyle(lineWidth: 1, dash: [6]))
                    .foregroundColor(.echoGray300)
            )
        }
    }
}

struct PersonaDetailCard: View {
    let persona: Persona
    let onEdit: () -> Void
    let onDelete: () -> Void
    let onSetDefault: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
            HStack {
                Text(persona.type.emoji)
                    .font(.system(size: 24))

                Text(persona.name)
                    .typographyStyle(.h4, color: .echoPrimaryText)

                if persona.isDefault {
                    Image(systemName: "star.fill")
                        .foregroundColor(.echoWarning)
                        .font(.system(size: 14))
                }

                Spacer()
            }

            Text(persona.displayName)
                .typographyStyle(.body, color: .echoSecondaryText)

            if let bio = persona.bio, !bio.isEmpty {
                Text(bio)
                    .typographyStyle(.caption, color: .echoGray400)
            }

            HStack {
                if persona.isDefault {
                    Text("Default persona")
                        .typographyStyle(.caption, color: .echoGray400)
                } else {
                    Text("\(persona.contactCount) contacts can see")
                        .typographyStyle(.caption, color: .echoGray400)
                }

                Spacer()

                Button("Edit", action: onEdit)
                    .typographyStyle(.caption, color: .echoPrimary)
            }
        }
        .padding(Spacing.lg.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(Color.echoGray200, lineWidth: 1)
        )
        .contextMenu {
            Button(action: onEdit) {
                Label("Edit", systemImage: "pencil")
            }
            if !persona.isDefault {
                Button(action: onSetDefault) {
                    Label("Set as Default", systemImage: "star")
                }
            }
            Button(role: .destructive, action: onDelete) {
                Label("Delete", systemImage: "trash")
            }
        }
    }
}

struct QuickActionRow: View {
    let icon: String
    let title: String
    var subtitle: String?
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: icon)
                    .foregroundColor(.echoPrimary)
                    .frame(width: 24)

                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .typographyStyle(.body, color: .echoPrimaryText)

                    if let subtitle = subtitle {
                        Text(subtitle)
                            .typographyStyle(.caption, color: .echoGray400)
                    }
                }

                Spacer()

                Image(systemName: "chevron.right")
                    .foregroundColor(.echoGray400)
                    .font(.system(size: 14))
            }
            .padding(Spacing.lg.rawValue)
            .background(Color.echoSurface)
            .cornerRadius(12)
        }
    }
}

struct SettingsNavRow: View {
    let icon: String
    let title: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: icon)
                    .foregroundColor(.echoGray500)
                    .frame(width: 24)

                Text(title)
                    .typographyStyle(.body, color: .echoPrimaryText)

                Spacer()

                Image(systemName: "chevron.right")
                    .foregroundColor(.echoGray400)
                    .font(.system(size: 14))
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .frame(height: 56)
        }
    }
}

struct ProfileTextField: View {
    let label: String
    @Binding var text: String
    var prefix: String?
    var isMultiline: Bool = false

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
            Text(label)
                .typographyStyle(.caption, color: .echoGray500)

            if isMultiline {
                TextEditor(text: $text)
                    .typographyStyle(.body, color: .echoPrimaryText)
                    .frame(minHeight: 80)
                    .padding(Spacing.sm.rawValue)
                    .background(Color.echoSurface)
                    .cornerRadius(8)
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.echoGray200, lineWidth: 1)
                    )
            } else {
                HStack {
                    if let prefix = prefix {
                        Text(prefix)
                            .typographyStyle(.body, color: .echoGray400)
                    }
                    TextField(label, text: $text)
                        .typographyStyle(.body, color: .echoPrimaryText)
                }
                .padding(Spacing.md.rawValue)
                .background(Color.echoSurface)
                .cornerRadius(8)
                .overlay(
                    RoundedRectangle(cornerRadius: 8)
                        .stroke(Color.echoGray200, lineWidth: 1)
                )
            }
        }
    }
}

struct RadioRow: View {
    let title: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: isSelected ? "largecircle.fill.circle" : "circle")
                    .foregroundColor(isSelected ? .echoPrimary : .echoGray400)

                Text(title)
                    .typographyStyle(.body, color: .echoPrimaryText)

                Spacer()
            }
        }
    }
}

struct PersonaTypeSelector: View {
    let type: PersonaType
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: Spacing.xs.rawValue) {
                Text(type.emoji)
                    .font(.system(size: 24))

                Text(type.label)
                    .typographyStyle(.caption, color: isSelected ? .echoPrimary : .echoPrimaryText)
                    .lineLimit(1)
            }
            .frame(maxWidth: .infinity)
            .padding(Spacing.sm.rawValue)
            .background(isSelected ? Color.echoPrimary.opacity(0.1) : Color.echoSurface)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(isSelected ? Color.echoPrimary : Color.echoGray200, lineWidth: isSelected ? 2 : 1)
            )
        }
    }
}

struct ThemeOption: View {
    let icon: String
    let label: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: Spacing.sm.rawValue) {
                Image(systemName: icon)
                    .font(.system(size: 24))
                    .foregroundColor(isSelected ? .echoPrimary : .echoGray400)

                Text(label)
                    .typographyStyle(.caption, color: isSelected ? .echoPrimary : .echoPrimaryText)

                if isSelected {
                    Image(systemName: "checkmark")
                        .foregroundColor(.echoPrimary)
                        .font(.system(size: 12, weight: .bold))
                }
            }
            .frame(maxWidth: .infinity)
            .padding(Spacing.md.rawValue)
            .background(isSelected ? Color.echoPrimary.opacity(0.1) : Color.echoSurface)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(isSelected ? Color.echoPrimary : Color.echoGray200, lineWidth: isSelected ? 2 : 1)
            )
        }
    }
}

struct AppIconOption: View {
    let color: Color
    let label: String
    var isPremium: Bool = false
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: Spacing.xs.rawValue) {
                ZStack {
                    RoundedRectangle(cornerRadius: 12)
                        .fill(color)
                        .frame(width: 48, height: 48)

                    if isSelected {
                        Image(systemName: "checkmark")
                            .foregroundColor(.white)
                            .font(.system(size: 16, weight: .bold))
                    }

                    if isPremium && !isSelected {
                        Image(systemName: "star.fill")
                            .foregroundColor(.white)
                            .font(.system(size: 14))
                    }
                }

                Text(label)
                    .typographyStyle(.caption, color: .echoPrimaryText)

                if isPremium {
                    Text("Premium")
                        .typographyStyle(.tiny, color: .echoWarning)
                }
            }
        }
    }
}

struct StorageBar: View {
    let used: Int64
    let total: Int64

    private var fraction: Double {
        guard total > 0 else { return 0 }
        return min(1, Double(used) / Double(total))
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 6)
                        .fill(Color.echoGray200)
                        .frame(height: 12)

                    RoundedRectangle(cornerRadius: 6)
                        .fill(Color.echoPrimary)
                        .frame(width: geo.size.width * fraction, height: 12)
                }
            }
            .frame(height: 12)

            let formatter = ByteCountFormatter()
            Text("\(formatter.string(fromByteCount: used)) of \(formatter.string(fromByteCount: total)) used")
                .typographyStyle(.caption, color: .echoGray500)
        }
        .padding(Spacing.lg.rawValue)
    }
}

// MARK: - Preview

#if DEBUG
struct ProfileScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            ProfileTabView(
                profile: ProfileData(
                    displayName: "Alex Echo",
                    username: "alexecho",
                    bio: "Product designer & crypto enthusiast",
                    trustScore: 72,
                    trustLevel: "Trusted",
                    isVerified: true,
                    messagesSent: 247,
                    contactsCount: 42,
                    echoRewards: 142.5
                )
            )
        }
    }
}
#endif
