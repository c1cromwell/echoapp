#if os(iOS)
import SwiftUI

/// Compose sheet for scheduling a message (WO-65).
struct ScheduleMessageSheet: View {
    let initialText: String
    let onSchedule: (Date) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var fireAt = Date().addingTimeInterval(3600)

    private let presets: [(String, TimeInterval)] = [
        ("1 hour", 3600),
        ("Tonight 8pm", 0), // computed in body
        ("Tomorrow 9am", 0),
        ("1 week", 604800),
    ]

    var body: some View {
        NavigationStack {
            Form {
                Section("Send at") {
                    DatePicker("Time", selection: $fireAt, in: Date()..., displayedComponents: [.date, .hourAndMinute])
                }
                Section("Quick picks") {
                    Button("In 1 hour") { fireAt = Date().addingTimeInterval(3600) }
                    Button("Tomorrow 9:00 AM") { fireAt = nextMorning(hour: 9) }
                    Button("In 1 week") { fireAt = Date().addingTimeInterval(604800) }
                }
                if !initialText.isEmpty {
                    Section("Preview") {
                        Text(initialText).font(.footnote).foregroundColor(.echoInk70)
                    }
                }
            }
            .navigationTitle("Schedule message")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Schedule") {
                        onSchedule(fireAt)
                        dismiss()
                    }
                }
            }
        }
    }

    private func nextMorning(hour: Int) -> Date {
        var cal = Calendar.current
        cal.timeZone = .current
        var comps = cal.dateComponents([.year, .month, .day], from: Date().addingTimeInterval(86400))
        comps.hour = hour
        comps.minute = 0
        return cal.date(from: comps) ?? Date().addingTimeInterval(86400)
    }
}
#endif
