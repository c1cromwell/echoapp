// Core/DesignSystem/Components/SecureThreadIndicator.swift
// 2px pulsating sky blue line indicating active encrypted connection

import SwiftUI

struct SecureThreadIndicator: View {
    @State private var opacity: Double = 0.6

    var body: some View {
        Rectangle()
            .fill(Color.Echo.primaryContainer)
            .frame(height: 2)
            .opacity(opacity)
            .onAppear {
                withAnimation(.easeInOut(duration: 2).repeatForever(autoreverses: true)) {
                    opacity = 1.0
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
