#if os(iOS)
import SwiftUI

/// Broadcast channels list (WO-77).
struct ChannelsListView: View {
    @State private var channels: [BroadcastChannelWire] = []
    @State private var isLoading = false

    var body: some View {
        Group {
            if isLoading && channels.isEmpty {
                ProgressView().frame(maxWidth: .infinity, minHeight: 120)
            } else if channels.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "dot.radiowaves.left.and.right")
                        .font(.system(size: 28))
                        .foregroundColor(.echoInk55)
                    Text("No channels yet")
                        .font(.system(size: 14))
                        .foregroundColor(.echoInk70)
                }
                .frame(maxWidth: .infinity, minHeight: 120)
            } else {
                ForEach(channels) { channel in
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(channel.name)
                                .font(.system(size: 15, weight: .medium))
                            if let topic = channel.topic, !topic.isEmpty {
                                Text(topic)
                                    .font(.system(size: 12))
                                    .foregroundColor(.echoInk55)
                            }
                        }
                        Spacer()
                        Button("Join") {
                            Task {
                                _ = await ChannelsAPIClient.subscribe(channelId: channel.id)
                            }
                        }
                        .font(.system(size: 13, weight: .semibold))
                    }
                    .padding(.horizontal, Spacing.lg.rawValue)
                    .padding(.vertical, Spacing.md.rawValue)
                    Divider().padding(.leading, Spacing.lg.rawValue)
                }
            }
        }
        .task { await reload() }
        .refreshable { await reload() }
    }

    private func reload() async {
        isLoading = true
        channels = await ChannelsAPIClient.listChannels()
        isLoading = false
    }
}
#endif
