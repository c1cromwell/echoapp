#if os(iOS)
import Foundation

/// Lightweight GIF search (Signal Parity Wave S4) — no Signal sticker CDN.
/// Uses the public Giphy/Tenor-style URL only when `ECHO_GIF_API_KEY` is set in the
/// process environment; otherwise returns curated local placeholder suggestions.
actor GifSearchService {
    struct GifResult: Identifiable, Sendable, Hashable {
        let id: String
        let previewURL: URL?
        let title: String
    }

    static let shared = GifSearchService()

    func search(query: String) async -> [GifResult] {
        let q = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !q.isEmpty else { return Self.placeholders }
        // Optional remote search — keep offline-safe defaults.
        if let key = ProcessInfo.processInfo.environment["ECHO_GIF_API_KEY"], !key.isEmpty,
           let encoded = q.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed),
           let url = URL(string: "https://api.giphy.com/v1/gifs/search?api_key=\(key)&q=\(encoded)&limit=12") {
            do {
                let (data, response) = try await URLSession.shared.data(from: url)
                guard let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode) else {
                    return Self.placeholders
                }
                return Self.parseGiphy(data) ?? Self.placeholders
            } catch {
                return Self.placeholders
            }
        }
        return Self.placeholders.filter { $0.title.localizedCaseInsensitiveContains(q) }
            + Self.placeholders
    }

    private static let placeholders: [GifResult] = [
        GifResult(id: "wave", previewURL: nil, title: "Wave"),
        GifResult(id: "thumbsup", previewURL: nil, title: "Thumbs up"),
        GifResult(id: "celebrate", previewURL: nil, title: "Celebrate"),
        GifResult(id: "thanks", previewURL: nil, title: "Thanks"),
    ]

    private static func parseGiphy(_ data: Data) -> [GifResult]? {
        guard
            let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
            let arr = json["data"] as? [[String: Any]]
        else { return nil }
        return arr.compactMap { item in
            guard let id = item["id"] as? String else { return nil }
            let title = (item["title"] as? String)?.nilIfEmpty ?? id
            let images = item["images"] as? [String: Any]
            let fixed = images?["fixed_height_small"] as? [String: Any]
            let urlStr = fixed?["url"] as? String
            return GifResult(id: id, previewURL: urlStr.flatMap(URL.init(string:)), title: title)
        }
    }
}

private extension String {
    var nilIfEmpty: String? { isEmpty ? nil : self }
}
#endif
