import SwiftUI

/// ECHO OTP Input Component
/// 6-digit input with auto-advance and paste support
public struct OTPInputView: View {
    @Binding var code: String
    let onComplete: (String) -> Void
    
    private let digitCount = 6
    @State private var digits: [String] = Array(repeating: "", count: 6)
    @FocusState private var focusedField: Int?
    
    public init(code: Binding<String>, onComplete: @escaping (String) -> Void) {
        self._code = code
        self.onComplete = onComplete
    }
    
    public var body: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            Text("Enter verification code")
                .typographyStyle(.h4, color: .echoPrimaryText)
            
            Text("We've sent a 6-digit code to your phone")
                .typographyStyle(.body, color: .echoSecondaryText)
            
            // OTP Input Grid
            HStack(spacing: Spacing.sm.rawValue) {
                ForEach(0..<digitCount, id: \.self) { index in
                    OTPDigitField(
                        digit: $digits[index],
                        isActive: focusedField == index,
                        isFilled: !digits[index].isEmpty,
                        onEditingChanged: { shouldFocus in
                            if shouldFocus {
                                focusedField = index
                            }
                        },
                        onBackspace: {
                            if digits[index].isEmpty && index > 0 {
                                focusedField = index - 1
                                digits[index - 1] = ""
                            }
                        }
                    )
                    .focused($focusedField, equals: index)
                    .onChange(of: digits[index]) { newValue in
                        // Only allow single digit
                        if newValue.count > 1 {
                            digits[index] = newValue.last.map(String.init) ?? ""
                        }
                        
                        // Only allow numeric input
                        if !newValue.isEmpty && !newValue.allSatisfy(\.isNumber) {
                            digits[index] = ""
                            return
                        }
                        
                        // Auto-advance to next field
                        if !newValue.isEmpty && index < digitCount - 1 {
                            focusedField = index + 1
                        }
                        
                        // Check if all filled
                        let fullCode = digits.joined()
                        code = fullCode
                        if fullCode.count == digitCount {
                            onComplete(fullCode)
                        }
                    }
                }
            }
            .frame(height: 56)
            
            // Resend Code
            HStack(spacing: Spacing.xs.rawValue) {
                Text("Didn't receive code?")
                    .typographyStyle(.body, color: .echoSecondaryText)
                
                Button(action: { resetOTP() }) {
                    Text("Resend")
                        .typographyStyle(.body, color: .echoPrimary)
                        .fontWeight(.semibold)
                }
                .accessibility(label: Text("Resend code"))
            }
        }
        .frame(maxWidth: .infinity)
        .echoSpacing(.lg)
    }
    
    private func resetOTP() {
        digits = Array(repeating: "", count: digitCount)
        code = ""
        focusedField = 0
    }
}

/// Individual OTP Digit Field
struct OTPDigitField: View {
    @Binding var digit: String
    let isActive: Bool
    let isFilled: Bool
    let onEditingChanged: (Bool) -> Void
    let onBackspace: () -> Void
    
    var body: some View {
        ZStack {
            // Background
            RoundedRectangle(cornerRadius: 8)
                .fill(Color.echoLightSurface)
                .stroke(
                    isActive ? Color.echoPrimary : Color.echoGray300,
                    lineWidth: isActive ? 2 : 1
                )
            
            // Digit Display or Cursor
            if !digit.isEmpty {
                Text(digit)
                    .typographyStyle(.h3, color: .echoPrimaryText)
                    .font(.system(size: 24, weight: .bold))
            } else if isActive {
                RoundedRectangle(cornerRadius: 2)
                    .fill(Color.echoPrimary)
                    .frame(width: 2, height: 20)
                    .opacity(0.7)
            }
        }
        .frame(height: 56)
        .frame(maxWidth: .infinity)
        .contentShape(Rectangle())
        .onTapGesture {
            onEditingChanged(true)
        }
        #if os(iOS)
        .onReceive(NotificationCenter.default.publisher(for: UIResponder.keyboardDidShowNotification)) { _ in
            if isActive {
                onEditingChanged(true)
            }
        }
        #endif
    }
}

// MARK: - Preview

#if DEBUG
struct OTPInputView_Previews: PreviewProvider {
    @State static var code = ""
    
    static var previews: some View {
        VStack {
            OTPInputView(code: $code) { fullCode in
                print("OTP Complete: \(fullCode)")
            }
            .echoSpacing(.lg)
            
            Spacer()
        }
        .frame(maxHeight: .infinity, alignment: .top)
        .background(Color.echoBackground)
    }
}
#endif
