#if os(iOS)
import SwiftUI

struct ReactionChipsView: View {
    let reactions: [ReactionCount]
    let currentUserDID: String
    let onTap: (String) -> Void

    var body: some View {
        if !reactions.isEmpty {
            HStack(spacing: 4) {
                ForEach(reactions) { reaction in
                    Button {
                        onTap(reaction.emoji)
                    } label: {
                        HStack(spacing: 3) {
                            Text(reaction.emoji)
                                .font(.system(size: 13))
                            if reaction.count > 1 {
                                Text("\(reaction.count)")
                                    .font(.system(size: 11, weight: .medium))
                                    .foregroundStyle(Color.Echo.onSurfaceVariant)
                            }
                        }
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(
                            reaction.reactors.contains(currentUserDID)
                                ? Color.Echo.primaryContainer.opacity(0.15)
                                : Color.Echo.surfaceContainerLow
                        )
                        .clipShape(Capsule())
                        .overlay(
                            Capsule().strokeBorder(
                                reaction.reactors.contains(currentUserDID)
                                    ? Color.Echo.primaryContainer.opacity(0.3)
                                    : Color.Echo.outlineVariant.opacity(0.15),
                                lineWidth: 1
                            )
                        )
                    }
                    .buttonStyle(.plain)
                }
            }
        }
    }
}

#Preview {
    ReactionChipsView(
        reactions: [
            ReactionCount(emoji: "👍", count: 2, reactors: ["did:key:me", "did:key:other"]),
            ReactionCount(emoji: "❤️", count: 1, reactors: ["did:key:other"])
        ],
        currentUserDID: "did:key:me",
        onTap: { _ in }
    )
    .padding()
}
#endif
