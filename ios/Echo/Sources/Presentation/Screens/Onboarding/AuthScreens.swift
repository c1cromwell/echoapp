import SwiftUI

/// Welcome/Splash Screen - First screen of onboarding
public struct WelcomeView: View {
    @Environment(\.dismiss) var dismiss
    let onContinueWithPhone: () -> Void
    let onUseVerifiableID: () -> Void
    
    public init(
        onContinueWithPhone: @escaping () -> Void = {},
        onUseVerifiableID: @escaping () -> Void = {}
    ) {
        self.onContinueWithPhone = onContinueWithPhone
        self.onUseVerifiableID = onUseVerifiableID
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: Spacing.lg.rawValue) {
                VStack(spacing: Spacing.xl.rawValue) {
                    // Logo/Icon
                    Image(systemName: "bubble.right.fill")
                        .font(.system(size: 64, weight: .bold))
                        .foregroundColor(.echoPrimary)
                    
                    VStack(spacing: Spacing.md.rawValue) {
                        Text("Welcome to Echo")
                            .typographyStyle(.display, color: .echoPrimaryText)
                        
                        Text("Secure messaging with verifiable identity")
                            .typographyStyle(.bodyLarge, color: .echoSecondaryText)
                            .multilineTextAlignment(.center)
                    }
                }
                .frame(maxHeight: .infinity, alignment: .center)
                
                VStack(spacing: Spacing.md.rawValue) {
                    // Primary CTA
                    EchoButton(
                        "Continue with Phone Number",
                        style: .primary,
                        size: .large,
                        icon: Image(systemName: "phone.fill"),
                        action: onContinueWithPhone
                    )
                    
                    // Secondary CTA
                    EchoButton(
                        "Use Verifiable ID",
                        style: .secondary,
                        size: .large,
                        icon: Image(systemName: "checkmark.shield.fill"),
                        action: onUseVerifiableID
                    )
                    
                    // Privacy Badge
                    HStack(spacing: Spacing.xs.rawValue) {
                        Image(systemName: "lock.shield.fill")
                            .font(.system(size: 12))
                        
                        Text("End-to-end encrypted")
                            .typographyStyle(.caption, color: .echoSecondaryText)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(Spacing.sm.rawValue)
                    .background(Color.echoSurface)
                    .cornerRadius(8)
                }
            }
            .echoSpacing(.lg)
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - Phone Entry Screen

public struct PhoneEntryView: View {
    @Environment(\.dismiss) var dismiss
    @State private var phone = ""
    @State private var selectedCountry = "US"
    @State private var isLoading = false
    
    let onSendCode: (String) -> Void
    
    public init(onSendCode: @escaping (String) -> Void = { _ in }) {
        self.onSendCode = onSendCode
    }
    
    var isValidPhone: Bool {
        phone.count >= 10 && phone.allSatisfy(\.isNumber)
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: Spacing.lg.rawValue) {
                EchoNavBar(
                    title: "Enter Phone Number",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )
                
                VStack(spacing: Spacing.xl.rawValue) {
                    VStack(spacing: Spacing.md.rawValue) {
                        Text("We'll send you a verification code")
                            .typographyStyle(.bodyLarge, color: .echoSecondaryText)
                    }
                    
                    VStack(spacing: Spacing.md.rawValue) {
                        // Country Picker
                        HStack {
                            Text("Country")
                                .typographyStyle(.caption, color: .echoSecondaryText)
                            
                            Spacer()
                            
                            Picker("Country", selection: $selectedCountry) {
                                Text("US +1").tag("US")
                                Text("UK +44").tag("UK")
                                Text("CA +1").tag("CA")
                                Text("AU +61").tag("AU")
                            }
                            .pickerStyle(.automatic)
                        }
                        .padding(Spacing.md.rawValue)
                        .background(Color.echoSurface)
                        .cornerRadius(12)
                        
                        // Phone Input
                        EchoTextField(
                            label: "Phone Number",
                            placeholder: "(555) 123-4567",
                            text: $phone,
                            maxLength: 20
                        )
                    }
                    
                    Spacer()
                }
                
                VStack(spacing: Spacing.md.rawValue) {
                    EchoButton(
                        "Send Verification Code",
                        style: .primary,
                        size: .large,
                        isLoading: isLoading,
                        isDisabled: !isValidPhone || isLoading,
                        action: {
                            isLoading = true
                            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                                isLoading = false
                                onSendCode(phone)
                            }
                        }
                    )
                    
                    VStack(spacing: Spacing.xs.rawValue) {
                        Text("By continuing, you agree to our")
                            .typographyStyle(.caption, color: .echoSecondaryText)
                        
                        HStack(spacing: Spacing.xs.rawValue) {
                            Link("Terms of Service", destination: URL(string: "https://example.com")!)
                                .typographyStyle(.caption, color: .echoPrimary)
                            
                            Text("and")
                                .typographyStyle(.caption, color: .echoSecondaryText)
                            
                            Link("Privacy Policy", destination: URL(string: "https://example.com")!)
                                .typographyStyle(.caption, color: .echoPrimary)
                        }
                        .multilineTextAlignment(.center)
                    }
                }
            }
            .echoSpacing(.lg)
        }
        .navigationBarBackButtonHidden(true)
    }
}

