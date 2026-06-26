// Features/Safety/ContactSafetyBanner.swift
//
// Non-blocking, dismissible safety banner shown at the top of a chat thread on a risky first
// contact. Mirrors ThreadSummaryBanner's style. Copy is friendly, never accusatory.

#if os(iOS)
import SwiftUI

public struct ContactSafetyBanner: View {
    let assessment: SafetyAssessment
    let onVerify: () -> Void
    let onBlock: () -> Void
    let onDismiss: () -> Void

    public init(assessment: SafetyAssessment,
                onVerify: @escaping () -> Void = {},
                onBlock: @escaping () -> Void = {},
                onDismiss: @escaping () -> Void = {}) {
        self.assessment = assessment
        self.onVerify = onVerify
        self.onBlock = onBlock
        self.onDismiss = onDismiss
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: icon).font(.system(size: 13, weight: .semibold))
                Text(title).font(.system(size: 13, weight: .semibold))
                Spacer()
                Button(action: onDismiss) {
                    Image(systemName: "xmark").font(.system(size: 11, weight: .semibold))
                }
                .foregroundColor(tint.opacity(0.7))
            }
            ForEach(Array(messages.enumerated()), id: \.offset) { _, line in
                Text("• \(line)").font(.system(size: 12)).foregroundColor(.echoInk70)
            }
            HStack(spacing: 10) {
                Button(action: onVerify) { actionLabel("View profile", filled: true) }
                Button(action: onBlock) { actionLabel("Block", filled: false) }
            }
            .padding(.top, 2)
        }
        .padding(12)
        .background(tint.opacity(0.10))
        .overlay(RoundedRectangle(cornerRadius: 12).stroke(tint.opacity(0.30), lineWidth: 1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .foregroundColor(tint)
    }

    @ViewBuilder private func actionLabel(_ text: String, filled: Bool) -> some View {
        Text(text)
            .font(.system(size: 12, weight: .semibold))
            .foregroundColor(filled ? .white : tint)
            .padding(.horizontal, 12).padding(.vertical, 6)
            .background(filled ? tint : Color.clear)
            .overlay(RoundedRectangle(cornerRadius: 8).stroke(tint.opacity(filled ? 0 : 0.5), lineWidth: 1))
            .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private var tint: Color { assessment.level == .warning ? .echoAlert : .echoWarning }
    private var icon: String { assessment.level == .warning ? "exclamationmark.triangle.fill" : "shield.lefthalf.filled" }
    private var title: String { assessment.level == .warning ? "Possible impersonation" : "New, unverified contact" }

    private var messages: [String] {
        assessment.reasons.map {
            switch $0 {
            case .firstContact: return "This is the first time this person has messaged you."
            case .unverifiedSender: return "They haven't verified their identity."
            case .lowTrustSender(let tier): return "Low trust level (Tier \(tier))."
            case .possibleImpersonation(let name): return "Their name looks like \"\(name)\", a verified contact — but it's a different account."
            }
        }
    }
}
#endif
