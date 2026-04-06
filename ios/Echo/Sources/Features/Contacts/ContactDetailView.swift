// Features/Contacts/ContactDetailView.swift
// Contact detail screen with trust info, credentials, shared media, and privacy settings

import SwiftUI
#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

struct ContactDetailView: View {
    @StateObject private var viewModel: ContactDetailViewModel
    @Environment(\.dismiss) private var dismiss

    init(contactId: String) {
        _viewModel = StateObject(wrappedValue: ContactDetailViewModel(contactId: contactId))
    }

    var body: some View {
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
                        .font(.custom("Inter", size: 28))
                        .fontWeight(.heavy)
                        .tracking(-0.5)

                    Text(viewModel.contact.echoHandle)
                        .font(.custom("Inter", size: 13))
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
                        // Navigate to chat
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
                    .font(.custom("Inter", size: 14)).fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.error)

                    Button("Report Contact") {
                        viewModel.showReportSheet = true
                    }
                    .font(.custom("Inter", size: 14)).fontWeight(.semibold)
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
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .foregroundStyle(Color.Echo.outline)
                }
            }
        }
        .task { await viewModel.loadContact() }
    }
}

// MARK: - Ghost Border Section

struct GhostBorderSection<Content: View>: View {
    let title: String
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(title)
                .font(.custom("Inter", size: 10))
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
                    .font(.custom("Inter", size: 10))
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
                    .font(.custom("Inter", size: 10))
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
