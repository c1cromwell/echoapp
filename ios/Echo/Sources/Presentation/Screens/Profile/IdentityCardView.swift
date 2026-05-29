#if os(iOS)
import SwiftUI

// Design review finding: "Make the cryptography legible."
// The user should see their actual public key — it IS their identity.
// Show it, let them copy it, let them share it as a QR.

public struct IdentityCardView: View {
    let username: String
    let did: String
    let joinDate: Date?
    let trustTier: Int?

    @State private var showQR = false
    @State private var copied = false

    public init(username: String, did: String, joinDate: Date? = nil, trustTier: Int? = nil) {
        self.username = username
        self.did = did
        self.joinDate = joinDate
        self.trustTier = trustTier
    }

    private var shortDID: String {
        guard did.count > 20 else { return did }
        let prefix = String(did.prefix(16))
        let suffix = String(did.suffix(6))
        return "\(prefix)···\(suffix)"
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Avatar + name
            HStack(spacing: 14) {
                Circle()
                    .fill(Color.echoTrustGreen.opacity(0.12))
                    .overlay(Circle().stroke(Color.echoTrustGreen.opacity(0.30), lineWidth: 1))
                    .frame(width: 52, height: 52)
                    .overlay(
                        Text(String(username.prefix(1)).uppercased())
                            .font(.system(size: 22, weight: .semibold))
                            .foregroundStyle(Color.echoTrustGreen)
                    )

                VStack(alignment: .leading, spacing: 2) {
                    Text("@\(username)")
                        .font(.system(size: 18, weight: .semibold))
                        .tracking(-0.3)
                        .foregroundStyle(Color.echoInk)
                    if let tier = trustTier {
                        Text("Trust tier T\(tier)")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoTrustGreen)
                    }
                    if let d = joinDate {
                        Text("joined \(d.formatted(.dateTime.year().month().day()))")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoInk55)
                    }
                }
            }

            Divider()
                .foregroundStyle(Color.echoHair)
                .padding(.vertical, 14)

            // The public key — the most important thing to show
            VStack(alignment: .leading, spacing: 8) {
                Text("YOUR PUBLIC KEY")
                    .font(.echomono(10))
                    .foregroundStyle(Color.echoInk40)

                Text(did)
                    .font(.echomono(12.5))
                    .foregroundStyle(Color.echoInk)
                    .lineLimit(3)
                    .fixedSize(horizontal: false, vertical: true)

                HStack(spacing: 8) {
                    Button(action: { showQR = true }) {
                        Text("Show QR")
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(Color.echoInk)
                            .padding(.horizontal, 14)
                            .padding(.vertical, 8)
                            .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 9))
                            .overlay(RoundedRectangle(cornerRadius: 9)
                                .stroke(Color.echoHair, lineWidth: 1))
                    }

                    Button(action: copyDID) {
                        Text(copied ? "Copied ✓" : "Copy")
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(copied ? Color.echoTrustGreen : Color.echoInk)
                            .padding(.horizontal, 14)
                            .padding(.vertical, 8)
                            .background(Color.echoPaper, in: RoundedRectangle(cornerRadius: 9))
                            .overlay(RoundedRectangle(cornerRadius: 9)
                                .stroke(Color.echoHair, lineWidth: 1))
                    }
                }
            }
        }
        .padding(20)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 16))
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
        .sheet(isPresented: $showQR) {
            QRIdentitySheet(did: did, username: username, trustTier: trustTier ?? 0)
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

// MARK: - QR sheet (uses QRIdentityViewModel payload)

private struct QRIdentitySheet: View {
    let did: String
    let username: String
    let trustTier: Int
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = QRIdentityViewModel()

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Text("Share your public key")
                    .font(.system(size: 17, weight: .semibold))

                if let qrImage = viewModel.qrCodeImage {
                    Image(uiImage: qrImage)
                        .interpolation(.none)
                        .resizable()
                        .frame(width: 200, height: 200)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Text("@\(username)")
                    .font(.system(size: 15, weight: .semibold))
                Text(did)
                    .font(.echomono(11))
                    .foregroundStyle(Color.echoInk55)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)

                Button("Share Link") { viewModel.shareLink() }
                    .font(.system(size: 14, weight: .medium))
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color.echoPaper.ignoresSafeArea())
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .cancellationAction) { Button("Done") { dismiss() } } }
            .onAppear {
                viewModel.configure(did: did, username: username, trustTier: trustTier)
            }
        }
    }
}

#Preview {
    ScrollView {
        VStack(spacing: 20) {
            IdentityCardView(
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
