// Features/Onboarding/Recovery/Models/RecoveryPhrase.swift

import Foundation
import CryptoKit

/// A 24-word BIP-39 recovery phrase. Never held in memory longer than required —
/// clear references as soon as derivation or display is complete.
struct RecoveryPhrase: Sendable {
    let words: [String] // always 24, lowercase

    /// Fails if count ≠ 24 or any word is not in the BIP-39 English wordlist.
    init?(words: [String]) {
        guard words.count == 24 else { return nil }
        let normalized = words.map { $0.lowercased().trimmingCharacters(in: .whitespaces) }
        guard normalized.allSatisfy({ BIP39Wordlist.contains($0) }) else { return nil }
        self.words = normalized
    }

    /// Returns the underlying 256-bit entropy if the phrase passes BIP-39 checksum validation.
    /// Returns nil on checksum failure — callers must surface the error to the user.
    var entropy: Data? {
        BIP39.toEntropy(words: words)
    }

    /// Three random word positions (1-indexed, sorted) for the confirmation challenge.
    func challengePositions() -> [Int] {
        Array(Array(1...24).shuffled().prefix(3).sorted())
    }

    func word(at position: Int) -> String {
        words[position - 1]
    }
}

// MARK: - BIP-39 Entropy Derivation

/// Minimal BIP-39 implementation for entropy extraction and checksum validation.
/// Full mnemonic-to-seed derivation for wallet restoration is delegated to StargazerBridge.
enum BIP39 {
    /// Converts a 24-word mnemonic to its 256-bit entropy, validating the embedded checksum.
    /// Returns nil if the mnemonic is invalid or the checksum does not match.
    static func toEntropy(words: [String]) -> Data? {
        guard words.count == 24 else { return nil }

        // Map each word to its 11-bit index in the BIP-39 wordlist.
        let sortedWords = BIP39Wordlist.words.sorted()
        var bits: [Bool] = []
        bits.reserveCapacity(264) // 24 × 11

        for word in words {
            guard let idx = sortedWords.firstIndex(of: word) else { return nil }
            for shift in stride(from: 10, through: 0, by: -1) {
                bits.append((idx >> shift) & 1 == 1)
            }
        }

        guard bits.count == 264 else { return nil }

        // Split: first 256 bits = entropy, last 8 bits = checksum.
        let entropyBits = Array(bits.prefix(256))
        let checksumBits = Array(bits.suffix(8))

        // Pack entropy bits into bytes.
        var entropyBytes = Data(count: 32)
        for i in 0..<32 {
            var byte: UInt8 = 0
            for j in 0..<8 {
                if entropyBits[i * 8 + j] { byte |= 1 << (7 - j) }
            }
            entropyBytes[i] = byte
        }

        // Verify checksum: first 8 bits of SHA-256(entropy).
        let digest = SHA256.hash(data: entropyBytes)
        let digestBytes = Array(digest)
        let firstByte = digestBytes[0]
        let expectedChecksumBits = (0..<8).map { (firstByte >> (7 - $0)) & 1 == 1 }

        guard checksumBits == expectedChecksumBits else { return nil }
        return entropyBytes
    }
}
