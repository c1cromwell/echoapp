// Core/Mesh/MeshEntitlements.swift
//
// Capability limits for the BLE-mesh transport. Two axes, deliberately decoupled:
//   - lane:   .verified (peers are IDV'd ECHO DIDs) vs .anonymous (ephemeral, no identity)
//   - isVIP:  whether the user has an active VIP subscription
//
// Anonymity is NEVER paywalled. The anonymous lane is free but protocol-limited for abuse
// control; VIP only *lifts* perks (hops, group size, priority relay, persistence). The
// anonymous lane stays ephemeral (no persistence) even for VIP, to preserve its guarantee.

import Foundation

public enum MeshLane: String, Sendable, CaseIterable {
    case verified
    case anonymous
}

public struct MeshLimits: Equatable, Sendable {
    /// Max relay hops a message may traverse (bitchat caps at 7).
    public let maxHops: Int
    /// Max participants in a mesh group chat.
    public let maxGroupSize: Int
    /// Whether this user's frames are prioritized when relays are congested.
    public let priorityRelay: Bool
    /// Whether messages persist to local history (anonymous lane is always ephemeral).
    public let persistence: Bool
}

public enum MeshEntitlements {
    /// Absolute ceiling shared by the protocol (bitchat mesh design).
    public static let protocolMaxHops = 7

    public static func limits(lane: MeshLane, isVIP: Bool) -> MeshLimits {
        switch (lane, isVIP) {
        case (.verified, false):
            return MeshLimits(maxHops: 7, maxGroupSize: 32, priorityRelay: false, persistence: true)
        case (.verified, true):
            return MeshLimits(maxHops: 7, maxGroupSize: 128, priorityRelay: true, persistence: true)
        case (.anonymous, false):
            // Free but limited: short reach, small rooms, no history — abuse control, not paywall.
            return MeshLimits(maxHops: 3, maxGroupSize: 8, priorityRelay: false, persistence: false)
        case (.anonymous, true):
            // VIP lifts reach/size but the lane stays ephemeral by design.
            return MeshLimits(maxHops: 7, maxGroupSize: 32, priorityRelay: true, persistence: false)
        }
    }
}
