// Features/Calling/CallViewModel.swift
// Manages voice/video call state and WebRTC signaling

import Foundation
import SwiftUI
import Combine

@MainActor
class CallViewModel: ObservableObject {
    let contactId: String
    let callType: CallType

    @Published var state: CallState = .idle
    @Published var stateLabel: String = "Calling..."
    @Published var contactName: String = ""
    @Published var contactAvatar: URL?
    @Published var contactTrustTier: TrustTier = .newcomer
    @Published var duration: TimeInterval = 0

    // Audio/Video controls
    @Published var isMuted = false
    @Published var isSpeaker = false
    @Published var isCameraOn = true
    @Published var isScreenSharing = false

    // Streams (placeholder types for WebRTC integration)
    @Published var localStream: Any?
    @Published var remoteStream: Any?

    private var durationTimer: Timer?

    init(contactId: String, callType: CallType) {
        self.contactId = contactId
        self.callType = callType
        self.isCameraOn = callType == .video
    }

    func startCall() async {
        state = .connecting
        stateLabel = "Connecting..."

        // Load contact info
        await loadContactInfo()

        // Simulate connection (replace with real WebRTC signaling)
        try? await Task.sleep(nanoseconds: 2_000_000_000)

        state = .active
        stateLabel = "00:00"
        startDurationTimer()
    }

    func endCall() {
        durationTimer?.invalidate()
        durationTimer = nil
        state = .ended
        stateLabel = "Call Ended"
    }

    func toggleMute() {
        isMuted.toggle()
    }

    func toggleSpeaker() {
        isSpeaker.toggle()
    }

    func toggleCamera() {
        isCameraOn.toggle()
    }

    func flipCamera() {
        // Toggle front/back camera
    }

    func startScreenShare() {
        isScreenSharing = true
    }

    func stopScreenShare() {
        isScreenSharing = false
    }

    // MARK: - Private

    private func loadContactInfo() async {
        // TODO: Load from contacts service
        contactName = "Contact"
        contactTrustTier = .verified
    }

    private func startDurationTimer() {
        durationTimer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
            Task { @MainActor in
                guard let self = self else { return }
                self.duration += 1
                let minutes = Int(self.duration) / 60
                let seconds = Int(self.duration) % 60
                self.stateLabel = String(format: "%02d:%02d", minutes, seconds)
            }
        }
    }
}
