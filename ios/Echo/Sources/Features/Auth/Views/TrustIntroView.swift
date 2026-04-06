import SwiftUI

struct TrustIntroView: View {
    let onDismiss: () -> Void

    var body: some View {
        ZStack {
            Color.echoBackground.ignoresSafeArea()

            VStack(spacing: 0) {
                ScrollView {
                    VStack(spacing: Spacing.xl.rawValue) {
                        // Header
                        VStack(spacing: Spacing.md.rawValue) {
                            Image(systemName: "shield.checkered")
                                .font(.system(size: 64))
                                .foregroundColor(.echoPrimary)

                            Text("Your Trust Score")
                                .typographyStyle(.display, color: .echoPrimaryText)

                            Text("Echo uses a transparent trust system to keep the community safe and reward positive participation.")
                                .typographyStyle(.body, color: .echoSecondaryText)
                                .multilineTextAlignment(.center)
                        }
                        .padding(.top, Spacing.xl.rawValue)

                        // Trust tiers
                        VStack(spacing: Spacing.md.rawValue) {
                            trustTierRow(
                                icon: "star",
                                color: .echoTrustNewcomer,
                                title: "Newcomer",
                                description: "Basic messaging. Complete verification to unlock more features."
                            )
                            trustTierRow(
                                icon: "star.fill",
                                color: .echoTrustBasic,
                                title: "Basic",
                                description: "Messaging + basic payments. Grow your trust score."
                            )
                            trustTierRow(
                                icon: "star.circle.fill",
                                color: .echoTrustTrusted,
                                title: "Trusted",
                                description: "Full payments, governance participation, staking."
                            )
                            trustTierRow(
                                icon: "checkmark.seal.fill",
                                color: .echoTrustVerified,
                                title: "Verified",
                                description: "All features unlocked. Higher earning multipliers."
                            )
                        }

                        // How to build trust
                        VStack(alignment: .leading, spacing: Spacing.sm.rawValue) {
                            Text("How to build trust:")
                                .typographyStyle(.h4, color: .echoPrimaryText)

                            buildTrustItem("Verify your phone number and passkey")
                            buildTrustItem("Send messages and stay active")
                            buildTrustItem("Get vouched for by trusted users")
                            buildTrustItem("Participate in governance")
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                }

                // Continue button
                EchoButton(
                    "Get Started",
                    style: .primary,
                    size: .large,
                    action: onDismiss
                )
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.bottom, Spacing.lg.rawValue)
            }
        }
    }

    private func trustTierRow(
        icon: String,
        color: Color,
        title: String,
        description: String
    ) -> some View {
        HStack(spacing: Spacing.md.rawValue) {
            Image(systemName: icon)
                .font(.system(size: 24))
                .foregroundColor(color)
                .frame(width: 40, height: 40)

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .typographyStyle(.bodyLarge, color: .echoPrimaryText)
                Text(description)
                    .typographyStyle(.caption, color: .echoSecondaryText)
            }

            Spacer()
        }
        .padding(Spacing.md.rawValue)
        .background(Color.echoSurface)
        .cornerRadius(12)
    }

    private func buildTrustItem(_ text: String) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "checkmark.circle")
                .font(.system(size: 16))
                .foregroundColor(.echoPrimary)
            Text(text)
                .typographyStyle(.body, color: .echoSecondaryText)
        }
    }
}
