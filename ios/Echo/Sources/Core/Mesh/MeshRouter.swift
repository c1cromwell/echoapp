// Core/Mesh/MeshRouter.swift
//
// TTL-bounded flooding with per-packet dedup, store-and-forward relay, and
// fragmentation/reassembly. Transport-agnostic (the BLE layer feeds it bytes and
// performs the actual broadcasts). Design derived from bitchat (public domain); see
// Sources/Core/Mesh/NOTICE. Pure Foundation — unit-testable without CoreBluetooth.

import Foundation

public protocol MeshRouterDelegate: AnyObject {
    /// A fully-reassembled payload addressed to us (or broadcast) is ready for the app.
    func meshRouter(_ router: MeshRouter, didReceive payload: Data, from sender: Data, messageID: Data)
    /// Push this frame out to every connected peer (origination and relay both use this).
    func meshRouter(_ router: MeshRouter, broadcast packet: MeshPacket)
}

public final class MeshRouter {
    public weak var delegate: MeshRouterDelegate?

    public let localID: Data           // our 8-byte peer id
    public let maxPacketSize: Int      // BLE write budget (header + payload)
    private let maxHops: UInt8

    private var seen: Set<Data> = []   // packet messageIDs already processed
    private var seenOrder: [Data] = []
    private let seenCap: Int
    private let reassembler = MeshReassembler()

    public init(
        localID: Data,
        maxHops: UInt8 = UInt8(MeshEntitlements.protocolMaxHops),
        maxPacketSize: Int = 512,
        seenCapacity: Int = 2048
    ) {
        self.localID = MeshPacket.fit(localID, to: MeshPacket.peerIDSize)
        self.maxHops = maxHops
        self.maxPacketSize = maxPacketSize
        self.seenCap = seenCapacity
    }

    // MARK: Origination

    /// Originate a message from this node: fragment if needed, mark seen (so we don't relay
    /// our own echoes), and broadcast each frame. Returns the frames sent (for tests/inspection).
    @discardableResult
    public func send(payload: Data, to recipient: Data, ttl: UInt8? = nil) -> [MeshPacket] {
        let hops = min(ttl ?? maxHops, maxHops)
        let maxChunk = max(1, maxPacketSize - MeshPacket.headerSize)
        let packets: [MeshPacket]
        if payload.count <= maxChunk {
            packets = [MeshPacket(type: .message, ttl: hops, messageID: MeshPacket.randomMessageID(),
                                  sender: localID, recipient: recipient, payload: payload)]
        } else {
            let groupID = MeshPacket.randomMessageID()
            packets = MeshFragmenter.fragment(payload, groupID: groupID, maxChunk: maxChunk).map { frag in
                MeshPacket(type: .message, ttl: hops, flags: .fragmented,
                           messageID: MeshPacket.randomMessageID(),
                           sender: localID, recipient: recipient, payload: frag)
            }
        }
        for p in packets {
            _ = markSeen(p.messageID)
            delegate?.meshRouter(self, broadcast: p)
        }
        return packets
    }

    // MARK: Reception (called by the BLE layer for every received frame)

    public func ingest(_ data: Data) {
        guard let packet = MeshPacket.decode(data) else { return }
        ingest(packet)
    }

    public func ingest(_ packet: MeshPacket) {
        // Dedup: a packet we've already handled is dropped (this is what bounds flooding).
        guard !markSeen(packet.messageID) else { return }
        // Ignore our own frames that looped back.
        guard packet.sender != localID else { return }

        let forMe = packet.recipient == localID || packet.isBroadcast
        if forMe {
            if packet.flags.contains(.fragmented) {
                if let full = reassembler.add(packet.payload) {
                    delegate?.meshRouter(self, didReceive: full, from: packet.sender, messageID: packet.messageID)
                }
            } else {
                delegate?.meshRouter(self, didReceive: packet.payload, from: packet.sender, messageID: packet.messageID)
            }
        }

        // Relay: keep flooding unless we are the sole unicast destination or TTL is spent.
        let weAreFinalDestination = packet.recipient == localID && !packet.isBroadcast
        if packet.ttl > 1 && !weAreFinalDestination {
            var relayed = packet
            relayed.ttl -= 1
            delegate?.meshRouter(self, broadcast: relayed)
        }
    }

    // MARK: Dedup set (bounded FIFO)

    /// Returns true if `id` had already been seen.
    private func markSeen(_ id: Data) -> Bool {
        if seen.contains(id) { return true }
        seen.insert(id)
        seenOrder.append(id)
        if seenOrder.count > seenCap {
            let old = seenOrder.removeFirst()
            seen.remove(old)
        }
        return false
    }
}
