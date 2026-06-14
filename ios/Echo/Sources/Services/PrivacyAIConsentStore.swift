#if os(iOS)
import Foundation

struct PrivacyAIConsent: Codable, Sendable, Equatable {
    var smartRepliesEnabled: Bool
    var summariesEnabled: Bool
    var translationEnabled: Bool

    static let `default` = PrivacyAIConsent(
        smartRepliesEnabled: true,
        summariesEnabled: false,
        translationEnabled: false
    )
}

enum PrivacyAIConsentStore {
    private static let key = "echo.privacy.ai.consent.v1"

    static func load() -> PrivacyAIConsent {
        guard let data = UserDefaults.standard.data(forKey: key),
              let decoded = try? JSONDecoder().decode(PrivacyAIConsent.self, from: data) else {
            return .default
        }
        return decoded
    }

    static func save(_ consent: PrivacyAIConsent) {
        guard let data = try? JSONEncoder().encode(consent) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
