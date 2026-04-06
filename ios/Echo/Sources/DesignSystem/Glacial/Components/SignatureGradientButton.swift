// Core/DesignSystem/Components/SignatureGradientButton.swift
// Pill CTA with deep navy → sky blue gradient

import SwiftUI

struct SignatureGradientButton: View {
    let title: String
    let subtitle: String?
    let icon: String?
    let action: () -> Void
    @State private var isPressed = false

    init(
        title: String,
        subtitle: String? = nil,
        icon: String? = nil,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.subtitle = subtitle
        self.icon = icon
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                if let icon {
                    Image(systemName: icon)
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(width: 40, height: 40)
                        .background(.white.opacity(0.1))
                        .clipShape(Circle())
                        .overlay(Circle().strokeBorder(.white.opacity(0.2), lineWidth: 1))
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .font(Font.Echo.titleLarge)
                        .foregroundStyle(.white)

                    if let subtitle {
                        Text(subtitle)
                            .font(Font.Echo.bodySm)
                            .foregroundStyle(Color.Echo.skyLight.opacity(0.7))
                    }
                }

                Spacer()

                Image(systemName: "chevron.right")
                    .foregroundStyle(.white.opacity(0.5))
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 16)
            .background(LinearGradient.signature)
            .clipShape(RoundedRectangle(cornerRadius: 32))
            .deepGlacialShadow()
        }
        .buttonStyle(GlacialPressStyle())
    }
}

// MARK: - Press Style

struct GlacialPressStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
            .animation(.glacialPress, value: configuration.isPressed)
    }
}

#Preview {
    VStack(spacing: 16) {
        SignatureGradientButton(
            title: "Login with Passkey",
            subtitle: "FaceID, TouchID, or PIN",
            icon: "faceid"
        ) {}

        SignatureGradientButton(
            title: "Stake ECHO",
            icon: "lock.shield"
        ) {}
    }
    .padding()
}
