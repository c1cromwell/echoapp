import SwiftUI

/// ECHO Design System - Typography
/// Comprehensive type scale from Display to Tiny with semantic styles
public enum TypographyStyle {
    // MARK: - Display & Headings
    
    /// Large display headings (36pt, bold, -0.5% letter spacing)
    case display
    
    /// Heading Level 1 (32pt, bold, -0.3% letter spacing)
    case h1
    
    /// Heading Level 2 (28pt, semibold, -0.2% letter spacing)
    case h2
    
    /// Heading Level 3 (24pt, semibold)
    case h3
    
    /// Heading Level 4 (20pt, semibold)
    case h4
    
    // MARK: - Body Text
    
    /// Body Large (16pt, regular, used for content blocks)
    case bodyLarge
    
    /// Body Regular (14pt, regular, default body text)
    case body
    
    /// Body Small (12pt, regular, secondary content)
    case bodySmall
    
    // MARK: - Semantic Styles
    
    /// Button text (14pt, semibold)
    case button
    
    /// Caption text (12pt, medium, for auxiliary text)
    case caption
    
    /// Tiny text (12pt, regular, for labels)
    case tiny
    
    /// Overline text (12pt, semibold, uppercase for section labels)
    case overline
    
    var font: Font {
        switch self {
        case .display:
            return .system(size: 36, weight: .bold, design: .default)
        case .h1:
            return .system(size: 32, weight: .bold, design: .default)
        case .h2:
            return .system(size: 28, weight: .semibold, design: .default)
        case .h3:
            return .system(size: 24, weight: .semibold, design: .default)
        case .h4:
            return .system(size: 20, weight: .semibold, design: .default)
        case .bodyLarge:
            return .system(size: 16, weight: .regular, design: .default)
        case .body:
            return .system(size: 14, weight: .regular, design: .default)
        case .bodySmall:
            return .system(size: 12, weight: .regular, design: .default)
        case .button:
            return .system(size: 14, weight: .semibold, design: .default)
        case .caption:
            return .system(size: 12, weight: .medium, design: .default)
        case .tiny:
            return .system(size: 12, weight: .regular, design: .default)
        case .overline:
            return .system(size: 12, weight: .semibold, design: .default)
        }
    }
    
    var lineHeight: CGFloat {
        switch self {
        case .display:
            return 44
        case .h1:
            return 40
        case .h2:
            return 36
        case .h3:
            return 32
        case .h4:
            return 28
        case .bodyLarge:
            return 24
        case .body:
            return 22
        case .bodySmall:
            return 18
        case .button:
            return 20
        case .caption:
            return 16
        case .tiny:
            return 16
        case .overline:
            return 16
        }
    }
    
    var letterSpacing: CGFloat {
        switch self {
        case .display:
            return -0.18 // -0.5% of 36pt
        case .h1:
            return -0.10 // -0.3% of 32pt
        case .h2:
            return -0.06 // -0.2% of 28pt
        case .overline:
            return 0.5 // +1% of 12pt
        default:
            return 0
        }
    }
}

// MARK: - Text Style Modifier

struct TextStyleModifier: ViewModifier {
    let style: TypographyStyle
    let color: Color
    
    func body(content: Content) -> some View {
        content
            .font(style.font)
            .tracking(style.letterSpacing)
            .lineSpacing(style.lineHeight - style.font.pointSize)
            .foregroundColor(color)
    }
}

// MARK: - View Extension for Typography

public extension View {
    /// Apply a typography style to the view
    /// - Parameters:
    ///   - style: The TypographyStyle to apply
    ///   - color: The text color (defaults to primary text)
    func typographyStyle(_ style: TypographyStyle, color: Color = .echoPrimaryText) -> some View {
        modifier(TextStyleModifier(style: style, color: color))
    }
}

// MARK: - Preset Text Styles

public struct StyledText: View {
    let text: String
    let style: TypographyStyle
    let color: Color
    
    public init(_ text: String, style: TypographyStyle, color: Color = .echoPrimaryText) {
        self.text = text
        self.style = style
        self.color = color
    }
    
    public var body: some View {
        Text(text)
            .typographyStyle(style, color: color)
    }
}

// MARK: - Common Typography Presets

public extension Text {
    /// Display heading (36pt bold)
    func displayStyle() -> some View {
        self.typographyStyle(.display)
    }
    
    /// Primary heading (32pt bold)
    func h1Style() -> some View {
        self.typographyStyle(.h1)
    }
    
    /// Secondary heading (28pt semibold)
    func h2Style() -> some View {
        self.typographyStyle(.h2)
    }
    
    /// Tertiary heading (24pt semibold)
    func h3Style() -> some View {
        self.typographyStyle(.h3)
    }
    
    /// Quaternary heading (20pt semibold)
    func h4Style() -> some View {
        self.typographyStyle(.h4)
    }
    
    /// Large body text (16pt)
    func bodyLargeStyle() -> some View {
        self.typographyStyle(.bodyLarge)
    }
    
    /// Regular body text (14pt)
    func bodyStyle() -> some View {
        self.typographyStyle(.body)
    }
    
    /// Small body text (12pt)
    func bodySmallStyle() -> some View {
        self.typographyStyle(.bodySmall)
    }
    
    /// Button text (14pt semibold)
    func buttonStyle() -> some View {
        self.typographyStyle(.button)
    }
    
    /// Caption text (12pt medium)
    func captionStyle() -> some View {
        self.typographyStyle(.caption)
    }
}
