#if os(iOS)
import Foundation

@MainActor
final class DeviceManagementViewModel: ObservableObject {
    @Published var currentDevice: DeviceSession?
    @Published var otherDevices: [DeviceSession] = []
    @Published var isLoading = true
    @Published var errorMessage: String?

    private let apiClient: AuthAPIClientProtocol
    private let tokenManager: TokenManager
    private var publicKeyHexByDeviceId: [String: String] = [:]

    init(apiClient: AuthAPIClientProtocol, tokenManager: TokenManager) {
        self.apiClient = apiClient
        self.tokenManager = tokenManager
    }

    convenience init() {
        let authAPI = AuthAPIClient(baseURL: EchoAPIBaseURL.resolved)
        let tokens = TokenManager(keychain: KeychainAdapter(), apiClient: authAPI)
        self.init(apiClient: authAPI, tokenManager: tokens)
    }

    func loadDevices() async {
        isLoading = true
        defer { isLoading = false }

        do {
            let token = try await tokenManager.getValidAccessToken()
            let devices = try await apiClient.listDevices(token: token)
            currentDevice = devices.first(where: \.isCurrentDevice)
            otherDevices = devices.filter { !$0.isCurrentDevice }
            await refreshIdentityPublicKeys(for: devices)
        } catch {
            errorMessage = "Could not load devices."
        }
    }

    func revokeDevice(id: String, elevatedToken: String) async -> Bool {
        do {
            try await apiClient.revokeDevice(
                id: id, elevatedToken: elevatedToken
            )
            if let hex = publicKeyHexByDeviceId[id] {
                try? await DIContainer.shared.resolveDeviceHistorySync()?.revokeDevice(publicKeyHex: hex)
            }
            otherDevices.removeAll { $0.id == id }
            publicKeyHexByDeviceId.removeValue(forKey: id)
            return true
        } catch {
            errorMessage = "Could not remove device."
            return false
        }
    }

    func logoutAllDevices() async {
        do {
            await revokeAllRemoteSyncStreams()
            let token = try await tokenManager.getValidAccessToken()
            try await apiClient.logout(token: token, allDevices: true)
            tokenManager.clearTokens()
            BackupSessionKeyStore.clear()
        } catch {
            errorMessage = "Could not log out devices."
        }
    }

    private func revokeAllRemoteSyncStreams() async {
        guard let did = await CurrentUserSession.currentDID(),
              let client = DIContainer.shared.resolveAPIClient(),
              let sync = DIContainer.shared.resolveDeviceHistorySync() else { return }
        let link = DeviceLinkAPIClient(apiClient: client)
        guard let devices = try? await link.listRegisteredDevices(did: did) else { return }
        let currentLabel = currentDevice?.friendlyName.lowercased()
        for device in devices {
            if let label = device.deviceLabel?.lowercased(), label == currentLabel { continue }
            try? await sync.revokeDevice(publicKeyHex: device.publicKeyHex)
        }
    }

    private func refreshIdentityPublicKeys(for sessions: [DeviceSession]) async {
        guard let did = await CurrentUserSession.currentDID(),
              let client = DIContainer.shared.resolveAPIClient() else {
            publicKeyHexByDeviceId = [:]
            return
        }
        let link = DeviceLinkAPIClient(apiClient: client)
        guard let identityDevices = try? await link.listRegisteredDevices(did: did) else {
            publicKeyHexByDeviceId = [:]
            return
        }

        var byLabel: [String: String] = [:]
        for device in identityDevices {
            if let label = device.deviceLabel?.lowercased(), !label.isEmpty {
                byLabel[label] = device.publicKeyHex
            }
        }

        publicKeyHexByDeviceId = sessions.reduce(into: [:]) { map, session in
            if let hex = byLabel[session.friendlyName.lowercased()] {
                map[session.id] = hex
            }
        }
    }
}
#endif
