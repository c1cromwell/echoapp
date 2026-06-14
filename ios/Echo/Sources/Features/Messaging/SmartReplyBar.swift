#if os(iOS)
import SwiftUI

struct SmartReplyBar: View {
    let suggestions: [SmartReplySuggestion]
    let onSelect: (String) -> Void

    var body: some View {
        if !suggestions.isEmpty {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(suggestions) { suggestion in
                        Button {
                            onSelect(suggestion.text)
                        } label: {
                            Text(suggestion.text)
                                .font(.system(size: 14, weight: .medium))
                                .foregroundColor(.echoInk)
                                .padding(.horizontal, 14)
                                .padding(.vertical, 8)
                                .background(Color.echoPaper)
                                .clipShape(Capsule())
                                .overlay(Capsule().stroke(Color.echoHair, lineWidth: 1))
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, Spacing.md.rawValue)
                .padding(.vertical, 6)
            }
            .background(Color.echoPaperDim.opacity(0.6))
        }
    }
}
#endif
