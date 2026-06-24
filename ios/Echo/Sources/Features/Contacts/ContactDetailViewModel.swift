#if os(iOS)
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
    private let preferredDisplayName: String?

    @Published var contact: ContactDetail = .empty
    @Published var mutualGroups: [ContactSocialAPIClient.MutualGroup] = []
    @Published var mutualContacts: [ContactSocialAPIClient.MutualContact] = []
    @Published var relationshipError: String?
    @Published var sharedMedia: [SharedMediaItem] = []
    @Published var notificationsEnabled = true
    @Published var disappearingEnabled = false

    @Published var showVoiceCall = false
    @Published var showVideoCall = false
    @Published var showSearch = false
    @Published var showMediaGallery = false
    @Published var showBlockConfirmation = false
    @Published var showReportSheet = false
    @Published var isBlocked = false
    @Published var blockError: String?

    private var socialAPI: ContactSocialAPIClient? {
        guard let client = DIContainer.shared.resolveAPIClient() else { return nil }
        return ContactSocialAPIClient(apiClient: client)
    }

    init(contactId: String, displayName: String? = nil) {
        self.contactId = contactId
        self.preferredDisplayName = displayName
    }

    func loadContact() async {
        let tier = ContactTrustIndex.shared.tier(conversationId: "", peerDID: contactId)
        var name = preferredDisplayName ?? ContactThreadHelper.truncatedDID(contactId)
        var handle = contactId

        if let client = DIContainer.shared.resolveAPIClient() {
            let discovery = ContactDiscoveryAPIClient(apiClient: client)
            if let profile = try? await discovery.resolveIdentity(did: contactId) {
                if let username = profile.username, !username.isEmpty {
                    handle = "@\(username)"
                    if preferredDisplayName == nil {
                        name = handle
                    }
                }
            }
        }

        contact = ContactDetail(
            id: contactId,
            name: name,
            echoHandle: handle,
            avatarURL: nil,
            trustTier: Self.trustTier(fromNumeric: tier),
            trustScore: min(100, tier * 20),
            did: contactId,
            isOnline: false,
            verifiedDate: "—",
            mutualGroups: 0,
            mutualContacts: 0,
            credentials: []
        )

        await loadRelationship()
    }

    func loadRelationship() async {
        guard let socialAPI else {
            relationshipError = nil
            return
        }
        relationshipError = nil
        do {
            let rel = try await socialAPI.fetchRelationship(peerDID: contactId)
            mutualGroups = rel.mutual_groups ?? []
            mutualContacts = rel.mutual_contacts ?? []
            contact = ContactDetail(
                id: contact.id,
                name: contact.name,
                echoHandle: contact.echoHandle,
                avatarURL: contact.avatarURL,
                trustTier: contact.trustTier,
                trustScore: contact.trustScore,
                did: contact.did,
                isOnline: contact.isOnline,
                verifiedDate: contact.verifiedDate,
                mutualGroups: rel.mutual_groups_count ?? mutualGroups.count,
                mutualContacts: rel.mutual_contacts_count ?? mutualContacts.count,
                credentials: contact.credentials
            )
        } catch {
            relationshipError = error.localizedDescription
            mutualGroups = []
            mutualContacts = []
        }
    }

    func blockContact() async {
        guard socialAPI != nil || DIContainer.shared.resolveBlockContactUseCase() != nil else {
            blockError = "Sign in required"
            return
        }
        do {
            if let blockUseCase = DIContainer.shared.resolveBlockContactUseCase() {
                try await blockUseCase.block(did: contactId)
            } else if let socialAPI {
                try await socialAPI.blockContact(did: contactId)
            }
            isBlocked = true
            showBlockConfirmation = false
        } catch {
            blockError = error.localizedDescription
        }
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

    private static func trustTier(fromNumeric tier: Int) -> TrustTier {
        switch tier {
        case 5: return .elite
        case 4: return .trusted
        case 3: return .verified
        case 2: return .basic
        default: return .newcomer
        }
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
#endif
