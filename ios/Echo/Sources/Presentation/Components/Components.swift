import SwiftUI

// MARK: - Primary Button

struct PrimaryButton: View {
    let title: String
    let action: () -> Void
    var isLoading: Bool = false
    var isEnabled: Bool = true
    
    var body: some View {
        Button(action: action) {
            HStack {
                if isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    Text(title)
                        .fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 48)
            .background(isEnabled ? Color.blue : Color.gray)
            .foregroundColor(.white)
            .cornerRadius(12)
        }
        .disabled(!isEnabled || isLoading)
    }
}

// MARK: - Secondary Button

struct SecondaryButton: View {
    let title: String
    let action: () -> Void
    var isLoading: Bool = false
    var isEnabled: Bool = true
    
    var body: some View {
        Button(action: action) {
            HStack {
                if isLoading {
                    ProgressView()
                        .tint(.blue)
                } else {
                    Text(title)
                        .fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 48)
            .background(Color(.systemGray6))
            .foregroundColor(.blue)
            .cornerRadius(12)
        }
        .disabled(!isEnabled || isLoading)
    }
}

// MARK: - Text Input Field

struct TextInputField: View {
    let label: String
    let placeholder: String
    @Binding var text: String
    var isSecure: Bool = false
    var keyboardType: UIKeyboardType = .default
    var error: String?
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(label)
                .font(.subheadline)
                .fontWeight(.medium)
            
            if isSecure {
                SecureField(placeholder, text: $text)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .keyboardType(keyboardType)
            } else {
                TextField(placeholder, text: $text)
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .keyboardType(keyboardType)
            }
            
            if let error = error {
                Text(error)
                    .font(.caption)
                    .foregroundColor(.red)
            }
        }
    }
}

// MARK: - Card

struct CardView<Content: View>: View {
    let content: Content
    var cornerRadius: CGFloat = 12
    var padding: CGFloat = 16
    
    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }
    
    var body: some View {
        content
            .padding(padding)
            .background(Color(.systemBackground))
            .cornerRadius(cornerRadius)
            .shadow(color: Color.Echo.onSurface.opacity(0.1), radius: 8, x: 0, y: 2)
    }
}

// MARK: - Trust Badge

struct TrustBadge: View {
    let level: TrustLevel
    var size: CGFloat = 32
    
    var backgroundColor: Color {
        switch level {
        case .newcomer: return Color.gray
        case .basic: return Color.green
        case .trusted: return Color.blue
        case .verified: return Color.purple
        case .elite: return Color.orange
        }
    }
    
    var displayName: String {
        level.rawValue.capitalized
    }
    
    var body: some View {
        VStack(spacing: 4) {
            Circle()
                .fill(backgroundColor)
                .frame(width: size, height: size)
                .overlay(
                    Text(String(level.rawValue.first!))
                        .font(.system(size: size * 0.4, weight: .bold))
                        .foregroundColor(.white)
                )
            
            Text(displayName)
                .font(.caption2)
                .fontWeight(.semibold)
        }
    }
}

// MARK: - Verification Badge

struct VerificationBadge: View {
    let type: VerificationType
    let status: VerificationStatus
    var showLabel: Bool = true
    
    var icon: String {
        switch type {
        case .email: return "envelope.fill"
        case .phone: return "phone.fill"
        case .identity: return "person.fill"
        case .business: return "briefcase.fill"
        }
    }
    
    var color: Color {
        switch status {
        case .verified: return .green
        case .pending: return .yellow
        case .rejected: return .red
        case .expired: return .gray
        }
    }
    
    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 12, weight: .semibold))
            
            if showLabel {
                Text(type.rawValue.capitalized)
                    .font(.caption)
                    .fontWeight(.medium)
            }
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(color.opacity(0.2))
        .foregroundColor(color)
        .cornerRadius(6)
    }
}

// MARK: - Achievement Card

