#if os(iOS)
// Features/Auth/Enrollment/DriversLicense/MDLQRScannerView.swift
//
// Scans an ISO 18013-5 §8.2 device engagement QR code and hands it off to
// the MDLVerifier for session establishment over BLE.

import SwiftUI
import AVFoundation

// MARK: - ViewModel

@MainActor
@Observable
final class MDLQRScannerViewModel {
    enum State {
        case preparing
        case scanning
        case engaged(engagementURI: String)
        case retrieving
        case verifying
        case success(VerifiedIdentityBundle)
        case failure(EnrollmentError)
    }

    var state: State = .preparing
    var torchOn: Bool = false

    private let verifier = MDLVerifier()
    private let api = EnrollmentAPIClient.shared

    func didScan(engagementURI: String, coordinator: EnrollmentCoordinator) {
        guard case .scanning = state else { return }  // debounce
        state = .engaged(engagementURI: engagementURI)
        Task { await continuePresentation(uri: engagementURI, coordinator: coordinator) }
    }

    func didFailToPrepare(_ error: EnrollmentError) {
        state = .failure(error)
    }

    func didStartScanning() {
        if case .preparing = state { state = .scanning }
    }

    private func continuePresentation(uri: String, coordinator: EnrollmentCoordinator) async {
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
                transport: .qrBLE
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

// MARK: - View

struct MDLQRScannerView: View {
    let coordinator: EnrollmentCoordinator
    @State private var viewModel = MDLQRScannerViewModel()

    var body: some View {
        ZStack {
            Color.Echo.deepNavy.ignoresSafeArea()

            switch viewModel.state {
            case .preparing, .scanning, .engaged:
                cameraLayer
                    .ignoresSafeArea()
                overlay
            case .retrieving, .verifying:
                statusPanel(
                    title: viewModel.state.isVerifying ? "Verifying issuer" : "Connecting over Bluetooth",
                    subtitle: viewModel.state.isVerifying
                        ? "Checking the mDL signature against the DMV's root."
                        : "Hold still — establishing an encrypted channel to your wallet.",
                    showSpinner: true
                )
            case .success:
                statusPanel(
                    title: "Verified",
                    subtitle: "One more step to finish enrolling.",
                    showSpinner: false
                )
            case .failure(let error):
                errorPanel(error)
            }
        }
        .navigationTitle("Scan QR")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(.hidden, for: .navigationBar)
        .toolbarColorScheme(.dark, for: .navigationBar)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button {
                    viewModel.torchOn.toggle()
                } label: {
                    Image(systemName: viewModel.torchOn ? "bolt.fill" : "bolt.slash.fill")
                        .foregroundStyle(.white)
                }
            }
        }
    }

    // MARK: - Subviews

    private var cameraLayer: some View {
        QRCameraPreview(
            torchOn: viewModel.torchOn,
            onReady: { viewModel.didStartScanning() },
            onPrepareError: { viewModel.didFailToPrepare($0) },
            onQR: { payload in
                if payload.lowercased().hasPrefix("mdoc:") {
                    viewModel.didScan(engagementURI: payload, coordinator: coordinator)
                }
            }
        )
    }

    private var overlay: some View {
        VStack {
            Spacer()
            scannerReticle
            Spacer()
            instructions
                .padding(.bottom, 32)
        }
    }

    private var scannerReticle: some View {
        RoundedRectangle(cornerRadius: 28, style: .continuous)
            .stroke(Color.white.opacity(0.85), lineWidth: 2)
            .frame(width: 260, height: 260)
            .shadow(color: Color.Echo.primaryContainer.opacity(0.35), radius: 20)
            .overlay(
                RoundedRectangle(cornerRadius: 28, style: .continuous)
                    .stroke(Color.Echo.primaryContainer, lineWidth: 0.5)
                    .padding(-2)
            )
    }

    private var instructions: some View {
        VStack(spacing: 6) {
            Text("Line up the mDL engagement QR")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
            Text("Ask the verifier terminal or your issuer app to display its engagement code.")
                .font(.system(size: 12))
                .foregroundStyle(.white.opacity(0.75))
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
                .lineSpacing(2)
        }
    }

