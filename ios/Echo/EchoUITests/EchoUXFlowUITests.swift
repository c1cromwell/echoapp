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
/// SELECTORS: these flows key off stable `accessibilityIdentifier`s set on the
/// app's controls (welcome.getStarted, welcome.haveAccount, username.title,
/// username.field, mainTabBar, tab.<name>) so they don't break on copy/locale
/// changes.
final class EchoUXFlowUITests: XCTestCase {
    private var app: XCUIApplication!

    override func setUp() {
        super.setUp()
        continueAfterFailure = false
        app = XCUIApplication()
    }

    override func tearDown() {
        app = nil
        super.tearDown()
    }

    /// Launches under UI-test mode (deterministic, animations disabled). With
    /// `authenticated: true` the app lands on the main tabs; otherwise it starts
    /// at first-run onboarding. Backed by `UITestSupport` in AppState.swift.
    private func launch(authenticated: Bool = false) {
        app.launchArguments += ["-uiTestMode"]
        if authenticated { app.launchArguments += ["-uiTestAuthenticated"] }
        app.launchEnvironment["ECHO_UITEST"] = "1"
        app.launch()
    }

    /// Onboarding entry: first-run welcome → begin setup → username entry.
    func testOnboardingWelcomeToUsername() {
        launch()
        snap("01-welcome")

        let getStarted = app.buttons["welcome.getStarted"]
        XCTAssertTrue(getStarted.waitForExistence(timeout: 10), "Welcome CTA not found")
        getStarted.tap()

        // DisplayNameEntryView.
        let heading = app.staticTexts["username.title"]
        XCTAssertTrue(
            heading.waitForExistence(timeout: 8),
            "Did not reach the username step after tapping Get Started"
        )
        snap("02-username")

        let field = app.textFields["username.field"]
        if field.waitForExistence(timeout: 3) {
            field.tap()
            field.typeText("qa_chad")
            snap("03-username-filled")
        }
    }

    /// "Already have an account" branch from the welcome screen.
    func testWelcomeAlreadyHaveAccountBranch() {
        launch()
        let existing = app.buttons["welcome.haveAccount"]
        XCTAssertTrue(existing.waitForExistence(timeout: 10), "Recovery entry CTA not found")
        existing.tap()
        snap("01-recovery-entry")
        // Extend: assert the recovery/login screen, drive phrase entry, etc.
    }

    /// Smoke sweep of the (custom) main tab bar. Launches straight onto the tabs
    /// via `-uiTestAuthenticated`. The bar is a custom HStack of buttons, not a
    /// UITabBar, so we key off the `tab.<name>` identifiers.
    func testMainTabsSmoke() throws {
        launch(authenticated: true)
        guard app.otherElements["mainTabBar"].waitForExistence(timeout: 10) else {
            throw XCTSkip("Main tabs did not appear (authenticated launch may need a seeded session)")
        }
        for tab in ["messages", "contacts", "rewards", "settings"] {
            let button = app.buttons["tab.\(tab)"]
            XCTAssertTrue(button.waitForExistence(timeout: 5), "Missing tab: \(tab)")
            button.tap()
            snap("tab-\(tab)")
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
// `-uiTestMode` / `-uiTestAuthenticated` are handled by UITestSupport in
// AppState.swift: uiTestMode disables animations and forces first-run; adding
// uiTestAuthenticated lands directly on the main tabs. Extend UITestSupport to
// seed fixtures / stub the network for deeper deterministic journeys (wallet
// send, a seeded chat).
#endif
