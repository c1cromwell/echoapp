#if os(iOS)
import Foundation

/// One-time DEBUG population of the Messages tab with realistic local test data:
/// 4 DM contacts (with threads + trust tiers), 2 group chats, and 2 hidden folders
/// each holding a hidden chat. Everything is injected through the same device-local
/// stores the Messages tab reads from (`ConversationStore`, `ConversationThreadStore`,
/// `ConversationPreferencesStore`, `HiddenFolderStore`, `ContactTrustIndex`) — no
/// backend dependency and no new persistence layer.
///
/// Compiles out entirely in release builds. Supersedes the single "Echo Support" row
/// from `ConversationStore.seedDemoIfNeeded()` on first run.
@MainActor
enum DebugSeedData {
    private static let seededFlag = "echo.debug.seeded.v1"

    /// Populate once. No-ops in release, if already seeded, or before the user DID is
    /// available (pre-auth). Safe to call repeatedly (e.g. every `.active` scene phase).
    static func seedIfNeeded() async {
        #if DEBUG
        guard !UserDefaults.standard.bool(forKey: seededFlag) else { return }
        guard let localDID = await CurrentUserSession.currentDID(), !localDID.isEmpty else { return }

        await seedContacts(localDID: localDID)
        seedGroups(localDID: localDID)
        await seedHiddenFolders(localDID: localDID)

        UserDefaults.standard.set(true, forKey: "echo.hasSentFirstMessage")
        UserDefaults.standard.set(true, forKey: seededFlag)
        #endif
    }

    #if DEBUG

    // MARK: - Contacts (4 DMs)

    private struct SeedContact {
        let name: String
        let peerDID: String
        let trustTier: Int
        let lastMessage: String
        let timestamp: String
        let unreadCount: Int
        let isOnline: Bool
        let muted: Bool
        let disappearing: DisappearingTimer
    }

    private static let contactFixtures: [SeedContact] = [
        SeedContact(
            name: "Jordan Lee", peerDID: "did:key:z6MkTestJordanLee00000000000000000000000000000",
            trustTier: 3, lastMessage: "See you at 3?", timestamp: "9:41 AM",
            unreadCount: 2, isOnline: true, muted: false, disappearing: .off
        ),
        SeedContact(
            name: "Sam Rivera", peerDID: "did:key:z6MkTestSamRivera0000000000000000000000000000",
            trustTier: 2, lastMessage: "Invite link worked!", timestamp: "Yesterday",
            unreadCount: 0, isOnline: false, muted: true, disappearing: .off
        ),
        SeedContact(
            name: "Riley Chen", peerDID: "did:key:z6MkTestRileyChen0000000000000000000000000000",
            trustTier: 1, lastMessage: "Thanks — talk soon.", timestamp: "Yesterday",
            unreadCount: 0, isOnline: false, muted: false, disappearing: .off
        ),
        SeedContact(
            name: "Aria Rao", peerDID: "did:key:z6MkTestAriaRao000000000000000000000000000000",
            trustTier: 2, lastMessage: "Disappearing on 🔥", timestamp: "8:58 AM",
            unreadCount: 1, isOnline: true, muted: false, disappearing: .h24
        ),
    ]

    private static func seedContacts(localDID: String) async {
        for contact in contactFixtures {
            guard let convo = await ContactThreadHelper.upsertDirectThread(
                peerDID: contact.peerDID,
                displayName: contact.name
            ) else { continue }

            // Enrich the list row with realistic preview / badge state.
            var row = convo
            row.lastMessage = contact.lastMessage
            row.timestamp = contact.timestamp
            row.unreadCount = contact.unreadCount
            row.isOnline = contact.isOnline
            ConversationStore.shared.upsert(row)

            ContactTrustIndex.shared.setTier(contact.trustTier, peerDID: contact.peerDID)

            if contact.muted {
                ConversationPreferencesStore.shared.setMuted(true, for: convo.id)
            }
            if contact.disappearing != .off {
                ConversationPreferencesStore.shared.setDisappearing(contact.disappearing, for: convo.id)
            }

            ConversationThreadStore.replaceStored(
                conversationId: convo.id,
                messages: directThread(localDID: localDID, peerDID: contact.peerDID, name: contact.name)
            )
        }
    }

