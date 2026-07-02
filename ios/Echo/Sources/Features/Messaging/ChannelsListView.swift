#if os(iOS)
import SwiftUI

/// Channels tab content in the Messages hub (WO-65).
/// Placeholder — full channel implementation is planned for a future phase.
struct ChannelsListView: View {
    var body: some View {
        EmptyStateView(
            icon: "megaphone",
            title: "Channels coming soon",
            subtitle: "Broadcast channels will let you share updates with large groups."
        )
    }
}
#endif
