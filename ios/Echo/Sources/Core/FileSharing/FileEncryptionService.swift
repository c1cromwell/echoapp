#if os(iOS)
import Foundation
import CryptoKit

// WO-9: File Encryption and Key Management
//
// # T0–T7 Data Classification
//
// | Data                  | Tier | Rationale |
// |-----------------------|------|-----------|
// | `fileData` (input)    | T0   | Memory-only; never logged |
// | `conversationKey`     | T1   | Derived in Secure Enclave context |
// | `fileKey` (derived)   | T1   | Per-file HKDF key; used once, then discarded |
// | `EncryptedFile`       | T2   | AES-256-GCM ciphertext; safe to store/send |
//
// T0 invariant: plaintext file bytes must never appear in logs, DB writes,
// or network payloads.  Only the T2 `EncryptedFile` blob leaves this service.

/// Per-file encryption using AES-256-GCM with HKDF-derived keys.
///
/// Each file gets a unique symmetric key derived from the conversation's
/// Kinnami key and the file's UUID:
/// ```
/// fileKey = HKDF-SHA256(ikm: conversationKey, info: "file_key_<uuid>")
/// ```
/// This means rotating the conversation key does not decrypt existing files,
/// and compromising one file key does not compromise others.
struct FileEncryptionService {

    // MARK: - Encrypted model

    struct EncryptedFile: Codable {
        /// AES.GCM.SealedBox combined representation (nonce || ciphertext || tag).
        let combined: Data
        /// UUID of the file — needed to re-derive the key during decryption.
        let fileId: UUID
        /// SHA-256 of the plaintext for integrity verification.
        let plaintextHash: String
        /// MIME type of the original file (encrypted separately if sensitive).
        let mimeType: String
        let encryptedAt: Date
    }

    // MARK: - Encryption

    /// Encrypts raw `Data` with a per-file key derived from `conversationKey`.
    ///
    /// - Parameters:
    ///   - data: Plaintext file bytes (T0 — discarded after this call).
    ///   - conversationKey: The conversation's Kinnami symmetric key (T1).
    ///   - fileId: UUID identifying this file (used as HKDF info).
    ///   - mimeType: MIME type of the file (stored in the encrypted envelope).
    /// - Returns: `EncryptedFile` (T2) ready for local storage or relay.
    func encryptData(
        _ data: Data,
        conversationKey: SymmetricKey,
        fileId: UUID,
        mimeType: String = "application/octet-stream"
    ) throws -> EncryptedFile {
        let fileKey = deriveFileKey(conversationKey: conversationKey, fileId: fileId)
        let sealedBox = try AES.GCM.seal(data, using: fileKey)
        guard let combined = sealedBox.combined else {
            throw FileEncryptionError.sealFailed
        }
        let hash = SHA256.hash(data: data)
        return EncryptedFile(
            combined: combined,
            fileId: fileId,
            plaintextHash: hash.compactMap { String(format: "%02x", $0) }.joined(),
            mimeType: mimeType,
            encryptedAt: Date()
        )
    }

    /// Encrypts a file at the given URL.
    func encryptFile(
        at url: URL,
        conversationKey: SymmetricKey,
        fileId: UUID
    ) async throws -> EncryptedFile {
        let data = try Data(contentsOf: url)
        let mimeType = url.mimeType
        return try encryptData(data, conversationKey: conversationKey, fileId: fileId, mimeType: mimeType)
    }

    // MARK: - Decryption

    /// Decrypts an `EncryptedFile` and verifies its plaintext integrity hash.
    func decryptFile(
        _ encrypted: EncryptedFile,
        conversationKey: SymmetricKey
    ) throws -> Data {
        let fileKey = deriveFileKey(conversationKey: conversationKey, fileId: encrypted.fileId)
        let sealedBox = try AES.GCM.SealedBox(combined: encrypted.combined)
        let plaintext = try AES.GCM.open(sealedBox, using: fileKey)

        let hash = SHA256.hash(data: plaintext)
            .compactMap { String(format: "%02x", $0) }.joined()
        guard hash == encrypted.plaintextHash else {
            throw FileEncryptionError.integrityCheckFailed
        }
        return plaintext
    }

    // MARK: - Key derivation

    /// Derives a 256-bit per-file key using HKDF-SHA256.
    /// `info = "file_key_<uuid>"` ensures key isolation across files.
    private func deriveFileKey(conversationKey: SymmetricKey, fileId: UUID) -> SymmetricKey {
        let info = Data("file_key_\(fileId.uuidString)".utf8)
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: conversationKey,
            info: info,
            outputByteCount: 32
        )
    }
}

// MARK: - Errors

enum FileEncryptionError: LocalizedError {
    case sealFailed
    case integrityCheckFailed
    case unsupportedFormat

    var errorDescription: String? {
        switch self {
        case .sealFailed:           return "AES-GCM seal produced no combined output"
        case .integrityCheckFailed: return "Decrypted file hash does not match; possible tampering"
        case .unsupportedFormat:    return "Unsupported file format for encryption"
        }
    }
}

// MARK: - URL helpers

private extension URL {
    var mimeType: String {
        switch pathExtension.lowercased() {
        case "jpg", "jpeg": return "image/jpeg"
        case "png":         return "image/png"
        case "gif":         return "image/gif"
        case "mp4":         return "video/mp4"
        case "mov":         return "video/quicktime"
        case "pdf":         return "application/pdf"
        case "txt":         return "text/plain"
        default:            return "application/octet-stream"
        }
    }
}
#endif