    /// A short, realistic 1:1 thread including a threaded reply and read receipts.
    private static func directThread(localDID: String, peerDID: String, name: String) -> [StoredThreadMessage] {
        let first = "Hey — did you get the invite link?"
        return [
            StoredThreadMessage(
                id: "seed-\(peerDID)-1", senderDID: peerDID,
                content: first, timestamp: "9:40 AM", deliveryStatus: .read
            ),
            StoredThreadMessage(
                id: "seed-\(peerDID)-2", senderDID: localDID,
                content: "Yes! End-to-end encrypted on this thread.", timestamp: "9:41 AM",
                deliveryStatus: .read
            ),
            StoredThreadMessage(
                id: "seed-\(peerDID)-3", senderDID: peerDID,
                content: "Perfect. I'll send the doc after standup.", timestamp: "9:42 AM",
                deliveryStatus: .read,
                replyToMessageId: "seed-\(peerDID)-2", replyPreview: "Yes! End-to-end encrypted on this thread."
            ),
            StoredThreadMessage(
                id: "seed-\(peerDID)-4", senderDID: localDID,
                content: "Sounds good, thanks \(name.split(separator: " ").first.map(String.init) ?? name)!",
                timestamp: "9:43 AM", deliveryStatus: .delivered
            ),
        ]
    }

    // MARK: - Groups (2)

    private struct SeedGroup {
        let id: String
        let name: String
        let lastMessage: String
        let timestamp: String
        let unreadCount: Int
        let memberDIDs: [(did: String, name: String)]
    }

    private static let groupFixtures: [SeedGroup] = [
        SeedGroup(
            id: "grp-team-echo", name: "Team Echo",
            lastMessage: "Ship checklist updated", timestamp: "Mon", unreadCount: 3,
            memberDIDs: [
                ("did:key:z6MkTestJordanLee00000000000000000000000000000", "Jordan"),
                ("did:key:z6MkTestSamRivera0000000000000000000000000000", "Sam"),
            ]
        ),
        SeedGroup(
            id: "grp-weekend-crew", name: "Weekend crew",
            lastMessage: "Anyone up for a hike?", timestamp: "Sun", unreadCount: 0,
            memberDIDs: [
                ("did:key:z6MkTestRileyChen0000000000000000000000000000", "Riley"),
                ("did:key:z6MkTestAriaRao000000000000000000000000000000", "Aria"),
            ]
        ),
    ]

    private static func seedGroups(localDID: String) {
        for group in groupFixtures {
            let conversationId = "group:\(group.id)"
            ConversationStore.shared.upsert(
                StoredConversation(
                    id: conversationId,
                    contactName: group.name,
                    peerDID: group.id,
                    lastMessage: group.lastMessage,
                    timestamp: group.timestamp,
                    unreadCount: group.unreadCount
                )
            )
            ConversationThreadStore.replaceStored(
                conversationId: conversationId,
                messages: groupThread(localDID: localDID, group: group)
            )
        }
    }

    /// A short multi-party group thread.
    private static func groupThread(localDID: String, group: SeedGroup) -> [StoredThreadMessage] {
        var messages: [StoredThreadMessage] = []
        for (index, member) in group.memberDIDs.enumerated() {
            messages.append(
                StoredThreadMessage(
                    id: "seed-\(group.id)-\(index)", senderDID: member.did,
                    content: index == 0 ? "Welcome to the group 👋" : "Glad to be here!",
                    timestamp: "10:0\(index) AM", deliveryStatus: .read
                )
            )
        }
        messages.append(
            StoredThreadMessage(
                id: "seed-\(group.id)-me", senderDID: localDID,
                content: group.lastMessage, timestamp: "10:05 AM", deliveryStatus: .delivered
            )
        )
        return messages
    }

    // MARK: - Hidden folders (2, each with a hidden chat)

    private struct SeedHiddenChat {
        let folderName: String
        let contactName: String
        let peerDID: String
    }

    private static let hiddenChatFixtures: [SeedHiddenChat] = [
        SeedHiddenChat(
            folderName: "Personal", contactName: "Morgan Diaz",
            peerDID: "did:key:z6MkTestMorganDiaz000000000000000000000000000"
        ),
        SeedHiddenChat(
            folderName: "Work", contactName: "Devi Patel",
            peerDID: "did:key:z6MkTestDeviPatel0000000000000000000000000000"
        ),
    ]

    private static func seedHiddenFolders(localDID: String) async {
        for hidden in hiddenChatFixtures {
            guard let convo = await ContactThreadHelper.upsertDirectThread(
                peerDID: hidden.peerDID,
                displayName: hidden.contactName
            ) else { continue }

            let folder = try? HiddenFolderStore.create(name: hidden.folderName)

            // Flip the hidden flag first: migrates thread storage to encrypted and
            // assigns to the default folder. Then move it into the intended folder.
            ConversationPreferencesStore.shared.setHidden(true, for: convo.id)
            if let folder {
                HiddenFolderStore.assign(conversationId: convo.id, folderId: folder.id)
            }

            // isHidden is now true, so this persists AES-GCM encrypted automatically.
            ConversationThreadStore.replaceStored(
                conversationId: convo.id,
                messages: directThread(localDID: localDID, peerDID: hidden.peerDID, name: hidden.contactName)
            )
        }
    }

    #endif
}
#endif
