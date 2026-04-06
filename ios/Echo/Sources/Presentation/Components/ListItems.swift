import SwiftUI

/// ECHO List Items - Reusable list item components
public struct ContactListItem: View {
    let name: String
    let username: String
    let trustLevel: String
    let imageURL: URL?
    let onTap: () -> Void
    let onSwipe: (() -> Void)?
    
    public init(
        name: String,
        username: String,
        trustLevel: String,
        imageURL: URL? = nil,
        onTap: @escaping () -> Void = {},
        onSwipe: (() -> Void)? = nil
    ) {
        self.name = name
        self.username = username
        self.trustLevel = trustLevel
        self.imageURL = imageURL
        self.onTap = onTap
        self.onSwipe = onSwipe
    }
    
    public var body: some View {
        Button(action: onTap) {
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(
                    imageURL: imageURL,
                    initials: initials,
                    size: .lg,
                    trustLevel: trustLevel
                )
                
                VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                    HStack {
                        Text(name)
                            .typographyStyle(.h4, color: .echoPrimaryText)
                        
                        Spacer()
                        
                        TrustBadge(level: trustLevel, size: .small)
                    }
                    
                    Text("@\(username)")
                        .typographyStyle(.body, color: .echoSecondaryText)
                }
                
                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.echoGray400)
            }
            .padding(Spacing.md.rawValue)
            .background(Color.echoSurface)
            .cornerRadius(12)
        }
        .buttonStyle(PlainButtonStyle())
        .accessibility(element: children: .ignore)
        .accessibility(label: Text("\(name), trust level: \(trustLevel)"))
    }
    
    private var initials: String {
        name.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

public struct ConversationListItem: View {
    let contactName: String
    let lastMessage: String
    let timestamp: String
    let unreadCount: Int
    let imageURL: URL?
    let isOnline: Bool
    let onTap: () -> Void
    
    public init(
        contactName: String,
        lastMessage: String,
        timestamp: String,
        unreadCount: Int = 0,
        imageURL: URL? = nil,
        isOnline: Bool = false,
        onTap: @escaping () -> Void = {}
    ) {
        self.contactName = contactName
        self.lastMessage = lastMessage
        self.timestamp = timestamp
        self.unreadCount = unreadCount
        self.imageURL = imageURL
        self.isOnline = isOnline
        self.onTap = onTap
    }
    
    public var body: some View {
        Button(action: onTap) {
            HStack(spacing: Spacing.md.rawValue) {
                AvatarView(
                    imageURL: imageURL,
                    initials: initials,
                    size: .lg,
                    status: isOnline ? .online : .offline
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
                        .typographyStyle(.body, color: unreadCount > 0 ? .echoPrimaryText : .echoSecondaryText)
                        .fontWeight(unreadCount > 0 ? .semibold : .regular)
                        .lineLimit(1)
                }
                
                if unreadCount > 0 {
                    VStack {
                        Circle()
                            .fill(Color.echoPrimary)
                            .overlay(
                                Text("\(min(unreadCount, 99))")
                                    .font(.system(size: 10, weight: .bold))
                                    .foregroundColor(.white)
                            )
                            .frame(width: 24, height: 24)
                        
                        Spacer()
                    }
                }
            }
            .padding(Spacing.md.rawValue)
            .background(Color.echoSurface)
            .cornerRadius(12)
        }
        .buttonStyle(PlainButtonStyle())
        .accessibility(element: children: .ignore)
        .accessibility(label: Text("\(contactName), \(lastMessage)"))
        if unreadCount > 0 {
            .accessibility(value: Text("\(unreadCount) unread messages"))
        }
    }
    
    private var initials: String {
        contactName.split(separator: " ")
            .prefix(2)
            .map { String($0.first ?? " ") }
            .joined()
    }
}

public struct SettingsListItem: View {
    let icon: Image
    let title: String
    let subtitle: String?
    let value: String?
    let hasToggle: Bool
    @Binding var toggleValue: Bool
    let onTap: () -> Void
    
    public init(
        icon: Image,
        title: String,
        subtitle: String? = nil,
        value: String? = nil,
        hasToggle: Bool = false,
        toggleValue: Binding<Bool> = .constant(false),
        onTap: @escaping () -> Void = {}
    ) {
        self.icon = icon
        self.title = title
        self.subtitle = subtitle
        self.value = value
        self.hasToggle = hasToggle
        self._toggleValue = toggleValue
        self.onTap = onTap
    }
    
    public var body: some View {
        HStack(spacing: Spacing.md.rawValue) {
            icon
                .font(.system(size: 18, weight: .semibold))
                .foregroundColor(.echoPrimary)
                .frame(width: 32)
            
            VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
                Text(title)
                    .typographyStyle(.body, color: .echoPrimaryText)
                
                if let subtitle = subtitle {
                    Text(subtitle)
                        .typographyStyle(.caption, color: .echoSecondaryText)
                }
            }
            
            Spacer()
            
            if hasToggle {
                Toggle("", isOn: $toggleValue)
                    .accessibility(label: Text(title))
            } else if let value = value {
                Text(value)
                    .typographyStyle(.body, color: .echoGray500)
                
                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.echoGray400)
            }
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
        .contentShape(Rectangle())
        .onTapGesture(perform: onTap)
        .accessibility(element: children: .ignore)
        .accessibility(label: Text(title))
        if let subtitle = subtitle {
            .accessibility(hint: Text(subtitle))
        }
    }
}

// MARK: - Preview

#if DEBUG
struct ListItems_Previews: PreviewProvider {
    @State static var notification = true
    
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            VStack(spacing: Spacing.md.rawValue) {
                ContactListItem(
                    name: "John Doe",
                    username: "johndoe",
                    trustLevel: "Verified"
                )
                
                ContactListItem(
                    name: "Jane Smith",
                    username: "janesmith",
                    trustLevel: "Trusted"
                )
            }
            
            VStack(spacing: Spacing.md.rawValue) {
                ConversationListItem(
                    contactName: "Alice Johnson",
                    lastMessage: "That sounds great!",
                    timestamp: "2:45 PM",
                    unreadCount: 3,
                    isOnline: true
                )
                
                ConversationListItem(
                    contactName: "Bob Wilson",
                    lastMessage: "See you tomorrow",
                    timestamp: "Yesterday",
                    unreadCount: 0,
                    isOnline: false
                )
            }
            
            VStack(spacing: Spacing.md.rawValue) {
                SettingsListItem(
                    icon: Image(systemName: "bell.fill"),
                    title: "Notifications",
                    subtitle: "Push notifications",
                    value: "Enabled"
                )
                
                SettingsListItem(
                    icon: Image(systemName: "lock.fill"),
                    title: "Two-Factor Auth",
                    hasToggle: true,
                    toggleValue: $notification
                )
                
                SettingsListItem(
                    icon: Image(systemName: "moon.fill"),
                    title: "Dark Mode",
                    hasToggle: true,
                    toggleValue: $notification
                )
            }
            
            Spacer()
        }
        .echoSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
