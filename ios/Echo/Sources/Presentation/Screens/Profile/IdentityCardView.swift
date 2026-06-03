#if os(iOS)
import SwiftUI

// Signal-style identity row: friendly name up front, DID secondary, QR affordance on the right.

public struct IdentityCardView: View {
    let displayName: String
    let username: String
    let did: String
    let joinDate: Date?
    let trustTier: Int?

    @State private var showQR = false
    @State private var copied = false

    public init(
        displayName: String,
        username: String,
        did: String,
        joinDate: Date? = nil,
        trustTier: Int? = nil
    ) {
        self.displayName = displayName
        self.username = username
        self.did = did
        self.joinDate = joinDate
        self.trustTier = trustTier
    }

    private var shortDID: String {
        guard did.count > 20 else { return did }
        return "\(did.prefix(12))···\(did.suffix(6))"
    }

    private var primaryLabel: String {
        let trimmed = displayName.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty, trimmed.lowercased() != username.lowercased() {
            return trimmed
        }
        return username.isEmpty ? "Your profile" : "@\(username)"
    }

    private var secondaryHandle: String? {
        let trimmed = displayName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !username.isEmpty else { return nil }
        if trimmed.isEmpty || trimmed.lowercased() == username.lowercased() {
            return nil
        }
        return "@\(username)"
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .center, spacing: 12) {
                Circle()
                    .fill(Color.echoTrustGreen.opacity(0.12))
                    .overlay(Circle().stroke(Color.echoTrustGreen.opacity(0.30), lineWidth: 1))
                    .frame(width: 48, height: 48)
                    .overlay(
                        Text(String((username.isEmpty ? displayName : username).prefix(1)).uppercased())
                            .font(.system(size: 20, weight: .semibold))
                            .foregroundStyle(Color.echoTrustGreen)
                    )

                VStack(alignment: .leading, spacing: 3) {
                    Text(primaryLabel)
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundStyle(Color.echoInk)
                        .lineLimit(1)

                    if let secondaryHandle {
                        Text(secondaryHandle)
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoInk55)
                    }

                    Text(shortDID)
                        .font(.echomono(11))
                        .foregroundStyle(Color.echoInk40)
                        .lineLimit(1)

                    if let tier = trustTier {
                        Text("Trust tier T\(tier)")
                            .font(.echomono(10))
                            .foregroundStyle(Color.echoTrustGreen)
                    }

                    Text("Self-sovereign did:key · no phone required")
                        .font(.system(size: 11))
                        .foregroundStyle(Color.echoInk55)
                        .padding(.top, 2)
                }

                Spacer(minLength: 4)

                Button {
                    showQR = true
                } label: {
                    Image(systemName: "qrcode")
                        .font(.system(size: 22, weight: .medium))
                        .foregroundStyle(Color.echoInk)
                        .frame(width: 44, height: 44)
                        .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 12))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.echoHair, lineWidth: 1))
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Show QR code to share your profile")
            }

            if let d = joinDate {
                Text("Joined \(d.formatted(.dateTime.year().month().day()))")
                    .font(.echomono(10))
                    .foregroundStyle(Color.echoInk40)
                    .padding(.top, 10)
            }

            Button(action: copyDID) {
                Text(copied ? "DID copied" : "Copy DID")
                    .font(.system(size: 12, weight: .medium))
                    .foregroundStyle(copied ? Color.echoTrustGreen : Color.echoInk55)
            }
            .padding(.top, 10)
        }
        .padding(16)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 16))
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
        .sheet(isPresented: $showQR) {
            QRIdentitySheet(
                did: did,
                username: username,
                displayName: displayName,
                trustTier: trustTier ?? 0
            )
        }
    }

    private func copyDID() {
        UIPasteboard.general.string = did
        withAnimation { copied = true }
        Task { try? await Task.sleep(nanoseconds: 2_000_000_000); copied = false }
    }
}

// MARK: - What's protected list

public struct IdentityProtectedList: View {
    let hasPhrase: Bool
    let hasSMS: Bool

