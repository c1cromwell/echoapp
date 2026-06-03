#if os(iOS)
import SwiftUI

/// Wave 0.5 moat — short integrity explainer (metagraph anchor, no oversell).
struct IntegrityExplainerView: View {
    var body: some View {
        List {
            Section {
                Text("ECHO threads are end-to-end encrypted on your device. Optional integrity anchors record only cryptographic hashes on the Constellation metagraph — never message bodies.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            Section("What is anchored") {
                Label("Username → DID binding", systemImage: "at")
                Label("Trust tier updates", systemImage: "checkmark.shield")
                Label("Evidence event hashes (when enabled)", systemImage: "link")
            }

            Section("What is never on chain") {
                Label("Message text or media", systemImage: "xmark.circle")
                Label("Phone contacts or address book", systemImage: "xmark.circle")
                Label("Your private keys", systemImage: "xmark.circle")
            }

            Section {
                Text("This is different from Meta, X, or Telegram accounts tied to a phone number and a central database.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .navigationTitle("Integrity & privacy")
    }
}
#endif
