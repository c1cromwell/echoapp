#if os(iOS)
import SwiftUI

struct ReactionPickerView: View {
    let onSelect: (String) -> Void
    let onDismiss: () -> Void

    private let quickEmojis = ["👍", "❤️", "😂", "😮", "😢", "🙏", "🔥", "🎉", "👏", "💯"]
    private let recentKey = "echo.recentReactionEmojis"

    private var displayEmojis: [String] {
        let recent = UserDefaults.standard.stringArray(forKey: recentKey) ?? []
        let merged = recent + quickEmojis
        var seen = Set<String>()
        return merged.filter { seen.insert($0).inserted }
    }

    var body: some View {
        HStack(spacing: 8) {
            ForEach(displayEmojis.prefix(10), id: \.self) { emoji in
                Button {
                    recordRecent(emoji)
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

    private func recordRecent(_ emoji: String) {
        var recent = UserDefaults.standard.stringArray(forKey: recentKey) ?? []
        recent.removeAll { $0 == emoji }
        recent.insert(emoji, at: 0)
        UserDefaults.standard.set(Array(recent.prefix(8)), forKey: recentKey)
    }
}

#Preview {
    ReactionPickerView(onSelect: { _ in }, onDismiss: {})
        .padding()
}
#endif
