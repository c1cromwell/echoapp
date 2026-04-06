import SwiftUI

/// ECHO Trust Badge Component
/// 5 trust-level pill variants
public struct TrustBadge: View {
    let level: String
    let size: BadgeSize
    
    public enum BadgeSize {
        case small
        case medium
        case large
        
        var fontSize: CGFloat {
            switch self {
            case .small:
                return 11
            case .medium:
                return 12
            case .large:
                return 14
            }
        }
        
        var padding: CGFloat {
            switch self {
            case .small:
                return 6
            case .medium:
                return 8
            case .large:
                return 10
            }
        }
    }
    
    public init(level: String, size: BadgeSize = .medium) {
        self.level = level
        self.size = size
    }
    
    var color: Color {
        Color.trustColor(for: level)
    }
    
    var displayText: String {
        level.prefix(1).uppercased() + level.dropFirst()
    }
    
    public var body: some View {
        HStack(spacing: Spacing.xs.rawValue) {
            // Icon
            Image(systemName: iconName)
                .font(.system(size: size.fontSize - 1, weight: .semibold))
            
            // Label
            Text(displayText)
                .font(.system(size: size.fontSize, weight: .semibold))
        }
        .foregroundColor(.white)
        .padding(.horizontal, size.padding * 1.5)
        .padding(.vertical, size.padding)
        .background(color)
        .cornerRadius(16)
        .accessibility(label: Text("\(level) trust level"))
    }
    
    private var iconName: String {
        switch level.lowercased() {
        case "newcomer":
            return "person.crop.circle"
        case "basic":
            return "checkmark.circle"
        case "trusted":
            return "star.circle"
        case "verified":
            return "shield.checkmark"
        case "highlytrusted":
            return "crown.circle"
        default:
            return "circle"
        }
    }
}

// MARK: - Preview

#if DEBUG
struct TrustBadge_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            HStack(spacing: Spacing.md.rawValue) {
                TrustBadge(level: "Newcomer", size: .small)
                TrustBadge(level: "Basic", size: .small)
                TrustBadge(level: "Trusted", size: .small)
                TrustBadge(level: "Verified", size: .small)
                TrustBadge(level: "HighlyTrusted", size: .small)
            }
            
            HStack(spacing: Spacing.md.rawValue) {
                TrustBadge(level: "Newcomer", size: .medium)
                TrustBadge(level: "Basic", size: .medium)
                TrustBadge(level: "Trusted", size: .medium)
            }
            
            HStack(spacing: Spacing.md.rawValue) {
                TrustBadge(level: "Newcomer", size: .large)
                TrustBadge(level: "Verified", size: .large)
            }
            
            Spacer()
        }
        .echoSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
