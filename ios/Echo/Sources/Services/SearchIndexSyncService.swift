#if os(iOS)
import Foundation

/// WO-73: push/pull encrypted search index blob via device sync streams.
actor SearchIndexSyncService {
    private let syncAPI: DeviceSyncAPIClient
    private let crypto: DeviceSyncCrypto
    private let indexStore = EncryptedIndexStore.shared

    init(syncAPI: DeviceSyncAPIClient, crypto: DeviceSyncCrypto = DeviceSyncCrypto()) {
        self.syncAPI = syncAPI
        self.crypto = crypto
    }

    /// Primary: export encrypted index file and push to linked device's stream.
    func pushIndexToDevice(publicKeyHex: String) async throws {
        let trimmed = publicKeyHex.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { throw DeviceHistorySyncError.missingPublicKey }
        let snapshot = await LocalMessageIndexer.shared.currentSnapshot()
        let plaintext = try JSONEncoder().encode(snapshot)
        let recipientKey = try TextMessageCrypto.dataFromPublicKeyHex(trimmed)
        let ciphertext = try await crypto.wrap(plaintext: plaintext, recipientPublicKey: recipientKey)
        let targetDeviceId = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: trimmed)
        _ = try await syncAPI.push(
            targetDeviceId: targetDeviceId,
            ciphertext: ciphertext,
            entryType: DeviceSyncEntryType.searchIndex
        )
    }

    /// Linked device: apply search_index sync entries after history pull.
    @discardableResult
    func pullAndApplySearchIndex() async throws -> Bool {
        let deviceId = await DeviceIdentityStore.currentDeviceId()
        var cursor = SyncCursorStore.load(deviceId: deviceId)
        var applied = false

        while true {
            let page = try await syncAPI.pull(deviceId: deviceId, after: cursor, limit: 20)
            guard !page.entries.isEmpty else { break }
            for entry in page.entries {
                guard entry.entryType == DeviceSyncEntryType.searchIndex else { continue }
                let plaintext = try await crypto.unwrapWithLocalKey(ciphertext: entry.ciphertext)
                let snapshot = try JSONDecoder().decode(SearchIndexSnapshot.self, from: plaintext)
                try await indexStore.save(snapshot)
                await LocalMessageIndexer.shared.importSnapshot(snapshot)
                applied = true
            }
            cursor = page.nextCursor
            SyncCursorStore.save(deviceId: deviceId, cursor: cursor)
            if page.entries.count < 20 { break }
        }
        return applied
    }
}

/// Fire-and-forget pull for WO-73 search index sync entries.
enum SearchIndexSyncBootstrap {
    static func pullIfNeeded() {
        Task {
            guard let syncAPI = DIContainer.shared.resolveDeviceSyncAPI(),
                  let crypto = DIContainer.shared.resolveDeviceSyncCrypto() else { return }
            let service = SearchIndexSyncService(syncAPI: syncAPI, crypto: crypto)
            _ = try? await service.pullAndApplySearchIndex()
        }
    }
}
#endif
