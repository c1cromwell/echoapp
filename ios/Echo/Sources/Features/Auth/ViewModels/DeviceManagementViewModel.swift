import SwiftUI

@MainActor
final class DeviceManagementViewModel: ObservableObject {
    @Published var currentDevice: DeviceSession?
    @Published var otherDevices: [DeviceSession] = []
    @Published var isLoading = true
    @Published var errorMessage: String?

    private let apiClient: AuthAPIClientProtocol
    private let tokenManager: TokenManager

    init(apiClient: AuthAPIClientProtocol, tokenManager: TokenManager) {
        self.apiClient = apiClient
        self.tokenManager = tokenManager
    }

    func loadDevices() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await tokenManager.getValidAccessToken()
            let devices = try await apiClient.listDevices(token: token)
            currentDevice = devices.first(where: \.isCurrentDevice)
            otherDevices = devices.filter { !$0.isCurrentDevice }
        } catch {
            errorMessage = "Could not load devices."
        }
    }

    func revokeDevice(id: String, elevatedToken: String) async -> Bool {
        do {
            try await apiClient.revokeDevice(
                id: id, elevatedToken: elevatedToken
            )
            otherDevices.removeAll { $0.id == id }
            return true
        } catch {
            errorMessage = "Could not remove device."
            return false
        }
    }

    func logoutAllDevices() async {
        do {
            let token = try await tokenManager.getValidAccessToken()
            try await apiClient.logout(token: token, allDevices: true)
            tokenManager.clearTokens()
        } catch {
            errorMessage = "Could not log out devices."
        }
    }
}
