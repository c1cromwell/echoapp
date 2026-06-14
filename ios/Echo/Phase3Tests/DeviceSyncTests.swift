import XCTest
import CryptoKit
@testable import Echo

#if os(iOS)
final class DeviceSyncCryptoTests: XCTestCase {
    func testWrapUnwrapRoundTrip() async throws {
        let recipientPrivate = P256.KeyAgreement.PrivateKey()
        let recipientPublic = recipientPrivate.publicKey.rawRepresentation

        let bundle = HistorySyncBundle(
            conversations: [
                StoredConversation(id: "dm:alice", contactName: "Alice", peerDID: "did:key:alice")
            ],
            threads: [
                "dm:alice": [
                    StoredThreadMessage(
                        id: "m1",
                        senderDID: "did:key:me",
                        content: "hello sync",
                        timestamp: "Now"
                    )
                ]
            ]
        )
        let plaintext = try bundle.encoded()

        let crypto = DeviceSyncCrypto()
        let recovered = try await crypto.roundTrip(
            plaintext: plaintext,
            recipientPublicKey: recipientPublic,
            recipientPrivateKey: recipientPrivate
        )

        let decoded = try HistorySyncBundle.decode(from: recovered)
        XCTAssertEqual(decoded, bundle)
    }

    func testUnwrapRejectsEmptyCiphertext() async throws {
        let crypto = DeviceSyncCrypto()
        let privateKey = P256.KeyAgreement.PrivateKey()
        do {
            _ = try await crypto.unwrap(ciphertext: Data(), ourPrivateKey: privateKey)
            XCTFail("expected error")
        } catch {
            XCTAssertTrue(error is DeviceSyncCryptoError)
        }
    }
}

final class HistorySyncBundleTests: XCTestCase {
    func testEncodeDecodeRoundTrip() throws {
        let bundle = HistorySyncBundle(
            conversations: [
                StoredConversation(id: "group:grp-1", contactName: "Team", peerDID: "grp-1")
            ],
            threads: [
                "group:grp-1": [
                    StoredThreadMessage(id: "g1", senderDID: "did:key:a", content: "hi", timestamp: "Now")
                ]
            ]
        )
        let data = try bundle.encoded()
        let decoded = try HistorySyncBundle.decode(from: data)
        XCTAssertEqual(decoded, bundle)
    }

    func testMergeIsIdempotent() {
        let conversation = StoredConversation(id: "dm:1", contactName: "Bob", peerDID: "did:key:bob")
        let msg = StoredThreadMessage(id: "m1", senderDID: "did:key:bob", content: "ping", timestamp: "Now")
        let bundle = HistorySyncBundle(conversations: [conversation], threads: ["dm:1": [msg]])

        ConversationThreadStore.mergeMessages(conversationId: "dm:1", incoming: [msg])
        ConversationThreadStore.mergeMessages(conversationId: "dm:1", incoming: [msg])

        let exported = ConversationThreadStore.exportMessages(conversationId: "dm:1")
        XCTAssertEqual(exported.count, 1)
        XCTAssertEqual(exported[0].id, "m1")
        _ = bundle // bundle shape validated in encode test
    }
}

final class DeviceIdentityStoreTests: XCTestCase {
    func testDeviceIdIsStable() async {
        await DeviceIdentityStore.resetForTesting()
        let a = await DeviceIdentityStore.currentDeviceId()
        let b = await DeviceIdentityStore.currentDeviceId()
        XCTAssertEqual(a, b)
        XCTAssertTrue(a.hasPrefix("dev-"))
        await DeviceIdentityStore.resetForTesting()
    }

    func testSyncDeviceIdIsDeterministic() {
        let hex = "04" + String(repeating: "ab", count: 32)
        let a = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: hex)
        let b = DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: hex.uppercased())
        XCTAssertEqual(a, b)
        XCTAssertTrue(a.hasPrefix("dev-"))
    }

    func testAssignSyncDeviceIdPinsStream() async {
        await DeviceIdentityStore.resetForTesting()
        let hex = "04" + String(repeating: "cd", count: 32)
        await DeviceIdentityStore.assignSyncDeviceId(fromPublicKeyHex: hex)
        let currentId = await DeviceIdentityStore.currentDeviceId()
        XCTAssertEqual(currentId, DeviceIdentityStore.syncDeviceId(fromPublicKeyHex: hex))
        await DeviceIdentityStore.resetForTesting()
    }
}

final class SyncCursorStoreTests: XCTestCase {
    func testCursorPersistence() {
        SyncCursorStore.reset(deviceId: "dev-test")
        XCTAssertEqual(SyncCursorStore.load(deviceId: "dev-test"), 0)
        SyncCursorStore.save(deviceId: "dev-test", cursor: 42)
        XCTAssertEqual(SyncCursorStore.load(deviceId: "dev-test"), 42)
        SyncCursorStore.reset(deviceId: "dev-test")
    }
}
#endif
