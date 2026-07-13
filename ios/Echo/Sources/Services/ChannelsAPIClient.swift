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
}
#endif
