#if os(iOS)
import XCTest

/// First automated end-to-end UX flows, driven through the live app in the
/// simulator. Pair with `scripts/qa-record.sh`, which screen-records the whole
/// run and pulls these screenshots out of the .xcresult bundle.
///
/// TARGET: this file belongs to a **UI Testing Bundle** target (e.g.
/// `EchoUITests`), NOT the app or unit-test target. XCUITests run out-of-process
/// and attach to the app, so they can't `@testable import Echo` — they see only
/// the on-screen UI.
///
/// SELECTOR NOTE: the app currently exposes no `accessibilityIdentifier`s, so
/// these flows key off visible text, which is brittle (breaks on copy/locale
/// changes). Hardening step: add `.accessibilityIdentifier("…")` to the key
/// controls and switch the queries below to `app.buttons["id"]`.
final class EchoUXFlowUITests: XCTestCase {
    private var app: XCUIApplication!

    override func setUp() {
        super.setUp()
        continueAfterFailure = false
        app = XCUIApplication()
        // Deterministic QA state. Wire these up in AppState to seed fixtures /
        // skip animations for stable, fast runs (see NOTES at bottom).
        app.launchArguments += ["-uiTestMode", "1"]
        app.launchEnvironment["ECHO_UITEST"] = "1"
        app.launch()
    }

    override func tearDown() {
        app = nil
        super.tearDown()
    }

    /// Onboarding entry: first-run welcome → begin setup → username entry.
    func testOnboardingWelcomeToUsername() {
        snap("01-welcome")

        // Welcome screen (EchoWelcomeView): primary CTA is "Let's Go".
        let letsGo = app.buttons["Let's Go"]
        XCTAssertTrue(letsGo.waitForExistence(timeout: 10), "Welcome CTA not found")
        letsGo.tap()

        // DisplayNameEntryView.
        let usernameHeading = app.staticTexts["Choose username"]
        XCTAssertTrue(
            usernameHeading.waitForExistence(timeout: 8),
            "Did not reach the username step after tapping Let's Go"
        )
        snap("02-username")

        // Type a username if a field is present (identifier TBD — add one).
        let field = app.textFields.firstMatch
        if field.waitForExistence(timeout: 3) {
            field.tap()
            field.typeText("qa_chad")
            snap("03-username-filled")
        }
    }

    /// "Already have an account" branch from the welcome screen.
    func testWelcomeAlreadyHaveAccountBranch() {
        let existing = app.buttons["I already have an account"]
        XCTAssertTrue(existing.waitForExistence(timeout: 10), "Recovery entry CTA not found")
        existing.tap()
        snap("01-recovery-entry")
        // Extend: assert the recovery/login screen, drive phrase entry, etc.
    }

    /// Smoke sweep of the main tab bar once past onboarding. Skips cleanly until
    /// -uiTestMode lands the app straight on the main tabs.
    func testMainTabsSmoke() throws {
        let tabBar = app.tabBars.firstMatch
        guard tabBar.waitForExistence(timeout: 8) else {
            throw XCTSkip("Not on main tabs — enable -uiTestMode to seed a signed-in session")
        }
        for button in tabBar.buttons.allElementsBoundByIndex {
            button.tap()
            snap("tab-\(button.label.replacingOccurrences(of: " ", with: "-").lowercased())")
        }
    }

    // MARK: - Helpers

    /// Screenshot attached to the test, kept in the .xcresult and exported by
    /// scripts/qa-record.sh.
    private func snap(_ name: String) {
        let shot = XCTAttachment(screenshot: app.screenshot())
        shot.name = name
        shot.lifetime = .keepAlways
        add(shot)
    }
}

// ---------------------------------------------------------------- NOTES -------
// Deterministic launch state: read the flag early in app startup, e.g. in
// AppState / EchoApp:
//
//   if ProcessInfo.processInfo.arguments.contains("-uiTestMode") {
//       // seed fixtures, disable network, skip to a known screen,
//       UIView.setAnimationsEnabled(false)
//   }
//
// Without it these flows only cover the first-run path; with it you can land
// directly on any journey (main tabs, wallet, a seeded chat) for stable QA.
#endif
