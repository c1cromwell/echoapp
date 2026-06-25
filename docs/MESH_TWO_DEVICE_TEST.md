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

## Live integration status (compile-checked; runtime pending)

`MeshMessagingStack` assembles the path opt-in (does NOT replace the WebSocket default):
- `BLEMeshService` + `MeshPeerCache` (fed by key announces) + `MeshSignalTransport` + `OfflineMessageQueue`.
- `stack.combined(with: websocketTransport)` → a `TransportRouter` that runs relay + mesh in parallel.
- `stack.start()` broadcasts our key-announce cert (via a `MeshCertProvider`) so peers resolve our key offline.

**Remaining on-device hooks (need hardware to verify):**
1. **Secure-Enclave cert provider.** `SoftwareMeshCertProvider` works in sim/tests; on device, sign
   `(did ‖ kaPublicKey)` with `SecureEnclaveManager.sign(keyId: "echo-identity-signing")` and convert its
   DER signature to raw r‖s before storing in `MeshKeyCert.signature`.
2. **Switch the transport.** When the user enables mesh (verified lane, `isIdentityVerified()`), construct
   `ConversationSignalService(transport: stack.combined(with: ws))` in the DI instead of the WS-only default.
3. **Queue on no-peers.** Enqueue outbound frames into `OfflineMessageQueue` when `onPeerCountChange == 0`;
   `markDelivered` on the recipient's ack.

**Groups over mesh:** no new mesh code needed — a flood mesh delivers the group envelope to everyone in
range, and only members decrypt it with the existing `GroupKeyManager` E2E key. Send the normal group
envelope over `stack.transport`; `OfflineMessageQueue` covers store-and-forward for groups too.
