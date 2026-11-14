// Features/Safety/ContactSafetyBanner.swift
//
// Non-blocking, dismissible safety banner shown at the top of a chat thread on a risky first
// contact. Mirrors ThreadSummaryBanner's style. Copy is friendly, never accusatory.

#if os(iOS)
import SwiftUI

public struct ContactSafetyBanner: View {
    let assessment: SafetyAssessment
    let onVerify: () -> Void
    let onBlock: () -> Void
    let onDismiss: () -> Void

    public init(assessment: SafetyAssessment,
                onVerify: @escaping () -> Void = {},
                onBlock: @escaping () -> Void = {},
                onDismiss: @escaping () -> Void = {}) {
        self.assessment = assessment
        self.onVerify = onVerify
        self.onBlock = onBlock
        self.onDismiss = onDismiss
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: icon).font(.system(size: 13, weight: .semibold))
                Text(title).font(.system(size: 13, weight: .semibold))
                Spacer()
                Button(action: onDismiss) {
                    Image(systemName: "xmark").font(.system(size: 11, weight: .semibold))
                }
                .foregroundColor(tint.opacity(0.7))
            }
            ForEach(Array(messages.enumerated()), id: \.offset) { _, line in
                Text("• \(line)").font(.system(size: 12)).foregroundColor(.echoInk70)
            }
            HStack(spacing: 10) {
                Button(action: onVerify) { actionLabel("View profile", filled: true) }
                Button(action: onBlock) { actionLabel("Block", filled: false) }
            }
            .padding(.top, 2)
        }
        .padding(12)
        .background(tint.opacity(0.10))
        .overlay(RoundedRectangle(cornerRadius: 12).stroke(tint.opacity(0.30), lineWidth: 1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .foregroundColor(tint)
    }

    @ViewBuilder private func actionLabel(_ text: String, filled: Bool) -> some View {
        Text(text)
            .font(.system(size: 12, weight: .semibold))
            .foregroundColor(filled ? .white : tint)
            .padding(.horizontal, 12).padding(.vertical, 6)
            .background(filled ? tint : Color.clear)
            .overlay(RoundedRectangle(cornerRadius: 8).stroke(tint.opacity(filled ? 0 : 0.5), lineWidth: 1))
            .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private var tint: Color { assessment.level == .warning ? .echoAlert : .echoWarning }
    private var icon: String { assessment.level == .warning ? "exclamationmark.triangle.fill" : "shield.lefthalf.filled" }
    private var title: String { assessment.level == .warning ? "Possible impersonation" : "New, unverified contact" }

    private var messages: [String] {
        assessment.reasons.map {
            switch $0 {
            case .firstContact: return "This is the first time this person has messaged you."
            case .unverifiedSender: return "They haven't verified their identity."
            case .lowTrustSender(let tier): return "Low trust level (Tier \(tier))."
            case .possibleImpersonation(let name): return "Their name looks like \"\(name)\", a verified contact — but it's a different account."
            }
        }
    }
}
// MARK: - Report sheet (contact, message, channel, group)

/// Reason picker, confirmation, and success/failure for a safety report.
/// Same shape as the contact report sheet; the report goes to `/v1/safety/reports`.
struct SafetyReportSheet: View {
    @StateObject private var viewModel: SafetyReportViewModel
    @Environment(\.dismiss) private var dismiss
    @State private var showConfirmation = false
    private let onSubmitted: () -> Void

    init(target: SafetyReportTarget, onSubmitted: @escaping () -> Void = {}) {
        _viewModel = StateObject(wrappedValue: SafetyReportSheet.liveViewModel(target: target))
        self.onSubmitted = onSubmitted
    }

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 12) {
                Text(viewModel.target.prompt)
                    .font(.system(size: 14))
                    .foregroundStyle(Color.Echo.outline)
                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.Echo.error)
                }
                ForEach(SafetyReportReason.allCases) { reason in
                    Button {
                        viewModel.selectedReason = reason
                    } label: {
                        HStack {
                            Text(reason.label)
                                .font(.system(size: 16, weight: .semibold))
                            Spacer()
                            if viewModel.selectedReason == reason {
                                Image(systemName: "checkmark")
                            }
                        }
                        .padding(.vertical, 8)
                    }
                    .accessibilityAddTraits(viewModel.selectedReason == reason ? .isSelected : [])
                }
                TextField("Add details (optional)", text: $viewModel.details, axis: .vertical)
                    .lineLimit(2...4)
                    .textFieldStyle(.roundedBorder)
                Button(role: .destructive) {
                    showConfirmation = true
                } label: {
                    HStack {
                        if viewModel.phase == .submitting { ProgressView() }
                        Text(viewModel.target.alsoBlocks ? "Block & report" : "Send report")
                            .font(.system(size: 16, weight: .semibold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 10)
                }
                .disabled(!viewModel.canSubmit)
                .accessibilityIdentifier("safetyReport.submit")
                Spacer()
            }
            .padding(24)
            .navigationTitle(viewModel.target.title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
        }
        .presentationDetents([.medium, .large])
        .confirmationDialog(
            viewModel.target.alsoBlocks ? "Block this person and send the report?" : "Send this report to Echo?",
            isPresented: $showConfirmation,
            titleVisibility: .visible
        ) {
            Button(viewModel.target.alsoBlocks ? "Block & report" : "Send report", role: .destructive) {
                Task { await viewModel.submitReport() }
            }
            Button("Cancel", role: .cancel) {}
        }
        .alert("Report sent", isPresented: Binding(
            get: { viewModel.phase == .succeeded },
            set: { _ in }
        )) {
            Button("OK", role: .cancel) {
                onSubmitted()
                dismiss()
            }
        } message: {
            Text(viewModel.successMessage)
        }
    }

    @MainActor
    private static func liveViewModel(target: SafetyReportTarget) -> SafetyReportViewModel {
        SafetyReportViewModel(
            target: target,
            submit: { request in
                guard let client = DIContainer.shared.resolveAPIClient() else {
                    throw SafetyReportError.notSignedIn
                }
                try await ContactSocialAPIClient(apiClient: client).submitReport(request)
            },
            block: { did in
                if let useCase = DIContainer.shared.resolveBlockContactUseCase() {
                    try await useCase.block(did: did)
                } else if let client = DIContainer.shared.resolveAPIClient() {
                    try await ContactSocialAPIClient(apiClient: client).blockContact(did: did)
                } else {
                    throw SafetyReportError.notSignedIn
                }
            }
        )
    }
}
#endif