struct AchievementCard: View {
    let achievement: Achievement
    var body: some View {
        VStack(spacing: 8) {
            if achievement.isUnlocked {
                Image(systemName: achievement.icon)
                    .font(.system(size: 32))
                    .foregroundColor(.yellow)
            } else {
                Image(systemName: achievement.icon)
                    .font(.system(size: 32))
                    .foregroundColor(.gray)
                    .opacity(0.5)
            }
            
            VStack(spacing: 4) {
                Text(achievement.name)
                    .font(.caption)
                    .fontWeight(.semibold)
                    .lineLimit(1)
                
                Text(achievement.description)
                    .font(.caption2)
                    .foregroundColor(.secondary)
                    .lineLimit(2)
                    .multilineTextAlignment(.center)
            }
        }
        .padding(12)
        .frame(maxWidth: .infinity)
        .background(Color(.systemGray6))
        .cornerRadius(12)
        .opacity(achievement.isUnlocked ? 1 : 0.5)
    }
}

// MARK: - Message Bubble

struct MessageBubble: View {
    let message: Message
    let isCurrentUser: Bool
    
    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            if isCurrentUser {
                Spacer()
            } else {
                if let avatar = message.sender.avatar {
                    AsyncImage(url: URL(string: avatar)) { phase in
                        switch phase {
                        case .success(let image):
                            image
                                .resizable()
                                .scaledToFill()
                                .frame(width: 32, height: 32)
                                .cornerRadius(16)
                        default:
                            Circle()
                                .fill(Color.gray)
                                .frame(width: 32, height: 32)
                        }
                    }
                }
            }
            
            VStack(alignment: isCurrentUser ? .trailing : .leading, spacing: 4) {
                if !isCurrentUser {
                    Text(message.sender.username)
                        .font(.caption)
                        .fontWeight(.semibold)
                        .foregroundColor(.secondary)
                }
                
                Text(message.content)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(isCurrentUser ? Color.blue : Color(.systemGray5))
                    .foregroundColor(isCurrentUser ? .white : Color.Echo.onSurface)
                    .cornerRadius(12)
                
                Text(message.timestamp, style: .time)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            
            if !isCurrentUser {
                Spacer()
            }
        }
        .padding(.horizontal)
    }
}

// MARK: - User List Item

struct UserListItem: View {
    let user: User
    let isOnline: Bool
    var onTap: (() -> Void)?
    
    var body: some View {
        Button(action: { onTap?() }) {
            HStack(spacing: 12) {
                if let avatar = user.avatar {
                    AsyncImage(url: URL(string: avatar)) { phase in
                        switch phase {
                        case .success(let image):
                            image
                                .resizable()
                                .scaledToFill()
                                .frame(width: 48, height: 48)
                                .cornerRadius(24)
                        default:
                            Circle()
                                .fill(Color.gray)
                                .frame(width: 48, height: 48)
                        }
                    }
                } else {
                    Circle()
                        .fill(Color.gray)
                        .frame(width: 48, height: 48)
                }
                
                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 4) {
                        Text(user.username)
                            .font(.subheadline)
                            .fontWeight(.semibold)
                            .foregroundColor(Color.Echo.onSurface)
                        
                        if user.isVerified {
                            Image(systemName: "checkmark.circle.fill")
                                .font(.caption)
                                .foregroundColor(.blue)
                        }
                    }
                    
                    Text(isOnline ? "Online" : "Offline")
                        .font(.caption)
                        .foregroundColor(isOnline ? .green : .secondary)
                }
                
                Spacer()
                
                if isOnline {
                    Circle()
                        .fill(Color.green)
                        .frame(width: 8, height: 8)
                }
            }
            .padding()
            .background(Color(.systemGray6))
            .cornerRadius(12)
        }
    }
}

// MARK: - Conversation Cell

struct ConversationCell: View {
    let conversation: Conversation
    var onTap: (() -> Void)?
    
