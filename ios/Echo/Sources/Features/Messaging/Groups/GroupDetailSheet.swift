#if os(iOS)
import SwiftUI

struct GroupDetailSheet: View {
    let groupId: String
    let groupName: String
    let currentUserDID: String

    @State private var viewModel: GroupDetailViewModel?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Group {
                if let viewModel {
                    GroupDetailView(viewModel: viewModel)
                } else {
                    ProgressView("Loading group…")
                        .task { await loadViewModel() }
                }
            }
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
            }
        }
    }

    private func loadViewModel() async {
        let api = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
        let groups = DIContainer.shared.resolveGroupsAPI() ?? LiveGroupsAPIClient(apiClient: api)
        let distribution = DIContainer.shared.resolveGroupKeyDistribution()
            ?? GroupKeyDistributionService(
                keyManager: GroupKeyManager(encryption: KinnamiEncryption()),
                groupsAPI: groups,
                encryption: KinnamiEncryption(),
                secureEnclave: SecureEnclaveManager.shared
            )
        viewModel = GroupDetailViewModel(
            groupId: groupId,
            groupName: groupName,
            currentUserDID: currentUserDID,
            groupsAPI: groups,
            keyDistribution: distribution,
            identityResolve: IdentityResolveClient(apiClient: api)
        )
    }
}
#endif
