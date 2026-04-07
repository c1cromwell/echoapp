// Features/Auth/Views/GlacialLoginScreen.swift
// Glacial Interface login screen with passkey + SMS alternative

import SwiftUI

struct GlacialLoginScreen: View {
    @State private var showSMSSection = false
    @State private var phoneNumber = ""
    let onPasskeyLogin: () -> Void
    let onSMSLogin: (String) -> Void
    let onGetStarted: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()

            ScrollView {
                VStack(spacing: 32) {
                    Spacer(minLength: 60)

                    // Frosted Glass Header
                    VStack(spacing: 8) {
                        EchoLogo(size: 64)

                        Text("ECHO")
                            .font(Font.Echo.displayMedium)
                            .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .padding(.vertical, 24)
                    .frame(maxWidth: .infinity)
                    .background(.ultraThinMaterial.opacity(0.6))
                    .clipShape(RoundedRectangle(cornerRadius: 32))
                    .ghostBorder(opacity: 0.15)
                    .glacialShadow(radius: 32, opacity: 0.04)
                    .padding(.horizontal)

                    // Passkey Button
                    SignatureGradientButton(
                        title: "Login with Passkey",
                        subtitle: "FaceID, TouchID, or PIN",
                        icon: "faceid"
                    ) {
                        onPasskeyLogin()
                    }
                    .padding(.horizontal)

                    // Secure Alternative Divider
                    Button {
                        withAnimation(.glacial) {
                            showSMSSection.toggle()
                        }
                    } label: {
                        HStack(spacing: 12) {
                            Rectangle()
                                .fill(Color.Echo.outlineVariant.opacity(0.2))
                                .frame(height: 1)

                            HStack(spacing: 4) {
                                Text("SECURE ALTERNATIVE")
                                    .font(Font.Echo.labelSm)
                                    .tracking(1.5)
                                    .foregroundStyle(Color.Echo.outline.opacity(0.6))
                                Image(systemName: showSMSSection ? "chevron.up" : "chevron.down")
                                    .font(.system(size: 8, weight: .bold))
                                    .foregroundStyle(Color.Echo.outline.opacity(0.6))
                            }

                            Rectangle()
                                .fill(Color.Echo.outlineVariant.opacity(0.2))
                                .frame(height: 1)
                        }
                        .padding(.horizontal)
                    }

                    // SMS Section (Expandable)
                    if showSMSSection {
                        VStack(spacing: 16) {
                            TextField("Phone number", text: $phoneNumber)
                                .font(Font.Echo.bodyLarge)
                                .foregroundStyle(Color.Echo.onSurface)
                                .padding(16)
                                .background(Color.Echo.surfaceContainerLowest)
                                .clipShape(RoundedRectangle(cornerRadius: 32))
                                .ghostBorder(opacity: 0.10)
                                #if os(iOS)
                                .keyboardType(.phonePad)
                                #endif

                            Button {
                                onSMSLogin(phoneNumber)
                            } label: {
                                Text("Send Code")
                                    .font(Font.Echo.bodyLarge)
                                    .foregroundStyle(Color.Echo.onSurface)
                                    .frame(maxWidth: .infinity)
                                    .padding(.vertical, 14)
                                    .background(Color.Echo.surfaceContainerHighest)
                                    .clipShape(RoundedRectangle(cornerRadius: 32))
                                    .ghostBorder(opacity: 0.20)
                            }
                            .disabled(phoneNumber.count < 10)
                            .opacity(phoneNumber.count < 10 ? 0.5 : 1.0)
                        }
                        .padding(20)
                        .background(Color.Echo.surfaceContainerLow)
                        .clipShape(RoundedRectangle(cornerRadius: 32))
                        .ghostBorder(opacity: 0.15)
                        .padding(.horizontal)
                        .transition(.opacity.combined(with: .move(edge: .top)))
                    }

                    Spacer(minLength: 40)

                    // Footer
                    HStack(spacing: 4) {
                        Text("New to ECHO?")
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                        Button("Get Started") {
                            onGetStarted()
                        }
                        .font(Font.Echo.bodyMedium.bold())
                        .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .padding(.bottom, 32)
                }
            }
        }
        .icyBackground()
    }
}

#Preview {
    GlacialLoginScreen(
        onPasskeyLogin: {},
        onSMSLogin: { _ in },
        onGetStarted: {}
    )
}
