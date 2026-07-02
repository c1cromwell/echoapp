#if os(iOS)
import Foundation

/// Applies trust-tier TTL policy locally when offline (WO-115).
enum DisappearingRestrictionsPolicy {
    static func minTTLSeconds(trustTier: Int) -> Int {
        switch trustTier {
        case ...1: return 3600
        case 2: return 300
        default: return 10
        }
    }

    static func isAllowed(ttlSeconds: Int, trustTier: Int) -> Bool {
        if ttlSeconds <= 0 { return true }
        return ttlSeconds >= minTTLSeconds(trustTier: trustTier)
    }

    static func reason(trustTier: Int) -> String {
        switch trustTier {
        case ...1:
            return "Unverified accounts cannot use timers shorter than 1 hour."
        case 2:
            return "Newcomer accounts cannot use timers shorter than 5 minutes."
        default:
            return "No disappearing-message restrictions for your trust tier."
        }
    }
}
#endif
