#if os(iOS)
import SwiftUI

/// Channels tab in the Messages hub — discover, create, and engage for ECHO rewards.
struct ChannelsListView: View {
    @State private var channels: [BroadcastChannelWire] = []
    @State private var isLoading = true
    @State private var errorMessage: String?
    @State private var showCreate = false
    @State private var selectedChannel: BroadcastChannelWire?
    @State private var claimStatus: String?

    var body: some View {
        VStack(spacing: 0) {
            engagementBanner

            if isLoading && channels.isEmpty {
                ProgressView("Loading channels…")
                    .frame(maxWidth: .infinity)
                    .padding(.top, 48)
            } else if let errorMessage, channels.isEmpty {
                EmptyStateView(
                    icon: "wifi.exclamationmark",
                    title: "Couldn't load channels",
                    subtitle: errorMessage,
                    actionTitle: "Retry",
                    action: { Task { await refresh() } }
                )
                .frame(minHeight: 280)
            } else if channels.isEmpty {
                EmptyStateView(
                    icon: "megaphone",
                    title: "No communities yet",
                    subtitle: "Create a community channel and earn ECHO when people engage.",
                    actionTitle: "Create channel",
                    action: { showCreate = true }
                )
                .frame(minHeight: 280)
            } else {
                LazyVStack(spacing: 0) {
                    ForEach(channels) { channel in
                        Button {
                            selectedChannel = channel
                        } label: {
                            channelRow(channel)
                        }
                        .buttonStyle(.plain)
                        Divider().padding(.leading, 72)
                    }
                }
            }
        }
        .task { await refresh() }
        .sheet(isPresented: $showCreate) {
            CreateChannelSheet { created in
                showCreate = false
                if let created {
                    channels.insert(created, at: 0)
                    selectedChannel = created
                }
            }
        }
        .sheet(item: $selectedChannel) { channel in
            NavigationStack {
                ChannelDetailView(channel: channel)
            }
        }
    }

    private var engagementBanner: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Image(systemName: "gift.fill")
                    .foregroundStyle(Color.echoTrustGreen)
                Text("Earn ECHO for engagement")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(Color.echoInk)
                Spacer()
                Button("Create") { showCreate = true }
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.echoSignal)
            }
            Text("Post, react, and subscribe in community channels to accrue rewards. Claim from Rewards when ready.")
                .font(.system(size: 13))
                .foregroundStyle(Color.echoInk55)
            if let claimStatus {
                Text(claimStatus)
                    .font(.caption)
                    .foregroundStyle(Color.echoTrustGreen)
            }
            Button {
                Task {
                    let ok = await ChannelsAPIClient.claimChannelEngagement(trustTier: 3)
                    claimStatus = ok ? "Engagement rewards claimed" : "Nothing to claim yet (need Tier 2+)"
                }
            } label: {
                Text("Claim channel rewards")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 10)
                    .background(Color.echoSignal)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
            .buttonStyle(.plain)
        }
        .padding(16)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }

    private func channelRow(_ channel: BroadcastChannelWire) -> some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(Color.echoSignal.opacity(0.15))
                    .frame(width: 48, height: 48)
                Image(systemName: "megaphone.fill")
                    .foregroundStyle(Color.echoSignal)
            }
            VStack(alignment: .leading, spacing: 2) {
                Text(channel.name)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(Color.echoInk)
                Text(channel.topic ?? channel.channelType ?? "community")
                    .font(.system(size: 13))
                    .foregroundStyle(Color.echoInk55)
                    .lineLimit(1)
            }
            Spacer()
            VStack(alignment: .trailing, spacing: 2) {
                Text("\(channel.subscriberCount ?? 0)")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(Color.echoInk)
                Text("members")
                    .font(.caption2)
                    .foregroundStyle(Color.echoInk40)
            }
            Image(systemName: "chevron.right")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.echoInk40)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }

    @MainActor
    private func refresh() async {
        isLoading = true
        errorMessage = nil
        let list = await ChannelsAPIClient.listChannels()
        channels = list
        isLoading = false
        if list.isEmpty {
            // Empty can mean no channels or auth failure — keep UI calm.
        }
    }
}

// MARK: - Create

