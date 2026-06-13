#if os(iOS)
import SwiftUI

/// Loads DI + current DID before presenting `GroupCreateView`.
struct GroupCreateSheet: View {
    var onCreated: (String, String) -> Void

    @State private var viewModel: GroupCreateViewModel?

    var body: some View {
        Group {
            if let viewModel {
                GroupCreateView(viewModel: viewModel, onCreated: onCreated)
            } else {
                ProgressView("Preparing group setup…")
                    .task { await loadViewModel() }
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
        let currentDID = await CurrentUserSession.currentDID() ?? ""
        viewModel = GroupCreateViewModel(
            groupsAPI: groups,
            keyDistribution: distribution,
            identityResolve: IdentityResolveClient(apiClient: api),
            currentDID: currentDID
        )
    }
}
#endif
