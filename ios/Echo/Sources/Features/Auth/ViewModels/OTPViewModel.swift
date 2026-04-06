import SwiftUI
import Combine

@MainActor
final class OTPViewModel: ObservableObject {
    @Published var code: [String] = Array(repeating: "", count: 6)
    @Published var focusedIndex = 0
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var resendCountdown = 60
    @Published var canResend = false

    let phoneNumber: String
    let verificationId: String

    private let apiClient: AuthAPIClientProtocol
    private let deviceService: DeviceFingerprintService
    private var timer: AnyCancellable?

    var fullCode: String { code.joined() }
    var isComplete: Bool { fullCode.count == 6 && fullCode.allSatisfy(\.isNumber) }

    init(
        phoneNumber: String,
        verificationId: String,
        apiClient: AuthAPIClientProtocol,
        deviceService: DeviceFingerprintService
    ) {
        self.phoneNumber = phoneNumber
        self.verificationId = verificationId
        self.apiClient = apiClient
        self.deviceService = deviceService
        startResendTimer()
    }

    func verifyOTP() async -> String? {
        guard isComplete else { return nil }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let response = try await apiClient.verifyOTP(
                verificationId: verificationId,
                code: fullCode,
                deviceInfo: deviceService.collectDeviceInfo()
            )
            return response.accessToken  // Temp token for passkey registration
        } catch let error as AuthAPIError {
            errorMessage = error.userMessage
            code = Array(repeating: "", count: 6)
            focusedIndex = 0
            return nil
        } catch {
            errorMessage = "Verification failed. Please try again."
            return nil
        }
    }

    func resendCode() async {
        guard canResend else { return }
        canResend = false
        resendCountdown = 60
        startResendTimer()
        _ = try? await apiClient.registerPhone(
            phone: phoneNumber.filter(\.isNumber),
            countryCode: String(phoneNumber.prefix(2)),
            deviceInfo: deviceService.collectDeviceInfo()
        )
    }

    func handleDigitInput(at index: Int, value: String) {
        if value.count == 1 && index < 5 {
            focusedIndex = index + 1
        }
    }

    private func startResendTimer() {
        timer = Timer.publish(every: 1, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                guard let self else { return }
                if self.resendCountdown > 0 {
                    self.resendCountdown -= 1
                } else {
                    self.canResend = true
                    self.timer?.cancel()
                }
            }
    }
}
