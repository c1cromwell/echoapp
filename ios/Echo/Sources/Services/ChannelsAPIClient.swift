#if os(iOS)
import Foundation

struct BroadcastChannelWire: Codable, Identifiable, Sendable {
    let id: String
    let name: String
    let topic: String?
    let channelType: String?
    let subscriberCount: Int?

    enum CodingKeys: String, CodingKey {
        case id, name, topic, channelType, subscriberCount
    }
}

struct BroadcastListResponse: Codable, Sendable {
    let channels: [BroadcastChannelWire]
}

struct BroadcastPostsResponse: Codable, Sendable {
    let channelId: String
    let posts: [BroadcastPostWire]
}

struct BroadcastPostWire: Codable, Identifiable, Sendable {
    let id: String
    let channelId: String
    let content: String
    let contentType: String?

    enum CodingKeys: String, CodingKey {
        case id, channelId, content, contentType
    }
}

enum ChannelsAPIClient {
    static func listChannels(query: String? = nil) async -> [BroadcastChannelWire] {
        guard let token = try? await KeychainManager.shared.getAuthToken() else { return [] }
        var path = "/v3/broadcasts/list"
        if let query, !query.isEmpty {
            path += "?q=\(query.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? query)"
        }
        guard let url = URL(string: APIConfiguration.default.baseURL.absoluteString + path) else { return [] }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse, http.statusCode == 200 else {
            return []
        }
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        guard let decoded = try? decoder.decode(BroadcastListResponse.self, from: data) else {
            return []
        }
        return decoded.channels
    }

    static func subscribe(channelId: String) async -> Bool {
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: APIConfiguration.default.baseURL.absoluteString + "/v3/broadcasts/subscribe") else {
            return false
        }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONEncoder().encode(["channelId": channelId])
        guard let (_, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse else { return false }
        return (200...299).contains(http.statusCode)
    }
}
#endif