    var body: some View {
        Button(action: { onTap?() }) {
            HStack(spacing: 12) {
                Circle()
                    .fill(Color.blue.opacity(0.3))
                    .frame(width: 48, height: 48)
                    .overlay(
                        Text(String(conversation.displayName.prefix(1)))
                            .font(.headline)
                            .foregroundColor(.blue)
                    )
                
                VStack(alignment: .leading, spacing: 4) {
                    Text(conversation.displayName)
                        .font(.subheadline)
                        .fontWeight(.semibold)
                        .foregroundColor(Color.Echo.onSurface)
                        .lineLimit(1)
                    
                    if let lastMessage = conversation.lastMessage {
                        Text(lastMessage.content)
                            .font(.caption)
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                    }
                }
                
                Spacer()
                
                VStack(alignment: .trailing, spacing: 4) {
                    if conversation.unreadCount > 0 {
                        Text("\(conversation.unreadCount)")
                            .font(.caption)
                            .fontWeight(.semibold)
                            .foregroundColor(.white)
                            .frame(width: 20, height: 20)
                            .background(Color.blue)
                            .cornerRadius(10)
                    }
                    
                    if let timestamp = conversation.lastMessage?.timestamp {
                        Text(timestamp, style: .date)
                            .font(.caption2)
                            .foregroundColor(.secondary)
                    }
                }
            }
            .padding()
            .background(Color(.systemBackground))
        }
    }
}

// MARK: - Loading Indicator

struct LoadingIndicator: View {
    var body: some View {
        VStack(spacing: 16) {
            ProgressView()
                .tint(.blue)
            Text("Loading...")
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(.systemBackground))
    }
}

// MARK: - Empty State

struct EmptyStateView: View {
    let icon: String
    let title: String
    let subtitle: String
    
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: icon)
                .font(.system(size: 48))
                .foregroundColor(.gray)
            
            Text(title)
                .font(.headline)
                .fontWeight(.semibold)
            
            Text(subtitle)
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(.systemBackground))
    }
}

// MARK: - Error Banner

struct ErrorBanner: View {
    let message: String
    var onDismiss: (() -> Void)?
    
    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: "exclamationmark.circle.fill")
                .foregroundColor(.white)
            
            Text(message)
                .font(.subheadline)
                .foregroundColor(.white)
            
            Spacer()
            
            Button(action: { onDismiss?() }) {
                Image(systemName: "xmark")
                    .foregroundColor(.white)
            }
        }
        .padding()
        .background(Color.red)
        .cornerRadius(8)
    }
}

// MARK: - Success Banner

struct SuccessBanner: View {
    let message: String
    var onDismiss: (() -> Void)?
    
    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: "checkmark.circle.fill")
                .foregroundColor(.white)
            
            Text(message)
                .font(.subheadline)
                .foregroundColor(.white)
            
            Spacer()
            
            Button(action: { onDismiss?() }) {
                Image(systemName: "xmark")
                    .foregroundColor(.white)
            }
        }
        .padding()
        .background(Color.green)
        .cornerRadius(8)
    }
}

// MARK: - Scale Button Style

public struct ScaleButtonStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
            .animation(.spring(response: 0.3, dampingFraction: 0.6), value: configuration.isPressed)
    }
}

// MARK: - Section Divider

public struct SectionDivider: View {
    @Environment(\.theme) var theme
    public let title: String

    public init(title: String) {
        self.title = title
    }

    public var body: some View {
        HStack(spacing: 12) {
            Rectangle()
                .fill(theme.colors.border)
                .frame(height: 1)

            Text(title)
                .font(.system(size: 12, weight: .semibold))
                .foregroundColor(theme.colors.textTertiary)
                .tracking(1)

            Rectangle()
                .fill(theme.colors.border)
                .frame(height: 1)
        }
        .padding(.horizontal, ScreenPadding.horizontal)
        .padding(.vertical, 8)
    }
}
