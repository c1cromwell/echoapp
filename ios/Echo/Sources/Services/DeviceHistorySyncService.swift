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

    /// Linked device: pull pending ciphertext entries, unwrap, merge into local stores.
    @discardableResult
    func pullAndApplyPendingHistory() async throws -> Int {
        let deviceId = await DeviceIdentityStore.currentDeviceId()
        var cursor = SyncCursorStore.load(deviceId: deviceId)
        var appliedBundles = 0

        while true {
            let page = try await syncAPI.pull(deviceId: deviceId, after: cursor, limit: pullPageSize)
            guard !page.entries.isEmpty else { break }

            for entry in page.entries {
                let entryType = entry.entryType ?? DeviceSyncEntryType.history
                guard entryType == DeviceSyncEntryType.history else { continue }
                let plaintext = try await crypto.unwrapWithLocalKey(ciphertext: entry.ciphertext)
                let bundle = try HistorySyncBundle.decode(from: plaintext)
                HistorySyncBundleMerger.apply(bundle, to: conversationStore)
                appliedBundles += 1
            }

            cursor = page.nextCursor
            SyncCursorStore.save(deviceId: deviceId, cursor: cursor)
            if page.entries.count < pullPageSize { break }
        }

        return appliedBundles
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

/// Fire-and-forget pull after unlock; safe to call from login and Messages tab.
enum DeviceHistorySyncBootstrap {
    static func pullIfNeeded() {
        Task { @MainActor in
            guard let service = DIContainer.shared.resolveDeviceHistorySync() else { return }
            _ = try? await service.pullAndApplyPendingHistory()
        }
    }
}
#endif
