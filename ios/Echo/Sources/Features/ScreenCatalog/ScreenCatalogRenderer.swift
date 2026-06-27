#if os(iOS)
import SwiftUI
import UIKit

/// Renders SwiftUI screens to PNG for `docs/screen_catalog` (headless via iOS Simulator tests).
enum ScreenCatalogRenderer {
    /// iPhone 15 Pro logical point size @3x export.
    static let canvasSize = CGSize(width: 393, height: 852)
    static let exportScale: CGFloat = 3.0

    struct ManifestEntry: Codable, Sendable {
        let journey: String
        let stepId: String
        let title: String
        let e2eRef: String
        let relativePath: String
    }

    static var isGenerationEnabled: Bool {
        ProcessInfo.processInfo.environment["SCREEN_CATALOG_GENERATE"] == "1"
    }

    static var catalogRoot: URL {
        if let root = ProcessInfo.processInfo.environment["SCREEN_CATALOG_ROOT"], !root.isEmpty {
            return URL(fileURLWithPath: root, isDirectory: true)
        }
        // Hosted XCTest on Simulator: write under app Documents; `make screen-catalog` copies out.
        let docs = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first
            ?? FileManager.default.temporaryDirectory
        return docs.appendingPathComponent("screen_catalog", isDirectory: true)
    }

    @MainActor
    static func render<V: View>(
        _ view: V,
        journey: String,
        stepId: String,
        title: String,
        e2eRef: String
    ) throws -> URL {
        let framed = view
            .frame(width: canvasSize.width, height: canvasSize.height)
            .background(Color.Echo.surface)

        let renderer = ImageRenderer(content: framed)
        renderer.scale = exportScale
        renderer.isOpaque = true

        guard let image = renderer.uiImage, let data = image.pngData() else {
            throw ScreenCatalogError.renderFailed(journey: journey, stepId: stepId)
        }

        let journeyDir = catalogRoot.appendingPathComponent(journey, isDirectory: true)
        try FileManager.default.createDirectory(at: journeyDir, withIntermediateDirectories: true)
        try FileManager.default.createDirectory(at: catalogRoot, withIntermediateDirectories: true)

        let fileName = "\(stepId).png"
        let fileURL = journeyDir.appendingPathComponent(fileName)
        try data.write(to: fileURL, options: .atomic)

        try appendManifest(
            ManifestEntry(
                journey: journey,
                stepId: stepId,
                title: title,
                e2eRef: e2eRef,
                relativePath: "\(journey)/\(fileName)"
            )
        )
        return fileURL
    }

    private static func appendManifest(_ entry: ManifestEntry) throws {
        let manifestURL = catalogRoot.appendingPathComponent("manifest.jsonl")
        let line = try String(data: JSONEncoder().encode(entry), encoding: .utf8)! + "\n"
        if FileManager.default.fileExists(atPath: manifestURL.path) {
            let handle = try FileHandle(forWritingTo: manifestURL)
            try handle.seekToEnd()
            try handle.write(contentsOf: Data(line.utf8))
            try handle.close()
        } else {
            try line.write(to: manifestURL, atomically: true, encoding: .utf8)
        }
    }

    @MainActor
    static func resetCatalog() throws {
        let fm = FileManager.default
        if fm.fileExists(atPath: catalogRoot.path) {
            try fm.removeItem(at: catalogRoot)
        }
        try fm.createDirectory(at: catalogRoot, withIntermediateDirectories: true)
    }

    enum ScreenCatalogError: Error, CustomStringConvertible {
        case renderFailed(journey: String, stepId: String)

        var description: String {
            switch self {
            case .renderFailed(let journey, let stepId):
                return "ScreenCatalogRenderer: failed to render \(journey)/\(stepId)"
            }
        }
    }
}
#endif
