#if os(iOS)
import XCTest
@testable import Echo

final class ContactPersonaStoreTests: XCTestCase {
    func testMinimalDisclosureHidden() {
        let did = "did:key:zTestPersona"
        ContactPersonaStore.setPersonaId("hidden", for: did)
        XCTAssertEqual(ContactPersonaStore.personaId(for: did), "hidden")
        XCTAssertEqual(ContactPersonaStore.minimalDisplayName(base: "Alice", personaId: "hidden"), "ECHO user")
        XCTAssertTrue(ContactPersonaStore.shouldHideAvatar(personaId: "hidden"))
        ContactPersonaStore.setPersonaId(nil, for: did)
    }
}
#endif
