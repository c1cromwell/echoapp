import SwiftUI

@main
struct EchoApp: App {
    @State private var isAuthenticated = false
    @State private var showEnrollment = false

    var body: some Scene {
        WindowGroup {
            if isAuthenticated {
                MainTabView()
            } else {
                NavigationStack {
                    GlacialLoginScreen(
                        onPasskeyLogin: { isAuthenticated = true },
                        onSMSLogin: { _ in isAuthenticated = true },
                        onGetStarted: { showEnrollment = true }
                    )
                }
                .fullScreenCover(isPresented: $showEnrollment) {
                    EnrollmentCoordinatorView(
                        coordinator: EnrollmentCoordinator(
                            onComplete: { _ in
                                showEnrollment = false
                                isAuthenticated = true
                            },
                            onCancel: { showEnrollment = false }
                        )
                    )
                }
            }
        }
    }
}
