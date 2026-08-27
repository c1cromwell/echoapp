#if os(iOS)
import SwiftUI
import UserNotifications

#if ECHO_PRODUCT_MESSAGING
final class EchoAppDelegate: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        return true
    }

    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        Task { await PushRegistrationService.shared.handleDeviceToken(deviceToken) }
    }

    func application(
        _ application: UIApplication,
        didReceiveRemoteNotification userInfo: [AnyHashable: Any],
        fetchCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void
    ) {
        Task {
            await MessageRelaySession.connect()
            completionHandler(.newData)
        }
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        Task {
            await MessageRelaySession.connect()
            completionHandler()
        }
    }
}
#endif

@main
struct EchoApp: App {
    #if ECHO_PRODUCT_MESSAGING
    @UIApplicationDelegateAdaptor(EchoAppDelegate.self) private var appDelegate
    @State private var appState: AppState
    #endif
    @Environment(\.scenePhase) private var scenePhase

    init() {
        UITestSupport.applyIfNeeded()
        #if ECHO_PRODUCT_MESSAGING
        DisappearingMessageBGTask.register()
        ScheduledMessageBGTask.register()
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
                DisappearingMessageBGTask.scheduleNext()
                ScheduledMessageBGTask.scheduleNext()
                HiddenChatsSession.shared.refreshForegroundLockIfNeeded()
                BackupScheduler.runIfDue()
                Task { await MessageRelaySession.connect() }
                Task { await ScheduledMessageStore.hydrateFromServer() }
                Task {
                    let storageKey = SecureEnclaveManager.shared.deriveStorageKey(
                        keyId: "echo-identity-signing"
                    )
                    await LocalDatabase.shared.unlock(storageKey: storageKey)
                }
                Task { await DebugSeedData.seedIfNeeded() }

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