private struct CreateChannelSheet: View {
    let onDone: (BroadcastChannelWire?) -> Void
    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var topic = ""
    @State private var isSaving = false
    @State private var error: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("Community") {
                    TextField("Name", text: $name)
                    TextField("Topic", text: $topic)
                }
                Section {
                    Text("Creating a community channel auto-subscribes you. Posts and reactions earn ECHO engagement rewards (Tier 2+).")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                if let error {
                    Text(error).foregroundStyle(.red)
                }
            }
            .navigationTitle("New channel")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss(); onDone(nil) }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task { await create() }
                    }
                    .disabled(name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || isSaving)
                }
            }
        }
    }

    private func create() async {
        isSaving = true
        error = nil
        let created = await ChannelsAPIClient.createChannel(
            name: name.trimmingCharacters(in: .whitespacesAndNewlines),
            topic: topic.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                ? "general"
                : topic.trimmingCharacters(in: .whitespacesAndNewlines),
            channelType: "community"
        )
        isSaving = false
        if let created {
            onDone(created)
            dismiss()
        } else {
            error = "Couldn't create channel. Check you're signed in."
        }
    }
}

// MARK: - Detail

struct ChannelDetailView: View {
    let channel: BroadcastChannelWire
    @State private var posts: [BroadcastPostWire] = []
    @State private var draft = ""
    @State private var isLoading = true
    @State private var isSubscribed = false
    @State private var statusMessage: String?
    @State private var currentDID = ""
    @State private var showAdmin = false
    @State private var commentsPost: BroadcastPostWire?

