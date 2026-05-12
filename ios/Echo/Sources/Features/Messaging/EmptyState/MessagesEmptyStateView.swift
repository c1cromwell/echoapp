#if os(iOS)
// Features/Messaging/EmptyState/MessagesEmptyStateView.swift
// Shown in the Messages pane when the user has zero conversations after first run.
//
// Composition:
//   • Centered ECHO logo mark
//   • "Welcome to ECHO, {name}" title
//   • One-sentence instruction
//   • TrustTierBanner below — tappable deep-link to credential upgrade
//   • Right-aligned 58pt FAB at the bottom, green gradient + glow shadow

import SwiftUI

public struct MessagesEmptyStateView: View {
    let displayName: String
    let trustTier: Int
    let onComposeTapped: () -> Void
    let onUpgradeTrustTapped: () -> Void

    public init(
        displayName: String,
        trustTier: Int,
        onComposeTapped: @escaping () -> Void = {},
        onUpgradeTrustTapped: @escaping () -> Void = {}
    ) {
        self.displayName = displayName
        self.trustTier = trustTier
        self.onComposeTapped = onComposeTapped
        self.onUpgradeTrustTapped = onUpgradeTrustTapped
    }

    public var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            centerContent

            ComposeFAB(onTap: onComposeTapped)
                .padding(.trailing, 20)
                .padding(.bottom, 26)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomTrailing)
        }
        .accessibilityElement(children: .contain)
    }

    // MARK: - Center content

    private var centerContent: some View {
        VStack(spacing: 0) {
            Spacer()

            EchoLogo(size: 56)
                .opacity(0.88)
                .padding(.bottom, 18)

            Text("Welcome to ECHO, \(displayName)")
                .font(.system(size: 18, weight: .semibold))
                .kerning(-0.2)
                .foregroundStyle(Color.Echo.onSurface)
                .multilineTextAlignment(.center)
                .padding(.bottom, 8)

            Text("Tap the + button to start your first conversation — by QR code, ECHO handle, or invite link.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .multilineTextAlignment(.center)
                .lineSpacing(2)
                .frame(maxWidth: 260)
                .padding(.bottom, 24)

            TrustTierBanner(tier: trustTier)
                .onTapGesture { onUpgradeTrustTapped() }
                .accessibilityHint("Opens credential enrollment to upgrade your trust tier")

            Spacer()
        }
        .padding(.horizontal, 28)
    }
}

// MARK: - Trust Tier Banner

/// Compact inline banner showing the user's current trust tier with an upgrade CTA.
struct TrustTierBanner: View {
    let tier: Int

    var body: some View {
        HStack(spacing: 10) {
            TrustTierBadge(tier: tier, size: 22)

            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(Color.Echo.onSurface)
                Text(subtitle)
                    .font(.system(size: 10.5))
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
            Spacer(minLength: 0)

            if tier <= 2 {
                Image(systemName: "chevron.right")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundStyle(tint.opacity(0.65))
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .frame(maxWidth: 260)
        .background(tint.opacity(0.08), in: RoundedRectangle(cornerRadius: 14, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .strokeBorder(tint.opacity(0.28), lineWidth: 0.5)
        )
        .accessibilityElement(children: .combine)
    }

    private var title: String {
        switch tier {
        case 0: return "Unverified · Tier 0"
        case 1: return "Standard trust · Tier 1"
        case 2: return "Phone verified · Tier 2"
        case 3: return "Active member · Tier 3"
        case 4: return "Credential verified · Tier 4"
        default: return "Identity verified · Tier 5"
        }
    }

    private var subtitle: String {
        switch tier {
        case 0: return "Provisioning your identity…"
        case 1, 2: return "Add a credential to upgrade →"
        case 3: return "Add a credential for higher trust →"
        case 4: return "Present a government ID for Tier 5 →"
        default: return "Maximum trust reached"
        }
    }

    private var tint: Color {
        switch tier {
        case 0:       return Color.Echo.onSurfaceVariant
        case 1, 2:    return Color.Echo.trustYellow
        case 3, 4:    return Color.Echo.primaryContainer
        default:      return Color.Echo.trustPremium
        }
    }
}

// MARK: - Compose FAB

/// Reusable green + FAB. Same gradient and glow used by MessagesEmptyStateView
/// and the populated conversation list overlay.
struct ComposeFAB: View {
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            ZStack {
                Circle()
                    .fill(Color.Echo.compose.opacity(0.30))
                    .frame(width: 76, height: 76)
                    .blur(radius: 12)

                ZStack {
                    Circle()
                        .fill(
                            LinearGradient(
                                colors: [Color.Echo.composeDeep, Color.Echo.compose],
                                startPoint: .topLeading, endPoint: .bottomTrailing
                            )
                        )
                        .overlay(
                            Circle().strokeBorder(Color.Echo.compose.opacity(0.5), lineWidth: 0.5)
                        )

                    Image(systemName: "plus")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(.white)
                }
                .frame(width: 58, height: 58)
            }
        }
        .buttonStyle(SpringPressStyle())
        .accessibilityLabel("Start a new conversation")
    }
}

#endif
