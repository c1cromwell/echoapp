#if os(iOS)
import Foundation

/// Shared WS reconnect used by the Messages tab and by APNs wake-up.
/// Opening the socket triggers server `flushOffline`, so queued messages land
/// even if the recipient was not in the app when they were sent.
enum MessageRelaySession {
    private static let inboundQueue = SerialAsyncQueue()

    @MainActor
    static func connect() async {
        guard let service = DIContainer.shared.resolveConversationSignalService() else { return }
        // A stored token may be missing (new install) or expired (they last 15
        // minutes); the bootstrap renews it or opens a session with the passkey.
        guard let token = await SessionTokenBootstrap.shared.validAccessToken(), !token.isEmpty else { return }

        service.setInboundTextHandler { event in
            // In arrival order: each message is decrypted and stored before the next starts.
            inboundQueue.enqueue {
                await ingestInbound(event)
            }
        }

        try? await service.connect(accessToken: token)
        // Publish this account's encryption key before anything else: until it is on the server,
        // a message from a new contact is encrypted to a key this device cannot decrypt with.
        if let did = await CurrentUserSession.currentDID() {
            await MessagingKeyRegistrar().ensureRegistered(did: did)
        }
        if let client = DIContainer.shared.resolveAPIClient() {
            await DeliveryKeyRegistration.registerIfNeeded(apiClient: client)
            await RecoveryLinkSync.syncIfNeeded(apiClient: client)
            await UsernameRegistration.syncIfNeeded(apiClient: client)
        }
        await IncomingCallPresenter.shared.configureIfNeeded()
        ScreenshotAlertService.shared.configure(signalService: service)
        ScreenshotAlertService.shared.startMonitoring()
        await PushRegistrationService.shared.registerIfNeeded()
    }

    @MainActor
    private static func ingestInbound(_ event: TextMessageSignalEvent) async {
        let resolved = await InboundTextMessageResolver.resolveBody(for: event)
        let store = ConversationStore.shared
        let match = store.conversations.first(where: { $0.id == event.conversationId })
            ?? store.conversations.first(where: { $0.peerDID == event.peerDID })
        let conversation: StoredConversation
        if let match {
            conversation = match
        } else if let created = await ContactThreadHelper.upsertDirectThread(
            peerDID: event.peerDID,
            displayName: event.peerDID
        ) {
            conversation = created
        } else {
            return
        }

        let currentDID = await CurrentUserSession.currentDID() ?? ""
        guard !resolved.senderDID.isEmpty, resolved.senderDID != currentDID else { return }

        let inbound = ChatDetailMessage(
            id: event.messageId,
            senderDID: resolved.senderDID.isEmpty ? event.peerDID : resolved.senderDID,
            currentUserDID: currentDID,
            content: resolved.body,
            timestamp: "Now",
            deliveryStatus: .delivered
        )
        ConversationThreadStore.appendIfNew(conversationId: conversation.id, message: inbound)
        let session = HiddenChatsSession.shared
        guard session.shouldSurfaceNotification(for: conversation.id) else { return }
        let preview = session.redactedPreviewIfNeeded(
            for: conversation.id,
            resolved: resolved.preview
        )
        store.appendMessagePreview(conversationId: conversation.id, preview: preview)
        if ActiveChatRegistry.openConversationId != conversation.id {
            store.incrementUnread(conversationId: conversation.id)
        }
    }
}

// MARK: - Username

/// Claims this account's username on the server, so other people can find it.
///
/// First run checks that the chosen name is free but, until this existed, never told the server
/// the account had taken it: nobody could be found by username. This sends it once per identity,
/// the first time the app is signed in. It never blocks sign-in; a failure is retried at the
/// next connect.
enum UsernameRegistration {
    private static let registeredKey = "echo.username.registered"
    /// Set when the server says the name now belongs to someone else (taken between the
    /// availability check and registration). The account then has no findable name until the
    /// person picks another.
    static let conflictKey = "echo.username.conflict"

    enum Outcome: Equatable {
        case nothingToRegister, alreadyDone, registered, taken, failed
    }

    struct Endpoint: APIEndpoint {
        var path: String { "/v3/auth/register" }
    }

    private struct Body: Encodable {
        let username: String
        let publicKey: String
    }

    private struct Reply: Decodable {
        let did: String?
        let username: String?
    }

    /// One entry per identity and name: a new account, or a changed name, registers again.
    static func marker(did: String, username: String) -> String {
        did + "|" + username.lowercased()
    }

    static func needsSync(did: String?, username: String?, lastRegistered: String?) -> Bool {
        guard let did, !did.isEmpty, let username, isRegistrable(username) else { return false }
        return marker(did: did, username: username) != lastRegistered
    }

