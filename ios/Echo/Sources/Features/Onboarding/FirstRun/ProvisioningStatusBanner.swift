#if os(iOS)
import SwiftUI

/// Non-blocking banner that surfaces a failed background provision (the enrollment tail
/// run by `SilentProvisionService` after first-run: Secure Enclave key → DID → wallet →
/// passkey).
///
/// Previously a failed provision was invisible: `firstRunCompleted` starts provisioning
/// and routes straight to `.authenticated`, so a failure left the user inside the app
/// with no identity and no way to recover. This shows only in the actionable `.failed`
/// state and wires the already-existing `SilentProvisionService.retry(displayName:)`.
struct ProvisioningStatusBanner: View {
    @Environment(AppState.self) private var appState

    var body: some View {
        Group {
            if isFailed {
                banner
            }
        }
        .animation(.easeInOut(duration: 0.2), value: isFailed)
    }

    private var isFailed: Bool {
        if case .failed = appState.provisionService.stage { return true }
        return false
    }

    private var banner: some View {
        HStack(spacing: 10) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 14))
                .foregroundStyle(Color.echoAlert)

            VStack(alignment: .leading, spacing: 1) {
                Text("Account setup didn't finish")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(Color.echoInk)
                Text("Some features stay locked until it completes.")
                    .font(.system(size: 11))
                    .foregroundStyle(Color.echoInk55)
            }

            Spacer()

            Button("Retry") {
                appState.provisionService.retry(displayName: appState.displayName)
            }
            .font(.system(size: 13, weight: .semibold))
            .foregroundStyle(Color.echoSignal)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
        .overlay(Rectangle().fill(Color.echoHair).frame(height: 1), alignment: .bottom)
        .transition(.move(edge: .top).combined(with: .opacity))
    }
}
#endif
