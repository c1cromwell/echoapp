#if os(iOS)
import SwiftUI

/// Reply or edit strip above the chat composer.
struct ChatComposerBanner: View {
    enum Mode {
        case reply(author: String, preview: String)
        case edit
    }

    let mode: Mode
    let onCancel: () -> Void

    var body: some View {
        HStack(spacing: 10) {
            Rectangle()
                .fill(Color.echoSignal)
                .frame(width: 3)
                .clipShape(Capsule())

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.echoSignal)
                if case .reply(_, let preview) = mode {
                    Text(preview)
                        .font(.system(size: 13))
                        .foregroundColor(.echoInk70)
                        .lineLimit(2)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)

            Button(action: onCancel) {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 20))
                    .foregroundColor(.echoInk40)
            }
            .buttonStyle(.plain)
            .accessibilityLabel("Cancel")
        }
        .padding(.horizontal, Spacing.md.rawValue)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
    }

    private var title: String {
        switch mode {
        case .reply: return "Reply"
        case .edit:  return "Edit message"
        }
    }
}

/// Pinned message strip at top of chat.
struct ChatPinnedMessageBanner: View {
    let authorLabel: String
    let preview: String
    let onUnpin: () -> Void
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 10) {
                Image(systemName: "pin.fill")
                    .font(.system(size: 14))
                    .foregroundColor(.echoSignal)
                VStack(alignment: .leading, spacing: 2) {
                    Text(authorLabel)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.echoSignal)
                    Text(preview)
                        .font(.system(size: 13))
                        .foregroundColor(.echoInk70)
                        .lineLimit(1)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                Button(action: onUnpin) {
                    Image(systemName: "xmark")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.echoInk40)
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, Spacing.md.rawValue)
            .padding(.vertical, 8)
            .background(Color.echoPaperDim)
        }
        .buttonStyle(.plain)
    }
}
#endif
