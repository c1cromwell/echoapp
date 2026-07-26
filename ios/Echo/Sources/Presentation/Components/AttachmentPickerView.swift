#if os(iOS)
import SwiftUI
import PhotosUI

struct AttachmentPickerView: View {
    @Binding var isPresented: Bool
    var onImageSelected: (Data, String) -> Void
    var onVideoSelected: (Data, String) -> Void
    var onFileSelected: (Data, String) -> Void
    var onCloudTapped: (() -> Void)? = nil
    var onVoiceNoteTapped: () -> Void
    var onPollTapped: (() -> Void)? = nil
    var onPaymentTapped: (() -> Void)? = nil
    var onGifTapped: (() -> Void)? = nil

    @State private var selectedPhotoItem: PhotosPickerItem?
    @State private var showPhotoPicker = false
    @State private var showFilePicker = false

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 24) {
                attachmentButton(icon: "photo.on.rectangle", label: "Photo") {
                    showPhotoPicker = true
                }
                attachmentButton(icon: "mic.fill", label: "Voice") {
                    isPresented = false
                    onVoiceNoteTapped()
                }
                attachmentButton(icon: "doc.fill", label: "File") {
                    showFilePicker = true
                }
                if let onCloudTapped {
                    attachmentButton(icon: "icloud.fill", label: "Cloud") {
                        isPresented = false
                        onCloudTapped()
                    }
                }
                if let onPollTapped {
                    attachmentButton(icon: "chart.bar", label: "Poll") {
                        isPresented = false
                        onPollTapped()
                    }
                }
                if let onGifTapped {
                    attachmentButton(icon: "face.smiling", label: "GIF") {
                        isPresented = false
                        onGifTapped()
                    }
                }
                if let onPaymentTapped {
                    attachmentButton(icon: "dollarsign.circle.fill", label: "Pay") {
                        isPresented = false
                        onPaymentTapped()
                    }
                }
            }
            .padding(.vertical, 16)
            .padding(.horizontal, 20)
        }
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .photosPicker(
            isPresented: $showPhotoPicker,
            selection: $selectedPhotoItem,
            matching: .any(of: [.images, .videos])
        )
        .onChange(of: selectedPhotoItem) { _, item in
            guard let item else { return }
            Task { await processPhotoItem(item) }
            selectedPhotoItem = nil
            isPresented = false
        }
        .fileImporter(
            isPresented: $showFilePicker,
            allowedContentTypes: [.data],
            allowsMultipleSelection: false
        ) { result in
            guard case .success(let urls) = result, let url = urls.first else { return }
            processFileURL(url)
            isPresented = false
        }
    }

    private func attachmentButton(icon: String, label: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            VStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 24))
                    .foregroundStyle(Color.echoSignal)
                    .frame(width: 52, height: 52)
                    .background(Color.echoSignal.opacity(0.12))
                    .clipShape(Circle())
                Text(label)
                    .font(.system(size: 12, weight: .medium))
                    .foregroundStyle(Color.echoInk70)
            }
        }
        .buttonStyle(.plain)
    }

    private func processPhotoItem(_ item: PhotosPickerItem) async {
        if let data = try? await item.loadTransferable(type: Data.self) {
            let contentType = item.supportedContentTypes.first
            if let contentType, contentType.conforms(to: .movie) {
                onVideoSelected(data, contentType.preferredMIMEType ?? "video/mp4")
            } else {
                onImageSelected(data, contentType?.preferredMIMEType ?? "image/jpeg")
            }
        }
    }

    private func processFileURL(_ url: URL) {
        guard url.startAccessingSecurityScopedResource() else { return }
        defer { url.stopAccessingSecurityScopedResource() }
        guard let data = try? Data(contentsOf: url) else { return }
        let mime = url.mimeType ?? "application/octet-stream"
        onFileSelected(data, mime)
    }
}

private extension URL {
    var mimeType: String? {
        switch pathExtension.lowercased() {
        case "pdf": return "application/pdf"
        case "doc", "docx": return "application/msword"
        case "txt": return "text/plain"
        case "zip": return "application/zip"
        case "png": return "image/png"
        case "jpg", "jpeg": return "image/jpeg"
        case "mp4", "m4v": return "video/mp4"
        case "mp3": return "audio/mpeg"
        case "m4a": return "audio/mp4"
        default: return nil
        }
    }
}
#endif
