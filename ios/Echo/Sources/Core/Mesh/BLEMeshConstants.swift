// Core/Mesh/BLEMeshConstants.swift
//
// ECHO's own GATT identifiers for the mesh. These are NOT the bitchat UUIDs — the
// ECHO mesh is a separate network. See Sources/Core/Mesh/NOTICE.

import Foundation
#if canImport(CoreBluetooth)
import CoreBluetooth

public enum BLEMeshConstants {
    /// Advertised primary service that marks a device as an ECHO mesh node.
    public static let serviceUUID = CBUUID(string: "E340C001-2A1B-4E55-9C2D-0E43484F4D45")
    /// Single write/notify characteristic carrying encoded `MeshPacket` frames.
    public static let frameCharacteristicUUID = CBUUID(string: "E340C002-2A1B-4E55-9C2D-0E43484F4D45")

    /// Conservative frame budget so a small message fits one default-MTU write without
    /// fragmentation; larger payloads are fragmented by `MeshRouter`.
    public static let defaultFrameSize = 180
}
#endif
