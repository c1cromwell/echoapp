#if os(iOS)
import SwiftUI

/// Live persona management: rename the primary persona (your display name),
/// and add / edit / delete additional personas. The primary persona cannot be
/// deleted; the hidden-vault persona is managed by the hidden-chats flow and is
/// not shown here. Backed by `PersonaSessionStore` via `AppState`.
struct PersonaSettingsView: View {
    @Environment(AppState.self) private var appState
    @Environment(\.dismiss) private var dismiss

    @State private var editorTarget: EditorTarget?
    @State private var pendingDelete: PersonaSummary?

    /// Preset accent colors offered in the editor.
    static let palette: [UInt32] = [
        0x0E7AB8, 0x1F7A4C, 0x7C3AED, 0xF43F5E, 0xF59E0B, 0x0D9488, 0x475569,
    ]

    private enum EditorTarget: Identifiable {
        case create
        case edit(PersonaSummary)
        var id: String {
            switch self {
            case .create: return "create"
            case .edit(let p): return p.id
            }
        }
    }

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(title: "Personas", showBackButton: true, onBackPressed: { dismiss() })

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        SettingsSectionView(title: "Your Personas") {
                            VStack(spacing: 0) {
                                ForEach(Array(appState.manageablePersonas.enumerated()), id: \.element.id) { index, persona in
                                    personaRow(persona)
                                    if index < appState.manageablePersonas.count - 1 {
                                        Divider().padding(.leading, 60)
                                    }
                                }
                            }
                        }

                        Button {
                            editorTarget = .create
                        } label: {
                            HStack(spacing: Spacing.sm.rawValue) {
                                Image(systemName: "plus.circle.fill")
                                Text("Add Persona")
                                    .typographyStyle(.body, color: .white)
                            }
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Color.echoSignal)
                            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                        }
                        .accessibilityIdentifier("personas.add")

                        Text("Your primary persona is your main profile — it can be renamed but not removed. Add personas to keep different sides of your life separate.")
                            .typographyStyle(.caption, color: .echoSecondaryText)
                            .fixedSize(horizontal: false, vertical: true)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }
                    .padding(Spacing.lg.rawValue)
                }
            }
        }
        .toolbar(.hidden, for: .navigationBar)
        .sheet(item: $editorTarget) { target in
            switch target {
            case .create:
                PersonaEditorSheet(mode: .create) { name, color in
                    appState.addPersona(name: name, colorHex: color)
                }
            case .edit(let persona):
                PersonaEditorSheet(mode: .edit(persona)) { name, color in
                    appState.updatePersona(
                        id: persona.id,
                        name: name,
                        colorHex: color,
                        trustLevel: persona.trustLevel
                    )
                }
            }
        }
        .confirmationDialog(
            "Delete this persona?",
            isPresented: Binding(get: { pendingDelete != nil }, set: { if !$0 { pendingDelete = nil } }),
            presenting: pendingDelete
        ) { persona in
            Button("Delete \(persona.name)", role: .destructive) {
                appState.deletePersona(id: persona.id)
                pendingDelete = nil
            }
            Button("Cancel", role: .cancel) { pendingDelete = nil }
        } message: { persona in
            Text("Conversations stay, but \"\(persona.name)\" will no longer be selectable.")
        }
    }

    private func personaRow(_ persona: PersonaSummary) -> some View {
        let isPrimary = persona.id == PersonaSessionStore.primaryPersonaId
        return HStack(spacing: Spacing.md.rawValue) {
            avatar(persona)

            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(persona.name)
                        .typographyStyle(.body, color: .echoPrimaryText)
                    if isPrimary {
                        Text("PRIMARY")
                            .font(.system(size: 9, weight: .bold))
                            .foregroundStyle(Color.echoSignal)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.echoSignal.opacity(0.12))
                            .clipShape(Capsule())
                    }
                }
                Text(persona.trustLevel)
                    .typographyStyle(.caption, color: .echoSecondaryText)
            }

            Spacer()

            Button("Edit") { editorTarget = .edit(persona) }
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(Color.echoSignal)

            if !isPrimary {
                Button {
                    pendingDelete = persona
                } label: {
                    Image(systemName: "trash")
                        .foregroundStyle(Color.echoError)
                }
                .padding(.leading, 4)
            }
        }
        .padding(Spacing.md.rawValue)
    }

    private func avatar(_ persona: PersonaSummary) -> some View {
        Circle()
            .fill(Color(hex: persona.colorHex))
            .frame(width: 40, height: 40)
            .overlay(
                Text(persona.initials)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.white)
            )
    }
}

// MARK: - Editor Sheet

private struct PersonaEditorSheet: View {
    enum Mode {
        case create
        case edit(PersonaSummary)
    }

    let mode: Mode
    let onSave: (String, UInt32) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var name: String
    @State private var colorHex: UInt32
    @FocusState private var nameFocused: Bool

    private let isPrimary: Bool

    init(mode: Mode, onSave: @escaping (String, UInt32) -> Void) {
        self.mode = mode
        self.onSave = onSave
        switch mode {
        case .create:
            _name = State(initialValue: "")
            _colorHex = State(initialValue: PersonaSettingsView.palette.first ?? 0x0E7AB8)
            isPrimary = false
        case .edit(let p):
            _name = State(initialValue: p.name)
            _colorHex = State(initialValue: p.colorHex)
            isPrimary = p.id == PersonaSessionStore.primaryPersonaId
        }
    }

    private var title: String {
        switch mode {
        case .create: return "New Persona"
        case .edit: return isPrimary ? "Edit Display Name" : "Edit Persona"
        }
    }

    private var canSave: Bool {
        !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Name") {
                    TextField("Persona name", text: $name)
                        .focused($nameFocused)
                        .autocorrectionDisabled(true)
                        .accessibilityIdentifier("persona.nameField")
                }

                // The primary persona's color is derived from the profile; only
                // custom personas expose a color picker.
                if !isPrimary {
                    Section("Color") {
                        LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 7), spacing: 12) {
                            ForEach(PersonaSettingsView.palette, id: \.self) { hex in
                                Circle()
                                    .fill(Color(hex: hex))
                                    .frame(width: 32, height: 32)
                                    .overlay(
                                        Circle()
                                            .strokeBorder(Color.echoPrimaryText, lineWidth: colorHex == hex ? 2 : 0)
                                    )
                                    .onTapGesture { colorHex = hex }
                            }
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
            .navigationTitle(title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        onSave(name, colorHex)
                        dismiss()
                    }
                    .disabled(!canSave)
                }
            }
            .onAppear { nameFocused = true }
        }
    }
}
#endif
