import XCTest
@testable import Echo

final class PaymentPayloadTests: XCTestCase {
    func testEncodeDecodeRoundTrip() {
        let p = PaymentPayload(kind: .request, amount: Decimal(string: "12.50")!, memo: "lunch")
        XCTAssertEqual(PaymentPayload.decode(p.encoded()), p)
    }

    func testAmountValidation() {
        XCTAssertTrue(PaymentAmount.isValid(Decimal(string: "0.01")!))
        XCTAssertFalse(PaymentAmount.isValid(0))
        XCTAssertFalse(PaymentAmount.isValid(-5))
        XCTAssertFalse(PaymentAmount.isValid(PaymentAmount.maxAmount + 1))
    }

    func testAmountParse() {
        XCTAssertEqual(PaymentAmount.parse(" 12.50 "), Decimal(string: "12.50"))
        XCTAssertNil(PaymentAmount.parse("abc"))
        XCTAssertNil(PaymentAmount.parse("-3"))
    }
}

final class PaymentCoordinatorTests: XCTestCase {
    private let coord = PaymentCoordinator(transfer: MockWalletTransfer(),
                                           evaluator: ContactSafetyEvaluator())

    func testMakeRequestRejectsInvalidAmount() {
        XCTAssertThrowsError(try coord.makeRequest(amount: 0, memo: nil))
    }

    func testPayGuardForLowTrust() {
        XCTAssertTrue(coord.needsExtraConfirmation(peerTier: 0))
        XCTAssertFalse(coord.needsExtraConfirmation(peerTier: 3))
    }

    func testRequestThenPayProducesSentWithTxId() async throws {
        let request = try coord.makeRequest(amount: Decimal(string: "9.99")!, memo: "coffee")
        XCTAssertEqual(request.kind, .request)
        XCTAssertEqual(request.status, .pending)

        let sent = try await coord.pay(request, toDID: "did:key:bob")
        XCTAssertEqual(sent.kind, .sent)
        XCTAssertEqual(sent.status, .paid)
        XCTAssertEqual(sent.id, request.id, "same payment id links request → sent")
        XCTAssertNotNil(sent.txId)
        XCTAssertTrue(sent.txId!.hasPrefix("mocktx-"))
    }

    func testPayRejectsInvalidAmount() async {
        let bad = PaymentPayload(kind: .request, amount: 0)
        do { _ = try await coord.pay(bad, toDID: "did:key:bob"); XCTFail("expected throw") }
        catch { XCTAssertEqual(error as? WalletTransferError, .invalidAmount) }
    }
}
