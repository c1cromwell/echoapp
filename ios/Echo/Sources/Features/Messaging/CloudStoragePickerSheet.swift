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
    @State private var connected: [CloudStorageIntegrationManager.Provider] = []
    @State private var browsingProvider: CloudStorageIntegrationManager.Provider?
    @State private var remoteFiles: [CloudStorageIntegrationManager.RemoteFile] = []

    var body: some View {
        NavigationStack {
            List {
                Section {
                    Text("Files stream through the backend proxy, then encrypt before upload.")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                Section("Providers") {
                    ForEach(CloudStorageIntegrationManager.Provider.allCases, id: \.self) { provider in
                        HStack {
                            VStack(alignment: .leading) {
                                Text(provider.title)
                                if connected.contains(provider) {
                                    Text("Connected").font(.caption).foregroundStyle(.green)
                                }
                            }
                            Spacer()
                            if connected.contains(provider) {
                                Button("Browse") { Task { await browse(provider: provider) } }
                            } else {
                                Button("Connect") { connect(provider: provider) }
                            }
                        }
                    }
                }
                if let browsingProvider, !remoteFiles.isEmpty {
                    Section("Files — \(browsingProvider.title)") {
                        ForEach(remoteFiles) { file in
                            Button {
                                Task { await pickFile(provider: browsingProvider, file: file) }
                            } label: {
                                VStack(alignment: .leading) {
                                    Text(file.name)
                                    Text(ByteCountFormatter.string(fromByteCount: file.size, countStyle: .file))
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
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
            .task { connected = await CloudStorageIntegrationManager.connectedProviders() }
        }
    }

    private func connect(provider: CloudStorageIntegrationManager.Provider) {
        Task {
            isLoading = true
            defer { isLoading = false }
            do {
                let url = try await CloudStorageIntegrationManager.fetchAuthorizeURL(provider: provider)
                await MainActor.run {
                    let session = ASWebAuthenticationSession(
                        url: url,
                        callbackURLScheme: "echo"
                    ) { callback, error in
                        if let error {
                            message = error.localizedDescription
                            return
                        }
                        guard let callback,
                              let code = URLComponents(url: callback, resolvingAgainstBaseURL: false)?
                                .queryItems?
                                .first(where: { $0.name == "code" })?
                                .value else {
                            message = "OAuth callback missing authorization code."
                            return
                        }
                        Task {
                            do {
                                try await CloudStorageIntegrationManager.exchangeCode(code, provider: provider)
                                connected = await CloudStorageIntegrationManager.connectedProviders()
                                message = "\(provider.title) connected."
                            } catch {
                                message = error.localizedDescription
                            }
                        }
                    }
                    session.presentationContextProvider = CloudAuthContext.shared
                    session.start()
                }
            } catch {
                message = error.localizedDescription
            }
        }
    }

    private func browse(provider: CloudStorageIntegrationManager.Provider) async {
        isLoading = true
        defer { isLoading = false }
        do {
            remoteFiles = try await CloudStorageIntegrationManager.listFiles(provider: provider)
            browsingProvider = provider
            if remoteFiles.isEmpty {
                message = "No files found in \(provider.title)."
            }
        } catch {
            message = error.localizedDescription
        }
    }

    private func pickFile(
        provider: CloudStorageIntegrationManager.Provider,
        file: CloudStorageIntegrationManager.RemoteFile
    ) async {
        isLoading = true
        defer { isLoading = false }
        do {
            let (data, mime) = try await CloudStorageIntegrationManager.downloadFile(
                provider: provider,
                fileID: file.id
            )
            let resolvedMime = file.mimeType.isEmpty ? mime : file.mimeType
            onFileData(data, resolvedMime)
            dismiss()
        } catch {
            message = error.localizedDescription
        }
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
