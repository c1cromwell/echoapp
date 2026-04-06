import SwiftUI

// MARK: - Group Member (SwiftUI-dependent, lives here instead of Models)

public struct GroupMember {
    public let initials: String
    public let color: Color

    public init(initials: String, color: Color) {
        self.initials = initials
        self.color = color
    }
}

// MARK: - Online Indicator

public struct OnlineIndicator: View {
    public init() {}

    public var body: some View {
        Circle()
            .fill(Color(hex: 0x10B981))
            .frame(width: 14, height: 14)
            .overlay(
                Circle()
                    .stroke(Color.white, lineWidth: 2.5)
            )
    }
}

// MARK: - Unread Badge

public struct UnreadBadge: View {
    public let count: Int

    public init(count: Int) {
        self.count = count
    }

    private var displayText: String {
        count > 99 ? "99+" : "\(count)"
    }

    public var body: some View {
        Text(displayText)
            .font(.system(size: 10, weight: .bold))
            .foregroundColor(.white)
            .padding(.horizontal, 6)
            .frame(minWidth: 20, minHeight: 20)
            .background(Color(hex: 0xF43F5E))
            .clipShape(Capsule())
            .overlay(
                Capsule()
                    .stroke(Color.white, lineWidth: 2)
            )
    }
}

// MARK: - Contact Avatar View

public struct ContactAvatarView: View {
    public let initials: String
    public let gradient: [Color]
    public let size: CGFloat

    public init(initials: String, gradient: [Color], size: CGFloat) {
        self.initials = initials
        self.gradient = gradient
        self.size = size
    }

