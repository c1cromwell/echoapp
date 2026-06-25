#if os(iOS)
import Foundation

/// Fans out group call offers to all members (M4 group-call foundation).
@MainActor
enum GroupCallSignalingService {
    static func fanOutOffer(
        groupId: String,
        callId: String,
        callType: CallType,
        sdp: String,
        excludingDID: String
    ) async throws {
        guard let groups = DIContainer.shared.resolveGroupsAPI(),
              let signaling = DIContainer.shared.resolveCallSignaling() else { return }
        let members = try await groups.listMembers(groupId: groupId)
        for member in members where member.memberId != excludingDID {
            try await signaling.sendOffer(
                callId: callId,
                to: member.memberId,
                callType: callType,
                sdp: sdp
            )
        }
    }
}
#endif
