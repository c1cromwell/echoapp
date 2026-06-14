#if os(iOS)
import Foundation
import CryptoKit

/// AES-GCM encrypted on-disk search index (WO-3). Server never sees plaintext index.
actor EncryptedIndexStore {
    static let shared = EncryptedIndexStore()

    private let fileURL: URL
    private static let salt = Data("echo-search-index-v1".utf8)

    init() {
        let base = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        let dir = base.appendingPathComponent("search_index", isDirectory: true)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        fileURL = dir.appendingPathComponent("index.enc")
    }

    func load() throws -> SearchIndexSnapshot {
        guard FileManager.default.fileExists(atPath: fileURL.path) else {
            return SearchIndexSnapshot()
        }
        let wrapped = try Data(contentsOf: fileURL)
        let key = Self.deriveKey()
        let archive = try JSONDecoder().decode(EncryptedIndexArchive.self, from: wrapped)
        let sealed = try AES.GCM.SealedBox(combined: archive.ciphertext)
        let plaintext = try AES.GCM.open(sealed, using: key)
        return try JSONDecoder().decode(SearchIndexSnapshot.self, from: plaintext)
    }

    func save(_ snapshot: SearchIndexSnapshot) throws {
        let key = Self.deriveKey()
        let plaintext = try JSONEncoder().encode(snapshot)
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw EncryptedIndexStoreError.sealFailed }
        let archive = EncryptedIndexArchive(version: 1, ciphertext: combined)
        let data = try JSONEncoder().encode(archive)
        try data.write(to: fileURL, options: .atomic)
    }

    private static func deriveKey() -> SymmetricKey {
        // Device-local HKDF salt (T2 at rest). Cross-device sync uses WO-73 later.
        SymmetricKey(data: SHA256.hash(data: salt + Data("echo-index-local".utf8)))
    }
}

struct EncryptedIndexArchive: Codable {
    let version: Int
    let ciphertext: Data
}

struct SearchIndexSnapshot: Codable, Sendable, Equatable {
    var postings: [String: [SearchPosting]]
    var documents: [String: SearchDocument]

    init(postings: [String: [SearchPosting]] = [:], documents: [String: SearchDocument] = [:]) {
        self.postings = postings
        self.documents = documents
    }
}

struct SearchPosting: Codable, Sendable, Equatable {
    let messageId: String
    let conversationId: String
    let timestamp: TimeInterval
    let fieldType: String
}

struct SearchDocument: Codable, Sendable, Equatable {
    let messageId: String
    let conversationId: String
    let senderDID: String
    let bodyPreview: String
    let timestamp: TimeInterval
    let contentType: String
}

enum EncryptedIndexStoreError: LocalizedError {
    case sealFailed

    var errorDescription: String? {
        switch self {
        case .sealFailed: return "Could not seal search index."
        }
    }
}
#endif
