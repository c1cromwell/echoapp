#if os(iOS)
import XCTest

/// Live end-to-end journeys: the real app, no stubs, against a running backend (Debug builds point
/// at http://localhost:8000). Driven by scripts/e2e-two-simulators.sh, which runs one step at a
/// time on each of two simulators and passes parameters through TEST_RUNNER_E2E_* variables.
/// Skipped unless E2E_USER is set, so the ordinary UI test run is unaffected.
final class EchoLiveE2EUITests: XCTestCase {
    private var app: XCUIApplication!
    private let env = ProcessInfo.processInfo.environment

    override func setUp() {
        super.setUp()
        continueAfterFailure = true
        app = XCUIApplication()
    }

    private func param(_ name: String) throws -> String {
        guard let value = env[name], !value.isEmpty else { throw XCTSkip("\(name) not set: live E2E not requested") }
        return value
    }

    private func snap(_ name: String) {
        let shot = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        shot.name = name
        shot.lifetime = .keepAlways
        add(shot)
    }

    /// What is on screen, as text: for the run log, so a failure can be understood without the device.
    @discardableResult
    private func describe(_ title: String) -> String {
        func labels(_ q: XCUIElementQuery, _ limit: Int = 40) -> String {
            q.allElementsBoundByIndex.prefix(limit).compactMap { e -> String? in
                guard e.exists else { return nil }
                let id = e.identifier
                var label = e.label
                if label.isEmpty, let placeholder = e.placeholderValue { label = "placeholder:" + placeholder }
                if let value = e.value as? String, !value.isEmpty, value != label { label += "=" + value }
                if id.isEmpty && label.isEmpty { return nil }
                return id.isEmpty ? label : "\(label)[\(id)]"
            }.joined(separator: " | ")
        }
        let text = """
        E2E-SCREEN \(title)
          buttons: \(labels(app.buttons))
          texts: \(labels(app.staticTexts, 30))
          fields: \(labels(app.textFields)) \(labels(app.secureTextFields)) \(labels(app.textViews))
          alerts: \(labels(app.alerts)) \(labels(app.sheets))
        """
        print(text)
        let note = XCTAttachment(string: text)
        note.name = "screen-\(title)"
        note.lifetime = .keepAlways
        add(note)
        return text
    }

    /// Taps the first of `labels` that is on screen and can be tapped. Some of the app's buttons
    /// are exposed to accessibility as text, so both are tried.
    private func tapIfPresent(_ labels: [String], timeout: TimeInterval = 1.5) -> String? {
        for label in labels {
            for element in [app.buttons[label].firstMatch, app.staticTexts[label].firstMatch] {
                if element.waitForExistence(timeout: timeout), element.isHittable {
                    element.tap()
                    return label
                }
            }
        }
        return nil
    }

    private var springboard: XCUIApplication { XCUIApplication(bundleIdentifier: "com.apple.springboard") }

    /// Answers a system permission prompt (notifications, contacts) if one is showing.
    private func dismissSystemPrompts() {
        for label in ["Allow", "OK", "Allow While Using App", "Don’t Allow"] {
            let button = springboard.buttons[label]
            if button.exists { button.tap(); return }
        }
    }

    private func buttonStarting(with prefix: String) -> XCUIElement {
        app.buttons.matching(NSPredicate(format: "label BEGINSWITH %@", prefix)).firstMatch
    }

    /// The hub's "New message" button, or the "+" on the empty screen a new account starts on.
    private var composeButton: XCUIElement {
        let hub = app.buttons["New message"]
        return hub.exists ? hub : app.buttons["Start a new conversation"]
    }

    private var onMainScreen: Bool {
        composeButton.exists || app.buttons["folders.manage"].exists
    }

