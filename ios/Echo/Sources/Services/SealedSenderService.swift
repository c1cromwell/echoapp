#if os(iOS)
import CryptoKit
import Foundation
import Security

private struct SealedSenderInner: Codable, Sendable {
    let senderDID: String
    let message: TextMessagePayload
    /// The sender's delivery key (hex). It travels only here, inside the encryption, so the
    /// recipient can reply without the server learning who is replying.
    let senderDeliveryKey: String?

    enum CodingKeys: String, CodingKey {
        case senderDID = "sender_did"
        case message
        case senderDeliveryKey = "sender_delivery_key"
    }
}

/// What a sealed message contained once opened on this device.
struct OpenedSealedMessage: Sendable {
    let senderDID: String
    let payload: TextMessagePayload
    let senderDeliveryKey: String?
}

struct SealedDeliveryTokenResponse: Codable, Sendable {
    let deliveryToken: String
    let expiresIn: Int?

    enum CodingKeys: String, CodingKey {
        case deliveryToken = "delivery_token"
        case expiresIn = "expires_in"
    }
}

enum SealedSenderEndpoint: APIEndpoint {
    case deliveryToken

    var path: String { "/v3/messages/sealed-token" }
}

/// Wraps 1:1 chat payloads so relay metadata omits sender DID (WO-219).
actor SealedSenderService {
    private let identityResolve: IdentityResolveClient
    private let apiClient: APIClient

    init(identityResolve: IdentityResolveClient, apiClient: APIClient) {
        self.identityResolve = identityResolve
        self.apiClient = apiClient
    }

    /// `withServerToken` is true only for first contact, when the message has to travel over the
    /// logged-in socket because the recipient's delivery key is not known yet.
    func wrap(
        payload: TextMessagePayload,
        recipientDID: String,
        senderDID: String,
        senderDeliveryKey: String? = nil,
        withServerToken: Bool = true
    ) async throws -> SealedTextPayload {
        let inner = SealedSenderInner(senderDID: senderDID, message: payload, senderDeliveryKey: senderDeliveryKey)
        let innerData = try JSONEncoder().encode(inner)
        guard let innerText = String(data: innerData, encoding: .utf8) else {
            throw SealedSenderError.encodingFailed
        }
        let hex = try await identityResolve.primaryPublicKeyHex(peerDID: recipientDID)
        let pubData = try TextMessageCrypto.dataFromPublicKeyHex(hex)
        let kinnami = KinnamiEncryption()
        let encrypted = try await kinnami.encryptWithKeyAgreement(
            plaintext: innerText,
            recipientPublicKeyData: pubData
        )
        let ciphertext = try JSONEncoder().encode(encrypted)
        let token = withServerToken ? try await fetchDeliveryToken() : nil
        return SealedTextPayload(deliveryToken: token, ciphertext: ciphertext)
    }

    func unwrap(_ sealed: SealedTextPayload) async throws -> (senderDID: String, payload: TextMessagePayload) {
        let opened = try await open(sealed)
        return (opened.senderDID, opened.payload)
    }

    func open(_ sealed: SealedTextPayload) async throws -> OpenedSealedMessage {
        let encrypted = try JSONDecoder().decode(EncryptedMessageWithPublicKey.self, from: sealed.ciphertext)
        let privateKey = try await TextMessageCrypto.loadAgreementPrivateKey()
        let kinnami = KinnamiEncryption()
        let plain = try await kinnami.decryptWithKeyAgreement(
            encryptedMessage: encrypted,
            ourPrivateKey: privateKey
        )
        guard let data = plain.data(using: .utf8) else {
            throw SealedSenderError.decodingFailed
        }
        let inner = try JSONDecoder().decode(SealedSenderInner.self, from: data)
        return OpenedSealedMessage(
            senderDID: inner.senderDID,
            payload: inner.message,
            senderDeliveryKey: inner.senderDeliveryKey
        )
    }

    private func fetchDeliveryToken() async throws -> String {
        struct EmptyBody: Encodable {}
        let response: SealedDeliveryTokenResponse = try await apiClient.post(
            endpoint: SealedSenderEndpoint.deliveryToken,
            body: EmptyBody()
        )
        guard !response.deliveryToken.isEmpty else { throw SealedSenderError.missingToken }
        return response.deliveryToken
    }
}

