#if os(iOS)
import Foundation

/// Loads shared media rows from locally persisted thread history.
@MainActor
enum SharedMediaLoader {
    static func conversationId(peerDID: String) -> String {
        ConversationStore.shared.conversations.first { $0.peerDID == peerDID }?.id ?? peerDID
    }

    static func loadGallery(conversationId: String) -> (
        photos: [GalleryMediaItem],
        videos: [GalleryMediaItem],
        files: [GalleryFileItem],
        links: [GalleryLinkItem]
    ) {
        let messages = ConversationThreadStore.exportMessages(conversationId: conversationId)
        var photos: [GalleryMediaItem] = []
        var videos: [GalleryMediaItem] = []
        var files: [GalleryFileItem] = []
        var links: [GalleryLinkItem] = []

        for message in messages {
            let timestamp = StoredThreadMessage.parseSentAt(message.sentAtISO) ?? Date()
            if let wire = MediaMessageService.parseWire(from: message.content) {
                let ref = wire.media
                let item = GalleryMediaItem(
                    id: message.id,
                    thumbnailURL: nil,
                    fullURL: nil,
                    isVideo: ref.mediaKind == MediaKind.video.rawValue,
                    duration: nil,
                    timestamp: timestamp
                )
                switch ref.mediaKind {
                case MediaKind.image.rawValue:
                    photos.append(item)
                case MediaKind.video.rawValue:
                    videos.append(item)
                case MediaKind.audio.rawValue:
                    files.append(GalleryFileItem(
                        id: message.id,
                        name: "Voice note",
                        size: ByteCountFormatter.string(fromByteCount: Int64(ref.byteSize), countStyle: .file),
                        icon: "waveform",
                        timestamp: timestamp
                    ))
                default:
                    files.append(GalleryFileItem(
                        id: message.id,
                        name: ref.caption ?? "File",
                        size: ByteCountFormatter.string(fromByteCount: Int64(ref.byteSize), countStyle: .file),
                        icon: "doc.fill",
                        timestamp: timestamp
                    ))
                }
                continue
            }

            if let url = firstURL(in: message.content) {
                links.append(GalleryLinkItem(
                    id: message.id,
                    url: url,
                    title: message.content,
                    description: nil,
                    imageURL: nil,
                    domain: url.host ?? url.absoluteString
                ))
            }
        }

        return (photos, videos, files, links)
    }

    static func previewItems(peerDID: String, limit: Int = 6) -> [SharedMediaItem] {
        let conversationId = conversationId(peerDID: peerDID)
        let bundle = loadGallery(conversationId: conversationId)
        let media = (bundle.photos + bundle.videos).prefix(limit)
        return media.map { item in
            SharedMediaItem(
                id: item.id,
                thumbnailURL: item.thumbnailURL,
                type: item.isVideo ? .video : .photo,
                timestamp: item.timestamp
            )
        }
    }

    private static func firstURL(in text: String) -> URL? {
        let detector = try? NSDataDetector(types: NSTextCheckingResult.CheckingType.link.rawValue)
        let range = NSRange(text.startIndex..<text.endIndex, in: text)
        let match = detector?.firstMatch(in: text, options: [], range: range)
        guard let match, let url = match.url else { return nil }
        return url
    }
}
#endif
