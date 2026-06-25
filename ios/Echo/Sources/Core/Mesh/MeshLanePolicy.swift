// Core/Mesh/MeshLanePolicy.swift
//
// Applies MeshEntitlements to a running lane: caps hop count, enforces group size, and (for the
// anonymous lane) rate-limits sends for abuse control. Anonymity is never paywalled — the limits
// exist to bound abuse on a permissionless lane; VIP only lifts them.

import Foundation

/// Sliding-window rate limiter (anonymous lane abuse control).
public final class MeshRateLimiter {
    private var timestamps: [Date] = []
    private let limit: Int
    private let window: TimeInterval

    public init(limit: Int, window: TimeInterval) {
        self.limit = limit
        self.window = window
    }

    /// Returns true (and records the event) if under the limit in the trailing window.
    public func allow(now: Date = Date()) -> Bool {
        timestamps.removeAll { now.timeIntervalSince($0) >= window }
        guard timestamps.count < limit else { return false }
        timestamps.append(now)
        return true
    }
}

public struct MeshLanePolicy {
    public let lane: MeshLane
    public let limits: MeshLimits
    private let rateLimiter: MeshRateLimiter?

    public init(lane: MeshLane, isVIP: Bool) {
        self.lane = lane
        self.limits = MeshEntitlements.limits(lane: lane, isVIP: isVIP)
        // Only the anonymous lane is rate-limited; verified peers are accountable.
        self.rateLimiter = lane == .anonymous ? MeshRateLimiter(limit: isVIP ? 60 : 20, window: 60) : nil
    }

    /// Clamp a requested TTL to the lane's hop ceiling.
    public func cappedTTL(_ requested: UInt8) -> UInt8 { min(requested, UInt8(limits.maxHops)) }

    public func allowsGroup(size: Int) -> Bool { size <= limits.maxGroupSize }

    public var persists: Bool { limits.persistence }

    /// Whether an outbound send is allowed now (always true on the verified lane).
    public func allowsSend(now: Date = Date()) -> Bool {
        guard let rateLimiter else { return true }
        return rateLimiter.allow(now: now)
    }
}
