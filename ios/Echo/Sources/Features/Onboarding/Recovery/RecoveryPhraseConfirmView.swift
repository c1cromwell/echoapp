#if os(iOS)
// Features/Onboarding/Recovery/RecoveryPhraseConfirmView.swift

import SwiftUI

struct RecoveryPhraseConfirmView: View {
    @Bindable var coordinator: RecoveryCoordinator
    @FocusState private var focusedField: Int?

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            ScrollView {
                VStack(spacing: 24) {
                    titleBlock
                        .padding(.top, 24)
                        .padding(.horizontal, 24)

                    ForEach(coordinator.challengePositions, id: \.self) { position in
                        challengeField(position: position)
                            .padding(.horizontal, 24)
                    }

                    verifyButton
                        .padding(.horizontal, 24)
                        .padding(.top, 12)
                }
                .padding(.bottom, 24)
            }
        }
        .navigationTitle("Confirm")
        .navigationBarTitleDisplayMode(.inline)
    }

    // MARK: - Subviews

    private var titleBlock: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Prove you wrote it down")
                .font(.system(size: 22, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
            Text("Enter these words from your recovery phrase.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func challengeField(position: Int) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Word #\(position)")
                .font(.system(size: 11, weight: .semibold))
                .tracking(1.5)
                .foregroundStyle(Color.Echo.onSurfaceVariant)

            TextField("", text: Binding(
                get: { coordinator.userAnswers[position] ?? "" },
                set: {
                    coordinator.userAnswers[position] = $0
                        .lowercased()
                        .trimmingCharacters(in: .whitespaces)
                }
            ))
            .font(.system(size: 15, weight: .medium, design: .monospaced))
            .foregroundStyle(Color.Echo.onSurface)
            .textInputAutocapitalization(.never)
            .autocorrectionDisabled(true)
            .focused($focusedField, equals: position)
            .padding(14)
            .background(Color.Echo.surfaceContainerLowest, in: RoundedRectangle(cornerRadius: 16))
            .overlay(
                RoundedRectangle(cornerRadius: 16)
                    .strokeBorder(
                        focusedField == position
                            ? Color.Echo.primaryContainer.opacity(0.45)
                            : Color.Echo.outlineVariant.opacity(0.3),
                        lineWidth: 0.5
                    )
            )
            .animation(.easeInOut(duration: 0.15), value: focusedField)
        }
    }

    private var verifyButton: some View {
        Button { verify() } label: {
            Text("Verify")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 16)
                .background(
                    LinearGradient(
                        colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    ),
                    in: RoundedRectangle(cornerRadius: 22, style: .continuous)
                )
                .opacity(allFieldsFilled ? 1 : 0.55)
        }
        .buttonStyle(SpringPressStyle())
        .disabled(!allFieldsFilled)
        .accessibilityHint(allFieldsFilled ? "Verify your phrase" : "Enter all three words to continue")
    }

    private var allFieldsFilled: Bool {
        coordinator.challengePositions.allSatisfy {
            !(coordinator.userAnswers[$0] ?? "").trimmingCharacters(in: .whitespaces).isEmpty
        }
    }

    private func verify() {
        guard let phrase = coordinator.phrase else { return }
        let allCorrect = coordinator.challengePositions.allSatisfy { pos in
            let entered = (coordinator.userAnswers[pos] ?? "")
                .lowercased()
                .trimmingCharacters(in: .whitespaces)
            return entered == phrase.word(at: pos)
        }
        if allCorrect {
            coordinator.confirmationSucceeded()
        } else {
            coordinator.confirmationFailed()
        }
    }
}
#endif
