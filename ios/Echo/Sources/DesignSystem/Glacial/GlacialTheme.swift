// DesignSystem/Glacial/GlacialTheme.swift
// Glacial token bridge — maps the original Color.Echo.* namespace to the
// Design Review warm paper / ink system defined in Colors.swift.
// All 51 existing call sites are updated automatically by this file.

import SwiftUI

// MARK: - Color Tokens (bridged to warm paper / ink system)

extension Color {
    enum Echo {
        // Primary → Signal blue
        static let primary              = Color.echoSignal
        static let primaryContainer     = Color.echoSignal
        static let onPrimary            = Color.white
        static let onPrimaryContainer   = Color.echoInk

        // Surface → Paper surfaces
        static let surface                    = Color.echoPaper
        static let surfaceBright              = Color.echoPaper
        static let surfaceDim                 = Color.echoPaperEdge
        static let surfaceContainer           = Color.echoPaperDim
        static let surfaceContainerLow        = Color.echoPaper
        static let surfaceContainerLowest     = Color.echoPaper
        static let surfaceContainerHigh       = Color.echoPaperDim
        static let surfaceContainerHighest    = Color.echoPaperEdge
        static let surfaceVariant             = Color.echoPaperDim

        // On-surface → Ink
        static let onSurface        = Color.echoInk
        static let onSurfaceVariant = Color.echoInk55

        // Secondary → subdued ink
        static let secondary          = Color.echoInk55
        static let secondaryContainer = Color.echoSignal.opacity(0.12)

        // Outline → hair / ink40
        static let outline        = Color.echoInk40
        static let outlineVariant = Color.echoHair

        // Inverse → Night surface
        static let inverseSurface   = Color.echoNight
        static let inverseOnSurface = Color.echoNightInk

        // Error
        static let error = Color.echoAlert

        // Gradient anchors — night → signal (replaces deepNavy → skyBlue)
        static let deepNavy  = Color.echoNight
        static let skyBlue   = Color.echoSignal
        static let skyLight  = Color.echoSignal.opacity(0.45)
    }
}

// MARK: - Signature Gradient (night → signal)

extension LinearGradient {
    static let signature = LinearGradient(
        colors: [Color.Echo.deepNavy, Color.Echo.skyBlue],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )
}

// MARK: - Typography (spec type scale)

extension Font {
    enum Echo {
        // Display / heading
        static let displayLarge  = Font.system(size: 34, weight: .semibold)  // hero
        static let displayMedium = Font.system(size: 26, weight: .semibold)  // screen title
        static let headlineSm    = Font.system(size: 22, weight: .semibold)  // section heading
        static let titleLarge    = Font.system(size: 20, weight: .semibold)  // card titles

        // Body
        static let bodyLarge  = Font.system(size: 16, weight: .regular)
        static let bodyMedium = Font.system(size: 15, weight: .regular)
        static let bodySm     = Font.system(size: 13, weight: .regular)

        // Labels
        static let labelMd = Font.system(size: 12, weight: .medium)
        static let labelSm = Font.system(size: 11, weight: .semibold)
    }
}

// MARK: - Shadows (tinted with ink)

extension View {
    func glacialShadow(radius: CGFloat = 16, opacity: Double = 0.05) -> some View {
        self.shadow(color: Color.echoInk.opacity(opacity), radius: radius, x: 0, y: 4)
    }

    func deepGlacialShadow() -> some View {
        self.shadow(color: Color.echoInk.opacity(0.08), radius: 24, x: 0, y: 8)
    }

    func ghostBorder(opacity: Double = 0.15) -> some View {
        self.overlay(
            RoundedRectangle(cornerRadius: 32)
                .strokeBorder(Color.echoHair.opacity(opacity), lineWidth: 1)
        )
    }
}

// MARK: - Hex String Color Initializer (backward compat for Color(hex: "#RRGGBB"))

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: .alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let r = Double((int >> 16) & 0xFF) / 255.0
        let g = Double((int >>  8) & 0xFF) / 255.0
        let b = Double( int        & 0xFF) / 255.0
        self.init(red: r, green: g, blue: b)
    }
}

// MARK: - Spring Animation

extension Animation {
    static let glacial      = Animation.spring(response: 0.38, dampingFraction: 0.82)
    static let glacialPress = Animation.spring(response: 0.3,  dampingFraction: 0.8)
}
