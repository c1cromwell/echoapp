// Core/Mesh/BLEMeshService.swift
//
// CoreBluetooth driver for the ECHO mesh: each device is BOTH a peripheral (advertises the
// mesh service, serves a write/notify characteristic) and a central (scans for peers,
// connects, subscribes). Incoming bytes flow into `MeshRouter`; the router's broadcasts flow
// back out to every connected peer/subscribed central. Design derived from bitchat (public
// domain); see Sources/Core/Mesh/NOTICE.
//
// Runtime behavior (advertising/scanning, multi-hop relay) can only be verified on real
// hardware — see docs/MESH_TWO_DEVICE_TEST.md. iOS background BLE has hard limits (overflow
// service-UUID advertising, restricted background scanning); tuning those is tracked for the
// on-device pass.

#if os(iOS)
import Foundation
import CoreBluetooth

public final class BLEMeshService: NSObject, MeshNode {
    /// Delivered when an application payload (`.message`) addressed to us (or broadcast) arrives.
    public var onMessage: ((_ payload: Data, _ fromPeer: Data) -> Void)?
    /// Delivered when a key-announce cert (`.keyAnnounce`) arrives — feed it to MeshPeerCache for
    /// offline verified-peer key resolution.
    public var onKeyAnnounce: ((_ certData: Data, _ fromPeer: Data) -> Void)?
    /// Observability: peer-count changes (for UI "N people nearby").
    public var onPeerCountChange: ((Int) -> Void)?

    private let router: MeshRouter
    private let queue = DispatchQueue(label: "app.echo.mesh.ble")

    private var central: CBCentralManager?
    private var peripheralManager: CBPeripheralManager?

    // Central role: connected peers and their writable characteristic.
    private var peers: [UUID: (peripheral: CBPeripheral, characteristic: CBMutableCharacteristic?)] = [:]
    private var discovered: [UUID: CBPeripheral] = [:]
    private var writeCharacteristics: [UUID: CBCharacteristic] = [:]
    // Peripheral role: the served characteristic + centrals subscribed for notify.
    private var frameCharacteristic: CBMutableCharacteristic?

    public init(localID: Data, frameSize: Int = BLEMeshConstants.defaultFrameSize) {
        self.router = MeshRouter(localID: localID, maxPacketSize: frameSize)
        super.init()
        self.router.delegate = self
    }

    public func start() {
        central = CBCentralManager(delegate: self, queue: queue)
        peripheralManager = CBPeripheralManager(delegate: self, queue: queue)
    }

    public func stop() {
        central?.stopScan()
        peripheralManager?.stopAdvertising()
        for (_, entry) in peers { central?.cancelPeripheralConnection(entry.peripheral) }
        peers.removeAll()
        discovered.removeAll()
        writeCharacteristics.removeAll()
        notifyPeerCount()
    }

    /// Originate an application payload to a peer id (or `MeshPacket.broadcast`).
    public func send(_ payload: Data, to recipient: Data) {
        queue.async { [weak self] in
            self?.router.send(payload: payload, to: recipient, type: .message)
        }
    }

    /// Broadcast our signed messaging-key cert so nearby peers can resolve our key offline.
    public func announce(_ certData: Data) {
        queue.async { [weak self] in
            self?.router.send(payload: certData, to: MeshPacket.broadcast, type: .keyAnnounce)
        }
    }

    private func notifyPeerCount() {
        let count = peers.count
        DispatchQueue.main.async { [weak self] in self?.onPeerCountChange?(count) }
    }

    /// Write/notify an encoded frame to every connected peer and subscribed central.
    fileprivate func transmit(_ frame: Data) {
        // Central role: write to each connected peer's characteristic.
        for (_, peripheral) in discovered {
            if let ch = writeCharacteristics[peripheral.identifier] {
                let type: CBCharacteristicWriteType =
                    ch.properties.contains(.writeWithoutResponse) ? .withoutResponse : .withResponse
                peripheral.writeValue(frame, for: ch, type: type)
            }
        }
        // Peripheral role: notify subscribed centrals.
        if let ch = frameCharacteristic {
            _ = peripheralManager?.updateValue(frame, for: ch, onSubscribedCentrals: nil)
        }
    }
}

