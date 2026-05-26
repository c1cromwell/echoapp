#if os(iOS)
import SwiftUI

struct ReactionPickerView: View {
    let onSelect: (String) -> Void
    let onDismiss: () -> Void

    private let emojis = ["👍", "❤️", "😂", "😮", "😢", "🙏"]

    var body: some View {
        HStack(spacing: 8) {
            ForEach(emojis, id: \.self) { emoji in
                Button {
                    onSelect(emoji)
                } label: {
                    Text(emoji)
                        .font(.system(size: 24))
                        .frame(width: 40, height: 40)
                }
                .buttonStyle(GlacialPressStyle())
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 24))
        .ghostBorder(opacity: 0.15)
        .glacialShadow(radius: 16, opacity: 0.08)
    }
}

#Preview {
    ReactionPickerView(onSelect: { _ in }, onDismiss: {})
        .padding()
}
#endif
