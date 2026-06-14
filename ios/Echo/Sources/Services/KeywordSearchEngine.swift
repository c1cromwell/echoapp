#if os(iOS)
import Foundation

struct SearchHit: Sendable, Equatable, Identifiable {
    var id: String { messageId }
    let messageId: String
    let conversationId: String
    let senderDID: String
    let matchedText: String
    let timestamp: Date
    let contentType: String
    let score: Double
}

/// Keyword search over local encrypted index (WO-16 / WO-197).
actor KeywordSearchEngine {
    static let shared = KeywordSearchEngine()

    private let indexer = LocalMessageIndexer.shared

    func search(query: String, filter: SearchFilter = .all) async -> [SearchHit] {
        let terms = MessageSearchTokenizer.parseQuery(query)
        guard !terms.isEmpty else { return [] }
        let snapshot = await indexer.currentSnapshot()
        var scores: [String: Double] = [:]
        var matchedTerms: [String: Set<String>] = [:]

        for (idx, term) in terms.enumerated() {
            let postings = snapshot.postings[term] ?? fuzzyPostings(for: term, in: snapshot)
            let idf = log(1 + Double(snapshot.documents.count) / Double(max(1, postings.count)))
            for posting in postings {
                guard let doc = snapshot.documents[posting.messageId] else { continue }
                if !matchesFilter(doc.contentType, filter: filter) { continue }
                let tf = 1.0 + Double(snapshot.postings[term]?.filter { $0.messageId == posting.messageId }.count ?? 1)
                let recency = recencyBoost(timestamp: posting.timestamp)
                let weight = tf * idf * recency * (1.0 / Double(idx + 1))
                scores[posting.messageId, default: 0] += weight
                matchedTerms[posting.messageId, default: []].insert(term)
            }
        }

        let requiredMatches = terms.count
        return scores.compactMap { messageId, score in
            guard matchedTerms[messageId]?.count == requiredMatches,
                  let doc = snapshot.documents[messageId] else { return nil }
            return SearchHit(
                messageId: messageId,
                conversationId: doc.conversationId,
                senderDID: doc.senderDID,
                matchedText: highlight(doc.bodyPreview, terms: Array(matchedTerms[messageId] ?? [])),
                timestamp: Date(timeIntervalSince1970: doc.timestamp),
                contentType: doc.contentType,
                score: score
            )
        }
        .sorted { $0.score > $1.score }
        .prefix(50)
        .map { $0 }
    }

    private func fuzzyPostings(for term: String, in snapshot: SearchIndexSnapshot) -> [SearchPosting] {
        guard term.count >= 5 else { return [] }
        var out: [SearchPosting] = []
        for (key, postings) in snapshot.postings where key != term {
            if levenshtein(key, term) <= 2 {
                out.append(contentsOf: postings)
            }
        }
        return out
    }

    private func levenshtein(_ a: String, _ b: String) -> Int {
        let aChars = Array(a), bChars = Array(b)
        var dist = Array(repeating: Array(repeating: 0, count: bChars.count + 1), count: aChars.count + 1)
        for i in 0...aChars.count { dist[i][0] = i }
        for j in 0...bChars.count { dist[0][j] = j }
        for i in 1...aChars.count {
            for j in 1...bChars.count {
                let cost = aChars[i - 1] == bChars[j - 1] ? 0 : 1
                dist[i][j] = min(dist[i - 1][j] + 1, dist[i][j - 1] + 1, dist[i - 1][j - 1] + cost)
            }
        }
        return dist[aChars.count][bChars.count]
    }

    private func recencyBoost(timestamp: TimeInterval) -> Double {
        let ageDays = max(0, Date().timeIntervalSince1970 - timestamp) / 86_400
        return 1.0 / (1.0 + ageDays / 30.0)
    }

    private func matchesFilter(_ contentType: String, filter: SearchFilter) -> Bool {
        switch filter {
        case .all: return true
        case .files: return contentType == "file"
        case .photos: return contentType == "image"
        case .links: return contentType == "link"
        case .voice: return contentType == "audio"
        }
    }

    private func highlight(_ text: String, terms: [String]) -> String {
        guard !terms.isEmpty else { return text }
        let lower = text.lowercased()
        for term in terms {
            if let range = lower.range(of: term) {
                let start = max(lower.startIndex, lower.index(range.lowerBound, offsetBy: -50, limitedBy: lower.startIndex) ?? lower.startIndex)
                let end = min(lower.endIndex, lower.index(range.upperBound, offsetBy: 50, limitedBy: lower.endIndex) ?? lower.endIndex)
                return String(text[start..<end])
            }
        }
        return String(text.prefix(100))
    }
}
#endif
