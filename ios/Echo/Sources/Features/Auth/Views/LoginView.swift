import SwiftUI

struct AuthLoginView: View {
    @StateObject var viewModel: LoginViewModel
    let onLoggedIn: (AuthUserProfile) -> Void
    let onSwitchAccount: () -> Void
    let onRecover: () -> Void

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                // Logo
                Image(systemName: "bubble.right.fill")
                    .font(.system(size: 64, weight: .bold))
                    .foregroundColor(.echoPrimary)
                    .padding(.bottom, 24)

                Text("Welcome back")
                    .typographyStyle(.display, color: .echoPrimaryText)
                    .padding(.bottom, 32)

                // Sign In Button
                EchoButton(
                    "Sign In",
                    style: .primary,
                    size: .large,
                    isLoading: viewModel.isAuthenticating,
                    isDisabled: viewModel.isAuthenticating,
                    icon: Image(systemName: "faceid"),
                    action: {
                        Task {
                            if let user = await viewModel.loginWithPasskey() {
                                onLoggedIn(user)
                            }
                        }
                    }
                )

                // Account info
                VStack(spacing: 4) {
                    if !viewModel.username.isEmpty {
                        Text("Signed in as @\(viewModel.username)")
                            .typographyStyle(.caption, color: .echoSecondaryText)
                    }
                    if !viewModel.maskedPhone.isEmpty {
                        Text(viewModel.maskedPhone)
                            .typographyStyle(.caption, color: .echoGray500)
                    }
                }
                .padding(.top, 16)

                // Error
                if let error = viewModel.errorMessage {
                    Text(error)
                        .typographyStyle(.caption, color: .red)
                        .multilineTextAlignment(.center)
                        .padding(.top, 12)
                }

                Spacer()

                // Switch account
                Button("Not you? Switch account", action: onSwitchAccount)
                    .typographyStyle(.body, color: .echoPrimary)
                    .padding(.bottom, 16)

                Divider()
                    .padding(.horizontal, 24)

                // Recovery
                Button("Lost access? Recover", action: onRecover)
                    .typographyStyle(.caption, color: .echoSecondaryText)
                    .padding(.vertical, 16)
            }
            .padding(.horizontal, 24)
        }
        .onAppear { viewModel.loadStoredAccountInfo() }
    }
}
