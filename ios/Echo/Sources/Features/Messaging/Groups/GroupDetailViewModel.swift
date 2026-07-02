#if os(iOS)
import Foundation

@MainActor
@Observable
final class GroupDetailViewModel {
    let groupId: String
    let groupName: String
    let currentUserDID: String

    var members: [GroupMemberWire] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    var statusMessage: String?

    private let groupsAPI: any GroupsAPIClient
    private let keyDistribution: GroupKeyDistributionService
    private let identityResolve: IdentityResolveClient

    init(
        groupId: String,
        groupName: String,
        currentUserDID: String,
        groupsAPI: any GroupsAPIClient,
        keyDistribution: GroupKeyDistributionService,
        identityResolve: IdentityResolveClient
    ) {
        self.groupId = groupId
        self.groupName = groupName
        self.currentUserDID = currentUserDID
        self.groupsAPI = groupsAPI
        self.keyDistribution = keyDistribution
        self.identityResolve = identityResolve
    }

    var canManageMembers: Bool {
        members.contains { $0.memberId == currentUserDID && $0.isAdmin }
    }

    func loadMembers() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            members = try await groupsAPI.listMembers(groupId: groupId)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func addMember(peerDID: String) async {
        guard canManageMembers else { return }
        guard peerDID != currentUserDID else { return }
        guard !members.contains(where: { $0.memberId == peerDID }) else { return }

        isSaving = true
        errorMessage = nil
        statusMessage = nil
        defer { isSaving = false }

        do {
            let requiresRekey = try await groupsAPI.addMember(groupId: groupId, memberDid: peerDID)
            await loadMembers()
            if requiresRekey {
                try await performRekey()
                statusMessage = "Member added and group key rotated."
            } else {
                statusMessage = "Member added."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func removeMember(_ peerDID: String) async {
        guard canManageMembers else { return }
        guard peerDID != currentUserDID else {
            errorMessage = "Transfer ownership before leaving as admin."
            return
        }

        isSaving = true
        errorMessage = nil
        statusMessage = nil
        defer { isSaving = false }

        do {
            let requiresRekey = try await groupsAPI.removeMember(groupId: groupId, memberDid: peerDID)
            await loadMembers()
            if requiresRekey {
                try await performRekey()
                statusMessage = "Member removed and group key rotated."
            } else {
                statusMessage = "Member removed."
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func banMember(_ peerDID: String) async {
        guard canManageMembers, peerDID != currentUserDID else { return }
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }
        do {
            try await groupsAPI.banMember(groupId: groupId, memberDid: peerDID)
            await loadMembers()
            statusMessage = "Member banned."
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func muteMember(_ peerDID: String, hours: Int) async {
        guard canManageMembers, peerDID != currentUserDID else { return }
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }
        do {
            try await groupsAPI.muteMember(groupId: groupId, memberDid: peerDID, durationHours: hours)
            await loadMembers()
            statusMessage = "Member muted."
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func performRekey() async throws {
        let dids = members.map(\.memberId)
        _ = try await keyDistribution.rekeyForMembers(
            groupId: groupId,
            memberDIDs: dids,
            distributedBy: currentUserDID,
            identityResolve: identityResolve
        )
    }
}
#endif
