import Foundation

/// Configuration for WebSocket client
struct WebSocketConfiguration {
    let baseURL: URL
    let reconnectAttempts: Int
    let reconnectDelay: TimeInterval
    let heartbeatInterval: TimeInterval
    
    static let `default` = WebSocketConfiguration(
        baseURL: URL(string: "wss://ws.echo.local")!,
        reconnectAttempts: 5,
        reconnectDelay: 2,
        heartbeatInterval: 30
    )
}

/// WebSocket message types
enum WebSocketMessage: Codable {
    case text(String)
    case data(Data)
    case control(ControlMessage)
    
    enum CodingKeys: String, CodingKey {
        case type, payload
    }
    
    enum MessageType: String, Codable {
        case text, data, control
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let type = try container.decode(MessageType.self, forKey: .type)
        
        switch type {
        case .text:
            let payload = try container.decode(String.self, forKey: .payload)
            self = .text(payload)
        case .data:
            let payload = try container.decode(Data.self, forKey: .payload)
            self = .data(payload)
        case .control:
            let payload = try container.decode(ControlMessage.self, forKey: .payload)
            self = .control(payload)
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        
        switch self {
        case .text(let payload):
            try container.encode(MessageType.text, forKey: .type)
            try container.encode(payload, forKey: .payload)
        case .data(let payload):
            try container.encode(MessageType.data, forKey: .type)
            try container.encode(payload, forKey: .payload)
        case .control(let payload):
            try container.encode(MessageType.control, forKey: .type)
            try container.encode(payload, forKey: .payload)
        }
    }
}

/// Control messages for WebSocket protocol
struct ControlMessage: Codable {
    let action: String
    let data: [String: String]?

    enum ActionType: String {
        case ping = "ping"
        case pong = "pong"
        case subscribe = "subscribe"
        case unsubscribe = "unsubscribe"
        case acknowledge = "acknowledge"
    }
}

// MARK: - Relay WebSocket Message Types (v3.0+)

/// Typed relay messages for the content-blind WebSocket relay
struct WSRelayMessage: Codable {
    let type: RelayMessageType
    let payload: Data
    let timestamp: Date

    enum RelayMessageType: String, Codable {
        case message            // E2E encrypted message blob
        case typing             // Typing indicator
        case presence           // Online/offline status
        case receipt            // Read/delivery receipt
        case ack                // Server acknowledgement
        case queueDrain         // Offline queue delivery on reconnect
        case confirmation       // On-chain anchoring confirmation
        case groupKey           // Group key distribution
    }
}

/// On-chain anchoring confirmation payload
struct WSConfirmation: Codable {
    let referenceId: String
    let snapshotHash: String
    let snapshotHeight: Int
    let merkleProof: [Data]?
}

/// Group key distribution payload
struct WSGroupKeyPayload: Codable {
    let groupId: String
    let version: Int
    let encryptedKey: Data
    let distributedBy: String
}

/// WebSocket delegate for handling events
protocol WebSocketDelegate: AnyObject {
    func webSocketDidConnect(_ client: WebSocketClient)
    func webSocketDidDisconnect(_ client: WebSocketClient, error: Error?)
    func webSocketDidReceiveMessage(_ client: WebSocketClient, message: String)
    func webSocketDidReceiveData(_ client: WebSocketClient, data: Data)
    func webSocketDidReceiveError(_ client: WebSocketClient, error: Error)
}

/// Real-time WebSocket client for messaging
actor WebSocketClient: NSObject, URLSessionWebSocketDelegate {
    
    // MARK: - Properties
    
    private let configuration: WebSocketConfiguration
    private var webSocket: URLSessionWebSocketTask?
    private var session: URLSession?
    private weak var delegate: WebSocketDelegate?
    
    // Connection state
    private var isConnected = false
    private var reconnectAttempt = 0
    private var heartbeatTask: Task<Void, Never>?
    private var receiveTask: Task<Void, Never>?
    
    // Message queue
    private var messageQueue: [WebSocketMessage] = []
    private var isQueueing = false
    
    // MARK: - Initialization
    
    init(configuration: WebSocketConfiguration = .default) {
        self.configuration = configuration
        super.init()
    }
    
    deinit {
        heartbeatTask?.cancel()
        receiveTask?.cancel()
    }
    
    // MARK: - Connection Management
    
    /// Connect to WebSocket server
    func connect(delegate: WebSocketDelegate) async throws {
        self.delegate = delegate
        
        let config = URLSessionConfiguration.default
        config.tlsMinimumSupportedProtocolVersion = .TLSv13
        self.session = URLSession(configuration: config, delegate: self, delegateQueue: nil)
        
        guard let session = session else {
            throw WebSocketError.sessionCreationFailed
        }
        
        let request = URLRequest(url: configuration.baseURL)
        webSocket = session.webSocketTask(with: request)
        webSocket?.resume()
        
        isConnected = true
        reconnectAttempt = 0
        
        DispatchQueue.main.async {
            self.delegate?.webSocketDidConnect(self)
        }
        
        // Start receiving messages and heartbeat
        startHeartbeat()
        startReceiving()
    }
    
    /// Disconnect from WebSocket server
    func disconnect() async {
        isConnected = false
        heartbeatTask?.cancel()
        receiveTask?.cancel()
        
        try? await webSocket?.cancel(with: .goingAway, reason: nil)
        webSocket = nil
        session?.invalidateAndCancel()
    }
    
    /// Reconnect with exponential backoff
    private func reconnect() async {
        guard reconnectAttempt < configuration.reconnectAttempts else {
            let error = WebSocketError.maxReconnectAttemptsExceeded
            DispatchQueue.main.async {
                self.delegate?.webSocketDidReceiveError(self, error: error)
            }
            return
        }
        
        reconnectAttempt += 1
        let delay = configuration.reconnectDelay * Double(reconnectAttempt)
        
        try? await Task.sleep(nanoseconds: UInt64(delay * 1_000_000_000))
        
        do {
            try await connect(delegate: delegate ?? DummyDelegate())
        } catch {
            DispatchQueue.main.async {
                self.delegate?.webSocketDidReceiveError(self, error: error)
            }
        }
    }
    
    // MARK: - Message Sending
    
    /// Send a text message
    func send(text: String) async throws {
        guard let webSocket = webSocket, isConnected else {
            messageQueue.append(.text(text))
            isQueueing = true
            throw WebSocketError.notConnected
        }
        
        try await webSocket.send(.string(text))
    }
    
    /// Send binary data
    func send(data: Data) async throws {
        guard let webSocket = webSocket, isConnected else {
            messageQueue.append(.data(data))
            isQueueing = true
            throw WebSocketError.notConnected
        }
        
        try await webSocket.send(.data(data))
    }
    
    /// Send a control message
    func send(controlMessage: ControlMessage) async throws {
        let encoder = JSONEncoder()
        let data = try encoder.encode(controlMessage)
        
        guard let jsonString = String(data: data, encoding: .utf8) else {
            throw WebSocketError.encodingError
        }
        
        try await send(text: jsonString)
    }
    
    // MARK: - Message Receiving
    
    private func startReceiving() {
        receiveTask = Task {
            while isConnected {
                do {
                    guard let webSocket = webSocket else { break }
                    
                    let message = try await webSocket.receive()
                    await handleMessage(message)
                } catch {
                    if isConnected {
                        DispatchQueue.main.async {
                            self.delegate?.webSocketDidReceiveError(self, error: error)
                        }
                        await reconnect()
                    }
                    break
                }
            }
        }
    }
    
    private func handleMessage(_ message: URLSessionWebSocketTask.Message) async {
        switch message {
        case .string(let text):
            DispatchQueue.main.async {
                self.delegate?.webSocketDidReceiveMessage(self, message: text)
            }
            
        case .data(let data):
            DispatchQueue.main.async {
                self.delegate?.webSocketDidReceiveData(self, data: data)
            }
            
        @unknown default:
            break
        }
    }
    
    // MARK: - Heartbeat
    
    private func startHeartbeat() {
        heartbeatTask = Task {
            while isConnected {
                try? await Task.sleep(nanoseconds: UInt64(configuration.heartbeatInterval * 1_000_000_000))
                
                if isConnected {
                    let ping = ControlMessage(action: "ping", data: nil)
                    try? await send(controlMessage: ping)
                }
            }
        }
    }
    
    // MARK: - URLSessionWebSocketDelegate
    
    nonisolated func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didOpenWithProtocol protocol: String?
    ) {
        // Handle connection opened
    }
    