    private var isOwner: Bool {
        !currentDID.isEmpty && channel.creatorId == currentDID
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            if isLoading {
                ProgressView().padding()
            }
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 12) {
                    if let statusMessage {
                        Text(statusMessage)
                            .font(.caption)
                            .foregroundStyle(Color.echoTrustGreen)
                            .padding(.horizontal)
                    }
                    ForEach(posts) { post in
                        postCard(post)
                    }
                    if posts.isEmpty && !isLoading {
                        Text("No posts yet — be the first.")
                            .font(.subheadline)
                            .foregroundStyle(Color.echoInk55)
                            .padding()
                    }
                }
                .padding(.vertical, 12)
            }
            composer
        }
        .background(Color.echoPaper)
        .navigationTitle(channel.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if isOwner {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        showAdmin = true
                    } label: {
                        Image(systemName: "slider.horizontal.3")
                    }
                    .accessibilityLabel("Manage channel")
                }
            }
        }
        .sheet(isPresented: $showAdmin) {
            NavigationStack { ChannelAdminView(channel: channel) }
        }
        .sheet(item: $commentsPost) { post in
            NavigationStack {
                ChannelPostCommentsView(channel: channel, post: post, currentDID: currentDID)
            }
        }
        .task {
            currentDID = await CurrentUserSession.currentDID() ?? ""
            await load()
            isSubscribed = await ChannelsAPIClient.subscribe(channelId: channel.id)
        }
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(channel.topic ?? "community")
                .font(.caption.weight(.semibold))
                .foregroundStyle(Color.echoSignal)
            Text("\(channel.subscriberCount ?? 0) members · earn ECHO for posts & reactions")
                .font(.caption)
                .foregroundStyle(Color.echoInk55)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
    }

    private func postCard(_ post: BroadcastPostWire) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(post.content)
                .font(.body)
                .foregroundStyle(Color.echoInk)
            HStack {
                Button {
                    Task {
                        if await ChannelsAPIClient.react(postId: post.id) {
                            statusMessage = "+ECHO for reaction"
                        }
                    }
                } label: {
                    Label("React", systemImage: "hand.thumbsup")
                        .font(.caption.weight(.semibold))
                }
                .buttonStyle(.plain)
                .foregroundStyle(Color.echoSignal)

                Button {
                    commentsPost = post
                } label: {
                    Label("Comment", systemImage: "bubble.left")
                        .font(.caption.weight(.semibold))
                }
                .buttonStyle(.plain)
                .foregroundStyle(Color.echoInk55)

                Spacer()
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, 16)
    }

    private var composer: some View {
        HStack(spacing: 10) {
            TextField("Post to \(channel.name)…", text: $draft, axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .lineLimit(1...4)
            Button {
                Task { await sendPost() }
            } label: {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(
                        draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                            ? Color.echoInk40
                            : Color.echoSignal
                    )
            }
            .disabled(draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }
        .padding(12)
        .background(Color.echoPaperDim)
    }

    private func load() async {
        isLoading = true
        posts = await ChannelsAPIClient.listPosts(channelId: channel.id)
        isLoading = false
    }

    private func sendPost() async {
        let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        if let post = await ChannelsAPIClient.createPost(channelId: channel.id, content: text) {
            posts.insert(post, at: 0)
            draft = ""
            statusMessage = "+ECHO for post"
        } else {
            statusMessage = "Couldn't post — subscribe first or check connection"
        }
    }
}

/// Discussion thread for a single channel post: view + add comments, delete own
/// (or as owner/admin). Backed by the /v3/broadcasts/comment(s) endpoints.
struct ChannelPostCommentsView: View {
    let channel: BroadcastChannelWire
    let post: BroadcastPostWire
    let currentDID: String

    @Environment(\.dismiss) private var dismiss
    @State private var comments: [ChannelCommentWire] = []
    @State private var draft = ""
    @State private var isLoading = true
    @State private var isSending = false
    @State private var errorMessage: String?
    @Bindable private var conversationStore = ConversationStore.shared

    private var isOwner: Bool { !currentDID.isEmpty && channel.creatorId == currentDID }
    private var trimmedDraft: String { draft.trimmingCharacters(in: .whitespacesAndNewlines) }

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 10) {
                    Text(post.content)
                        .font(.subheadline)
                        .foregroundStyle(Color.echoInk55)
                        .padding(.bottom, 4)
                    Divider()
                    if comments.isEmpty && !isLoading {
                        Text("No comments yet — start the discussion.")
                            .font(.subheadline)
                            .foregroundStyle(Color.echoInk55)
                            .padding(.top, 8)
                    }
                    ForEach(comments) { comment in
                        commentRow(comment)
                    }
                }
                .padding(16)
            }
            composer
        }
        .background(Color.echoPaper)
        .navigationTitle("Comments")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .confirmationAction) { Button("Done") { dismiss() } }
        }
        .overlay { if isLoading && comments.isEmpty { ProgressView() } }
        .alert(
            "Something went wrong",
            isPresented: Binding(get: { errorMessage != nil }, set: { if !$0 { errorMessage = nil } }),
            presenting: errorMessage
        ) { _ in
            Button("OK", role: .cancel) { errorMessage = nil }
        } message: { message in
            Text(message)
        }
        .task { await load() }
    }

    private func commentRow(_ comment: ChannelCommentWire) -> some View {
        let canDelete = isOwner || comment.authorId == currentDID
        return HStack(alignment: .top, spacing: 8) {
            VStack(alignment: .leading, spacing: 2) {
                Text(ChannelParticipantLabel.resolve(
                    did: comment.authorId ?? "",
                    displayName: comment.displayName,
                    contactName: conversationStore.conversations.first(where: { $0.peerDID == comment.authorId })?.contactName,
                    currentDID: currentDID
                ))
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(Color.echoInk70)
                Text(comment.content)
                    .font(.body)
                    .foregroundStyle(Color.echoInk)
            }
            Spacer()
            if canDelete {
                Button {
                    Task {
                        let ok = await ChannelsAPIClient.deleteComment(commentId: comment.id)
                        if ok { await load() } else { errorMessage = "Couldn't delete this comment. Try again." }
                    }
                } label: {
                    Image(systemName: "trash")
                        .font(.system(size: 15))
                        .frame(width: 44, height: 44)
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .foregroundStyle(Color.echoError)
            }
        }
        .padding(.vertical, 4)
    }

    private var composer: some View {
        HStack(spacing: 10) {
            TextField("Add a comment…", text: $draft, axis: .vertical)
                .lineLimit(1...4)
                .padding(.horizontal, 14)
                .padding(.vertical, 9)
                .background(Color.echoPaper)
                .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
                .overlay(
                    RoundedRectangle(cornerRadius: 18, style: .continuous)
                        .strokeBorder(Color.echoInk.opacity(0.12), lineWidth: 1)
                )
            Button {
                Task { await send() }
            } label: {
                if isSending {
                    ProgressView().frame(width: 30, height: 30)
                } else {
                    Image(systemName: "arrow.up.circle.fill")
                        .font(.system(size: 30))
                        .foregroundStyle(trimmedDraft.isEmpty ? Color.echoInk40 : Color.echoSignal)
                }
            }
            .disabled(trimmedDraft.isEmpty || isSending)
        }
        .padding(12)
        .background(Color.echoPaperDim)
    }

    private func load() async {
        isLoading = true
        comments = await ChannelsAPIClient.listComments(postId: post.id)
        isLoading = false
    }

    private func send() async {
        let text = trimmedDraft
        guard !text.isEmpty, !isSending else { return }
        isSending = true
        defer { isSending = false }
        if let c = await ChannelsAPIClient.addComment(channelId: channel.id, postId: post.id, content: text) {
            comments.append(c)
            draft = ""
        } else {
            errorMessage = "Your comment couldn't be posted. Check your connection and try again."
        }
    }

}
#endif
