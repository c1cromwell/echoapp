import SwiftUI

/// A persona shown in the hub header / switcher. Hidden personas reveal nothing from
/// the outside until a biometric gate passes (ux-spec §2.3, principle 6).
public struct PersonaSummary: Identifiable, Equatable, Hashable, Sendable, Codable {
    public let id: String
    public let name: String
    public let initials: String
    public let trustLevel: String
    public let isHidden: Bool
    public let colorHex: UInt32

    public init(
        id: String,
        name: String,
        initials: String,
        trustLevel: String = "Verified",
        isHidden: Bool = false,
        colorHex: UInt32 = 0x0E7AB8
    ) {
        self.id = id
        self.name = name
        self.initials = initials
        self.trustLevel = trustLevel
        self.isHidden = isHidden
        self.colorHex = colorHex
    }
}

/// Compact hub-header control (see docs/design-previews/messagehub1.png): centered
/// avatar + name + verified seal + chevron, tap to switch. No subtitle, no underline.
public struct PersonaSwitcherHeader: View {
    let active: PersonaSummary
    let onTap: () -> Void

    public init(active: PersonaSummary, onTap: @escaping () -> Void) {
        self.active = active
        self.onTap = onTap
    }

    /// Show the verified seal for identity-confirmed personas (T2+).
    private var isVerified: Bool {
        switch active.trustLevel.lowercased() {
        case "verified", "trusted", "highlytrusted", "premium", "elite": return true
        default: return false
        }
    }

    public var body: some View {
        Button(action: onTap) {
            HStack(spacing: 6) {
                avatar(active, size: 24)
                Text(active.name)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.echoInk)
                if isVerified {
                    Image(systemName: "checkmark")
                        .font(.system(size: 9, weight: .bold))
                        .foregroundColor(.white)
                        .frame(width: 16, height: 16)
                        .background(Color.echoTrustGreen)
                        .clipShape(Circle())
                }
                Image(systemName: "chevron.down")
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(.echoInk40)
                Spacer(minLength: 0)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.sm.rawValue)
        }
        .buttonStyle(.plain)
    }
}

/// Persona-switch sheet. Hidden personas show a lock and route through `onSelectHidden`
/// (caller presents the biometric gate); visible ones switch immediately.
public struct PersonaSwitcherSheet: View {
    let personas: [PersonaSummary]
    let activeID: String
    let onSelect: (PersonaSummary) -> Void
    let onSelectHidden: (PersonaSummary) -> Void
    let onCreate: () -> Void

    public init(
        personas: [PersonaSummary],
        activeID: String,
        onSelect: @escaping (PersonaSummary) -> Void,
        onSelectHidden: @escaping (PersonaSummary) -> Void,
        onCreate: @escaping () -> Void = {}
    ) {
        self.personas = personas
        self.activeID = activeID
        self.onSelect = onSelect
        self.onSelectHidden = onSelectHidden
        self.onCreate = onCreate
    }

    @State private var globalSilent: Bool = UserDefaults.standard.bool(forKey: "echo.globalSilent")
    @State private var defaultTimer: DisappearingTimer = {
        let raw = UserDefaults.standard.string(forKey: "echo.defaultDisappearing") ?? "off"
        return DisappearingTimer(rawValue: raw) ?? .off
    }()

    public var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                Capsule()
                    .fill(Color.echoHair)
                    .frame(width: 38, height: 5)
                    .padding(.top, 8)
                    .padding(.bottom, 12)

                Text("Settings & Personas")
                    .font(.system(size: 17, weight: .semibold))
                    .foregroundColor(.echoInk)
                    .padding(.bottom, 12)

                chatSettingsSection

                Divider().padding(.vertical, 8)

