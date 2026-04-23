// Features/Onboarding/Recovery/RecoveryPromptScheduler.swift

import Foundation
import UserNotifications

@MainActor
final class RecoveryPromptScheduler {
    static let shared = RecoveryPromptScheduler()

    private let reminderIds = ["echo.recovery.reminder.0",
                               "echo.recovery.reminder.1",
                               "echo.recovery.reminder.2"]
    // Intervals: 7 days, 30 days, 90 days post first-run.
    private let reminderIntervals: [TimeInterval] = [
        7  * 24 * 3600,
        30 * 24 * 3600,
        90 * 24 * 3600,
    ]

    private init() {}

    /// Schedules local in-app-only reminders for recovery phrase export.
    /// Silent category: no sound, no badge, no banner — just an in-app check on next launch.
    func scheduleReminders() async {
        guard phraseNotYetExported else { return }
        let center = UNUserNotificationCenter.current()
        let granted = (try? await center.requestAuthorization(options: [.alert])) ?? false
        guard granted else { return }

        for (i, interval) in reminderIntervals.enumerated() {
            let content = UNMutableNotificationContent()
            content.title = "Back up your ECHO identity"
            content.body = "Write down your 24-word recovery phrase so you can restore your account if you lose this device."
            content.sound = nil

            let trigger = UNTimeIntervalNotificationTrigger(timeInterval: interval, repeats: false)
            let req = UNNotificationRequest(identifier: reminderIds[i], content: content, trigger: trigger)
            try? await center.add(req)
        }
    }

    func cancelAllPendingReminders() {
        UNUserNotificationCenter.current().removePendingNotificationRequests(withIdentifiers: reminderIds)
    }

    /// Call on every app launch from Messages tab. If an overdue reminder exists and
    /// the phrase has not been exported, fires the presenter closure (e.g., show a sheet).
    func checkAndPresentIfOverdue(presenter: @escaping () -> Void) {
        guard phraseNotYetExported else { return }
        guard let firstRunAt = UserDefaults.standard.object(forKey: "echo.firstRunCompletedAt") as? Date else { return }

        let elapsed = Date().timeIntervalSince(firstRunAt)
        guard reminderIntervals.contains(where: { $0 <= elapsed }) else { return }

        // Throttle to once per 24 hours so the prompt isn't shown on every launch.
        let lastSeen = UserDefaults.standard.object(forKey: "echo.recoveryPromptLastPresentedAt") as? Date ?? .distantPast
        guard Date().timeIntervalSince(lastSeen) > 24 * 3600 else { return }

        UserDefaults.standard.set(Date(), forKey: "echo.recoveryPromptLastPresentedAt")
        presenter()
    }

    private var phraseNotYetExported: Bool {
        UserDefaults.standard.object(forKey: "echo.recoveryPhraseExportedAt") == nil
    }
}
