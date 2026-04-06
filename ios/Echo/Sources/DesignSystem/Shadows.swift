import SwiftUI

/// ECHO Design System - Shadows
/// Shadow modifiers for depth and elevation
public extension View {
    /// Small shadow (subtle elevation)
    /// - Shadow radius: 2, Y offset: 2, opacity: 0.05
    func shadowSm() -> some View {
        self.shadow(color: Color.Echo.onSurface.opacity(0.05), radius: 2, x: 0, y: 2)
    }
    
    /// Medium shadow (standard elevation)
    /// - Shadow radius: 6, Y offset: 4, opacity: 0.08
    func shadowMd() -> some View {
        self.shadow(color: Color.Echo.onSurface.opacity(0.08), radius: 6, x: 0, y: 4)
    }
    
    /// Large shadow (prominent elevation)
    /// - Shadow radius: 12, Y offset: 8, opacity: 0.12
    func shadowLg() -> some View {
        self.shadow(color: Color.Echo.onSurface.opacity(0.12), radius: 12, x: 0, y: 8)
    }
    
    /// Primary shadow (primary interactive elements)
    /// - Shadow radius: 8, Y offset: 4, opacity: 0.10, color: brand primary
    func shadowPrimary() -> some View {
        self.shadow(color: Color.echoPrimary.opacity(0.1), radius: 8, x: 0, y: 4)
    }
}

// MARK: - Shadow Elevation Levels

/// Enumeration for standard shadow elevations
public enum EchoShadowLevel {
    /// Subtle shadow (cards, buttons in default state)
    case subtle
    
    /// Standard shadow (cards on hover, elevated content)
    case standard
    
    /// Prominent shadow (modals, popovers)
    case prominent
    
    /// Floating shadow (FAB, floating action items)
    case floating
    
    func apply<Content: View>(to view: Content) -> some View {
        switch self {
        case .subtle:
            return AnyView(view.shadowSm())
        case .standard:
            return AnyView(view.shadowMd())
        case .prominent:
            return AnyView(view.shadowLg())
        case .floating:
            return AnyView(view.shadowPrimary())
        }
    }
}

// MARK: - Shadow Modifier View

struct ShadowModifier: ViewModifier {
    let level: EchoShadowLevel
    
    func body(content: Content) -> some View {
        switch level {
        case .subtle:
            return AnyView(content.shadowSm())
        case .standard:
            return AnyView(content.shadowMd())
        case .prominent:
            return AnyView(content.shadowLg())
        case .floating:
            return AnyView(content.shadowPrimary())
        }
    }
}

public extension View {
    /// Apply an elevation shadow level to the view
    func echoShadow(_ level: EchoShadowLevel) -> some View {
        modifier(ShadowModifier(level: level))
    }
}

// MARK: - Layered Shadow Compound

/// Complex shadow for elevated surfaces
public extension View {
    /// Apply layered shadows for depth (combines multiple shadow layers)
    func shadowElevated() -> some View {
        self
            .shadow(color: Color.Echo.onSurface.opacity(0.05), radius: 2, x: 0, y: 1)
            .shadow(color: Color.Echo.onSurface.opacity(0.08), radius: 4, x: 0, y: 2)
            .shadow(color: Color.Echo.onSurface.opacity(0.12), radius: 8, x: 0, y: 4)
    }
}
