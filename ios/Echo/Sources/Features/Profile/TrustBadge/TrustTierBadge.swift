// Features/Profile/TrustBadge/TrustTierBadge.swift
// Six-state tier badge used on Messages header, Profile, and MessagesEmptyState.
//
// Tier 0  Grey dashed circle, no check     Provisioning / pre-completion
// Tier 1  Amber solid check                Standard trust (display name only)
// Tier 2  Amber solid check                Phone verified (same visual as Tier 1)
// Tier 3  Sky blue solid check             Active member
// Tier 4  Sky blue filled badge with halo  Credential verified
// Tier 5  Gradient-ringed check            IAL2+ identity verified
//
// Named TrustTierBadge to distinguish from the existing string-based TrustBadge pill.

import SwiftUI

struct TrustTierBadge: View {
    let tier: Int
    let size: CGFloat

    var body: some View {
        content
            .frame(width: size, height: size)
            .accessibilityLabel(accessibilityText)
    }

    @ViewBuilder
    private var content: some View {
        switch tier {
        case 0:  tier0
        case 1, 2: tier1_2
        case 3:  tier3
        case 4:  tier4
        default: tier5
        }
    }

    // MARK: Tier 0 — dashed grey circle

    private var tier0: some View {
        Circle()
            .strokeBorder(
                Color.Echo.onSurfaceVariant.opacity(0.45),
                style: StrokeStyle(lineWidth: 1.0, lineCap: .round, dash: [2.5, 2])
            )
    }

    // MARK: Tier 1 & 2 — amber circle with white check

    private var tier1_2: some View {
        ZStack {
            Circle().fill(Color.Echo.trustYellow)
            checkmark(color: .white, scale: 0.45)
        }
    }

    // MARK: Tier 3 — sky blue circle with white check

    private var tier3: some View {
        ZStack {
            Circle().fill(Color.Echo.primaryContainer)
            checkmark(color: .white, scale: 0.45)
        }
    }

    // MARK: Tier 4 — sky blue with soft halo ring

    private var tier4: some View {
        ZStack {
            Circle()
                .fill(Color.Echo.primaryContainer.opacity(0.25))
                .scaleEffect(1.28)
            Circle().fill(Color.Echo.primaryContainer)
            checkmark(color: .white, scale: 0.45)
        }
    }

    // MARK: Tier 5 — gradient ring, inner sky-blue fill

    private var tier5: some View {
        ZStack {
            Circle()
                .strokeBorder(
                    LinearGradient(
                        colors: [Color.Echo.trustPremium, Color.Echo.primaryContainer],
                        startPoint: .topLeading, endPoint: .bottomTrailing
                    ),
                    lineWidth: size * 0.11
                )
            Circle()
                .fill(Color.Echo.primaryContainer)
                .padding(size * 0.18)
            checkmark(color: .white, scale: 0.38)
        }
    }

    private func checkmark(color: Color, scale: CGFloat) -> some View {
        Image(systemName: "checkmark")
            .font(.system(size: size * scale, weight: .black))
            .foregroundStyle(color)
    }

    private var accessibilityText: String {
        switch tier {
        case 0: return "Provisioning identity"
        case 1: return "Standard trust, Tier 1"
        case 2: return "Phone verified, Tier 2"
        case 3: return "Active member, Tier 3"
        case 4: return "Credential verified, Tier 4"
        default: return "Identity verified, Tier 5"
        }
    }
}

// MARK: - Membership mark

/// A mark beside a name that says what kind of account this is. It is not the trust tier:
/// trust is earned by verification, a mark can be paid for, and each is drawn so one cannot be
/// taken for the other. Marks are seals, trust tiers are circles; no mark uses the trust blue.
enum MembershipMark: String, Equatable, CaseIterable {
    /// An active ECHO VIP subscription.
    case vip
    /// A business or organisation Echo has checked. Not issued yet.
    case organization
    /// A government or public body Echo has checked. Not issued yet.
    case official

    /// The server's value; anything unknown (including "") is no mark.
    init?(serverValue: String?) {
        guard let serverValue, let mark = MembershipMark(rawValue: serverValue) else { return nil }
        self = mark
    }

    var color: Color {
        switch self {
        case .vip: return Color(red: 0.49, green: 0.36, blue: 0.96)          // violet
        case .organization: return Color(red: 0.83, green: 0.62, blue: 0.09) // gold
        case .official: return Color(red: 0.51, green: 0.55, blue: 0.60)     // grey
        }
    }

    /// Says what the mark means and, for a paid one, what it does not mean.
    var accessibilityText: String {
        switch self {
        case .vip: return "VIP member. A paid subscription, not identity verification"
        case .organization: return "Verified organisation"
        case .official: return "Government or official account"
        }
    }
}

struct MembershipMarkView: View {
    let mark: MembershipMark
    let size: CGFloat

    var body: some View {
        Image(systemName: "checkmark.seal.fill")
            .font(.system(size: size))
            .foregroundStyle(mark.color)
            .accessibilityLabel(mark.accessibilityText)
            .accessibilityIdentifier("membership.mark.\(mark.rawValue)")
    }
}
