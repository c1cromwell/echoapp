import Foundation

struct AgentGrantReviewItem: Identifiable, Equatable, Sendable {
    var id: String
    var agentName: String
    var verb: String
    var threadID: String
    var expiresAt: Date
}

struct AgentChallenge: Equatable, Sendable {
    var id: String
    var verb: String
    var threadID: String
}

/// Local review state for agent grants. Revoke and confirm are one-shot.
struct GrantReviewState: Equatable, Sendable {
    var grants: [AgentGrantReviewItem]
    var pending: AgentChallenge?

    mutating func revoke(id: String) {
        grants.removeAll { $0.id == id }
    }

    /// Returns false when the challenge is missing or already consumed.
    mutating func confirm(_ challenge: AgentChallenge) -> Bool {
        guard let pending, pending.id == challenge.id, pending.verb == challenge.verb else {
            return false
        }
        self.pending = nil
        return true
    }
}

#if os(iOS)
import SwiftUI

/// Grant review lives under Privacy. It is not a tab.
struct GrantReviewView: View {
    @State private var review: GrantReviewState
    @State private var showConfirmed = false

    init(review: GrantReviewState) {
        _review = State(initialValue: review)
    }

    var body: some View {
        List {
            if review.grants.isEmpty && review.pending == nil {
                Text("No agents can send, read, or search on your behalf.")
                    .foregroundStyle(.secondary)
            }
            Section("Active grants") {
                ForEach(review.grants) { grant in
                    VStack(alignment: .leading, spacing: 4) {
                        Text(grant.agentName)
                        Text("\(grant.verb) · \(grant.threadID)")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                        Text(grant.expiresAt.formatted(date: .abbreviated, time: .shortened))
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                        Button("Revoke", role: .destructive) {
                            review.revoke(id: grant.id)
                        }
                    }
                }
            }
            if let pending = review.pending {
                Section("Needs your confirmation") {
                    Text("\(pending.verb) on \(pending.threadID)")
                    Text("Pay, forward, and key changes finish only on this device.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                    Button("Confirm") {
                        if review.confirm(pending) {
                            showConfirmed = true
                        }
                    }
                }
            }
        }
        .navigationTitle("Agent grants")
        .alert("Confirmed", isPresented: $showConfirmed) {
            Button("OK", role: .cancel) {}
        } message: {
            Text("This device approved it once. The agent cannot repeat it.")
        }
    }
}
#endif
