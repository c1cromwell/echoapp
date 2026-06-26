import Foundation
import CryptoKit

/// Signal-style Double Ratchet (WO-SX1) — P-256 DH ratchet + symmetric chain ratchet.
enum DoubleRatchet {
    static let algorithm = "DR-P256-AES256GCM"
    static let maxSkip = 1000
    private static let rootInfo = "ECHO-RATCHET-ROOT"
    private static let chainInfo = "ECHO-RATCHET-CHAIN"
    private static let bootstrapInfo = "ECHO-RATCHET-X3DH-v1"
    private static let messageKeyConst: UInt8 = 0x01
    private static let chainKeyConst: UInt8 = 0x02

    struct WireMessage: Codable, Equatable, Sendable {
        let ratchetPublicKey: String
        let pn: UInt32
        let n: UInt32
        let nonce: String
        let ciphertext: String
        let tag: String
        let algorithm: String

        enum CodingKeys: String, CodingKey {
            case ratchetPublicKey
            case pn
            case n
            case nonce, ciphertext, tag, algorithm
        }
    }

    struct SkippedKey: Hashable {
        let ratchetPubB64: String
        let n: UInt32
    }

    /// Mutable session state — serialize via `PersistedState`.
    final class Session {
        var rootKey: Data
        var dhSelf: P256.KeyAgreement.PrivateKey
        var dhRemote: Data?
        var sendChainKey: Data?
        var recvChainKey: Data?
        var sendN: UInt32 = 0
        var recvN: UInt32 = 0
        var prevSendN: UInt32 = 0
        var skipped: [SkippedKey: Data] = [:]
        let maxSkip: Int

        init(
            rootKey: Data,
            dhSelf: P256.KeyAgreement.PrivateKey,
            dhRemote: Data? = nil,
            sendChainKey: Data? = nil,
            recvChainKey: Data? = nil,
            maxSkip: Int = DoubleRatchet.maxSkip
        ) {
            self.rootKey = rootKey
            self.dhSelf = dhSelf
            self.dhRemote = dhRemote
            self.sendChainKey = sendChainKey
            self.recvChainKey = recvChainKey
            self.maxSkip = maxSkip
        }

        func selfRatchetPublicRaw() -> Data {
            let pub = dhSelf.publicKey.rawRepresentation
            if pub.count == 65, pub.first == 0x04 { return pub.dropFirst() }
            return pub
        }
    }

    struct PersistedState: Codable {
        var rootKey: Data
        var dhSelfPriv: Data
        var dhRemote: Data?
        var sendChainKey: Data?
        var recvChainKey: Data?
        var sendN: UInt32
        var recvN: UInt32
        var prevSendN: UInt32
        var skipped: [[String: String]]
    }

