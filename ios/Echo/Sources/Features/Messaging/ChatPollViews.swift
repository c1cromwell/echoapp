#if os(iOS)
import SwiftUI

struct CreatePollSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var question = ""
    @State private var options: [String] = ["", ""]

    let onCreate: (String, [String]) -> Void

    private var canCreate: Bool {
        let trimmedQ = question.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedOpts = options.map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }.filter { !$0.isEmpty }
        return trimmedQ.count >= 2 && trimmedOpts.count >= 2
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Question") {
                    TextField("Ask something…", text: $question, axis: .vertical)
                        .lineLimit(2...4)
                }
                Section("Options") {
                    ForEach(options.indices, id: \.self) { idx in
                        TextField("Option \(idx + 1)", text: $options[idx])
                    }
                    if options.count < 6 {
                        Button("Add option") {
                            options.append("")
                        }
                    }
                }
            }
            .navigationTitle("New poll")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        let trimmedOpts = options
                            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
                            .filter { !$0.isEmpty }
                        onCreate(question.trimmingCharacters(in: .whitespacesAndNewlines), trimmedOpts)
                        dismiss()
                    }
                    .disabled(!canCreate)
                }
            }
        }
    }
}

struct PollBubbleView: View {
    let poll: ChatPoll
    let currentUserDID: String
    let isSent: Bool
    let onVote: (String) -> Void
    let onClose: () -> Void

    private var totalVotes: Int {
        poll.options.reduce(0) { $0 + $1.voteCount }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 6) {
                Image(systemName: "chart.bar.doc.horizontal")
                    .font(.system(size: 13, weight: .semibold))
                Text("Poll")
                    .font(.system(size: 12, weight: .semibold))
                if poll.isClosed {
                    Text("Closed")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.echoInk55)
                }
            }
            .foregroundColor(isSent ? .white.opacity(0.9) : .echoInk55)

            Text(poll.question)
                .font(.system(size: 15, weight: .semibold))
                .foregroundColor(isSent ? .white : .echoInk)

            ForEach(poll.options) { option in
                pollOptionRow(option)
            }

            if totalVotes > 0 {
                Text("\(totalVotes) vote\(totalVotes == 1 ? "" : "s")")
                    .font(.system(size: 11))
                    .foregroundColor(isSent ? .white.opacity(0.75) : .echoInk40)
            }

            if poll.creatorDID == currentUserDID && !poll.isClosed {
                Button("Close poll", action: onClose)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(isSent ? .white : .echoSignal)
            }
        }
        .padding(12)
        .frame(maxWidth: 280, alignment: .leading)
        .background(isSent ? Color.echoSignal : Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    @ViewBuilder
    private func pollOptionRow(_ option: ChatPollOption) -> some View {
        let voted = option.voters.contains(currentUserDID)
        let pct = totalVotes > 0 ? Double(option.voteCount) / Double(totalVotes) : 0
        Button {
            guard !poll.isClosed, !voted else { return }
            onVote(option.id)
        } label: {
            ZStack(alignment: .leading) {
                GeometryReader { geo in
                    RoundedRectangle(cornerRadius: 8)
                        .fill((isSent ? Color.white : Color.echoSignal).opacity(0.18))
                        .frame(width: geo.size.width * pct)
                }
                HStack {
                    Text(option.text)
                        .font(.system(size: 14))
                        .foregroundColor(isSent ? .white : .echoInk)
                    Spacer()
                    if option.voteCount > 0 {
                        Text("\(option.voteCount)")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(isSent ? .white.opacity(0.85) : .echoInk55)
                    }
                    if voted {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 12))
                            .foregroundColor(isSent ? .white : .echoSignal)
                    }
                }
                .padding(.horizontal, 10)
                .padding(.vertical, 8)
            }
            .frame(height: 36)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .stroke(isSent ? Color.white.opacity(0.35) : Color.echoHair, lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
        .disabled(poll.isClosed || voted)
    }
}
#endif
