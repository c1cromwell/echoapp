import Foundation
import CryptoKit

/// Manages message relay through the stateless WebSocket server.
/// The relay server transports E2E encrypted blobs it cannot read.
actor MessageRelayManager {

    private let webSocket: WebSocketClient
    private let encryption: KinnamiEncryption
    private let secureEnclave: SecureEnclaveManager
    private let offlineQueue: OfflineQueueManager
    private let anchoringTracker: AnchoringTracker

    init(
        webSocket: WebSocketClient,
        encryption: KinnamiEncryption,
        secureEnclave: SecureEnclaveManager,
        offlineQueue: OfflineQueueManager,
        anchoringTracker: AnchoringTracker
    ) {
        self.webSocket = webSocket
        self.encryption = encryption
        self.secureEnclave = secureEnclave
        self.offlineQueue = offlineQueue
        self.anchoringTracker = anchoringTracker
    }

    // MARK: - Send Flow

    /// Full message send pipeline: encrypt -> commit -> sign -> relay
    func sendMessage(
        plaintext: Data,
        contentType: String,
        recipientPublicKey: Data,
        conversationId: String
    ) async throws -> SendResult {
        // 1. E2E encrypt with Kinnami (X25519 + ChaCha20-Poly1305)
        let encrypted = try await encryption.encryptWithKeyAgreement(
            plaintext: String(data: plaintext, encoding: .utf8) ?? "",
            recipientPublicKeyData: recipientPublicKey
        )

        let payloadData = try JSONEncoder().encode(encrypted)

        // 2. Compute commitment hash (SHA-256 of encrypted payload)
        let commitment = SHA256.hash(data: payloadData)
        let commitmentData = Data(commitment)

        // 3. Sign the encrypted payload with Secure Enclave (P-256)
        let signature = try await secureEnclave.sign(
            data: payloadData,
            keyId: "messaging-signing-key"
        )

        // 4. Build relay request
        let messageId = UUID().uuidString
        let request = RelayRequest(
            messageId: messageId,
            conversationId: conversationId,
            contentType: contentType,
            encryptedPayload: payloadData,
            commitment: commitmentData,
            signature: signature,
            timestamp: Date()
        )

        do {
            // 5. Submit to relay via WebSocket
            let encoded = try JSONEncoder().encode(request)
            try await webSocket.send(data: encoded)

            // 6. Track commitment for on-chain anchoring confirmation
            await anchoringTracker.track(messageId: messageId, commitment: commitmentData)

            return SendResult(messageId: messageId, status: .sent)
        } catch {
            // 7. If relay unavailable, queue locally for retry
            await offlineQueue.enqueue(request)
            return SendResult(messageId: messageId, status: .sending)
        }
    }

    // MARK: - Receive Flow

    /// Process incoming relay message
    func handleIncomingMessage(_ payload: Data) async throws -> Data {
        let relayMsg = try JSONDecoder().decode(IncomingRelayMessage.self, from: payload)

        // 1. Verify sender signature (P-256)
        let senderKey = try P256.Signing.PublicKey(x963Representation: relayMsg.senderPublicKey)
        let isValid = await secureEnclave.verify(
            signature: relayMsg.signature,
            data: relayMsg.encryptedPayload,
            publicKey: senderKey
        )
        guard isValid else {
            throw MessageRelayError.invalidSignature
        }

        // 2. Verify commitment integrity
        let expectedCommitment = Data(SHA256.hash(data: relayMsg.encryptedPayload))
        guard expectedCommitment == relayMsg.commitment else {
            throw MessageRelayError.commitmentMismatch
        }

        // 3. Decrypt with own private key (Kinnami)
        // The decryption key is derived from the Secure Enclave master key.
        // In production, this key agreement private key is stored in the keychain
        // and retrieved via the SecureEnclaveManager's derived key hierarchy.
        let encryptedMessage = try JSONDecoder().decode(
            EncryptedMessageWithPublicKey.self,
            from: relayMsg.encryptedPayload
        )

        // Retrieve the messaging private key from Secure Enclave-derived storage
        let privateKeyData = try await secureEnclave.getPublicKey(id: "messaging-key-agreement")
        let privateKey = try P256.KeyAgreement.PrivateKey(rawRepresentation: privateKeyData)

        let decrypted = try await encryption.decryptWithKeyAgreement(
            encryptedMessage: encryptedMessage,
            ourPrivateKey: privateKey
        )

        return Data(decrypted.utf8)
    }

    // MARK: - Offline Queue Drain

    /// Called on WebSocket reconnect - drain any queued outbound messages
    func drainOfflineQueue() async {
        let queued = await offlineQueue.dequeueAll()
        for request in queued {
            do {
                let encoded = try JSONEncoder().encode(request)
                try await webSocket.send(data: encoded)
            } catch {
                // Re-queue if still failing
                await offlineQueue.enqueue(request)
            }
        }
    }

    // MARK: - WebSocket Message Routing

    /// Route incoming WebSocket relay messages by type
    func handleWSRelayMessage(_ wsMessage: WSRelayMessage) async {
        switch wsMessage.type {
        case .message, .queueDrain:
            // Decrypt and process (same flow for live and queued messages)
            try? await handleIncomingMessage(wsMessage.payload)

        case .confirmation:
            // On-chain anchoring confirmed
            if let conf = try? JSONDecoder().decode(WSConfirmation.self, from: wsMessage.payload) {
                await anchoringTracker.confirmAnchoring(
                    messageId: conf.referenceId,
                    snapshotHash: conf.snapshotHash,
                    snapshotHeight: conf.snapshotHeight,
                    merkleProof: conf.merkleProof
                )
            }

        case .typing, .presence, .receipt, .ack, .groupKey:
            break // Handled by dedicated managers
        }
    }
}

// MARK: - Supporting Types

struct RelayRequest: Codable {
    let messageId: String
    let conversationId: String
    let contentType: String
    let encryptedPayload: Data
    let commitment: Data
    let signature: Data
    let timestamp: Date
}

struct IncomingRelayMessage: Codable {
    let messageId: String
    let senderDID: String
    let senderPublicKey: Data
    let encryptedPayload: Data
    let commitment: Data
    let signature: Data
    let timestamp: Date
}

struct SendResult {
    let messageId: String
    let status: DeliveryStatus
}

enum MessageRelayError: LocalizedError {
    case invalidSignature
    case commitmentMismatch
    case decryptionFailed
    case relayUnavailable

    var errorDescription: String? {
        switch self {
        case .invalidSignature: return "Sender signature verification failed"
        case .commitmentMismatch: return "Commitment hash does not match content"
        case .decryptionFailed: return "Failed to decrypt message"
        case .relayUnavailable: return "Relay server is unavailable"
        }
    }
}
