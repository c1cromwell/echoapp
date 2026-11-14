import Foundation
import os

/// Resolves the Echo REST API base URL for TestFlight and local dev.
///
/// Precedence (matches `SilentProvisionService` / Xcode scheme docs):
/// 1. `ECHO_API_URL` environment variable
/// 2. `API_URL` environment variable
/// 3. `API_URL` in the app Info.plist (build setting substitution)
/// 4. `https://api.echo.local` — a host that does not resolve. Reaching this
///    means the build has no API URL; it is logged as a fault, never silent.
enum EchoAPIBaseURL {

    /// Where the resolved URL came from.
    enum Source: Equatable {
        case environment
        case infoPlist
        /// Nothing usable was configured; `fallback` is in use.
        case unconfigured
    }

    static let fallback = URL(string: "https://api.echo.local")!

    /// Pure resolution, so the precedence and the unconfigured case are testable.
    static func resolve(environment env: [String: String], infoPlistValue: String?) -> (url: URL, source: Source) {
        if let raw = env["ECHO_API_URL"] ?? env["API_URL"],
           let url = usableURL(raw) {
            return (url, .environment)
        }
        if let raw = infoPlistValue, let url = usableURL(raw) {
            return (url, .infoPlist)
        }
        return (fallback, .unconfigured)
    }

    /// True for the placeholder a build ships with until the owner sets the
    /// real production URL (`.invalid` is reserved and never resolves).
    static func isPlaceholder(_ url: URL) -> Bool {
        guard let host = url.host?.lowercased() else { return true }
        return host.hasSuffix(".invalid") || host == fallback.host
    }

    static var resolved: URL { current.url }

    /// False when requests cannot reach a real backend: no `API_URL` at all,
    /// or the placeholder production value is still in the build settings.
    static var isConfigured: Bool { current.source != .unconfigured && !isPlaceholder(current.url) }

    private static let current: (url: URL, source: Source) = {
        let result = resolve(
            environment: ProcessInfo.processInfo.environment,
            infoPlistValue: Bundle.main.object(forInfoDictionaryKey: "API_URL") as? String
        )
        if result.source == .unconfigured {
            Logger(subsystem: "com.echo.app", category: "config").fault(
                "API_URL is not set for this build; using \(fallback.absoluteString, privacy: .public), which does not resolve. Set API_URL in the build settings."
            )
        } else if isPlaceholder(result.url) {
            Logger(subsystem: "com.echo.app", category: "config").fault(
                "API_URL is still the placeholder \(result.url.absoluteString, privacy: .public). Set the production API URL in the Release build settings."
            )
        }
        return result
    }()

    private static func usableURL(_ raw: String) -> URL? {
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        // An unsubstituted `$(API_URL)` means the build setting is missing.
        guard !trimmed.isEmpty, !trimmed.hasPrefix("$("),
              let url = URL(string: trimmed), url.scheme != nil, url.host != nil else {
            return nil
        }
        return url
    }

    /// Builds an absolute URL from a path such as `/v1/auth/sms-recovery/register`.
    static func url(path: String) -> URL {
        let trimmed = path.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return resolved }
        if trimmed.hasPrefix("http://") || trimmed.hasPrefix("https://") {
            return URL(string: trimmed) ?? resolved
        }
        var components = URLComponents(url: resolved, resolvingAgainstBaseURL: false)!
        let suffix = trimmed.hasPrefix("/") ? String(trimmed.dropFirst()) : trimmed
        let basePath = components.path.hasSuffix("/") ? String(components.path.dropLast()) : components.path
        components.path = basePath.isEmpty ? "/\(suffix)" : "\(basePath)/\(suffix)"
        return components.url ?? resolved.appendingPathComponent(suffix)
    }
}

/// Legal and support destinations. One place, read from the app Info.plist
/// (`EchoTermsURL`, `EchoPrivacyURL`, `EchoSupportEmail`) so each product can
/// ship its own. The fallbacks are placeholders the owner must replace.
enum LegalLinks {
    static let placeholderURL = URL(string: "https://example.com")!
    static let placeholderSupportEmail = "support@example.com"

    static var termsURL: URL { url(forInfoKey: "EchoTermsURL", in: infoDictionary) }
    static var privacyURL: URL { url(forInfoKey: "EchoPrivacyURL", in: infoDictionary) }
    static var supportEmail: String { supportEmail(in: infoDictionary) }

    /// `mailto:` link for "Report a Problem" / "Contact Support".
    static var reportProblemURL: URL { mailtoURL(address: supportEmail, subject: "Echo: report a problem") }
    static var contactSupportURL: URL { mailtoURL(address: supportEmail, subject: "Echo support") }

    static func url(forInfoKey key: String, in info: [String: Any]) -> URL {
        guard let raw = (info[key] as? String)?.trimmingCharacters(in: .whitespacesAndNewlines),
              !raw.isEmpty,
              let url = URL(string: raw),
              let scheme = url.scheme?.lowercased(), scheme == "https" || scheme == "http",
              url.host != nil else {
            return placeholderURL
        }
        return url
    }

    static func supportEmail(in info: [String: Any]) -> String {
        guard let raw = (info["EchoSupportEmail"] as? String)?.trimmingCharacters(in: .whitespacesAndNewlines),
              raw.contains("@"), !raw.contains(" ") else {
            return placeholderSupportEmail
        }
        return raw
    }

    static func mailtoURL(address: String, subject: String) -> URL {
        var components = URLComponents()
        components.scheme = "mailto"
        components.path = address
        components.queryItems = [URLQueryItem(name: "subject", value: subject)]
        return components.url ?? URL(string: "mailto:\(placeholderSupportEmail)")!
    }

    private static var infoDictionary: [String: Any] { Bundle.main.infoDictionary ?? [:] }
}


// MARK: - v1 release scope

/// What the v1 App Store build shows. See docs/V1_RELEASE.md.
///
/// ECHO launches as a value-free, gamified reward (no redemption, no transfer,
/// no promise of value). Yield percentages and token-supply internals are a
/// value promise and the pattern App Review rejects under 3.1.5, so Release
/// builds leave them out. Debug builds keep everything for development.
enum V1ReleaseScope {
    #if DEBUG
    static let isFullSurface = true
    #else
    static let isFullSurface = false
    #endif

    /// "% APR" / "APY" figures on staking tiers and validators.
    static var showsYieldRates: Bool { isFullSurface }
    /// Emission gauge and other token-supply internals on the wallet.
    static var showsTokenInternals: Bool { isFullSurface }
}
