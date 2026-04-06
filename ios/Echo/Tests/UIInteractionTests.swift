import XCTest
@testable import Echo

// MARK: - Welcome View Interaction Tests

final class WelcomeViewInteractionTests: XCTestCase {

    func testWelcomeViewCallsPhoneCallback() {
        var phoneTapped = false
        let view = WelcomeView(
            onContinueWithPhone: { phoneTapped = true },
            onUseVerifiableID: {}
        )
        XCTAssertNotNil(view)
        // Verify the callback is wired (invoke directly)
        view.onContinueWithPhone()
        XCTAssertTrue(phoneTapped)
    }

    func testWelcomeViewCallsVerifiableIDCallback() {
        var idTapped = false
        let view = WelcomeView(
            onContinueWithPhone: {},
            onUseVerifiableID: { idTapped = true }
        )
        view.onUseVerifiableID()
        XCTAssertTrue(idTapped)
    }

    func testWelcomeViewDefaultCallbacksDoNotCrash() {
        let view = WelcomeView()
        view.onContinueWithPhone()
        view.onUseVerifiableID()
        // No crash = pass
    }
}

// MARK: - Phone Entry View Tests

final class PhoneEntryViewInteractionTests: XCTestCase {

    func testPhoneValidation_TooShort() {
        let phone = "12345"
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertFalse(isValid)
    }

    func testPhoneValidation_ExactlyTenDigits() {
        let phone = "5551234567"
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertTrue(isValid)
    }

    func testPhoneValidation_ContainsLetters() {
        let phone = "555123abcd"
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertFalse(isValid)
    }

    func testPhoneValidation_ContainsSpaces() {
        let phone = "555 123 4567"
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertFalse(isValid)
    }

    func testPhoneValidation_LongerThanTen() {
        let phone = "15551234567"
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertTrue(isValid)
    }

    func testPhoneValidation_Empty() {
        let phone = ""
        let isValid = phone.count >= 10 && phone.allSatisfy(\.isNumber)
        XCTAssertFalse(isValid)
    }

    func testPhoneEntryCallbackReceivesPhone() {
        var receivedPhone: String?
        let view = PhoneEntryView(onSendCode: { phone in
            receivedPhone = phone
        })
        view.onSendCode("5551234567")
        XCTAssertEqual(receivedPhone, "5551234567")
    }
}

// MARK: - OTP Verification View Tests

final class OTPVerificationViewTests: XCTestCase {

    func testOTPViewDisplaysPhoneNumber() {
        let view = OTPVerificationView(phoneNumber: "+1 (555) 123-4567")
        XCTAssertEqual(view.phoneNumber, "+1 (555) 123-4567")
    }

    func testOTPViewCallsVerifyCallback() {
        var verifiedCode: String?
        let view = OTPVerificationView(
            phoneNumber: "+1 (555) 123-4567",
            onVerify: { code in verifiedCode = code }
        )
        view.onVerify("123456")
        XCTAssertEqual(verifiedCode, "123456")
    }

    func testOTPViewCallsResendCallback() {
        var resendCalled = false
        let view = OTPVerificationView(
            phoneNumber: "+1 (555) 123-4567",
            onResendCode: { resendCalled = true }
        )
        view.onResendCode()
        XCTAssertTrue(resendCalled)
    }
}

// MARK: - Conversation List View Tests

final class ConversationListViewInteractionTests: XCTestCase {

    func testConversationListSelectCallback() {
        var selectedId: String?
        let view = ConversationListView(
            onSelectConversation: { id in selectedId = id }
        )
        view.onSelectConversation("conv_42")
        XCTAssertEqual(selectedId, "conv_42")
    }

    func testConversationListPinnedItemCallback() {
        var tappedItem: PinnedItem?
        let view = ConversationListView(
            onPinnedItemTap: { item in tappedItem = item }
        )
        let item = PinnedItem(
            id: "1", type: .contact, name: "Alice", initials: "AL", gradientIndex: 0
        )
        view.onPinnedItemTap(item)
        XCTAssertEqual(tappedItem?.id, "1")
        XCTAssertEqual(tappedItem?.name, "Alice")
    }

    func testConversationListEditPinnedCallback() {
        var editCalled = false
        let view = ConversationListView(
            onEditPinned: { editCalled = true }
        )
        view.onEditPinned()
        XCTAssertTrue(editCalled)
    }

    func testConversationListWithPinnedItems() {
        let pinnedItems = [
            PinnedItem(id: "1", type: .contact, name: "Alice", initials: "AL", gradientIndex: 0, isOnline: true),
            PinnedItem(id: "2", type: .group, name: "Team", initials: "TM", gradientIndex: 1)
        ]
        let view = ConversationListView(pinnedItems: pinnedItems)
        XCTAssertNotNil(view)
    }

    func testConversationSearchFiltering() {
        // Test the filtering logic directly
        let conversations = [
            ConversationItem(id: "1", contactName: "John Doe", lastMessage: "Hi", timestamp: "Now", unreadCount: 0, isOnline: true),
            ConversationItem(id: "2", contactName: "Jane Smith", lastMessage: "Hello", timestamp: "Now", unreadCount: 0, isOnline: false),
            ConversationItem(id: "3", contactName: "Alice Johnson", lastMessage: "Hey", timestamp: "Now", unreadCount: 0, isOnline: true)
        ]

        let searchText = "john"
        let filtered = conversations.filter { $0.contactName.localizedCaseInsensitiveContains(searchText) }
        XCTAssertEqual(filtered.count, 2) // "John Doe" and "Alice Johnson"
    }

