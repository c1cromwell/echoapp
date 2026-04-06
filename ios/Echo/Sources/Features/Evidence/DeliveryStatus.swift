// Features/Evidence/DeliveryStatus.swift
// Message delivery lifecycle with anchoring and verification states

import SwiftUI

/// Message delivery lifecycle.
/// Each status is progressive — a message moves forward through these states.
enum DeliveryStatus: String, Codable, Comparable {
    case sending      // Encrypting / queued locally (offline)
    case sent         // Accepted by relay server, recipient offline
    case delivered    // Delivered to recipient's device
    case read         // Recipient opened the message
    case failed       // Relay rejected or unrecoverable error
    case anchored     // Commitment included in finalized metagraph snapshot (all users)
    case verified     // Digital Evidence fingerprint anchored (Org tier + Smart Checkmark)

    /// Icon displayed next to message timestamp.
    var icon: String {
        switch self {
        case .sending:   return "arrow.up.circle"
        case .sent:      return "checkmark"
        case .delivered: return "checkmark.circle"
        case .read:      return "eye"
        case .failed:    return "exclamationmark.circle"
        case .anchored:  return "link"
        case .verified:  return "checkmark.seal"
        }
    }

    /// Color for the status icon.
    var iconColor: Color {
        switch self {
        case .sending:   return Color.Echo.onSurfaceVariant
        case .sent:      return Color.Echo.onSurfaceVariant
        case .delivered: return Color.Echo.primaryContainer
        case .read:      return Color.Echo.primaryContainer
        case .failed:    return Color.Echo.error
        case .anchored:  return Color.Echo.secondary
        case .verified:  return Color.Echo.primaryContainer
        }
    }

    /// Whether tapping the icon opens a verification URL.
    var hasVerificationURL: Bool {
        self == .verified
    }

    /// Display label for the status.
    var displayLabel: String {
        switch self {
        case .sending:   return "Sending"
        case .sent:      return "Sent"
        case .delivered: return "Delivered"
        case .read:      return "Read"
        case .failed:    return "Failed"
        case .anchored:  return "Anchored"
        case .verified:  return "Verified"
        }
    }

    // MARK: - Comparable

    private var sortOrder: Int {
        switch self {
        case .sending:   return 0
        case .sent:      return 1
        case .delivered: return 2
        case .read:      return 3
        case .failed:    return -1
        case .anchored:  return 4
        case .verified:  return 5
        }
    }

    static func < (lhs: DeliveryStatus, rhs: DeliveryStatus) -> Bool {
        lhs.sortOrder < rhs.sortOrder
    }
}

// MARK: - Smart Checkmark View

struct SmartCheckmarkView: View {
    let status: DeliveryStatus
    let eventId: String?
    let onTapVerified: ((String) -> Void)?

    var body: some View {
        Button {
            if status == .verified, let eventId {
                onTapVerified?(eventId)
            }
        } label: {
            Image(systemName: status.icon)
                .font(.system(size: 12))
                .foregroundStyle(status.iconColor)
        }
        .disabled(!status.hasVerificationURL)
        .buttonStyle(.plain)
    }
}
