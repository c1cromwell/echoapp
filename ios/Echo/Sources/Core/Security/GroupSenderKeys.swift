import Foundation
import CryptoKit

/// Echo Sender Keys (Signal Parity Wave S2) — per-sender chain keys over the group root key.
/// Scales group send beyond sealing a fresh package for every member on every message:
/// each sender advances a chain key; receivers derive message keys from distribution.
///
/// Does **not** replace `GroupKeyManager` membership rekey; it layers on top of the
/// current group symmetric root (versioned) for message encryption efficiency.
actor GroupSenderKeyStore {
    struct ChainState: Sendable {
        let groupId: String
        let senderDID: String
        let groupKeyVersion: Int
        var chainKey: SymmetricKey
        var iteration: UInt32
    }

    private var chains: [String: ChainState] = [:]

    private static func chainId(groupId: String, senderDID: String, version: Int) -> String {
        "\(groupId)|\(senderDID)|\(version)"
    }

    /// Seed a sender chain from the group root key (admin distribute or self after generate).
    func seedChain(groupId: String, senderDID: String, groupKeyVersion: Int, rootKey: SymmetricKey) {
        let id = Self.chainId(groupId: groupId, senderDID: senderDID, version: groupKeyVersion)
        let salt = Data("echo.senderkey.v1.\(groupId).\(senderDID)".utf8)
        let derived = HKDF<SHA256>.deriveKey(
            inputKeyMaterial: rootKey,
            salt: salt,
            info: Data("chain".utf8),
            outputByteCount: 32
        )
        chains[id] = ChainState(
            groupId: groupId,
            senderDID: senderDID,
            groupKeyVersion: groupKeyVersion,
            chainKey: derived,
            iteration: 0
        )
    }

    /// Encrypt plaintext for the group using the local sender chain; advances chain.
    func encrypt(
        plaintext: Data,
        groupId: String,
        senderDID: String,
        groupKeyVersion: Int
    ) throws -> (ciphertext: Data, iteration: UInt32) {
        let id = Self.chainId(groupId: groupId, senderDID: senderDID, version: groupKeyVersion)
        guard var state = chains[id] else { throw GroupSenderKeyError.missingChain }
        let messageKey = Self.deriveMessageKey(chainKey: state.chainKey, iteration: state.iteration)
        let nonce = AES.GCM.Nonce()
        let sealed = try AES.GCM.seal(plaintext, using: messageKey, nonce: nonce)
        guard let combined = sealed.combined else { throw GroupSenderKeyError.encryptionFailed }
        let iter = state.iteration
        state.chainKey = Self.advanceChain(state.chainKey)
        state.iteration += 1
        chains[id] = state
        return (combined, iter)
    }

    /// Decrypt using peer sender chain (must be seeded from same root).
    func decrypt(
        ciphertext: Data,
        groupId: String,
        senderDID: String,
        groupKeyVersion: Int,
        iteration: UInt32
    ) throws -> Data {
        let id = Self.chainId(groupId: groupId, senderDID: senderDID, version: groupKeyVersion)
        guard var state = chains[id] else { throw GroupSenderKeyError.missingChain }
        // Fast-forward chain if needed (out-of-order limited; advance until iteration).
        while state.iteration < iteration {
            state.chainKey = Self.advanceChain(state.chainKey)
            state.iteration += 1
        }
        guard state.iteration == iteration else { throw GroupSenderKeyError.iterationMismatch }
        let messageKey = Self.deriveMessageKey(chainKey: state.chainKey, iteration: iteration)
        let box = try AES.GCM.SealedBox(combined: ciphertext)
        let plain = try AES.GCM.open(box, using: messageKey)
        state.chainKey = Self.advanceChain(state.chainKey)
        state.iteration += 1
        chains[id] = state
        return plain
    }

    func hasChain(groupId: String, senderDID: String, groupKeyVersion: Int) -> Bool {
        chains[Self.chainId(groupId: groupId, senderDID: senderDID, version: groupKeyVersion)] != nil
    }

    private static func deriveMessageKey(chainKey: SymmetricKey, iteration: UInt32) -> SymmetricKey {
        var info = Data("msg".utf8)
        var be = iteration.bigEndian
        withUnsafeBytes(of: &be) { info.append(contentsOf: $0) }
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: chainKey,
            salt: Data(),
            info: info,
            outputByteCount: 32
        )
    }

    private static func advanceChain(_ key: SymmetricKey) -> SymmetricKey {
        HKDF<SHA256>.deriveKey(
            inputKeyMaterial: key,
            salt: Data("echo.senderkey.advance".utf8),
            info: Data("next".utf8),
            outputByteCount: 32
        )
    }
}

enum GroupSenderKeyError: LocalizedError {
    case missingChain
    case encryptionFailed
    case iterationMismatch

    var errorDescription: String? {
        switch self {
        case .missingChain: return "Group sender chain not established — wait for key distribution."
        case .encryptionFailed: return "Sender-key encryption failed."
        case .iterationMismatch: return "Sender-key iteration out of sync."
        }
    }
}
