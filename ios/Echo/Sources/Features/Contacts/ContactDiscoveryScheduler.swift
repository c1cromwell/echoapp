#if os(iOS)
import Foundation

/// Runs OPRF-PSI scans on the cadence chosen in Privacy settings (WO-221 / S5).
enum ContactDiscoveryScheduler {
    @MainActor
    static func runIfDue() {
        guard ContactDiscoverySyncPreferences.shouldRunAutomaticSync() else { return }
        Task {
            guard let useCase = DIContainer.shared.resolveContactDiscoveryUseCase() else { return }
            do {
                _ = try await useCase.discoverContacts()
                ContactDiscoverySyncPreferences.markSyncedNow()
            } catch ContactDiscoveryError.noMatches {
                ContactDiscoverySyncPreferences.markSyncedNow()
            } catch {
                // Permission denied or network — try again on next foreground.
            }
        }
    }
}
#endif
