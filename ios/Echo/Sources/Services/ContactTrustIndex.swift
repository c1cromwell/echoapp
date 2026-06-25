#if os(iOS)
import Foundation
import Observation

/// Peer DID → trust tier (T0–T4) for hub folder filters (Phase B).
@MainActor
@Observable
public final class ContactTrustIndex {
    public static let shared = ContactTrustIndex()

    private var tierByDID: [String: Int] = [:]
    private let storageKey = "echo.contact.trust.tiers.v1"

    private init() {
        load()
    }

    public func tier(peerDID: String) -> Int {
        tierByDID[peerDID] ?? 1
    }

    public func tier(conversationId: String, peerDID: String) -> Int {
        tierByDID[peerDID] ?? tierByDID[conversationId] ?? 1
    }

    public func setTier(_ tier: Int, peerDID: String) {
        tierByDID[peerDID] = min(4, max(0, tier))
        persist()
    }

    func ingestRemoteContacts(_ contacts: [RemoteContact]) {
        for contact in contacts {
            guard let did = contact.contactDid, !did.isEmpty else { continue }
            tierByDID[did] = Self.tier(fromTrustBadge: contact.trustBadge)
        }
        persist()
    }

    func ingestSearchHits(_ hits: [UsernameSearchHit]) {
        for hit in hits {
            if let tier = hit.tier {
                tierByDID[hit.did] = min(4, max(0, tier))
            }
        }
        persist()
    }

    public static func tier(fromTrustBadge badge: String?) -> Int {
        switch badge?.lowercased() ?? "" {
        case "highlytrusted", "highly_trusted", "tier4", "t4": return 4
        case "trusted", "tier3", "t3": return 3
        case "verified", "tier2", "t2": return 2
        case "basic", "newcomer", "tier1", "t1": return 1
        default: return 1
        }
    }

    public static func trustLevelLabel(tier: Int) -> String {
        switch min(4, max(0, tier)) {
        case 4: return "HighlyTrusted"
        case 3: return "Trusted"
        case 2: return "Verified"
        default: return "Basic"
        }
    }

    private func persist() {
        UserDefaults.standard.set(tierByDID, forKey: storageKey)
    }

    private func load() {
        guard let raw = UserDefaults.standard.dictionary(forKey: storageKey) else { return }
        tierByDID = raw.compactMapValues { value in
            if let n = value as? Int { return n }
            if let n = value as? NSNumber { return n.intValue }
            return nil
        }
    }
}
#endif
