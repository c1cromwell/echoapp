import SwiftUI
import Combine

// MARK: - Color Palette Protocol

public protocol ColorPalette {
    // Primary
    var primary: Color { get }
    var primaryDark: Color { get }
    var primaryLight: Color { get }
    var primarySubtle: Color { get }

    // Accent
    var accent: Color { get }
    var accentDark: Color { get }
    var accentLight: Color { get }

    // Secondary
    var secondary: Color { get }
    var secondaryDark: Color { get }

    // Semantic
    var success: Color { get }
    var successLight: Color { get }
    var warning: Color { get }
    var warningLight: Color { get }
    var error: Color { get }
    var errorLight: Color { get }

    // Background
    var background: Color { get }
    var backgroundSecondary: Color { get }
    var surface: Color { get }
    var surfaceElevated: Color { get }

    // Text
    var textPrimary: Color { get }
    var textSecondary: Color { get }
    var textTertiary: Color { get }
    var textInverse: Color { get }
    var textOnPrimary: Color { get }

    // Border
    var border: Color { get }
    var borderLight: Color { get }
    var borderFocus: Color { get }

    // Gradients
    var primaryGradient: LinearGradient { get }
    var secondaryGradient: LinearGradient { get }
    var warmGradient: LinearGradient { get }
}

// MARK: - Electric Violet Theme (Default)

public struct ElectricVioletPalette: ColorPalette {
    public let primary = Color(hex: 0x7C3AED)
    public let primaryDark = Color(hex: 0x6D28D9)
    public let primaryLight = Color(hex: 0xA78BFA)
    public let primarySubtle = Color(hex: 0xEDE9FE)

    public let accent = Color(hex: 0x06B6D4)
    public let accentDark = Color(hex: 0x0891B2)
    public let accentLight = Color(hex: 0x67E8F9)

    public let secondary = Color(hex: 0xF43F5E)
    public let secondaryDark = Color(hex: 0xE11D48)

    public let success = Color(hex: 0x10B981)
    public let successLight = Color(hex: 0xD1FAE5)
    public let warning = Color(hex: 0xF59E0B)
    public let warningLight = Color(hex: 0xFEF3C7)
    public let error = Color(hex: 0xEF4444)
    public let errorLight = Color(hex: 0xFEE2E2)

    public let background = Color(hex: 0xFFFFFF)
    public let backgroundSecondary = Color(hex: 0xFAFAFA)
    public let surface = Color(hex: 0xF4F4F5)
    public let surfaceElevated = Color(hex: 0xFFFFFF)

    public let textPrimary = Color(hex: 0x09090B)
    public let textSecondary = Color(hex: 0x52525B)
    public let textTertiary = Color(hex: 0xA1A1AA)
    public let textInverse = Color(hex: 0xFAFAFA)
    public let textOnPrimary = Color(hex: 0xFFFFFF)

    public let border = Color(hex: 0xE4E4E7)
    public let borderLight = Color(hex: 0xF4F4F5)
    public let borderFocus = Color(hex: 0x7C3AED)

    public var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0x7C3AED), Color(hex: 0x06B6D4)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0x06B6D4), Color(hex: 0x10B981)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0xF43F5E), Color(hex: 0xF59E0B)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public init() {}
}

// MARK: - Teal + Warm Neutral Theme

public struct TealWarmPalette: ColorPalette {
    public let primary = Color(hex: 0x0D9488)
    public let primaryDark = Color(hex: 0x0F766E)
    public let primaryLight = Color(hex: 0x2DD4BF)
    public let primarySubtle = Color(hex: 0xCCFBF1)

    public let accent = Color(hex: 0xF59E0B)
    public let accentDark = Color(hex: 0xD97706)
    public let accentLight = Color(hex: 0xFBBF24)

    public let secondary = Color(hex: 0xC4A77D)
    public let secondaryDark = Color(hex: 0xA68B5B)

    public let success = Color(hex: 0x059669)
    public let successLight = Color(hex: 0xD1FAE5)
    public let warning = Color(hex: 0xEA580C)
    public let warningLight = Color(hex: 0xFFEDD5)
    public let error = Color(hex: 0xDC2626)
    public let errorLight = Color(hex: 0xFEE2E2)

    public let background = Color(hex: 0xFFFCF7)
    public let backgroundSecondary = Color(hex: 0xFBF8F3)
    public let surface = Color(hex: 0xF5F0E8)
    public let surfaceElevated = Color(hex: 0xFFFFFF)

    public let textPrimary = Color(hex: 0x1C1917)
    public let textSecondary = Color(hex: 0x57534E)
    public let textTertiary = Color(hex: 0x78716C)
    public let textInverse = Color(hex: 0xFFFCF7)
    public let textOnPrimary = Color(hex: 0xFFFFFF)

