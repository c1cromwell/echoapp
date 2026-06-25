// Core/Mesh/MeshDebugView.swift
//
// DEBUG-only harness to drive the P0 two-device BLE spike before P1 wires the chat UI.
// Reachable from Settings under #if DEBUG. See docs/MESH_TWO_DEVICE_TEST.md.

#if os(iOS) && DEBUG
import SwiftUI

struct MeshDebugView: View {
    @State private var service: BLEMeshService?
    @State private var running = false
    @State private var peerCount = 0
    @State private var localName = "deviceA"
    @State private var message = "ping"
    @State private var log: [String] = []

    var body: some View {
        List {
            Section("Node") {
                TextField("Local name (use different names per device)", text: $localName)
                    .disabled(running)
                    .autocorrectionDisabled()
                HStack {
                    Text("Peers nearby")
                    Spacer()
                    Text("\(peerCount)").monospacedDigit().foregroundColor(.secondary)
                }
                Button(running ? "Stop mesh" : "Start mesh") { running ? stop() : start() }
            }
            Section("Send (broadcast)") {
                TextField("Message", text: $message)
                Button("Send") { send() }.disabled(!running)
            }
            Section("Log") {
                if log.isEmpty {
                    Text("—").foregroundColor(.secondary)
                } else {
                    ForEach(Array(log.enumerated()), id: \.offset) { _, line in
                        Text(line).font(.system(size: 13, design: .monospaced))
                    }
                }
            }
        }
        .navigationTitle("Mesh debug")
        .navigationBarTitleDisplayMode(.inline)
        .onDisappear { stop() }
    }

    private func start() {
        let svc = BLEMeshService(localID: MeshPacket.peerID(from: Data(localName.utf8)))
        svc.onPeerCountChange = { peerCount = $0 }
        svc.onMessage = { payload, _ in
            let text = String(data: payload, encoding: .utf8) ?? "<\(payload.count) bytes>"
            log.insert("RX  \(text)", at: 0)
        }
        svc.start()
        service = svc
        running = true
        log.insert("started as \"\(localName)\"", at: 0)
    }

    private func stop() {
        service?.stop()
        service = nil
        running = false
        peerCount = 0
    }

    private func send() {
        service?.send(Data(message.utf8), to: MeshPacket.broadcast)
        log.insert("TX  \(message)", at: 0)
    }
}
#endif
