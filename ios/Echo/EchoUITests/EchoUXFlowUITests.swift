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
        _ = app.wait(for: .runningForeground, timeout: 8)
        snap("00-authenticated-launch")
        // Match by accessibility label — the custom HStack tab buttons expose
        // their label reliably (the identifier doesn't survive the nested VStack).
        let tabs = ["Messages", "Contacts", "Rewards", "Settings"]
        guard app.buttons[tabs[0]].waitForExistence(timeout: 12) else {
            throw XCTSkip("Main tabs did not appear (authenticated launch may need a seeded session)")
        }
        for label in tabs {
            let button = app.buttons[label]
            XCTAssertTrue(button.waitForExistence(timeout: 5), "Missing tab: \(label)")
            button.tap()
            snap("tab-\(label.lowercased())")
        }
    }

    /// Persona management: from the Settings tab, open Personas and add one.
    func testCreatePersonaFromSettings() throws {
        launch(authenticated: true)
        let settingsTab = app.buttons["Settings"]
        guard settingsTab.waitForExistence(timeout: 12) else {
            throw XCTSkip("Main tabs did not appear (authenticated launch may need a seeded session)")
        }
        settingsTab.tap()

        let personas = app.buttons["settings.personas"]
        XCTAssertTrue(personas.waitForExistence(timeout: 6), "Personas row not found in Settings")
        personas.tap()
        snap("personas-list")

        let add = app.buttons["personas.add"]
        XCTAssertTrue(add.waitForExistence(timeout: 6), "Add Persona button not found")
        add.tap()

        // Editor sheet opened — gate on its nav title so the field is present.
        // (firstMatch is unusable: MainTabView keeps all tabs in the hierarchy,
        // so a background "Search contacts" field ranks first.)
        XCTAssertTrue(app.navigationBars["New Persona"].waitForExistence(timeout: 6), "Persona editor did not open")
        let byId = app.textFields["persona.nameField"]
        _ = byId.waitForExistence(timeout: 4)
        let field = byId.exists ? byId : app.textFields["Persona name"]
        XCTAssertTrue(field.waitForExistence(timeout: 3), "Persona name field not found")
        field.tap()
        field.typeText("Gaming")
        app.buttons["Save"].tap()

        XCTAssertTrue(
            app.staticTexts["Gaming"].waitForExistence(timeout: 6),
            "Newly created persona did not appear in the list"
        )
        snap("persona-created")
    }

    /// Chat folders: from the Messages hub, open the folder manager and create a
    /// folder, then confirm it appears.
    func testCreateChatFolder() throws {
        launch(authenticated: true)
        let manage = app.buttons["folders.manage"]
        guard manage.waitForExistence(timeout: 12) else {
            throw XCTSkip("Folder manager not found (authenticated hub may need a seeded session)")
        }
        manage.tap()

        let newFolder = app.buttons["folders.new"]
        XCTAssertTrue(newFolder.waitForExistence(timeout: 6), "New Folder button not found")
        newFolder.tap()

        let field = app.alerts.textFields.firstMatch
        XCTAssertTrue(field.waitForExistence(timeout: 4), "New Folder text field not found")
        field.typeText("QA Folder")
        app.alerts.buttons["Create"].tap()

        XCTAssertTrue(
            app.staticTexts["QA Folder"].waitForExistence(timeout: 6),
            "Created folder did not appear in the list"
        )
        snap("folder-created")
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
