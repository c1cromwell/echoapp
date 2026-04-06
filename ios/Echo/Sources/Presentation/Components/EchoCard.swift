import SwiftUI

/// ECHO Card Component
/// Standard, Elevated, and Selection variants
public enum CardVariant {
    case standard      // Light background, light border
    case elevated      // Light background, shadow
    case selection     // Can be selected/highlighted
}

public struct EchoCard<Content: View>: View {
    let variant: CardVariant
    let isSelected: Bool
    let onTap: (() -> Void)?
    let content: () -> Content
    
    public init(
        variant: CardVariant = .standard,
        isSelected: Bool = false,
        onTap: (() -> Void)? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.variant = variant
        self.isSelected = isSelected
        self.onTap = onTap
        self.content = content
    }
    
    var backgroundColor: Color {
        switch variant {
        case .standard, .elevated:
            return .echoSurface
        case .selection:
            return isSelected ? Color.echoPrimary.opacity(0.05) : .echoSurface
        }
    }
    
    var borderColor: Color {
        switch variant {
        case .standard:
            return .echoGray200
        case .elevated:
            return .clear
        case .selection:
            return isSelected ? .echoPrimary : .echoGray200
        }
    }
    
    public var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 12)
                .fill(backgroundColor)
                .stroke(borderColor, lineWidth: 1)
            
            content()
        }
        .padding(Spacing.md.rawValue)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(backgroundColor)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(borderColor, lineWidth: 1)
        )
        .conditionalShadow(shouldApply: variant == .elevated)
        .contentShape(Rectangle())
        .onTapGesture {
            if variant == .selection {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) {
                    onTap?()
                }
            } else {
                onTap?()
            }
        }
    }
}

// MARK: - Helper

struct ConditionalShadowModifier: ViewModifier {
    let shouldApply: Bool
    
    func body(content: Content) -> some View {
        if shouldApply {
            content.shadowMd()
        } else {
            content
        }
    }
}

extension View {
    func conditionalShadow(shouldApply: Bool) -> some View {
        modifier(ConditionalShadowModifier(shouldApply: shouldApply))
    }
}

// MARK: - Preset Cards

public struct ContactCard: View {
    let name: String
    let username: String
    let trustLevel: String
    let imageURL: URL?
    
    public init(
        name: String,
        username: String,
        trustLevel: String,
        imageURL: URL? = nil
    ) {
        self.name = name
        self.username = username
        self.trustLevel = trustLevel
        self.imageURL = imageURL
    }
    
    public var body: some View {
        EchoCard(variant: .standard) {
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(
                    imageURL: imageURL,
                    initials: initials,
                    size: .lg,
                    trustLevel: trustLevel
                )
                
                VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                    Text(name)
                        .typographyStyle(.h4, color: .echoPrimaryText)
                    
                    Text("@\(username)")
                        .typographyStyle(.body, color: .echoSecondaryText)
                    
                    TrustBadge(level: trustLevel, size: .small)
                }
                
                Spacer()
            }
        }
    }
    
    private var initials: String {
        name.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

public struct ConversationCard: View {
    let contactName: String
    let lastMessage: String
    let timestamp: String
    let unreadCount: Int
    let imageURL: URL?
    
    public init(
        contactName: String,
        lastMessage: String,
        timestamp: String,
        unreadCount: Int = 0,
        imageURL: URL? = nil
    ) {
        self.contactName = contactName
        self.lastMessage = lastMessage
        self.timestamp = timestamp
        self.unreadCount = unreadCount
        self.imageURL = imageURL
    }
    
    public var body: some View {
        EchoCard(variant: .standard) {
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(
                    imageURL: imageURL,
                    initials: initials,
                    size: .md,
                    status: .online
                )
                
                VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                    HStack {
                        Text(contactName)
                            .typographyStyle(.h4, color: .echoPrimaryText)
                        
                        Spacer()
                        
                        Text(timestamp)
                            .typographyStyle(.caption, color: .echoGray500)
                    }
                    
                    Text(lastMessage)
                        .typographyStyle(.body, color: .echoSecondaryText)
                        .lineLimit(1)
                }
                
                if unreadCount > 0 {
                    VStack {
                        Circle()
                            .fill(Color.echoPrimary)
                            .overlay(
                                Text("\(unreadCount)")
                                    .font(.system(size: 10, weight: .bold))
                                    .foregroundColor(.white)
                            )
                            .frame(width: 24, height: 24)
                        
                        Spacer()
                    }
                }
            }
        }
    }
    
    private var initials: String {
        contactName.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

// MARK: - Preview

#if DEBUG
struct EchoCard_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            EchoCard(variant: .standard) {
                VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                    Text("Standard Card")
                        .typographyStyle(.h4)
                    
                    Text("This is a standard card with a light border")
                        .typographyStyle(.body, color: .echoSecondaryText)
                }
            }
            
            EchoCard(variant: .elevated) {
                VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
                    Text("Elevated Card")
                        .typographyStyle(.h4)
                    
                    Text("This card has a shadow for elevation")
                        .typographyStyle(.body, color: .echoSecondaryText)
                }
            }
            
            ContactCard(
                name: "John Doe",
                username: "johndoe",
                trustLevel: "Verified"
            )
            
            ConversationCard(
                contactName: "Jane Smith",
                lastMessage: "That sounds great! Let's catch up soon.",
                timestamp: "2:45 PM",
                unreadCount: 3
            )
            
            Spacer()
        }
        .echoSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
