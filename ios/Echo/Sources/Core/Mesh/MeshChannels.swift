// Core/Mesh/MeshChannels.swift
//
// Geohash-defined location channels (P4): a room is the geohash of an area; coarser precision =
// larger area (block / neighborhood / city / region), à la bitchat. Pure — unit-testable.

import Foundation

public enum Geohash {
    private static let base32 = Array("0123456789bcdefghjkmnpqrstuvwxyz")

    public static func encode(latitude: Double, longitude: Double, precision: Int = 6) -> String {
        var latRange = (-90.0, 90.0)
        var lonRange = (-180.0, 180.0)
        var hash = ""
        var even = true        // longitude bit first
        var bit = 0
        var ch = 0
        while hash.count < precision {
            if even {
                let mid = (lonRange.0 + lonRange.1) / 2
                if longitude >= mid { ch = (ch << 1) | 1; lonRange.0 = mid } else { ch <<= 1; lonRange.1 = mid }
            } else {
                let mid = (latRange.0 + latRange.1) / 2
                if latitude >= mid { ch = (ch << 1) | 1; latRange.0 = mid } else { ch <<= 1; latRange.1 = mid }
            }
            even.toggle()
            bit += 1
            if bit == 5 {
                hash.append(base32[ch])
                bit = 0
                ch = 0
            }
        }
        return hash
    }
}

public enum MeshChannelPrecision {
    public static let block = 7
    public static let neighborhood = 6
    public static let city = 5
    public static let region = 4
}

public enum MeshChannel {
    /// The geohash room id for a location at the given precision.
    public static func channelID(latitude: Double, longitude: Double, precision: Int) -> String {
        Geohash.encode(latitude: latitude, longitude: longitude, precision: precision)
    }

    /// Mesh recipient id for a channel — frames flood to everyone in range; members filter by channel.
    public static func peerID(forChannel channel: String) -> Data {
        MeshPacket.peerID(from: Data(("geo:" + channel).utf8))
    }
}
