#if os(iOS)
import SwiftUI

/// Tap reaction count to see who reacted (WO-10).
struct ReactionReactorDetailSheet: View {
    let emoji: String
    let reactors: [String]
    let currentUserDID: String
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text(emoji)
                        .font(.system(size: 32))
                        .frame(maxWidth: .infinity, alignment: .center)
                }
                Section("Reacted") {
                    ForEach(reactors, id: \.self) { did in
                        HStack {
                            Text(displayLabel(for: did))
                                .font(.body)
                            Spacer()
                            if did == currentUserDID {
                                Text("You")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
            }
            .navigationTitle("Reactions")
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
        }
    }

    private func displayLabel(for did: String) -> String {
        if did == currentUserDID { return "You" }
        if did.count > 24 {
            return "\(did.prefix(14))…\(did.suffix(6))"
        }
        return did
    }
}
#endif
