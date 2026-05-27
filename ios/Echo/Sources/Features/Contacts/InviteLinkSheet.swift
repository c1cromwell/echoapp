#if os(iOS)
import SwiftUI

struct InviteLinkSheet: View {
    @Environment(\.dismiss) private var dismiss

    @State private var invite: InviteLinkResponse?
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Group {
                if isLoading {
                    ProgressView("Creating invite…")
                } else if let invite, let url = invite.shareURL {
                    VStack(spacing: 16) {
                        Text("Share this link")
                            .font(.headline)
                        Text(url.absoluteString)
                            .font(.caption.monospaced())
                            .textSelection(.enabled)
                            .padding()
                            .background(Color.echoSurface)
                            .cornerRadius(8)
                        ShareLink(item: url) {
                            Label("Share invite", systemImage: "square.and.arrow.up")
                                .frame(maxWidth: .infinity)
                        }
                        .buttonStyle(.borderedProminent)
                    }
                    .padding()
                } else if let errorMessage {
                    ContentUnavailableView("Could not create invite", systemImage: "exclamationmark.triangle", description: Text(errorMessage))
                } else {
                    ContentUnavailableView("No invite yet", systemImage: "link")
                }
            }
            .navigationTitle("Invite to ECHO")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
            .task { await loadInvite() }
        }
    }

    private func loadInvite() async {
        isLoading = true
        defer { isLoading = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = "Sign in required"
            return
        }
        do {
            invite = try await ContactSocialAPIClient(apiClient: client).createInviteLink()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
