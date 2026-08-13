#if os(iOS)
import Foundation

struct BroadcastChannelWire: Codable, Identifiable, Sendable, Hashable {
    let id: String
    let name: String
    let topic: String?
    let channelType: String?
    let subscriberCount: Int?
    let description: String?
    let creatorId: String?
}

struct BroadcastListResponse: Codable, Sendable {
    let channels: [BroadcastChannelWire]
}

struct BroadcastPostsResponse: Codable, Sendable {
    let channelId: String
    let posts: [BroadcastPostWire]
}

struct BroadcastPostWire: Codable, Identifiable, Sendable, Hashable {
    let id: String
    let channelId: String
    let content: String
    let contentType: String?
    let creatorId: String?
}

// MARK: - Channel admin wire models

struct ChannelMemberWire: Codable, Identifiable, Sendable, Hashable {
    let subscriberId: String
    let role: String?
    let isMuted: Bool?
    var id: String { subscriberId }
}

struct ChannelMembersResponse: Codable, Sendable {
    let members: [ChannelMemberWire]
}

struct ChannelJoinRequestWire: Codable, Identifiable, Sendable, Hashable {
    let channelId: String
    let subscriberId: String
    let status: String?
    var id: String { subscriberId }
}

struct ChannelJoinRequestsResponse: Codable, Sendable {
    let requests: [ChannelJoinRequestWire]
}

struct ChannelPostsListResponse: Codable, Sendable {
    let posts: [BroadcastPostWire]
}

struct ChannelCommentWire: Codable, Identifiable, Sendable, Hashable {
    let id: String
    let channelId: String?
    let postId: String?
    let authorId: String?
    let content: String
}

struct ChannelCommentsResponse: Codable, Sendable {
    let comments: [ChannelCommentWire]
}

struct ChannelAnalyticsWire: Codable, Sendable {
    let totalSubscribers: Int?
    let postCount: Int?
    let averageEngagement: Double?
    let viewCount: Int?
}

/// Live broadcast channels client (JWT Bearer — matches gateway auth middleware).
enum ChannelsAPIClient {
    private static var baseURL: URL { APIConfiguration.default.baseURL }

