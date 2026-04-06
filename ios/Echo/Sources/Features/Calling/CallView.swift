// Features/Calling/CallView.swift
// Voice/Video call screen with WebRTC integration

import SwiftUI

// MARK: - Call View

struct CallView: View {
    @StateObject private var viewModel: CallViewModel
    @Environment(\.dismiss) private var dismiss

    init(contactId: String, callType: CallType) {
        _viewModel = StateObject(wrappedValue: CallViewModel(
            contactId: contactId, callType: callType
        ))
    }

    var body: some View {
        ZStack {
            // Background
            if viewModel.callType == .video && viewModel.state == .active {
                // Remote video fills screen
                VideoStreamPlaceholder()
                    .ignoresSafeArea()

                // Self PiP — top left, draggable
                SelfPreviewPiP()
                    .frame(width: 100, height: 140)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                    .glacialShadow()
                    .position(x: 70, y: 120)
            } else {
                // Voice call — icy background
                Color.Echo.surface
                    .icyBackground()
                    .ignoresSafeArea()
            }

            VStack(spacing: 0) {
                // Encryption badge — top
                EncryptionBadge()
                    .padding(.top, 60)

                Spacer()

                // Contact info — center (voice mode)
                if viewModel.callType == .voice || viewModel.state != .active {
                    VStack(spacing: 16) {
                        TrustRingAvatar(
                            imageURL: viewModel.contactAvatar,
                            trustTier: viewModel.contactTrustTier,
                            size: 80,
                            isAnimating: viewModel.state == .connecting
                        )

                        Text(viewModel.contactName)
                            .font(Font.Echo.headlineSm)
                            .foregroundStyle(Color.Echo.onSurface)

                        TrustTierPill(tier: viewModel.contactTrustTier)

                        // State label
                        Text(viewModel.stateLabel)
                            .font(.custom("Inter", size: 32))
                            .fontWeight(.light)
                            .foregroundStyle(Color.Echo.outline)
                            .monospacedDigit()
                    }
                }

                Spacer()

                // Screen sharing banner
                if viewModel.isScreenSharing {
                    ScreenShareBanner(onStop: { viewModel.stopScreenShare() })
                }

                // Call controls
                CallControlsBar(
                    isMuted: viewModel.isMuted,
                    isSpeaker: viewModel.isSpeaker,
                    isCameraOn: viewModel.isCameraOn,
                    callType: viewModel.callType,
                    onToggleMute: { viewModel.toggleMute() },
                    onToggleSpeaker: { viewModel.toggleSpeaker() },
                    onToggleCamera: { viewModel.toggleCamera() },
                    onShareScreen: { viewModel.startScreenShare() },
                    onFlipCamera: { viewModel.flipCamera() },
                    onEndCall: {
                        viewModel.endCall()
                        dismiss()
                    }
                )
                .padding(.bottom, 40)
            }
        }
        .task { await viewModel.startCall() }
        #if os(iOS)
        .statusBarHidden()
        #endif
    }
}

// MARK: - Call Controls Bar

struct CallControlsBar: View {
    let isMuted: Bool
    let isSpeaker: Bool
    let isCameraOn: Bool
    let callType: CallType
    let onToggleMute: () -> Void
    let onToggleSpeaker: () -> Void
    let onToggleCamera: () -> Void
    let onShareScreen: () -> Void
    let onFlipCamera: () -> Void
    let onEndCall: () -> Void

