#if os(iOS)
import AVFoundation
import SwiftUI
import UIKit

/// Camera-backed QR scanner (profile / invite codes).
struct LiveQRCodeScannerView: View {
    let onScan: (String) -> Void
    @Environment(\.dismiss) private var dismiss
    @State private var cameraDenied = false

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            if cameraDenied {
                VStack(spacing: 16) {
                    Text("Camera access is required to scan QR codes.")
                        .font(.system(size: 15))
                        .foregroundStyle(.white)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                    Button("Close") { dismiss() }
                        .buttonStyle(.borderedProminent)
                }
            } else {
                QRScannerRepresentable(onCode: onScan)
                    .ignoresSafeArea()

                VStack {
                    Spacer()
                    Text("Align Echo profile QR in the frame")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.white.opacity(0.85))
                        .padding(.bottom, 8)
                    Button("Cancel") { dismiss() }
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(.white)
                        .padding(.horizontal, 28)
                        .padding(.vertical, 12)
                        .background(.ultraThinMaterial, in: Capsule())
                        .padding(.bottom, 32)
                }
            }
        }
    }
}

private struct QRScannerRepresentable: UIViewControllerRepresentable {
    let onCode: (String) -> Void

    func makeUIViewController(context: Context) -> QRScannerViewController {
        let controller = QRScannerViewController()
        controller.onCode = onCode
        return controller
    }

    func updateUIViewController(_ uiViewController: QRScannerViewController, context: Context) {}
}

private final class QRScannerViewController: UIViewController, AVCaptureMetadataOutputObjectsDelegate {
    var onCode: ((String) -> Void)?
    private let session = AVCaptureSession()
    private var didEmit = false

    override func viewDidLoad() {
        super.viewDidLoad()
        view.backgroundColor = .black
        switch AVCaptureDevice.authorizationStatus(for: .video) {
        case .authorized:
            configureSession()
        case .notDetermined:
            AVCaptureDevice.requestAccess(for: .video) { [weak self] granted in
                DispatchQueue.main.async {
                    guard granted else { return }
                    self?.configureSession()
                }
            }
        default:
            break
        }
    }

    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        (view.layer.sublayers?.first { $0 is AVCaptureVideoPreviewLayer })?.frame = view.bounds
    }

    private func configureSession() {
        guard let device = AVCaptureDevice.default(for: .video),
              let input = try? AVCaptureDeviceInput(device: device),
              session.canAddInput(input) else { return }

        session.beginConfiguration()
        session.addInput(input)
        let output = AVCaptureMetadataOutput()
        if session.canAddOutput(output) {
            session.addOutput(output)
            output.setMetadataObjectsDelegate(self, queue: DispatchQueue.main)
            output.metadataObjectTypes = [.qr]
        }
        session.commitConfiguration()

        let preview = AVCaptureVideoPreviewLayer(session: session)
        preview.videoGravity = .resizeAspectFill
        preview.frame = view.bounds
        view.layer.insertSublayer(preview, at: 0)

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            self?.session.startRunning()
        }
    }

    func metadataOutput(
        _ output: AVCaptureMetadataOutput,
        didOutput metadataObjects: [AVMetadataObject],
        from connection: AVCaptureConnection
    ) {
        guard !didEmit,
              let object = metadataObjects.first as? AVMetadataMachineReadableCodeObject,
              object.type == .qr,
              let value = object.stringValue else { return }
        didEmit = true
        session.stopRunning()
        onCode?(value)
    }
}
#endif
