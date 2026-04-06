import SwiftUI

/// ECHO Trust Score Circular View
/// Animated circular progress (140×140pt, 10pt stroke)
public struct TrustScoreView: View {
    let score: Int
    let level: String
    let showAnimation: Bool
    
    @State private var animationProgress: CGFloat = 0
    
    private let size: CGFloat = 140
    private let strokeWidth: CGFloat = 10
    
    public init(
        score: Int,
        level: String = "Newcomer",
        showAnimation: Bool = true
    ) {
        self.score = min(max(score, 0), 100)
        self.level = level
        self.showAnimation = showAnimation
    }
    
    var scorePercentage: Double {
        Double(score) / 100.0
    }
    
    var scoreColor: Color {
        Color.trustColor(for: level)
    }
    
    public var body: some View {
        VStack(spacing: Spacing.md.rawValue) {
            ZStack {
                // Background circle
                Circle()
                    .stroke(Color.echoGray200, lineWidth: strokeWidth)
                
                // Progress circle
                Circle()
                    .trim(from: 0, to: animationProgress)
                    .stroke(scoreColor, style: StrokeStyle(lineWidth: strokeWidth, lineCap: .round))
                    .rotationEffect(.degrees(-90))
                    .animation(.easeOut(duration: 1.0), value: animationProgress)
                
                // Score Text
                VStack(spacing: Spacing.xs.rawValue) {
                    Text("\(score)")
                        .typographyStyle(.display, color: scoreColor)
                        .font(.system(size: 48, weight: .bold))
                    
                    Text("out of 100")
                        .typographyStyle(.caption, color: .echoGray500)
                }
            }
            .frame(width: size, height: size)
            
            // Trust Level
            VStack(spacing: Spacing.xs.rawValue) {
                Text(level)
                    .typographyStyle(.h4, color: scoreColor)
                    .fontWeight(.semibold)
                
                // Score breakdown hint
                Text("Based on identity, behavior, network & activity")
                    .typographyStyle(.caption, color: .echoSecondaryText)
                    .multilineTextAlignment(.center)
            }
        }
        .onAppear {
            if showAnimation {
                animationProgress = scorePercentage
            } else {
                animationProgress = scorePercentage
            }
        }
        .accessibility(element: children: .ignore)
        .accessibility(label: Text("Trust Score"))
        .accessibility(value: Text("\(score) out of 100, \(level) level"))
    }
}

// MARK: - Preview

#if DEBUG
struct TrustScoreView_Previews: PreviewProvider {
    static var previews: some View {
        VStack(spacing: Spacing.xl.rawValue) {
            TrustScoreView(score: 15, level: "Newcomer")
            TrustScoreView(score: 45, level: "Basic")
            TrustScoreView(score: 65, level: "Trusted")
            TrustScoreView(score: 85, level: "Verified")
            TrustScoreView(score: 95, level: "Highly Trusted")
        }
        .frame(maxWidth: .infinity)
        .echoSpacing(.xl)
        .background(Color.echoBackground)
    }
}
#endif
