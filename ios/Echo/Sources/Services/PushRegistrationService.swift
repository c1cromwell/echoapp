#if os(iOS)
import Foundation
import UIKit
import UserNotifications

/// Registers this device for content-blind APNs wake-ups (`POST /v3/notifications/register`).
actor PushRegistrationService {
    static let shared = PushRegistrationService()

    private var registeredTokenHex: String?

    func registerIfNeeded() async {
        let granted = await requestAuthorization()
        guard granted else { return }
        await MainActor.run {
            UIApplication.shared.registerForRemoteNotifications()
        }
    }

    func handleDeviceToken(_ token: Data) async {
        let hex = token.map { String(format: "%02x", $0) }.joined()
        guard hex.count >= 8 else { return }
        if registeredTokenHex == hex { return }
        let deviceLabel = await MainActor.run { UIDevice.current.name }
        guard let client = await MainActor.run(body: { DIContainer.shared.resolveAPIClient() }) else { return }
        struct Body: Encodable {
            let deviceLabel: String
            let publicKey: String
            let apnsToken: String
        }
        let did = await CurrentUserSession.currentDID() ?? ""
        do {
            let _: UserDeviceRegistration = try await client.post(
                endpoint: NotificationsEndpoint.register,
                body: Body(
                    deviceLabel: deviceLabel,
                    publicKey: did,
                    apnsToken: hex
                )
            )
            registeredTokenHex = hex
        } catch {
            // Best-effort: queued messages still flush when the app next opens WS.
        }
    }

    private func requestAuthorization() async -> Bool {
        await withCheckedContinuation { continuation in
            UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, _ in
                continuation.resume(returning: granted)
            }
        }
    }
}

struct UserDeviceRegistration: Decodable, Sendable {
    let deviceId: String?
    enum CodingKeys: String, CodingKey {
        case deviceId = "deviceId"
        case alt = "device_id"
    }
    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        deviceId = try c.decodeIfPresent(String.self, forKey: .deviceId)
            ?? c.decodeIfPresent(String.self, forKey: .alt)
    }
}

enum NotificationsEndpoint: APIEndpoint {
    case register
    var path: String {
        "/v3/notifications/register"
    }
}
#endif
