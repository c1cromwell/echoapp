#if os(iOS)
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
    @Published var localVideoTrack: Any?
    @Published var remoteVideoTrack: Any?
    @Published var usesLiveWebRTC = false

    private let callId: String
    private var durationTimer: Timer?
    private let session = WebRTCCallSession()
    private var signaling: CallSignalingService?
    private var callKitUUID: UUID?
    private var localDID: String = ""
    private var startedAt = Date()
    private let pendingOfferSDP: String?
    private var hasAppliedIncomingOffer = false

    init(
        peerDID: String,
        callType: CallType,
        contactName: String = "",
        isOutgoing: Bool = true,
        incomingCallId: String? = nil,
        pendingOfferSDP: String? = nil
    ) {
        self.peerDID = peerDID
        self.callType = callType
        self.contactName = contactName
        self.isOutgoing = isOutgoing
        self.callId = incomingCallId ?? UUID().uuidString
        self.pendingOfferSDP = pendingOfferSDP
        self.isCameraOn = callType == .video
    }

    func startCall() async {
        state = .connecting
        stateLabel = isOutgoing ? "Calling..." : "Connecting..."
        startedAt = Date()
        await loadContactInfo()
        await configureSignaling()

        do {
            let iceServers = try await signaling?.fetchICEServers() ?? []
            session.configure(
                iceServers: iceServers,
                callType: callType,
                onIceCandidate: { [weak self] candidate in
                    Task { @MainActor in
                        guard let self else { return }
                        try? await self.signaling?.sendIce(
                            callId: self.callId,
                            to: self.peerDID,
                            candidate: candidate
                        )
                    }
                },
                onConnected: { [weak self] in
                    Task { @MainActor in await self?.markActive() }
                }
            )
            usesLiveWebRTC = session.usesLiveWebRTC
            localVideoTrack = session.localVideoTrack
            remoteVideoTrack = session.remoteVideoTrack

            if isOutgoing {
                callKitUUID = CallKitCoordinator.shared.reportOutgoingCall(
                    peerName: contactName.isEmpty ? peerDID : contactName,
                    hasVideo: callType == .video,
                    onEnd: { [weak self] in self?.endCall() }
                )
                let sdp = try await session.createOffer(callType: callType)
                try await signaling?.sendOffer(
                    callId: callId,
                    to: peerDID,
                    callType: callType,
                    sdp: sdp
                )
            } else if let pendingOfferSDP, !hasAppliedIncomingOffer {
                hasAppliedIncomingOffer = true
                state = .ringing
                stateLabel = "Incoming call..."
                let answer = try await session.applyRemoteOffer(pendingOfferSDP, callType: callType)
                try await signaling?.sendAnswer(callId: callId, to: peerDID, sdp: answer)
                localVideoTrack = session.localVideoTrack
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
        CallKitCoordinator.shared.endCall()
        state = .ended
        stateLabel = "Call Ended"
        recordCall(missed: false)
        Task {
            try? await signaling?.sendHangup(callId: callId, to: peerDID)
        }
    }

    func toggleMute() {
        isMuted.toggle()
        session.setMuted(isMuted)
    }

    func toggleSpeaker() {
        isSpeaker.toggle()
        session.setSpeakerEnabled(isSpeaker)
    }

    func toggleCamera() {
        isCameraOn.toggle()
        session.setVideoEnabled(isCameraOn)
        localVideoTrack = session.localVideoTrack
    }

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
        guard event.callId == callId else { return }
        switch event.action {
        case .offer:
            guard !isOutgoing, let sdp = event.sdp, !hasAppliedIncomingOffer else { return }
            hasAppliedIncomingOffer = true
            Task {
                do {
                    let answer = try await session.applyRemoteOffer(sdp, callType: callType)
                    try await signaling?.sendAnswer(callId: event.callId, to: event.peerDID, sdp: answer)
                    localVideoTrack = session.localVideoTrack
                } catch {
                    state = .ended
                    stateLabel = "Call failed"
                }
            }
        case .answer:
            guard let sdp = event.sdp else { return }
            Task {
                do {
                    try await session.applyAnswer(sdp)
                } catch {
                    state = .ended
                    stateLabel = "Call failed"
                }
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
        guard state != .active else { return }
        state = .active
        stateLabel = "00:00"
        remoteVideoTrack = session.remoteVideoTrack
        localVideoTrack = session.localVideoTrack
        CallKitCoordinator.shared.reportConnected()
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
#endif
