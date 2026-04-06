// Presentation/Screens/Notifications/NotificationCenterView.swift
// Notification center with grouped notifications, swipe actions, and badge management

import SwiftUI

// MARK: - Notification Center View

struct NotificationCenterView: View {
    @StateObject private var viewModel = NotificationViewModel()

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 0) {
                ForEach(viewModel.groupedNotifications.keys.sorted().reversed(), id: \.self) { date in
                    // Section header
                    Text(viewModel.sectionTitle(for: date))
                        .font(.custom("Inter", size: 10))
                        .fontWeight(.bold)
                        .tracking(2)
                        .textCase(.uppercase)
                        .foregroundStyle(Color.Echo.outline)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, 20)
                        .padding(.top, 24)
                        .padding(.bottom, 8)

                    ForEach(viewModel.groupedNotifications[date] ?? []) { notification in
                        NotificationRow(notification: notification)
                            .swipeActions(edge: .trailing) {
                                Button("Delete", role: .destructive) {
                                    viewModel.delete(notification)
                                }
                                Button("Mute") {
                                    viewModel.mute(notification)
                                }
                                .tint(Color.Echo.outline)
                            }
                    }
                }
            }
            .padding(.bottom, 100)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Notifications")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button("Mark All") {
                    viewModel.markAllRead()
                }
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.primaryContainer)
            }
        }
        .task { await viewModel.loadNotifications() }
    }
}

// MARK: - Notification Row

struct NotificationRow: View {
    let notification: AppNotification

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            // Category icon
            Circle()
                .fill(notification.category.color.opacity(0.15))
                .frame(width: 40, height: 40)
                .overlay(
                    Image(systemName: notification.category.icon)
                        .font(.system(size: 16))
                        .foregroundStyle(notification.category.color)
                )

            VStack(alignment: .leading, spacing: 4) {
                Text(notification.title)
                    .font(.custom("Inter", size: 14))
                    .fontWeight(notification.isRead ? .regular : .bold)
                    .foregroundStyle(Color.Echo.onSurface)

                if let subtitle = notification.subtitle {
                    Text(subtitle)
                        .font(Font.Echo.bodyMedium)
                        .foregroundStyle(Color.Echo.outline)
                        .lineLimit(1)
                }
            }

            Spacer()

            Text(notification.timeAgo)
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.outline)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
        .background(
            notification.isRead
                ? Color.Echo.surface
                : Color.Echo.surfaceContainerLowest
        )
        .overlay(alignment: .leading) {
            if !notification.isRead {
                Rectangle()
                    .fill(Color.Echo.primaryContainer)
                    .frame(width: 3)
            }
        }
    }
}

// MARK: - App Notification Model

struct AppNotification: Identifiable {
    let id: String
    let title: String
    let subtitle: String?
    let category: NotificationCategory
    let timestamp: Date
    var isRead: Bool
    let deepLink: String?

    var timeAgo: String {
        RelativeDateTimeFormatter().localizedString(for: timestamp, relativeTo: .now)
    }
}

enum NotificationCategory: String, CaseIterable {
    case message, call, group, channel, trust, wallet, system

    var icon: String {
        switch self {
        case .message: return "message.fill"
        case .call: return "phone.fill"
        case .group: return "person.3.fill"
        case .channel: return "megaphone.fill"
        case .trust: return "shield.checkered"
        case .wallet: return "trophy.fill"
        case .system: return "gear"
        }
    }

    var color: Color {
        switch self {
        case .message: return Color.Echo.primaryContainer
        case .call: return Color.Echo.success
        case .group: return Color(hex: "#8B5CF6")
        case .channel: return Color.Echo.warning
        case .trust: return Color.Echo.primaryContainer
        case .wallet: return Color.Echo.success
        case .system: return Color.Echo.outline
        }
    }
}

// MARK: - Notification ViewModel

@MainActor
class NotificationViewModel: ObservableObject {
    @Published var notifications: [AppNotification] = []
    @Published var groupedNotifications: [Date: [AppNotification]] = [:]

    func loadNotifications() async {
        // TODO: Load from notification service
        groupNotifications()
    }

    func markAllRead() {
        for i in notifications.indices {
            notifications[i].isRead = true
        }
        groupNotifications()
    }

    func delete(_ notification: AppNotification) {
        notifications.removeAll { $0.id == notification.id }
        groupNotifications()
    }

    func mute(_ notification: AppNotification) {
        // TODO: Mute notification source
    }

    func sectionTitle(for date: Date) -> String {
        let calendar = Calendar.current
        if calendar.isDateInToday(date) { return "Today" }
        if calendar.isDateInYesterday(date) { return "Yesterday" }
        return date.formatted(date: .abbreviated, time: .omitted)
    }

    private func groupNotifications() {
        let calendar = Calendar.current
        groupedNotifications = Dictionary(grouping: notifications) { notification in
            calendar.startOfDay(for: notification.timestamp)
        }
    }
}
