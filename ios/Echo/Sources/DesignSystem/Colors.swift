import SwiftUI

/// ECHO Design System - Color Palette
/// Comprehensive color system with trust level support and semantic colors
public extension Color {
    // MARK: - Primary Colors
    
    /// Primary brand color - Indigo
    static let echoPrimary = Color(hex: 0x6366F1)
    
    /// Secondary brand color - Slate
    static let echoSecondary = Color(hex: 0x64748B)
    
    // MARK: - Semantic Colors
    
    /// Success state
    static let echoSuccess = Color(hex: 0x10B981)
    
    /// Warning state
    static let echoWarning = Color(hex: 0xF59E0B)
    
    /// Error state
    static let echoError = Color(hex: 0xEF4444)
    
    /// Information state
    static let echoInfo = Color(hex: 0x3B82F6)
    
    // MARK: - Gray Scale (9 levels)
    
    /// Lightest gray - backgrounds
    static let echoGray50 = Color(hex: 0xF9FAFB)
    
    /// Light gray - subtle backgrounds
    static let echoGray100 = Color(hex: 0xF3F4F6)
    
    /// Very light gray
    static let echoGray200 = Color(hex: 0xE5E7EB)
    
    /// Light gray - borders
    static let echoGray300 = Color(hex: 0xD1D5DB)
    
    /// Medium gray - secondary text
    static let echoGray400 = Color(hex: 0x9CA3AF)
    
    /// Medium gray - default text
    static let echoGray500 = Color(hex: 0x6B7280)
    
    /// Dark gray - primary text
    static let echoGray600 = Color(hex: 0x4B5563)
    
    /// Very dark gray
    static let echoGray700 = Color(hex: 0x374151)
    
    /// Darkest gray - headings
    static let echoGray900 = Color(hex: 0x111827)
    
    // MARK: - Trust Level Colors
    
    /// Newcomer - Red
    static let echoTrustNewcomer = Color(hex: 0xEF4444)
    
    /// Basic - Orange
    static let echoTrustBasic = Color(hex: 0xF59E0B)
    
    /// Trusted - Yellow
    static let echoTrustTrusted = Color(hex: 0xEAB308)
    
    /// Verified - Green
    static let echoTrustVerified = Color(hex: 0x10B981)
    
    /// Highly Trusted - Blue
    static let echoTrustHighlyTrusted = Color(hex: 0x3B82F6)
    
    // MARK: - Functional Colors
    
    /// Dark background (dark mode)
    static let echoDarkBg = Color(hex: 0x0F172A)
    
    /// Dark surface (dark mode)
    static let echoDarkSurface = Color(hex: 0x1E293B)
    
    /// Light background (light mode)
    static let echoLightBg = Color(hex: 0xFFFFFF)
    
    /// Light surface (light mode)
    static let echoLightSurface = Color(hex: 0xF8FAFC)
}

// MARK: - Helper Initializers

extension Color {
    /// Create Color from hex value
    /// - Parameter hex: Hex color value (e.g., 0x6366F1)
    init(hex: UInt32) {
        let r = Double((hex >> 16) & 0xFF) / 255.0
        let g = Double((hex >> 8) & 0xFF) / 255.0
        let b = Double(hex & 0xFF) / 255.0
        self.init(red: r, green: g, blue: b)
    }
    
    /// Get trust level color based on TrustLevel enum
    /// - Parameter level: The trust level to get color for
    /// - Returns: Color for the given trust level
    static func trustColor(for level: String) -> Color {
        switch level.lowercased() {
        case "newcomer":
            return .echoTrustNewcomer
        case "basic":
            return .echoTrustBasic
        case "trusted":
            return .echoTrustTrusted
        case "verified":
            return .echoTrustVerified
        case "highlytrusted":
            return .echoTrustHighlyTrusted
        default:
            return .echoGray400
        }
    }
}

// MARK: - Adaptive Colors (Light/Dark Mode)

extension Color {
    /// Adaptive primary text color
    static let echoPrimaryText = Color(light: .echoGray900, dark: .echoGray50)
    
    /// Adaptive secondary text color
    static let echoSecondaryText = Color(light: .echoGray600, dark: .echoGray300)
    
    /// Adaptive tertiary text color
    static let echoTertiaryText = Color(light: .echoGray500, dark: .echoGray400)
    
    /// Adaptive background color
    static let echoBackground = Color(light: .echoLightBg, dark: .echoDarkBg)
    
    /// Adaptive surface color
    static let echoSurface = Color(light: .echoLightSurface, dark: .echoDarkSurface)
}

// Helper function for adaptive colors
private extension Color {
    init(light: Color, dark: Color) {
        #if os(iOS)
        self.init(uiColor: UIColor(
            light: UIColor(light),
            dark: UIColor(dark)
        ))
        #else
        self = light
        #endif
    }
}
