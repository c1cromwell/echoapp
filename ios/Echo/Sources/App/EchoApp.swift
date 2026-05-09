#if os(iOS)
import SwiftUI

@main
struct EchoApp: App {
    @State private var appState: AppState
    @Environment(\.scenePhase) private var scenePhase

    init() {
        let provisionService = SilentProvisionService(
            secureEnclave: StubProvisionSecureEnclave(),
            api: StubProvisionAPI(),
            stargazer: StubProvisionStargazer(),
            passkeyProvider: StubProvisionPasskey()
        )
        _appState = State(initialValue: AppState(provisionService: provisionService))
    }

    var body: some Scene {
        WindowGroup {
            EchoRootView(appState: appState)
        }
        // WO-208 / WO-223: purge all Secure Enclave derived-key caches on background.
        // This ensures T1 secrets never persist in memory while the app is suspended.
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .background {
                Task {
                    await SecureEnclaveManager.shared.purgeOnBackground()
                }
            }
        }
    }
}
#endif
