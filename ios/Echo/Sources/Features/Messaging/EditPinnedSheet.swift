#if os(iOS)
import SwiftUI

/// Reorder and toggle pins for the Messages hub (Phase B §5.5).
struct EditPinnedSheet: View {
    let conversations: [StoredConversation]
    @Environment(\.dismiss) private var dismiss
    @Bindable private var pinnedStore = PinnedConversationsStore.shared

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text("Pin up to \(PinnedConversationsStore.maxPins) conversations. Pinned chats stay at the top of your hub.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                Section("Conversations") {
                    ForEach(conversations) { conv in
                        HStack {
                            VStack(alignment: .leading, spacing: 2) {
                                Text(conv.contactName)
                                    .font(.headline)
                                Text(ContactThreadHelper.truncatedDID(conv.peerDID))
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            if pinnedStore.isPinned(conv.id) {
                                Image(systemName: "pin.fill")
                                    .foregroundStyle(Color.echoSignal)
                            }
                        }
                        .contentShape(Rectangle())
                        .onTapGesture { pinnedStore.toggle(conv.id) }
                    }
                }

                if !pinnedStore.orderedIDs.isEmpty {
                    Section("Pinned order") {
                        ForEach(pinnedStore.orderedIDs, id: \.self) { id in
                            if let conv = conversations.first(where: { $0.id == id }) {
                                Text(conv.contactName)
                            }
                        }
                        .onMove { from, to in
                            pinnedStore.move(from: from, to: to)
                        }
                        .onDelete { offsets in
                            for index in offsets {
                                let id = pinnedStore.orderedIDs[index]
                                pinnedStore.unpin(id)
                            }
                        }
                    }
                }
            }
            .navigationTitle("Pinned chats")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
                ToolbarItem(placement: .primaryAction) {
                    EditButton()
                }
            }
        }
    }
}
#endif
