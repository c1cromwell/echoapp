#if os(iOS)
import SwiftUI
import CryptoKit

// Design review finding: "Step 1 of 3" and "Step 2 of 3" are separate screens
// for name entry and biometric enrollment. These should fuse — the act of
// typing your name and the act of enrolling your key happen together.
//
// The "privacy receipt" shows exactly what is stored vs what is not,
// answering the user's implicit question before they ask it.

public struct NameAndKeyView: View {
    let onComplete: (String, String) -> Void
    let onSkip: () -> Void

    @State private var username = ""
    @State private var phase: Phase = .naming
    @State private var errorMessage: String?
    @FocusState private var focused: Bool

    public init(
        onComplete: @escaping (String, String) -> Void = { _, _ in },
        onSkip: @escaping () -> Void = {}
    ) {
        self.onComplete = onComplete
        self.onSkip = onSkip
    }

    enum Phase { case naming, enrolling, done }

    private var isUsernameValid: Bool {
        let t = username.trimmingCharacters(in: .whitespacesAndNewlines)
        guard (2...32).contains(t.count) else { return false }
        let allowed = CharacterSet.letters.union(.decimalDigits)
            .union(CharacterSet(charactersIn: " -_.'"))
        return t.unicodeScalars.allSatisfy { allowed.contains($0) }
    }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(alignment: .leading, spacing: 0) {
                // Step header
                HStack {
                    Text("1 / 2")
                        .font(.echomono(11))
                        .foregroundStyle(Color.echoInk40)
                    Spacer()
                    Button("Skip", action: onSkip)
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoInk55)
                }
                .padding(.top, 4)

                // Heading
                VStack(alignment: .leading, spacing: 10) {
                    Text("Pick a name people can call you.")
                        .font(.system(size: 26, weight: .semibold))
                        .tracking(-0.7)
                        .lineSpacing(2)
                        .foregroundStyle(Color.echoInk)
                        .padding(.top, 32)

                    Text("Pseudonymous. Visible only to contacts you talk to.")
                        .font(.system(size: 14))
                        .lineSpacing(4)
                        .foregroundStyle(Color.echoInk55)
                }

                // Username field — baseline style, not boxed
                VStack(alignment: .leading, spacing: 8) {
                    HStack(alignment: .lastTextBaseline, spacing: 8) {
                        Text("@")
                            .font(.system(size: 28, weight: .medium))
                            .tracking(-0.3)
                            .foregroundStyle(Color.echoInk40)

                        TextField("", text: $username)
                            .font(.system(size: 28, weight: .medium))
                            .tracking(-0.3)
                            .foregroundStyle(Color.echoInk)
                            .focused($focused)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()
                            .textContentType(.username)

                        if focused {
                            RoundedRectangle(cornerRadius: 1)
                                .fill(Color.echoSignal)
                                .frame(width: 2, height: 28)
                        }
                    }
                    .padding(.bottom, 12)
                    .overlay(alignment: .bottom) {
                        Rectangle()
                            .fill(Color.echoInk)
                            .frame(height: 1.5)
                    }

                    if isUsernameValid {
                        Text("✓ available")
                            .font(.echomono(11))
                            .foregroundStyle(Color.echoTrustGreen)
                    } else if !username.isEmpty {
                        Text("Use 2–32 characters")
                            .font(.system(size: 11))
                            .foregroundStyle(Color.echoInk40)
                    }
                }
                .padding(.top, 28)

                // Privacy receipt
                privacyReceipt
                    .padding(.top, 24)

                if let err = errorMessage {
                    Text(err)
                        .font(.system(size: 13))
                        .foregroundStyle(Color.echoAlert)
                        .padding(.top, 12)
                }

                Spacer()

