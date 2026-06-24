#if os(iOS)
import SwiftUI

/// Collapsible on-device thread summary (WO-CA1 / M7).
struct ThreadSummaryBanner: View {
    let summary: ThreadSummary
    @State private var expanded = true

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Button {
                withAnimation(.easeInOut(duration: 0.2)) { expanded.toggle() }
            } label: {
                HStack(spacing: 6) {
                    Image(systemName: "sparkles")
                        .font(.system(size: 12, weight: .semibold))
                    Text("Thread summary")
                        .font(.system(size: 12, weight: .semibold))
                    Spacer()
                    Image(systemName: expanded ? "chevron.up" : "chevron.down")
                        .font(.system(size: 11, weight: .semibold))
                }
                .foregroundColor(.echoInk70)
            }
            .buttonStyle(.plain)

            if expanded {
                Text(summary.summary)
                    .font(.system(size: 13))
                    .foregroundColor(.echoInk)
                    .fixedSize(horizontal: false, vertical: true)
                Text("On-device · \(summary.sentenceCount) messages")
                    .font(.system(size: 11))
                    .foregroundColor(.echoInk55)
            }
        }
        .padding(.horizontal, Spacing.lg.rawValue)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
        .overlay(
            Rectangle()
                .frame(height: 1)
                .foregroundColor(.echoHair),
            alignment: .bottom
        )
    }
}
#endif
