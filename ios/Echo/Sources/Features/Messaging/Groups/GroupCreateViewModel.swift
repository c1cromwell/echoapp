#if os(iOS)
import Foundation

@MainActor
@Observable
final class GroupCreateViewModel {
    var name = ""
    var selectedPeerDIDs: Set<String> = []
    var isSaving = false
    var errorMessage: String?

    private let groupsAPI: any GroupsAPIClient
    private let keyDistribution: GroupKeyDistributionService
    private let identityResolve: IdentityResolveClient
    private let currentDID: String

    init(
        groupsAPI: any GroupsAPIClient,
        keyDistribution: GroupKeyDistributionService,
        identityResolve: IdentityResolveClient,
        currentDID: String
    ) {
        self.groupsAPI = groupsAPI
        self.keyDistribution = keyDistribution
        self.identityResolve = identityResolve
        self.currentDID = currentDID
    }

    func toggleMember(_ peerDID: String) {
        if selectedPeerDIDs.contains(peerDID) {
            selectedPeerDIDs.remove(peerDID)
        } else {
            selectedPeerDIDs.insert(peerDID)
        }
    }

    /// Creates a private group and distributes v1 key to admin + selected members.
    func createGroup() async -> (groupId: String, name: String)? {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            errorMessage = "Enter a group name."
            return nil
        }
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        let groupId = "grp-\(UUID().uuidString.lowercased())"
        do {
            try await groupsAPI.createGroup(groupId: groupId, name: trimmed, type: "private")
            for peer in selectedPeerDIDs {
                _ = try await groupsAPI.addMember(groupId: groupId, memberDid: peer)
            }
            let targets = try await GroupMemberResolver.memberTargets(
                peerDIDs: Array(selectedPeerDIDs),
                currentDID: currentDID,
                identityResolve: identityResolve
            )
            guard !targets.isEmpty else {
                errorMessage = "Add at least one member."
                return nil
            }
            _ = try await keyDistribution.rotateAndDistribute(
                groupId: groupId,
                members: targets,
                distributedBy: currentDID
            )
            return (groupId, trimmed)
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }
}
#endif