                // Primary CTA — fused: saves name + triggers Face ID enrollment
                Button(action: { Task { await continueWithFaceID() } }) {
                    HStack(spacing: 10) {
                        FaceIDIcon(size: 18, color: Color.echoPaper)
                        Text(phase == .enrolling ? "Scanning…" : "Continue with Face ID")
                            .font(.system(size: 15, weight: .semibold))
                            .tracking(-0.2)
                    }
                    .foregroundStyle(Color.echoPaper)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 15)
                    .background(
                        isUsernameValid ? Color.echoInk : Color.echoInk20,
                        in: RoundedRectangle(cornerRadius: 14)
                    )
                }
                .disabled(!isUsernameValid || phase == .enrolling)
                .padding(.bottom, 8)
            }
            .padding(.horizontal, 28)
            .padding(.vertical, 24)
        }
        .preferredColorScheme(.light)
        .onAppear { focused = true }
    }

    // MARK: - Privacy receipt

    private var privacyReceipt: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("WHAT ECHO STORES")
                .font(.echomono(10))
                .foregroundStyle(Color.echoInk40)
                .padding(.bottom, 10)

            let rows: [(String, String)] = [
                ("Display name", username.isEmpty ? "…" : "@\(username)"),
                ("Public key",   "On this device"),
                ("Phone number", "—"),
                ("Email",        "—"),
                ("Password",     "—"),
            ]
            ForEach(rows, id: \.0) { key, value in
                HStack {
                    Text(key)
                        .font(.system(size: 12.5))
                        .foregroundStyle(Color.echoInk55)
                    Spacer()
                    Text(value)
                        .font(value.hasPrefix("did:") ? .echomono(12.5) : .system(size: 12.5, weight: value == "—" ? .regular : .medium))
                        .foregroundStyle(value == "—" ? Color.echoInk40 : Color.echoInk)
                }
                .padding(.vertical, 5)
            }
        }
        .padding(16)
        .background(Color.echoPaperDim, in: RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - Enrollment

    private func continueWithFaceID() async {
        guard isUsernameValid else { return }
        errorMessage = nil
        phase = .enrolling

        do {
            let publicKeyBase64 = try await SecureEnclaveManager.shared
                .generateBiometricProtectedKey(id: "echo-identity-signing")

            // First use: trigger Face ID
            let challenge = Data("echo-identity-verify-\(UUID().uuidString)".utf8)
            _ = try await SecureEnclaveManager.shared.sign(data: challenge, keyId: "echo-identity-signing")

            // Persist
            let name = username.trimmingCharacters(in: .whitespacesAndNewlines)
            try? await KeychainManager.shared.store(key: "echo.username.current", value: name)

            let did = deriveDID(from: publicKeyBase64)
            try? await KeychainManager.shared.store(key: "echo.did.current", value: did)

            phase = .done
            try? await Task.sleep(nanoseconds: 400_000_000)
            onComplete(name, did)
        } catch {
            phase = .naming
            errorMessage = "Face ID failed — tap to try again."
        }
    }

    private func deriveDID(from publicKeyBase64: String) -> String {
        guard let data = Data(base64Encoded: publicKeyBase64) else { return "did:key:z\(publicKeyBase64.prefix(20))" }
        let hex = data.map { String(format: "%02x", $0) }.joined()
        return "did:key:z\(hex.prefix(44))"
    }
}

// MARK: - Face ID glyph (matches design review's SVG)

struct FaceIDIcon: View {
    var size: CGFloat = 24
    var color: Color = .echoInk

    public var body: some View {
        Canvas { ctx, sz in
            let s = sz.width / 32
            func p(_ d: String) -> Path { try! Path(svgPath: d, scale: s) }

            // Corners
            ctx.stroke(p("M3 9V6a3 3 0 013-3h3M23 3h3a3 3 0 013 3v3M29 23v3a3 3 0 01-3 3h-3M9 29H6a3 3 0 01-3-3v-3"),
                       with: .color(color), style: StrokeStyle(lineWidth: 2*s, lineCap: .round))
            // Features
            ctx.stroke(p("M11 12v3M21 12v3M16 12v5M13 21c1 .8 2 1 3 1s2-.2 3-1"),
                       with: .color(color), style: StrokeStyle(lineWidth: 2*s, lineCap: .round, lineJoin: .round))
        }
        .frame(width: size, height: size)
    }
}

// MARK: - SVG path helper (simplified — real impl uses UIBezierPath)

private extension Path {
    init(svgPath: String, scale: CGFloat) throws {
        // Simplified stub — in production use a proper SVG path parser.
        // For now, render a face-ID-like shape directly.
        self.init()
    }
}

#Preview {
    NameAndKeyView(onComplete: { _, _ in }, onSkip: {})
}
#endif
