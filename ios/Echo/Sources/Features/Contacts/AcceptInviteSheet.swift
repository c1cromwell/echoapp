#if os(iOS)
import SwiftUI

/// Accept `echo://invite?code=` contact invites (WO-222).
struct AcceptInviteSheet: View {
    @Environment(\.dismiss) private var dismiss

    let inviteCode: String
    var onAccepted: (() -> Void)?

    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var success = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                if success {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(.green)
                    Text("Invite accepted")
                        .font(.headline)
                    Text("They've been added to your contacts.")
                        .multilineTextAlignment(.center)
                        .foregroundStyle(.secondary)
                } else {
                    Text("Accept contact invite?")
                        .font(.headline)
                    Text("Code: \(inviteCode)")
                        .font(.caption.monospaced())
                        .foregroundStyle(.secondary)
                    if let errorMessage {
                        Text(errorMessage)
                            .font(.footnote)
                            .foregroundStyle(.red)
                    }
                    Button {
                        Task { await accept() }
                    } label: {
                        if isLoading {
                            ProgressView()
                        } else {
                            Text("Accept invite")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(isLoading)
                }
            }
            .padding()
            .navigationTitle("Contact invite")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(success ? "Done" : "Cancel") {
                        if success { onAccepted?() }
                        dismiss()
                    }
                }
            }
        }
    }

    private func accept() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        guard let client = DIContainer.shared.resolveAPIClient() else {
            errorMessage = "Sign in required"
            return
        }
        let social = ContactSocialAPIClient(apiClient: client)
        do {
            _ = try await social.acceptInvite(code: inviteCode)
            success = true
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
#endif
