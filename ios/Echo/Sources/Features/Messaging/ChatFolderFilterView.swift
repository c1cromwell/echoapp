import SwiftUI

/// Trust-based conversation folders (spec §5.2). Telegram folder-tab pattern adapted
/// to Echo trust tiers — one green hue at five opacities, not rainbow categories.
public enum ChatFolder: String, CaseIterable, Identifiable, Sendable {
    case all, verified, trusted, unverified

    public var id: String { rawValue }

    public var title: String {
        switch self {
        case .all:         return "All"
        case .verified:    return "Verified"
        case .trusted:     return "Trusted"
        case .unverified:  return "Unverified"
        }
    }

    /// Inclusive predicate over a numeric trust tier (T0–T4).
    public func includes(tier: Int) -> Bool {
        switch self {
        case .all:        return true
        case .verified:   return tier >= 2   // T2+
        case .trusted:    return tier >= 3   // T3+
        case .unverified: return tier <= 1   // T0–T1
        }
    }
}

/// Horizontal trust-folder chips. Active chip = signal fill / white; inactive =
/// paper-dim / ink55.
public struct ChatFolderFilterView: View {
    @Binding var selection: ChatFolder
    var unreadCounts: [ChatFolder: Int]

    public init(selection: Binding<ChatFolder>, unreadCounts: [ChatFolder: Int] = [:]) {
        self._selection = selection
        self.unreadCounts = unreadCounts
    }

    public var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: Spacing.sm.rawValue) {
                ForEach(ChatFolder.allCases) { folder in
                    chip(folder)
                }
            }
            .padding(.horizontal, Spacing.lg.rawValue)
            .padding(.vertical, Spacing.sm.rawValue)
        }
    }

    @ViewBuilder
    private func chip(_ folder: ChatFolder) -> some View {
        let isActive = folder == selection
        Button {
            withAnimation(.spring(response: 0.38, dampingFraction: 0.82)) {
                selection = folder
            }
        } label: {
            HStack(spacing: 6) {
                Text(folder.title)
                    .font(.system(size: 13, weight: .semibold))
                if let count = unreadCounts[folder], count > 0 {
                    Text(count > 99 ? "99+" : "\(count)")
                        .font(.system(size: 11, weight: .bold))
                        .padding(.horizontal, 6)
                        .padding(.vertical, 1)
                        .background(isActive ? Color.white.opacity(0.25) : Color.echoSignal)
                        .foregroundColor(.white)
                        .clipShape(Capsule())
                }
            }
            .foregroundColor(isActive ? .white : .echoInk55)
            .padding(.horizontal, 14)
            .padding(.vertical, 7)
            .background(isActive ? Color.echoSignal : Color.echoPaperDim)
            .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }
}

#if DEBUG
struct ChatFolderFilterView_Previews: PreviewProvider {
    struct Wrap: View {
        @State var sel: ChatFolder = .all
        var body: some View {
            ChatFolderFilterView(selection: $sel, unreadCounts: [.all: 5, .verified: 2])
                .background(Color.echoPaper)
        }
    }
    static var previews: some View { Wrap() }
}
#endif