    public var body: some View {
        ZStack {
            Circle()
                .fill(
                    LinearGradient(
                        colors: gradient,
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .frame(width: size, height: size)
                .shadow(color: Color.Echo.onSurface.opacity(0.15), radius: 8, x: 0, y: 4)

            Text(initials)
                .font(.system(size: size * 0.4, weight: .semibold))
                .foregroundColor(.white)
        }
    }
}

// MARK: - Group Avatar View (2x2 Grid)

public struct GroupAvatarView: View {
    @Environment(\.theme) var theme
    public let members: [GroupMember]
    public let size: CGFloat

    private let spacing: CGFloat = 2
    private let padding: CGFloat = 4

    private var faceSize: CGFloat {
        (size - padding * 2 - spacing) / 2
    }

    public init(members: [GroupMember], size: CGFloat) {
        self.members = members
        self.size = size
    }

    public var body: some View {
        ZStack {
            Circle()
                .fill(theme.colors.surface)
                .frame(width: size, height: size)
                .overlay(
                    Circle()
                        .stroke(theme.colors.border, lineWidth: 2)
                )

            LazyVGrid(
                columns: [
                    GridItem(.fixed(faceSize), spacing: spacing),
                    GridItem(.fixed(faceSize), spacing: spacing)
                ],
                spacing: spacing
            ) {
                ForEach(0..<4, id: \.self) { index in
                    if index < members.count {
                        let member = members[index]
                        let displayText = index == 3 && members.count > 4
                            ? "+\(members.count - 3)"
                            : member.initials

                        RoundedRectangle(cornerRadius: faceSize * 0.5)
                            .fill(member.color)
                            .frame(width: faceSize, height: faceSize)
                            .overlay(
                                Text(displayText)
                                    .font(.system(size: 10, weight: .semibold))
                                    .foregroundColor(.white)
                            )
                    }
                }
            }
            .padding(padding)
        }
    }
}

// MARK: - Pinned Avatar View

public struct PinnedAvatarView: View {
    @Environment(\.theme) var theme
    public let item: PinnedItem
    public let size: CGFloat
    public var members: [GroupMember]?

    private var gradients: [[Color]] {
        [
            [Color(hex: 0x7C3AED), Color(hex: 0x06B6D4)],
            [Color(hex: 0x06B6D4), Color(hex: 0x10B981)],
            [Color(hex: 0xF43F5E), Color(hex: 0xF59E0B)]
        ]
    }

    public init(item: PinnedItem, size: CGFloat = 56, members: [GroupMember]? = nil) {
        self.item = item
        self.size = size
        self.members = members
    }

    public var body: some View {
        ZStack {
            if item.type == .group, let members = members {
                GroupAvatarView(members: members, size: size)
            } else {
                ContactAvatarView(
                    initials: item.initials,
                    gradient: gradients[item.gradientIndex % gradients.count],
                    size: size
                )
            }

            // Online indicator
            if item.isOnline && item.type == .contact {
                OnlineIndicator()
                    .position(x: size - 7, y: size - 7)
            }

            // Unread badge
            if item.unreadCount > 0 {
                UnreadBadge(count: item.unreadCount)
                    .position(x: size - 4, y: 4)
            }
        }
        .frame(width: size, height: size)
    }
}

// MARK: - Pinned Item Card

public struct PinnedItemCard: View {
    @Environment(\.theme) var theme
    public let item: PinnedItem
    public var members: [GroupMember]?
    public let onTap: () -> Void

    public init(item: PinnedItem, members: [GroupMember]? = nil, onTap: @escaping () -> Void) {
        self.item = item
        self.members = members
        self.onTap = onTap
    }

    public var body: some View {
        Button(action: onTap) {
            VStack(spacing: 8) {
                PinnedAvatarView(item: item, size: 56, members: members)

                Text(item.name)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(theme.colors.textSecondary)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .frame(width: 72)
            }
        }
        .buttonStyle(ScaleButtonStyle())
    }
}

// MARK: - Pinned Section View

public struct PinnedSectionView: View {
    @Environment(\.theme) var theme
    public let items: [PinnedItem]
    public let onItemTap: (PinnedItem) -> Void
    public let onEditTap: () -> Void
    public let maxItems: Int

    public init(
        items: [PinnedItem],
        maxItems: Int = 9,
        onItemTap: @escaping (PinnedItem) -> Void,
        onEditTap: @escaping () -> Void
    ) {
        self.items = items
        self.maxItems = maxItems
        self.onItemTap = onItemTap
        self.onEditTap = onEditTap
    }

    public var body: some View {
        if items.isEmpty {
            EmptyView()
        } else {
            VStack(spacing: 12) {
                // Header
                HStack {
                    HStack(spacing: 6) {
                        Image(systemName: "pin.fill")
                            .font(.system(size: 12))
                            .foregroundColor(theme.colors.textTertiary)

                        Text("PINNED")
                            .font(.system(size: 13, weight: .bold))
                            .foregroundColor(theme.colors.textTertiary)
                            .tracking(1)
                    }

                    Spacer()

                    Button(action: onEditTap) {
                        Text("Edit")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundColor(theme.colors.primary)
                    }
                }
                .padding(.horizontal, ScreenPadding.horizontal)

                // Horizontal scroll of pinned items
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(items.prefix(maxItems)) { item in
                            PinnedItemCard(item: item) {
                                onItemTap(item)
                            }
                        }
                    }
                    .padding(.horizontal, ScreenPadding.horizontal)
                }
            }
            .padding(.vertical, 12)
        }
    }
}

// MARK: - Preview

#if DEBUG
struct PinnedSection_Previews: PreviewProvider {
    static var previews: some View {
        let items = [
            PinnedItem(id: "1", type: .contact, name: "Alice", avatar: nil, initials: "AL", gradientIndex: 0, isOnline: true, unreadCount: 3),
            PinnedItem(id: "2", type: .contact, name: "Bob", avatar: nil, initials: "BW", gradientIndex: 1, isOnline: false),
            PinnedItem(id: "3", type: .group, name: "Design Team", avatar: nil, initials: "DT", gradientIndex: 2, unreadCount: 1)
        ]

        VStack {
            PinnedSectionView(
                items: items,
                onItemTap: { _ in },
                onEditTap: {}
            )
            Spacer()
        }
        .background(Color.echoBackground)
    }
}
#endif