    private func statusPanel(title: String, subtitle: String, showSpinner: Bool) -> some View {
        VStack(spacing: 18) {
            if showSpinner {
                ProgressView().controlSize(.large).tint(.white)
            } else {
                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 48, weight: .semibold))
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            VStack(spacing: 6) {
                Text(title)
                    .font(.system(size: 20, weight: .semibold))
                    .foregroundStyle(.white)
                Text(subtitle)
                    .font(.system(size: 14))
                    .foregroundStyle(.white.opacity(0.75))
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }
        }
    }

    private func errorPanel(_ error: EnrollmentError) -> some View {
        VStack(spacing: 20) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 40, weight: .semibold))
                .foregroundStyle(.yellow)
            Text(error.errorDescription ?? "Something went wrong")
                .font(.system(size: 14))
                .foregroundStyle(.white)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 24)
            Button {
                coordinator.path.removeLast()
            } label: {
                Text("Pick a different method")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 22)
                    .padding(.vertical, 13)
                    .background(Color.Echo.primaryContainer, in: Capsule())
            }
        }
    }
}

// MARK: - AVFoundation camera preview

struct QRCameraPreview: UIViewRepresentable {
    let torchOn: Bool
    let onReady: () -> Void
    let onPrepareError: (EnrollmentError) -> Void
    let onQR: (String) -> Void

    func makeUIView(context: Context) -> QRPreviewUIView {
        let view = QRPreviewUIView(frame: .zero)
        view.delegate = context.coordinator
        view.start()
        return view
    }

    func updateUIView(_ uiView: QRPreviewUIView, context: Context) {
        uiView.setTorch(torchOn)
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(onReady: onReady, onPrepareError: onPrepareError, onQR: onQR)
    }

    final class Coordinator: NSObject, QRPreviewDelegate {
        let onReady: () -> Void
        let onPrepareError: (EnrollmentError) -> Void
        let onQR: (String) -> Void

        init(onReady: @escaping () -> Void,
             onPrepareError: @escaping (EnrollmentError) -> Void,
             onQR: @escaping (String) -> Void) {
            self.onReady = onReady
            self.onPrepareError = onPrepareError
            self.onQR = onQR
        }

        func qrPreviewDidStart() { onReady() }
        func qrPreviewDidFail(_ error: EnrollmentError) { onPrepareError(error) }
        func qrPreviewDidRead(_ payload: String) { onQR(payload) }
    }
}

protocol QRPreviewDelegate: AnyObject {
    func qrPreviewDidStart()
    func qrPreviewDidFail(_ error: EnrollmentError)
    func qrPreviewDidRead(_ payload: String)
}

final class QRPreviewUIView: UIView, AVCaptureMetadataOutputObjectsDelegate {
    weak var delegate: QRPreviewDelegate?

    private let session = AVCaptureSession()
    private let queue = DispatchQueue(label: "app.echo.enrollment.qr")
    private var device: AVCaptureDevice?

    override class var layerClass: AnyClass { AVCaptureVideoPreviewLayer.self }
    var previewLayer: AVCaptureVideoPreviewLayer { layer as! AVCaptureVideoPreviewLayer }

    func start() {
        guard let device = AVCaptureDevice.default(for: .video) else {
            delegate?.qrPreviewDidFail(.cameraUnavailable)
            return
        }
        self.device = device

        do {
            let input = try AVCaptureDeviceInput(device: device)
            session.beginConfiguration()
            session.sessionPreset = .high
            if session.canAddInput(input) { session.addInput(input) }

            let output = AVCaptureMetadataOutput()
            if session.canAddOutput(output) {
                session.addOutput(output)
                output.setMetadataObjectsDelegate(self, queue: queue)
                output.metadataObjectTypes = [.qr]
            }
            session.commitConfiguration()

            previewLayer.session = session
            previewLayer.videoGravity = .resizeAspectFill

            queue.async { [weak self] in
                self?.session.startRunning()
                Task { @MainActor in self?.delegate?.qrPreviewDidStart() }
            }
        } catch {
            delegate?.qrPreviewDidFail(.transportFailed(underlying: error.localizedDescription))
        }
    }

    func setTorch(_ on: Bool) {
        guard let device, device.hasTorch else { return }
        do {
            try device.lockForConfiguration()
            device.torchMode = on ? .on : .off
            device.unlockForConfiguration()
        } catch {
            // silent — torch is non-critical
        }
    }

    override func layoutSubviews() {
        super.layoutSubviews()
        previewLayer.frame = bounds
    }

    func metadataOutput(
        _ output: AVCaptureMetadataOutput,
        didOutput metadataObjects: [AVMetadataObject],
        from connection: AVCaptureConnection
    ) {
        guard let first = metadataObjects.first as? AVMetadataMachineReadableCodeObject,
              let payload = first.stringValue else { return }

        Task { @MainActor [weak self] in
            self?.delegate?.qrPreviewDidRead(payload)
        }
    }

    deinit {
        session.stopRunning()
    }
}

// MARK: - State helpers

private extension MDLQRScannerViewModel.State {
    var isVerifying: Bool {
        if case .verifying = self { return true }
        return false
    }
}
#endif
