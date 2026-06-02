import SwiftUI

/// A trust tier on the T0–T4 ladder (ux-spec §4.1 — one hue, five opacities).
public struct TrustLadderTier: Identifiable, Equatable, Sendable {
    public let id: Int          // 0...4
    public let name: String
    public let meaning: String
    public let opacity: Double

    public init(id: Int, name: String, meaning: String, opacity: Double) {
        self.id = id
        self.name = name
        self.meaning = meaning
        self.opacity = opacity
    }

    public static let ladder: [TrustLadderTier] = [
        TrustLadderTier(id: 0, name: "Unverified", meaning: "Not verified", opacity: 0.10),
        TrustLadderTier(id: 1, name: "Basic",      meaning: "Profile visible", opacity: 0.25),
        TrustLadderTier(id: 2, name: "Verified",   meaning: "Identity confirmed", opacity: 0.45),
        TrustLadderTier(id: 3, name: "Trusted",    meaning: "Community trusted", opacity: 0.70),
        TrustLadderTier(id: 4, name: "Elite",      meaning: "Full trust", opacity: 1.0),
    ]
}

/// A verification source contributing to trust (mock now; bind to a trust API in P2).
public struct VerificationSource: Identifiable, Equatable, Sendable {
    public let id: String
    public let title: String
    public let icon: String
    public let satisfied: Bool

    public init(id: String, title: String, icon: String, satisfied: Bool) {
        self.id = id
        self.title = title
        self.icon = icon
        self.satisfied = satisfied
    }
}

/// Trust dashboard (spec §4.9 — P2 deep link from Profile, not a tab). Functional with
/// mock data; "Increase trust" routes to the VIP verification path (kept for later phases).
public struct TrustDashboardView: View {
    let currentTier: Int
    let sources: [VerificationSource]
    let onStartVerification: () -> Void

    public init(
        currentTier: Int = 2,
        sources: [VerificationSource] = TrustDashboardView.mockSources,
        onStartVerification: @escaping () -> Void = {}
    ) {
        self.currentTier = currentTier
        self.sources = sources
        self.onStartVerification = onStartVerification
    }

    private var tier: TrustLadderTier {
        TrustLadderTier.ladder.first { $0.id == currentTier } ?? TrustLadderTier.ladder[0]
    }

    public var body: some View {
        ScrollView {
            VStack(spacing: 0) {
                tierCard
                ladderSection
                sourcesSection
                if currentTier < 4 {
                    Button(action: onStartVerification) {
                        Text("Increase trust")
                            .font(.system(size: 15, weight: .semibold))
                            .foregroundColor(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 14)
                            .background(Color.echoSignal)
                            .clipShape(RoundedRectangle(cornerRadius: 22))
                    }
                    .buttonStyle(.plain)
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.top, Spacing.sm.rawValue)
                    Text("Verify with a digital ID or standard IDV to raise your tier.")
                        .font(.system(size: 12))
                        .foregroundColor(.echoInk40)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, Spacing.xl.rawValue)
                        .padding(.top, 6)
                }
            }
            .padding(.bottom, Spacing.xl.rawValue)
        }
        .background(Color.echoPaper.ignoresSafeArea())
        .navigationTitle("Trust")
    }

    private var tierCard: some View {
        VStack(spacing: 10) {
            ZStack {
                Circle()
                    .fill(Color.echoTrustGreen.opacity(tier.opacity))
                    .frame(width: 84, height: 84)
                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 32))
                    .foregroundColor(tier.opacity > 0.4 ? .white : .echoTrustGreen)
            }
            Text(tier.name)
                .font(.system(size: 22, weight: .semibold))
                .foregroundColor(.echoInk)
            Text("Tier T\(tier.id) · \(tier.meaning)")
                .font(.system(size: 13))
                .foregroundColor(.echoInk55)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 24)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(RoundedRectangle(cornerRadius: 16).stroke(Color.echoHair, lineWidth: 1))
        .padding(Spacing.lg.rawValue)
    }

    private var ladderSection: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("TRUST LADDER")
                .font(.system(size: 10.5, weight: .semibold, design: .monospaced))
                .tracking(1)
                .foregroundColor(.echoInk40)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.bottom, 6)
            ForEach(TrustLadderTier.ladder.reversed()) { t in
                HStack(spacing: Spacing.md.rawValue) {
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color.echoTrustGreen.opacity(t.opacity))
                        .frame(width: 28, height: 28)
                    Text(t.name)
                        .font(.system(size: 14, weight: t.id == currentTier ? .bold : .regular))
                        .foregroundColor(t.id == currentTier ? .echoInk : .echoInk70)
                    Spacer()
                    if t.id == currentTier {
                        Text("You're here")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(.echoTrustGreen)
                    } else {
                        Text("T\(t.id)")
                            .font(.system(size: 12))
                            .foregroundColor(.echoInk40)
                    }
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, 9)
                .overlay(Divider(), alignment: .bottom)
            }
        }
        .padding(.top, Spacing.sm.rawValue)
    }

    private var sourcesSection: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("VERIFICATION SOURCES")
                .font(.system(size: 10.5, weight: .semibold, design: .monospaced))
                .tracking(1)
                .foregroundColor(.echoInk40)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.top, Spacing.lg.rawValue)
                .padding(.bottom, 6)
            ForEach(sources) { source in
                HStack(spacing: Spacing.md.rawValue) {
                    Image(systemName: source.icon)
                        .frame(width: 22)
                        .foregroundColor(.echoInk55)
                    Text(source.title)
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk70)
                    Spacer()
                    Image(systemName: source.satisfied ? "checkmark.circle.fill" : "circle")
                        .foregroundColor(source.satisfied ? .echoTrustGreen : .echoInk40)
                }
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, 11)
                .overlay(Divider(), alignment: .bottom)
            }
        }
    }

    public static let mockSources: [VerificationSource] = [
        VerificationSource(id: "did", title: "Secure Enclave key (DID)", icon: "key.fill", satisfied: true),
        VerificationSource(id: "phone", title: "SMS backup", icon: "phone.fill", satisfied: true),
        VerificationSource(id: "idv", title: "Standard IDV (scan + selfie)", icon: "person.text.rectangle", satisfied: false),
        VerificationSource(id: "wallet", title: "Digital ID / mDL", icon: "wallet.pass.fill", satisfied: false),
    ]
}

#if DEBUG
struct TrustDashboardView_Previews: PreviewProvider {
    static var previews: some View {
        NavigationStack { TrustDashboardView(currentTier: 2) }
    }
}
#endif