// MARK: - MeshRouterDelegate

extension BLEMeshService: MeshRouterDelegate {
    public func meshRouter(_ router: MeshRouter, didReceive payload: Data, from sender: Data,
                           type: MeshPacketType, messageID: Data) {
        DispatchQueue.main.async { [weak self] in
            switch type {
            case .keyAnnounce: self?.onKeyAnnounce?(payload, sender)
            default:           self?.onMessage?(payload, sender)
            }
        }
    }
    public func meshRouter(_ router: MeshRouter, broadcast packet: MeshPacket) {
        transmit(packet.encoded())
    }
}

// MARK: - Central role (scan, connect, subscribe, receive notifications)

extension BLEMeshService: CBCentralManagerDelegate {
    public func centralManagerDidUpdateState(_ central: CBCentralManager) {
        guard central.state == .poweredOn else { return }
        central.scanForPeripherals(
            withServices: [BLEMeshConstants.serviceUUID],
            options: [CBCentralManagerScanOptionAllowDuplicatesKey: false]
        )
    }

    public func centralManager(_ central: CBCentralManager, didDiscover peripheral: CBPeripheral,
                               advertisementData: [String: Any], rssi RSSI: NSNumber) {
        guard discovered[peripheral.identifier] == nil else { return }
        discovered[peripheral.identifier] = peripheral
        central.connect(peripheral, options: nil)
    }

    public func centralManager(_ central: CBCentralManager, didConnect peripheral: CBPeripheral) {
        peripheral.delegate = self
        peripheral.discoverServices([BLEMeshConstants.serviceUUID])
        notifyPeerCount()
    }

    public func centralManager(_ central: CBCentralManager, didDisconnectPeripheral peripheral: CBPeripheral,
                               error: Error?) {
        discovered[peripheral.identifier] = nil
        writeCharacteristics[peripheral.identifier] = nil
        peers[peripheral.identifier] = nil
        notifyPeerCount()
        // Best-effort reconnect; the mesh is meant to be resilient.
        central.connect(peripheral, options: nil)
    }
}

extension BLEMeshService: CBPeripheralDelegate {
    public func peripheral(_ peripheral: CBPeripheral, didDiscoverServices error: Error?) {
        for service in peripheral.services ?? [] where service.uuid == BLEMeshConstants.serviceUUID {
            peripheral.discoverCharacteristics([BLEMeshConstants.frameCharacteristicUUID], for: service)
        }
    }

    public func peripheral(_ peripheral: CBPeripheral, didDiscoverCharacteristicsFor service: CBService,
                           error: Error?) {
        for ch in service.characteristics ?? [] where ch.uuid == BLEMeshConstants.frameCharacteristicUUID {
            writeCharacteristics[peripheral.identifier] = ch
            peripheral.setNotifyValue(true, for: ch)
        }
    }

    public func peripheral(_ peripheral: CBPeripheral, didUpdateValueFor characteristic: CBCharacteristic,
                           error: Error?) {
        guard let data = characteristic.value else { return }
        router.ingest(data)
    }
}

// MARK: - Peripheral role (advertise, serve characteristic, receive writes)

extension BLEMeshService: CBPeripheralManagerDelegate {
    public func peripheralManagerDidUpdateState(_ peripheral: CBPeripheralManager) {
        guard peripheral.state == .poweredOn else { return }
        let characteristic = CBMutableCharacteristic(
            type: BLEMeshConstants.frameCharacteristicUUID,
            properties: [.write, .writeWithoutResponse, .notify],
            value: nil,
            permissions: [.writeable]
        )
        let service = CBMutableService(type: BLEMeshConstants.serviceUUID, primary: true)
        service.characteristics = [characteristic]
        frameCharacteristic = characteristic
        peripheral.add(service)
        peripheral.startAdvertising([CBAdvertisementDataServiceUUIDsKey: [BLEMeshConstants.serviceUUID]])
    }

    public func peripheralManager(_ peripheral: CBPeripheralManager,
                                  didReceiveWrite requests: [CBATTRequest]) {
        for request in requests {
            if let data = request.value { router.ingest(data) }
        }
        peripheral.respond(to: requests[0], withResult: .success)
    }
}
#endif
