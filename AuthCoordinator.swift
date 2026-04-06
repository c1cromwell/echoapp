// Features/Auth/AuthCoordinator.swift
// Manages the pre-authentication flow: Login → SMS OTP → Onboarding

import SwiftUI

struct AuthCoordinator: View {
    @StateObject private var router = AuthRouter()
    
    var body: some View {
        ZStack {
            switch router.currentScreen {
            case .login:
                LoginScreen()
                    .transition(.opacity)
                
            case .smsVerification(let phone):
                SMSVerificationScreen(phoneNumber: phone)
                    .transition(.move(edge: .trailing))
                
            case .onboarding:
                OnboardingFlow()
                    .transition(.move(edge: .trailing))
            }
        }
        .animation(.glacial, value: router.currentScreen)
        .environmentObject(router)
        .onReceive(NotificationCenter.default.publisher(for: .navigateToOnboarding)) { _ in
            router.currentScreen = .onboarding
        }
    }
}

@MainActor
class AuthRouter: ObservableObject {
    @Published var currentScreen: AuthScreen = .login
    
    enum AuthScreen: Equatable {
        case login
        case smsVerification(phone: String)
        case onboarding
    }
}

// MARK: - SMS Verification Screen

struct SMSVerificationScreen: View {
    let phoneNumber: String
    @State private var otpCode: String = ""
    @State private var isVerifying = false
    @FocusState private var isOTPFocused: Bool
    
    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            
            VStack(spacing: 32) {
                Spacer().frame(height: 100)
                
                // Header
                VStack(spacing: 12) {
                    Image(systemName: "bubble.left.and.text.bubble.right")
                        .font(.system(size: 48, weight: .light))
                        .foregroundStyle(Color.Echo.primaryContainer)
                    
                    Text("Enter Verification Code")
                        .font(.custom("Inter", size: 24))
                        .fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)
                    
                    Text("Sent to \(phoneNumber)")
                        .font(.custom("Inter", size: 14))
                        .foregroundStyle(Color.Echo.outline)
                }
                
                // OTP Input
                ZStack {
                    RoundedRectangle(cornerRadius: 16)
                        .fill(Color.Echo.surfaceContainerLowest)
                    
                    RoundedRectangle(cornerRadius: 16)
                        .strokeBorder(
                            isOTPFocused
                                ? Color.Echo.primaryContainer.opacity(0.40)
                                : Color.Echo.outlineVariant.opacity(0.10),
                            lineWidth: 1
                        )
                    
                    TextField("000000", text: $otpCode)
                        .font(.custom("Inter", size: 32))
                        .fontWeight(.bold)
                        .tracking(12)
                        .multilineTextAlignment(.center)
                        .keyboardType(.numberPad)
                        .foregroundStyle(Color.Echo.onSurface)
                        .focused($isOTPFocused)
                        .padding(.horizontal, 24)
                }
                .frame(height: 72)
                .padding(.horizontal, 48)
                
                // Verify Button (Signature Gradient)
                Button {
                    isVerifying = true
                } label: {
                    Text("Verify")
                        .font(.custom("Inter", size: 18))
                        .fontWeight(.bold)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 18)
                        .background(
                            RoundedRectangle(cornerRadius: 9999)
                                .fill(LinearGradient.signature)
                        )
                        .deepGlacialShadow()
                }
                .padding(.horizontal, 48)
                .disabled(otpCode.count < 6)
                .opacity(otpCode.count < 6 ? 0.6 : 1.0)
                
                // Resend
                Button {
                    // Resend code
                } label: {
                    Text("Resend Code")
                        .font(.custom("Inter", size: 14))
                        .fontWeight(.medium)
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
                
                Spacer()
            }
            .padding(.horizontal, 24)
            
            // Secure thread line
            VStack {
                Rectangle()
                    .fill(Color.Echo.primaryContainer)
                    .frame(height: 2)
                    .blur(radius: 1)
                    .opacity(0.8)
                    .ignoresSafeArea()
                Spacer()
            }
        }
        .onAppear { isOTPFocused = true }
    }
}

// MARK: - Secure Thread Indicator (Reusable)

/// Custom component: thin 2px glowing line pulsating at top of screen.
/// Indicates active encrypted connection per DESIGN.md spec.
struct SecureThreadIndicator: View {
    @State private var opacity: Double = 0.6
    
    var body: some View {
        Rectangle()
            .fill(Color.Echo.primaryContainer)
            .frame(height: 2)
            .blur(radius: 1)
            .opacity(opacity)
            .ignoresSafeArea()
            .onAppear {
                withAnimation(.easeInOut(duration: 2.0).repeatForever(autoreverses: true)) {
                    opacity = 1.0
                }
            }
    }
}

// MARK: - Frosted Glass Navigation Bar (Reusable)

/// Glassmorphism nav bar per DESIGN.md: backdrop blur 20-32px, 
/// surface_variant at 60% opacity, ghost border at 15%.
struct GlacialNavigationBar<Content: View>: View {
    let content: () -> Content
    
    init(@ViewBuilder content: @escaping () -> Content) {
        self.content = content
    }
    
    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()
            
            HStack {
                content()
                Spacer()
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 16)
            .background(.ultraThinMaterial.opacity(0.6))
            .background(Color.white.opacity(0.6))
            .overlay(alignment: .bottom) {
                Rectangle()
                    .fill(Color.Echo.skyLight.opacity(0.15))
                    .frame(height: 1)
            }
            .shadow(color: Color.Echo.onSurface.opacity(0.04), radius: 32, x: 0, y: 8)
        }
        .ignoresSafeArea()
    }
}

#Preview("Login") {
    LoginScreen()
}

#Preview("SMS Verification") {
    SMSVerificationScreen(phoneNumber: "+1 (555) 123-4567")
}