// MARK: - Delivery keys

/// Where delivery keys are kept. The Keychain in the app; memory in tests.
protocol DeliveryKeyStorage: Sendable {
    func value(for name: String) -> String?
    func set(_ value: String?, for name: String)
}

struct KeychainDeliveryKeyStorage: DeliveryKeyStorage {
    private static let service = "com.echo.delivery"

    private func query(_ name: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: Self.service,
            kSecAttrAccount as String: name,
        ]
    }

    func value(for name: String) -> String? {
        var lookup = query(name)
        lookup[kSecReturnData as String] = true
        var result: AnyObject?
        guard SecItemCopyMatching(lookup as CFDictionary, &result) == errSecSuccess,
              let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    func set(_ value: String?, for name: String) {
        SecItemDelete(query(name) as CFDictionary)
        guard let value else { return }
        var item = query(name)
        item[kSecValueData as String] = Data(value.utf8)
        // Readable after the first unlock so a message arriving while the phone is locked can
        // still be answered; never leaves this device.
        item[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        SecItemAdd(item as CFDictionary, nil)
    }
}

final class MemoryDeliveryKeyStorage: DeliveryKeyStorage, @unchecked Sendable {
    private let lock = NSLock()
    private var values: [String: String] = [:]

    func value(for name: String) -> String? {
        lock.lock(); defer { lock.unlock() }
        return values[name]
    }

    func set(_ value: String?, for name: String) {
        lock.lock(); defer { lock.unlock() }
        values[name] = value
    }
}

/// This device's delivery key and the delivery keys of the people it talks to.
///
/// A delivery key is 32 random bytes. Whoever holds yours can send you a message without the
/// server learning who they are; the server is only ever given its SHA-256. Keys are handed to
/// contacts inside encrypted messages and nowhere else.
final class DeliveryKeys: @unchecked Sendable {
    static let shared = DeliveryKeys(storage: KeychainDeliveryKeyStorage())

    private static let ownName = "own"
    private static let registeredName = "own.registered"
    private let storage: DeliveryKeyStorage
    private let lock = NSLock()
    /// Keys received in a message that has not been authenticated yet.
    private var pending: [String: String] = [:]

    init(storage: DeliveryKeyStorage) {
        self.storage = storage
    }

    static func isValid(_ hex: String) -> Bool {
        hex.count == 64 && hex.allSatisfy { $0.isHexDigit }
    }

    static func hash(ofKeyHex hex: String) -> String? {
        guard isValid(hex), let data = Data(hexEncoded: hex) else { return nil }
        return SHA256.hash(data: data).map { String(format: "%02x", $0) }.joined()
    }

    /// This device's key, created on first use.
    func ownKey() -> String {
        lock.lock(); defer { lock.unlock() }
        if let existing = storage.value(for: Self.ownName), Self.isValid(existing) { return existing }
        return makeOwnKeyLocked()
    }

    /// Replaces this device's key. The caller registers the new one with the server.
    @discardableResult
    func rotateOwnKey() -> String {
        lock.lock(); defer { lock.unlock() }
        return makeOwnKeyLocked()
    }

    private func makeOwnKeyLocked() -> String {
        var bytes = [UInt8](repeating: 0, count: 32)
        let status = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        precondition(status == errSecSuccess, "no secure random bytes")
        let hex = bytes.map { String(format: "%02x", $0) }.joined()
        storage.set(hex, for: Self.ownName)
        storage.set(nil, for: Self.registeredName)
        return hex
    }

    /// True when the server has not yet been told the hash of the current key.
    var needsRegistration: Bool {
        guard let hash = Self.hash(ofKeyHex: ownKey()) else { return false }
        return storage.value(for: Self.registeredName) != hash
    }

    func markRegistered() {
        guard let hash = Self.hash(ofKeyHex: ownKey()) else { return }
        storage.set(hash, for: Self.registeredName)
    }

    private func peerName(_ did: String) -> String { "peer." + did }

    /// The key to send to `peerDID` with, if that peer has given us one.
    func peerKey(for peerDID: String) -> String? {
        guard let key = storage.value(for: peerName(peerDID)), Self.isValid(key) else { return nil }
        return key
    }

    /// A key arrived inside a sealed message claiming to be from `peerDID`. It is not used until
    /// that message has been authenticated: otherwise anyone able to send to us could redirect
    /// our replies to a mailbox of their choosing.
    func notePending(peerDID: String, key: String?) {
        guard let key, Self.isValid(key), !peerDID.isEmpty else { return }
        lock.lock(); defer { lock.unlock() }
        pending[peerDID] = key.lowercased()
    }

    /// A message from `peerDID` was authenticated; start using the key that came with it.
    func confirm(peerDID: String) {
        lock.lock()
        let key = pending.removeValue(forKey: peerDID)
        lock.unlock()
        guard let key else { return }
        storage.set(key, for: peerName(peerDID))
    }

    func forget(peerDID: String) {
        lock.lock()
        pending.removeValue(forKey: peerDID)
        lock.unlock()
        storage.set(nil, for: peerName(peerDID))
    }
}

private extension Data {
    init?(hexEncoded hex: String) {
        guard hex.count % 2 == 0 else { return nil }
        var data = Data(capacity: hex.count / 2)
        var index = hex.startIndex
        while index < hex.endIndex {
            let next = hex.index(index, offsetBy: 2)
            guard let byte = UInt8(hex[index..<next], radix: 16) else { return nil }
            data.append(byte)
            index = next
        }
        self = data
    }
}

// MARK: - Sending without a login

enum SealedDeliveryError: Error, Equatable {
    /// The recipient no longer accepts the key we hold (they replaced it).
    case unknownDeliveryKey
    case rejected(status: Int)
}

/// Sends a sealed message on a connection that carries no login, no cookie and no device header.
/// The server learns which mailbox the message is for and nothing about who sent it.
///
/// The request still comes from this device's network address, as every request does. Hiding
/// that needs a relay or Tor in front of it; the SOCKS proxy setting applies here.
enum SealedDelivery {
    static let path = "/v1/sealed/deliver"

    /// The exact request sent. Separate from `send` so a test can check nothing identifying is in it.
    static func request(
        baseURL: URL,
        deliveryKey: String,
        payload: SealedTextPayload,
        messageId: String?,
        commitmentHash: String?,
        silent: Bool
    ) throws -> URLRequest {
        struct Body: Encodable {
            let delivery_key: String
            let payload: SealedTextPayload
            let silent: Bool?
            let message_id: String?
            let commitment_hash: String?
        }
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpShouldHandleCookies = false
        let anchored = (messageId?.isEmpty == false) && (commitmentHash?.isEmpty == false)
        request.httpBody = try JSONEncoder().encode(Body(
            delivery_key: deliveryKey,
            payload: payload,
            silent: silent ? true : nil,
            message_id: anchored ? messageId : nil,
            commitment_hash: anchored ? commitmentHash : nil
        ))
        return request
    }

    private static func makeSession() -> URLSession {
        // Ephemeral: no cookies, no cache, no credentials shared with the logged-in session.
        let configuration = URLSessionConfiguration.ephemeral
        configuration.httpCookieStorage = nil
        configuration.urlCredentialStorage = nil
        configuration.httpShouldSetCookies = false
        TransportProxySettings.apply(to: configuration)
        return URLSession(configuration: configuration)
    }

    static func send(
        deliveryKey: String,
        payload: SealedTextPayload,
        messageId: String?,
        commitmentHash: String?,
        silent: Bool,
        baseURL: URL = EchoAPIBaseURL.resolved
    ) async throws {
        let request = try request(
            baseURL: baseURL,
            deliveryKey: deliveryKey,
            payload: payload,
            messageId: messageId,
            commitmentHash: commitmentHash,
            silent: silent
        )
        let session = makeSession()
        defer { session.finishTasksAndInvalidate() }
        let (_, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw SealedDeliveryError.rejected(status: 0) }
        switch http.statusCode {
        case 200..<300: return
        case 404: throw SealedDeliveryError.unknownDeliveryKey
        default: throw SealedDeliveryError.rejected(status: http.statusCode)
        }
    }
}

enum DeliveryKeyEndpoint: APIEndpoint {
    case register

    var path: String { "/v3/messages/delivery-key" }
}

/// Tells the server the hash of this device's delivery key. The key itself is never sent to it.
enum DeliveryKeyRegistration {
    private struct Body: Encodable {
        let key_hash: String
        let revoke_others: Bool?
    }

    private struct Reply: Decodable {
        let registered: Bool?
    }

    static func registerIfNeeded(apiClient: APIClient, keys: DeliveryKeys = .shared) async {
        guard keys.needsRegistration else { return }
        await register(apiClient: apiClient, keys: keys, revokeOthers: false)
    }

    /// After blocking someone: a new key, and every earlier key stops working at once, so the
    /// copy the blocked person holds is useless. Other contacts fall back to the logged-in path
    /// for one message and pick up the new key from the next message we send them.
    static func rotateAfterBlocking(apiClient: APIClient, keys: DeliveryKeys = .shared) async {
        keys.rotateOwnKey()
        await register(apiClient: apiClient, keys: keys, revokeOthers: true)
    }

    private static func register(apiClient: APIClient, keys: DeliveryKeys, revokeOthers: Bool) async {
        guard let hash = DeliveryKeys.hash(ofKeyHex: keys.ownKey()) else { return }
        do {
            let _: Reply = try await apiClient.put(
                endpoint: DeliveryKeyEndpoint.register,
                body: Body(key_hash: hash, revoke_others: revokeOthers ? true : nil)
            )
            keys.markRegistered()
        } catch {
            // Left unregistered; the next connect tries again. Until then contacts reach us over
            // the logged-in path.
        }
    }
}

// MARK: - One place that sends a 1:1 message

enum DirectMessageSender {
    /// How a 1:1 message will leave this device.
    enum Route: Equatable {
        /// No login: the server cannot tell who sent it.
        case sealedDelivery(key: String)
        /// First contact: sealed for the recipient, but over the logged-in socket.
        case sealedOverSocket
        /// Sealed sender is switched off for this contact.
        case plain
    }

    static func route(sealedAllowed: Bool, isGroup: Bool, peerKey: String?) -> Route {
        guard sealedAllowed, !isGroup else { return .plain }
        if let peerKey { return .sealedDelivery(key: peerKey) }
        return .sealedOverSocket
    }

    @MainActor
    static func send(
        payload: TextMessagePayload,
        conversationId: String,
        peerDID: String,
        senderDID: String,
        silent: Bool = false,
        signalService: ConversationSignalService,
        keys: DeliveryKeys = .shared
    ) async throws {
        let isGroup = conversationId.hasPrefix("group:")
        let allowed = SealedSenderPolicy.shouldUseSealed(peerDID: peerDID, conversationId: conversationId)
        var chosen = route(sealedAllowed: allowed, isGroup: isGroup, peerKey: keys.peerKey(for: peerDID))
        if chosen == .plain {
            try await signalService.sendTextMessage(
                conversationId: conversationId,
                peerDID: peerDID,
                payload: payload,
                silent: silent
            )
            return
        }

        let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
        let sealedService = SealedSenderService(
            identityResolve: IdentityResolveClient(apiClient: client),
            apiClient: client
        )
        // Our key goes with every sealed message so the other side can always reply sealed,
        // including after we have replaced it.
        let ownKey = keys.ownKey()

        if case .sealedDelivery(let peerKey) = chosen {
            let sealed = try await sealedService.wrap(
                payload: payload,
                recipientDID: peerDID,
                senderDID: senderDID,
                senderDeliveryKey: ownKey,
                withServerToken: false
            )
            do {
                try await SealedDelivery.send(
                    deliveryKey: peerKey,
                    payload: sealed,
                    messageId: payload.messageId,
                    commitmentHash: payload.commitmentHash,
                    silent: silent
                )
                return
            } catch SealedDeliveryError.unknownDeliveryKey {
                // They replaced their key. Forget ours and reach them the first-contact way;
                // their reply brings the new one.
                keys.forget(peerDID: peerDID)
                chosen = .sealedOverSocket
            }
        }

        let sealed = try await sealedService.wrap(
            payload: payload,
            recipientDID: peerDID,
            senderDID: senderDID,
            senderDeliveryKey: ownKey,
            withServerToken: true
        )
        try await signalService.sendSealedTextMessage(
            conversationId: ConversationID.sealedWire,
            peerDID: peerDID,
            payload: sealed,
            silent: silent
        )
    }
}

// MARK: - Receiving

enum SealedInbound {
    /// Turns an opened sealed message into the event the rest of the app handles. The thread is
    /// worked out here, on the device, from the sender named inside the encryption: the wire
    /// names no conversation.
    static func event(
        from opened: OpenedSealedMessage,
        localDID: String,
        fallbackMessageId: String
    ) -> TextMessageSignalEvent? {
        let sender = opened.senderDID
        guard !sender.isEmpty, !localDID.isEmpty, sender != localDID else { return nil }
        let messageId = opened.payload.messageId.isEmpty ? fallbackMessageId : opened.payload.messageId
        return TextMessageSignalEvent(
            conversationId: ConversationID.direct(localDID: localDID, peerDID: sender),
            peerDID: sender,
            messageId: messageId,
            text: TextMessagePayload.encryptedPlaceholder,
            wirePayload: opened.payload
        )
    }

    /// Opens a sealed event. Returns nil when it cannot be opened or is from someone blocked;
    /// such a message cannot be attributed to a thread and is dropped.
    static func open(_ event: TextMessageSignalEvent, keys: DeliveryKeys = .shared) async -> TextMessageSignalEvent? {
        guard let sealed = event.sealedPayload else { return event }
        let client = await DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
        let service = SealedSenderService(
            identityResolve: IdentityResolveClient(apiClient: client),
            apiClient: client
        )
        guard let opened = try? await service.open(sealed) else { return nil }
        if BlockedSenders.contains(opened.senderDID) { return nil }
        let localDID = await CurrentUserSession.currentDID() ?? ""
        guard let resolved = Self.event(from: opened, localDID: localDID, fallbackMessageId: event.messageId) else {
            return nil
        }
        keys.notePending(peerDID: opened.senderDID, key: opened.senderDeliveryKey)
        return resolved
    }
}

/// People this device has blocked. The server cannot enforce a block on a message whose sender
/// it cannot see, so the device does.
enum BlockedSenders {
    private static let key = "echo.blocked_senders"

    static func contains(_ did: String) -> Bool {
        (UserDefaults.standard.stringArray(forKey: key) ?? []).contains(did)
    }

    static func set(_ did: String, blocked: Bool) {
        var all = Set(UserDefaults.standard.stringArray(forKey: key) ?? [])
        if blocked { all.insert(did) } else { all.remove(did) }
        UserDefaults.standard.set(Array(all).sorted(), forKey: key)
    }
}

enum SealedSenderError: LocalizedError {
    case encodingFailed
    case decodingFailed
    case missingToken

    var errorDescription: String? {
        switch self {
        case .encodingFailed: return "Could not encode sealed message."
        case .decodingFailed: return "Could not decode sealed message."
        case .missingToken: return "Sealed sender token unavailable."
        }
    }
}
#endif
