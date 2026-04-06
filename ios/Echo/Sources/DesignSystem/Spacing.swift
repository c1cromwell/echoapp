import SwiftUI

/// ECHO Design System - Spacing
/// Consistent spacing scale based on 4px base unit
public enum Spacing: CGFloat {
    /// 4pt - Extra small spacing
    case xs = 4
    
    /// 8pt - Small spacing
    case sm = 8
    
    /// 12pt - Medium-small spacing
    case md = 12
    
    /// 16pt - Medium spacing (default padding)
    case lg = 16
    
    /// 20pt - Medium-large spacing
    case xl = 20
    
    /// 24pt - Large spacing
    case xxl = 24
    
    /// 32pt - Extra large spacing
    case xxxl = 32
    
    /// 40pt - 2XL spacing
    case huge = 40
    
    /// 48pt - Extra huge spacing
    case massive = 48
}

// MARK: - SwiftUI Modifier Extensions

public extension View {
    /// Apply padding with semantic spacing
    func echoSpacing(_ spacing: Spacing) -> some View {
        self.padding(spacing.rawValue)
    }
    
    /// Apply horizontal padding only
    func echoHorizontalSpacing(_ spacing: Spacing) -> some View {
        self.padding(.horizontal, spacing.rawValue)
    }
    
    /// Apply vertical padding only
    func echoVerticalSpacing(_ spacing: Spacing) -> some View {
        self.padding(.vertical, spacing.rawValue)
    }
    
    /// Apply spacing to specific edge(s)
    func echoSpacing(_ spacing: Spacing, edges: Edge.Set) -> some View {
        self.padding(edges, spacing.rawValue)
    }
    
    /// Add spacing between content
    func echoSpaced(_ spacing: Spacing) -> some View {
        VStack(spacing: spacing.rawValue) {
            self
        }
    }
}

// MARK: - Common Spacing Patterns

public struct SpacedVStack<Content: View>: View {
    let spacing: Spacing
    let content: () -> Content
    
    public init(spacing: Spacing = .md, @ViewBuilder content: @escaping () -> Content) {
        self.spacing = spacing
        self.content = content
    }
    
    public var body: some View {
        VStack(spacing: spacing.rawValue) {
            content()
        }
    }
}

public struct SpacedHStack<Content: View>: View {
    let spacing: Spacing
    let alignment: VerticalAlignment
    let content: () -> Content
    
    public init(spacing: Spacing = .md, alignment: VerticalAlignment = .center, @ViewBuilder content: @escaping () -> Content) {
        self.spacing = spacing
        self.alignment = alignment
        self.content = content
    }
    
    public var body: some View {
        HStack(spacing: spacing.rawValue, content: content)
    }
}

// MARK: - Spacing Dividers

public struct SpacingDivider: View {
    let size: Spacing
    let axis: Axis = .vertical
    
    public init(_ size: Spacing) {
        self.size = size
    }
    
    public var body: some View {
        if axis == .vertical {
            Spacer().frame(height: size.rawValue)
        } else {
            Spacer().frame(width: size.rawValue)
        }
    }
}

// MARK: - Corner Radius

public enum CornerRadius {
    static let sm: CGFloat = 4
    static let md: CGFloat = 8
    static let lg: CGFloat = 12
    static let xl: CGFloat = 16
    static let xxl: CGFloat = 24
    static let full: CGFloat = 9999
}

// MARK: - Screen Padding

public enum ScreenPadding {
    static let horizontal: CGFloat = 24
    static let vertical: CGFloat = 24
    static let cardPadding: CGFloat = 16
}

// MARK: - Geometry Constants Using Spacing

public extension CGFloat {
    /// Standard button height (56pt = 4 × lg + border)
    static let standardButtonHeight = Spacing.lg.rawValue * 3.5

    /// Standard input field height (44pt)
    static let standardInputHeight = Spacing.lg.rawValue + Spacing.md.rawValue + Spacing.md.rawValue

    /// Standard icon size (24pt)
    static let standardIconSize = Spacing.lg.rawValue + Spacing.xs.rawValue + Spacing.xs.rawValue + Spacing.xs.rawValue + Spacing.xs.rawValue

    /// Standard corner radius (12pt)
    static let standardCornerRadius = Spacing.md.rawValue
}
