#if os(iOS)
import SwiftUI

struct GroupDetailView: View {
    @Bindable var viewModel: GroupDetailViewModel
    @Bindable private var conversationStore = ConversationStore.shared
    @State private var showAddMember = false

    private var contacts: [StoredConversation] {
        conversationStore.conversations
            .filter { conv in
                !conv.peerDID.isEmpty
                    && !conv.id.hasPrefix("group:")
                    && !viewModel.members.contains(where: { $0.memberId == conv.peerDID })
            }
            .sorted { $0.contactName.localizedCompare($1.contactName) == .orderedAscending }
    }

    var body: some View {
        List {
            Section("Members (\(viewModel.members.count))") {
                if viewModel.isLoading {
                    ProgressView()
                } else {
                    ForEach(viewModel.members) { member in
                        HStack {
                            VStack(alignment: .leading, spacing: 2) {
                                Text(displayName(for: member))
                                    .font(.body)
                                Text(member.role.capitalized)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                                Text(shortDID(member.memberId))
                                    .font(.caption2)
                                    .foregroundStyle(.tertiary)
                            }
                            Spacer()
                            if viewModel.canManageMembers,
                               member.memberId != viewModel.currentUserDID {
                                Button(role: .destructive) {
                                    Task { await viewModel.removeMember(member.memberId) }
                                } label: {
                                    Text("Remove")
                                }
                                .buttonStyle(.borderless)
                                .disabled(viewModel.isSaving)
                            }
                        }
                    }
                }
            }

            if viewModel.canManageMembers {
                Section("Add member") {
                    if contacts.isEmpty {
                        Text("No contacts available to add.")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    } else {
                        ForEach(contacts) { conv in
                            Button {
                                Task { await viewModel.addMember(peerDID: conv.peerDID) }
                            } label: {
                                HStack {
                                    Text(conv.contactName)
                                    Spacer()
                                    Image(systemName: "plus.circle")
                                }
                            }
                            .disabled(viewModel.isSaving)
                        }
                    }
                }
            }

            if let status = viewModel.statusMessage {
                Section {
                    Text(status).font(.footnote).foregroundStyle(.green)
                }
            }
            if let error = viewModel.errorMessage {
                Section {
                    Text(error).font(.footnote).foregroundStyle(.red)
                }
            }
        }
        .navigationTitle(viewModel.groupName)
        .navigationBarTitleDisplayMode(.inline)
        .task { await viewModel.loadMembers() }
        .refreshable { await viewModel.loadMembers() }
    }

    private func displayName(for member: GroupMemberWire) -> String {
        if member.memberId == viewModel.currentUserDID { return "You" }
        if let name = member.displayName, !name.isEmpty { return name }
        if let conv = conversationStore.conversations.first(where: { $0.peerDID == member.memberId }) {
            return conv.contactName
        }
        return shortDID(member.memberId)
    }

    private func shortDID(_ did: String) -> String {
        guard did.count > 20 else { return did }
        return String(did.prefix(18)) + "…"
    }
}
#endif
