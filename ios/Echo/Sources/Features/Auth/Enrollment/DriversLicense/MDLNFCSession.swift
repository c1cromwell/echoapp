// Features/Auth/Enrollment/DriversLicense/MDLNFCSession.swift
//
// ISO 18013-5 §8.3 NFC device engagement.

import SwiftUI
import CoreNFC

// MARK: - ViewModel

@MainActor
@Observable
final class MDLNFCViewModel {
    enum State {
        case idle
        case polling
        case engaged(engagementURI: String)
        case retrieving
        case verifying
        case success(VerifiedIdentityBundle)
        case failure(EnrollmentError)
    }

    var state: State = .idle

    private var nfcSession: NFCTagReaderSession?
    private var sessionDelegate: NFCDelegate?
    private let verifier = MDLVerifier()
    private let api = EnrollmentAPIClient.shared

    func start(coordinator: EnrollmentCoordinator) {
        guard NFCTagReaderSession.readingAvailable else {
            state = .failure(.nfcUnavailable)
            return
        }

        state = .polling

        let delegate = NFCDelegate(
            onEngagement: { [weak self] uri in
                Task { [weak self] in
                    await self?.continueWithEngagement(uri: uri, coordinator: coordinator)
                }
            },
            onInvalidate: { [weak self] error in
                Task { @MainActor [weak self] in
                    guard let self else { return }
                    if case .polling = self.state {
                        self.state = .failure(error)
                    }
                }
            }
        )
        self.sessionDelegate = delegate

        let session = NFCTagReaderSession(
            pollingOption: [.iso14443],
            delegate: delegate,
            queue: nil
        )
        session?.alertMessage = "Hold your phone near the mDL reader."
        session?.begin()
        self.nfcSession = session
    }

    private func continueWithEngagement(uri: String, coordinator: EnrollmentCoordinator) async {
        state = .engaged(engagementURI: uri)
        do {
            state = .retrieving
            let presentation = try await verifier.retrieveOverBLE(
                engagementURI: uri,
                requestedClaims: .minimumForTier4
            )

            state = .verifying
            let bundle = try await api.finishMDLPresentation(
                sessionID: nil,
                deviceResponse: presentation.deviceResponse,
                sessionTranscript: presentation.sessionTranscript,
                transport: .nfcBLE
            )

            state = .success(bundle)
            coordinator.credentialVerified(bundle)
        } catch let error as EnrollmentError {
            state = .failure(error)
        } catch {
            state = .failure(.transportFailed(underlying: error.localizedDescription))
        }
    }
}

// MARK: - NFC Delegate

final class NFCDelegate: NSObject, NFCTagReaderSessionDelegate {
    let onEngagement: (String) -> Void
    let onInvalidate: (EnrollmentError) -> Void

    init(onEngagement: @escaping (String) -> Void,
         onInvalidate: @escaping (EnrollmentError) -> Void) {
        self.onEngagement = onEngagement
        self.onInvalidate = onInvalidate
    }

    func tagReaderSessionDidBecomeActive(_ session: NFCTagReaderSession) {}

    func tagReaderSession(_ session: NFCTagReaderSession, didInvalidateWithError error: Error) {
        let nfcError = error as? NFCReaderError
        switch nfcError?.code {
        case .readerSessionInvalidationErrorUserCanceled:
            onInvalidate(.userCancelled)
        case .readerSessionInvalidationErrorSessionTimeout:
            onInvalidate(.transportFailed(underlying: "NFC session timed out"))
        default:
            onInvalidate(.transportFailed(underlying: error.localizedDescription))
        }
    }

    func tagReaderSession(_ session: NFCTagReaderSession, didDetect tags: [NFCTag]) {
        guard let tag = tags.first else {
            session.invalidate(errorMessage: "No tag found.")
            return
        }

        session.connect(to: tag) { connectError in
            if let connectError {
                session.invalidate(errorMessage: "Connection failed: \(connectError.localizedDescription)")
                return
            }

            guard case let .iso7816(iso7816Tag) = tag else {
                session.invalidate(errorMessage: "Unsupported tag type.")
                return
            }

            Task {
                do {
                    let engagementURI = try await Self.readDeviceEngagement(from: iso7816Tag)
                    session.alertMessage = "Engagement received. Starting secure transfer…"
                    session.invalidate()
                    self.onEngagement(engagementURI)
                } catch {
                    session.invalidate(errorMessage: "Couldn't read engagement: \(error.localizedDescription)")
                }
            }
        }
    }

