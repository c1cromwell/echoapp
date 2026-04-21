// Features/Auth/Enrollment/Shared/EnrollmentComponents.swift
//
// Small reusable SwiftUI components shared by the enrollment screens.
// Note: SecureThreadIndicator lives in DesignSystem/Glacial/Components/ and
// is not duplicated here.

import SwiftUI
import AuthenticationServices

// MARK: - Atmospheric Background

/// Two blurred circles that give the "frozen lake depth" feel. Must sit
/// behind other content via ZStack.
struct AtmosphericBackground: View {
    var body: some View {
        ZStack {
            Circle()
                .fill(Color.Echo.primaryContainer.opacity(0.10))
                .frame(width: 320, height: 320)
                .blur(radius: 120)
                .offset(x: -120, y: -280)
            Circle()
                .fill(Color.Echo.secondaryContainer.opacity(0.10))
                .frame(width: 280, height: 280)
                .blur(radius: 100)
                .offset(x: 140, y: 260)
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }
}

// MARK: - Section Divider with Tracked Label

/// Editorial small-caps label flanked by 1px ghost lines.
struct TrackedSectionDivider: View {
    let label: String
    let tint: Color

    var body: some View {
        HStack(spacing: 10) {
            Rectangle()
                .fill(tint.opacity(0.22))
                .frame(height: 1)
            Text(label)
                .font(.system(size: 11, weight: .semibold))
                .tracking(1.8)
                .foregroundStyle(tint)
            Rectangle()
                .fill(tint.opacity(0.22))
                .frame(height: 1)
        }
        .accessibilityElement(children: .combine)
    }
}

// MARK: - Enrollment Card Styles

/// Three visual tiers for enrollment option cards, mapped to promotion order.
enum EnrollmentCardStyle {
    case primaryGradient     // Most promoted — navy→sky gradient fill
    case secondaryTinted     // Frosted white with sky-tinted ghost border
    case tertiaryNeutral     // Muted surface, de-emphasized
}

struct EnrollmentOptionCard<Icon: View>: View {
    let style: EnrollmentCardStyle
    let title: String
    let subtitle: String
    let tags: [String]
    let badge: String?               // "New", "Recommended"
    let icon: Icon
    let action: () -> Void

    init(
        style: EnrollmentCardStyle,
        title: String,
        subtitle: String,
        tags: [String] = [],
        badge: String? = nil,
        @ViewBuilder icon: () -> Icon,
        action: @escaping () -> Void
    ) {
        self.style = style
        self.title = title
        self.subtitle = subtitle
        self.tags = tags
        self.badge = badge
        self.icon = icon()
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            HStack(alignment: .top, spacing: 12) {
                iconBadge

                VStack(alignment: .leading, spacing: 3) {
                    Text(title)
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(primaryTextColor)
                    Text(subtitle)
                        .font(.system(size: 12.5))
                        .foregroundStyle(secondaryTextColor)
                        .lineSpacing(1.5)
                        .multilineTextAlignment(.leading)

                    if !tags.isEmpty {
                        HStack(spacing: 6) {
                            ForEach(tags, id: \.self) { tag in
                                Text(tag)
                                    .font(.system(size: 11, weight: .semibold))
                                    .foregroundStyle(primaryTextColor)
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 2)
                                    .background(tagFillColor)
                                    .clipShape(Capsule())
                            }
                        }
                        .padding(.top, 6)
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)

                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(chevronColor)
                    .padding(.top, 4)
            }
            .padding(14)
            .background {
                cardBackground
                    .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
            }
            .overlay(cardBorder)
            .overlay(alignment: .topTrailing) { badgePill }
            .contentShape(RoundedRectangle(cornerRadius: 22))
        }
        .buttonStyle(SpringPressStyle())
        .accessibilityElement(children: .combine)
        .accessibilityLabel("\(title). \(subtitle).\(tags.isEmpty ? "" : " \(tags.joined(separator: ", ")).")")
    }

    // MARK: - Style mapping

    @ViewBuilder private var iconBadge: some View {
        ZStack {
            Circle().fill(iconFillColor)
            Circle().stroke(iconStrokeColor, lineWidth: 0.5)
            icon
        }
        .frame(width: 34, height: 34)
    }

