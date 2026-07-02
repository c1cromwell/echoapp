#if os(iOS)
import SwiftUI

// MARK: - Shimmer Effect

struct ShimmerModifier: ViewModifier {
    @State private var phase: CGFloat = 0

    func body(content: Content) -> some View {
        content
            .overlay(
                LinearGradient(
                    colors: [
                        .clear,
                        Color.white.opacity(0.3),
                        .clear
                    ],
                    startPoint: .leading,
                    endPoint: .trailing
                )
                .offset(x: phase)
                .mask(content)
            )
            .onAppear {
                withAnimation(
                    .linear(duration: 1.5)
                    .repeatForever(autoreverses: false)
                ) {
                    phase = UIScreen.main.bounds.width
                }
            }
    }
}

extension View {
    func shimmer() -> some View {
        modifier(ShimmerModifier())
    }
}

// MARK: - Skeleton Pill

struct SkeletonPill: View {
    let width: CGFloat
    let height: CGFloat

    init(width: CGFloat = 120, height: CGFloat = 14) {
        self.width = width
        self.height = height
    }

    var body: some View {
        RoundedRectangle(cornerRadius: height / 2)
            .fill(Color.echoPaperEdge)
            .frame(width: width, height: height)
    }
}

// MARK: - Conversation Row Skeleton

struct ConversationRowSkeleton: View {
    var body: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.echoPaperEdge)
                .frame(width: 48, height: 48)

            VStack(alignment: .leading, spacing: 6) {
                SkeletonPill(width: 120, height: 14)
                SkeletonPill(width: 180, height: 12)
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 6) {
                SkeletonPill(width: 40, height: 10)
            }
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 12)
        .shimmer()
    }
}

// MARK: - Contact Row Skeleton

struct ContactRowSkeleton: View {
    var body: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.echoPaperEdge)
                .frame(width: 44, height: 44)

            VStack(alignment: .leading, spacing: 6) {
                SkeletonPill(width: 100, height: 14)
                SkeletonPill(width: 70, height: 10)
            }

            Spacer()
        }
        .padding(.horizontal, 18)
        .padding(.vertical, 10)
        .shimmer()
    }
}

// MARK: - Balance Card Skeleton

struct BalanceCardSkeleton: View {
    var body: some View {
        GhostBorderCard {
            VStack(spacing: 10) {
                SkeletonPill(width: 90, height: 12)
                SkeletonPill(width: 140, height: 28)
                SkeletonPill(width: 50, height: 12)
            }
            .frame(maxWidth: .infinity)
        }
        .shimmer()
    }
}

// MARK: - Skeleton List

struct SkeletonList<Skeleton: View>: View {
    let count: Int
    let skeleton: () -> Skeleton

    init(count: Int = 6, @ViewBuilder skeleton: @escaping () -> Skeleton) {
        self.count = count
        self.skeleton = skeleton
    }

    var body: some View {
        ForEach(0..<count, id: \.self) { _ in
            skeleton()
        }
    }
}

#endif
