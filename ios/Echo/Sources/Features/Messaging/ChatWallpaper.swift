#if os(iOS)
import SwiftUI

/// Optional per-thread wallpaper (Signal Parity Wave S4) using Echo design tokens only.
enum ChatWallpaperStyle: String, CaseIterable, Identifiable, Sendable {
    case none
    case paperDim
    case signalWash
    case inkMist

    var id: String { rawValue }

    var title: String {
        switch self {
        case .none: return "Default"
        case .paperDim: return "Soft paper"
        case .signalWash: return "Signal wash"
        case .inkMist: return "Ink mist"
        }
    }

    @ViewBuilder
    var background: some View {
        switch self {
        case .none:
            Color.echoPaper
        case .paperDim:
            Color.echoPaperDim
        case .signalWash:
            LinearGradient(
                colors: [Color.echoPaper, Color.echoSignal.opacity(0.08)],
                startPoint: .top,
                endPoint: .bottom
            )
        case .inkMist:
            LinearGradient(
                colors: [Color.echoPaper, Color.echoInk.opacity(0.06)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        }
    }
}

enum ChatWallpaperStore {
    private static func key(_ conversationId: String) -> String {
        "echo.wallpaper.\(conversationId)"
    }

    static func style(for conversationId: String) -> ChatWallpaperStyle {
        guard let raw = UserDefaults.standard.string(forKey: key(conversationId)),
              let style = ChatWallpaperStyle(rawValue: raw) else {
            return .none
        }
        return style
    }

    static func set(_ style: ChatWallpaperStyle, for conversationId: String) {
        UserDefaults.standard.set(style.rawValue, forKey: key(conversationId))
    }
}
#endif
