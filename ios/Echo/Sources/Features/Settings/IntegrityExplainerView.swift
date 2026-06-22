#if os(iOS)
import SwiftUI

/// Wave 0.5 moat — short integrity explainer (metagraph anchor, no oversell).
struct IntegrityExplainerView: View {
    var body: some View {
        List {
            Section {
                Text("Your messages are private and stay on your device. We can optionally save a tamper-proof fingerprint of a conversation (never the message text) so it can be verified later.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            Section("What we can verify") {
                Label("Your username matches your account", systemImage: "at")
                Label("Trust level updates", systemImage: "checkmark.shield")
                Label("Conversation fingerprints (when enabled)", systemImage: "link")
            }

            Section("What we never store") {
                Label("Message text or media", systemImage: "xmark.circle")
                Label("Phone contacts or address book", systemImage: "xmark.circle")
                Label("Your private keys", systemImage: "xmark.circle")
            }

            Section {
                Text("You don't need a phone number or email to use ECHO — your account lives on your device.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .navigationTitle("Integrity & privacy")
    }
}
#endif
