import SwiftUI

import SwiftUI

/// ECHO Custom Tab Bar
/// 5 tabs with unread badges and custom styling
public struct EchoTabBar: View {
    @Binding var selectedTab: Int
    let tabs: [TabBarItem]
    
    public init(selectedTab: Binding<Int>, tabs: [TabBarItem]) {
        self._selectedTab = selectedTab
        self.tabs = tabs
    }
    
    public var body: some View {
        VStack(spacing: 0) {
            // Tonal shift separator instead of hard Divider (Glacial Interface rule)
            Rectangle()
                .fill(Color.Echo.surfaceContainerHigh)
                .frame(height: 1)
                .opacity(0.5)
            
            HStack(spacing: 0) {
                ForEach(0..<tabs.count, id: \.self) { index in
                    TabBarItemView(
                        item: tabs[index],
                        isSelected: selectedTab == index,
                        action: {
                            withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) {
                                selectedTab = index
                            }
                        }
                    )
                }
            }
            .frame(height: 56)
            .background(Color.Echo.surfaceContainerLowest)
        }
    }
}

/// Tab Bar Item
public struct TabBarItem: Identifiable {
    public let id = UUID()
    public let title: String
    public let icon: Image
    public let badge: Int?
    
    public init(title: String, icon: Image, badge: Int? = nil) {
        self.title = title
        self.icon = icon
        self.badge = badge
    }
}

/// Individual Tab Bar Item View
struct TabBarItemView: View {
    let item: TabBarItem
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 4) {
                ZStack(alignment: .topTrailing) {
                    item.icon
                        .font(.system(size: 18, weight: .semibold))
                    
                    if let badge = item.badge, badge > 0 {
                        Circle()
                            .fill(Color.echoError)
                            .overlay(
                                Text("\(min(badge, 9))")
                                    .font(.system(size: 9, weight: .bold))
                                    .foregroundColor(.white)
                            )
                            .frame(width: 16, height: 16)
                            .offset(x: 4, y: -4)
                    }
                }
                
                Text(item.title)
                    .font(.system(size: 10, weight: .semibold))
            }
            .foregroundColor(isSelected ? .echoPrimary : .echoGray500)
            .frame(maxWidth: .infinity)
        }
        .accessibilityLabel(item.title)
        .accessibilityValue(item.badge.map { $0 > 0 ? "\($0) unread" : "" } ?? "")
    }
}

// MARK: - Preview

#if DEBUG
struct EchoTabBar_Previews: PreviewProvider {
    @State static var selectedTab = 0
    
    static var previews: some View {
        VStack {
            Spacer()
            
            EchoTabBar(
                selectedTab: $selectedTab,
                tabs: [
                    TabBarItem(title: "Messages", icon: Image(systemName: "message.fill"), badge: 3),
                    TabBarItem(title: "Contacts", icon: Image(systemName: "person.2.fill")),
                    TabBarItem(title: "Trust", icon: Image(systemName: "shield.checkmark.fill")),
                    TabBarItem(title: "Rewards", icon: Image(systemName: "star.fill")),
                    TabBarItem(title: "Profile", icon: Image(systemName: "person.circle.fill"))
                ]
            )
        }
        .background(Color.echoBackground)
    }
}
#endif
