#if os(iOS)
import SwiftUI

@main
struct EchoApp: App {
    @State private var appState: AppState
    @Environment(\.scenePhase) private var scenePhase

    init() {
        let provisionService = SilentProvisionService(
            secureEnclave: RealProvisionSecureEnclave(),
            api: RealProvisionAPI(),
            stargazer: StubProvisionStargazer(),     // Phase 2: real Stargazer SDK
            passkeyProvider: StubProvisionPasskey()  // Phase 2: real WebAuthn
        )
        _appState = State(initialValue: AppState(provisionService: provisionService))
    }

    var body: some Scene {
        WindowGroup {
            EchoRootView(appState: appState)
        }
        .onChange(of: scenePhase) { _, newPhase in
            switch newPhase {
            case .active:
                // WO-224: Re-derive storage key when app returns to foreground.
                // deriveStorageKey is nonisolated + synchronous — no biometric prompt
                // (it uses the key label as IKM proxy). Full biometric re-derivation
                // happens inside SecureEnclaveManager for the identity key path.
                Task {
                    let storageKey = SecureEnclaveManager.shared.deriveStorageKey(
                        keyId: "echo-identity-signing"
                    )
                    await LocalDatabase.shared.unlock(storageKey: storageKey)
                }

            case .background:
                // WO-208 / WO-223 / WO-224: zero all in-memory key material.
                Task {
                    await SecureEnclaveManager.shared.purgeOnBackground()
                    await LocalDatabase.shared.lockStorage()
                }

            default:
                break
            }
        }
    }
}
#endif
