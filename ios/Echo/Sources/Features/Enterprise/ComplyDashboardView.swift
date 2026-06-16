import SwiftUI

/// WO-314 read-mostly compliance posture dashboard for org administrators.
struct ComplyDashboardView: View {
    let orgDID: String
    @StateObject private var viewModel = ComplyDashboardViewModel()

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                if let summary = viewModel.summary {
                    postureGrid(summary)
                } else if viewModel.isLoading {
                    ProgressView()
                        .padding(.top, 40)
                } else if let error = viewModel.errorMessage {
                    Text(error)
                        .font(Font.Echo.bodyMedium)
                        .foregroundStyle(Color.Echo.outline)
                        .padding()
                }

                if let audit = viewModel.audit {
                    GhostBorderSection(title: "AUDIT SNAPSHOT") {
                        Text(audit.verificationNotice)
                            .font(Font.Echo.labelMd)
                            .foregroundStyle(Color.Echo.outline)
                        InfoRow(label: "Anchor health", value: audit.anchorHealth.capitalized)
                        InfoRow(label: "Active holds", value: "\(audit.activeLitigationHolds)")
                    }
                }

                NavigationLink {
                    ComplyPolicyListView(orgDID: orgDID, viewModel: viewModel)
                } label: {
                    HStack {
                        Text("Retention policies")
                            .font(Font.Echo.bodyMedium)
                        Spacer()
                        Image(systemName: "chevron.right")
                            .font(.system(size: 12))
                    }
                    .foregroundStyle(Color.Echo.onSurface)
                    .padding()
                    .background(Color.Echo.surfaceContainerHigh)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                }
            }
            .padding()
        }
        .background(Color.Echo.surface)
        .navigationTitle("Compliance")
        .task { await viewModel.load(orgDID: orgDID) }
    }

    @ViewBuilder
    private func postureGrid(_ summary: ComplyDashboardSummary) -> some View {
        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
            ComplyMetricCard(title: "DE coverage", value: summary.deCoverageRate)
            ComplyMetricCard(title: "Retention", value: "\(summary.activeRetentionPolicies)")
            ComplyMetricCard(title: "Holds", value: "\(summary.litigationHolds)")
            ComplyMetricCard(title: "Exports", value: "\(summary.pendingExports)")
        }
        ComplyMetricCard(title: "Anchor health", value: summary.anchorHealth.capitalized)
    }
}

private struct ComplyMetricCard: View {
    let title: String
    let value: String

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.outline)
            Text(value)
                .font(.system(size: 22, weight: .semibold))
                .foregroundStyle(Color.Echo.onSurface)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(Color.Echo.surfaceContainerHigh)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

struct ComplyPolicyListView: View {
    let orgDID: String
    @ObservedObject var viewModel: ComplyDashboardViewModel

    var body: some View {
        List(viewModel.policies) { policy in
            VStack(alignment: .leading, spacing: 4) {
                Text(policy.policyType.replacingOccurrences(of: "_", with: " ").capitalized)
                    .font(Font.Echo.bodyMedium)
                if let conv = policy.conversationId {
                    Text(conv)
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.Echo.outline)
                }
            }
        }
        .navigationTitle("Policies")
        .task { await viewModel.loadPolicies(orgDID: orgDID) }
    }
}

@MainActor
final class ComplyDashboardViewModel: ObservableObject {
    @Published var summary: ComplyDashboardSummary?
    @Published var audit: ComplyAuditReport?
    @Published var policies: [ComplyRetentionPolicy] = []
    @Published var isLoading = false
    @Published var errorMessage: String?

    private var complyAPI: (any ComplyAPIClient)?

    func load(orgDID: String) async {
        isLoading = true
        defer { isLoading = false }
        guard let client = resolveClient() else {
            errorMessage = "Comply API unavailable"
            return
        }
        do {
            async let dash = client.dashboard(orgDID: orgDID)
            async let auditReport = client.auditReport(orgDID: orgDID)
            summary = try await dash
            audit = try await auditReport
        } catch {
            errorMessage = "Could not load compliance posture"
        }
    }

    func loadPolicies(orgDID: String) async {
        guard let client = resolveClient() else { return }
        policies = (try? await client.listPolicies(orgDID: orgDID)) ?? []
    }

    private func resolveClient() -> (any ComplyAPIClient)? {
        if let complyAPI { return complyAPI }
        #if os(iOS)
        if let api = DIContainer.shared.resolveAPIClient() {
            let client = ComplyAPI(apiClient: api)
            complyAPI = client
            return client
        }
        #endif
        return nil
    }
}
