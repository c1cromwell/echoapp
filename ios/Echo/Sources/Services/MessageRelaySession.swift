#if os(iOS)
import Foundation

/// Shared WS reconnect used by the Messages tab and by APNs wake-up.
/// Opening the socket triggers server `flushOffline`, so queued messages land
/// even if the recipient was not in the app when they were sent.
enum MessageRelaySession {
    @MainActor
    static func connect() async {
        guard let service = DIContainer.shared.resolveConversationSignalService() else { return }
        // A stored token may be missing (new install) or expired (they last 15
        // minutes); the bootstrap renews it or opens a session with the passkey.
        guard let token = await SessionTokenBootstrap.shared.validAccessToken(), !token.isEmpty else { return }

        service.setInboundTextHandler { event in
            Task { @MainActor in
                await ingestInbound(event)
            }
        }

        try? await service.connect(accessToken: token)
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

    private static func resolve() async -> String? {
        if let stored = try? await KeychainManager.shared.getAuthToken(),
           !stored.isEmpty, !isExpiring(stored) {
            return stored
        }
        if let refreshed = await refresh() {
            return refreshed
        }
        return await openSession()
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
