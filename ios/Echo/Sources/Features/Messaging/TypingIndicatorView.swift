#if os(iOS)
import SwiftUI

struct TypingIndicatorView: View {
    @State private var phase: Int = 0

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3) { index in
                Circle()
                    .fill(Color.Echo.onSurfaceVariant.opacity(0.5))
                    .frame(width: 7, height: 7)
                    .offset(y: phase == index ? -4 : 0)
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
        .background(Color.Echo.surfaceContainerLow)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .ghostBorder(opacity: 0.10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, Spacing.md.rawValue)
        .onAppear { startAnimation() }
    }

    private func startAnimation() {
        Timer.scheduledTimer(withTimeInterval: 0.3, repeats: true) { _ in
            withAnimation(.spring(response: 0.3, dampingFraction: 0.5)) {
                phase = (phase + 1) % 3
            }
        }
    }
}

#Preview {
    VStack {
        Spacer()
        TypingIndicatorView()
    }
    .background(Color.Echo.surface)
}
#endif
