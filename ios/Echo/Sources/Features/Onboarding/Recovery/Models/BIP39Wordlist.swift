// Features/Onboarding/Recovery/Models/BIP39Wordlist.swift
// BIP-39 English wordlist (2048 words) loaded from the app bundle.
// Resource file: Echo/Resources/bip39_english.txt (one word per line, UTF-8).
// Source: https://github.com/bitcoin/bips/blob/master/bip-0039/english.txt

import Foundation

enum BIP39Wordlist {
    // Loaded once at first use and cached.
    static let words: Set<String> = {
        guard let url = Bundle.main.url(forResource: "bip39_english", withExtension: "txt"),
              let contents = try? String(contentsOf: url, encoding: .utf8)
        else {
            assertionFailure("bip39_english.txt missing from app bundle — add it to Echo/Resources/")
            return []
        }
        let loaded = Set(
            contents
                .components(separatedBy: .newlines)
                .map { $0.trimmingCharacters(in: .whitespaces).lowercased() }
                .filter { !$0.isEmpty }
        )
        assert(loaded.count == 2048, "BIP-39 wordlist should have exactly 2048 words, got \(loaded.count)")
        return loaded
    }()

    static func contains(_ word: String) -> Bool {
        words.contains(word.lowercased().trimmingCharacters(in: .whitespaces))
    }
}
