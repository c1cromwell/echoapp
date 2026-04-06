// Presentation/Screens/Media/MediaGalleryView.swift
// Shared media gallery with tabbed grid/list layouts

import SwiftUI

// MARK: - Media Gallery View

struct MediaGalleryView: View {
    @StateObject private var viewModel: MediaGalleryViewModel
    @State private var selectedTab: MediaTab = .photos

    enum MediaTab: CaseIterable {
        case photos, videos, files, links
        var label: String {
            switch self {
            case .photos: return "Photos"
            case .videos: return "Videos"
            case .files: return "Files"
            case .links: return "Links"
            }
        }
    }

    init(conversationId: String) {
        _viewModel = StateObject(wrappedValue: MediaGalleryViewModel(conversationId: conversationId))
    }

    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()

            // Tab selector
            HStack(spacing: 0) {
                ForEach(MediaTab.allCases, id: \.self) { tab in
                    Button {
                        withAnimation(.spring(response: 0.3, dampingFraction: 0.85)) {
                            selectedTab = tab
                        }
                    } label: {
                        Text(tab.label)
                            .font(.custom("Inter", size: 14))
                            .fontWeight(selectedTab == tab ? .bold : .medium)
                            .foregroundStyle(selectedTab == tab ? Color.Echo.primaryContainer : Color.Echo.outline)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 12)
                    }
                }
            }
            .background(Color.Echo.surfaceContainerLow)

            // Content
            switch selectedTab {
            case .photos, .videos:
                let columns = Array(repeating: GridItem(.flexible(), spacing: 2), count: 4)
                ScrollView {
                    LazyVGrid(columns: columns, spacing: 2) {
                        ForEach(viewModel.mediaItems(for: selectedTab)) { item in
                            MediaThumbnail(item: item)
                                .aspectRatio(1, contentMode: .fill)
                        }
                    }
                }

            case .files:
                ScrollView {
                    LazyVStack(spacing: 0) {
                        ForEach(viewModel.files) { file in
                            FileRow(file: file)
                        }
                    }
                }

            case .links:
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(viewModel.links) { link in
                            LinkPreviewCard(link: link)
                                .padding(.horizontal, 16)
                        }
                    }
                    .padding(.top, 12)
                }
            }
        }
        .background(Color.Echo.surface)
        .navigationTitle("Shared Media")
        .task { await viewModel.loadMedia() }
    }
}

// MARK: - Media Gallery ViewModel

@MainActor
class MediaGalleryViewModel: ObservableObject {
    let conversationId: String

    @Published var photos: [GalleryMediaItem] = []
    @Published var videos: [GalleryMediaItem] = []
    @Published var files: [GalleryFileItem] = []
    @Published var links: [GalleryLinkItem] = []

    init(conversationId: String) {
        self.conversationId = conversationId
    }

    func loadMedia() async {
        // TODO: Load from media service
    }

    func mediaItems(for tab: MediaGalleryView.MediaTab) -> [GalleryMediaItem] {
        switch tab {
        case .photos: return photos
        case .videos: return videos
        default: return []
        }
    }
}

// MARK: - Gallery Models

struct GalleryMediaItem: Identifiable {
    let id: String
    let thumbnailURL: URL?
    let fullURL: URL?
    let isVideo: Bool
    let duration: TimeInterval?
    let timestamp: Date
}

struct GalleryFileItem: Identifiable {
    let id: String
    let name: String
    let size: String
    let icon: String
    let timestamp: Date
}

struct GalleryLinkItem: Identifiable {
    let id: String
    let url: URL
    let title: String?
    let description: String?
    let imageURL: URL?
    let domain: String
}

// MARK: - Media Thumbnail

struct MediaThumbnail: View {
    let item: GalleryMediaItem

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 4)
                .fill(Color.Echo.surfaceContainerHigh)

            if item.isVideo {
                Image(systemName: "play.circle.fill")
                    .font(.system(size: 24))
                    .foregroundStyle(.white.opacity(0.8))
            }
        }
    }
}

// MARK: - File Row

struct FileRow: View {
    let file: GalleryFileItem

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: file.icon)
                .font(.system(size: 20))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 40, height: 40)
                .background(
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color.Echo.surfaceContainerLow)
                )

            VStack(alignment: .leading, spacing: 2) {
                Text(file.name)
                    .font(Font.Echo.bodyMedium)
                    .fontWeight(.medium)
                    .foregroundStyle(Color.Echo.onSurface)
                    .lineLimit(1)
                Text(file.size)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.outline)
            }

            Spacer()

            Text(file.timestamp.formatted(date: .abbreviated, time: .omitted))
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.outline)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }
}

// MARK: - Link Preview Card

struct LinkPreviewCard: View {
    let link: GalleryLinkItem

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Image placeholder
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.Echo.surfaceContainerHigh)
                .frame(height: 120)

            VStack(alignment: .leading, spacing: 4) {
                if let title = link.title {
                    Text(title)
                        .font(Font.Echo.bodyMedium)
                        .fontWeight(.semibold)
                        .foregroundStyle(Color.Echo.onSurface)
                        .lineLimit(2)
                }
                Text(link.domain)
                    .font(Font.Echo.labelMd)
                    .foregroundStyle(Color.Echo.outline)
            }
            .padding(.horizontal, 12)
            .padding(.bottom, 12)
        }
        .background(
            RoundedRectangle(cornerRadius: 20)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .ghostBorder(opacity: 0.15)
    }
}
