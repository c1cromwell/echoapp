// Core/DesignSystem/Components/GlacialNavigationBar.swift
// Frosted glass navigation bar with ghost border

import SwiftUI

struct GlacialNavigationBar<Leading: View, Trailing: View>: View {
    let title: String
    let leading: Leading
    let trailing: Trailing

    init(
        title: String,
        @ViewBuilder leading: () -> Leading = { EmptyView() },
        @ViewBuilder trailing: () -> Trailing = { EmptyView() }
    ) {
        self.title = title
        self.leading = leading()
        self.trailing = trailing()
    }

    var body: some View {
        HStack {
            leading
            Spacer()
            Text(title)
                .font(Font.Echo.headlineSm)
                .foregroundStyle(Color.Echo.onSurface)
            Spacer()
            trailing
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(.ultraThinMaterial.opacity(0.6))
        .ghostBorder(opacity: 0.15)
        .glacialShadow(radius: 32, opacity: 0.04)
    }
}

#Preview {
    VStack {
        GlacialNavigationBar(title: "Messages") {
            Image(systemName: "chevron.left")
                .foregroundStyle(Color.Echo.primaryContainer)
        } trailing: {
            Image(systemName: "gear")
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        Spacer()
    }
}