    @ViewBuilder private var cardBackground: some View {
        switch style {
        case .primaryGradient:
            LinearGradient(
                colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        case .secondaryTinted:
            Color.Echo.surfaceContainerLowest
        case .tertiaryNeutral:
            Color.Echo.surfaceContainerLow.opacity(0.7)
        }
    }

    @ViewBuilder private var cardBorder: some View {
        RoundedRectangle(cornerRadius: 22, style: .continuous)
            .strokeBorder(
                style == .secondaryTinted
                    ? Color.Echo.primaryContainer.opacity(0.28)
                    : Color.clear,
                lineWidth: 0.5
            )
    }

    @ViewBuilder private var badgePill: some View {
        if let badge {
            Text(badge)
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(.white)
                .tracking(0.4)
                .padding(.horizontal, 9)
                .padding(.vertical, 2)
                .background(Color.Echo.primaryContainer, in: Capsule())
                .offset(x: -12, y: -8)
        }
    }

    private var primaryTextColor: Color {
        style == .primaryGradient ? .white : Color.Echo.onSurface
    }
    private var secondaryTextColor: Color {
        style == .primaryGradient
            ? Color.white.opacity(0.82)
            : Color.Echo.onSurfaceVariant
    }
    private var chevronColor: Color {
        style == .primaryGradient
            ? Color.white.opacity(0.65)
            : Color.Echo.primaryContainer
    }
    private var iconFillColor: Color {
        style == .primaryGradient
            ? Color.white.opacity(0.13)
            : Color.Echo.primaryContainer.opacity(0.12)
    }
    private var iconStrokeColor: Color {
        style == .primaryGradient
            ? Color.white.opacity(0.25)
            : Color.clear
    }
    private var tagFillColor: Color {
        style == .primaryGradient
            ? Color.white.opacity(0.2)
            : Color.Echo.primaryContainer.opacity(0.14)
    }
}

// MARK: - Spring Press Button Style

struct SpringPressStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.96 : 1.0)
            .animation(.spring(response: 0.28, dampingFraction: 0.85), value: configuration.isPressed)
    }
}

// MARK: - Get Started Card

/// Replaces the plain "New to ECHO? Get Started" footer link on the login screen.
/// Shows a ghost-outline card with a credential icon and a "New" badge.
struct GetStartedCard: View {
    let action: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 10) {
                Text("New to ECHO")
                    .font(.system(size: 11, weight: .semibold))
                    .tracking(1.8)
                    .foregroundStyle(Color.Echo.primaryContainer)
                Rectangle()
                    .fill(Color.Echo.primaryContainer.opacity(0.22))
                    .frame(height: 1)
            }

            Button(action: action) {
                HStack(spacing: 11) {
                    ZStack {
                        Circle().fill(Color.Echo.primaryContainer.opacity(0.14))
                        Image(systemName: "person.badge.key.fill")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(Color.Echo.primaryContainer)
                    }
                    .frame(width: 32, height: 32)

                    VStack(alignment: .leading, spacing: 2) {
                        Text("Get started with a credential")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(Color.Echo.onSurface)
                        Text("Wallet, driver's license, or phone")
                            .font(.system(size: 11.5))
                            .foregroundStyle(Color.Echo.onSurfaceVariant)
                    }
                    Spacer()
                    Image(systemName: "chevron.right")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
                .padding(14)
                .background(
                    Color.Echo.primaryContainer.opacity(0.06),
                    in: RoundedRectangle(cornerRadius: 22, style: .continuous)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: 22, style: .continuous)
                        .strokeBorder(Color.Echo.primaryContainer.opacity(0.38), lineWidth: 0.5)
                )
                .overlay(alignment: .topTrailing) {
                    Text("New")
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundStyle(.white)
                        .tracking(0.4)
                        .padding(.horizontal, 9)
                        .padding(.vertical, 2)
                        .background(Color.Echo.primaryContainer, in: Capsule())
                        .offset(x: -12, y: -8)
                }
                .contentShape(RoundedRectangle(cornerRadius: 22))
            }
            .buttonStyle(SpringPressStyle())
        }
    }
}

// MARK: - ASWebAuthentication Presentation Context

/// Shared presentation context provider for ASWebAuthenticationSession.
final class AuthPresentationContextProvider: NSObject, ASWebAuthenticationPresentationContextProviding {
    static let shared = AuthPresentationContextProvider()

    func presentationAnchor(for session: ASWebAuthenticationSession) -> ASPresentationAnchor {
        let scenes = UIApplication.shared.connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .filter { $0.activationState == .foregroundActive }
        return scenes.first?.windows.first(where: \.isKeyWindow)
            ?? ASPresentationAnchor()
    }
}
