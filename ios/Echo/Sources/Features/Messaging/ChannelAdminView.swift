#if os(iOS)
import SwiftUI

/// Owner/admin console for a broadcast channel: review analytics, approve/deny
/// pending join requests (for approval-gated channels), and manage member roles
/// (promote moderator/admin, demote) and removals. Backed by the /v3/broadcasts
/// admin endpoints. Presented from ChannelDetailView when the viewer owns it.
struct ChannelAdminView: View {
    let channel: BroadcastChannelWire

    @Environment(\.dismiss) private var dismiss
    @State private var members: [ChannelMemberWire] = []
    @State private var requests: [ChannelJoinRequestWire] = []
    @State private var pendingPosts: [BroadcastPostWire] = []
    @State private var analytics: ChannelAnalyticsWire?
    @State private var isLoading = true
    @State private var busyId: String?
    @State private var errorMessage: String?
    @State private var pendingRemoval: ChannelMemberWire?
    @State private var pendingBlock: ChannelMemberWire?

    private let roles = ["subscriber", "moderator", "admin"]

    var body: some View {
        List {
            if let analytics {
                Section("Overview") {
                    statRow("Subscribers", analytics.totalSubscribers ?? channel.subscriberCount ?? 0)
                    statRow("Posts", analytics.postCount ?? 0)
                    statRow("Views", analytics.viewCount ?? 0)
                    HStack {
                        Text("Avg. engagement")
                        Spacer()
                        Text(String(format: "%.0f%%", (analytics.averageEngagement ?? 0) * 100))
                            .foregroundStyle(Color.echoSecondaryText)
                    }
                }
            }

            if !requests.isEmpty {
                Section {
                    ForEach(requests) { req in
                        requestRow(req)
                    }
                } header: {
                    Text("Join requests (\(requests.count))")
                } footer: {
                    Text("This channel requires approval — new members wait here until you approve them.")
                }
            }

            if !pendingPosts.isEmpty {
                Section("Posts awaiting approval (\(pendingPosts.count))") {
                    ForEach(pendingPosts) { post in
                        pendingPostRow(post)
                    }
                }
            }

            Section("Members (\(members.count))") {
                if members.isEmpty && !isLoading {
                    Text("No members yet.").foregroundStyle(Color.echoSecondaryText)
                }
                ForEach(members) { member in
                    memberRow(member)
                }
            }
        }
        .navigationTitle("Manage \(channel.name)")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .confirmationAction) {
                Button("Done") { dismiss() }
            }
        }
        .overlay {
            if isLoading && members.isEmpty && requests.isEmpty {
                ProgressView()
            }
        }
        .confirmationDialog(
            "Remove this member?",
            isPresented: Binding(get: { pendingRemoval != nil }, set: { if !$0 { pendingRemoval = nil } }),
            presenting: pendingRemoval
        ) { member in
            Button("Remove", role: .destructive) { remove(member.subscriberId) }
            Button("Cancel", role: .cancel) {}
        } message: { member in
            Text("\(shortDID(member.subscriberId)) will lose access to this channel. They can rejoin later.")
        }
        .confirmationDialog(
            "Block this member?",
            isPresented: Binding(get: { pendingBlock != nil }, set: { if !$0 { pendingBlock = nil } }),
            presenting: pendingBlock
        ) { member in
            Button("Block", role: .destructive) { block(member.subscriberId) }
            Button("Cancel", role: .cancel) {}
        } message: { member in
            Text("\(shortDID(member.subscriberId)) will be removed and prevented from rejoining.")
        }
        .alert(
            "Action failed",
            isPresented: Binding(get: { errorMessage != nil }, set: { if !$0 { errorMessage = nil } }),
            presenting: errorMessage
        ) { _ in
            Button("OK", role: .cancel) { errorMessage = nil }
        } message: { message in
            Text(message)
        }
        .task { await reload() }
    }

    private func requestRow(_ req: ChannelJoinRequestWire) -> some View {
        let busy = busyId == req.subscriberId
        return HStack {
            Text(shortDID(req.subscriberId))
                .lineLimit(1).truncationMode(.middle)
            Spacer()
            Button("Approve") { decide(req.subscriberId, approve: true) }
                .buttonStyle(.borderedProminent)
                .tint(.echoSignal)
                .disabled(busy)
            Button("Deny") { decide(req.subscriberId, approve: false) }
                .buttonStyle(.bordered)
                .tint(.echoError)
                .disabled(busy)
        }
    }

    private func pendingPostRow(_ post: BroadcastPostWire) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(post.content).typographyStyle(.bodySmall, color: .echoPrimaryText)
            HStack {
                Button("Approve") { decidePost(post.id, approve: true) }
                    .buttonStyle(.borderedProminent).tint(.echoSignal)
                Button("Reject") { decidePost(post.id, approve: false) }
                    .buttonStyle(.bordered).tint(.echoError)
                Spacer()
            }
            .disabled(busyId == post.id)
        }
        .padding(.vertical, 2)
    }

    @ViewBuilder
    private func memberRow(_ member: ChannelMemberWire) -> some View {
        let isOwner = member.subscriberId == channel.creatorId
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(shortDID(member.subscriberId))
                    .lineLimit(1).truncationMode(.middle)
                Text(isOwner ? "Owner" : (member.role ?? "subscriber").capitalized)
                    .font(.caption)
                    .foregroundStyle(Color.echoSecondaryText)
            }
            Spacer()
            if busyId == member.subscriberId {
                ProgressView()
                    .frame(width: 44, height: 44)
            } else if !isOwner {
                Menu {
                    ForEach(roles, id: \.self) { role in
                        Button {
                            setRole(member.subscriberId, role: role)
                        } label: {
                            if member.role == role { Label(role.capitalized, systemImage: "checkmark") }
                            else { Text(role.capitalized) }
                        }
                    }
                    Divider()
                    Button((member.isMuted ?? false) ? "Unmute" : "Mute") {
                        setMuted(member.subscriberId, muted: !(member.isMuted ?? false))
                    }
                    Button("Block", role: .destructive) { pendingBlock = member }
                    Button("Remove", role: .destructive) { pendingRemoval = member }
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .font(.system(size: 20))
                        .foregroundStyle(Color.echoSignal)
                        .frame(width: 44, height: 44)
                        .contentShape(Rectangle())
                }
            }
        }
    }

    private func statRow(_ label: String, _ value: Int) -> some View {
        HStack {
            Text(label)
            Spacer()
            Text("\(value)").foregroundStyle(Color.echoSecondaryText)
        }
    }

    private func shortDID(_ did: String) -> String {
        guard did.count > 20 else { return did }
        return String(did.prefix(12)) + "…" + String(did.suffix(6))
    }

    // MARK: - Actions

    private func reload() async {
        isLoading = true
        async let m = ChannelsAPIClient.listMembers(channelId: channel.id)
        async let r = ChannelsAPIClient.listJoinRequests(channelId: channel.id)
        async let p = ChannelsAPIClient.listPendingPosts(channelId: channel.id)
        async let a = ChannelsAPIClient.analytics(channelId: channel.id)
        members = await m
        requests = await r
        pendingPosts = await p
        analytics = await a
        isLoading = false
    }

    /// Runs an admin action keyed by `id` (shows the row's spinner), surfaces a
    /// failure message if the call returns `false`, then refreshes.
    private func perform(_ id: String, failure: String, _ action: @escaping () async -> Bool) {
        busyId = id
        Task {
            let ok = await action()
            if !ok { errorMessage = failure }
            await reload()
            busyId = nil
        }
    }

    private func setMuted(_ memberId: String, muted: Bool) {
        perform(memberId, failure: muted ? "Couldn't mute this member." : "Couldn't unmute this member.") {
            await ChannelsAPIClient.setMuted(channelId: channel.id, memberId: memberId, muted: muted)
        }
    }

    private func block(_ memberId: String) {
        perform(memberId, failure: "Couldn't block this member.") {
            await ChannelsAPIClient.blockMember(channelId: channel.id, memberId: memberId)
        }
    }

    private func decidePost(_ postId: String, approve: Bool) {
        perform(postId, failure: "Couldn't update this post.") {
            await ChannelsAPIClient.decidePost(channelId: channel.id, postId: postId, approve: approve)
        }
    }

    private func decide(_ memberId: String, approve: Bool) {
        perform(memberId, failure: approve ? "Couldn't approve this request." : "Couldn't deny this request.") {
            approve
                ? await ChannelsAPIClient.approveMember(channelId: channel.id, memberId: memberId)
                : await ChannelsAPIClient.denyMember(channelId: channel.id, memberId: memberId)
        }
    }

    private func setRole(_ memberId: String, role: String) {
        perform(memberId, failure: "Couldn't change this member's role.") {
            await ChannelsAPIClient.setRole(channelId: channel.id, memberId: memberId, role: role)
        }
    }

    private func remove(_ memberId: String) {
        perform(memberId, failure: "Couldn't remove this member.") {
            await ChannelsAPIClient.removeMember(channelId: channel.id, memberId: memberId)
        }
    }
}
#endif
