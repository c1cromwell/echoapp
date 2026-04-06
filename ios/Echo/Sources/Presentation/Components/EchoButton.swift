import SwiftUI

import SwiftUI

/// Button style enumeration
public enum EchoButtonStyle {
    case primary
    case secondary
    case ghost
    case destructive
}

/// Button size enumeration
public enum EchoButtonSize {
    case large    // 56pt height
    case medium   // 44pt height
    case small    // 36pt height
}

/// ECHO Custom Button Component
/// Supports 4 styles × 3 sizes with loading and disabled states
public struct EchoButton: View {
    let title: String
    let style: EchoButtonStyle
    let size: EchoButtonSize
    let isLoading: Bool
    let isDisabled: Bool
    let icon: Image?
    let action: () -> Void
    
    @State private var isPressed = false
    
    public init(
        _ title: String,
        style: EchoButtonStyle = .primary,
        size: EchoButtonSize = .medium,
        isLoading: Bool = false,
        isDisabled: Bool = false,
        icon: Image? = nil,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.style = style
        self.size = size
        self.isLoading = isLoading
        self.isDisabled = isDisabled
        self.icon = icon
        self.action = action
    }
    
    var height: CGFloat {
        switch size {
        case .large:
            return 56
        case .medium:
            return 44
        case .small:
            return 36
        }
    }
    
    var fontSize: CGFloat {
        switch size {
        case .large:
            return 16
        case .medium:
            return 14
        case .small:
            return 12
        }
    }
    
    var backgroundColor: Color {
        if isDisabled {
            return .echoGray300
        }
        
        switch style {
        case .primary:
            return isPressed ? Color.echoPrimary.opacity(0.8) : .echoPrimary
        case .secondary:
            return isPressed ? Color.echoSecondary.opacity(0.8) : .echoSecondary
        case .ghost:
            return isPressed ? Color.echoGray200 : .transparent
        case .destructive:
            return isPressed ? Color.echoError.opacity(0.8) : .echoError
        }
    }
    
    var foregroundColor: Color {
        if isDisabled {
            return .echoGray500
        }
        
        switch style {
        case .primary, .secondary, .destructive:
            return .white
        case .ghost:
            return .echoPrimaryText
        }
    }
    
    var borderColor: Color? {
        switch style {
        case .ghost:
            return isDisabled ? .echoGray300 : .echoGray300
        default:
            return nil
        }
    }
    
    public var body: some View {
        HStack(spacing: Spacing.sm.rawValue) {
            if !isLoading, let icon = icon {
                icon
                    .font(.system(size: fontSize))
            }
            
            if isLoading {
                ProgressView()
                    .tint(foregroundColor)
            } else {
                Text(title)
                    .font(.system(size: fontSize, weight: .semibold))
            }
        }
        .frame(height: height)
        .frame(maxWidth: .infinity)
        .background(backgroundColor)
        .foregroundColor(foregroundColor)
        .cornerRadius(12)
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(borderColor ?? .clear, lineWidth: 1)
        )
        .opacity(isDisabled && !isLoading ? 0.6 : 1.0)
        .scaleEffect(isPressed && !isDisabled ? 0.98 : 1.0)
        .onLongPressGesture(
            minimumDuration: 0.01,
            pressing: { pressing in
                withAnimation(.easeOut(duration: 0.1)) {
                    isPressed = pressing && !isDisabled
                }
            },
            perform: {
                if !isLoading && !isDisabled {
                    action()
                }
            }
        )
        .accessibilityElement(children: .ignore)
        .accessibility(label: Text(title))
        .accessibility(hint: Text(isLoading ? "Loading" : "Button"))
        .accessibility(enabled: !isDisabled)
    }
}

// MARK: - Preview

#if DEBUG
struct EchoButton_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            // Primary buttons
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Large Primary", style: .primary, size: .large) {}
                EchoButton("Medium Primary", style: .primary, size: .medium) {}
                EchoButton("Small Primary", style: .primary, size: .small) {}
            }
            
            // Secondary buttons
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Secondary", style: .secondary, size: .medium) {}
            }
            
            // Ghost buttons
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Ghost", style: .ghost, size: .medium) {}
            }
            
            // Destructive buttons
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Delete", style: .destructive, size: .medium) {}
            }
            
            // Loading state
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Loading...", style: .primary, size: .medium, isLoading: true) {}
            }
            
            // Disabled state
            VStack(spacing: Spacing.md.rawValue) {
                EchoButton("Disabled", style: .primary, size: .medium, isDisabled: true) {}
            }
            
            Spacer()
        }
        .echoVerticalSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
