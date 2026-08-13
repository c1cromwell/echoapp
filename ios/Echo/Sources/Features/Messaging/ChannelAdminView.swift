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
                            .foregroundStyle(.secondary)
                    }
                }
            }

            if !requests.isEmpty {
                Section {
                    ForEach(requests) { req in
                        HStack {
                            Text(shortDID(req.subscriberId))
                                .lineLimit(1).truncationMode(.middle)
                            Spacer()
                            Button("Approve") { decide(req.subscriberId, approve: true) }
                                .buttonStyle(.borderedProminent)
                                .tint(.echoSignal)
                                .disabled(busyId == req.subscriberId)
                            Button("Deny") { decide(req.subscriberId, approve: false) }
                                .buttonStyle(.bordered)
                                .tint(.echoError)
                                .disabled(busyId == req.subscriberId)
                        }
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
                        VStack(alignment: .leading, spacing: 8) {
                            Text(post.content).font(.subheadline)
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
                }
            }

            Section("Members (\(members.count))") {
                if members.isEmpty && !isLoading {
                    Text("No members yet.").foregroundStyle(.secondary)
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
        .task { await reload() }
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
                    .foregroundStyle(.secondary)
            }
            Spacer()
            if !isOwner {
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
                    Button("Block", role: .destructive) { block(member.subscriberId) }
                    Button("Remove", role: .destructive) { remove(member.subscriberId) }
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .foregroundStyle(Color.echoSignal)
                }
                .disabled(busyId == member.subscriberId)
            }
        }
    }

    private func statRow(_ label: String, _ value: Int) -> some View {
        HStack {
            Text(label)
            Spacer()
            Text("\(value)").foregroundStyle(.secondary)
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

    private func setMuted(_ memberId: String, muted: Bool) {
        busyId = memberId
        Task {
            _ = await ChannelsAPIClient.setMuted(channelId: channel.id, memberId: memberId, muted: muted)
            await reload()
            busyId = nil
        }
    }

    private func block(_ memberId: String) {
        busyId = memberId
        Task {
            _ = await ChannelsAPIClient.blockMember(channelId: channel.id, memberId: memberId)
            await reload()
            busyId = nil
        }
    }

    private func decidePost(_ postId: String, approve: Bool) {
        busyId = postId
        Task {
            _ = await ChannelsAPIClient.decidePost(channelId: channel.id, postId: postId, approve: approve)
            await reload()
            busyId = nil
        }
    }

    private func decide(_ memberId: String, approve: Bool) {
        busyId = memberId
        Task {
            _ = approve
                ? await ChannelsAPIClient.approveMember(channelId: channel.id, memberId: memberId)
                : await ChannelsAPIClient.denyMember(channelId: channel.id, memberId: memberId)
            await reload()
            busyId = nil
        }
    }

    private func setRole(_ memberId: String, role: String) {
        busyId = memberId
        Task {
            _ = await ChannelsAPIClient.setRole(channelId: channel.id, memberId: memberId, role: role)
            await reload()
            busyId = nil
        }
    }

    private func remove(_ memberId: String) {
        busyId = memberId
        Task {
            _ = await ChannelsAPIClient.removeMember(channelId: channel.id, memberId: memberId)
            await reload()
            busyId = nil
        }
    }
}
#endif