    static func deriveBootstrapSecret(
        local: P256.KeyAgreement.PrivateKey,
        peerPubData: Data
    ) throws -> Data {
        let peer = try P256.KeyAgreement.PublicKey(rawRepresentation: normalizedPub(peerPubData))
        let ss = try local.sharedSecretFromKeyAgreement(with: peer)
        let key = ss.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data(),
            sharedInfo: Data(bootstrapInfo.utf8),
            outputByteCount: 32
        )
        return key.withUnsafeBytes { Data($0) }
    }

    static func newInitiator(sharedSecret: Data, remoteRatchetPub: Data) throws -> Session {
        guard sharedSecret.count == 32 else { throw DoubleRatchetError.invalidSecret }
        let dhSelf = P256.KeyAgreement.PrivateKey()
        var s = Session(rootKey: sharedSecret, dhSelf: dhSelf, dhRemote: remoteRatchetPub)
        let dh = try dhRaw(privateKey: dhSelf, remoteRaw: remoteRatchetPub)
        let (rk, ck) = try kdfRootKey(rootKey: s.rootKey, dhOutput: dh)
        s.rootKey = rk
        s.sendChainKey = ck
        return s
    }

    static func newResponder(sharedSecret: Data, selfRatchet: P256.KeyAgreement.PrivateKey) throws -> Session {
        guard sharedSecret.count == 32 else { throw DoubleRatchetError.invalidSecret }
        return Session(rootKey: sharedSecret, dhSelf: selfRatchet)
    }

    static func encrypt(_ session: Session, plaintext: Data) throws -> WireMessage {
        guard var chain = session.sendChainKey else { throw DoubleRatchetError.noSendingChain }
        let (nextChain, mk) = try kdfChainKey(chainKey: chain)
        session.sendChainKey = nextChain
        let enc = try aesGCMEncrypt(plaintext: plaintext, key: SymmetricKey(data: mk))
        let msg = WireMessage(
            ratchetPublicKey: session.selfRatchetPublicRaw().base64EncodedString(),
            pn: session.prevSendN,
            n: session.sendN,
            nonce: enc.nonce.base64EncodedString(),
            ciphertext: enc.ciphertext.base64EncodedString(),
            tag: enc.tag.base64EncodedString(),
            algorithm: algorithm
        )
        session.sendN += 1
        return msg
    }

    static func decrypt(_ session: Session, message: WireMessage) throws -> Data {
        guard let remotePub = Data(base64Encoded: message.ratchetPublicKey) else {
            throw DoubleRatchetError.invalidWire
        }
        if let pt = try consumeSkipped(session: session, remotePub: remotePub, message: message) {
            return pt
        }
        if !pubEqual(remotePub, session.dhRemote) {
            try skipMessageKeys(session: session, from: session.recvN, until: message.pn)
            try dhRatchet(session: session, remotePub: remotePub)
        }
        try skipMessageKeys(session: session, from: session.recvN, until: message.n)
        guard var chain = session.recvChainKey else { throw DoubleRatchetError.noReceivingChain }
        let (nextChain, mk) = try kdfChainKey(chainKey: chain)
        session.recvChainKey = nextChain
        session.recvN += 1
        return try aesGCMDecrypt(
            nonce: message.nonce,
            ciphertext: message.ciphertext,
            tag: message.tag,
            key: SymmetricKey(data: mk)
        )
    }

    static func export(_ session: Session) -> PersistedState {
        let skippedArr = session.skipped.map { key, mk -> [String: String] in
            ["pub": key.ratchetPubB64, "n": String(key.n), "mk": mk.base64EncodedString()]
        }
        return PersistedState(
            rootKey: session.rootKey,
            dhSelfPriv: session.dhSelf.rawRepresentation,
            dhRemote: session.dhRemote,
            sendChainKey: session.sendChainKey,
            recvChainKey: session.recvChainKey,
            sendN: session.sendN,
            recvN: session.recvN,
            prevSendN: session.prevSendN,
            skipped: skippedArr
        )
    }

    static func restore(_ state: PersistedState) throws -> Session {
        let dhSelf = try P256.KeyAgreement.PrivateKey(rawRepresentation: state.dhSelfPriv)
        let s = Session(
            rootKey: state.rootKey,
            dhSelf: dhSelf,
            dhRemote: state.dhRemote,
            sendChainKey: state.sendChainKey,
            recvChainKey: state.recvChainKey
        )
        s.sendN = state.sendN
        s.recvN = state.recvN
        s.prevSendN = state.prevSendN
        for entry in state.skipped {
            guard let pub = entry["pub"], let nStr = entry["n"], let n = UInt32(nStr),
                  let mkB64 = entry["mk"], let mk = Data(base64Encoded: mkB64) else { continue }
            s.skipped[SkippedKey(ratchetPubB64: pub, n: n)] = mk
        }
        return s
    }

    // MARK: - internals

    private static func consumeSkipped(
        session: Session,
        remotePub: Data,
        message: WireMessage
    ) throws -> Data? {
        let key = SkippedKey(ratchetPubB64: message.ratchetPublicKey, n: message.n)
        guard let mk = session.skipped.removeValue(forKey: key) else { return nil }
        return try aesGCMDecrypt(
            nonce: message.nonce,
            ciphertext: message.ciphertext,
            tag: message.tag,
            key: SymmetricKey(data: mk)
        )
    }

    private static func skipMessageKeys(session: Session, from: UInt32, until: UInt32) throws {
        guard session.recvChainKey != nil else { return }
        if until < from { return }
        if until - from > UInt32(session.maxSkip) {
            throw DoubleRatchetError.tooManySkipped
        }
        let remoteB64 = session.dhRemote?.base64EncodedString() ?? ""
        var chain = session.recvChainKey!
        for i in from..<until {
            let (next, mk) = try kdfChainKey(chainKey: chain)
            chain = next
            session.skipped[SkippedKey(ratchetPubB64: remoteB64, n: i)] = mk
        }
        session.recvChainKey = chain
        session.recvN = until
    }

    private static func dhRatchet(session: Session, remotePub: Data) throws {
        session.prevSendN = session.sendN
        session.sendN = 0
        session.recvN = 0
        session.dhRemote = remotePub
        let dh1 = try dhRaw(privateKey: session.dhSelf, remoteRaw: remotePub)
        let (rk1, recvCK) = try kdfRootKey(rootKey: session.rootKey, dhOutput: dh1)
        session.rootKey = rk1
        session.recvChainKey = recvCK
        let newSelf = P256.KeyAgreement.PrivateKey()
        session.dhSelf = newSelf
        let dh2 = try dhRaw(privateKey: newSelf, remoteRaw: remotePub)
        let (rk2, sendCK) = try kdfRootKey(rootKey: session.rootKey, dhOutput: dh2)
        session.rootKey = rk2
        session.sendChainKey = sendCK
    }

    private static func kdfRootKey(rootKey: Data, dhOutput: Data) throws -> (Data, Data) {
        let prk = HKDF<SHA256>.extract(inputKeyMaterial: SymmetricKey(data: dhOutput), salt: rootKey)
        let outKey = HKDF<SHA256>.expand(
            pseudoRandomKey: prk,
            info: Data(rootInfo.utf8),
            outputByteCount: 64
        )
        let out = outKey.withUnsafeBytes { Data($0) }
        return (out.prefix(32), out.suffix(32))
    }

    private static func kdfChainKey(chainKey: Data) throws -> (Data, Data) {
        let mk = HMAC<SHA256>.authenticationCode(for: Data([messageKeyConst]), using: SymmetricKey(data: chainKey))
        let next = HMAC<SHA256>.authenticationCode(for: Data([chainKeyConst]), using: SymmetricKey(data: chainKey))
        let derived = HKDF<SHA256>.deriveKey(
            inputKeyMaterial: SymmetricKey(data: Data(mk)),
            info: Data(chainInfo.utf8),
            outputByteCount: 32
        )
        return (Data(next), derived.withUnsafeBytes { Data($0) })
    }

    private static func dhRaw(privateKey: P256.KeyAgreement.PrivateKey, remoteRaw: Data) throws -> Data {
        let peer = try P256.KeyAgreement.PublicKey(rawRepresentation: normalizedPub(remoteRaw))
        let ss = try privateKey.sharedSecretFromKeyAgreement(with: peer)
        return ss.withUnsafeBytes { Data($0) }
    }

    private static func normalizedPub(_ raw: Data) -> Data {
        if raw.count == 64 { return Data([0x04]) + raw }
        return raw
    }

    private static func pubEqual(_ a: Data?, _ b: Data?) -> Bool {
        guard let a, let b else { return false }
        return a == b
    }

    private static func aesGCMEncrypt(plaintext: Data, key: SymmetricKey) throws -> (nonce: Data, ciphertext: Data, tag: Data) {
        let sealed = try AES.GCM.seal(plaintext, using: key)
        return (Data(sealed.nonce), sealed.ciphertext, sealed.tag)
    }

    private static func aesGCMDecrypt(nonce: String, ciphertext: String, tag: String, key: SymmetricKey) throws -> Data {
        guard let n = Data(base64Encoded: nonce),
              let ct = Data(base64Encoded: ciphertext),
              let tg = Data(base64Encoded: tag) else {
            throw DoubleRatchetError.invalidWire
        }
        let box = try AES.GCM.SealedBox(nonce: try AES.GCM.Nonce(data: n), ciphertext: ct, tag: tg)
        return try AES.GCM.open(box, using: key)
    }
}

enum DoubleRatchetError: LocalizedError {
    case invalidSecret
    case invalidWire
    case noSendingChain
    case noReceivingChain
    case tooManySkipped

    var errorDescription: String? {
        switch self {
        case .invalidSecret: return "Invalid ratchet bootstrap secret."
        case .invalidWire: return "Invalid ratchet wire message."
        case .noSendingChain: return "Ratchet has no sending chain yet."
        case .noReceivingChain: return "Ratchet has no receiving chain yet."
        case .tooManySkipped: return "Too many skipped ratchet messages."
        }
    }
}
