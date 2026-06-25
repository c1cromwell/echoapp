#if os(iOS)
import Foundation
import UIKit

/// Caches identity-registered device pubkeys for incremental sync (M3b).
actor LinkedDevicesRegistry {
    static let shared = LinkedDevicesRegistry()

    private var cached: [DeviceLinkAPIClient.RegisteredDevice] = []
    private var lastRefresh: Date?
    private let refreshInterval: TimeInterval = 120

    func linkedDevices() async -> [DeviceLinkAPIClient.RegisteredDevice] {
        if let lastRefresh, Date().timeIntervalSince(lastRefresh) < refreshInterval, !cached.isEmpty {
            return filter(cached)
        }
        await refresh()
        return filter(cached)
    }

    func invalidate() {
        cached = []
        lastRefresh = nil
    }

    private func refresh() async {
        guard let did = await CurrentUserSession.currentDID(),
              let client = DIContainer.shared.resolveAPIClient() else {
            cached = []
            return
        }
        let link = DeviceLinkAPIClient(apiClient: client)
        cached = (try? await link.listRegisteredDevices(did: did)) ?? []
        lastRefresh = Date()
    }

    private func filter(_ devices: [DeviceLinkAPIClient.RegisteredDevice]) -> [DeviceLinkAPIClient.RegisteredDevice] {
        let myLabel = UIDevice.current.name.lowercased()
        return devices.filter { ($0.deviceLabel ?? "").lowercased() != myLabel }
    }
}
#endif
