#if os(iOS)
import XCTest
import UIKit
@testable import Echo

/// Guards against the bundled Geist Mono fonts silently un-bundling (the EchoApp target
/// had no resources build phase at all before 2026-06-24). If these fail, check that the
/// .ttf files are in the app target's Copy Bundle Resources phase and listed under
/// UIAppFonts in the product Info.plist.
final class FontResourceTests: XCTestCase {
    func testGeistMonoFamilyIsRegistered() {
        for name in ["GeistMono-Regular", "GeistMono-Medium", "GeistMono-SemiBold", "GeistMono-Bold"] {
            XCTAssertNotNil(
                UIFont(name: name, size: 12),
                "\(name) should be bundled and registered via UIAppFonts"
            )
        }
    }
}
#endif
