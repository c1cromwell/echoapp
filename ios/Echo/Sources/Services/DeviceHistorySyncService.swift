#if os(iOS)
import Foundation

/// Orchestrates WO-CA3 history seed (primary → linked device) and pull/apply (new device).
@MainActor
final class DeviceHistorySyncService {
    private let syncAPI: DeviceSyncAPIClient
    private let crypto: DeviceSyncCrypto
    private let conversationStore: ConversationStore
    private let pullPageSize = 50

    init(
        syncAPI: DeviceSyncAPIClient,
        crypto: DeviceSyncCrypto,
        conversationStore: ConversationStore = .shared
    ) {
        self.syncAPI = syncAPI
        self.crypto = crypto
        self.conversationStore = conversationStore
    }

    /// Primary device: export local history + search index to a newly linked device.
    func seedAllToDevice(publicKeyHex: String) async throws {
        try await seedHistoryToDevice(publicKeyHex: publicKeyHex)
        let indexSync = SearchIndexSyncService(syncAPI: syncAPI, crypto: crypto)
        try? await indexSync.pushIndexToDevice(publicKeyHex: publicKeyHex)
    }

    /// Primary device: export local history, ECDH-wrap to target pubkey, push to its sync stream.
    func seedHistoryToDevice(publicKeyHex: String) async throws {
        let trimmed = publicKeyHex.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { throw DeviceHistorySyncError.missingPublicKey }

        let targetDeviceId = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: trimmed)
        let recipientKey = try TextMessageCrypto.dataFromPublicKeyHex(trimmed)
        let bundle = HistorySyncBundleBuilder.build(from: conversationStore)
        let plaintext = try bundle.encoded()
        let ciphertext = try await crypto.wrap(plaintext: plaintext, recipientPublicKey: recipientKey)
        _ = try await syncAPI.push(
            targetDeviceId: targetDeviceId,
            ciphertext: ciphertext,
            entryType: DeviceSyncEntryType.history
        )
    }

    /// Linked device: pull pending sync entries (history + WO-73 search index) and apply.
    @discardableResult
    func pullAndApplyPendingHistory() async throws -> Int {
        let deviceId = await DeviceIdentityStore.currentDeviceId()
        var cursor = SyncCursorStore.load(deviceId: deviceId)
        var appliedEntries = 0
        let indexStore = EncryptedIndexStore.shared

        while true {
            let page = try await syncAPI.pull(deviceId: deviceId, after: cursor, limit: pullPageSize)
            guard !page.entries.isEmpty else { break }

            for entry in page.entries {
                switch entry.entryType ?? DeviceSyncEntryType.history {
                case DeviceSyncEntryType.history:
                    let plaintext = try await crypto.unwrapWithLocalKey(ciphertext: entry.ciphertext)
                    let bundle = try HistorySyncBundle.decode(from: plaintext)
                    HistorySyncBundleMerger.apply(bundle, to: conversationStore)
                    appliedEntries += 1
                case DeviceSyncEntryType.message, DeviceSyncEntryType.tombstone:
                    let plaintext = try await crypto.unwrapWithLocalKey(ciphertext: entry.ciphertext)
                    let envelope = try JSONDecoder().decode(DeviceSyncMessageEnvelope.self, from: plaintext)
                    applyIncremental(envelope)
                    appliedEntries += 1
                case DeviceSyncEntryType.searchIndex:
                    let plaintext = try await crypto.unwrapWithLocalKey(ciphertext: entry.ciphertext)
                    let snapshot = try JSONDecoder().decode(SearchIndexSnapshot.self, from: plaintext)
                    try indexStore.save(snapshot)
                    await LocalMessageIndexer.shared.importSnapshot(snapshot)
                    appliedEntries += 1
                default:
                    continue
                }
            }

            cursor = page.nextCursor
            SyncCursorStore.save(deviceId: deviceId, cursor: cursor)
            if page.entries.count < pullPageSize { break }
        }

        return appliedEntries
    }

    private func applyIncremental(_ envelope: DeviceSyncMessageEnvelope) {
        switch envelope.operation {
        case .upsert:
            guard let message = envelope.message else { return }
            ConversationThreadStore.upsertMessage(conversationId: envelope.conversationId, message: message)
        case .tombstone:
            guard let messageId = envelope.messageId else { return }
            ConversationThreadStore.removeMessage(conversationId: envelope.conversationId, messageId: messageId)
        }
    }

    /// Revoke a linked device's sync stream (e.g. when removing the device).
    func revokeDevice(publicKeyHex: String) async throws {
        let targetDeviceId = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: publicKeyHex)
        try await syncAPI.revoke(targetDeviceId: targetDeviceId)
    }
}

enum DeviceHistorySyncError: LocalizedError {
    case missingPublicKey

    var errorDescription: String? {
        switch self {
        case .missingPublicKey: return "Linked device public key is missing."
        }
    }
}

/// Fire-and-forget pull after unlock; applies history + WO-73 search index (S4).
enum DeviceHistorySyncBootstrap {
    static func pullIfNeeded() {
        Task { @MainActor in
            guard let service = DIContainer.shared.resolveDeviceHistorySync() else { return }
            _ = try? await service.pullAndApplyPendingHistory()
        }
    }
}
#endif