    func testConversationSortingUnreadFirst() {
        let conversations = [
            ConversationItem(id: "1", contactName: "No Unread", lastMessage: "Hi", timestamp: "Now", unreadCount: 0, isOnline: true),
            ConversationItem(id: "2", contactName: "Has Unread", lastMessage: "Hello", timestamp: "Now", unreadCount: 3, isOnline: false)
        ]

        let sorted = conversations.sorted { ($0.unreadCount > 0) && ($1.unreadCount == 0) }
        XCTAssertEqual(sorted.first?.contactName, "Has Unread")
    }
}

// MARK: - Chat View Tests

final class ChatViewInteractionTests: XCTestCase {

    func testChatViewInitWithContactName() {
        let view = ChatView(contactName: "John Doe")
        XCTAssertEqual(view.contactName, "John Doe")
    }

    func testChatViewSendMessageCallback() {
        var sentMessage: String?
        let view = ChatView(
            contactName: "John Doe",
            onSendMessage: { msg in sentMessage = msg }
        )
        view.onSendMessage("Hello!")
        XCTAssertEqual(sentMessage, "Hello!")
    }

    func testChatViewDefaultCallbackDoesNotCrash() {
        let view = ChatView(contactName: "Test")
        view.onSendMessage("test message")
        // No crash = pass
    }
}

// MARK: - Conversation Item Model Tests

final class ConversationItemTests: XCTestCase {

    func testConversationItemCreation() {
        let item = ConversationItem(
            id: "conv_1",
            contactName: "John",
            lastMessage: "Hey there!",
            timestamp: "2:30 PM",
            unreadCount: 5,
            isOnline: true
        )
        XCTAssertEqual(item.id, "conv_1")
        XCTAssertEqual(item.contactName, "John")
        XCTAssertEqual(item.lastMessage, "Hey there!")
        XCTAssertEqual(item.unreadCount, 5)
        XCTAssertTrue(item.isOnline)
    }
}

// MARK: - Chat Message Model Tests

final class ChatMessageTests: XCTestCase {

    func testChatMessageCreation() {
        let message = ChatMessage(
            id: "msg_1",
            content: "Hello!",
            isSent: true,
            status: .sent,
            timestamp: "10:30 AM"
        )
        XCTAssertEqual(message.id, "msg_1")
        XCTAssertEqual(message.content, "Hello!")
        XCTAssertTrue(message.isSent)
    }

    func testReceivedMessageCreation() {
        let message = ChatMessage(
            id: "msg_2",
            content: "Hi back!",
            isSent: false,
            status: .read,
            timestamp: "10:31 AM"
        )
        XCTAssertFalse(message.isSent)
    }
}

// MARK: - Auth Flow Integration Tests

@MainActor
final class AuthFlowIntegrationTests: XCTestCase {

    func testOnboardingFlowSequence() {
        // Simulate the full auth flow through coordinators
        let authCoordinator = AuthCoordinator()
        authCoordinator.start()
        XCTAssertEqual(authCoordinator.navigationPath.count, 1, "Should start at authentication")

        // User enters phone → navigate to onboarding
        authCoordinator.showOnboarding()
        XCTAssertEqual(authCoordinator.navigationPath.count, 1, "Should be at onboarding")

        // User completes authentication
        var authCompleted = false
        authCoordinator.onAuthenticationComplete = { authCompleted = true }
        authCoordinator.completeAuthentication()
        XCTAssertTrue(authCompleted, "Auth completion callback should fire")
    }

    func testMainFlowNavigation() {
        // Simulate navigating through the main app
        let coordinator = MainCoordinator()
        coordinator.start()

        // Open a chat
        coordinator.openChat(conversationId: "conv_1")
        XCTAssertEqual(coordinator.navigationPath.count, 2)

        // Go back
        coordinator.popBack()
        XCTAssertEqual(coordinator.navigationPath.count, 1)

        // Switch to contacts tab
        coordinator.selectTab(.contacts)
        XCTAssertEqual(coordinator.selectedTab, .contacts)
        XCTAssertEqual(coordinator.navigationPath.count, 1) // popToRoot

        // Open a profile from contacts
        coordinator.openProfile(userId: "user_42")
        XCTAssertEqual(coordinator.navigationPath.count, 2)

        // Switch tab should reset
        coordinator.selectTab(.wallet)
        XCTAssertEqual(coordinator.selectedTab, .wallet)
        XCTAssertEqual(coordinator.navigationPath.count, 1)
    }

    func testDeepLinkIntoChatWhileAuthenticated() {
        let appCoordinator = AppCoordinator()
        appCoordinator.isAuthenticated = true
        appCoordinator.navigate(to: .main)

        // Deep link arrives
        let handler = DeepLinkHandler(appCoordinator: appCoordinator)
        handler.handle(URL(string: "echo://chat/conv_deep_123")!)

        // Should navigate (replacing the stack)
        XCTAssertEqual(appCoordinator.navigationPath.count, 1)
    }
}
