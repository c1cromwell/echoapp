// Features/Calling/CallState.swift
// Voice/Video call state models

import Foundation

enum CallType: String, Codable {
    case voice, video
}

enum CallState: String {
    case idle, ringing, connecting, active, onHold, ended
}

struct CallParticipant: Identifiable {
    let id: String
    let name: String
    let avatarURL: URL?
    let trustTier: TrustTier
    let isMuted: Bool
    let isVideoOn: Bool
}

enum TrustTier: String, Codable, CaseIterable {
    case newcomer, basic, verified, trusted, elite

    var displayName: String { rawValue.capitalized }

    var color: String {
        switch self {
        case .newcomer: return "#6E7881"
        case .basic: return "#3C627D"
        case .verified: return "#0EA5E9"
        case .trusted: return "#006591"
        case .elite: return "#7DD3FC"
        }
    }
}
