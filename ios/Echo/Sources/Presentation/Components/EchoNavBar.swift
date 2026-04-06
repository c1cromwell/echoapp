import SwiftUI

import SwiftUI

/// ECHO Custom Navigation Bar Component
/// Back button, title, and trailing action
public struct EchoNavBar: View {
    let title: String
    let showBackButton: Bool
    let onBackPressed: () -> Void
    let trailingAction: (() -> Void)?
    let trailingIcon: Image?
    
    public init(
        title: String,
        showBackButton: Bool = true,
        onBackPressed: @escaping () -> Void = {},
        trailingAction: (() -> Void)? = nil,
        trailingIcon: Image? = nil
    ) {
        self.title = title
        self.showBackButton = showBackButton
        self.onBackPressed = onBackPressed
        self.trailingAction = trailingAction
        self.trailingIcon = trailingIcon
    }
    
    public var body: some View {
        HStack(spacing: Spacing.md.rawValue) {
            // Back Button
            if showBackButton {
                Button(action: onBackPressed) {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundColor(.echoPrimaryText)
                        .frame(width: 44, height: 44)
                }
                .accessibility(label: Text("Back"))
            }
            
            // Title
            Text(title)
                .typographyStyle(.h4, color: .echoPrimaryText)
                .lineLimit(1)
            
            Spacer()
            
            // Trailing Action
            if let trailingIcon = trailingIcon, let action = trailingAction {
                Button(action: action) {
                    trailingIcon
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.echoPrimaryText)
                        .frame(width: 44, height: 44)
                }
                .accessibility(label: Text("Action"))
            }
        }
        .frame(height: 56)
        .padding(.horizontal, Spacing.md.rawValue)
        .background(Color.echoSurface)
        .border(Color.echoGray200, width: 1)
    }
}

// MARK: - Preview

#if DEBUG
struct EchoNavBar_Previews: PreviewProvider {
    static var previews: some View {
        VStack {
            EchoNavBar(
                title: "Messages",
                showBackButton: false
            )
            
            EchoNavBar(
                title: "Chat Details",
                showBackButton: true,
                onBackPressed: {}
            )
            
            EchoNavBar(
                title: "Contact Info",
                showBackButton: true,
                onBackPressed: {},
                trailingAction: {},
                trailingIcon: Image(systemName: "ellipsis")
            )
            
            Spacer()
        }
        .background(Color.echoBackground)
    }
}
#endif
