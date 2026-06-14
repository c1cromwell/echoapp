#if os(iOS)
import SwiftUI

struct WaveformView: View {
    let samples: [CGFloat]
    var progress: Double = 0
    var accentColor: Color = .echoSignal
    var inactiveColor: Color = .echoInk40

    private let barWidth: CGFloat = 3
    private let barSpacing: CGFloat = 2

    var body: some View {
        GeometryReader { geo in
            let maxBars = Int(geo.size.width / (barWidth + barSpacing))
            let displayed = resample(to: maxBars)
            let progressIndex = Int(Double(displayed.count) * min(max(progress, 0), 1))

            HStack(alignment: .center, spacing: barSpacing) {
                ForEach(Array(displayed.enumerated()), id: \.offset) { index, amplitude in
                    RoundedRectangle(cornerRadius: barWidth / 2)
                        .fill(index < progressIndex ? accentColor : inactiveColor)
                        .frame(
                            width: barWidth,
                            height: max(4, amplitude * geo.size.height)
                        )
                }
            }
            .frame(maxHeight: .infinity, alignment: .center)
        }
    }

    private func resample(to count: Int) -> [CGFloat] {
        guard count > 0 else { return [] }
        guard !samples.isEmpty else {
            return Array(repeating: CGFloat(0.15), count: count)
        }
        if samples.count == count { return samples }
        var result = [CGFloat]()
        let step = Double(samples.count) / Double(count)
        for i in 0..<count {
            let sampleIndex = min(Int(Double(i) * step), samples.count - 1)
            result.append(samples[sampleIndex])
        }
        return result
    }
}
#endif
