// Features/Safety/LookalikeDetector.swift
//
// Detects impersonation by name: folds diacritics + common Unicode homoglyphs (Cyrillic/Greek →
// Latin), strips whitespace/case, then compares with a length-scaled edit distance. Pure logic.

import Foundation

public enum LookalikeDetector {
    /// Common homoglyphs that render like Latin letters but are different code points.
    private static let confusables: [Character: Character] = [
        "а": "a", "е": "e", "о": "o", "р": "p", "с": "c", "х": "x", "у": "y", "к": "k",
        "т": "t", "в": "b", "н": "h", "м": "m", "і": "i", "ѕ": "s", "ј": "j", "ԁ": "d",
        "Α": "A", "Β": "B", "Ε": "E", "Ζ": "Z", "Η": "H", "Ι": "I", "Κ": "K", "Μ": "M",
        "Ν": "N", "Ο": "O", "Ρ": "P", "Τ": "T", "Υ": "Y", "Χ": "X",
        "α": "a", "ο": "o", "ν": "v", "ι": "i", "ρ": "p", "ε": "e", "τ": "t",
    ]

    /// Canonical form for confusable comparison.
    public static func normalize(_ name: String) -> String {
        let folded = name.folding(options: [.diacriticInsensitive, .widthInsensitive], locale: .current)
        let mapped = String(folded.map { confusables[$0] ?? $0 })
        return mapped.lowercased().filter { !$0.isWhitespace }
    }

    public static func levenshtein(_ s1: String, _ s2: String) -> Int {
        let a = Array(s1), b = Array(s2)
        if a.isEmpty { return b.count }
        if b.isEmpty { return a.count }
        var prev = Array(0...b.count)
        for i in 1...a.count {
            var cur = [i] + Array(repeating: 0, count: b.count)
            for j in 1...b.count {
                cur[j] = a[i - 1] == b[j - 1] ? prev[j - 1] : Swift.min(prev[j - 1], prev[j], cur[j - 1]) + 1
            }
            prev = cur
        }
        return prev[b.count]
    }

    /// True if `candidate` looks confusingly like `reference` after folding. Identical-after-folding
    /// (pure homoglyph swap) always matches; otherwise allow a small, length-scaled edit distance.
    public static func isLookalike(_ candidate: String, of reference: String, maxDistance: Int = 1) -> Bool {
        let c = normalize(candidate)
        let r = normalize(reference)
        guard !c.isEmpty, !r.isEmpty else { return false }
        if c == r { return true }
        let allowed = Swift.max(maxDistance, r.count / 6)
        return levenshtein(c, r) <= allowed
    }
}