    nonisolated func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
        reason: Data?
    ) {
        Task {
            await self.handleDisconnection()
        }
    }
    
    // MARK: - Connection State
    
    private func handleDisconnection() async {
        isConnected = false
        heartbeatTask?.cancel()
        receiveTask?.cancel()
        
        DispatchQueue.main.async {
            self.delegate?.webSocketDidDisconnect(self, error: nil)
        }
        
        if reconnectAttempt < configuration.reconnectAttempts {
            await reconnect()
        }
    }
    
    // MARK: - State Accessors
    
    var connected: Bool {
        isConnected
    }
    
    var queuedMessages: [WebSocketMessage] {
        messageQueue
    }
    
    /// Flush queued messages
    func flushQueue() async throws {
        for message in messageQueue {
            switch message {
            case .text(let text):
                try await send(text: text)
            case .data(let data):
                try await send(data: data)
            case .control(let control):
                try await send(controlMessage: control)
            }
        }
        
        messageQueue.removeAll()
        isQueueing = false
    }
}

// MARK: - Dummy Delegate for reconnection

private class DummyDelegate: WebSocketDelegate {
    func webSocketDidConnect(_ client: WebSocketClient) {}
    func webSocketDidDisconnect(_ client: WebSocketClient, error: Error?) {}
    func webSocketDidReceiveMessage(_ client: WebSocketClient, message: String) {}
    func webSocketDidReceiveData(_ client: WebSocketClient, data: Data) {}
    func webSocketDidReceiveError(_ client: WebSocketClient, error: Error) {}
}

// MARK: - WebSocket Errors

enum WebSocketError: LocalizedError {
    case notConnected
    case sessionCreationFailed
    case encodingError
    case decodingError
    case invalidURL
    case maxReconnectAttemptsExceeded
    case connectionClosed
    case unknown(String)
    
    var errorDescription: String? {
        switch self {
        case .notConnected:
            return "WebSocket is not connected"
        case .sessionCreationFailed:
            return "Failed to create URLSession"
        case .encodingError:
            return "Failed to encode message"
        case .decodingError:
            return "Failed to decode message"
        case .invalidURL:
            return "Invalid WebSocket URL"
        case .maxReconnectAttemptsExceeded:
            return "Maximum reconnection attempts exceeded"
        case .connectionClosed:
            return "WebSocket connection was closed"
        case .unknown(let message):
            return "Unknown error: \(message)"
        }
    }
}