    /// The server's rule: 3-30 letters, numbers or underscores. A placeholder such as "You" for
    /// an account with no name yet must not be claimed.
    static func isRegistrable(_ username: String) -> Bool {
        guard (3...30).contains(username.count), username != "You" else { return false }
        return username.allSatisfy { $0.isASCII && ($0.isLetter || $0.isNumber || $0 == "_") }
    }

    private static func storedUsername() async -> String? {
        if let name = try? await KeychainManager.shared.retrieve(key: "echo.username.current", as: String.self), !name.isEmpty {
            return name
        }
        return UserDefaults.standard.string(forKey: "echo.displayName")
    }

    @discardableResult
    static func syncIfNeeded(apiClient: APIClient, defaults: UserDefaults = .standard) async -> Outcome {
        let did = await CurrentUserSession.currentDID()
        let username = await storedUsername()
        guard let did, !did.isEmpty, let username, isRegistrable(username) else { return .nothingToRegister }
        guard needsSync(did: did, username: username, lastRegistered: defaults.string(forKey: registeredKey)) else {
            return .alreadyDone
        }
        guard let key = try? await SecureEnclaveManager.shared.getPublicKey(id: "echo-identity-signing") else {
            return .failed
        }
        let publicKeyHex = key.map { String(format: "%02x", $0) }.joined()
        do {
            let _: Reply = try await apiClient.post(endpoint: Endpoint(), body: Body(username: username, publicKey: publicKeyHex))
            defaults.set(marker(did: did, username: username), forKey: registeredKey)
            defaults.removeObject(forKey: conflictKey)
            return .registered
        } catch APIError.unexpectedStatusCode(409) {
            // The name is someone else's now, or this account already holds a different one.
            // Asking again will not change the answer.
            defaults.set(marker(did: did, username: username), forKey: registeredKey)
            defaults.set(username, forKey: conflictKey)
            return .taken
        } catch {
            return .failed
        }
    }
}

// MARK: - Recovery link

/// Tells the server which recovery address belongs to this account.
///
/// Restoring an account on a new device starts from the recovery phrase: the app derives the
/// recovery key, and the server looks the account up by that key's address. The server can only
/// do that if it was told the address. This sends it once: at first run when it can, otherwise
/// the first time the app is signed in afterwards. It never blocks sign-in or first run: a
/// failure is retried at the next connect.
enum RecoveryLinkSync {
    private static let syncedKey = "echo.recovery.linkedAddress"

    enum Outcome: Equatable {
        case nothingToSend, alreadySent, sent, refused, failed
    }

    struct LinkEndpoint: APIEndpoint {
        var path: String { "/v1/identity/link-wallet" }
    }

    private struct Body: Encodable {
        let did: String
        let wallet_address: String
    }

    private struct Reply: Decodable {
        let linked: Bool?
    }

    /// Whether `address` still needs sending, given what was last sent.
    static func needsSync(address: String?, lastSent: String?) -> Bool {
        guard let address, !address.isEmpty else { return false }
        return address != lastSent
    }

    @discardableResult
    static func syncIfNeeded(apiClient: APIClient, did knownDID: String? = nil, defaults: UserDefaults = .standard) async -> Outcome {
        // Saved with the typed Keychain helper; the untyped one looks elsewhere and finds nothing.
        let address = (try? await KeychainManager.shared.retrieve(key: WalletKeychain.dagAddressKey, as: String.self)) ?? nil
        guard let address, !address.isEmpty else { return .nothingToSend }
        guard needsSync(address: address, lastSent: defaults.string(forKey: syncedKey)) else { return .alreadySent }
        var resolved = knownDID
        if resolved == nil || resolved?.isEmpty == true { resolved = await CurrentUserSession.currentDID() }
        guard let did = resolved, !did.isEmpty else { return .failed }
        do {
            let _: Reply = try await apiClient.post(
                endpoint: LinkEndpoint(),
                body: Body(did: did, wallet_address: address)
            )
            defaults.set(address, forKey: syncedKey)
            return .sent
        } catch APIError.unexpectedStatusCode(409) {
            // The address belongs to another account. Sending it again will not change that.
            defaults.set(address, forKey: syncedKey)
            return .refused
        } catch {
            return .failed
        }
    }
}

// MARK: - Session bootstrap

