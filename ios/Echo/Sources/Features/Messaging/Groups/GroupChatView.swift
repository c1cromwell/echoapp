#if os(iOS)
import SwiftUI

struct GroupChatView: View {
    @Bindable var viewModel: GroupChatViewModel
    @State private var showGroupDetail = false
    @State private var showAttachmentPicker = false
    @StateObject private var voiceRecorder = VoiceNoteRecorder()

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
            if voiceRecorder.isRecording {
                groupVoiceRecordingBar
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
                    onVoiceNoteTapped: {
                        try? voiceRecorder.startRecording()
                    }
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

    private var groupVoiceRecordingBar: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(Color.red)
                .frame(width: 10, height: 10)

            WaveformView(
                samples: voiceRecorder.waveformSamples,
                progress: 1.0,
                accentColor: .echoSignal
            )
            .frame(height: 28)

            Text(voiceRecorderElapsed)
                .font(.system(size: 14, weight: .medium).monospacedDigit())
                .foregroundStyle(Color.echoInk)

            Spacer()

            Button {
                voiceRecorder.cancelRecording()
            } label: {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 22))
                    .foregroundStyle(Color.echoInk40)
            }
            .buttonStyle(.plain)

            Button {
                Task { await viewModel.sendVoiceNote(from: voiceRecorder) }
            } label: {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.system(size: 28))
                    .foregroundStyle(Color.echoSignal)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, Spacing.md.rawValue)
        .padding(.vertical, 10)
        .background(Color.echoPaperDim)
    }

    private var voiceRecorderElapsed: String {
        let total = Int(voiceRecorder.elapsed)
        return String(format: "%d:%02d", total / 60, total % 60)
    }
}
#endif
