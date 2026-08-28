import Foundation

/// Human-readable label for channel members, join requests, and comment authors.
enum ChannelParticipantLabel {
    static func resolve(
        did: String,
        displayName: String?,
        contactName: String? = nil,
        currentDID: String? = nil
    ) -> String {
        if let currentDID, !currentDID.isEmpty, did == currentDID {
            return "You"
        }
        if let displayName {
            let trimmed = displayName.trimmingCharacters(in: .whitespacesAndNewlines)
            if !trimmed.isEmpty { return trimmed }
        }
        if let contactName {
            let trimmed = contactName.trimmingCharacters(in: .whitespacesAndNewlines)
            if !trimmed.isEmpty { return trimmed }
        }
        return truncatedDID(did)
    }

    static func truncatedDID(_ did: String) -> String {
        guard did.count > 20 else { return did }
        return String(did.prefix(12)) + "…" + String(did.suffix(6))
    }
}
