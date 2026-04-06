import XCTest
@testable import Echo

// MARK: - App Coordinator Tests

@MainActor
final class AppCoordinatorTests: XCTestCase {

    func testInitialState() {
        let coordinator = AppCoordinator()
        XCTAssertTrue(coordinator.navigationPath.isEmpty)
        XCTAssertFalse(coordinator.isAuthenticated)
    }

    func testNavigateReplacesStack() {
        let coordinator = AppCoordinator()
        coordinator.navigate(to: .main)
        XCTAssertEqual(coordinator.navigationPath.count, 1)

        coordinator.navigate(to: .settings)
        // navigate() calls removeAll then append — stack should have exactly 1 item
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testPopBackRemovesLast() {
        let coordinator = AppCoordinator()
        coordinator.navigationPath = [.main, .settings]
        coordinator.popBack()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testPopBackOnEmptyDoesNotCrash() {
        let coordinator = AppCoordinator()
        coordinator.popBack() // should not crash
        XCTAssertTrue(coordinator.navigationPath.isEmpty)
    }

    func testPopToRootClearsStack() {
        let coordinator = AppCoordinator()
        coordinator.navigationPath = [.main, .chat(conversationId: "123"), .profile(userId: "abc")]
        coordinator.popToRoot()
        XCTAssertTrue(coordinator.navigationPath.isEmpty)
    }
}

// MARK: - Auth Coordinator Tests

@MainActor
final class AuthCoordinatorTests: XCTestCase {

    func testStartSetsAuthentication() {
        let coordinator = AuthCoordinator()
        coordinator.start()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testNavigateAppendsToStack() {
        let coordinator = AuthCoordinator()
        coordinator.start()
        coordinator.navigate(to: .onboarding)
        XCTAssertEqual(coordinator.navigationPath.count, 2)
    }

    func testPopToRootResetsToAuthentication() {
        let coordinator = AuthCoordinator()
        coordinator.start()
        coordinator.navigate(to: .onboarding)
        coordinator.popToRoot()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testCompleteAuthenticationCallsCallback() {
        let coordinator = AuthCoordinator()
        var callbackCalled = false
        coordinator.onAuthenticationComplete = { callbackCalled = true }
        coordinator.completeAuthentication()
        XCTAssertTrue(callbackCalled)
    }

    func testShowOnboardingReplacesPath() {
        let coordinator = AuthCoordinator()
        coordinator.start()
        coordinator.showOnboarding()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }
}

// MARK: - Main Coordinator Tests

@MainActor
final class MainCoordinatorTests: XCTestCase {

    func testStartSetsMain() {
        let coordinator = MainCoordinator()
        coordinator.start()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testDefaultTabIsConversations() {
        let coordinator = MainCoordinator()
        XCTAssertEqual(coordinator.selectedTab, .conversations)
    }

    func testSelectTabChangesTabAndResetsStack() {
        let coordinator = MainCoordinator()
        coordinator.start()
        coordinator.navigate(to: .chat(conversationId: "123"))
        XCTAssertEqual(coordinator.navigationPath.count, 2)

        coordinator.selectTab(.contacts)
        XCTAssertEqual(coordinator.selectedTab, .contacts)
        // selectTab calls popToRoot which sets stack to [.main]
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testOpenChatAppends() {
        let coordinator = MainCoordinator()
        coordinator.start()
        coordinator.openChat(conversationId: "conv_42")
        XCTAssertEqual(coordinator.navigationPath.count, 2)
    }

    func testOpenProfileAppends() {
        let coordinator = MainCoordinator()
        coordinator.start()
        coordinator.openProfile(userId: "user_99")
        XCTAssertEqual(coordinator.navigationPath.count, 2)
    }

    func testOpenWalletAppends() {
        let coordinator = MainCoordinator()
        coordinator.start()
        coordinator.openWallet()
        XCTAssertEqual(coordinator.navigationPath.count, 2)
    }

    func testOpenSettingsAppends() {
        let coordinator = MainCoordinator()
        coordinator.start()
        coordinator.openSettings()
        XCTAssertEqual(coordinator.navigationPath.count, 2)
    }

    func testAllMainTabs() {
        let tabs = MainCoordinator.MainTab.allCases
        XCTAssertEqual(tabs.count, 4)
        XCTAssertTrue(tabs.contains(.conversations))
        XCTAssertTrue(tabs.contains(.contacts))
        XCTAssertTrue(tabs.contains(.profile))
        XCTAssertTrue(tabs.contains(.wallet))
    }
}

// MARK: - Settings Coordinator Tests

@MainActor
final class SettingsCoordinatorTests: XCTestCase {

    func testStartSetsSettings() {
        let coordinator = SettingsCoordinator()
        coordinator.start()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testPopToRootResetsToSettings() {
        let coordinator = SettingsCoordinator()
        coordinator.start()
        coordinator.navigate(to: .profile(userId: "me"))
        coordinator.popToRoot()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }
}

// MARK: - Chat Coordinator Tests

@MainActor
final class ChatCoordinatorTests: XCTestCase {

    func testInitWithConversationId() {
        let coordinator = ChatCoordinator(conversationId: "conv_42")
        XCTAssertEqual(coordinator.conversationId, "conv_42")
    }

    func testStartSetsChat() {
        let coordinator = ChatCoordinator(conversationId: "conv_42")
        coordinator.start()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testPopToRootResetsToChat() {
        let coordinator = ChatCoordinator(conversationId: "conv_42")
        coordinator.start()
        coordinator.navigate(to: .profile(userId: "contact_1"))
        XCTAssertEqual(coordinator.navigationPath.count, 2)
        coordinator.popToRoot()
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }
}

// MARK: - Route Manager Tests

@MainActor
final class RouteManagerTests: XCTestCase {

    func testPushAddsToStack() {
        let manager = RouteManager()
        manager.push(.main)
        manager.push(.settings)
        XCTAssertEqual(manager.navigationStack.count, 2)
    }

    func testPopRemovesLast() {
        let manager = RouteManager()
        manager.push(.main)
        manager.push(.settings)
        manager.pop()
        XCTAssertEqual(manager.navigationStack.count, 1)
    }

    func testPopOnEmptyDoesNotCrash() {
        let manager = RouteManager()
        manager.pop() // should not crash
        XCTAssertTrue(manager.navigationStack.isEmpty)
    }

    func testPopToRootClearsStack() {
        let manager = RouteManager()
        manager.push(.main)
        manager.push(.settings)
        manager.push(.contacts)
        manager.popToRoot()
        XCTAssertTrue(manager.navigationStack.isEmpty)
    }

    func testReplaceReplacesEntireStack() {
        let manager = RouteManager()
        manager.push(.main)
        manager.push(.settings)
        manager.replace(.authentication)
        XCTAssertEqual(manager.navigationStack.count, 1)
    }
}

// MARK: - Navigation Stack Tests

@MainActor
final class NavigationStackTests: XCTestCase {

    func testPushAndCurrent() {
        let stack = NavigationStack()
        XCTAssertNil(stack.current)
        XCTAssertFalse(stack.canPop)

        stack.push(.main)
        XCTAssertNotNil(stack.current)
        XCTAssertTrue(stack.canPop)
    }

    func testPopReturnsLast() {
        let stack = NavigationStack()
        stack.push(.main)
        stack.push(.settings)

        let popped = stack.pop()
        XCTAssertNotNil(popped)
        XCTAssertEqual(stack.stack.count, 1)
    }

    func testPopAllClearsStack() {
        let stack = NavigationStack()
        stack.push(.main)
        stack.push(.settings)
        stack.push(.contacts)
        stack.popAll()
        XCTAssertTrue(stack.stack.isEmpty)
        XCTAssertFalse(stack.canPop)
    }
}

// MARK: - Deep Link Handler Tests

@MainActor
final class DeepLinkHandlerTests: XCTestCase {

    func testHandleChatDeepLink() {
        let coordinator = AppCoordinator()
        let handler = DeepLinkHandler(appCoordinator: coordinator)

        let url = URL(string: "echo://chat/conv_123")!
        handler.handle(url)

        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testHandleProfileDeepLink() {
        let coordinator = AppCoordinator()
        let handler = DeepLinkHandler(appCoordinator: coordinator)

        let url = URL(string: "echo://profile/user_456")!
        handler.handle(url)

        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testHandleWalletDeepLink() {
        let coordinator = AppCoordinator()
        let handler = DeepLinkHandler(appCoordinator: coordinator)

        let url = URL(string: "echo://wallet")!
        handler.handle(url)

        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testHandleUnknownDeepLinkDoesNothing() {
        let coordinator = AppCoordinator()
        let handler = DeepLinkHandler(appCoordinator: coordinator)

        let url = URL(string: "echo://unknown/path")!
        handler.handle(url)

        XCTAssertTrue(coordinator.navigationPath.isEmpty)
    }

    func testHandleInvalidURLDoesNotCrash() {
        let coordinator = AppCoordinator()
        let handler = DeepLinkHandler(appCoordinator: coordinator)

        let url = URL(string: "not-echo://something")!
        handler.handle(url)
        // Should not crash, just does nothing
    }
}
