import Foundation

/// Abstraction over WebSocket send/receive for conversation signals (testable).
protocol ConversationSignalTransport: AnyObject {
    func connect(accessToken: String) async throws
    func disconnect() async
    func send(text: String) async throws
    var onTextMessage: (@Sendable (String) -> Void)? { get set }
}

#if os(iOS)

/// Bridges `WebSocketClient` to [ConversationSignalTransport].
final class WebSocketConversationSignalTransport: NSObject, ConversationSignalTransport, @unchecked Sendable {
    private let webSocket: WebSocketClient
    private let apiBaseURL: URL
    private var bridge: WebSocketDelegateBridge?

    var onTextMessage: (@Sendable (String) -> Void)?

    init(webSocket: WebSocketClient = WebSocketClient(), apiBaseURL: URL) {
        self.webSocket = webSocket
        self.apiBaseURL = apiBaseURL
        super.init()
    }

    func connect(accessToken: String) async throws {
        let bridge = WebSocketDelegateBridge { [weak self] text in
            self?.onTextMessage?(text)
        }
        self.bridge = bridge
        try await webSocket.connect(accessToken: accessToken, apiBaseURL: apiBaseURL, delegate: bridge)
    }

    func disconnect() async {
        await webSocket.disconnect()
        bridge = nil
    }

    func send(text: String) async throws {
        try await webSocket.send(text: text)
    }
}

private final class WebSocketDelegateBridge: WebSocketDelegate {
    private let onMessage: (String) -> Void

    init(onMessage: @escaping (String) -> Void) {
        self.onMessage = onMessage
    }

    func webSocketDidConnect(_ client: WebSocketClient) {}
    func webSocketDidDisconnect(_ client: WebSocketClient, error: Error?) {}
    func webSocketDidReceiveMessage(_ client: WebSocketClient, message: String) {
        onMessage(message)
    }
    func webSocketDidReceiveData(_ client: WebSocketClient, data: Data) {}
    func webSocketDidReceiveError(_ client: WebSocketClient, error: Error) {}
}

#endif
