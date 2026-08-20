#if os(iOS)
import SwiftUI

// MARK: - Quest ViewModel (WO-271)

@MainActor
final class QuestViewModel: ObservableObject {
    @Published private(set) var starterQuests: [QuestItem] = []
    @Published private(set) var advancedQuests: [QuestItem] = []
    @Published private(set) var claimedBadges: [String] = []
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?
    @Published var claimingQuestId: String?

    private let api: GamificationAPIClient

    init(api: GamificationAPIClient) {
        self.api = api
    }

    var nextAvailableQuest: QuestItem? {
        (starterQuests + advancedQuests).first { !$0.isCompleted }
    }

    var nextClaimableQuest: QuestItem? {
        (starterQuests + advancedQuests).first(\.isClaimable)
    }

    func load() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            let catalog = try await api.fetchQuestCatalog()
            starterQuests = catalog.starter
            advancedQuests = catalog.advanced
            claimedBadges = (catalog.starter + catalog.advanced)
                .filter(\.rewardClaimed)
                .map(\.badge)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func claim(_ quest: QuestItem) async {
        guard quest.isClaimable else { return }
        claimingQuestId = quest.id
        errorMessage = nil
        defer { claimingQuestId = nil }
        do {
            _ = try await api.claimQuest(quest.questId)
            await load()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Profile Quest Section

struct QuestGamificationSection: View {
    @StateObject private var viewModel: QuestViewModel
    @State private var showAdvanced = false

    init(api: GamificationAPIClient) {
        _viewModel = StateObject(wrappedValue: QuestViewModel(api: api))
    }

    var body: some View {
        VStack(alignment: .leading, spacing: Spacing.md.rawValue) {
            HStack {
                Text("Quests")
                    .typographyStyle(.h3, color: .echoPrimaryText)
                Spacer()
                if viewModel.isLoading {
                    ProgressView()
                }
            }

            if !viewModel.claimedBadges.isEmpty {
                badgeRow
            }

            if let next = viewModel.nextAvailableQuest {
                nextQuestCallout(next)
            }

            ForEach(viewModel.starterQuests) { quest in
                QuestRow(quest: quest, isClaiming: viewModel.claimingQuestId == quest.id) {
                    Task { await viewModel.claim(quest) }
                }
            }

            if !viewModel.advancedQuests.isEmpty {
                DisclosureGroup(isExpanded: $showAdvanced) {
                    ForEach(viewModel.advancedQuests) { quest in
                        QuestRow(quest: quest, isClaiming: viewModel.claimingQuestId == quest.id) {
                            Task { await viewModel.claim(quest) }
                        }
                    }
                } label: {
                    Text("Advanced Quests")
                        .typographyStyle(.bodySmall, color: .echoSecondaryText)
                }
            }

            if let error = viewModel.errorMessage {
                Text(error)
                    .typographyStyle(.caption, color: .echoError)
            }
        }
        .padding(Spacing.lg.rawValue)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.echoSurface)
        .clipShape(RoundedRectangle(cornerRadius: Spacing.lg.rawValue, style: .continuous))
        .glacialShadow()
        .task { await viewModel.load() }
    }

    private var badgeRow: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: Spacing.sm.rawValue) {
                ForEach(viewModel.claimedBadges, id: \.self) { badge in
                    Text(badge)
                        .typographyStyle(.caption, color: .echoPrimaryText)
                        .padding(.horizontal, 10)
                        .padding(.vertical, 6)
                        .background(Color.echoSignal.opacity(0.15))
                        .cornerRadius(999)
                }
            }
        }
    }

    private func nextQuestCallout(_ quest: QuestItem) -> some View {
        HStack(spacing: Spacing.sm.rawValue) {
            Image(systemName: "sparkles")
                .foregroundColor(.echoSignal)
            VStack(alignment: .leading, spacing: 2) {
                Text("Next up: \(quest.title)")
                    .typographyStyle(.bodySmall, color: .echoPrimaryText)
                Text(quest.description)
                    .typographyStyle(.caption, color: .echoSecondaryText)
                    .lineLimit(2)
            }
        }
        .padding(Spacing.md.rawValue)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: Spacing.md.rawValue, style: .continuous))
    }
}

private struct QuestRow: View {
    let quest: QuestItem
    let isClaiming: Bool
    let onClaim: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .top) {
                VStack(alignment: .leading, spacing: 4) {
                    Text(quest.title)
                        .typographyStyle(.body, color: .echoPrimaryText)
                    Text(quest.description)
                        .typographyStyle(.caption, color: .echoSecondaryText)
                        .lineLimit(2)
                }
                Spacer()
                Text("+\(formatEcho(quest.rewardAmount))")
                    .typographyStyle(.bodySmall, color: .echoSignal)
            }

            ProgressView(value: quest.progressFraction)
                .tint(quest.isCompleted ? Color.echoTrustGreen : Color.echoSignal)

            HStack {
                if quest.rewardClaimed {
                    Label(quest.badge, systemImage: "checkmark.seal.fill")
                        .typographyStyle(.caption, color: .echoTrustGreen)
                } else if quest.isClaimable {
                    Button(action: onClaim) {
                        if isClaiming {
                            ProgressView()
                        } else {
                            Text("Claim")
                                .typographyStyle(.bodySmall, color: .white)
                                .padding(.horizontal, 14)
                                .padding(.vertical, 8)
                                .background(Color.echoSignal)
                                .clipShape(RoundedRectangle(cornerRadius: Spacing.sm.rawValue, style: .continuous))
                        }
                    }
                    .disabled(isClaiming)
                } else if quest.isCompleted {
                    Text("Claimed pending")
                        .typographyStyle(.caption, color: .echoSecondaryText)
                } else {
                    Text("\(quest.progress)/\(quest.requiredCount)")
                        .typographyStyle(.caption, color: .echoSecondaryText)
                }
                Spacer()
                Text(quest.badge)
                    .typographyStyle(.caption, color: .echoSecondaryText)
            }
        }
        .padding(.vertical, 4)
    }
}
#endif
