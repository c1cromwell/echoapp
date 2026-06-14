// Features/Contacts/ContactDetailView.swift
// Contact detail screen with trust info, credentials, shared media, and privacy settings

import SwiftUI
#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

public struct ContactDetailView: View {
    @StateObject private var viewModel: ContactDetailViewModel
    @Environment(\.dismiss) private var dismiss
    var onMessage: (() -> Void)?

    public init(contactId: String, displayName: String? = nil, onMessage: (() -> Void)? = nil) {
        self.onMessage = onMessage
        _viewModel = StateObject(wrappedValue: ContactDetailViewModel(contactId: contactId, displayName: displayName))
    }

    public var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Hero — large avatar with trust ring
                VStack(spacing: 12) {
                    TrustRingAvatar(
                        imageURL: viewModel.contact.avatarURL,
                        trustTier: viewModel.contact.trustTier,
                        size: 140
                    )

                    Text(viewModel.contact.name)
                        .font(.system(size: 28))
                        .fontWeight(.heavy)
                        .tracking(-0.5)

                    Text(viewModel.contact.echoHandle)
                        .font(.system(size: 13))
                        .fontWeight(.semibold)
                        .foregroundStyle(Color.Echo.primaryContainer)

                    HStack(spacing: 8) {
                        TrustTierPill(tier: viewModel.contact.trustTier)

                        if viewModel.contact.isOnline {
                            HStack(spacing: 4) {
                                Circle().fill(Color.Echo.success).frame(width: 6, height: 6)
                                Text("Online").font(Font.Echo.labelMd).foregroundStyle(Color.Echo.success)
                            }
                        }
                    }
                }
                .padding(.top, 16)

                // Action buttons — 4-column
                HStack(spacing: 12) {
                    ContactActionButton(icon: "message.fill", label: "Message") {
                        dismiss()
                        onMessage?()
                    }
                    ContactActionButton(icon: "phone.fill", label: "Voice") {
                        viewModel.showVoiceCall = true
                    }
                    ContactActionButton(icon: "video.fill", label: "Video") {
                        viewModel.showVideoCall = true
                    }
                    ContactActionButton(icon: "magnifyingglass", label: "Search") {
                        viewModel.showSearch = true
                    }
                }
                .padding(.horizontal, 20)

                // Trust & Identity card
                GhostBorderSection(title: "TRUST & IDENTITY") {
                    TrustRow(label: "Trust Score", value: "\(viewModel.contact.trustScore)/100")
                    TrustRow(label: "DID", value: viewModel.contact.didShort, copyable: true)
                    TrustRow(label: "Verified Since", value: viewModel.contact.verifiedDate)
                    TrustRow(label: "Mutual Groups", value: "\(viewModel.contact.mutualGroups)")
                    TrustRow(label: "Mutual Contacts", value: "\(viewModel.contact.mutualContacts)")
                }

                if !viewModel.mutualGroups.isEmpty {
                    GhostBorderSection(title: "GROUPS IN COMMON") {
                        ForEach(viewModel.mutualGroups) { group in
                            HStack(spacing: 12) {
                                Image(systemName: "person.3.fill")
                                    .foregroundStyle(Color.Echo.primaryContainer)
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(group.name ?? "Group")
                                        .font(Font.Echo.bodyMedium)
                                    if let count = group.member_count {
                                        Text("\(count) members")
                                            .font(Font.Echo.labelMd)
                                            .foregroundStyle(Color.Echo.outline)
                                    }
                                }
                                Spacer()
                            }
                        }
                    }
                } else if viewModel.relationshipError == nil {
                    GhostBorderSection(title: "GROUPS IN COMMON") {
                        Text("No shared groups yet")
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.outline)
                    }
                }

                if !viewModel.mutualContacts.isEmpty {
                    GhostBorderSection(title: "MUTUAL CONTACTS") {
                        ForEach(viewModel.mutualContacts) { person in
                            HStack(spacing: 12) {
                                Image(systemName: "person.circle.fill")
                                    .foregroundStyle(Color.Echo.primaryContainer)
                                VStack(alignment: .leading, spacing: 2) {
                                    if let username = person.username, !username.isEmpty {
                                        Text("@\(username)")
                                            .font(Font.Echo.bodyMedium)
                                    }
                                    Text(ContactThreadHelper.truncatedDID(person.did))
                                        .font(Font.Echo.labelMd)
                                        .foregroundStyle(Color.Echo.outline)
                                }
                                Spacer()
                            }
                        }
                    }
                }

                // Credentials card
                GhostBorderSection(title: "CREDENTIALS") {
                    ForEach(viewModel.contact.credentials) { cred in
                        HStack(spacing: 12) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(Color.Echo.success)
                            Text(cred.name)
                                .font(Font.Echo.bodyMedium)
                            Spacer()
                        }
                    }
                }

                // Shared Media preview (horizontal scroll)
                SharedMediaPreview(
                    media: viewModel.sharedMedia,
                    onSeeAll: { viewModel.showMediaGallery = true }
                )

                // Privacy settings for this contact
                GhostBorderSection(title: "PRIVACY FOR THIS CONTACT") {
                    SettingsRow(icon: "bell.fill", label: "Custom Notifications",
                                value: viewModel.notificationsEnabled ? "On" : "Off")
                    SettingsRow(icon: "timer", label: "Disappearing Messages",
                                value: viewModel.disappearingEnabled ? "On" : "Off")

                    Spacer().frame(height: 8)

                    Button("Block Contact") {
                        viewModel.showBlockConfirmation = true
                    }
                    .font(.system(size: 14)).fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.error)

                    Button("Report Contact") {
                        viewModel.showReportSheet = true
                    }
                    .font(.system(size: 14)).fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.error.opacity(0.7))
                }
            }
            .padding(.bottom, 40)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationBarBackButtonHidden()
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button { dismiss() } label: {
                    Image(systemName: "arrow.left")
                        .foregroundStyle(Color.Echo.onSurface)
                }
            }
            ToolbarItem(placement: .primaryAction) {
                Menu {
                    Button("Copy DID") { viewModel.copyDID() }
                    Button("Share Contact") { viewModel.shareContact() }
                    Button(
                        ContactFavoritesStore.isFavorite(did: viewModel.contactId)
                            ? "Remove from favorites" : "Add to favorites"
                    ) {
                        ContactFavoritesStore.toggle(did: viewModel.contactId)
                    }
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .foregroundStyle(Color.Echo.outline)
                }
            }
        }
        .task { await viewModel.loadContact() }
        #if os(iOS)
        .fullScreenCover(isPresented: $viewModel.showVoiceCall) {
            CallView(
                peerDID: viewModel.contactId,
                callType: .voice,
                contactName: viewModel.contact.name
            )
        }
        .fullScreenCover(isPresented: $viewModel.showVideoCall) {
            CallView(
                peerDID: viewModel.contactId,
                callType: .video,
                contactName: viewModel.contact.name
            )
        }
        #endif
        .alert("Block contact?", isPresented: $viewModel.showBlockConfirmation) {
            Button("Cancel", role: .cancel) {}
            Button("Block", role: .destructive) {
                Task { await viewModel.blockContact() }
            }
        } message: {
            Text("They won't be able to message you or see your online status.")
        }
        .alert("Contact blocked", isPresented: $viewModel.isBlocked) {
            Button("OK") { dismiss() }
        }
    }
}

