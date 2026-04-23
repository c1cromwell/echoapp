// Features/Onboarding/Recovery/RecoveryPhraseDisplayView.swift

import SwiftUI

struct RecoveryPhraseDisplayView: View {
    @Bindable var coordinator: RecoveryCoordinator
    @State private var screenCaptureDetected = false

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 0) {
                Spacer(minLength: 12)

                titleBlock
                    .padding(.horizontal, 24)

                if let phrase = coordinator.phrase {
                    wordGrid(phrase: phrase)
                        .padding(.horizontal, 20)
                        .padding(.top, 24)
                } else {
                    ProgressView()
                        .padding(.top, 40)
                }

                Spacer()

                footer
                    .padding(.horizontal, 24)
                    .padding(.bottom, 24)
            }
            .blur(radius: screenCaptureDetected ? 20 : 0)

            if screenCaptureDetected {
                screenCaptureWarning
            }
        }
        .navigationTitle("Recovery phrase")
        .navigationBarTitleDisplayMode(.inline)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .onReceive(
            NotificationCenter.default.publisher(for: UIScreen.capturedDidChangeNotification)
        ) { _ in
            screenCaptureDetected = UIScreen.main.isCaptured
        }
    }

    // MARK: - Subviews

    private var titleBlock: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Write these 24 words down in order")
                .font(.system(size: 22, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
            Text("Store them somewhere only you can access. Anyone with this phrase can restore your ECHO identity.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .lineSpacing(2)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func wordGrid(phrase: RecoveryPhrase) -> some View {
        let columns = Array(repeating: GridItem(.flexible(), spacing: 10), count: 4)
        return LazyVGrid(columns: columns, spacing: 10) {
            ForEach(Array(phrase.words.enumerated()), id: \.offset) { index, word in
                wordCell(position: index + 1, word: word)
            }
        }
    }

    private func wordCell(position: Int, word: String) -> some View {
        HStack(spacing: 6) {
            Text("\(position)")
                .font(.system(size: 10, weight: .semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant.opacity(0.6))
                .monospacedDigit()
            Text(word)
                .font(.system(size: 14, weight: .medium, design: .monospaced))
                .foregroundStyle(Color.Echo.onSurface)
        }
        .frame(maxWidth: .infinity)
        .padding(.horizontal, 8)
        .padding(.vertical, 10)
        .background(Color.Echo.surfaceContainerLowest, in: RoundedRectangle(cornerRadius: 10))
        .overlay(
            RoundedRectangle(cornerRadius: 10)
                .strokeBorder(Color.Echo.outlineVariant.opacity(0.4), lineWidth: 0.5)
        )
    }

    private var footer: some View {
        VStack(spacing: 10) {
            Button {
                coordinator.phraseDisplayContinued()
            } label: {
                Text("I've written it down")
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
            }
            .buttonStyle(SpringPressStyle())
            .accessibilityLabel("I've written the recovery phrase down")

            Button("Cancel") { coordinator.onCancel() }
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
    }

    private var screenCaptureWarning: some View {
        VStack(spacing: 12) {
            Image(systemName: "eye.slash.fill")
                .font(.system(size: 40))
                .foregroundStyle(Color.Echo.primaryContainer)
            Text("Recording detected")
                .font(.system(size: 18, weight: .semibold))
                .foregroundStyle(Color.Echo.onSurface)
            Text("Stop screen recording to view your phrase.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
        }
        .padding(24)
        .background(Color.Echo.surface, in: RoundedRectangle(cornerRadius: 20))
    }
}
