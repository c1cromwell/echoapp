// Presentation/Screens/Search/SearchView.swift
// Advanced message search with filters and recent searches

import SwiftUI

// MARK: - Search View

struct SearchView: View {
    @StateObject private var viewModel = SearchViewModel()
    @FocusState private var isSearchFocused: Bool
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 0) {
            SecureThreadIndicator()

            // Search bar
            HStack(spacing: 12) {
                Button("Cancel") { dismiss() }
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.primaryContainer)

                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundStyle(Color.Echo.outline)
                    TextField("Search messages, files, links...", text: $viewModel.query)
                        .font(Font.Echo.bodyLarge)
                        .focused($isSearchFocused)
                    if !viewModel.query.isEmpty {
                        Button { viewModel.query = "" } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(Color.Echo.outline)
                        }
                    }
                }
                .padding(12)
                .background(
                    RoundedRectangle(cornerRadius: 16)
                        .fill(Color.Echo.surfaceContainerLow)
                )
                .ghostBorder(opacity: 0.15)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)

            // Filter tabs
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(SearchFilter.allCases, id: \.self) { filter in
                        FilterChip(
                            label: filter.displayName,
                            isSelected: viewModel.activeFilter == filter,
                            action: { viewModel.activeFilter = filter }
                        )
                    }
                }
                .padding(.horizontal, 16)
            }
            .padding(.vertical, 8)

            // Results / Recent searches
            ScrollView {
                if viewModel.query.isEmpty {
                    RecentSearchesSection(
                        searches: viewModel.recentSearches,
                        onTap: { viewModel.query = $0 },
                        onClear: { viewModel.clearRecentSearches() }
                    )
                } else {
                    LazyVStack(spacing: 0) {
                        ForEach(viewModel.results) { result in
                            SearchResultRow(result: result)
                        }
                    }

                    if viewModel.results.isEmpty && !viewModel.isSearching {
                        EmptySearchState()
                    }
                }
            }
        }
        .background(Color.Echo.surface)
        .onAppear { isSearchFocused = true }
    }
}

// MARK: - Search Filter

enum SearchFilter: String, CaseIterable {
    case all, files, photos, links, voice

    var displayName: String {
        switch self {
        case .all: return "All"
        case .files: return "Files"
        case .photos: return "Photos"
        case .links: return "Links"
        case .voice: return "Voice"
        }
    }
}

// MARK: - Search Result

struct SearchResult: Identifiable {
    let id: String
    let conversationId: String
    let contactName: String
    let contactAvatar: URL?
    let matchedText: String
    let timestamp: Date
    let messageType: String
    let attachmentName: String?
    let attachmentSize: String?
}

// MARK: - Search ViewModel

@MainActor
class SearchViewModel: ObservableObject {
    @Published var query = ""
    @Published var activeFilter: SearchFilter = .all
    @Published var results: [SearchResult] = []
    @Published var recentSearches: [String] = []
    @Published var isSearching = false
    @Published var filterContact: String?
    @Published var filterDateRange: ClosedRange<Date>?
    @Published var filterChat: String?
    @Published var showAdvancedFilters = false

    func clearRecentSearches() {
        recentSearches.removeAll()
    }
}

// MARK: - Filter Chip

struct FilterChip: View {
    let label: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(label)
                .font(.custom("Inter", size: 13))
                .fontWeight(isSelected ? .bold : .medium)
                .foregroundStyle(isSelected ? .white : Color.Echo.outline)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    Capsule()
                        .fill(isSelected ? Color.Echo.primaryContainer : Color.Echo.surfaceContainerLow)
                )
        }
    }
}

// MARK: - Search Result Row

struct SearchResultRow: View {
    let result: SearchResult

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Circle()
                .fill(Color.Echo.surfaceContainerHigh)
                .frame(width: 40, height: 40)
                .overlay(
                    Image(systemName: "person.fill")
                        .font(.system(size: 16))
                        .foregroundStyle(Color.Echo.outline)
                )

            VStack(alignment: .leading, spacing: 4) {
                Text(result.contactName)
                    .font(.custom("Inter", size: 14))
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.Echo.onSurface)
                Text(result.matchedText)
                    .font(Font.Echo.bodyMedium)
                    .foregroundStyle(Color.Echo.outline)
                    .lineLimit(2)
            }

            Spacer()

            Text(result.timestamp.formatted(date: .abbreviated, time: .shortened))
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.outline)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
    }
}

// MARK: - Recent Searches Section

struct RecentSearchesSection: View {
    let searches: [String]
    let onTap: (String) -> Void
    let onClear: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("RECENT SEARCHES")
                    .font(.custom("Inter", size: 10))
                    .fontWeight(.bold)
                    .tracking(2)
                    .foregroundStyle(Color.Echo.outline)
                Spacer()
                if !searches.isEmpty {
                    Button("Clear") { onClear() }
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.primaryContainer)
                }
            }
            .padding(.horizontal, 20)
            .padding(.top, 20)

            ForEach(searches, id: \.self) { search in
                Button { onTap(search) } label: {
                    HStack(spacing: 12) {
                        Image(systemName: "clock.arrow.circlepath")
                            .foregroundStyle(Color.Echo.outline)
                        Text(search)
                            .font(Font.Echo.bodyMedium)
                            .foregroundStyle(Color.Echo.onSurface)
                        Spacer()
                    }
                    .padding(.horizontal, 20)
                    .padding(.vertical, 10)
                }
            }
        }
    }
}

// MARK: - Empty Search State

struct EmptySearchState: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 48))
                .foregroundStyle(Color.Echo.outline.opacity(0.5))
            Text("No results found")
                .font(Font.Echo.bodyLarge)
                .foregroundStyle(Color.Echo.outline)
        }
        .frame(maxWidth: .infinity)
        .padding(.top, 60)
    }
}
