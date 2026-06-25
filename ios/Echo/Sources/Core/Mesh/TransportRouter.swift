// Core/Mesh/TransportRouter.swift
//
// A ConversationSignalTransport that fans out across several child transports (e.g. the WebSocket
// relay + the BLE mesh) and merges their inbound streams into one. This is how ECHO runs online
// and offline paths in parallel: send over both, and accept a message from whichever delivers it
// first. ConversationSignalService's own message-id handling dedups a frame seen on two paths.

import Foundation

public final class TransportRouter: ConversationSignalTransport, @unchecked Sendable {
    public var onTextMessage: (@Sendable (String) -> Void)?
    private let children: [ConversationSignalTransport]

    init(_ children: [ConversationSignalTransport]) {
        self.children = children
        for child in children {
            child.onTextMessage = { [weak self] text in self?.onTextMessage?(text) }
        }
    }

    /// Connect every child independently — a failure on one path (e.g. no internet for the relay)
    /// must not stop the others (e.g. the mesh, which works offline).
    public func connect(accessToken: String) async throws {
        for child in children { try? await child.connect(accessToken: accessToken) }
    }

    public func disconnect() async {
        for child in children { await child.disconnect() }
    }

    /// Send over all paths; succeeds if at least one accepts. Throws only if every path failed.
    public func send(text: String) async throws {
        var lastError: Error?
        var anySucceeded = false
        for child in children {
            do { try await child.send(text: text); anySucceeded = true }
            catch { lastError = error }
        }
        if !anySucceeded, let lastError { throw lastError }
    }
}
