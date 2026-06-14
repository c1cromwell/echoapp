// Features/Calling/CallViewModel.swift
// Manages voice/video call state and WebRTC signaling

import Foundation
import SwiftUI
import Combine

@MainActor
class CallViewModel: ObservableObject {
    let peerDID: String
    let callType: CallType
    let isOutgoing: Bool

    @Published var state: CallState = .idle
    @Published var stateLabel: String = "Calling..."
    @Published var contactName: String = ""
    @Published var contactAvatar: URL?
    @Published var contactTrustTier: TrustTier = .newcomer
    @Published var duration: TimeInterval = 0

    @Published var isMuted = false
    @Published var isSpeaker = false
    @Published var isCameraOn = true
    @Published var isScreenSharing = false
    @Published var localStream: Any?
    @Published var remoteStream: Any?

    private let callId = UUID().uuidString
    private var durationTimer: Timer?
    private let session = WebRTCCallSession()
    private var signaling: CallSignalingService?
    private var localDID: String = ""
    private var startedAt = Date()

    init(peerDID: String, callType: CallType, contactName: String = "", isOutgoing: Bool = true) {
        self.peerDID = peerDID
        self.callType = callType
        self.contactName = contactName
        self.isOutgoing = isOutgoing
        self.isCameraOn = callType == .video
    }

    func startCall() async {
        state = .connecting
        stateLabel = isOutgoing ? "Calling..." : "Connecting..."
        startedAt = Date()
        await loadContactInfo()
        await configureSignaling()

        do {
            _ = try await signaling?.fetchICEServers()
            if isOutgoing {
                let sdp = try await session.createOffer(callType: callType)
                try await signaling?.sendOffer(
                    callId: callId,
                    to: peerDID,
                    callType: callType,
                    sdp: sdp
                )
            } else {
                state = .ringing
                stateLabel = "Incoming call..."
            }
        } catch {
            state = .ended
            stateLabel = "Call failed"
            recordCall(missed: true)
        }
    }

    func endCall() {
        durationTimer?.invalidate()
        durationTimer = nil
        session.hangup()
        state = .ended
        stateLabel = "Call Ended"
        recordCall(missed: false)
        Task {
            try? await signaling?.sendHangup(callId: callId, to: peerDID)
        }
    }

    func toggleMute() { isMuted.toggle() }
    func toggleSpeaker() { isSpeaker.toggle() }
    func toggleCamera() { isCameraOn.toggle() }
    func flipCamera() {}
    func startScreenShare() { isScreenSharing = true }
    func stopScreenShare() { isScreenSharing = false }

    private func configureSignaling() async {
        guard let signalService = DIContainer.shared.resolveConversationSignalService(),
              let apiClient = DIContainer.shared.resolveAPIClient() else { return }
        localDID = await CurrentUserSession.currentDID() ?? ""
        let service = CallSignalingService(
            signalService: signalService,
            iceAPI: CallICEAPIClient(apiClient: apiClient)
        )
        service.configure(localDID: localDID) { [weak self] event in
            Task { @MainActor in self?.handleCallSignal(event) }
        }
        signaling = service
    }

    private func handleCallSignal(_ event: CallSignalEvent) {
        guard event.callId == callId || isOutgoing == false else { return }
        switch event.action {
        case .offer:
            guard !isOutgoing, let sdp = event.sdp else { return }
            Task {
                do {
                    let answer = try await session.applyRemoteOffer(sdp, callType: callType)
                    try await signaling?.sendAnswer(callId: event.callId, to: event.peerDID, sdp: answer)
                    await session.applyAnswer(answer)
                    await markActive()
                } catch {
                    state = .ended
                    stateLabel = "Call failed"
                }
            }
        case .answer:
            guard let sdp = event.sdp else { return }
            Task {
                await session.applyAnswer(sdp)
                await markActive()
            }
        case .ice:
            if let ice = event.iceCandidate {
                Task { await session.addIceCandidate(ice) }
            }
        case .hangup, .reject:
            state = .ended
            stateLabel = event.action == .reject ? "Declined" : "Call Ended"
            recordCall(missed: event.action == .reject)
        case .ring:
            state = .ringing
            stateLabel = "Ringing..."
        }
    }

    private func markActive() async {
        state = .active
        stateLabel = "00:00"
        startDurationTimer()
    }

    private func recordCall(missed: Bool) {
        guard !localDID.isEmpty else { return }
        CallHistoryStore.append(CallRecord(
            callType: callType,
            peerDID: peerDID,
            peerName: contactName.isEmpty ? peerDID : contactName,
            startedAt: startedAt,
            duration: duration,
            missedByCurrentUser: missed && !isOutgoing,
            initiatorDID: isOutgoing ? localDID : peerDID
        ))
    }

    private func loadContactInfo() async {
        if !contactName.isEmpty { return }
        contactName = ContactThreadHelper.truncatedDID(peerDID)
        let tier = ContactTrustIndex.shared.tier(conversationId: "", peerDID: peerDID)
        contactTrustTier = TrustTier.from(tierInt: tier)
    }

    private func startDurationTimer() {
        durationTimer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
            Task { @MainActor in
                guard let self else { return }
                self.duration += 1
                let minutes = Int(self.duration) / 60
                let seconds = Int(self.duration) % 60
                self.stateLabel = String(format: "%02d:%02d", minutes, seconds)
            }
        }
    }
}
