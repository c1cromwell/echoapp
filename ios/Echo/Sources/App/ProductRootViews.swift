#if os(iOS)
import SwiftUI

// MARK: - Echo Comply Companion (read-mostly org compliance posture)

struct ComplyCompanionRootView: View {
    @AppStorage("echo.comply.orgDID") private var orgDID = ""
    @State private var draftOrgDID = ""

    var body: some View {
        NavigationStack {
            Group {
                if orgDID.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    orgSetup
                } else {
                    ComplyDashboardView(orgDID: orgDID)
                }
            }
            .background(Color.Echo.surface)
            .navigationTitle("Echo Comply")
            .toolbar {
                if !orgDID.isEmpty {
                    ToolbarItem(placement: .topBarTrailing) {
                        Button("Switch org") {
                            draftOrgDID = orgDID
                            orgDID = ""
                        }
                        .font(Font.Echo.labelMd)
                    }
                }
            }
        }
        .onAppear {
            draftOrgDID = orgDID
        }
    }

    private var orgSetup: some View {
        VStack(spacing: 20) {
            EchoLogo(size: 48)
            Text("Organization compliance companion")
                .font(Font.Echo.titleLarge)
                .multilineTextAlignment(.center)
            Text("Enter your organization DID to view retention, holds, and audit exports. No message content is shown.")
                .font(Font.Echo.bodyMedium)
                .foregroundStyle(Color.Echo.outline)
                .multilineTextAlignment(.center)

            EchoTextField(
                label: "Organization DID",
                placeholder: "did:key:z6Mk…",
                text: $draftOrgDID
            )

            EchoButton("Continue", style: .primary) {
                let trimmed = draftOrgDID.trimmingCharacters(in: .whitespacesAndNewlines)
                guard trimmed.hasPrefix("did:") else { return }
                orgDID = trimmed
            }
            .disabled(!draftOrgDID.trimmingCharacters(in: .whitespacesAndNewlines).hasPrefix("did:"))
        }
        .padding(24)
    }
}

// MARK: - Echo Passport (standalone verifiable-credential wallet)

struct PassportRootView: View {
    @AppStorage("echo.passport.hasCredential") private var enrolled = false
    @State private var coordinator = EnrollmentCoordinator(
        onComplete: { _ in
            UserDefaults.standard.set(true, forKey: "echo.passport.hasCredential")
        },
        onCancel: {}
    )

    var body: some View {
        NavigationStack {
            passportHome
        }
        .onChange(of: coordinator.stage) { _, stage in
            if case .complete = stage {
                enrolled = true
            }
        }
    }

    private var passportHome: some View {
        ScrollView {
            VStack(spacing: 20) {
                HStack(spacing: 8) {
                    EchoLogo(size: 28)
                    Text("Echo Passport")
                        .font(Font.Echo.headlineSm)
                }
                .frame(maxWidth: .infinity, alignment: .leading)

                GhostBorderSection(title: "YOUR CREDENTIALS") {
                    if enrolled {
                        credentialRow(
                            title: "Echo identity",
                            subtitle: "Verified via OIDC4VC",
                            verified: true
                        )
                    } else {
                        Text("No credentials yet. Add a government ID or mobile wallet credential to unlock Echo trust tiers.")
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.outline)
                    }
                }

                NavigationLink {
                    EnrollmentCoordinatorView(coordinator: coordinator)
                } label: {
                    HStack {
                        Image(systemName: "plus.circle.fill")
                        Text(enrolled ? "Add another credential" : "Add credential")
                            .font(Font.Echo.bodyMedium)
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.system(size: 12))
                    }
                    .foregroundStyle(Color.Echo.onSurface)
                    .padding()
                    .background(Color.Echo.surfaceContainerHigh)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                }
            }
            .padding()
        }
        .background(Color.Echo.surface)
    }

    private func credentialRow(title: String, subtitle: String, verified: Bool) -> some View {
        HStack(spacing: 12) {
            Image(systemName: verified ? "checkmark.seal.fill" : "seal")
                .foregroundStyle(verified ? Color.Echo.success : Color.Echo.outline)
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(Font.Echo.bodyMedium)
                Text(subtitle)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.outline)
            }
            Spacer()
        }
    }
}
#endif
