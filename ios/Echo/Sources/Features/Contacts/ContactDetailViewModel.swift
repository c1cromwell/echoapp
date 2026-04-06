// Features/Contacts/ContactDetailViewModel.swift
// Manages contact detail screen state

import Foundation
import SwiftUI
#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

@MainActor
class ContactDetailViewModel: ObservableObject {
    let contactId: String

    @Published var contact: ContactDetail = .empty
    @Published var sharedMedia: [SharedMediaItem] = []
    @Published var notificationsEnabled = true
    @Published var disappearingEnabled = false

    @Published var showVoiceCall = false
    @Published var showVideoCall = false
    @Published var showSearch = false
    @Published var showMediaGallery = false
    @Published var showBlockConfirmation = false
    @Published var showReportSheet = false

    init(contactId: String) {
        self.contactId = contactId
    }

    func loadContact() async {
        // TODO: Load from contacts repository
    }

    func copyDID() {
        #if canImport(UIKit)
        UIPasteboard.general.string = contact.did
        #elseif canImport(AppKit)
        NSPasteboard.general.setString(contact.did, forType: .string)
        #endif
    }

    func shareContact() {
        // TODO: Share via UIActivityViewController
    }
}

// MARK: - Contact Detail Model

struct ContactDetail {
    let id: String
    let name: String
    let echoHandle: String
    let avatarURL: URL?
    let trustTier: TrustTier
    let trustScore: Int
    let did: String
    let isOnline: Bool
    let verifiedDate: String
    let mutualGroups: Int
    let mutualContacts: Int
    let credentials: [ContactCredential]

    var didShort: String {
        guard did.count > 20 else { return did }
        return "\(did.prefix(12))...\(did.suffix(6))"
    }

    static let empty = ContactDetail(
        id: "", name: "", echoHandle: "", avatarURL: nil,
        trustTier: .newcomer, trustScore: 0, did: "",
        isOnline: false, verifiedDate: "", mutualGroups: 0,
        mutualContacts: 0, credentials: []
    )
}

struct ContactCredential: Identifiable {
    let id: String
    let name: String
    let issuer: String
    let isVerified: Bool
}

struct SharedMediaItem: Identifiable {
    let id: String
    let thumbnailURL: URL?
    let type: MediaItemType
    let timestamp: Date
}

enum MediaItemType: String {
    case photo, video, file, link
}
