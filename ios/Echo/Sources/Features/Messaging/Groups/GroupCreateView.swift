#if os(iOS)
import SwiftUI

/// Create a private group, pick members, and distribute the initial symmetric key (M2).
struct GroupCreateView: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable private var conversationStore = ConversationStore.shared
    @State private var viewModel: GroupCreateViewModel
    var onCreated: (String, String) -> Void

    init(viewModel: GroupCreateViewModel, onCreated: @escaping (String, String) -> Void) {
        _viewModel = State(initialValue: viewModel)
        self.onCreated = onCreated
    }

    private var contacts: [StoredConversation] {
        conversationStore.conversations
            .filter { !$0.peerDID.isEmpty && !$0.id.hasPrefix("group:") }
            .sorted { $0.contactName.localizedCompare($1.contactName) == .orderedAscending }
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Group") {
                    TextField("Name", text: $viewModel.name)
                }
                Section("Members") {
                    if contacts.isEmpty {
                        Text("Add contacts first, then create a group.")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    } else {
                        ForEach(contacts) { conv in
                            Button {
                                viewModel.toggleMember(conv.peerDID)
                            } label: {
                                HStack {
                                    Text(conv.contactName)
                                    Spacer()
                                    if viewModel.selectedPeerDIDs.contains(conv.peerDID) {
                                        Image(systemName: "checkmark.circle.fill")
                                            .foregroundStyle(Color.accentColor)
                                    }
                                }
                            }
                            .buttonStyle(.plain)
                        }
                    }
                }
                if let error = viewModel.errorMessage {
                    Section {
                        Text(error).foregroundStyle(.red).font(.footnote)
                    }
                }
            }
            .navigationTitle("New group")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task {
                            if let result = await viewModel.createGroup() {
                                onCreated(result.groupId, result.name)
                                dismiss()
                            }
                        }
                    }
                    .disabled(
                        viewModel.isSaving
                            || viewModel.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                            || viewModel.selectedPeerDIDs.isEmpty
                    )
                }
            }
        }
    }
}
#endif
