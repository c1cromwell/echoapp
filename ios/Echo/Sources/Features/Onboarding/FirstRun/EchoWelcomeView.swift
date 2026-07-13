#if os(iOS)
import SwiftUI

public struct EchoWelcomeView: View {
    let onSetUp: () -> Void
    let onAlreadyHaveAccount: () -> Void

    public init(onSetUp: @escaping () -> Void, onAlreadyHaveAccount: @escaping () -> Void) {
        self.onSetUp = onSetUp; self.onAlreadyHaveAccount = onAlreadyHaveAccount
    }

    public var body: some View {
        ZStack {
            Color.echoPaper.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                EchoRippleMark(size: 160, color: .echoSignal)
                    .padding(.bottom, 32)

                Text("ECHO")
                    .font(.system(size: 36, weight: .bold))
                    .tracking(1.5)
                    .foregroundStyle(Color.echoInk)
                    .padding(.bottom, 16)

                Text("Always Secure, Always Private")
                    .font(.system(size: 16))
                    .foregroundStyle(Color.echoInk70)

                Text("Your Data, Your Community")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(Color.echoInk70)
                    .padding(.top, 4)

                Spacer()
                Spacer()

                VStack(spacing: 12) {
                    Button(action: onSetUp) {
                        HStack(spacing: 8) {
                            Text("Let's Go")
                                .font(.system(size: 17, weight: .semibold))
                            Image(systemName: "arrow.right")
                                .font(.system(size: 15, weight: .semibold))
                        }
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(Color.echoSignal, in: RoundedRectangle(cornerRadius: 14))
                    }
                    .accessibilityIdentifier("welcome.getStarted")

                    Button("I already have an account", action: onAlreadyHaveAccount)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(Color.echoInk55)
                        .padding(.vertical, 6)
                        .accessibilityIdentifier("welcome.haveAccount")
                }
                .padding(.horizontal, 28)
                .padding(.bottom, 16)
            }
        }
        .preferredColorScheme(.light)
    }
}

#Preview {
    EchoWelcomeView(onSetUp: {}, onAlreadyHaveAccount: {})
}
#endif