                Text("PERSONAS")
                    .font(.system(size: 11, weight: .bold))
                    .tracking(1.5)
                    .foregroundColor(.echoInk40)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.bottom, 4)

                ForEach(personas) { persona in
                    row(persona)
                }

                Button(action: onCreate) {
                    HStack(spacing: Spacing.md.rawValue) {
                        Image(systemName: "plus.circle")
                            .font(.system(size: 22))
                            .foregroundColor(.echoSignal)
                        Text("Create a persona")
                            .font(.system(size: 15, weight: .medium))
                            .foregroundColor(.echoSignal)
                        Spacer()
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, Spacing.md.rawValue)
                }
                .buttonStyle(.plain)

                Spacer(minLength: 20)
            }
        }
        .background(Color.echoPaper)
    }

    private var chatSettingsSection: some View {
        VStack(spacing: 0) {
            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: "bell.slash").frame(width: 24).foregroundColor(.echoInk55)
                Text("Silent notifications")
                    .font(.system(size: 15))
                    .foregroundColor(.echoInk)
                Spacer()
                Toggle("", isOn: $globalSilent)
                    .labelsHidden()
                    .tint(.echoSignal)
                    .onChange(of: globalSilent) { _, val in
                        UserDefaults.standard.set(val, forKey: "echo.globalSilent")
                    }
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
            .overlay(Divider(), alignment: .bottom)

            VStack(alignment: .leading, spacing: 8) {
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: "timer").frame(width: 24).foregroundColor(.echoInk55)
                    Text("Disappearing messages")
                        .font(.system(size: 15))
                        .foregroundColor(.echoInk)
                    Spacer()
                }
                Picker("", selection: $defaultTimer) {
                    ForEach(DisappearingTimer.allCases, id: \.self) { t in
                        Text(t.label).tag(t)
                    }
                }
                .pickerStyle(.segmented)
                .onChange(of: defaultTimer) { _, val in
                    UserDefaults.standard.set(val.rawValue, forKey: "echo.defaultDisappearing")
                }
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
            .overlay(Divider(), alignment: .bottom)

            HStack(spacing: Spacing.md.rawValue) {
                Image(systemName: "checkmark.shield").frame(width: 24).foregroundColor(.echoInk55)
                Text("Verify identity")
                    .font(.system(size: 15))
                    .foregroundColor(.echoInk)
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.echoInk40)
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
            .overlay(Divider(), alignment: .bottom)
        }
    }

    @ViewBuilder
    private func row(_ persona: PersonaSummary) -> some View {
        Button {
            if persona.isHidden { onSelectHidden(persona) } else { onSelect(persona) }
        } label: {
            HStack(spacing: Spacing.md.rawValue) {
                avatar(persona, size: 44)
                VStack(alignment: .leading, spacing: 2) {
                    Text(persona.isHidden ? "Hidden persona" : persona.name)
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundColor(.echoInk)
                    if !persona.isHidden {
                        TrustBadge(level: persona.trustLevel, size: .small)
                    }
                }
                Spacer()
                if persona.id == activeID {
                    Image(systemName: "checkmark.circle.fill").foregroundColor(.echoSignal)
                } else if persona.isHidden {
                    Image(systemName: "lock.fill").foregroundColor(.echoInk40)
                }
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.md.rawValue)
        }
        .buttonStyle(.plain)
    }
}

@ViewBuilder
private func avatar(_ persona: PersonaSummary, size: CGFloat) -> some View {
    Text(persona.isHidden ? "" : persona.initials)
        .font(.system(size: size * 0.4, weight: .semibold))
        .foregroundColor(.white)
        .frame(width: size, height: size)
        .background(persona.isHidden ? Color.echoNight : Color(hex: persona.colorHex))
        .clipShape(Circle())
        .overlay(
            persona.isHidden
                ? Image(systemName: "eye.slash.fill").font(.system(size: size * 0.34)).foregroundColor(.echoNightInk)
                : nil
        )
}

#if DEBUG
struct PersonaSwitcher_Previews: PreviewProvider {
    static let personas = [
        PersonaSummary(id: "1", name: "Aria (public)", initials: "AR", trustLevel: "Verified", colorHex: 0x0E7AB8),
        PersonaSummary(id: "2", name: "Builder", initials: "BD", trustLevel: "Trusted", colorHex: 0x1F7A4C),
        PersonaSummary(id: "3", name: "Hidden", initials: "", isHidden: true),
    ]
    static var previews: some View {
        VStack {
            PersonaSwitcherHeader(active: personas[0]) {}
            PersonaSwitcherSheet(personas: personas, activeID: "1", onSelect: { _ in }, onSelectHidden: { _ in })
        }
        .background(Color.echoPaper)
    }
}
#endif
