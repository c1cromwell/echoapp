#if os(iOS)
import SwiftUI

/// Sheet for scheduling a message to send at a future time (WO-65).
struct ScheduleMessageSheet: View {
    let initialText: String
    let onSchedule: (Date) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var fireAt = Date().addingTimeInterval(3600)

    var body: some View {
        NavigationStack {
            Form {
                Section("Message") {
                    Text(initialText)
                        .foregroundStyle(.secondary)
                        .lineLimit(3)
                }
                Section("Send at") {
                    DatePicker("Date & Time", selection: $fireAt, in: Date()..., displayedComponents: [.date, .hourAndMinute])
                }
            }
            .navigationTitle("Schedule Message")
            .navigationBarTitleDisplayMode(.inline)
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
        .presentationDetents([.medium])
    }
}
#endif
