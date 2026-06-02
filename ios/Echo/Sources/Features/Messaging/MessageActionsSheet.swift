import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

/// A message action (spec §5.3). Telegram long-press grid + Signal action list.
public enum MessageAction: String, Identifiable, CaseIterable {
    case reply, copy, forward, pin, edit, delete

    public var id: String { rawValue }

    public var title: String {
        switch self {
        case .reply:   return "Reply"
        case .copy:    return "Copy"
        case .forward: return "Forward"
        case .pin:     return "Pin"
        case .edit:    return "Edit"
        case .delete:  return "Delete"
        }
    }

    public var icon: String {
        switch self {
        case .reply:   return "arrowshape.turn.up.left"
        case .copy:    return "doc.on.doc"
        case .forward: return "arrowshape.turn.up.right"
        case .pin:     return "pin"
        case .edit:    return "pencil"
        case .delete:  return "trash"
        }
    }

    public var isDestructive: Bool { self == .delete }
}

/// Bottom-detent action grid for a long-pressed message. `isOwnMessage` and `sentAt`
/// gate Edit (<15 min, own) and Delete-for-everyone (own).
public struct MessageActionsSheet: View {
    let messagePreview: String
    let isOwnMessage: Bool
    let sentWithinEditWindow: Bool
    let onAction: (MessageAction) -> Void

    public init(
        messagePreview: String,
        isOwnMessage: Bool,
        sentWithinEditWindow: Bool = false,
        onAction: @escaping (MessageAction) -> Void
    ) {
        self.messagePreview = messagePreview
        self.isOwnMessage = isOwnMessage
        self.sentWithinEditWindow = sentWithinEditWindow
        self.onAction = onAction
    }

    private var actions: [MessageAction] {
        MessageAction.allCases.filter { action in
            switch action {
            case .edit:   return isOwnMessage && sentWithinEditWindow
            case .delete: return isOwnMessage
            default:      return true
            }
        }
    }

    public var body: some View {
        VStack(spacing: 0) {
            Capsule()
                .fill(Color.echoHair)
                .frame(width: 38, height: 5)
                .padding(.top, 8)
                .padding(.bottom, 12)

            // Message preview strip
            Text(messagePreview)
                .font(.system(size: 13))
                .foregroundColor(.echoInk70)
                .lineLimit(2)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(Spacing.md.rawValue)
                .background(Color.echoPaperDim)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.bottom, Spacing.md.rawValue)

            LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 4), spacing: 4) {
                ForEach(actions) { action in
                    cell(action)
                }
            }
            .padding(.horizontal, Spacing.md.rawValue)
            .padding(.bottom, Spacing.xl.rawValue)
        }
        .background(Color.echoPaper)
    }

    @ViewBuilder
    private func cell(_ action: MessageAction) -> some View {
        Button {
            performSideEffects(action)
            onAction(action)
        } label: {
            VStack(spacing: 6) {
                Image(systemName: action.icon)
                    .font(.system(size: 22))
                Text(action.title)
                    .font(.system(size: 12))
            }
            .foregroundColor(action.isDestructive ? .echoAlert : .echoInk70)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
        }
        .buttonStyle(.plain)
    }

    private func performSideEffects(_ action: MessageAction) {
        guard action == .copy else { return }
        #if canImport(UIKit)
        UIPasteboard.general.string = messagePreview
        #endif
    }
}

#if DEBUG
struct MessageActionsSheet_Previews: PreviewProvider {
    static var previews: some View {
        MessageActionsSheet(
            messagePreview: "Receipt is signed & anchored.",
            isOwnMessage: true,
            sentWithinEditWindow: true,
            onAction: { _ in }
        )
    }
}
#endif
