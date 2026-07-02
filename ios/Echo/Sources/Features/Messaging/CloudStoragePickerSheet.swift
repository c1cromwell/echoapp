#if os(iOS)
import AuthenticationServices
import SwiftUI
import UIKit

/// Connect cloud providers and pick a remote file for encrypted upload (WO-46).
struct CloudStoragePickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    let onFileData: (Data, String) -> Void

    @State private var message: String?
    @State private var isLoading = false

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text("Files stream through encryption — nothing is stored unencrypted on device.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                Section("Providers") {
                    ForEach(CloudStorageIntegrationManager.Provider.allCases, id: \.self) { provider in
                        HStack {
                            VStack(alignment: .leading) {
                                Text(provider.title)
                                if CloudStorageIntegrationManager.connectedProviders().contains(provider) {
                                    Text("Connected").font(.caption).foregroundStyle(.green)
                                }
                            }
                            Spacer()
                            if CloudStorageIntegrationManager.connectedProviders().contains(provider) {
                                Button("Pick file") { Task { await pickStub(provider: provider) } }
                            } else {
                                Button("Connect") { connect(provider: provider) }
                            }
                        }
                    }
                }
                if let message {
                    Section {
                        Text(message).font(.footnote).foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Cloud files")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
            .overlay {
                if isLoading { ProgressView().controlSize(.large) }
            }
        }
    }

    private func connect(provider: CloudStorageIntegrationManager.Provider) {
        let url = CloudStorageIntegrationManager.authorizationURL(provider: provider)
        let session = ASWebAuthenticationSession(
            url: url,
            callbackURLScheme: "echo"
        ) { callback, error in
            if let error {
                message = error.localizedDescription
                return
            }
            guard let callback,
                  let token = URLComponents(url: callback, resolvingAgainstBaseURL: false)?
                    .queryItems?
                    .first(where: { $0.name == "access_token" })?
                    .value else {
                message = "OAuth completed — paste token via Settings if redirect lacks access_token."
                return
            }
            Task {
                try? await CloudStorageIntegrationManager.saveToken(token, provider: provider)
                message = "\(provider.title) connected."
            }
        }
        session.presentationContextProvider = CloudAuthContext.shared
        session.start()
    }

    private func pickStub(provider: CloudStorageIntegrationManager.Provider) async {
        isLoading = true
        defer { isLoading = false }
        // MVP: cloud picker streams via backend proxy in a later WO; demo encrypted payload path.
        let sample = Data("echo-cloud-file-placeholder".utf8)
        onFileData(sample, "application/octet-stream")
        message = "Selected file from \(provider.title) (stub stream)."
        dismiss()
    }
}

private final class CloudAuthContext: NSObject, ASWebAuthenticationPresentationContextProviding {
    static let shared = CloudAuthContext()
    func presentationAnchor(for session: ASWebAuthenticationSession) -> ASPresentationAnchor {
        UIApplication.shared.connectedScenes
            .compactMap { ($0 as? UIWindowScene)?.keyWindow }
            .first ?? ASPresentationAnchor()
    }
}
#endif
