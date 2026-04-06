// Core/DesignSystem/Components/GhostBorderCard.swift
// Surface-container card with 15% opacity border

import SwiftUI

struct GhostBorderCard<Content: View>: View {
    let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        content
            .padding(16)
            .background(Color.Echo.surfaceContainer)
            .clipShape(RoundedRectangle(cornerRadius: 32))
            .ghostBorder()
            .glacialShadow()
    }
}

#Preview {
    GhostBorderCard {
        VStack(alignment: .leading, spacing: 8) {
            Text("Balance")
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Text("1,250.00 ECHO")
                .font(Font.Echo.headlineSm)
                .foregroundStyle(Color.Echo.onSurface)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
    .padding()
}
