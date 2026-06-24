#if os(iOS)
import Foundation
#if canImport(Translation)
import Translation
#endif

/// Apple Translation framework bridge (iOS 18+). Falls back when unavailable.
enum OnDeviceTranslationBridge {
    static func translate(_ text: String, toLanguageCode: String) async -> String? {
        guard !text.isEmpty else { return nil }
        if #available(iOS 18.0, *) {
            return await TranslationBridge_iOS18.translate(text, toLanguageCode: toLanguageCode)
        }
        return nil
    }
}

@available(iOS 18.0, *)
private enum TranslationBridge_iOS18 {
    static func translate(_ text: String, toLanguageCode: String) async -> String? {
        #if canImport(Translation)
        let target = Locale.Language(identifier: toLanguageCode)
        do {
            let session = try TranslationSession(installedSource: nil, target: target)
            let response = try await session.translate(text)
            return response.targetText
        } catch {
            return nil
        }
        #else
        return nil
        #endif
    }
}
#endif
