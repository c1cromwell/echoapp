// Features/Safety/ContactSafetyEvaluator.swift
//
// Trust-aware scam/impersonation risk scoring using ECHO's IDV trust tiers — the protection no
// other messenger can offer. Pure logic; the chat surfaces the result as a dismissible banner and
// payments use the pay-guard. Reuses ContactTrustIndex tiers (0 = unverified … higher = verified).

import Foundation

public enum SafetyLevel: Int, Comparable, Sendable {
    case ok, caution, warning
    public static func < (lhs: SafetyLevel, rhs: SafetyLevel) -> Bool { lhs.rawValue < rhs.rawValue }
}

public enum SafetyReason: Equatable, Sendable {
    case firstContact
    case unverifiedSender
    case lowTrustSender(tier: Int)
    case possibleImpersonation(of: String)   // display name of the trusted contact being mimicked
}

public struct SafetyAssessment: Equatable, Sendable {
    public let level: SafetyLevel
    public let reasons: [SafetyReason]
}

public struct KnownContact: Equatable, Sendable {
    public let did: String
    public let displayName: String
    public let tier: Int
    public init(did: String, displayName: String, tier: Int) {
        self.did = did
        self.displayName = displayName
        self.tier = tier
    }
}

public struct ContactSafetyEvaluator {
    /// Tier at/above which a known contact is "trusted" (the baseline impersonation targets).
    private let trustedTierThreshold: Int

    public init(trustedTierThreshold: Int = 3) {
        self.trustedTierThreshold = trustedTierThreshold
    }

    public func evaluate(
        peerDID: String,
        peerName: String,
        peerTier: Int,
        isFirstContact: Bool,
        knownContacts: [KnownContact]
    ) -> SafetyAssessment {
        var reasons: [SafetyReason] = []
        var level: SafetyLevel = .ok

        // Impersonation is the strongest signal: a low-trust peer whose name mimics a *trusted*
        // contact but with a different DID.
        if peerTier <= 1 {
            for known in knownContacts
            where known.tier >= trustedTierThreshold
                && known.did != peerDID
                && LookalikeDetector.isLookalike(peerName, of: known.displayName) {
                reasons.append(.possibleImpersonation(of: known.displayName))
                level = Swift.max(level, .warning)
                break
            }
        }

        if isFirstContact {
            reasons.append(.firstContact)
            level = Swift.max(level, .caution)
        }

        if peerTier == 0 {
            reasons.append(.unverifiedSender)
            level = Swift.max(level, .caution)
        } else if peerTier == 1 {
            reasons.append(.lowTrustSender(tier: peerTier))
            level = Swift.max(level, .caution)
        }

        return SafetyAssessment(level: level, reasons: reasons)
    }

    /// Pay-guard: require an extra confirmation before sending money to an unverified/low-trust peer.
    public func requiresExtraConfirmationToPay(peerTier: Int) -> Bool { peerTier <= 1 }
}
