import SwiftUI

@MainActor
final class PhoneEntryViewModel: ObservableObject {
    @Published var countryCode = "+1"
    @Published var phoneNumber = ""
    @Published var isLoading = false
    @Published var errorMessage: String?

    var isValid: Bool {
        let digits = phoneNumber.filter(\.isNumber)
        return digits.count >= 10 && digits.count <= 15
    }

    var formattedDisplay: String {
        let digits = phoneNumber.filter(\.isNumber)
        guard digits.count == 10, countryCode == "+1" else { return phoneNumber }
        let area = digits.prefix(3)
        let mid = digits.dropFirst(3).prefix(3)
        let last = digits.suffix(4)
        return "(\(area)) \(mid)-\(last)"
    }

    private let apiClient: AuthAPIClientProtocol
    private let deviceService: DeviceFingerprintService

    init(apiClient: AuthAPIClientProtocol, deviceService: DeviceFingerprintService) {
        self.apiClient = apiClient
        self.deviceService = deviceService
    }

    func sendOTP() async -> (phone: String, verificationId: String)? {
        guard isValid else { return nil }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            let digits = phoneNumber.filter(\.isNumber)
            let response = try await apiClient.registerPhone(
                phone: digits,
                countryCode: countryCode,
                deviceInfo: deviceService.collectDeviceInfo()
            )
            return (phone: "\(countryCode)\(digits)", verificationId: response.verificationId)
        } catch let error as AuthAPIError {
            errorMessage = error.userMessage
            return nil
        } catch {
            errorMessage = "Something went wrong. Please try again."
            return nil
        }
    }
}
