// Core/DesignSystem/Modifiers/IcyBackground.swift
// Blurred gradient orbs creating the frozen lake depth effect

import SwiftUI

struct IcyBackground: ViewModifier {
    func body(content: Content) -> some View {
        content
            .background {
                ZStack {
                    Color.Echo.surface.ignoresSafeArea()

                    Circle()
                        .fill(Color.Echo.primaryContainer.opacity(0.10))
                        .frame(width: 300, height: 300)
                        .blur(radius: 120)
                        .offset(x: -80, y: -200)

                    Circle()
                        .fill(Color.Echo.secondaryContainer.opacity(0.10))
                        .frame(width: 250, height: 250)
                        .blur(radius: 100)
                        .offset(x: 100, y: 300)
                }
            }
    }
}

extension View {
    /// Applies the Glacial Interface icy background atmosphere.
    func icyBackground() -> some View {
        modifier(IcyBackground())
    }
}

#Preview {
    Text("Icy")
        .font(Font.Echo.displayLarge)
        .foregroundStyle(Color.Echo.onSurface)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .icyBackground()
}
