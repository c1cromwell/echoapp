#if os(iOS)
import Foundation

/// Tracks in-progress resumable media uploads for offline retry (WO-21/34).
struct MediaUploadResumeState: Codable, Sendable, Equatable {
    let fileId: String
    let totalChunks: Int
    let mimeType: String
    let trustTier: Int
    let encryptedSize: Int
    var receivedChunks: Set<Int>
    let createdAt: Date
}

enum MediaUploadResumeStore {
    private static let key = "echo.media.upload.resume.v1"

    static func save(_ state: MediaUploadResumeState) {
        var all = loadAll()
        all[state.fileId] = state
        persist(all)
    }

    static func load(fileId: String) -> MediaUploadResumeState? {
        loadAll()[fileId]
    }

    static func clear(fileId: String) {
        var all = loadAll()
        all.removeValue(forKey: fileId)
        persist(all)
    }

    static func pending() -> [MediaUploadResumeState] {
        Array(loadAll().values)
    }

    private static func loadAll() -> [String: MediaUploadResumeState] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let map = try? JSONDecoder().decode([String: MediaUploadResumeState].self, from: data) else {
            return [:]
        }
        return map
    }

    private static func persist(_ map: [String: MediaUploadResumeState]) {
        guard let data = try? JSONEncoder().encode(map) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
