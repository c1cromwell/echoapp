# BLE Mesh — Two-Device Runtime Test (P0 gate)

The mesh protocol/routing logic is unit-tested (`Phase3Tests/MeshCoreTests.swift`:
frame round-trip, fragmentation/reassembly, TTL/dedup/relay). The **CoreBluetooth driver**
(`Sources/Core/Mesh/BLEMeshService.swift`) can only be verified on real hardware — the
simulator has no Bluetooth radio. This is the P0 gate.

## What P0 proves
Two iPhones, **offline** (airplane mode + Bluetooth on, no Wi-Fi/cellular), discover each
other and exchange one encoded `MeshPacket` over BLE GATT.

## Setup
- 2 physical iPhones (iOS 17+), each with the `EchoMessaging` scheme installed (real signing,
  not the unsigned CI build — CoreBluetooth needs the app to run on-device).
- Both: **Airplane Mode ON, Bluetooth ON.** Grant the Bluetooth permission prompt
  (`NSBluetoothAlwaysUsageDescription`, already in the Info.plists).

## Driving a send (until P1 wires the chat UI)
`BLEMeshService` has no UI trigger yet (that lands in P1 via `MeshSignalTransport`). To spike
P0, add a temporary debug button that calls:

```swift
let mesh = BLEMeshService(localID: MeshPacket.peerID(from: Data("deviceA".utf8)))
mesh.onMessage = { payload, peer in print("MESH RX:", String(decoding: payload, as: UTF8.self)) }
mesh.start()
// on the other device use localID "deviceB"
mesh.send(Data("ping from A".utf8), to: MeshPacket.broadcast)
```

## Expected
1. Within a few seconds the central on each device discovers + connects to the other
   (`onPeerCountChange` → 1).
2. Tapping send on A logs `MESH RX: ping from A` on B (and vice-versa), with **all radios but
   Bluetooth disabled** — proving the offline path.
3. A third device in range relays a broadcast it is not the destination of (multi-hop): place
   B out of A's range but in C's range, with C between them.

## Known on-device follow-ups (not blocking P0)
- **Background BLE limits:** iOS moves the service UUID to the advertising *overflow* area in
  the background and restricts scanning for unknown peers; foreground works first, then tune
  background (`bluetooth-central`/`bluetooth-peripheral` modes are already declared).
- **MTU/fragmentation:** `defaultFrameSize` is conservative (180B); negotiate
  `maximumWriteValueLength` per peer and raise it.
- **Battery:** add duty-cycled scanning (bitchat's adaptive modes).
- **Reconnect storms / dedup memory:** validate the bounded `seen` set under sustained flooding.
