#if os(iOS)
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

                auditExportRow

                if !viewModel.segments.isEmpty {
                    GhostBorderSection(title: "SEGMENT REPORTS") {
                        ForEach(viewModel.segments) { seg in
                            VStack(alignment: .leading, spacing: 4) {
                                Text(seg.label)
                                    .font(Font.Echo.bodyMedium)
                                ForEach(seg.metrics, id: \.key) { m in
                                    InfoRow(label: m.label, value: m.value)
                                }
                            }
                            .padding(.vertical, 4)
                        }
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
        .sheet(isPresented: $viewModel.showShareSheet) {
            if let url = viewModel.auditPDFURL {
                ShareLink(item: url) {
                    Label("Share audit PDF", systemImage: "square.and.arrow.up")
                        .font(Font.Echo.bodyMedium)
                        .padding()
                }
                .presentationDetents([.medium])
            }
        }
    }

    @ViewBuilder
    private var auditExportRow: some View {
        Button {
            Task { await viewModel.exportAuditPDF(orgDID: orgDID) }
        } label: {
            HStack {
                if viewModel.isExportingPDF {
                    ProgressView()
                        .controlSize(.small)
                } else {
                    Image(systemName: "doc.richtext")
                }
                Text("Export audit PDF")
                    .font(Font.Echo.bodyMedium)
                Spacer()
            }
            .foregroundStyle(Color.Echo.onSurface)
            .padding()
            .background(Color.Echo.surfaceContainerHigh)
            .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        .disabled(viewModel.isExportingPDF)
    }

    @ViewBuilder
    private func postureGrid(_ summary: ComplyDashboardSummary) -> some View {
        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
            ComplyMetricCard(title: "DE coverage", value: summary.deCoverageRate)
            ComplyMetricCard(title: "Retention", value: "\(summary.activeRetentionPolicies)")
            ComplyMetricCard(title: "Holds", value: "\(summary.litigationHolds)")
            ComplyMetricCard(title: "Exports", value: "\(summary.pendingExports)")
        }
        ComplyMetricCard(
            title: "Anchor health",
            value: summary.anchorHealth.capitalized,
            tone: anchorTone(summary.anchorHealth)
        )
    }

    private func anchorTone(_ health: String) -> Color {
        switch health.lowercased() {
        case "healthy": return .green
        case "degraded": return .orange
        default: return .red
        }
    }
}

private struct ComplyMetricCard: View {
    let title: String
    let value: String
    var tone: Color = Color.Echo.onSurface

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(Font.Echo.labelMd)
                .foregroundStyle(Color.Echo.outline)
            Text(value)
                .font(.system(size: 22, weight: .semibold))
                .foregroundStyle(tone)
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
                if let ref = policy.dataL1Ref, !ref.isEmpty {
                    Text("anchor \(String(ref.prefix(12)))…")
                        .font(Font.Echo.labelMd)
                        .foregroundStyle(Color.green.opacity(0.85))
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
    @Published var segments: [ComplySegmentReport] = []
    @Published var isLoading = false
    @Published var isExportingPDF = false
    @Published var showShareSheet = false
    @Published var auditPDFURL: URL?
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
            async let segmentReport = client.segmentDashboard(orgDID: orgDID)
            summary = try await dash
            audit = try await auditReport
            if let segDash = try? await segmentReport {
                segments = segDash.segments
            }
        } catch {
            errorMessage = "Could not load compliance posture"
        }
    }

    func loadPolicies(orgDID: String) async {
        guard let client = resolveClient() else { return }
        policies = (try? await client.listPolicies(orgDID: orgDID)) ?? []
    }

    func exportAuditPDF(orgDID: String) async {
        guard let client = resolveClient() else { return }
        isExportingPDF = true
        defer { isExportingPDF = false }
        do {
            let data = try await client.auditReportPDF(orgDID: orgDID)
            let url = FileManager.default.temporaryDirectory
                .appendingPathComponent("echo-comply-audit-\(orgDID.suffix(8)).pdf")
            try data.write(to: url, options: .atomic)
            auditPDFURL = url
            showShareSheet = true
        } catch {
            errorMessage = "Audit PDF export failed"
        }
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
#endif
