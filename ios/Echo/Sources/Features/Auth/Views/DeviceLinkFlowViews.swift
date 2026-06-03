#if os(iOS)
import SwiftUI
import CoreImage.CIFilterBuiltins

/// Trusted device: show QR with 5-minute registration token (WO-288).
struct LinkNewDeviceQRView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var qrImage: UIImage?
    @State private var expiresIn = 300
    @State private var errorMessage: String?
    @State private var isLoading = true

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                if isLoading {
                    ProgressView("Preparing secure link…")
                } else if let qrImage {
                    Image(uiImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .scaledToFit()
                        .frame(width: 220, height: 220)
                        .padding()
                        .background(Color.white)
                        .clipShape(RoundedRectangle(cornerRadius: 16))

                    Text("Scan with your new iPhone or iPad")
                        .font(.headline)
                    Text("Expires in \(expiresIn / 60) minutes. Message history stays on this device.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal)
                } else {
                    ContentUnavailableView(
                        "Couldn't create link",
                        systemImage: "qrcode",
                        description: Text(errorMessage ?? "Try again from Account settings.")
                    )
                }
            }
            .padding()
            .navigationTitle("Link new device")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .task { await loadToken() }
        }
    }

    private func loadToken() async {
        isLoading = true
        defer { isLoading = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = DeviceLinkError.notAuthenticated.localizedDescription
            return
        }
        let link = DeviceLinkAPIClient(apiClient: client)
        do {
            let response = try await link.issueLinkToken()
            expiresIn = response.expiresIn
            guard let url = DeviceLinkAPIClient.linkQRURL(token: response.token) else {
                errorMessage = "Invalid link payload"
                return
            }
            qrImage = QRCodeGenerator.image(from: url.absoluteString)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

/// New device: scan `echo://link-device?token=` and register Secure Enclave key (WO-288).
struct LinkDeviceScanView: View {
    var onLinked: (() -> Void)?

    @Environment(\.dismiss) private var dismiss
    @State private var manualToken = ""
    @State private var isRegistering = false
    @State private var errorMessage: String?
    @State private var success = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 16) {
                if success {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(.green)
                    Text("Device linked")
                        .font(.headline)
                    Text("Sign in with Face ID on this device. Your message history stays on your old device.")
                        .multilineTextAlignment(.center)
                        .foregroundStyle(.secondary)
                        .padding(.horizontal)
                    Button("Continue") {
                        onLinked?()
                        dismiss()
                    }
                    .buttonStyle(.borderedProminent)
                } else {
                    Text("Scan the QR on your signed-in device, or paste the link token.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal)

                    LiveQRCodeScannerView { raw in
                        guard let url = URL(string: raw.trimmingCharacters(in: .whitespacesAndNewlines)),
                              case .linkDevice(let token) = EchoDeepLink.parse(url) else { return }
                        Task { await register(token: token) }
                    }
                    .frame(height: 240)
                    .clipShape(RoundedRectangle(cornerRadius: 12))

                    TextField("Registration token", text: $manualToken)
                        .textFieldStyle(.roundedBorder)
                        .font(.caption.monospaced())
                        .autocapitalization(.none)
                        .disableAutocorrection(true)

                    if let errorMessage {
                        Text(errorMessage)
                            .font(.footnote)
                            .foregroundStyle(.red)
                    }

                    Button {
                        Task { await register(token: manualToken.trimmingCharacters(in: .whitespacesAndNewlines)) }
                    } label: {
                        if isRegistering {
                            ProgressView()
                        } else {
                            Text("Link this device")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(isRegistering || manualToken.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                }
            }
            .padding()
            .navigationTitle("New device")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }

    private func register(token: String) async {
        guard !token.isEmpty else { return }
        isRegistering = true
        errorMessage = nil
        defer { isRegistering = false }

        do {
            let pubBase64 = try await SecureEnclaveManager.shared
                .generateBiometricProtectedKey(id: "echo-identity-signing-linked")
            let hex = try DeviceLinkAPIClient.publicKeyHex(fromBase64PublicKey: pubBase64)
            let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            _ = try await DeviceLinkAPIClient(apiClient: client).completeLink(
                token: token,
                newPublicKeyHex: hex,
                deviceLabel: UIDevice.current.name
            )
            success = true
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

enum QRCodeGenerator {
    static func image(from string: String) -> UIImage? {
        let context = CIContext()
        let filter = CIFilter.qrCodeGenerator()
        filter.message = Data(string.utf8)
        guard let output = filter.outputImage else { return nil }
        let scaled = output.transformed(by: CGAffineTransform(scaleX: 8, y: 8))
        guard let cg = context.createCGImage(scaled, from: scaled.extent) else { return nil }
        return UIImage(cgImage: cg)
    }
}
#endif
