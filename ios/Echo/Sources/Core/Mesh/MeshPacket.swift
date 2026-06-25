// Core/Mesh/MeshPacket.swift
//
// Compact wire frame for the ECHO BLE mesh. Design derived from bitchat (public domain);
// see Sources/Core/Mesh/NOTICE. Pure Foundation so it is unit-testable on every platform.

import Foundation

public enum MeshPacketType: UInt8, Sendable {
    case message = 1       // an application payload (possibly a fragment, see flags)
    case ack = 2           // delivery acknowledgement
    case keyAnnounce = 3   // signed messaging-key cert for offline verified bootstrap (P1)
}

public struct MeshPacketFlags: OptionSet, Sendable {
    public let rawValue: UInt8
    public init(rawValue: UInt8) { self.rawValue = rawValue }
    /// The payload is one fragment of a larger message (decode with MeshReassembler).
    public static let fragmented = MeshPacketFlags(rawValue: 1 << 0)
}

/// One hop-able mesh frame. Big-endian, fixed 38-byte header + variable payload:
///   version(1) type(1) ttl(1) flags(1) messageID(16) sender(8) recipient(8) length(2) payload
public struct MeshPacket: Equatable, Sendable {
    public static let version: UInt8 = 1
    public static let idSize = 16
    public static let peerIDSize = 8
    public static let headerSize = 1 + 1 + 1 + 1 + idSize + peerIDSize + peerIDSize + 2  // 38
    /// All-zero recipient = flood to every peer.
    public static let broadcast = Data(repeating: 0, count: peerIDSize)

    public var type: MeshPacketType
    public var ttl: UInt8
    public var flags: MeshPacketFlags
    public var messageID: Data   // 16 bytes — dedup key, stable across relays
    public var sender: Data      // 8 bytes — short peer id
    public var recipient: Data   // 8 bytes — short peer id, or broadcast
    public var payload: Data

    public init(
        type: MeshPacketType,
        ttl: UInt8,
        flags: MeshPacketFlags = [],
        messageID: Data,
        sender: Data,
        recipient: Data,
        payload: Data
    ) {
        self.type = type
        self.ttl = ttl
        self.flags = flags
        self.messageID = MeshPacket.fit(messageID, to: MeshPacket.idSize)
        self.sender = MeshPacket.fit(sender, to: MeshPacket.peerIDSize)
        self.recipient = MeshPacket.fit(recipient, to: MeshPacket.peerIDSize)
        self.payload = payload
    }

    public var isBroadcast: Bool { recipient == MeshPacket.broadcast }

    public func encoded() -> Data {
        var out = Data(capacity: MeshPacket.headerSize + payload.count)
        out.append(MeshPacket.version)
        out.append(type.rawValue)
        out.append(ttl)
        out.append(flags.rawValue)
        out.append(messageID)
        out.append(sender)
        out.append(recipient)
        let len = UInt16(min(payload.count, Int(UInt16.max)))
        out.append(UInt8(len >> 8))
        out.append(UInt8(len & 0xFF))
        out.append(payload.prefix(Int(len)))
        return out
    }

    public static func decode(_ data: Data) -> MeshPacket? {
        guard data.count >= headerSize else { return nil }
        let b = [UInt8](data)
        guard b[0] == version, let type = MeshPacketType(rawValue: b[1]) else { return nil }
        var i = 4
        let messageID = Data(b[i..<i+idSize]); i += idSize
        let sender = Data(b[i..<i+peerIDSize]); i += peerIDSize
        let recipient = Data(b[i..<i+peerIDSize]); i += peerIDSize
        let len = (Int(b[i]) << 8) | Int(b[i+1]); i += 2
        guard b.count >= i + len else { return nil }
        let payload = Data(b[i..<i+len])
        return MeshPacket(
            type: type, ttl: b[2], flags: MeshPacketFlags(rawValue: b[3]),
            messageID: messageID, sender: sender, recipient: recipient, payload: payload
        )
    }

    /// Pad (with zeros) or truncate `data` to exactly `n` bytes.
    static func fit(_ data: Data, to n: Int) -> Data {
        if data.count == n { return data }
        if data.count > n { return data.prefix(n) }
        return data + Data(repeating: 0, count: n - data.count)
    }

    /// Derive a stable 8-byte peer id from arbitrary identity bytes (e.g. a public key or DID).
    public static func peerID(from identity: Data) -> Data {
        // FNV-1a 64-bit — small, dependency-free, good enough for a routing/dedup id.
        var hash: UInt64 = 0xcbf29ce484222325
        for byte in identity {
            hash ^= UInt64(byte)
            hash = hash &* 0x100000001b3
        }
        var be = hash.bigEndian
        return withUnsafeBytes(of: &be) { Data($0) }
    }

    public static func randomMessageID() -> Data {
        var bytes = [UInt8](repeating: 0, count: idSize)
        for i in 0..<idSize { bytes[i] = UInt8.random(in: 0...255) }
        return Data(bytes)
    }
}
