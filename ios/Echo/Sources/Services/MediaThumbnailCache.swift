#if os(iOS)
import UIKit

/// Local thumbnail cache for encrypted media blobs (M5b). Never stores plaintext off-device paths.
enum MediaThumbnailCache {
    private static let folderName = "media_thumbnails"

    static func thumbnail(for fileId: String, from imageData: Data, maxDimension: CGFloat = 320) -> UIImage? {
        if let cached = load(fileId: fileId) { return cached }
        guard let image = downsample(imageData: imageData, maxDimension: maxDimension) else { return nil }
        save(fileId: fileId, image: image)
        return image
    }

    static func load(fileId: String) -> UIImage? {
        let url = fileURL(fileId: fileId)
        guard let data = try? Data(contentsOf: url) else { return nil }
        return UIImage(data: data)
    }

    private static func save(fileId: String, image: UIImage) {
        guard let data = image.jpegData(compressionQuality: 0.75) else { return }
        let url = fileURL(fileId: fileId)
        try? FileManager.default.createDirectory(at: url.deletingLastPathComponent(), withIntermediateDirectories: true)
        try? data.write(to: url, options: .atomic)
    }

    private static func fileURL(fileId: String) -> URL {
        let base = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        return base.appendingPathComponent(folderName, isDirectory: true).appendingPathComponent("\(fileId).jpg")
    }

    private static func downsample(imageData: Data, maxDimension: CGFloat) -> UIImage? {
        let options: [CFString: Any] = [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceThumbnailMaxPixelSize: maxDimension,
            kCGImageSourceCreateThumbnailWithTransform: true,
        ]
        guard let source = CGImageSourceCreateWithData(imageData as CFData, nil),
              let cgImage = CGImageSourceCreateThumbnailAtIndex(source, 0, options as CFDictionary) else {
            return UIImage(data: imageData)
        }
        return UIImage(cgImage: cgImage)
    }
}
#endif
