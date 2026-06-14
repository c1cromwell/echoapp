#if os(iOS)
import Foundation
import NaturalLanguage

struct SmartReplySuggestion: Sendable, Equatable, Identifiable {
    let id: String
    let text: String
}

struct ThreadSummary: Sendable, Equatable {
    let sentenceCount: Int
    let summary: String
    let generatedAt: Date
}

/// On-device privacy AI (M7a / WO-CA1). Uses NaturalLanguage only — no server plaintext.
actor OnDeviceAIService {
    static let shared = OnDeviceAIService()

    func smartReplies(from messages: [String], limit: Int = 3) -> [SmartReplySuggestion] {
        guard PrivacyAIConsentStore.load().smartRepliesEnabled else { return [] }
        let recent = messages.suffix(12).map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }.filter { !$0.isEmpty }
        guard !recent.isEmpty else { return [] }

        var candidates: [String] = []
        if let last = recent.last?.lowercased() {
            if last.contains("?") {
                candidates += ["Yes", "No", "Maybe later"]
            }
            if last.contains("thank") {
                candidates += ["You're welcome", "Anytime"]
            }
            if last.contains("meet") || last.contains("coffee") || last.contains("lunch") {
                candidates += ["Sounds good", "What time?", "Can't make it"]
            }
        }
        candidates += ["👍", "On my way", "Got it", "Let me check"]

        let deduped = Array(Set(candidates)).prefix(limit)
        return deduped.enumerated().map { idx, text in
            SmartReplySuggestion(id: "sr-\(idx)-\(text.hashValue)", text: text)
        }
    }

    func summarizeThread(messages: [String], maxSentences: Int = 2) -> ThreadSummary? {
        guard PrivacyAIConsentStore.load().summariesEnabled else { return nil }
        let joined = messages.joined(separator: " ")
        guard joined.count >= 40 else { return nil }

        let tokenizer = NLTokenizer(unit: .sentence)
        tokenizer.string = joined
        var sentences: [String] = []
        tokenizer.enumerateTokens(in: joined.startIndex..<joined.endIndex) { range, _ in
            let sentence = String(joined[range]).trimmingCharacters(in: .whitespacesAndNewlines)
            if !sentence.isEmpty { sentences.append(sentence) }
            return true
        }
        let head = sentences.prefix(maxSentences).joined(separator: " ")
        guard !head.isEmpty else { return nil }
        return ThreadSummary(sentenceCount: sentences.count, summary: head, generatedAt: Date())
    }

    func translateOnDevice(text: String, to languageCode: String) -> String? {
        guard PrivacyAIConsentStore.load().translationEnabled else { return nil }
        guard !text.isEmpty else { return nil }
        // Phase 1: identity passthrough until Apple Translation framework is wired in Xcode.
        _ = languageCode
        return text
    }
}
#endif
