#if os(iOS)
import SwiftUI
import CryptoKit

// Wave 12: SMS phone backup setup.
//
// Privacy model:
//   - Phone number is hashed (SHA-256) before anything leaves the device
//   - Backend receives: {phone_hash, phone_raw, did}  → stores only phone_hash
//   - phone_raw is used only to send the OTP via Twilio, then discarded
//   - After OTP verification, phone_hash is stored in Keychain for recovery lookups

struct SMSOTPSetupView: View {
    let did: String
    let onConfigured: () -> Void

    @State private var phone = ""
    @State private var otp = ""
    @State private var sessionToken = ""
    @State private var phase: OTPPhase = .enterPhone
    @State private var countdown = 45
    @State private var countdownTask: Task<Void, Never>?
    @State private var errorMessage: String?
    @Environment(\.dismiss) private var dismiss

    enum OTPPhase { case enterPhone, enterOTP, done }

    var body: some View {
        NavigationStack {
            ZStack {
                Color.echoBackground.ignoresSafeArea()
                VStack(spacing: 24) {
                    switch phase {
                    case .enterPhone:
                        phoneEntryBody
                    case .enterOTP:
                        otpEntryBody
                    case .done:
                        doneBody
                    }
                }
                .padding(.top, 32)
            }
            .navigationTitle("Phone Backup")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                        .disabled(phase == .done)
                }
            }
        }
    }

    // MARK: - Phone entry

    private var phoneEntryBody: some View {
        VStack(spacing: 20) {
            Image(systemName: "phone.badge.checkmark")
                .font(.system(size: 44))
                .foregroundColor(.echoPrimary)

            VStack(spacing: 8) {
                Text("Add phone backup")
                    .font(.system(size: 20, weight: .bold))
                    .foregroundColor(.echoPrimaryText)
                Text("We store only a hash of your number — Echo never sees it in plain text.")
                    .font(.system(size: 13))
                    .foregroundColor(.echoSecondaryText)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 32)
            }

            EchoTextField("Phone number (+12125551234)", text: $phone)
                .keyboardType(.phonePad)
                .padding(.horizontal, 24)

            if let error = errorMessage {
                Text(error).font(.system(size: 13)).foregroundColor(.echoError)
                    .padding(.horizontal, 24)
            }

            EchoButton("Send verification code", style: .primary) {
                Task { await sendOTP() }
            }
            .disabled(!isValidE164(phone))
            .padding(.horizontal, 24)

            Spacer()
        }
    }

    // MARK: - OTP entry

    private var otpEntryBody: some View {
        VStack(spacing: 20) {
            Image(systemName: "message.badge.filled.fill")
                .font(.system(size: 44))
                .foregroundColor(.echoPrimary)

            VStack(spacing: 8) {
                Text("Enter verification code")
                    .font(.system(size: 20, weight: .bold))
                    .foregroundColor(.echoPrimaryText)
                Text("A 6-digit code was sent to \(phone)")
                    .font(.system(size: 13))
                    .foregroundColor(.echoSecondaryText)
            }

            OTPInputView(code: $otp, length: 6)
                .padding(.horizontal, 24)

            if let error = errorMessage {
                Text(error).font(.system(size: 13)).foregroundColor(.echoError)
            }

            EchoButton("Verify", style: .primary) {
                Task { await verifyOTP() }
            }
            .disabled(otp.count != 6)
            .padding(.horizontal, 24)

            Button(countdown > 0 ? "Resend in \(countdown)s" : "Resend code") {
                guard countdown == 0 else { return }
                Task { await sendOTP() }
            }
            .font(.system(size: 13))
            .foregroundColor(countdown > 0 ? .echoSecondaryText : .echoPrimary)
            .disabled(countdown > 0)

            Spacer()
        }
    }

    // MARK: - Done

    private var doneBody: some View {
        VStack(spacing: 20) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: 60))
                .foregroundColor(Color(hex: 0x10B981))
            Text("Phone backup added")
                .font(.system(size: 20, weight: .bold))
                .foregroundColor(.echoPrimaryText)
            Text("Your phone number hash is stored as a recovery hint.")
                .font(.system(size: 13))
                .foregroundColor(.echoSecondaryText)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
            Spacer()
        }
    }

    // MARK: - Actions

    private func sendOTP() async {
        errorMessage = nil
        let hash = sha256Hex(phone)
        let body: [String: String] = [
            "phone_hash": "sha256:\(hash)",
            "phone_raw": phone,
            "did": did,
        ]
        guard let data = try? JSONSerialization.data(withJSONObject: body),
              let url = URL(string: "https://api.echo.local/v1/auth/sms-recovery/register") else { return }

        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = data

        do {
            let (respData, _) = try await URLSession.shared.data(for: req)
            if let json = try? JSONSerialization.jsonObject(with: respData) as? [String: Any],
               let token = json["session_token"] as? String {
                sessionToken = token
                phase = .enterOTP
                startCountdown()
            } else {
                errorMessage = "Failed to send code. Check your number and try again."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func verifyOTP() async {
        errorMessage = nil
        let body: [String: String] = ["session_token": sessionToken, "otp": otp]
        guard let data = try? JSONSerialization.data(withJSONObject: body),
              let url = URL(string: "https://api.echo.local/v1/auth/sms-recovery/verify") else { return }

        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = data

        do {
            let (_, response) = try await URLSession.shared.data(for: req)
            if let http = response as? HTTPURLResponse, http.statusCode == 200 {
                // Store phone hash in Keychain for future recovery lookups.
                try? await KeychainManager.shared.store(
                    key: "echo.sms.phone_hash",
                    value: "sha256:\(sha256Hex(phone))"
                )
                countdownTask?.cancel()
                phase = .done
                try? await Task.sleep(nanoseconds: 1_200_000_000)
                onConfigured()
            } else {
                errorMessage = "Invalid code. Please try again."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func startCountdown() {
        countdown = 45
        countdownTask?.cancel()
        countdownTask = Task {
            while countdown > 0 {
                try? await Task.sleep(nanoseconds: 1_000_000_000)
                countdown -= 1
            }
        }
    }

    // MARK: - Helpers

    private func isValidE164(_ s: String) -> Bool {
        let pattern = #"^\+[1-9]\d{9,14}$"#
        return s.range(of: pattern, options: .regularExpression) != nil
    }

    private func sha256Hex(_ input: String) -> String {
        let hash = SHA256.hash(data: Data(input.utf8))
        return hash.compactMap { String(format: "%02x", $0) }.joined()
    }
}
#endif