/// Supplies the access token the realtime socket (and Bearer-only routes) need.
///
/// A device that enrolled through first-run has a registered Secure Enclave key
/// but was never issued a token, and tokens from account recovery expire after
/// 15 minutes with no refresh token. Without this, `MessageRelaySession.connect()`
/// found nothing in the keychain and returned without opening the socket.
///
/// Order of preference: a stored token that is still valid, then the refresh
/// token (no biometric prompt, so it also works on a background push wake), then
/// a passkey-signed `POST /v1/auth/session`.
actor SessionTokenBootstrap {
    static let shared = SessionTokenBootstrap()

    private static let deviceIDKey = "echo.session.deviceId"
    private var inFlight: Task<String?, Never>?

    /// A stable per-install identifier; refresh tokens are bound to it server-side.
    static var deviceID: String {
        if let existing = UserDefaults.standard.string(forKey: deviceIDKey), !existing.isEmpty {
            return existing
        }
        let fresh = UUID().uuidString
        UserDefaults.standard.set(fresh, forKey: deviceIDKey)
        return fresh
    }

    /// Returns a token valid for at least another minute, or nil if none can be
    /// obtained right now (offline, signed out, or the key is not registered).
    /// Concurrent callers share one request: refresh tokens are single-use, and
    /// presenting the same one twice revokes the whole session server-side.
    func validAccessToken() async -> String? {
        if let inFlight {
            return await inFlight.value
        }
        let task = Task<String?, Never> { await Self.resolve() }
        inFlight = task
        let value = await task.value
        inFlight = nil
        return value
    }

    /// Forgets the access token after the server refused it, so the next request for a token
    /// refreshes or opens a new session instead of presenting the same one again.
    func discardAccessToken() async {
        try? await KeychainManager.shared.delete(key: "authToken")
    }

    private static func resolve() async -> String? {
        let currentDID = await CurrentUserSession.currentDID()
        if let stored = try? await KeychainManager.shared.getAuthToken(), !stored.isEmpty {
            if belongsToSomeoneElse(stored, currentDID: currentDID) {
                // A session left over from a different identity. The Keychain outlives the app:
                // after deleting the app and creating a new account, the old account's tokens
                // are still here. Using them would sign the socket in as the old identity while
                // every signed request speaks as the new one. Drop them and start again.
                try? await KeychainManager.shared.delete(key: "authToken")
                try? await KeychainManager.shared.delete(key: "refreshToken")
                return await openSession()
            }
            if !isExpiring(stored) { return stored }
        }
        if let refreshed = await refresh(), !belongsToSomeoneElse(refreshed, currentDID: currentDID) {
            return refreshed
        }
        return await openSession()
    }

    /// The identity a session token was issued to (its `sub` claim), if readable.
    static func subject(of jwt: String) -> String? {
        let parts = jwt.split(separator: ".")
        guard parts.count == 3 else { return nil }
        var payload = String(parts[1])
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        while payload.count % 4 != 0 { payload.append("=") }
        guard let data = Data(base64Encoded: payload),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else { return nil }
        return json["sub"] as? String
    }

    /// True when the token names an identity other than the one this device now is.
    static func belongsToSomeoneElse(_ jwt: String, currentDID: String?) -> Bool {
        guard let currentDID, !currentDID.isEmpty, let subject = subject(of: jwt), !subject.isEmpty else {
            return false
        }
        return subject != currentDID
    }

    /// True when the JWT's `exp` is missing, unreadable, or under a minute away.
    static func isExpiring(_ jwt: String, now: Date = Date()) -> Bool {
        let parts = jwt.split(separator: ".")
        guard parts.count == 3 else { return true }
        var payload = String(parts[1])
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")
        while payload.count % 4 != 0 { payload.append("=") }
        guard let data = Data(base64Encoded: payload),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let exp = (json["exp"] as? NSNumber)?.doubleValue else {
            return true
        }
        return exp - now.timeIntervalSince1970 < 60
    }

    private static func refresh() async -> String? {
        guard let refreshToken = try? await KeychainManager.shared.getRefreshToken(),
              !refreshToken.isEmpty else { return nil }
        var req = URLRequest(url: EchoAPIBaseURL.url(path: "/v3/auth/refresh"))
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.timeoutInterval = 15
        req.httpBody = try? JSONSerialization.data(withJSONObject: [
            "refresh_token": refreshToken,
            "device_id": deviceID,
        ])
        return await send(req)
    }

    private static func openSession() async -> String? {
        var req = URLRequest(url: EchoAPIBaseURL.url(path: "/v1/auth/session"))
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.timeoutInterval = 15
        req.httpBody = try? JSONSerialization.data(withJSONObject: ["device_id": deviceID])
        do {
            // Adds X-Sender-DID / X-Signature / X-Timestamp from the Secure Enclave key.
            try await PasskeySigningInterceptor().intercept(&req)
        } catch {
            return nil
        }
        guard req.value(forHTTPHeaderField: "X-Signature") != nil else { return nil }
        return await send(req)
    }

    /// Sends a token request and persists the pair it returns.
    private static func send(_ req: URLRequest) async -> String? {
        guard let (data, response) = try? await URLSession.shared.data(for: req),
              let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let access = json["access_token"] as? String, !access.isEmpty else {
            return nil
        }
        try? await KeychainManager.shared.storeAuthToken(access)
        if let refresh = json["refresh_token"] as? String, !refresh.isEmpty {
            try? await KeychainManager.shared.storeRefreshToken(refresh)
        }
        return access
    }
}
#endif
