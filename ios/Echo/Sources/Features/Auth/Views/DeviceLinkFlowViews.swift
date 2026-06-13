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
    @State private var syncStatus: String?
    @State private var knownPublicKeys: Set<String> = []
    @State private var pollTask: Task<Void, Never>?

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
                    Text("Expires in \(expiresIn / 60) minutes. Message history syncs after the new device links.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal)

                    if let syncStatus {
                        Text(syncStatus)
                            .font(.footnote)
                            .foregroundStyle(.green)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal)
                    }
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
            .onDisappear { pollTask?.cancel() }
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
            await startPollingForLinkedDevice(link: link)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func startPollingForLinkedDevice(link: DeviceLinkAPIClient) async {
        guard let did = await CurrentUserSession.currentDID(),
              let sync = DIContainer.shared.resolveDeviceHistorySync() else { return }

        if let devices = try? await link.listRegisteredDevices(did: did) {
            knownPublicKeys = Set(devices.map { $0.publicKeyHex.lowercased() })
        }

        pollTask?.cancel()
        pollTask = Task {
            for _ in 0..<100 {
                try? await Task.sleep(nanoseconds: 3_000_000_000)
                if Task.isCancelled { return }
                guard let devices = try? await link.listRegisteredDevices(did: did) else { continue }
                let fresh = devices.filter { !knownPublicKeys.contains($0.publicKeyHex.lowercased()) }
                for device in fresh {
                    do {
                        try await sync.seedHistoryToDevice(publicKeyHex: device.publicKeyHex)
                        knownPublicKeys.insert(device.publicKeyHex.lowercased())
                        syncStatus = "Message history sent to \(device.deviceLabel ?? "new device")."
                    } catch {
                        syncStatus = "New device linked. Open Messages to finish history sync."
                    }
                }
            }
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
                    Text("Sign in with Face ID on this device. Your message history will download after unlock.")
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
            DeviceIdentityStore.assignSyncDeviceId(fromPublicKeyHex: hex)
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
