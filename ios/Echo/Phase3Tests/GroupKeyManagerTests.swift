import XCTest
import CryptoKit
@testable import Echo

#if os(iOS)
final class GroupKeyManagerTests: XCTestCase {
    func testDistributeAndDecryptRoundTrip() async throws {
        let adminEncryption = KinnamiEncryption()
        let memberPrivate = P256.KeyAgreement.PrivateKey()
        let memberPublic = memberPrivate.publicKey.rawRepresentation

        let manager = GroupKeyManager(encryption: adminEncryption)
        let info = await manager.generateGroupKey(groupId: "grp-test")

        let packages = try await manager.distributeGroupKey(
            groupId: "grp-test",
            keyInfo: info,
            members: [.init(did: "did:key:member", keyAgreementPublicKey: memberPublic)],
            encryption: adminEncryption
        )
        XCTAssertEqual(packages.count, 1)

        let memberEncryption = KinnamiEncryption()
        let keyData = try await manager.decryptKeyPackage(
            packages[0].encryptedKey,
            ourPrivateKey: memberPrivate,
            encryption: memberEncryption
        )
        await manager.storeReceivedKey(groupId: "grp-test", version: info.version, keyData: keyData)

        let plaintext = Data("hello group".utf8)
        let ciphertext = try manager.encryptForGroup(plaintext: plaintext, groupId: "grp-test")
        let decrypted = try manager.decryptFromGroup(
            ciphertext: ciphertext,
            groupId: "grp-test",
            keyVersion: info.version
        )
        XCTAssertEqual(String(data: decrypted, encoding: .utf8), "hello group")
    }

    func testGroupKeyPayloadCodec() throws {
        let payload = GroupKeyPayload(
            groupId: "grp-1",
            version: 2,
            encryptedKey: Data("opaque".utf8),
            distributedBy: "did:key:admin"
        )
        let envelope = WSEnvelope(
            type: ConversationSignalType.groupKey,
            to: "did:key:member",
            from: "did:key:admin",
            conversationId: "group:grp-1",
            payload: payload,
            timestamp: ConversationSignalCodec.isoTimestamp()
        )
        let data = try JSONEncoder().encode(envelope)
        let text = String(data: data, encoding: .utf8)!
        let event = try XCTUnwrap(try ConversationSignalCodec.decodeEvent(from: text))
        guard case .groupKey(let ev) = event else {
            return XCTFail("expected groupKey event")
        }
        XCTAssertEqual(ev.groupId, "grp-1")
        XCTAssertEqual(ev.version, 2)
    }
}
#endif
