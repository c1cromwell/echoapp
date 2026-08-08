#if os(iOS)
import SwiftUI

/// Local, on-device notification preferences (UserDefaults-backed). These gate
/// what the app surfaces locally; server-driven push routing is a separate
/// backend concern. Mirrors the `PrivacySettingsStore` pattern.
enum NotificationPreferencesStore {
    enum Key: String, CaseIterable {
        case messages = "echo.notify.messages"
        case previews = "echo.notify.previews"
        case sounds = "echo.notify.sounds"
        case groupMentionsOnly = "echo.notify.groupMentionsOnly"
        case contactRequests = "echo.notify.contactRequests"
        case rewards = "echo.notify.rewards"

        /// Default when the user hasn't set a value yet.
        var defaultValue: Bool {
            switch self {
            case .groupMentionsOnly: return false
            default: return true
            }
        }
    }

    static func value(_ key: Key) -> Bool {
        UserDefaults.standard.object(forKey: key.rawValue) as? Bool ?? key.defaultValue
    }

    static func set(_ value: Bool, for key: Key) {
        UserDefaults.standard.set(value, forKey: key.rawValue)
    }
}

struct NotificationPreferencesView: View {
    @Environment(\.dismiss) private var dismiss

    @State private var messages = NotificationPreferencesStore.value(.messages)
    @State private var previews = NotificationPreferencesStore.value(.previews)
    @State private var sounds = NotificationPreferencesStore.value(.sounds)
    @State private var groupMentionsOnly = NotificationPreferencesStore.value(.groupMentionsOnly)
    @State private var contactRequests = NotificationPreferencesStore.value(.contactRequests)
    @State private var rewards = NotificationPreferencesStore.value(.rewards)

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                EchoNavBar(title: "Notifications", showBackButton: true, onBackPressed: { dismiss() })

                ScrollView {
                    VStack(spacing: Spacing.lg.rawValue) {
                        SettingsSectionView(title: "Messages") {
                            VStack(spacing: 0) {
                                toggleRow("Message notifications", "Alert me about new direct messages.",
                                          $messages, .messages)
                                Divider().padding(.leading, Spacing.lg.rawValue)
                                toggleRow("Show previews", "Include a snippet of the message.",
                                          $previews, .previews)
                                    .disabled(!messages)
                                Divider().padding(.leading, Spacing.lg.rawValue)
                                toggleRow("In-app sounds", "Play a sound for new activity.",
                                          $sounds, .sounds)
                            }
                        }

                        SettingsSectionView(title: "Groups") {
                            toggleRow("Mentions only", "Only notify me when I'm @mentioned in groups.",
                                      $groupMentionsOnly, .groupMentionsOnly)
                        }

                        SettingsSectionView(title: "Other") {
                            VStack(spacing: 0) {
                                toggleRow("Contact requests", "Notify me about new message requests.",
                                          $contactRequests, .contactRequests)
                                Divider().padding(.leading, Spacing.lg.rawValue)
                                toggleRow("Reward activity", "Streaks, quests, and ECHO rewards.",
                                          $rewards, .rewards)
                            }
                        }
                    }
                    .padding(Spacing.lg.rawValue)
                }
            }
        }
        .toolbar(.hidden, for: .navigationBar)
    }

    private func toggleRow(
        _ title: String,
        _ subtitle: String,
        _ binding: Binding<Bool>,
        _ key: NotificationPreferencesStore.Key
    ) -> some View {
        Toggle(isOn: binding.onChange { NotificationPreferencesStore.set($0, for: key) }) {
            VStack(alignment: .leading, spacing: 2) {
                Text(title).typographyStyle(.body, color: .echoPrimaryText)
                Text(subtitle)
                    .typographyStyle(.caption, color: .echoSecondaryText)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
        .tint(.echoSignal)
        .padding(Spacing.md.rawValue)
    }
}

private extension Binding where Value == Bool {
    /// Runs `action` after the underlying value changes (to persist it).
    func onChange(_ action: @escaping (Bool) -> Void) -> Binding<Bool> {
        Binding(
            get: { wrappedValue },
            set: { newValue in
                wrappedValue = newValue
                action(newValue)
            }
        )
    }
}
#endif
