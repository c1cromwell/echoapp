#if os(iOS)
import SwiftUI

/// Unified people + messages search entry (Telegram-style global search).
struct GlobalSearchSheet: View {
    enum Tab: String, CaseIterable, Identifiable {
        case people, messages
        var id: String { rawValue }
        var title: String {
            switch self {
            case .people: return "People"
            case .messages: return "Messages"
            }
        }
    }

    var initialQuery: String = ""
    var initialTab: Tab = .messages

    @Environment(\.dismiss) private var dismiss
    @State private var tab: Tab = .messages

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                Picker("Search", selection: $tab) {
                    ForEach(Tab.allCases) { t in
                        Text(t.title).tag(t)
                    }
                }
                .pickerStyle(.segmented)
                .padding(.horizontal, Spacing.lg.rawValue)
                .padding(.vertical, Spacing.sm.rawValue)

                Group {
                    switch tab {
                    case .people:
                        UsernameSearchView()
                    case .messages:
                        SearchView(initialQuery: initialQuery.isEmpty ? nil : initialQuery)
                    }
                }
            }
            .background(Color.echoPaper.ignoresSafeArea())
            .navigationTitle("Search")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .onAppear {
                tab = initialTab
            }
        }
    }
}
#endif