    private static func readDeviceEngagement(from tag: NFCISO7816Tag) async throws -> String {
        try await MDLVerifier().readNFCDeviceEngagement(tag: tag)
    }
}

// MARK: - View

struct MDLNFCSessionView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = MDLNFCViewModel()

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 24) {
                Spacer(minLength: 20)

                animatedNFCBadge
                    .frame(width: 180, height: 180)

                VStack(spacing: 8) {
                    Text(title)
                        .font(.system(size: 22, weight: .semibold))
                        .kerning(-0.3)
                        .foregroundStyle(Color.Echo.onSurface)
                        .multilineTextAlignment(.center)
                    Text(subtitle)
                        .font(.system(size: 14))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                        .padding(.horizontal, 28)
                }

                content

                Spacer()
            }
        }
        .navigationTitle("Tap to read")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
    }

    // MARK: - Subviews

    private var animatedNFCBadge: some View {
        ZStack {
            ForEach(0..<3, id: \.self) { i in
                Circle()
                    .stroke(Color.Echo.primaryContainer.opacity(0.15), lineWidth: 1)
                    .scaleEffect(1.0 + Double(i) * 0.22)
            }
            ZStack {
                LinearGradient(
                    colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                    startPoint: .topLeading, endPoint: .bottomTrailing
                )
                .clipShape(Circle())
                Image(systemName: "wave.3.right")
                    .font(.system(size: 38, weight: .semibold))
                    .foregroundStyle(.white)
            }
            .frame(width: 96, height: 96)
        }
    }

    private var title: String {
        switch viewModel.state {
        case .idle:                   return "Ready to tap"
        case .polling:                return "Searching for a reader"
        case .engaged, .retrieving:   return "Transferring"
        case .verifying:              return "Verifying issuer"
        case .success:                return "Verified"
        case .failure:                return "Couldn't complete"
        }
    }

    private var subtitle: String {
        switch viewModel.state {
        case .idle:
            return "Tap the button below, then hold the top of your phone to the reader until the light confirms."
        case .polling:
            return "Hold your phone within 2 cm of the mDL reader. Keep still."
        case .engaged, .retrieving:
            return "Engagement received. Completing transfer over a private BLE channel."
        case .verifying:
            return "Checking the mDL signature against the DMV's IACA root."
        case .success:
            return "Your mDL is verified. One more step to finish enrolling."
        case .failure(let error):
            return error.errorDescription ?? "Something went wrong."
        }
    }

    @ViewBuilder
    private var content: some View {
        switch viewModel.state {
        case .idle:
            startButton("Start NFC session")
                .padding(.horizontal, 20)
        case .polling, .engaged, .retrieving, .verifying:
            ProgressView().controlSize(.large).tint(Color.Echo.primaryContainer).padding(.top, 8)
        case .success:
            EmptyView()
        case .failure:
            VStack(spacing: 10) {
                startButton("Try again")
                Button("Pick a different method") { coordinator.path.removeLast() }
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(.horizontal, 20)
        }
    }

    private func startButton(_ label: String) -> some View {
        Button {
            viewModel.start(coordinator: coordinator)
        } label: {
            HStack {
                Text(label)
                    .font(.system(size: 15, weight: .semibold))
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.6))
            }
            .foregroundStyle(.white)
            .padding(.horizontal, 18)
            .padding(.vertical, 16)
            .frame(maxWidth: .infinity)
            .background(
                LinearGradient(
                    colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                    startPoint: .topLeading, endPoint: .bottomTrailing
                ),
                in: RoundedRectangle(cornerRadius: 28, style: .continuous)
            )
        }
        .buttonStyle(SpringPressStyle())
    }
}