// MARK: - Ghost Border Section

struct GhostBorderSection<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(title)
                .font(.system(size: 10))
                .fontWeight(.bold)
                .tracking(2)
                .foregroundStyle(Color.Echo.outline)
                .padding(.leading, 8)

            VStack(alignment: .leading, spacing: 12) {
                content()
            }
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 32)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder()
        }
        .padding(.horizontal, 20)
    }
}

// MARK: - Contact Action Button

struct ContactActionButton: View {
    let icon: String
    let label: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 20))
                    .foregroundStyle(Color.Echo.primaryContainer)
                Text(label)
                    .font(.system(size: 10))
                    .fontWeight(.bold)
                    .textCase(.uppercase)
                    .tracking(0.5)
                    .foregroundStyle(Color.Echo.outline)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 64)
            .background(
                RoundedRectangle(cornerRadius: 20)
                    .fill(Color.Echo.surfaceContainerLow)
            )
            .ghostBorder(opacity: 0.15)
        }
        .buttonStyle(SpringButtonStyle())
    }
}

// MARK: - Trust Row

struct TrustRow: View {
    let label: String
    let value: String
    var copyable: Bool = false

    var body: some View {
        HStack {
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
            Spacer()
            HStack(spacing: 6) {
                Text(value)
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.onSurface)
                    .lineLimit(1)
                if copyable {
                    Button {
                        #if canImport(UIKit)
                        UIPasteboard.general.string = value
                        #elseif canImport(AppKit)
                        NSPasteboard.general.setString(value, forType: .string)
                        #endif
                    } label: {
                        Image(systemName: "doc.on.doc")
                            .font(.system(size: 10))
                            .foregroundStyle(Color.Echo.outline)
                    }
                }
            }
        }
    }
}

// MARK: - Settings Row

struct SettingsRow: View {
    let icon: String
    let label: String
    let value: String

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 14))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 24)
            Text(label)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
            Spacer()
            Text(value)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
        }
    }
}

// MARK: - Shared Media Preview

struct SharedMediaPreview: View {
    let media: [SharedMediaItem]
    let onSeeAll: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("SHARED MEDIA")
                    .font(.system(size: 10))
                    .fontWeight(.bold)
                    .tracking(2)
                    .foregroundStyle(Color.Echo.outline)
                Spacer()
                Button("See All") { onSeeAll() }
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(.horizontal, 28)

            if media.isEmpty {
                Text("No shared media yet")
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.outline)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 24)
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 4) {
                        ForEach(media.prefix(10)) { item in
                            RoundedRectangle(cornerRadius: 8)
                                .fill(Color.Echo.surfaceContainerHigh)
                                .frame(width: 80, height: 80)
                                .overlay(
                                    Image(systemName: item.type == .video ? "play.circle.fill" : "photo")
                                        .foregroundStyle(Color.Echo.outline)
                                )
                        }
                    }
                    .padding(.horizontal, 20)
                }
            }
        }
    }
}

// MARK: - Success Color Extension

extension Color.Echo {
    static let success = Color(hex: "#16A34A")
    static let warning = Color(hex: "#F59E0B")
}
