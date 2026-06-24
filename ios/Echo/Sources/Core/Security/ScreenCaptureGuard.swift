#if os(iOS)
import SwiftUI
import UIKit

/// Blurs sensitive UI while screen recording or mirroring (WO-125 / hidden-folder spec §7.3).
struct ScreenCaptureGuard: ViewModifier {
    let enabled: Bool
    @State private var isCaptured = false

    func body(content: Content) -> some View {
        ZStack {
            content
                .blur(radius: enabled && isCaptured ? 18 : 0)
            if enabled && isCaptured {
                VStack(spacing: 10) {
                    Image(systemName: "eye.slash.fill")
                        .font(.system(size: 32, weight: .light))
                    Text("Content protected")
                        .font(.system(size: 15, weight: .semibold))
                }
                .foregroundStyle(Color.echoInk70)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(Color.echoPaper.opacity(0.92))
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: UIScreen.capturedDidChangeNotification)) { _ in
            isCaptured = UIScreen.main.isCaptured
        }
        .onAppear {
            isCaptured = UIScreen.main.isCaptured
        }
    }
}

extension View {
    func screenCaptureGuard(enabled: Bool) -> some View {
        modifier(ScreenCaptureGuard(enabled: enabled))
    }
}
#endif
