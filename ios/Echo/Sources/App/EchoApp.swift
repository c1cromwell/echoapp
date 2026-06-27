#if os(iOS)
import SwiftUI

@main
struct EchoApp: App {
    #if ECHO_PRODUCT_MESSAGING
    @State private var appState: AppState
    #endif
    @Environment(\.scenePhase) private var scenePhase

    init() {
        #if ECHO_PRODUCT_MESSAGING
        let provisionService = SilentProvisionService(
            secureEnclave: RealProvisionSecureEnclave(),
            api: RealProvisionAPI(),
            stargazer: RealProvisionStargazer(),
            passkeyProvider: StubProvisionPasskey()
        )
        _appState = State(initialValue: AppState(provisionService: provisionService))
        #endif
    }

    var body: some Scene {
        WindowGroup {
            #if ECHO_PRODUCT_MESSAGING
            EchoRootView(appState: appState)
            #elseif ECHO_PRODUCT_COMPLY
            ComplyCompanionRootView()
            #else
            PassportRootView()
            #endif
        }
        .onChange(of: scenePhase) { _, newPhase in
            #if ECHO_PRODUCT_MESSAGING
            switch newPhase {
            case .active:
                Task {
                    let storageKey = SecureEnclaveManager.shared.deriveStorageKey(
                        keyId: "echo-identity-signing"
                    )
                    await LocalDatabase.shared.unlock(storageKey: storageKey)
                }

            case .background:
                Task {
                    await SecureEnclaveManager.shared.purgeOnBackground()
                    await LocalDatabase.shared.lockStorage()
                }

            default:
                break
            }
            #endif
        }
    }
}
#endif
