import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

/// A message action (spec §5.3). Telegram long-press grid + Signal action list.
public enum MessageAction: String, Identifiable, CaseIterable {
    case retry, reply, copy, forward, pin, edit, delete, translate, schedule, report

    public var id: String { rawValue }

    public var title: String {
        switch self {
        case .retry:     return "Try Again"
        case .reply:     return "Reply"
        case .copy:      return "Copy"
        case .forward:   return "Forward"
        case .pin:       return "Pin"
        case .edit:      return "Edit"
        case .delete:    return "Delete"
        case .translate: return "Translate"
        case .schedule:  return "Schedule"
        case .report:    return "Report"
        }
    }

    public var icon: String {
        switch self {
        case .retry:     return "arrow.clockwise"
        case .reply:     return "arrowshape.turn.up.left"
        case .copy:      return "doc.on.doc"
        case .forward:   return "arrowshape.turn.up.right"
        case .pin:       return "pin"
        case .edit:      return "pencil"
        case .delete:    return "trash"
        case .translate: return "character.bubble"
        case .schedule:  return "calendar.badge.clock"
        case .report:    return "exclamationmark.bubble"
        }
    }

    public var isDestructive: Bool { self == .delete || self == .report }

    /// Which actions a long-pressed message offers. Report is for other people's messages only.
    /// A message that failed to send never reached anyone, so it only offers Try Again
    /// (when `canRetry`), Copy and Delete.
    public static func available(
        isOwnMessage: Bool,
        sentWithinEditWindow: Bool,
        showTranslate: Bool,
        isFailed: Bool = false,
        canRetry: Bool = false
    ) -> [MessageAction] {
        if isFailed {
            return canRetry ? [.retry, .copy, .delete] : [.copy, .delete]
        }
        return allCases.filter { action in
            switch action {
            case .retry:     return false
            case .edit:      return isOwnMessage && sentWithinEditWindow
            case .delete:    return isOwnMessage
            case .schedule:  return isOwnMessage
            case .translate: return showTranslate
            case .report:    return !isOwnMessage
            default:         return true
            }
        }
    }
}

/// Bottom-detent action grid for a long-pressed message. `isOwnMessage` and `sentAt`
/// gate Edit (<15 min, own) and Delete-for-everyone (own).
public struct MessageActionsSheet: View {
    let messagePreview: String
    let isOwnMessage: Bool
    let sentWithinEditWindow: Bool
    var showTranslate: Bool = false
    var isFailed: Bool = false
    var canRetry: Bool = false
    let onAction: (MessageAction) -> Void

    /// Delete is permanent (and, for a sent message, removes it for everyone), so it is
    /// confirmed here before `onAction(.delete)` fires.
    @State private var confirmingDelete = false

    public init(
        messagePreview: String,
        isOwnMessage: Bool,
        sentWithinEditWindow: Bool = false,
        showTranslate: Bool = false,
        isFailed: Bool = false,
        canRetry: Bool = false,
        onAction: @escaping (MessageAction) -> Void
    ) {
        self.messagePreview = messagePreview
        self.isOwnMessage = isOwnMessage
        self.sentWithinEditWindow = sentWithinEditWindow
        self.showTranslate = showTranslate
        self.isFailed = isFailed
        self.canRetry = canRetry
        self.onAction = onAction
    }

    private var actions: [MessageAction] {
        MessageAction.available(
            isOwnMessage: isOwnMessage,
            sentWithinEditWindow: sentWithinEditWindow,
            showTranslate: showTranslate,
            isFailed: isFailed,
            canRetry: canRetry
        )
    }

    /// Wording for the delete confirmation: an unsent message only ever lived here.
    static func deleteConfirmation(isFailed: Bool) -> (title: String, message: String) {
        if isFailed {
            return ("Delete this unsent message?", "It was never sent, so it will only be removed from this device.")
        }
        return ("Delete this message for everyone?", "It will be removed from this chat for you and the other person. This can't be undone.")
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
        .confirmationDialog(
            Self.deleteConfirmation(isFailed: isFailed).title,
            isPresented: $confirmingDelete,
            titleVisibility: .visible
        ) {
            Button("Delete", role: .destructive) { onAction(.delete) }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text(Self.deleteConfirmation(isFailed: isFailed).message)
        }
    }

    @ViewBuilder
    private func cell(_ action: MessageAction) -> some View {
        Button {
            if action == .delete {
                confirmingDelete = true
                return
            }
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
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("messageAction.\(action.rawValue)")
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
