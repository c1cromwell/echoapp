#if os(iOS)
import SwiftUI

// MARK: - Quest & Gamification Section (WO-271)

/// Shows active quests and earned badges on the Profile screen.
struct QuestGamificationSection: View {
    let api: GamificationAPI
    @State private var quests: [GamificationAPI.Quest] = []
    @State private var badges: [GamificationAPI.Badge] = []

    var body: some View {
        if !quests.isEmpty || !badges.isEmpty {
            GhostBorderSection(title: "QUESTS & BADGES") {
                ForEach(quests) { quest in
                    HStack(spacing: 12) {
                        Image(systemName: quest.isCompleted ? "checkmark.seal.fill" : "target")
                            .foregroundStyle(quest.isCompleted ? Color.echoTrustGreen : Color.echoInk55)
                        VStack(alignment: .leading, spacing: 2) {
                            Text(quest.title)
                                .font(Font.Echo.bodyMedium)
                                .foregroundStyle(Color.Echo.onSurface)
                            Text(quest.description)
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.outline)
                        }
                        Spacer()
                        if !quest.isCompleted {
                            ProgressView(value: quest.progress)
                                .frame(width: 40)
                                .tint(Color.echoTrustGreen)
                        }
                    }
                }

                if !badges.isEmpty {
                    Divider()
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 12) {
                            ForEach(badges) { badge in
                                VStack(spacing: 4) {
                                    Image(systemName: badge.icon)
                                        .font(.system(size: 24))
                                        .foregroundStyle(badge.isEarned ? Color.echoSignal : Color.echoInk40)
                                    Text(badge.name)
                                        .font(.system(size: 10, weight: .medium))
                                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                                }
                                .frame(width: 64)
                                .opacity(badge.isEarned ? 1.0 : 0.4)
                            }
                        }
                    }
                }
            }
        }
    }
}

#endif
