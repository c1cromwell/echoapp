import SwiftUI

/// Text field state enumeration
public enum EchoTextFieldState {
    case `default`
    case focused
    case error(String?)
    case success
    case disabled
}

/// ECHO Custom Text Field Component
/// Supports 7 states with secure input, prefix/suffix, and character counting
public struct EchoTextField: View {
    let label: String?
    let placeholder: String
    @Binding var text: String
    let state: EchoTextFieldState
    let isSecure: Bool
    let maxLength: Int?
    let prefix: String?
    let suffix: String?
    let helperText: String?
    let onEditingChanged: (Bool) -> Void
    
    @State private var isFocused = false
    @State private var showPassword = false
    
    public init(
        label: String? = nil,
        placeholder: String = "",
        text: Binding<String>,
        state: EchoTextFieldState = .default,
        isSecure: Bool = false,
        maxLength: Int? = nil,
        prefix: String? = nil,
        suffix: String? = nil,
        helperText: String? = nil,
        onEditingChanged: @escaping (Bool) -> Void = { _ in }
    ) {
        self.label = label
        self.placeholder = placeholder
        self._text = text
        self.state = state
        self.isSecure = isSecure
        self.maxLength = maxLength
        self.prefix = prefix
        self.suffix = suffix
        self.helperText = helperText
        self.onEditingChanged = onEditingChanged
    }
    
    var borderColor: Color {
        switch state {
        case .default, .disabled:
            return isFocused ? .echoPrimary : .echoGray300
        case .focused:
            return .echoPrimary
        case .error:
            return .echoError
        case .success:
            return .echoSuccess
        }
    }
    
    var backgroundColor: Color {
        switch state {
        case .disabled:
            return .echoGray100
        default:
            return .echoLightSurface
        }
    }
    
    var textColor: Color {
        switch state {
        case .disabled:
            return .echoGray400
        default:
            return .echoPrimaryText
        }
    }
    
    var isDisabled: Bool {
        if case .disabled = state {
            return true
        }
        return false
    }
    
    public var body: some View {
        VStack(alignment: .leading, spacing: Spacing.xs.rawValue) {
            // Label
            if let label = label {
                Text(label)
                    .typographyStyle(.caption, color: .echoSecondaryText)
                    .accessibility(label: Text(label))
            }
            
            // Text Field Container
            HStack(spacing: Spacing.sm.rawValue) {
                // Prefix
                if let prefix = prefix {
                    Text(prefix)
                        .typographyStyle(.body, color: .echoGray500)
                }
                
                // Text Input
                if isSecure && !showPassword {
                    SecureField(placeholder, text: $text)
                        .typographyStyle(.body, color: textColor)
                        .disabled(isDisabled)
                        .onChange(of: text) { newValue in
                            if let maxLength = maxLength, newValue.count > maxLength {
                                text = String(newValue.prefix(maxLength))
                            }
                        }
                        .onFocusChange { focused in
                            isFocused = focused
                            onEditingChanged(focused)
                        }
                } else {
                    TextField(placeholder, text: $text)
                        .typographyStyle(.body, color: textColor)
                        .disabled(isDisabled)
                        .onChange(of: text) { newValue in
                            if let maxLength = maxLength, newValue.count > maxLength {
                                text = String(newValue.prefix(maxLength))
                            }
                        }
                        .onFocusChange { focused in
                            isFocused = focused
                            onEditingChanged(focused)
                        }
                }
                
                // Suffix and Controls
                HStack(spacing: Spacing.xs.rawValue) {
                    // Character Counter
                    if let maxLength = maxLength {
                        Text("\(text.count)/\(maxLength)")
                            .typographyStyle(.tiny, color: .echoGray400)
                    }
                    
                    // Password Toggle
                    if isSecure {
                        Button(action: { showPassword.toggle() }) {
                            Image(systemName: showPassword ? "eye.slash.fill" : "eye.fill")
                                .font(.system(size: 16))
                                .foregroundColor(.echoGray500)
                        }
                        .accessibility(label: Text(showPassword ? "Hide password" : "Show password"))
                    }
                    
                    // Success indicator
                    if case .success = state {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 18))
                            .foregroundColor(.echoSuccess)
                            .accessibility(label: Text("Verified"))
                    }
                    
                    // Error indicator
                    if case .error = state {
                        Image(systemName: "exclamationmark.circle.fill")
                            .font(.system(size: 18))
                            .foregroundColor(.echoError)
                            .accessibility(label: Text("Error"))
                    }
                    
                    // Custom Suffix
                    if let suffix = suffix {
                        Text(suffix)
                            .typographyStyle(.body, color: .echoGray500)
                    }
                }
            }
            .frame(height: 44)
            .padding(.horizontal, Spacing.md.rawValue)
            .background(backgroundColor)
            .border(borderColor, width: 1)
            .cornerRadius(12)
            
