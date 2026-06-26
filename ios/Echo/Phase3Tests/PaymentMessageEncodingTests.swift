import XCTest
@testable import Echo

final class PaymentMessageEncodingTests: XCTestCase {
    func testEncodeDecodeRoundTrip() {
        let p = PaymentPayload(kind: .request, amount: Decimal(string: "7.25")!, memo: "tacos")
        let wire = PaymentMessageEncoding.encode(p)
        XCTAssertTrue(PaymentMessageEncoding.isPayment(wire))
        XCTAssertEqual(PaymentMessageEncoding.decode(wire), p)
    }

    func testPlainTextIsNotAPayment() {
        XCTAssertFalse(PaymentMessageEncoding.isPayment("hello there"))
        XCTAssertNil(PaymentMessageEncoding.decode("hello there"))
    }

    func testSentPaymentRoundTrip() {
        let p = PaymentPayload(kind: .sent, amount: 100, memo: nil, txId: "mocktx-abc", status: .paid)
        XCTAssertEqual(PaymentMessageEncoding.decode(PaymentMessageEncoding.encode(p)), p)
    }
}
