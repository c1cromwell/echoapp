#if os(iOS)
import Foundation

/// Local VIP subscription state (Phase 1 — App Store billing deferred to WO-286).
enum VIPSubscriptionStore {
    private static let activeKey = "echo.vip.subscription.active"
    private static let expiresKey = "echo.vip.subscription.expiresAt"
    static let monthlyPriceUSD = Decimal(string: "9.99")!
    static let monthlyPriceLabel = "$9.99"

    static var isActive: Bool {
        guard UserDefaults.standard.bool(forKey: activeKey) else { return false }
        if let expires = expiresAt, expires < Date() {
            return false
        }
        return true
    }

    static var expiresAt: Date? {
        UserDefaults.standard.object(forKey: expiresKey) as? Date
    }

    static var statusLabel: String {
        isActive ? "Active" : "\(monthlyPriceLabel)/mo"
    }

    static func markActive(expiresAt: Date) {
        UserDefaults.standard.set(true, forKey: activeKey)
        UserDefaults.standard.set(expiresAt, forKey: expiresKey)
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: activeKey)
        UserDefaults.standard.removeObject(forKey: expiresKey)
    }
}

@MainActor
enum VIPSubscriptionService {
    enum ActivationError: LocalizedError {
        case notSignedIn
        case networkFailed

        var errorDescription: String? {
            switch self {
            case .notSignedIn: return "Sign in before upgrading to VIP."
            case .networkFailed: return "Could not activate VIP. Check your connection and try again."
            }
        }
    }

    /// Phase 1: records VIP on device + `POST /v1/auth/vip-verify` (no StoreKit yet).
    static func activateMonthly() async throws {
        guard let did = await CurrentUserSession.currentDID(), !did.isEmpty else {
            throw ActivationError.notSignedIn
        }

        let expires = Calendar.current.date(byAdding: .month, value: 1, to: Date()) ?? Date()
        let body: [String: Any] = [
            "did": did,
            "trust_tier": 3,
            "evidence_type": "vip_subscription"
        ]
        var req = URLRequest(url: EchoAPIBaseURL.url(path: "/v1/auth/vip-verify"))
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        req.timeoutInterval = 15

        let (_, response) = try await URLSession.shared.data(for: req)
        guard let http = response as? HTTPURLResponse, (200...299).contains(http.statusCode) else {
            throw ActivationError.networkFailed
        }

        VIPSubscriptionStore.markActive(expiresAt: expires)
        UserDefaults.standard.set(3, forKey: "echo.trustTier")
    }
}
#endif