            // Helper text / Error message
            if let helperText = helperText, case .error(let errorMsg) = state {
                Text(errorMsg ?? helperText)
                    .typographyStyle(.caption, color: .echoError)
                    .accessibility(label: Text("Error: \(errorMsg ?? helperText)"))
            } else if let helperText = helperText {
                Text(helperText)
                    .typographyStyle(.caption, color: .echoGray500)
            }
        }
    }
}

// MARK: - FocusModifier Helper

struct FocusChangeModifier: ViewModifier {
    let handler: (Bool) -> Void
    @FocusState private var isFocused: Bool
    
    func body(content: Content) -> some View {
        content
            .focused($isFocused)
            .onChange(of: isFocused) { newValue in
                handler(newValue)
            }
    }
}

extension View {
    func onFocusChange(perform: @escaping (Bool) -> Void) -> some View {
        modifier(FocusChangeModifier(handler: perform))
    }
}

// MARK: - Custom Border

struct BorderShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let cornerRadius: CGFloat = 12
        
        path.move(to: CGPoint(x: cornerRadius, y: 0))
        path.addLine(to: CGPoint(x: rect.width - cornerRadius, y: 0))
        path.addQuadCurve(to: CGPoint(x: rect.width, y: cornerRadius), control: CGPoint(x: rect.width, y: 0))
        path.addLine(to: CGPoint(x: rect.width, y: rect.height - cornerRadius))
        path.addQuadCurve(to: CGPoint(x: rect.width - cornerRadius, y: rect.height), control: CGPoint(x: rect.width, y: rect.height))
        path.addLine(to: CGPoint(x: cornerRadius, y: rect.height))
        path.addQuadCurve(to: CGPoint(x: 0, y: rect.height - cornerRadius), control: CGPoint(x: 0, y: rect.height))
        path.addLine(to: CGPoint(x: 0, y: cornerRadius))
        path.addQuadCurve(to: CGPoint(x: cornerRadius, y: 0), control: CGPoint(x: 0, y: 0))
        
        return path
    }
}

extension View {
    func border(_ color: Color, width: CGFloat) -> some View {
        self.overlay(
            BorderShape()
                .stroke(color, lineWidth: width)
        )
    }
}

// MARK: - Preview

#if DEBUG
struct EchoTextField_Previews: PreviewProvider {
    @State static var text = ""
    @State static var secureText = ""
    
    static var previews: some View {
        VStack(spacing: Spacing.lg.rawValue) {
            // Default state
            EchoTextField(
                label: "Full Name",
                placeholder: "Enter your name",
                text: $text
            )
            
            // Focused state
            EchoTextField(
                label: "Email",
                placeholder: "user@example.com",
                text: $text,
                state: .focused
            )
            
            // Success state
            EchoTextField(
                label: "Username",
                placeholder: "johndoe",
                text: $text,
                state: .success,
                helperText: "Available"
            )
            
            // Error state
            EchoTextField(
                label: "Password",
                placeholder: "••••••••",
                text: $secureText,
                state: .error("Password must be at least 8 characters"),
                isSecure: true
            )
            
            // With max length
            EchoTextField(
                label: "Bio",
                placeholder: "Tell us about yourself",
                text: $text,
                maxLength: 150,
                helperText: "Max 150 characters"
            )
            
            Spacer()
        }
        .echoSpacing(.lg)
        .background(Color.echoBackground)
    }
}
#endif