    public init(hasPhrase: Bool, hasSMS: Bool) {
        self.hasPhrase = hasPhrase
        self.hasSMS = hasSMS
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("WHAT'S PROTECTED")
                .font(.echomono(10))
                .foregroundStyle(Color.echoInk40)
                .padding(.horizontal, 4)
                .padding(.bottom, 8)

            let rows: [(String, String, Bool)] = [
                ("Identity key",    "Secure Enclave · Face ID",      true),
                ("Storage key",     "Derived · biometric-gated",     true),
                ("Recovery phrase", "24-word backup · " + (hasPhrase ? "saved" : "not set"), hasPhrase),
                ("SMS recovery",    hasSMS ? "Configured" : "Not configured", hasSMS),
            ]

            VStack(spacing: 0) {
                ForEach(Array(rows.enumerated()), id: \.offset) { idx, row in
                    HStack(spacing: 12) {
                        Circle()
                            .fill(row.2 ? Color.echoTrustGreen : Color.echoInk20)
                            .frame(width: 6, height: 6)

                        VStack(alignment: .leading, spacing: 2) {
                            Text(row.0)
                                .font(.system(size: 13.5, weight: .medium))
                                .foregroundStyle(Color.echoInk)
                            Text(row.1)
                                .font(.system(size: 11.5))
                                .foregroundStyle(Color.echoInk55)
                        }

                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.system(size: 13))
                            .foregroundStyle(Color.echoInk40)
                    }
                    .padding(.horizontal, 14)
                    .padding(.vertical, 12)

                    if idx < rows.count - 1 {
                        Divider().padding(.leading, 32)
                    }
                }
            }
            .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 14))
            .overlay(RoundedRectangle(cornerRadius: 14).stroke(Color.echoHair, lineWidth: 1))
        }
    }
}

// MARK: - QR sheet (share + scan)

private struct QRIdentitySheet: View {
    let did: String
    let username: String
    let displayName: String
    let trustTier: Int
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = QRIdentityViewModel()
    @State private var addCoordinator = QRContactAddCoordinator()
    @State private var showScanner = false

    private var headline: String {
        let trimmed = displayName.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty { return trimmed }
        return username.isEmpty ? "Your profile" : "@\(username)"
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                Text("Others can scan this code to find you on Echo. Your DID stays behind the scenes.")
                    .font(.system(size: 13))
                    .foregroundStyle(Color.echoInk55)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 24)

                if let qrImage = viewModel.qrCodeImage {
                    Image(uiImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .frame(width: 220, height: 220)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                        .padding(12)
                        .background(Color.white, in: RoundedRectangle(cornerRadius: 16))
                }

                Text(headline)
                    .font(.system(size: 17, weight: .semibold))
                if !username.isEmpty {
                    Text("@\(username)")
                        .font(.system(size: 14))
                        .foregroundStyle(Color.echoInk55)
                }
                Text(did)
                    .font(.echomono(10))
                    .foregroundStyle(Color.echoInk40)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)

                HStack(spacing: 12) {
                    Button {
                        showScanner = true
                    } label: {
                        Label("Scan", systemImage: "camera.viewfinder")
                            .frame(maxWidth: .infinity)
                    }
                    .buttonStyle(.bordered)

                    Button("Share link") { viewModel.shareLink() }
                        .frame(maxWidth: .infinity)
                        .buttonStyle(.borderedProminent)
                        .tint(Color.echoSignal)
                }
                .padding(.horizontal, 24)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color.echoPaper.ignoresSafeArea())
            .navigationTitle("Share profile")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .cancellationAction) { Button("Done") { dismiss() } } }
            .onAppear {
                viewModel.configure(did: did, username: username, trustTier: trustTier)
            }
            .fullScreenCover(isPresented: $showScanner) {
                LiveQRCodeScannerView { scanned in
                    showScanner = false
                    Task { await addCoordinator.handleScan(scanned) }
                }
            }
            .alert(
                addCoordinator.resultIsError ? "Couldn’t add contact" : "Contact added",
                isPresented: Binding(
                    get: { addCoordinator.resultMessage != nil },
                    set: { if !$0 { addCoordinator.reset() } }
                )
            ) {
                Button("OK", role: .cancel) { addCoordinator.reset() }
            } message: {
                Text(addCoordinator.resultMessage ?? "")
            }
        }
    }
}

#Preview {
    ScrollView {
        VStack(spacing: 20) {
            IdentityCardView(
                displayName: "Ada Lovelace",
                username: "ada",
                did: "did:key:z6MkhRfL7r9NB8Y3xQpJh···aV2nf9Q2",
                joinDate: Date()
            )
            IdentityProtectedList(hasPhrase: true, hasSMS: false)
        }
        .padding(24)
    }
    .background(Color.echoPaper)
    .preferredColorScheme(.light)
}
#endif
