#if os(iOS)
import Foundation
import CryptoKit

/// Per-peer Double Ratchet session lifecycle (WO-SX1).
actor DoubleRatchetCoordinator {
    static let shared = DoubleRatchetCoordinator()

    private var sessions: [String: DoubleRatchet.Session] = [:]
    private var peerPreKeys: [String: Data] = [:]
    private let store = DoubleRatchetStore()

    func cachePeerPreKey(peerDID: String, ratchetPubRaw: Data) {
        peerPreKeys[peerDID] = ratchetPubRaw
    }

    func publishedPreKeyRaw() async throws -> Data {
        if let existing = try await store.loadLocalPreKeyPublic() {
            return existing
        }
        let key = P256.KeyAgreement.PrivateKey()
        let pub = key.publicKey.rawRepresentation
        let raw = pub.count == 65 && pub.first == 0x04 ? Data(pub.dropFirst()) : pub
        try await store.saveLocalPreKey(privateKey: key, publicRaw: raw)
        return raw
    }

    func encrypt(plaintext: String, peerDID: String, peerMessagingPub: Data) async throws -> DoubleRatchet.WireMessage {
        let session = try await sessionForSend(peerDID: peerDID, peerMessagingPub: peerMessagingPub)
        guard let data = plaintext.data(using: .utf8) else { throw DoubleRatchetError.invalidWire }
        let wire = try DoubleRatchet.encrypt(session, plaintext: data)
        try await store.saveSession(peerDID: peerDID, state: DoubleRatchet.export(session))
        return wire
    }

    func decrypt(_ wire: DoubleRatchet.WireMessage, peerDID: String, peerMessagingPub: Data) async throws -> String {
        let session = try await sessionForReceive(peerDID: peerDID, peerMessagingPub: peerMessagingPub, remoteRatchetPub: Data(base64Encoded: wire.ratchetPublicKey))
        let pt = try DoubleRatchet.decrypt(session, message: wire)
        try await store.saveSession(peerDID: peerDID, state: DoubleRatchet.export(session))
        guard let text = String(data: pt, encoding: .utf8) else { throw DoubleRatchetError.invalidWire }
        return text
    }

    private func sessionForSend(peerDID: String, peerMessagingPub: Data) async throws -> DoubleRatchet.Session {
        if let cached = sessions[peerDID] { return cached }
        if let persisted = try await store.loadSession(peerDID: peerDID),
           let restored = try? DoubleRatchet.restore(persisted) {
            sessions[peerDID] = restored
            return restored
        }
        let local = try await TextMessageCrypto.loadAgreementPrivateKey()
        let secret = try DoubleRatchet.deriveBootstrapSecret(local: local, peerPubData: peerMessagingPub)
        let selfDID = await CurrentUserSession.currentDID() ?? ""
        let isInitiator = selfDID > peerDID
        if isInitiator {
            guard let remotePre = peerPreKeys[peerDID] else {
                throw DoubleRatchetError.noSendingChain
            }
            let sess = try DoubleRatchet.newInitiator(sharedSecret: secret, remoteRatchetPub: remotePre)
            sessions[peerDID] = sess
            return sess
        }
        let pre = try await store.loadLocalPreKeyPrivate() ?? P256.KeyAgreement.PrivateKey()
        let sess = try DoubleRatchet.newResponder(sharedSecret: secret, selfRatchet: pre)
        sessions[peerDID] = sess
        return sess
    }

    private func sessionForReceive(
        peerDID: String,
        peerMessagingPub: Data,
        remoteRatchetPub: Data?
    ) async throws -> DoubleRatchet.Session {
        if let cached = sessions[peerDID] { return cached }
        if let persisted = try await store.loadSession(peerDID: peerDID),
           let restored = try? DoubleRatchet.restore(persisted) {
            sessions[peerDID] = restored
            return restored
        }
        let local = try await TextMessageCrypto.loadAgreementPrivateKey()
        let secret = try DoubleRatchet.deriveBootstrapSecret(local: local, peerPubData: peerMessagingPub)
        let selfDID = await CurrentUserSession.currentDID() ?? ""
        let isInitiator = selfDID > peerDID
        if isInitiator {
            // Initiator receiving first reply — should already have session; bootstrap responder side.
            let pre = try await store.loadLocalPreKeyPrivate() ?? P256.KeyAgreement.PrivateKey()
            let sess = try DoubleRatchet.newResponder(sharedSecret: secret, selfRatchet: pre)
            sessions[peerDID] = sess
            return sess
        }
        guard let remotePre = remoteRatchetPub ?? peerPreKeys[peerDID] else {
            throw DoubleRatchetError.noReceivingChain
        }
        let sess = try DoubleRatchet.newInitiator(sharedSecret: secret, remoteRatchetPub: remotePre)
        sessions[peerDID] = sess
        return sess
    }
}

/// Keychain-backed ratchet state (T1 — device-local only).
struct DoubleRatchetStore {
    private let sessionPrefix = "echo.ratchet.session."
    private let localPrePriv = "echo.ratchet.local.prekey.priv"
    private let localPrePub = "echo.ratchet.local.prekey.pub"

    func saveSession(peerDID: String, state: DoubleRatchet.PersistedState) async throws {
        let data = try JSONEncoder().encode(state)
        UserDefaults.standard.set(data, forKey: sessionPrefix + peerDID)
    }

    func loadSession(peerDID: String) async throws -> DoubleRatchet.PersistedState? {
        guard let data = UserDefaults.standard.data(forKey: sessionPrefix + peerDID) else { return nil }
        return try JSONDecoder().decode(DoubleRatchet.PersistedState.self, from: data)
    }

    func saveLocalPreKey(privateKey: P256.KeyAgreement.PrivateKey, publicRaw: Data) async throws {
        UserDefaults.standard.set(privateKey.rawRepresentation, forKey: localPrePriv)
        UserDefaults.standard.set(publicRaw, forKey: localPrePub)
    }

    func loadLocalPreKeyPublic() async -> Data? {
        UserDefaults.standard.data(forKey: localPrePub)
    }

    func loadLocalPreKeyPrivate() async throws -> P256.KeyAgreement.PrivateKey? {
        guard let raw = UserDefaults.standard.data(forKey: localPrePriv) else { return nil }
        return try P256.KeyAgreement.PrivateKey(rawRepresentation: raw)
    }
}
#endif