    /// Waits for the Messages hub, answering system prompts on the way.
    private func waitForMain(timeout: TimeInterval) -> Bool {
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            if onMainScreen { return true }
            dismissSystemPrompts()
            sleep(1)
        }
        return onMainScreen
    }

    // MARK: - Step 1: create an account (clean install)

    func testCreateAccount() throws {
        let user = try param("E2E_USER")
        XCUIDevice.shared.orientation = .portrait
        app.launch()
        snap("01-welcome")

        let start = app.buttons["welcome.getStarted"]
        XCTAssertTrue(start.waitForExistence(timeout: 20), "welcome screen did not appear")
        start.tap()

        let field = app.textFields["username.field"]
        XCTAssertTrue(field.waitForExistence(timeout: 10), "username step did not appear")
        field.tap()
        field.typeText(user)
        XCTAssertTrue(app.staticTexts["Username is available"].waitForExistence(timeout: 15),
                      "the server did not confirm the username is available")
        snap("02-username-available")
        app.buttons["Next"].firstMatch.tap()

        XCTAssertTrue(app.buttons["Set Up"].waitForExistence(timeout: 10), "security step did not appear")
        snap("03-secure-your-account")
        app.buttons["Set Up"].tap()
        let enable = app.staticTexts["Set Up Face ID"]
        XCTAssertTrue(enable.waitForExistence(timeout: 8), "Face ID sheet did not appear")
        snap("04-face-id-sheet")
        enable.tap()
        let enrolled = app.staticTexts["Enrolled — your key is set"].waitForExistence(timeout: 40)
        if !enrolled { describe("enrollment-stuck") }
        XCTAssertTrue(enrolled, "Face ID enrollment / account registration did not complete")
        snap("05-face-id-enrolled")
        buttonStarting(with: "Continue").tap()

        let skip = buttonStarting(with: "Skip for now")
        XCTAssertTrue(skip.waitForExistence(timeout: 15), "recovery step did not appear")
        snap("06-recovery-options")
        skip.tap()

        let reached = waitForMain(timeout: 60)
        snap("07-after-first-run")
        describe("after-first-run")
        XCTAssertTrue(reached, "did not land on the Messages hub after first run")
    }

    // MARK: - Step 2 onward: open a 1:1 chat, check what arrived, send

    private func list(_ name: String) -> [String] {
        (env[name] ?? "").split(separator: "|").map(String.init).filter { !$0.isEmpty }
    }

    private var composer: XCUIElement {
        app.textFields.matching(NSPredicate(format: "placeholderValue == %@", "Message…")).firstMatch
    }

    /// Opens the 1:1 chat with `peer`: from its row on the hub if there is one, otherwise through
    /// New message → Find by Username.
    private func openChat(with peer: String, tag: String) {
        let row = app.buttons.matching(NSPredicate(format: "label CONTAINS %@", peer)).firstMatch
        if row.waitForExistence(timeout: 6), row.isHittable {
            print("E2E-OPEN via hub row: \(row.label)")
            row.tap()
            return
        }
        print("E2E-OPEN via New message (no hub row for \(peer))")
        for _ in 0..<10 where !composeButton.exists { sleep(1) }
        let newMessage = composeButton
        XCTAssertTrue(newMessage.exists, "New message button not found")
        newMessage.tap()
        let usernameField = app.textFields.matching(NSPredicate(format: "placeholderValue == %@", "Username")).firstMatch
        for _ in 0..<4 where !usernameField.exists {
            let find = app.buttons["Find by Username"]
            // Tap on the row's label, not its centre: the middle of the row is empty space that
            // does not respond to taps.
            if find.waitForExistence(timeout: 5) {
                find.coordinate(withNormalizedOffset: CGVector(dx: 0.3, dy: 0.5)).tap()
            }
            _ = usernameField.waitForExistence(timeout: 3)
        }
        snap("\(tag)-find-by-username")
        XCTAssertTrue(usernameField.exists, "Find by Username did not open the username search")
        usernameField.tap()
        usernameField.typeText(peer)
        let hit = app.staticTexts["@\(peer)"]
        let found = hit.waitForExistence(timeout: 20)
        snap("\(tag)-search-result")
        if !found { describe("\(tag)-search-no-result") }
        XCTAssertTrue(found, "searching for @\(peer) returned no result")
        hit.tap()
    }

    func testConversation() throws {
        let peer = try param("E2E_PEER")
        let tag = env["E2E_TAG"] ?? "chat"
        let patience = TimeInterval(env["E2E_WAIT"] ?? "") ?? 45
        XCUIDevice.shared.orientation = .portrait
        app.launch()
        XCTAssertTrue(waitForMain(timeout: 60), "the app did not open on the Messages hub")
        sleep(3) // let the socket connect and any waiting messages arrive
        snap("\(tag)-1-hub")

        openChat(with: peer, tag: tag)
        let ready = composer.waitForExistence(timeout: 20)
        snap("\(tag)-2-chat-opened")
        if !ready { describe("\(tag)-chat-not-open") }
        XCTAssertTrue(ready, "the chat with \(peer) did not open")

        for text in list("E2E_EXPECT") {
            let arrived = app.staticTexts[text].waitForExistence(timeout: patience)
            print("E2E-EXPECT \(arrived ? "RECEIVED" : "MISSING") \(text)")
            XCTAssertTrue(arrived, "did not receive: \(text)")
        }
        if !list("E2E_EXPECT").isEmpty {
            snap("\(tag)-3-received")
            describe("\(tag)-received")
        }

        for (index, text) in list("E2E_SEND").enumerated() {
            composer.tap()
            composer.typeText(text)
            let send = app.buttons["Send message"]
            XCTAssertTrue(send.waitForExistence(timeout: 5), "Send button not found")
            send.tap()
            let shown = app.staticTexts[text].waitForExistence(timeout: 15)
            print("E2E-SENT \(shown ? "SHOWN" : "NOT-SHOWN") \(text)")
            XCTAssertTrue(shown, "sent message is not in the thread: \(text)")
            sleep(2)
            snap("\(tag)-4-sent-\(index + 1)")
        }
        if !list("E2E_SEND").isEmpty {
            sleep(4) // delivery status settles
            snap("\(tag)-5-after-sending")
            describe("\(tag)-after-sending")
            if app.otherElements["chat.errorBanner"].exists || app.staticTexts["chat.errorBanner"].exists {
                print("E2E-ERROR-BANNER shown after sending")
            }
        }
        // Stay in the chat a while when asked, so messages sent from the other simulator arrive live.
        for text in list("E2E_EXPECT_LIVE") {
            let arrived = app.staticTexts[text].waitForExistence(timeout: patience)
            print("E2E-LIVE \(arrived ? "RECEIVED" : "MISSING") \(text)")
            XCTAssertTrue(arrived, "did not receive live: \(text)")
            snap("\(tag)-6-live-received")
        }
    }

    // MARK: - Sign out and sign back in

    // MARK: - Username taken (scripts/e2e-username-taken.sh)

    /// The account's username belongs to someone else on the server. The app must ask for
    /// another, refuse the taken one, register the new one, and not ask again.
    func testPickAnotherUsername() throws {
        let taken = try param("E2E_USER")
        let fresh = try param("E2E_NEW_NAME")
        XCUIDevice.shared.orientation = .portrait
        app.launch()

        let field = app.textFields["usernameConflict.field"]
        let deadline = Date().addingTimeInterval(90)
        while Date() < deadline, !field.exists {
            dismissSystemPrompts()
            sleep(1)
        }
        snap("taken-1-prompt")
        if !field.exists { describe("taken-no-prompt") }
        XCTAssertTrue(field.exists, "the app did not ask for another username")
        XCTAssertTrue(app.staticTexts.matching(NSPredicate(format: "label CONTAINS %@", taken)).firstMatch.exists,
                      "the prompt does not name the taken username")

        let submit = app.buttons["usernameConflict.submit"]
        field.tap()
        field.typeText(taken)
        XCTAssertTrue(app.staticTexts["That username is taken."].waitForExistence(timeout: 15),
                      "the server did not report the taken name as taken")
        XCTAssertFalse(submit.isEnabled, "a taken name can be submitted")
        snap("taken-2-still-taken")
        print("E2E-TAKEN refused \(taken)")

        field.typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: taken.count))
        field.typeText(fresh)
        XCTAssertTrue(app.staticTexts["Available"].waitForExistence(timeout: 15),
                      "the server did not report the new name as available")
        snap("taken-3-new-name-available")
        submit.tap()

        var closed = false
        for _ in 0..<20 where !closed {
            closed = !field.exists
            if !closed { sleep(1) }
        }
        snap("taken-4-registered")
        if !closed { describe("taken-not-closed") }
        XCTAssertTrue(closed, "the prompt stayed open after choosing \(fresh)")
        print("E2E-TAKEN chose \(fresh)")

        // A second launch must not ask again.
        app.terminate()
        app.launch()
        XCTAssertTrue(waitForMain(timeout: 60), "the app did not reopen on the main screen")
        sleep(8) // long enough for sign-in and the registration check
        snap("taken-5-relaunch")
        XCTAssertFalse(field.exists, "the app asked for a username again after one was registered")
        print("E2E-TAKEN no prompt on relaunch")
    }

    func testSignOutAndSignIn() throws {
        _ = try param("E2E_USER")
        XCUIDevice.shared.orientation = .portrait
        app.launch()
        XCTAssertTrue(waitForMain(timeout: 60), "the app did not open on the Messages hub")
        app.buttons["Settings"].firstMatch.tap()
        let signOut = app.buttons["Sign Out"].exists ? app.buttons["Sign Out"] : app.staticTexts["Sign Out"]
        snap("login-1-settings")
        // Sign Out is at the bottom of Settings, below the fold.
        app.swipeUp()
        app.swipeUp()
        snap("login-1a-settings-bottom")
        XCTAssertTrue(signOut.exists, "Sign Out not found in Settings")
        signOut.tap()
        let confirm = app.alerts.buttons["Sign Out"]
        XCTAssertTrue(confirm.waitForExistence(timeout: 6), "sign-out confirmation did not appear")
        snap("login-1b-confirm")
        confirm.tap()
        sleep(3)
        snap("login-2-signed-out")
        describe("login-signed-out")
        XCTAssertFalse(onMainScreen, "still on the Messages hub after signing out")
        XCTAssertTrue(app.staticTexts["Secure Login"].waitForExistence(timeout: 10) || app.buttons["Secure Login"].exists,
                      "the sign-in screen did not appear after signing out")

        var tapped: [String] = []
        for _ in 0..<5 where !onMainScreen {
            // The sign-in control is exposed as text, with its caption beneath it.
            if let label = tapIfPresent(["Secure Login", "Face ID, Touch ID, or PIN", "Use device passcode"]) {
                tapped.append(label)
            }
            dismissSystemPrompts()
            sleep(5)
        }
        print("E2E-LOGIN tapped \(tapped)")
        let back = waitForMain(timeout: 30)
        snap("login-3-signed-in")
        if !back { describe("login-not-signed-in") }
        XCTAssertTrue(back, "did not get back to the Messages hub after signing in")
    }
}
#endif
