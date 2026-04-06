import SwiftUI

import SwiftUI

/// ECHO Message Bubble Component
/// Sent/Received with 7 delivery status indicators (v3.1)
public enum MessageStatus: String {
    case sending = "Sending..."
    case sent = "Sent"
    case delivered = "Delivered"
    case read = "Read"
    case failed = "Failed"
    case anchored = "Anchored"     // Commitment in finalized metagraph snapshot
    case verified = "Verified"     // Digital Evidence fingerprint anchored (Smart Checkmark)
}

public struct MessageBubble: View {
    let message: String
    let isSent: Bool
    let status: MessageStatus
    let timestamp: String
    let showDeliveryStatus: Bool
    
    public init(
        message: String,
        isSent: Bool,
        status: MessageStatus = .sent,
        timestamp: String,
        showDeliveryStatus: Bool = true
    ) {
        self.message = message
        self.isSent = isSent
        self.status = status
        self.timestamp = timestamp
        self.showDeliveryStatus = showDeliveryStatus
    }
    
    var bubbleColor: Color {
        isSent ? .echoPrimary : .echoGray200
    }
    
    var textColor: Color {
        isSent ? .white : .echoPrimaryText
    }
    
    var statusIcon: String? {
        switch status {
        case .sending:
            return "clock"
        case .sent:
            return "checkmark"
        case .delivered:
            return "checkmark.2"
        case .read:
            return "checkmark.2"
        case .failed:
            return "exclamationmark.circle.fill"
        case .anchored:
            return "link"                      // Chain-link icon for on-chain anchoring
        case .verified:
            return "checkmark.seal.fill"       // Smart Checkmark for Digital Evidence
        }
    }

    var statusColor: Color {
        switch status {
        case .failed:
            return .echoError
        case .read:
            return .echoPrimary
        case .anchored:
            return .echoInfo
        case .verified:
            return .echoSuccess
        default:
            return isSent ? .white.opacity(0.7) : .echoGray500
        }
    }
    
    public var body: some View {
        VStack(alignment: isSent ? .trailing : .leading, spacing: Spacing.xs.rawValue) {
            // Message Bubble
            HStack {
                Text(message)
                    .typographyStyle(.body, color: textColor)
                    .lineLimit(nil)
                    .frame(maxWidth: UIScreen.main.bounds.width * 0.7, alignment: .leading)
            }
            .padding(.horizontal, Spacing.md.rawValue)
            .padding(.vertical, Spacing.sm.rawValue)
            .background(bubbleColor)
            .cornerRadius(16)
            .overlay(
                RoundedRectangle(cornerRadius: 16)
                    .stroke(bubbleColor, lineWidth: 0)
            )
            
            // Status and Timestamp
            HStack(spacing: Spacing.xs.rawValue) {
                if isSent {
                    Spacer()
                }
                
                Text(timestamp)
                    .typographyStyle(.caption, color: .echoGray500)
                
                if showDeliveryStatus, let statusIcon = statusIcon {
                    Image(systemName: statusIcon)
                        .font(.system(size: 10, weight: .semibold))
                        .foregroundColor(statusColor)
                        .accessibility(label: Text("Status: \(status.rawValue)"))
                }
                
                if !isSent {
                    Spacer()
                }
            }
            .frame(maxWidth: UIScreen.main.bounds.width * 0.7)
        }
        .frame(maxWidth: .infinity, alignment: isSent ? .trailing : .leading)
        .padding(.horizontal, Spacing.md.rawValue)
    }
}

// MARK: - Chat Bubble Group

public struct MessageBubbleGroup: View {
    let messages: [(String, MessageStatus, String, Bool)]
    
    public init(messages: [(String, MessageStatus, String, Bool)]) {
        self.messages = messages
    }
    
    public var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            ForEach(0..<messages.count, id: \.self) { index in
                let (message, status, timestamp, isSent) = messages[index]
                MessageBubble(
                    message: message,
                    isSent: isSent,
                    status: status,
                    timestamp: timestamp
                )
            }
        }
    }
}

// MARK: - Preview

#if DEBUG
struct MessageBubble_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            // Received messages
            MessageBubble(
                message: "Hey! How are you?",
                isSent: false,
                status: .read,
                timestamp: "10:30 AM"
            )
            
            MessageBubble(
                message: "I'm doing great! Just finished a project. What about you?",
                isSent: false,
                status: .read,
                timestamp: "10:32 AM"
            )
            
            // Sent messages with different statuses
            VStack(spacing: Spacing.sm.rawValue) {
                MessageBubble(
                    message: "Awesome! Let's grab coffee soon.",
                    isSent: true,
                    status: .sending,
                    timestamp: "10:33 AM"
                )
                
                MessageBubble(
                    message: "Sounds perfect!",
                    isSent: true,
                    status: .sent,
                    timestamp: "10:33 AM"
                )
                
                MessageBubble(
                    message: "How about tomorrow at 2 PM?",
                    isSent: true,
                    status: .delivered,
                    timestamp: "10:34 AM"
                )
                
                MessageBubble(
                    message: "Let me check my schedule.",
                    isSent: true,
                    status: .read,
                    timestamp: "10:34 AM"
                )
                
                MessageBubble(
                    message: "I'll send you the details later.",
                    isSent: true,
                    status: .failed,
                    timestamp: "10:35 AM"
                )
            }
            
            Spacer()
        }
        .echoVerticalSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
