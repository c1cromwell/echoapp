// Core/Mesh/MeshFragmenter.swift
//
// Splits payloads larger than a single BLE write into ordered fragments and reassembles
// them. Design derived from bitchat (public domain); see Sources/Core/Mesh/NOTICE.
// Fragment payload layout (big-endian): groupID(16) index(2) count(2) data(...)

import Foundation

public enum MeshFragmenter {
    public static let groupIDSize = 16
    public static let fragHeaderSize = groupIDSize + 2 + 2  // 20

    /// Split `payload` into fragment payloads, each ≤ `maxChunk` bytes including the
    /// fragment header. Returns a single un-fragmented chunk when it already fits.
    public static func fragment(_ payload: Data, groupID: Data, maxChunk: Int) -> [Data] {
        let gid = MeshPacket.fit(groupID, to: groupIDSize)
        let room = max(1, maxChunk - fragHeaderSize)
        let chunks = stride(from: 0, to: max(payload.count, 1), by: room).map { start -> Data in
            payload.subdata(in: start..<min(start + room, payload.count))
        }
        let count = UInt16(min(chunks.count, Int(UInt16.max)))
        return chunks.enumerated().map { idx, chunk in
            var out = Data(capacity: fragHeaderSize + chunk.count)
            out.append(gid)
            let i = UInt16(idx)
            out.append(UInt8(i >> 8)); out.append(UInt8(i & 0xFF))
            out.append(UInt8(count >> 8)); out.append(UInt8(count & 0xFF))
            out.append(chunk)
            return out
        }
    }
}

/// Collects fragments (bounded) and returns the full payload once a group is complete.
public final class MeshReassembler {
    private struct Group { var count: Int; var parts: [Int: Data] }
    private var groups: [Data: Group] = [:]
    private var order: [Data] = []
    private let maxGroups: Int

    public init(maxGroups: Int = 64) { self.maxGroups = maxGroups }

    /// Feed a fragment payload (as produced by `MeshFragmenter.fragment`).
    /// Returns the reassembled payload when the last missing fragment arrives, else nil.
    public func add(_ fragmentPayload: Data) -> Data? {
        guard fragmentPayload.count >= MeshFragmenter.fragHeaderSize else { return nil }
        let b = [UInt8](fragmentPayload)
        let gid = Data(b[0..<MeshFragmenter.groupIDSize])
        var i = MeshFragmenter.groupIDSize
        let index = (Int(b[i]) << 8) | Int(b[i+1]); i += 2
        let count = (Int(b[i]) << 8) | Int(b[i+1]); i += 2
        guard count > 0, index < count else { return nil }
        let data = Data(b[i...])

        if groups[gid] == nil {
            evictIfNeeded()
            groups[gid] = Group(count: count, parts: [:])
            order.append(gid)
        }
        groups[gid]?.parts[index] = data
        guard let g = groups[gid], g.parts.count == g.count else { return nil }

        var full = Data()
        for idx in 0..<g.count { full.append(g.parts[idx] ?? Data()) }
        remove(gid)
        return full
    }

    private func evictIfNeeded() {
        while order.count >= maxGroups, let oldest = order.first { remove(oldest) }
    }
    private func remove(_ gid: Data) {
        groups[gid] = nil
        if let idx = order.firstIndex(of: gid) { order.remove(at: idx) }
    }
}
