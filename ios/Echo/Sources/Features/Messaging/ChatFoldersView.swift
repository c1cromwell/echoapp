#if os(iOS)
import SwiftUI

/// Manage Telegram-style chat folders: create / rename / delete folders and
/// choose which conversations belong to each. Backed by `ChatFolderStore`
/// (local + `/v3/chat-folders` sync). Presented as a sheet from the Messages hub.
struct ChatFoldersView: View {
    /// Direct conversations available to assign (names for the membership picker).
    let conversations: [StoredConversation]

    @Bindable private var store = ChatFolderStore.shared
    @Environment(\.dismiss) private var dismiss

    @State private var showCreate = false
    @State private var newFolderName = ""
    @State private var renameTarget: CustomChatFolder?
    @State private var renameText = ""
    @State private var deleteTarget: CustomChatFolder?

    var body: some View {
        NavigationStack {
            List {
                Section {
                    if store.sortedFolders.isEmpty {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("No folders yet")
                                .foregroundStyle(Color.echoPrimaryText)
                            Text("Group your chats into tabs — like Work, Family, or Unread — that appear above your conversation list.")
                                .font(.caption)
                                .foregroundStyle(Color.echoSecondaryText)
                        }
                        .padding(.vertical, 4)
                    }
                    ForEach(store.sortedFolders) { folder in
                        NavigationLink {
                            FolderMembershipView(folderId: folder.id, conversations: conversations)
                        } label: {
                            HStack {
                                Image(systemName: "folder.fill").foregroundStyle(Color.echoSignal)
                                Text(folder.name)
                                Spacer()
                                Text("\(folder.conversationIds.count)")
                                    .foregroundStyle(Color.echoSecondaryText)
                            }
                        }
                        .swipeActions(edge: .trailing) {
                            Button(role: .destructive) { deleteTarget = folder } label: {
                                Label("Delete", systemImage: "trash")
                            }
                            Button {
                                renameText = folder.name
                                renameTarget = folder
                            } label: {
                                Label("Rename", systemImage: "pencil")
                            }
                            .tint(.echoSignal)
                        }
                    }
                } header: {
                    Text("Folders")
                } footer: {
                    Text("Folders appear as tabs above your chats. A conversation can be in more than one folder.")
                }

                Section {
                    Button {
                        newFolderName = ""
                        showCreate = true
                    } label: {
                        Label("New Folder", systemImage: "plus.circle.fill")
                    }
                    .accessibilityIdentifier("folders.new")
                }
            }
            .navigationTitle("Chat Folders")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .alert("New Folder", isPresented: $showCreate) {
                TextField("Folder name", text: $newFolderName)
                Button("Cancel", role: .cancel) {}
                Button("Create") {
                    let name = newFolderName.trimmingCharacters(in: .whitespacesAndNewlines)
                    if !name.isEmpty { store.createFolder(name: name) }
                }
            }
            .alert("Rename Folder", isPresented: Binding(
                get: { renameTarget != nil },
                set: { if !$0 { renameTarget = nil } }
            )) {
                TextField("Folder name", text: $renameText)
                Button("Cancel", role: .cancel) { renameTarget = nil }
                Button("Save") {
                    if let t = renameTarget { store.renameFolder(id: t.id, name: renameText) }
                    renameTarget = nil
                }
            }
            .confirmationDialog(
                "Delete folder?",
                isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }),
                presenting: deleteTarget
            ) { folder in
                Button("Delete \(folder.name)", role: .destructive) {
                    store.deleteFolder(id: folder.id)
                    deleteTarget = nil
                }
                Button("Cancel", role: .cancel) { deleteTarget = nil }
            } message: { _ in
                Text("Your conversations stay — only the folder is removed.")
            }
        }
    }
}

/// Toggle which conversations belong to a folder.
private struct FolderMembershipView: View {
    let folderId: String
    let conversations: [StoredConversation]

    @Bindable private var store = ChatFolderStore.shared

    var body: some View {
        List {
            if conversations.isEmpty {
                Text("No conversations yet.").foregroundStyle(Color.echoSecondaryText)
            }
            ForEach(conversations) { conv in
                Button {
                    store.toggleConversation(conv.id, in: folderId)
                } label: {
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(conv.contactName).foregroundStyle(Color.echoPrimaryText)
                            if !conv.lastMessage.isEmpty {
                                Text(conv.lastMessage)
                                    .font(.caption)
                                    .foregroundStyle(Color.echoSecondaryText)
                                    .lineLimit(1)
                            }
                        }
                        Spacer()
                        if store.contains(conv.id, in: folderId) {
                            Image(systemName: "checkmark.circle.fill").foregroundStyle(Color.echoSignal)
                        } else {
                            Image(systemName: "circle").foregroundStyle(Color.echoSecondaryText)
                        }
                    }
                }
                .buttonStyle(.plain)
            }
        }
        .navigationTitle(store.folder(id: folderId)?.name ?? "Folder")
        .navigationBarTitleDisplayMode(.inline)
    }
}
#endif
