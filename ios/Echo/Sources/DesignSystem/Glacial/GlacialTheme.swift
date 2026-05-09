// Core/DesignSystem/GlacialTheme.swift
// The Glacial Interface — ECHO Design System
// Crystalline clarity, frosted glass panes, architectural stillness

import SwiftUI

// MARK: - Color Tokens

extension Color {
    enum Echo {
        // Primary
        static let primary = Color(hex: "#006591")
        static let primaryContainer = Color(hex: "#0EA5E9")  // Signature Sky Blue
        static let onPrimary = Color.white
        static let onPrimaryContainer = Color(hex: "#003751")
        
        // Surface (The Glacial Layers)
        static let surface = Color(hex: "#FAF8FF")
        static let surfaceBright = Color(hex: "#FAF8FF")
        static let surfaceDim = Color(hex: "#D2D9F4")
        static let surfaceContainer = Color(hex: "#EAEDFF")
        static let surfaceContainerLow = Color(hex: "#F2F3FF")
        static let surfaceContainerLowest = Color.white
        static let surfaceContainerHigh = Color(hex: "#E2E7FF")
        static let surfaceContainerHighest = Color(hex: "#DAE2FD")
        static let surfaceVariant = Color(hex: "#DAE2FD")
        
        // On-Surface
        static let onSurface = Color(hex: "#131B2E")
        static let onSurfaceVariant = Color(hex: "#3E4850")
        
        // Secondary
        static let secondary = Color(hex: "#3C627D")
        static let secondaryContainer = Color(hex: "#B8DFFE")
        
        // Outline
        static let outline = Color(hex: "#6E7881")
        static let outlineVariant = Color(hex: "#BEC8D2")
        
        // Inverse
        static let inverseSurface = Color(hex: "#283044")
        static let inverseOnSurface = Color(hex: "#EEF0FF")
        
        // Error
        static let error = Color(hex: "#BA1A1A")
        
        // Signature Gradient Anchors
        static let deepNavy = Color(hex: "#0F172A")
        static let skyBlue = Color(hex: "#0EA5E9")
        static let skyLight = Color(hex: "#7DD3FC")
    }
}

// MARK: - Hex Color Initializer

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: .alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let r, g, b: Double
        switch hex.count {
        case 6:
            (r, g, b) = (Double((int >> 16) & 0xFF) / 255, Double((int >> 8) & 0xFF) / 255, Double(int & 0xFF) / 255)
        default:
            (r, g, b) = (0, 0, 0)
        }
        self.init(red: r, green: g, blue: b)
    }
}

// MARK: - Signature Gradient

extension LinearGradient {
    /// The Glacial Interface signature gradient: Deep Navy → Sky Blue at 135°
    static let signature = LinearGradient(
        colors: [Color.Echo.deepNavy, Color.Echo.skyBlue],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )
}

// MARK: - Typography

extension Font {
    enum Echo {
        static let displayLarge = Font.system(size: 56, weight: .bold)
        static let displayMedium = Font.system(size: 45, weight: .bold)
        static let headlineSm = Font.system(size: 24, weight: .bold)
        static let titleLarge = Font.system(size: 20, weight: .bold)
        static let bodyLarge = Font.system(size: 16, weight: .medium)
        static let bodyMedium = Font.system(size: 14, weight: .medium)
        static let bodySm = Font.system(size: 13, weight: .regular)
        static let labelMd = Font.system(size: 12, weight: .medium)
        static let labelSm = Font.system(size: 10, weight: .bold)
    }
}

// MARK: - Ambient Shadows (Tinted, Never Grey)

extension View {
    /// Soft glacial shadow — tinted with on-surface at 4-8% opacity
    func glacialShadow(radius: CGFloat = 24, opacity: Double = 0.04) -> some View {
        self.shadow(color: Color.Echo.onSurface.opacity(opacity), radius: radius, x: 0, y: 8)
    }
    
    /// Deep shadow for hero elements
    func deepGlacialShadow() -> some View {
        self.shadow(color: Color.Echo.deepNavy.opacity(0.15), radius: 32, x: 0, y: 16)
    }
    
    /// Ghost border — glint on glass edge, never a hard line
    func ghostBorder(opacity: Double = 0.15) -> some View {
        self.overlay(
            RoundedRectangle(cornerRadius: 32)
                .strokeBorder(Color.Echo.outlineVariant.opacity(opacity), lineWidth: 1)
        )
    }
}

// MARK: - Spring Animation (Premium Glass Weight)

extension Animation {
    /// Glacial spring — high damping, mimics weight of premium glass
    static let glacial = Animation.spring(response: 0.5, dampingFraction: 0.85, blendDuration: 0)
    /// Quick glacial press
    static let glacialPress = Animation.spring(response: 0.3, dampingFraction: 0.8, blendDuration: 0)
}
