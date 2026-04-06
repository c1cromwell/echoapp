// Features/Enterprise/EnterpriseProfileView.swift
// Enterprise profile with verified organization credentials and authorized representatives

import SwiftUI

struct EnterpriseProfileView: View {
    let organizationId: String
    @StateObject private var viewModel = EnterpriseProfileViewModel()

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Organization header
                VStack(spacing: 12) {
                    // Logo placeholder
                    RoundedRectangle(cornerRadius: 24)
                        .fill(Color.Echo.surfaceContainerHigh)
                        .frame(width: 80, height: 80)
                        .overlay(
                            Image(systemName: "building.2.fill")
                                .font(.system(size: 32))
                                .foregroundStyle(Color.Echo.primaryContainer)
                        )

                    Text(viewModel.orgName)
                        .font(.custom("Inter", size: 24))
                        .fontWeight(.heavy)
                        .tracking(-0.5)

                    if viewModel.isVerified {
                        HStack(spacing: 6) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(Color.Echo.success)
                            Text("Verified Organization")
                                .font(Font.Echo.labelMd)
                                .foregroundStyle(Color.Echo.success)
                        }
                    }
                }
                .padding(.top, 16)

                // Verified Credentials
                GhostBorderSection(title: "VERIFIED CREDENTIALS") {
                    VerificationRow(name: "Business Registration", verified: viewModel.hasBusinessReg)
                    VerificationRow(name: "Domain Verification", verified: viewModel.hasDomainVerification)
                    VerificationRow(name: "EV TLS Certificate", verified: viewModel.hasEVTLS)
                    VerificationRow(name: "KYB Compliance", verified: viewModel.hasKYB)
                }

                // Official Channels
                GhostBorderSection(title: "OFFICIAL CHANNELS") {
                    ForEach(viewModel.channels, id: \.name) { channel in
                        HStack(spacing: 12) {
                            Image(systemName: "megaphone.fill")
                                .foregroundStyle(Color.Echo.primaryContainer)
                                .frame(width: 24)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(channel.name)
                                    .font(Font.Echo.bodyMedium)
                                    .fontWeight(.medium)
                                    .foregroundStyle(Color.Echo.onSurface)
                                Text("\(channel.subscribers) subscribers")
                                    .font(Font.Echo.labelMd)
                                    .foregroundStyle(Color.Echo.outline)
                            }
                            Spacer()
                            Image(systemName: "chevron.right")
                                .font(.system(size: 12))
                                .foregroundStyle(Color.Echo.outline)
                        }
                    }
                }

                // Authorized Representatives
                GhostBorderSection(title: "AUTHORIZED REPRESENTATIVES") {
                    ForEach(viewModel.representatives, id: \.name) { rep in
                        HStack(spacing: 12) {
                            Circle()
                                .fill(Color.Echo.surfaceContainerHigh)
                                .frame(width: 36, height: 36)
                                .overlay(
                                    Image(systemName: "person.fill")
                                        .font(.system(size: 14))
                                        .foregroundStyle(Color.Echo.outline)
                                )
                            VStack(alignment: .leading, spacing: 2) {
                                Text(rep.name)
                                    .font(Font.Echo.bodyMedium)
                                    .fontWeight(.medium)
                                    .foregroundStyle(Color.Echo.onSurface)
                                Text(rep.role)
                                    .font(Font.Echo.labelMd)
                                    .foregroundStyle(Color.Echo.outline)
                            }
                            Spacer()
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundStyle(Color.Echo.success)
                        }
                    }
                }

                // DID Info
                GhostBorderSection(title: "ORGANIZATION IDENTITY") {
                    InfoRow(label: "DID", value: viewModel.didShort)
                    InfoRow(label: "Member Since", value: viewModel.memberSince)
                    InfoRow(label: "Trust Score", value: "\(viewModel.trustScore)/100")
                }
            }
            .padding(.bottom, 40)
        }
        .background(Color.Echo.surface)
        .overlay(alignment: .top) { SecureThreadIndicator() }
        .navigationTitle("Enterprise")
        .task { await viewModel.loadOrganization(id: organizationId) }
    }
}

// MARK: - Verification Row

struct VerificationRow: View {
    let name: String
    let verified: Bool

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: verified ? "checkmark.circle.fill" : "circle")
                .foregroundStyle(verified ? Color.Echo.success : Color.Echo.outline)
            Text(name)
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.onSurface)
            Spacer()
            if verified {
                Text("Verified")
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.success)
            }
        }
    }
}

// MARK: - Enterprise ViewModel

struct OrgChannel {
    let name: String
    let subscribers: Int
}

struct OrgRepresentative {
    let name: String
    let role: String
}

@MainActor
class EnterpriseProfileViewModel: ObservableObject {
    @Published var orgName = ""
    @Published var isVerified = false
    @Published var hasBusinessReg = false
    @Published var hasDomainVerification = false
    @Published var hasEVTLS = false
    @Published var hasKYB = false
    @Published var channels: [OrgChannel] = []
    @Published var representatives: [OrgRepresentative] = []
    @Published var did = ""
    @Published var memberSince = ""
    @Published var trustScore = 0

    var didShort: String {
        guard did.count > 20 else { return did }
        return "\(did.prefix(12))...\(did.suffix(6))"
    }

    func loadOrganization(id: String) async {
        // TODO: Load from enterprise service
    }
}
