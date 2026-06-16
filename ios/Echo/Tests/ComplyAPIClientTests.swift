import Foundation
import XCTest
@testable import Echo

#if os(iOS)
final class ComplyAPIClientTests: XCTestCase {
    func testDashboardDecoding() throws {
        let json = """
        {"deCoverageRate":"12%","activeRetentionPolicies":2,"litigationHolds":1,"pendingExports":0,"anchorHealth":"healthy"}
        """
        let decoded = try JSONDecoder().decode(ComplyDashboardSummary.self, from: Data(json.utf8))
        XCTAssertEqual(decoded.activeRetentionPolicies, 2)
        XCTAssertEqual(decoded.litigationHolds, 1)
    }

    func testComplyEndpointOrgHeader() {
        let endpoint = ComplyEndpoint.dashboard(orgDID: "did:org:acme")
        XCTAssertEqual(endpoint.headers["X-Org-DID"], "did:org:acme")
        XCTAssertEqual(endpoint.path, "/comply/dashboard")
    }

    func testAuditPDFEndpoint() {
        let endpoint = ComplyEndpoint.auditPDF(orgDID: "did:org:acme")
        XCTAssertEqual(endpoint.path, "/comply/audit/report?format=pdf")
        XCTAssertEqual(endpoint.headers["X-Org-DID"], "did:org:acme")
    }
}
#endif
