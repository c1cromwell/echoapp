#if os(iOS)
import Foundation

/// Tokenizer for local message search index (WO-3 / M6).
enum MessageSearchTokenizer {
    static let minTokenLength = 3

    /// Lowercase alphanumeric tokens from message body (privacy: local only).
    static func tokenize(_ text: String) -> [String] {
        let cleaned = text.lowercased()
        var tokens: [String] = []
        var current = ""
        for scalar in cleaned.unicodeScalars {
            if CharacterSet.alphanumerics.contains(scalar) {
                current.unicodeScalars.append(scalar)
            } else if !current.isEmpty {
                if current.count >= minTokenLength {
                    tokens.append(current)
                }
                current = ""
            }
        }
        if current.count >= minTokenLength {
            tokens.append(current)
        }
        return tokens
    }

    /// Query terms for AND search; quoted segments become single terms.
    static func parseQuery(_ raw: String) -> [String] {
        var terms: [String] = []
        var buffer = ""
        var inQuotes = false
        for ch in raw {
            if ch == "\"" {
                if inQuotes {
                    let term = buffer.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
                    if term.count >= minTokenLength { terms.append(term) }
                    buffer = ""
                }
                inQuotes.toggle()
                continue
            }
            if inQuotes {
                buffer.append(ch)
                continue
            }
            if ch.isWhitespace {
                let term = buffer.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
                if term.count >= minTokenLength { terms.append(term) }
                buffer = ""
            } else {
                buffer.append(ch)
            }
        }
        let tail = buffer.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if tail.count >= minTokenLength { terms.append(tail) }
        return terms
    }
}
#endif
