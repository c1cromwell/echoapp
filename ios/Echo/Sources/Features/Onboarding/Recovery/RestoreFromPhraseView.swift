#if os(iOS)
// Features/Onboarding/Recovery/RestoreFromPhraseView.swift

import SwiftUI

public struct RestoreFromPhraseView: View {
    @Bindable var coordinator: RecoveryCoordinator
    @State private var words: [String] = Array(repeating: "", count: 24)
    @State private var isRestoring = false
    @State private var errorMessage: String?

    public init(coordinator: RecoveryCoordinator) {
        self.coordinator = coordinator
    }

    public var body: some View {
        ZStack {
            Color.echoNight.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 20) {
                    titleBlock
                        .padding(.horizontal, 24)
                        .padding(.top, 16)

                    wordGrid
                        .padding(.horizontal, 20)

                    if let errorMessage {
                        errorBanner(errorMessage)
                            .padding(.horizontal, 24)
                    }

                    restoreButton
                        .padding(.horizontal, 24)
                        .padding(.top, 8)
                }
                .padding(.bottom, 24)
            }
        }
        .navigationTitle("Restore account")
        .navigationBarTitleDisplayMode(.inline)
        .disabled(isRestoring)
    }

    // MARK: - Subviews

    private var titleBlock: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Enter your 24-word phrase")
                .font(.system(size: 22, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
            Text("Enter the words in the order you wrote them down.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var wordGrid: some View {
        let columns = Array(repeating: GridItem(.flexible(), spacing: 8), count: 2)
        return LazyVGrid(columns: columns, spacing: 10) {
            ForEach(0..<24, id: \.self) { index in
                wordField(index: index)
            }
        }
    }

    private func wordField(index: Int) -> some View {
        HStack(spacing: 6) {
            Text("\(index + 1)")
                .font(.system(size: 10, weight: .semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant.opacity(0.6))
                .monospacedDigit()
                .frame(width: 18)
            TextField("", text: $words[index])
                .font(.system(size: 14, weight: .medium, design: .monospaced))
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled(true)
                .textContentType(.newPassword) // suppresses iCloud Keychain autofill
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 10)
        .background(Color.Echo.surfaceContainerLowest, in: RoundedRectangle(cornerRadius: 10))
        .overlay(
            RoundedRectangle(cornerRadius: 10)
                .strokeBorder(
                    wordIsValid(index)
                        ? Color.Echo.outlineVariant.opacity(0.3)
                        : Color.red.opacity(0.4),
                    lineWidth: 0.5
                )
        )
    }

    private func errorBanner(_ message: String) -> some View {
        Text(message)
            .font(.system(size: 13))
            .foregroundStyle(Color.Echo.onSurface)
            .padding(14)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.red.opacity(0.08), in: RoundedRectangle(cornerRadius: 14))
            .overlay(
                RoundedRectangle(cornerRadius: 14)
                    .strokeBorder(Color.red.opacity(0.3), lineWidth: 0.5)
            )
    }

    private var restoreButton: some View {
        Button { Task { await restore() } } label: {
            HStack {
                if isRestoring {
                    ProgressView()
                        .tint(.white)
                        .padding(.trailing, 6)
                }
                Text(isRestoring ? "Restoring…" : "Restore account")
                    .font(.system(size: 15, weight: .semibold))
            }
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
            .opacity(allWordsFilled ? 1 : 0.55)
        }
        .buttonStyle(SpringPressStyle())
        .disabled(!allWordsFilled || isRestoring)
        .accessibilityHint(allWordsFilled ? "Restore your ECHO account from your recovery phrase" : "Enter all 24 words to continue")
    }

    // MARK: - Helpers

    private var allWordsFilled: Bool {
        words.allSatisfy { !$0.trimmingCharacters(in: .whitespaces).isEmpty }
    }

    private func wordIsValid(_ index: Int) -> Bool {
        let w = words[index].lowercased().trimmingCharacters(in: .whitespaces)
        return w.isEmpty || BIP39Wordlist.contains(w)
    }

    private func restore() async {
        isRestoring = true
        errorMessage = nil
        defer { isRestoring = false }

        let normalized = words.map { $0.lowercased().trimmingCharacters(in: .whitespaces) }
        guard let phrase = RecoveryPhrase(words: normalized), phrase.entropy != nil else {
            errorMessage = "That doesn't look like a valid recovery phrase — check your words and try again."
            return
        }

        do {
            let restored = try await RecoveryService.shared.restore(from: phrase)
            coordinator.onRestoreComplete(restored)
        } catch RecoveryError.deviceAlreadyEnrolled {
            errorMessage = "This device already has an ECHO account. Use your existing credentials instead."
        } catch {
            errorMessage = "Couldn't restore your account. \(error.localizedDescription)"
        }
    }
}
#endif
