#if os(iOS)
// Core/DesignSystem/Components/SecureThreadIndicator.swift
// 2px pulsating sky blue line indicating active encrypted connection

import SwiftUI

public struct SecureThreadIndicator: View {
    public init() {}
    @State private var opacity: Double = 0.6

    public var body: some View {
        Group {
            if PrivacySettingsStore.showsEncryptionIndicator {
                Rectangle()
                    .fill(Color.echoSignal)
                    .frame(height: 2)
                    .opacity(opacity)
                    .onAppear {
                        withAnimation(.easeInOut(duration: 2).repeatForever(autoreverses: true)) {
                            opacity = 1.0
                        }
                    }
            }
        }
    }
}

#Preview {
    VStack {
        SecureThreadIndicator()
        Spacer()
    }
}
#endif
