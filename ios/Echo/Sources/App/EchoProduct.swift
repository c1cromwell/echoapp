#if os(iOS)
import SwiftUI

/// Build-time product identity for the three App Store deliverables.
/// Each Xcode scheme sets `SWIFT_ACTIVE_COMPILATION_CONDITIONS` and a matching Info.plist.
enum EchoProduct: String, Sendable {
    case messaging
    case comply
    case passport

    static var current: EchoProduct {
        #if ECHO_PRODUCT_COMPLY
        return .comply
        #elseif ECHO_PRODUCT_PASSPORT
        return .passport
        #else
        if let raw = Bundle.main.object(forInfoDictionaryKey: "ECHO_PRODUCT") as? String,
           let product = EchoProduct(rawValue: raw) {
            return product
        }
        return .messaging
        #endif
    }

    var displayName: String {
        switch self {
        case .messaging: return "Echo"
        case .comply: return "Echo Comply"
        case .passport: return "Echo Passport"
        }
    }

    var bundleIdentifier: String {
        switch self {
        case .messaging: return "com.echo.app"
        case .comply: return "com.echo.comply"
        case .passport: return "com.echo.passport"
        }
    }
}

enum EchoProductRouter {
    @ViewBuilder
    static func rootView(appState: AppState) -> some View {
        switch EchoProduct.current {
        case .messaging:
            EchoRootView(appState: appState)
        case .comply:
            ComplyCompanionRootView()
        case .passport:
            PassportRootView()
        }
    }
}
#endif
