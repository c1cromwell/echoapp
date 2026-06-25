// Core/Mesh/MeshKeyAnnounce.swift
//
// Serialization for the `.keyAnnounce` mesh payload: a node broadcasts its signed MeshKeyCert so
// nearby peers can verify + cache its messaging key for offline verified messaging. Larger than
// one BLE frame, so MeshRouter fragments/reassembles it automatically.

import Foundation

public enum MeshKeyAnnounce {
    public static func encode(_ cert: MeshKeyCert) -> Data {
        (try? JSONEncoder().encode(cert)) ?? Data()
    }
    public static func decode(_ data: Data) -> MeshKeyCert? {
        try? JSONDecoder().decode(MeshKeyCert.self, from: data)
    }
}