    public let border = Color(hex: 0xE7E5E4)
    public let borderLight = Color(hex: 0xF5F5F4)
    public let borderFocus = Color(hex: 0x0D9488)

    public var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0x0D9488), Color(hex: 0x2DD4BF)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0xC4A77D), Color(hex: 0xD4C8B8)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0xF59E0B), Color(hex: 0xEA580C)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public init() {}
}

// MARK: - Icy Minimal Theme

public struct IcyMinimalPalette: ColorPalette {
    public let primary = Color(hex: 0x64748B)
    public let primaryDark = Color(hex: 0x475569)
    public let primaryLight = Color(hex: 0x94A3B8)
    public let primarySubtle = Color(hex: 0xF1F5F9)

    public let accent = Color(hex: 0x0EA5E9)
    public let accentDark = Color(hex: 0x0284C7)
    public let accentLight = Color(hex: 0x7DD3FC)

    public let secondary = Color(hex: 0x78716C)
    public let secondaryDark = Color(hex: 0x57534E)

    public let success = Color(hex: 0x10B981)
    public let successLight = Color(hex: 0xD1FAE5)
    public let warning = Color(hex: 0xF59E0B)
    public let warningLight = Color(hex: 0xFEF3C7)
    public let error = Color(hex: 0xEF4444)
    public let errorLight = Color(hex: 0xFEE2E2)

    public let background = Color(hex: 0xF8FAFC)
    public let backgroundSecondary = Color(hex: 0xF1F5F9)
    public let surface = Color(hex: 0xE2E8F0)
    public let surfaceElevated = Color(hex: 0xFFFFFF)

    public let textPrimary = Color(hex: 0x0F172A)
    public let textSecondary = Color(hex: 0x475569)
    public let textTertiary = Color(hex: 0x64748B)
    public let textInverse = Color(hex: 0xF8FAFC)
    public let textOnPrimary = Color(hex: 0xFFFFFF)

    public let border = Color(hex: 0xE2E8F0)
    public let borderLight = Color(hex: 0xF1F5F9)
    public let borderFocus = Color(hex: 0x0EA5E9)

    public var primaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0x0EA5E9), Color(hex: 0x7DD3FC)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var secondaryGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0x78716C), Color(hex: 0xA8A29E)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public var warmGradient: LinearGradient {
        LinearGradient(
            colors: [Color(hex: 0xF59E0B), Color(hex: 0xFBBF24)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    public init() {}
}

// MARK: - Theme Type

public enum ThemeType: String, CaseIterable, Identifiable {
    case electricViolet = "electric_violet"
    case tealWarm = "teal_warm"
    case icyMinimal = "icy_minimal"

    public var id: String { rawValue }

    public var name: String {
        switch self {
        case .electricViolet: return "Electric Violet"
        case .tealWarm: return "Teal + Warm Neutral"
        case .icyMinimal: return "Icy Minimal"
        }
    }

    public var description: String {
        switch self {
        case .electricViolet: return "Bold & vibrant"
        case .tealWarm: return "Modern & professional"
        case .icyMinimal: return "Fresh & techy"
        }
    }

    public var palette: ColorPalette {
        switch self {
        case .electricViolet: return ElectricVioletPalette()
        case .tealWarm: return TealWarmPalette()
        case .icyMinimal: return IcyMinimalPalette()
        }
    }
}

// MARK: - Theme

public struct Theme {
    public let type: ThemeType
    public let colors: ColorPalette

    public init(type: ThemeType) {
        self.type = type
        self.colors = type.palette
    }
}

// MARK: - Theme Manager

public class ThemeManager: ObservableObject {
    @Published public var currentTheme: Theme

    private let userDefaults = UserDefaults.standard
    private let themeKey = "selected_theme"

    public init() {
        let savedTheme = userDefaults.string(forKey: themeKey) ?? ThemeType.electricViolet.rawValue
        let themeType = ThemeType(rawValue: savedTheme) ?? .electricViolet
        self.currentTheme = Theme(type: themeType)
    }

    public func setTheme(_ type: ThemeType) {
        currentTheme = Theme(type: type)
        userDefaults.set(type.rawValue, forKey: themeKey)
    }
}

// MARK: - Environment Key

public struct ThemeKey: EnvironmentKey {
    public static let defaultValue: Theme = Theme(type: .electricViolet)
}

public extension EnvironmentValues {
    var theme: Theme {
        get { self[ThemeKey.self] }
        set { self[ThemeKey.self] = newValue }
    }
}

// MARK: - View Modifier

public extension View {
    func themed(_ themeManager: ThemeManager) -> some View {
        self.environment(\.theme, themeManager.currentTheme)
    }
}
