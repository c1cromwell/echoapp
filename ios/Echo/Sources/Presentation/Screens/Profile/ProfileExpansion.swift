// Presentation/Screens/Profile/ProfileExpansion.swift
// Profile page expansion per iOS Spec v4.2.1 §17
// Adds DID display, wallet summary, QR share, and credential badges

import SwiftUI
#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

// MARK: - Profile DID Section

struct ProfileDIDSection: View {
    let echoHandle: String
    let fullDID: String

    var body: some View {
        VStack(spacing: 4) {
            Text(echoHandle)
                .font(.custom("Inter", size: 13))
                .fontWeight(.semibold)
                .foregroundStyle(Color.Echo.primaryContainer)

            HStack(spacing: 6) {
                Text(didShort)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.outline)
                    .lineLimit(1)
                    .truncationMode(.middle)
                Button {
                    #if canImport(UIKit)
                    UIPasteboard.general.string = fullDID
                    #elseif canImport(AppKit)
                    NSPasteboard.general.setString(fullDID, forType: .string)
                    #endif
                } label: {
                    Image(systemName: "doc.on.doc")
                        .font(.system(size: 10))
                        .foregroundStyle(Color.Echo.outline)
                }
            }
        }
    }

    private var didShort: String {
        guard fullDID.count > 24 else { return fullDID }
        return "\(fullDID.prefix(16))...\(fullDID.suffix(6))"
    }
}

// MARK: - Profile Wallet Summary Card

struct ProfileWalletSummary: View {
    let balance: String
    let onNavigateToWallet: () -> Void

    var body: some View {
        Button(action: onNavigateToWallet) {
            HStack {
                VStack(alignment: .leading) {
                    Text("ECHO BALANCE")
                        .font(Font.Echo.labelSm).tracking(1)
                        .foregroundStyle(Color.Echo.outline)
                    Text("\(balance) ECHO")
                        .font(.custom("Inter", size: 20)).fontWeight(.bold)
                        .foregroundStyle(Color.Echo.onSurface)
                }
                Spacer()
                Text("View Wallet →")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
            .padding(20)
            .background(RoundedRectangle(cornerRadius: 32).fill(Color.Echo.surfaceContainerLow))
            .ghostBorder(opacity: 0.15)
        }
        .buttonStyle(SpringButtonStyle())
        .padding(.horizontal, 20)
    }
}

// MARK: - Profile QR Share Button

struct ProfileQRShareButton: View {
    let onNavigateToQR: () -> Void

    var body: some View {
        Button(action: onNavigateToQR) {
            Label("Share My Identity", systemImage: "qrcode")
                .font(.custom("Inter", size: 14)).fontWeight(.bold)
                .foregroundStyle(Color.Echo.onSurface)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 14)
                .background(RoundedRectangle(cornerRadius: 9999).fill(Color.Echo.surfaceContainerLow))
                .ghostBorder(opacity: 0.15)
        }
        .padding(.horizontal, 20)
    }
}

// MARK: - Profile Credential Badges

struct ProfileCredentialBadges: View {
    let credentials: [ProfileCredentialItem]

    var body: some View {
        GhostBorderSection(title: "VERIFIED CREDENTIALS") {
            ForEach(credentials) { cred in
                CredentialRow(
                    icon: cred.icon,
                    name: cred.name,
                    verified: cred.isVerified,
                    action: cred.isVerified ? nil : "Verify →"
                )
            }
        }
    }
}

struct ProfileCredentialItem: Identifiable {
    let id = UUID()
    let icon: String
    let name: String
    let isVerified: Bool
}

struct CredentialRow: View {
    let icon: String
    let name: String
    let verified: Bool
    let action: String?

    init(icon: String, name: String, verified: Bool, action: String? = nil) {
        self.icon = icon
        self.name = name
        self.verified = verified
        self.action = action
    }

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 16))
                .foregroundStyle(verified ? Color.Echo.success : Color.Echo.outline)
                .frame(width: 24)
            Text(name)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
            Spacer()
            if verified {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundStyle(Color.Echo.success)
            } else if let action = action {
                Text(action)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.primaryContainer)
            }
        }
    }
}
