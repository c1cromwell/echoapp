#if os(iOS)
import SwiftUI

struct GroupChatView: View {
    @Bindable var viewModel: GroupChatViewModel
    @State private var showGroupDetail = false
    @State private var showAttachmentPicker = false

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 12) {
                    ForEach(viewModel.messages) { msg in
                        HStack {
                            if msg.isOutgoing { Spacer(minLength: 40) }
                            if let mediaRef = msg.mediaRef {
                                MediaBubbleView(
                                    mediaRef: mediaRef,
                                    peerDID: "",
                                    isSent: msg.isOutgoing,
                                    timestamp: "Now",
                                    deliveryStatus: nil
                                )
                            } else {
                                Text(msg.text)
                                    .padding(12)
                                    .background(msg.isOutgoing ? Color.accentColor.opacity(0.15) : Color.secondary.opacity(0.12))
                                    .clipShape(RoundedRectangle(cornerRadius: 16))
                            }
                            if !msg.isOutgoing { Spacer(minLength: 40) }
                        }
                    }
                }
                .padding()
            }
            if let error = viewModel.errorMessage {
                Text(error).font(.caption).foregroundStyle(.red).padding(.horizontal)
            }
            if showAttachmentPicker {
                AttachmentPickerView(
                    isPresented: $showAttachmentPicker,
                    onImageSelected: { data, mime in
                        Task { await viewModel.sendGroupMedia(data: data, mimeType: mime, mediaKind: .image) }
                    },
                    onVideoSelected: { data, mime in
                        Task { await viewModel.sendGroupMedia(data: data, mimeType: mime, mediaKind: .video) }
                    },
                    onFileSelected: { data, mime in
                        Task { await viewModel.sendGroupMedia(data: data, mimeType: mime, mediaKind: .file) }
                    },
                    onVoiceNoteTapped: {}
                )
            }
            HStack {
                Button {
                    withAnimation { showAttachmentPicker.toggle() }
                } label: {
                    Image(systemName: showAttachmentPicker ? "xmark.circle.fill" : "plus.circle.fill")
                        .font(.system(size: 24))
                        .foregroundStyle(Color.echoSignal)
                }
                TextField("Message", text: $viewModel.composerText, axis: .vertical)
                    .textFieldStyle(.roundedBorder)
                Button("Send") { Task { await viewModel.sendMessage() } }
                    .disabled(viewModel.composerText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }
            .padding()
        }
        .navigationTitle(viewModel.groupName)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button { showGroupDetail = true } label: {
                    Image(systemName: "person.3")
                }
            }
        }
        .sheet(isPresented: $showGroupDetail) {
            GroupDetailSheet(
                groupId: viewModel.groupId,
                groupName: viewModel.groupName,
                currentUserDID: viewModel.currentUserDID
            )
        }
    }
}
#endif
