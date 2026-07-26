import XCTest
#if canImport(CryptoKit)
import CryptoKit
#endif
@testable import Echo

/// Unit coverage for Signal Parity Waves S2/S4 pure logic (no Xcode UI).
final class SignalParityWaveTests: XCTestCase {
    func testSafetyNumberIsSymmetric() {
        let a = "did:key:z6MkAlice"
        let b = "did:key:z6MkBob"
        let ab = SafetyNumber.digits(localDID: a, peerDID: b)
        let ba = SafetyNumber.digits(localDID: b, peerDID: a)
        XCTAssertEqual(ab, ba)
        XCTAssertEqual(ab.count, 60)
        XCTAssertFalse(SafetyNumber.formatted(localDID: a, peerDID: b).isEmpty)
    }

    func testSafetyNumberChangesWhenPeerChanges() {
        let local = "did:key:z6MkAlice"
        let peer1 = "did:key:z6MkBob"
        let peer2 = "did:key:z6MkCarol"
        XCTAssertNotEqual(
            SafetyNumber.digits(localDID: local, peerDID: peer1),
            SafetyNumber.digits(localDID: local, peerDID: peer2)
        )
    }

    func testGroupSenderKeyRoundTrip() async throws {
        let store = GroupSenderKeyStore()
        let root = SymmetricKey(size: .bits256)
        let groupId = "group-test"
        let alice = "did:key:alice"
        let bob = "did:key:bob"
        await store.seedChain(groupId: groupId, senderDID: alice, groupKeyVersion: 1, rootKey: root)
        await store.seedChain(groupId: groupId, senderDID: alice, groupKeyVersion: 1, rootKey: root)
        // Bob needs same seed to decrypt Alice's chain
        let bobStore = GroupSenderKeyStore()
        await bobStore.seedChain(groupId: groupId, senderDID: alice, groupKeyVersion: 1, rootKey: root)

        let plain = Data("hello group".utf8)
        let (cipher, iter) = try await store.encrypt(
            plaintext: plain,
            groupId: groupId,
            senderDID: alice,
            groupKeyVersion: 1
        )
        let out = try await bobStore.decrypt(
            ciphertext: cipher,
            groupId: groupId,
            senderDID: alice,
            groupKeyVersion: 1,
            iteration: iter
        )
        XCTAssertEqual(out, plain)
    }

    func testViewOnceBurn() {
        let id = UUID().uuidString
        XCTAssertFalse(ViewOnceStore.isBurned(messageId: id))
        ViewOnceStore.markBurned(messageId: id)
        XCTAssertTrue(ViewOnceStore.isBurned(messageId: id))
    }

    func testMessageRequestAcceptFlow() {
        let did = "did:key:z6MkRequestPeer-\(UUID().uuidString)"
        MessageRequestStore.enqueue(peerDID: did)
        XCTAssertTrue(MessageRequestStore.isPending(peerDID: did))
        MessageRequestStore.accept(peerDID: did)
        XCTAssertTrue(MessageRequestStore.isAccepted(peerDID: did))
        XCTAssertFalse(MessageRequestStore.isPending(peerDID: did))
    }
}