    var body: some View {
        VStack(spacing: 24) {
            // Secondary controls
            HStack(spacing: 24) {
                CallControlButton(
                    icon: isMuted ? "mic.slash.fill" : "mic.fill",
                    isActive: !isMuted,
                    action: onToggleMute
                )

                CallControlButton(
                    icon: callType == .video
                        ? (isCameraOn ? "video.fill" : "video.slash.fill")
                        : (isSpeaker ? "speaker.wave.3.fill" : "speaker.fill"),
                    isActive: callType == .video ? isCameraOn : isSpeaker,
                    action: callType == .video ? onToggleCamera : onToggleSpeaker
                )

                CallControlButton(
                    icon: "rectangle.on.rectangle",
                    isActive: false,
                    action: onShareScreen
                )

                if callType == .video {
                    CallControlButton(
                        icon: "camera.rotate",
                        isActive: false,
                        action: onFlipCamera
                    )
                }
            }

            // End call button — 64px red circle
            Button(action: onEndCall) {
                Image(systemName: "phone.down.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(.white)
                    .frame(width: 64, height: 64)
                    .background(Circle().fill(Color.Echo.error))
                    .shadow(color: Color.Echo.error.opacity(0.4), radius: 16)
            }
            .buttonStyle(SpringButtonStyle())
        }
    }
}

struct CallControlButton: View {
    let icon: String
    let isActive: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: icon)
                .font(.system(size: 22))
                .foregroundStyle(.white)
                .frame(width: 56, height: 56)
                .background(
                    Circle()
                        .fill(.ultraThinMaterial)
                        .opacity(0.6)
                )
                .ghostBorder(opacity: 0.15)
        }
    }
}

// MARK: - Supporting Components

struct EncryptionBadge: View {
    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: "lock.fill")
                .font(.system(size: 10))
            Text("End-to-End Encrypted")
                .font(Font.Echo.labelSm)
        }
        .foregroundStyle(Color.Echo.outline)
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
        .background(
            Capsule()
                .fill(.ultraThinMaterial)
                .opacity(0.5)
        )
    }
}

struct TrustRingAvatar: View {
    let imageURL: URL?
    let trustTier: TrustTier
    let size: CGFloat
    var isAnimating: Bool = false

    @State private var ringScale: CGFloat = 1.0

    var body: some View {
        ZStack {
            // Trust ring
            Circle()
                .stroke(Color(hex: trustTier.color), lineWidth: 3)
                .frame(width: size + 8, height: size + 8)
                .scaleEffect(ringScale)

            // Avatar placeholder
            Circle()
                .fill(Color.Echo.surfaceContainerHigh)
                .frame(width: size, height: size)
                .overlay(
                    Image(systemName: "person.fill")
                        .font(.system(size: size * 0.4))
                        .foregroundStyle(Color.Echo.outline)
                )
        }
        .onAppear {
            if isAnimating {
                withAnimation(.easeInOut(duration: 1.5).repeatForever(autoreverses: true)) {
                    ringScale = 1.1
                }
            }
        }
    }
}

struct TrustTierPill: View {
    let tier: TrustTier

    var body: some View {
        Text(tier.displayName)
            .font(Font.Echo.labelMd)
            .foregroundStyle(Color(hex: tier.color))
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
            .background(
                Capsule()
                    .fill(Color(hex: tier.color).opacity(0.15))
            )
    }
}

struct ScreenShareBanner: View {
    let onStop: () -> Void

    var body: some View {
        HStack {
            Image(systemName: "rectangle.on.rectangle.angled")
                .foregroundStyle(Color.Echo.primaryContainer)
            Text("Screen Sharing")
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
            Spacer()
            Button("Stop") { onStop() }
                .font(Font.Echo.bodyMedium)
                .fontWeight(.bold)
                .foregroundStyle(Color.Echo.error)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
        .background(Color.Echo.surfaceContainerLow)
    }
}

struct VideoStreamPlaceholder: View {
    var body: some View {
        Color.Echo.deepNavy
            .overlay(
                Text("Video Stream")
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.outline)
            )
    }
}

struct SelfPreviewPiP: View {
    var body: some View {
        Color.Echo.surfaceContainerHigh
            .overlay(
                Image(systemName: "person.fill")
                    .font(.system(size: 24))
                    .foregroundStyle(Color.Echo.outline)
            )
    }
}

// MARK: - Spring Button Style

struct SpringButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
            .animation(.spring(response: 0.3, dampingFraction: 0.85), value: configuration.isPressed)
    }
}
