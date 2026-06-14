#if os(iOS)
import CallKit
import Foundation

/// CallKit integration for 1:1 voice/video calls (M4b). Reports call state to the system UI.
@MainActor
final class CallKitCoordinator: NSObject {
    static let shared = CallKitCoordinator()

    private let provider: CXProvider
    private let callController = CXCallController()
    private var activeCallUUID: UUID?
    private var onAnswer: (() -> Void)?
    private var onEnd: (() -> Void)?

    private override init() {
        let config = CXProviderConfiguration()
        config.supportsVideo = true
        config.maximumCallsPerCallGroup = 1
        config.supportedHandleTypes = [.generic]
        config.iconTemplateImageData = nil
        provider = CXProvider(configuration: config)
        super.init()
        provider.setDelegate(self, queue: nil)
    }

    func reportOutgoingCall(
        peerName: String,
        hasVideo: Bool,
        onEnd: @escaping () -> Void
    ) -> UUID {
        let uuid = UUID()
        activeCallUUID = uuid
        self.onEnd = onEnd
        let handle = CXHandle(type: .generic, value: peerName)
        let start = CXStartCallAction(call: uuid, handle: handle)
        start.isVideo = hasVideo
        callController.request(CXTransaction(action: start)) { _ in }
        provider.reportOutgoingCall(with: uuid, startedConnectingAt: Date())
        return uuid
    }

    func reportIncomingCall(
        peerName: String,
        hasVideo: Bool,
        onAnswer: @escaping () -> Void,
        onEnd: @escaping () -> Void
    ) {
        let uuid = UUID()
        activeCallUUID = uuid
        self.onAnswer = onAnswer
        self.onEnd = onEnd
        let update = CXCallUpdate()
        update.remoteHandle = CXHandle(type: .generic, value: peerName)
        update.hasVideo = hasVideo
        provider.reportNewIncomingCall(with: uuid, update: update) { _ in }
    }

    func reportConnected() {
        guard let uuid = activeCallUUID else { return }
        provider.reportOutgoingCall(with: uuid, connectedAt: Date())
    }

    func endCall() {
        guard let uuid = activeCallUUID else { return }
        let action = CXEndCallAction(call: uuid)
        callController.request(CXTransaction(action: action)) { _ in }
        activeCallUUID = nil
    }
}

extension CallKitCoordinator: CXProviderDelegate {
    nonisolated func providerDidReset(_ provider: CXProvider) {}

    nonisolated func provider(_ provider: CXProvider, perform action: CXAnswerCallAction) {
        Task { @MainActor in
            self.onAnswer?()
            action.fulfill()
        }
    }

    nonisolated func provider(_ provider: CXProvider, perform action: CXEndCallAction) {
        Task { @MainActor in
            self.onEnd?()
            self.activeCallUUID = nil
            action.fulfill()
        }
    }
}
#endif
