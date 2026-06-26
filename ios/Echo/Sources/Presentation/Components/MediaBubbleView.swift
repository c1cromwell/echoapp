#if os(iOS)
import SwiftUI

struct MediaBubbleView: View {
    let mediaRef: MediaAttachmentRef
    let peerDID: String
    let isSent: Bool
    let timestamp: String
    let deliveryStatus: DeliveryStatus?
    var onTapImage: ((UIImage) -> Void)?

    @State private var imageData: Data?
    @State private var audioData: Data?
    @State private var isLoading = true
    @State private var loadError: String?
    @State private var waveformSamples: [CGFloat] = []

    private var mediaKind: MediaKind {
        MediaKind(rawValue: mediaRef.mediaKind) ?? .file
    }

    var body: some View {
        VStack(alignment: isSent ? .trailing : .leading, spacing: Spacing.xs.rawValue) {
            Group {
                switch mediaKind {
                case .image:
                    imageContent
                case .audio:
                    audioContent
                case .video:
                    videoPlaceholder
                case .file:
                    filePlaceholder
                }
            }

            if let caption = mediaRef.caption, !caption.isEmpty {
                Text(caption)
                    .typographyStyle(.body, color: isSent ? .white : .echoPrimaryText)
                    .padding(.horizontal, Spacing.md.rawValue)
            }

            HStack(spacing: Spacing.xs.rawValue) {
                if isSent { Spacer() }
                Text(timestamp)
                    .typographyStyle(.caption, color: .echoInk40)
                if isSent, let deliveryStatus {
                    SmartCheckmarkView(status: deliveryStatus, eventId: nil, onTapVerified: nil)
                }
                if !isSent { Spacer() }
            }
            .frame(maxWidth: 280)
        }
        .frame(maxWidth: .infinity, alignment: isSent ? .trailing : .leading)
        .padding(.horizontal, Spacing.md.rawValue)
        .task { await loadMedia() }
    }

    @ViewBuilder
    private var imageContent: some View {
        if let cached = MediaThumbnailCache.load(fileId: mediaRef.fileId),
           imageData == nil {
            Image(uiImage: cached)
                .resizable()
                .aspectRatio(contentMode: .fill)
                .frame(maxWidth: 260, maxHeight: 300)
                .clipShape(RoundedRectangle(cornerRadius: 16))
        } else if let imageData, let uiImage = UIImage(data: imageData) {
            Image(uiImage: uiImage)
                .resizable()
                .aspectRatio(contentMode: .fill)
                .frame(maxWidth: 260, maxHeight: 300)
                .clipShape(RoundedRectangle(cornerRadius: 16))
                .onTapGesture { onTapImage?(uiImage) }
        } else if isLoading {
            mediaSkeleton(width: 200, height: 150)
        } else {
            mediaErrorView
        }
    }

    @ViewBuilder
    private var audioContent: some View {
        HStack {
            AudioPlayerView(
                waveformSamples: waveformSamples.isEmpty
                    ? defaultWaveform
                    : waveformSamples,
                audioData: audioData
            )
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(isSent ? Color.echoSignal.opacity(0.9) : Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .frame(maxWidth: 280)
    }

    @ViewBuilder
    private var videoPlaceholder: some View {
        ZStack {
            if let imageData, let uiImage = UIImage(data: imageData) {
                Image(uiImage: uiImage)
                    .resizable()
                    .aspectRatio(contentMode: .fill)
                    .frame(maxWidth: 260, maxHeight: 200)
            } else {
                Rectangle()
                    .fill(Color.echoPaperDim)
                    .frame(width: 200, height: 150)
            }
            Image(systemName: "play.circle.fill")
                .font(.system(size: 44))
                .foregroundStyle(.white.opacity(0.9))
        }
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    @ViewBuilder
    private var filePlaceholder: some View {
        HStack(spacing: 10) {
            Image(systemName: "doc.fill")
                .font(.system(size: 28))
                .foregroundStyle(isSent ? .white : .echoSignal)
            VStack(alignment: .leading, spacing: 2) {
                Text("Attachment")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(isSent ? .white : .echoInk)
                Text(ByteCountFormatter.string(fromByteCount: Int64(mediaRef.byteSize), countStyle: .file))
                    .font(.system(size: 12))
                    .foregroundStyle(isSent ? .white.opacity(0.7) : .echoInk55)
            }
        }
        .padding(12)
        .background(isSent ? Color.echoSignal : Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private func mediaSkeleton(width: CGFloat, height: CGFloat) -> some View {
        RoundedRectangle(cornerRadius: 16)
            .fill(Color.echoPaperDim)
            .frame(width: width, height: height)
            .overlay(ProgressView())
    }

    private var mediaErrorView: some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.triangle")
                .foregroundStyle(Color.echoError)
            Text(loadError ?? "Failed to load")
                .font(.system(size: 13))
                .foregroundStyle(Color.echoInk55)
        }
        .padding(12)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var defaultWaveform: [CGFloat] {
        (0..<30).map { _ in CGFloat.random(in: 0.15...0.85) }
    }

    private func loadMedia() async {
        guard let mediaService = DIContainer.shared.resolveMediaMessage() else {
            isLoading = false
            loadError = "Media service unavailable"
            return
        }
        do {
            let data = try await mediaService.downloadAndDecrypt(ref: mediaRef, peerDID: peerDID)
            switch mediaKind {
            case .image, .video:
                imageData = data
                _ = MediaThumbnailCache.thumbnail(for: mediaRef.fileId, from: data)
            case .audio:
                audioData = data
                if let bars = mediaRef.waveformBars, !bars.isEmpty {
                    waveformSamples = bars.map { CGFloat($0) }
                } else {
                    waveformSamples = WaveformExtractor.samples(from: data)
                }
            case .file:
                break
            }
            isLoading = false
        } catch {
            isLoading = false
            loadError = error.localizedDescription
        }
    }
}
#endif
