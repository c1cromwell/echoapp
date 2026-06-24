#if os(iOS)
import SwiftUI

/// Permission grant sheet shown before installing a bot (WO-63 / Stage 4).
struct BotInstallPermissionsSheet: View {
    let bot: BotManifestDTO
    let onInstall: ([String]) -> Void
    let onCancel: () -> Void

    @State private var granted: Set<String>

    init(bot: BotManifestDTO, onInstall: @escaping ([String]) -> Void, onCancel: @escaping () -> Void) {
        self.bot = bot
        self.onInstall = onInstall
        self.onCancel = onCancel
        _granted = State(initialValue: Set(bot.requiredPermissions))
    }

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text(bot.description)
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk70)
                } header: {
                    Text(bot.name)
                } footer: {
                    Text("Trust score \(bot.trustScore)/100 · runs sandboxed on your device")
                }

                Section("Permissions") {
                    ForEach(bot.requiredPermissions, id: \.self) { perm in
                        Toggle(permissionLabel(perm), isOn: Binding(
                            get: { granted.contains(perm) },
                            set: { on in
                                if on { granted.insert(perm) } else { granted.remove(perm) }
                            }
                        ))
                    }
                }
            }
            .navigationTitle("Install bot")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel", action: onCancel)
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Install") { onInstall(grantedPermissions) }
                        .disabled(!allRequiredGranted)
                }
            }
        }
    }

    private var allRequiredGranted: Bool {
        Set(bot.requiredPermissions).isSubset(of: granted)
    }

    var grantedPermissions: [String] {
        bot.requiredPermissions.filter { granted.contains($0) }
    }

    private func permissionLabel(_ perm: String) -> String {
        switch perm {
        case "send_message": return "Send messages"
        case "read_messages": return "Read messages"
        case "request_payment": return "Request payments"
        case "upload_file": return "Upload files"
        case "read_chain_state": return "Read public chain state"
        default: return perm.replacingOccurrences(of: "_", with: " ").capitalized
        }
    }
}
#endif
