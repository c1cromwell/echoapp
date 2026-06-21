#if os(iOS)
// Presentation/Screens/Search/SearchView.swift
// Advanced message search with filters and recent searches

import SwiftUI

// MARK: - Search View

public struct SearchView: View {
    @StateObject private var viewModel = SearchViewModel()
    private let initialQuery: String?
    @FocusState private var isSearchFocused: Bool
    @Environment(\.dismiss) private var dismiss

    public init(initialQuery: String? = nil) {
        self.initialQuery = initialQuery
    }

    public var body: some View {
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
                    FilterChip(
                        label: "Filters",
                        isSelected: viewModel.showAdvancedFilters,
                        action: { viewModel.showAdvancedFilters.toggle() }
                    )
                }
                .padding(.horizontal, 16)
            }
            .padding(.vertical, 8)

            if viewModel.showAdvancedFilters {
                SearchAdvancedFiltersPanel(viewModel: viewModel)
                    .padding(.horizontal, 16)
                    .padding(.bottom, 8)
            }

            if let chatId = viewModel.filterChat,
               let name = viewModel.filteredConversationName {
                HStack {
                    Text("In: \(name)")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(Color.Echo.outline)
                    Spacer()
                    Button("Clear") { viewModel.clearChatFilter() }
                        .font(.system(size: 12, weight: .semibold))
                }
                .padding(.horizontal, 16)
                .padding(.bottom, 4)
            }

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
        .onAppear {
            isSearchFocused = true
            if let initialQuery, !initialQuery.isEmpty, viewModel.query.isEmpty {
                viewModel.query = initialQuery
                viewModel.performSearch()
            }
        }
        .onChange(of: viewModel.query) { _, _ in viewModel.performSearch() }
        .onChange(of: viewModel.activeFilter) { _, _ in viewModel.performSearch() }
        .onChange(of: viewModel.filterChat) { _, _ in viewModel.performSearch() }
        .onChange(of: viewModel.filterDateRange) { _, _ in viewModel.performSearch() }
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

    private var searchTask: Task<Void, Never>?

    var filteredConversationName: String? {
        guard let chatId = filterChat else { return nil }
        return ConversationStore.shared.conversations.first(where: { $0.id == chatId })?.contactName
    }

    init() {
        recentSearches = SearchHistoryStore.load()
        Task { await LocalMessageIndexer.shared.rebuildFromLocalThreads() }
    }

    func clearRecentSearches() {
        recentSearches.removeAll()
        SearchHistoryStore.clear()
    }

    func clearChatFilter() {
        filterChat = nil
    }

    func performSearch() {
        searchTask?.cancel()
        let q = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard q.count >= 2 else {
            results = []
            isSearching = false
            return
        }
        isSearching = true
        let filter = activeFilter
        searchTask = Task {
            try? await Task.sleep(nanoseconds: 300_000_000)
            guard !Task.isCancelled else { return }
            var options = SearchQueryOptions()
            options.conversationId = filterChat
            options.since = filterDateRange?.lowerBound
            options.until = filterDateRange?.upperBound
            options.includeArchived = activeFilter == .all && filterChat != nil
            let hits = await KeywordSearchEngine.shared.search(
                query: q,
                filter: activeFilter,
                options: options
            )
            guard !Task.isCancelled else { return }
            results = hits.map { hit in
                SearchResult(
                    id: hit.messageId,
                    conversationId: hit.conversationId,
                    contactName: conversationName(for: hit.conversationId, senderDID: hit.senderDID),
                    contactAvatar: nil,
                    matchedText: hit.matchedText,
                    timestamp: hit.timestamp,
                    messageType: hit.contentType,
                    attachmentName: nil,
                    attachmentSize: nil
                )
            }
            isSearching = false
            SearchHistoryStore.append(q)
            recentSearches = SearchHistoryStore.load()
        }
    }

    private func conversationName(for conversationId: String, senderDID: String) -> String {
        if let conv = ConversationStore.shared.conversations.first(where: { $0.id == conversationId }) {
            return conv.contactName
        }
        return ContactThreadHelper.truncatedDID(senderDID)
    }
}

// MARK: - Advanced filters

struct SearchAdvancedFiltersPanel: View {
    @ObservedObject var viewModel: SearchViewModel
    @State private var startDate = Calendar.current.date(byAdding: .month, value: -1, to: Date()) ?? Date()
    @State private var endDate = Date()
    @State private var useDateRange = false

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Conversation")
                .font(.system(size: 12, weight: .semibold))
                .foregroundStyle(Color.Echo.outline)
            Picker("Chat", selection: Binding(
                get: { viewModel.filterChat ?? "" },
                set: { viewModel.filterChat = $0.isEmpty ? nil : $0 }
            )) {
                Text("All chats").tag("")
                ForEach(ConversationStore.shared.conversations) { conv in
                    Text(conv.contactName).tag(conv.id)
                }
            }
            .pickerStyle(.menu)

            Toggle("Limit to date range", isOn: $useDateRange)
                .font(.system(size: 13))
            if useDateRange {
                DatePicker("From", selection: $startDate, displayedComponents: .date)
                DatePicker("To", selection: $endDate, displayedComponents: .date)
                    .onChange(of: startDate) { _, _ in applyDateRange(useDateRange) }
                    .onChange(of: endDate) { _, _ in applyDateRange(useDateRange) }
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color.Echo.surfaceContainerLow)
        )
        .onChange(of: useDateRange) { _, enabled in
            applyDateRange(enabled)
        }
    }

    private func applyDateRange(_ enabled: Bool) {
        if enabled {
            viewModel.filterDateRange = startDate...endDate
        } else {
            viewModel.filterDateRange = nil
        }
    }
}

/// Sheet wrapper for hub → full message search (WO-197).
struct MessageSearchSheet: View {
    let initialQuery: String
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            SearchView(initialQuery: initialQuery)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Done") { dismiss() }
                    }
                }
        }
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
                .font(.system(size: 13))
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
                    .font(.system(size: 14))
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
                    .font(.system(size: 10))
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
#endif
