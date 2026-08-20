#if os(iOS)
import SwiftUI

struct InviteLinkSheet: View {
    @Environment(\.dismiss) private var dismiss

    @State private var handle = ""
    @State private var shareURL: URL?
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Group {
                if isLoading {
                    ProgressView("Loading invite…")
                } else if let url = shareURL {
                    VStack(spacing: Spacing.lg.rawValue) {
                        Text("Your @username is the invite")
                            .typographyStyle(.h4, color: .echoInk)
                        if !handle.isEmpty {
                            Text(handle)
                                .typographyStyle(.h3, color: .echoSignal)
                        }
                        Text(url.absoluteString)
                            .font(Font.echomono(13))
                            .textSelection(.enabled)
                            .foregroundStyle(Color.echoInk55)
                            .padding(Spacing.md.rawValue)
                            .frame(maxWidth: .infinity)
                            .background(Color.echoSurface)
                            .clipShape(RoundedRectangle(cornerRadius: Spacing.md.rawValue, style: .continuous))
                        ShareLink(item: url) {
                            Label("Share invite", systemImage: "square.and.arrow.up")
                                .frame(maxWidth: .infinity)
                        }
                        .buttonStyle(.borderedProminent)
                    }
                    .padding(Spacing.lg.rawValue)
                } else if let errorMessage {
                    ContentUnavailableView("Could not create invite", systemImage: "exclamationmark.triangle", description: Text(errorMessage))
                } else {
                    ContentUnavailableView("No invite yet", systemImage: "at")
                }
            }
            .navigationTitle("Invite")
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
        let username = await CurrentUserSession.currentUsername()
        handle = InviteHandle.display(username)
        guard let useCase = DIContainer.shared.resolveInviteLinkUseCase() else {
            errorMessage = "Sign in required"
            return
        }
        do {
            shareURL = try await useCase.generateInviteLink()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
