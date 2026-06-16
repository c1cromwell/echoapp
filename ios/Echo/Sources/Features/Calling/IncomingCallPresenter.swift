#if os(iOS)
import Foundation
import SwiftUI

/// Presents incoming 1:1 calls from global WS `call_signal` offers (M4c).
@MainActor
final class IncomingCallPresenter: ObservableObject {
    static let shared = IncomingCallPresenter()

    struct IncomingCall: Identifiable, Equatable {
        let id: String
        let callId: String
        let peerDID: String
        let callType: CallType
        let offerSDP: String
    }

    @Published private(set) var activeCall: IncomingCall?

    private var signaling: CallSignalingService?
    private var isConfigured = false

    private init() {}

    func configureIfNeeded() async {
        guard !isConfigured,
              let signalService = DIContainer.shared.resolveConversationSignalService(),
              let apiClient = DIContainer.shared.resolveAPIClient() else { return }
        isConfigured = true
        let localDID = await CurrentUserSession.currentDID() ?? ""
        let service = CallSignalingService(
            signalService: signalService,
            iceAPI: CallICEAPIClient(apiClient: apiClient)
        )
        service.configure(localDID: localDID) { [weak self] event in
            Task { @MainActor in self?.handle(event) }
        }
        signaling = service
    }

    func acceptActiveCall() {
        guard activeCall != nil else { return }
    }

    func rejectActiveCall() async {
        guard let call = activeCall,
              let signaling else {
            activeCall = nil
            return
        }
        try? await signaling.sendReject(callId: call.callId, to: call.peerDID)
        CallKitCoordinator.shared.endCall()
        activeCall = nil
    }

    func clearAfterDismiss() {
        activeCall = nil
    }

    private func handle(_ event: CallSignalEvent) {
        switch event.action {
        case .offer:
            guard let sdp = event.sdp,
                  let callType = event.callType,
                  activeCall == nil else { return }
            let peerName = ContactThreadHelper.truncatedDID(event.peerDID)
            activeCall = IncomingCall(
                id: event.callId,
                callId: event.callId,
                peerDID: event.peerDID,
                callType: callType,
                offerSDP: sdp
            )
            CallKitCoordinator.shared.reportIncomingCall(
                peerName: peerName,
                hasVideo: callType == .video,
                onAnswer: {},
                onEnd: { [weak self] in
                    Task { await self?.rejectActiveCall() }
                }
            )
        case .hangup, .reject:
            if event.callId == activeCall?.callId {
                activeCall = nil
                CallKitCoordinator.shared.endCall()
            }
        default:
            break
        }
    }
}
#endif