    private static let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        return d
    }()

    static func listChannels(query: String? = nil) async -> [BroadcastChannelWire] {
        var path = "/v3/broadcasts/list"
        if let query, !query.isEmpty {
            let q = query.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? query
            path += "?q=\(q)"
        }
        guard let data = await get(path: path) else { return [] }
        return (try? decoder.decode(BroadcastListResponse.self, from: data))?.channels ?? []
    }

    static func listPosts(channelId: String) async -> [BroadcastPostWire] {
        let encoded = channelId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? channelId
        let path = "/v3/broadcasts/posts?channelId=\(encoded)"
        guard let data = await get(path: path) else { return [] }
        return (try? decoder.decode(BroadcastPostsResponse.self, from: data))?.posts ?? []
    }

    static func createChannel(name: String, topic: String, channelType: String = "community") async -> BroadcastChannelWire? {
        let body: [String: String] = [
            "name": name,
            "topic": topic,
            "channelType": channelType,
        ]
        guard let data = await post(path: "/v3/broadcasts/create", body: body) else { return nil }
        return try? decoder.decode(BroadcastChannelWire.self, from: data)
    }

    static func createPost(channelId: String, content: String) async -> BroadcastPostWire? {
        let body: [String: String] = [
            "channelId": channelId,
            "content": content,
            "contentType": "text",
        ]
        guard let data = await post(path: "/v3/broadcasts/post", body: body) else { return nil }
        return try? decoder.decode(BroadcastPostWire.self, from: data)
    }

    @discardableResult
    static func subscribe(channelId: String) async -> Bool {
        await post(path: "/v3/broadcasts/subscribe", body: ["channelId": channelId]) != nil
    }

    @discardableResult
    static func react(postId: String, emoji: String = "👍") async -> Bool {
        await post(path: "/v3/broadcasts/react", body: ["postId": postId, "emoji": emoji]) != nil
    }

    static func claimChannelEngagement(trustTier: Int) async -> Bool {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: baseURL.absoluteString + "/v3/rewards/claim"),
              let httpBody = try? JSONSerialization.data(withJSONObject: [
                  "rewardType": "channel_engagement",
                  "trustTier": trustTier,
              ] as [String: Any]) else {
            return false
        }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = httpBody
        guard let (_, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse else { return false }
        return (200...299).contains(http.statusCode)
    }

    // MARK: - Channel admin (owner/admin only)

    static func listMembers(channelId: String) async -> [ChannelMemberWire] {
        let encoded = channelId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? channelId
        guard let data = await get(path: "/v3/broadcasts/members?channelId=\(encoded)") else { return [] }
        return (try? decoder.decode(ChannelMembersResponse.self, from: data))?.members ?? []
    }

    @discardableResult
    static func setRole(channelId: String, memberId: String, role: String) async -> Bool {
        await post(path: "/v3/broadcasts/role",
                   body: ["channelId": channelId, "memberId": memberId, "role": role]) != nil
    }

    @discardableResult
    static func removeMember(channelId: String, memberId: String) async -> Bool {
        await post(path: "/v3/broadcasts/remove-member",
                   body: ["channelId": channelId, "memberId": memberId]) != nil
    }

    static func listJoinRequests(channelId: String) async -> [ChannelJoinRequestWire] {
        let encoded = channelId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? channelId
        guard let data = await get(path: "/v3/broadcasts/join-requests?channelId=\(encoded)") else { return [] }
        return (try? decoder.decode(ChannelJoinRequestsResponse.self, from: data))?.requests ?? []
    }

    @discardableResult
    static func approveMember(channelId: String, memberId: String) async -> Bool {
        await post(path: "/v3/broadcasts/approve",
                   body: ["channelId": channelId, "memberId": memberId]) != nil
    }

    @discardableResult
    static func denyMember(channelId: String, memberId: String) async -> Bool {
        await post(path: "/v3/broadcasts/deny",
                   body: ["channelId": channelId, "memberId": memberId]) != nil
    }

    static func analytics(channelId: String) async -> ChannelAnalyticsWire? {
        let encoded = channelId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? channelId
        guard let data = await get(path: "/v3/broadcasts/analytics?channelId=\(encoded)") else { return nil }
        return try? decoder.decode(ChannelAnalyticsWire.self, from: data)
    }

    // MARK: - Comments (discussion)

    static func listComments(postId: String) async -> [ChannelCommentWire] {
        let encoded = postId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? postId
        guard let data = await get(path: "/v3/broadcasts/comments?postId=\(encoded)") else { return [] }
        return (try? decoder.decode(ChannelCommentsResponse.self, from: data))?.comments ?? []
    }

    static func addComment(channelId: String, postId: String, content: String) async -> ChannelCommentWire? {
        guard let data = await post(path: "/v3/broadcasts/comment",
                                    body: ["channelId": channelId, "postId": postId, "content": content]) else { return nil }
        return try? decoder.decode(ChannelCommentWire.self, from: data)
    }

    @discardableResult
    static func deleteComment(commentId: String) async -> Bool {
        await post(path: "/v3/broadcasts/comment-delete", body: ["commentId": commentId]) != nil
    }

    // MARK: - Moderation (owner/admin only)

    @discardableResult
    static func setMuted(channelId: String, memberId: String, muted: Bool) async -> Bool {
        await postJSON(path: "/v3/broadcasts/mute",
                       body: ["channelId": channelId, "memberId": memberId, "muted": muted]) != nil
    }

    @discardableResult
    static func blockMember(channelId: String, memberId: String) async -> Bool {
        await post(path: "/v3/broadcasts/block",
                   body: ["channelId": channelId, "memberId": memberId]) != nil
    }

    static func listPendingPosts(channelId: String) async -> [BroadcastPostWire] {
        let encoded = channelId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? channelId
        guard let data = await get(path: "/v3/broadcasts/pending-posts?channelId=\(encoded)") else { return [] }
        return (try? decoder.decode(ChannelPostsListResponse.self, from: data))?.posts ?? []
    }

    @discardableResult
    static func decidePost(channelId: String, postId: String, approve: Bool, reason: String = "") async -> Bool {
        await postJSON(path: "/v3/broadcasts/post-decision",
                       body: ["channelId": channelId, "postId": postId, "approve": approve, "reason": reason]) != nil
    }

    private static func get(path: String) async -> Data? {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: baseURL.absoluteString + path) else { return nil }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse,
              (200...299).contains(http.statusCode) else { return nil }
        return data
    }

    private static func post(path: String, body: [String: String]) async -> Data? {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: baseURL.absoluteString + path),
              let httpBody = try? JSONEncoder().encode(body) else { return nil }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = httpBody
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse,
              (200...299).contains(http.statusCode) else { return nil }
        return data
    }

    /// POST with a mixed-type JSON body (bools/strings), for admin/moderation endpoints.
    private static func postJSON(path: String, body: [String: Any]) async -> Data? {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: baseURL.absoluteString + path),
              let httpBody = try? JSONSerialization.data(withJSONObject: body) else { return nil }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = httpBody
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse,
              (200...299).contains(http.statusCode) else { return nil }
        return data
    }
}
#endif