// MARK: - OTP Verification Screen

public struct OTPVerificationView: View {
    @Environment(\.dismiss) var dismiss
    @State private var code = ""
    @State private var secondsRemaining = 45
    @State private var timer: Timer?
    
    let phoneNumber: String
    let onVerify: (String) -> Void
    let onResendCode: () -> Void
    
    public init(
        phoneNumber: String = "",
        onVerify: @escaping (String) -> Void = { _ in },
        onResendCode: @escaping () -> Void = {}
    ) {
        self.phoneNumber = phoneNumber
        self.onVerify = onVerify
        self.onResendCode = onResendCode
    }
    
    public var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()
            
            VStack(spacing: Spacing.lg.rawValue) {
                EchoNavBar(
                    title: "Verify Code",
                    showBackButton: true,
                    onBackPressed: { dismiss() }
                )
                
                VStack(spacing: Spacing.xl.rawValue) {
                    VStack(spacing: Spacing.md.rawValue) {
                        Text("Code sent to")
                            .typographyStyle(.bodyLarge, color: .echoSecondaryText)
                        
                        Text(phoneNumber)
                            .typographyStyle(.h4, color: .echoPrimaryText)
                        
                        Button(action: { dismiss() }) {
                            Text("Change number")
                                .typographyStyle(.body, color: .echoPrimary)
                        }
                    }
                    
                    OTPInputView(code: $code) { fullCode in
                        onVerify(fullCode)
                    }
                    
                    // Resend section
                    VStack(spacing: Spacing.xs.rawValue) {
                        if secondsRemaining > 0 {
                            Text("Resend code in \(secondsRemaining)s")
                                .typographyStyle(.caption, color: .echoGray500)
                        } else {
                            Button(action: {
                                onResendCode()
                                secondsRemaining = 45
                                startTimer()
                            }) {
                                Text("Resend code")
                                    .typographyStyle(.body, color: .echoPrimary)
                                    .fontWeight(.semibold)
                            }
                        }
                    }
                    
                    Spacer()
                }
            }
            .echoSpacing(.lg)
        }
        .navigationBarBackButtonHidden(true)
        .onAppear {
            startTimer()
        }
        .onDisappear {
            timer?.invalidate()
        }
    }
    
    private func startTimer() {
        timer?.invalidate()
        timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
            if secondsRemaining > 0 {
                secondsRemaining -= 1
            }
        }
    }
}

// MARK: - Preview

#if DEBUG
struct OnboardingScreens_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack {
            WelcomeView()
        }
    }
}
#endif
