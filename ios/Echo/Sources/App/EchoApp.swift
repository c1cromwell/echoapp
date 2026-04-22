import SwiftUI

@main
struct EchoApp: App {
    @State private var appState: AppState

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
    }
}
