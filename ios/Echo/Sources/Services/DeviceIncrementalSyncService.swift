#if os(iOS)
import Foundation

/// Pushes incremental message deltas to linked devices (M3b / WO-CA3).
@MainActor
final class DeviceIncrementalSyncService {
    private let syncAPI: DeviceSyncAPIClient
    private let crypto: DeviceSyncCrypto
    private let registry: LinkedDevicesRegistry

    init(
        syncAPI: DeviceSyncAPIClient,
        crypto: DeviceSyncCrypto,
        registry: LinkedDevicesRegistry = .shared
    ) {
        self.syncAPI = syncAPI
        self.crypto = crypto
        self.registry = registry
    }

    func pushUpsert(conversationId: String, message: StoredThreadMessage) async {
        await push(envelope: .upsert(conversationId: conversationId, message: message), entryType: DeviceSyncEntryType.message)
    }

    func pushTombstone(conversationId: String, messageId: String) async {
        await push(envelope: .tombstone(conversationId: conversationId, messageId: messageId), entryType: DeviceSyncEntryType.tombstone)
    }

    private func push(envelope: DeviceSyncMessageEnvelope, entryType: String) async {
        guard !envelope.conversationId.isEmpty else { return }
        guard let plaintext = try? JSONEncoder().encode(envelope) else { return }

        let targets = await registry.linkedDevices()
        guard !targets.isEmpty else { return }

        for device in targets {
            do {
                let recipientKey = try TextMessageCrypto.dataFromPublicKeyHex(device.publicKeyHex)
                let ciphertext = try await crypto.wrap(plaintext: plaintext, recipientPublicKey: recipientKey)
                let targetDeviceId = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: device.publicKeyHex)
                _ = try await syncAPI.push(
                    targetDeviceId: targetDeviceId,
                    ciphertext: ciphertext,
                    entryType: entryType
                )
            } catch {
                continue
            }
        }
    }
}

/// Fire-and-forget incremental sync after local thread mutations.
enum DeviceSyncOutboundCoordinator {
    static func notifyMessageSaved(conversationId: String, message: StoredThreadMessage) {
        Task { @MainActor in
            guard let service = DIContainer.shared.resolveDeviceIncrementalSync() else { return }
            await service.pushUpsert(conversationId: conversationId, message: message)
        }
    }

    static func notifyMessageDeleted(conversationId: String, messageId: String) {
        Task { @MainActor in
            guard let service = DIContainer.shared.resolveDeviceIncrementalSync() else { return }
            await service.pushTombstone(conversationId: conversationId, messageId: messageId)
        }
    }
}
#endif
