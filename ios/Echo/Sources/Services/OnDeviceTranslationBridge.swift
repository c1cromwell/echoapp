#if os(iOS)
import Foundation
import SwiftUI
#if canImport(Translation)
import Translation
#endif

/// Apple Translation framework bridge (iOS 18+). Falls back to `nil` when unavailable.
///
/// The Translation framework does **not** expose an app-constructable `TranslationSession`;
/// a session is only vended through the SwiftUI `.translationTask` modifier. So this bridge
/// hands requests to `TranslationGateway`, which drives a configuration that a host modifier
/// mounted at the app root (`.echoTranslationHost()`) turns into a real session. If no host is
/// mounted (e.g. a product without translation, or pre-iOS-18), translation returns `nil` and
/// callers degrade gracefully.
enum OnDeviceTranslationBridge {
    static func translate(_ text: String, toLanguageCode: String) async -> String? {
        guard !text.isEmpty else { return nil }
        #if canImport(Translation)
        if #available(iOS 18.0, *) {
            return await TranslationGateway.shared.translate(text, to: toLanguageCode)
        }
        #endif
        return nil
    }
}

#if canImport(Translation)

/// Bridges the SwiftUI-only Translation API to an `async -> String?` call. MainActor-confined.
@available(iOS 18.0, *)
@MainActor
@Observable
final class TranslationGateway {
    static let shared = TranslationGateway()
    private init() {}

    /// Observed by the host modifier; setting it (re)triggers a `.translationTask`.
    private(set) var configuration: TranslationSession.Configuration?

    private struct Request {
        let id: UUID
        let text: String
        let lang: String
        let cont: CheckedContinuation<String?, Never>
    }
    private var queue: [Request] = []
    private var hostMounted = false

    /// Called by the host modifier as it appears/disappears in the view tree.
    func setHostMounted(_ mounted: Bool) {
        hostMounted = mounted
        if mounted { startNextIfIdle() }
    }

    func translate(_ text: String, to lang: String) async -> String? {
        // No host in the tree means the framework can't vend a session — fail soft.
        guard hostMounted else { return nil }
        let id = UUID()
        return await withCheckedContinuation { (cont: CheckedContinuation<String?, Never>) in
            queue.append(Request(id: id, text: text, lang: lang, cont: cont))
            startNextIfIdle()
            // Safety net: never leave a caller awaiting forever if a session never arrives.
            Task { [weak self] in
                try? await Task.sleep(for: .seconds(12))
                self?.resolveTimeout(id)
            }
        }
    }

    /// Invoked by the host modifier's `.translationTask` once a session exists.
    func run(_ session: TranslationSession) async {
        guard let req = queue.first else { configuration = nil; return }
        queue.removeFirst()
        let translated = try? await session.translate(req.text)
        req.cont.resume(returning: translated?.targetText)
        configuration = nil
        startNextIfIdle()
    }

    private func startNextIfIdle() {
        guard hostMounted, configuration == nil, let next = queue.first else { return }
        configuration = TranslationSession.Configuration(
            target: Locale.Language(identifier: next.lang)
        )
    }

    private func resolveTimeout(_ id: UUID) {
        guard let idx = queue.firstIndex(where: { $0.id == id }) else { return }
        let req = queue.remove(at: idx)
        req.cont.resume(returning: nil)
        if configuration != nil && queue.isEmpty { configuration = nil }
        startNextIfIdle()
    }
}

@available(iOS 18.0, *)
private struct TranslationHostModifier: ViewModifier {
    @State private var gateway = TranslationGateway.shared
    func body(content: Content) -> some View {
        content
            .translationTask(gateway.configuration) { session in
                await gateway.run(session)
            }
            .onAppear { gateway.setHostMounted(true) }
            .onDisappear { gateway.setHostMounted(false) }
    }
}
#endif

extension View {
    /// Mount once near the app root so `OnDeviceTranslationBridge` can vend translation sessions.
    @ViewBuilder
    func echoTranslationHost() -> some View {
        #if canImport(Translation)
        if #available(iOS 18.0, *) {
            self.modifier(TranslationHostModifier())
        } else {
            self
        }
        #else
        self
        #endif
    }
}
#endif
