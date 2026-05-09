#if os(iOS)
// Features/Onboarding/FirstRun/WelcomeCarouselView.swift
// Page 1 of the first-run flow: ECHO logo + rotating privacy-first copy lines.
// Auto-advances every 3.5 s; user can swipe or tap dots to drive manually.

import SwiftUI

struct CarouselSlide: Identifiable, Hashable {
    let id = UUID()
    let title: String
    let body: String
}

enum CarouselCopy {
    static let slides: [CarouselSlide] = [
        CarouselSlide(
            title: "Private messaging, always",
            body: "Built with privacy and security first. You own your identity — nobody else."
        ),
        CarouselSlide(
            title: "End-to-end encrypted by default",
            body: "Zero metadata. No tracking, no profiling. Not even ECHO can read your messages."
        ),
        CarouselSlide(
            title: "No phone, no email, no password",
            body: "Sign up with just a display name. Your keys never leave this device."
        )
    ]
}

struct WelcomeCarouselView: View {
    let coordinator: FirstRunCoordinator

    @State private var currentIndex: Int = 0
    @State private var autoAdvance: Task<Void, Never>?
    private let autoAdvanceInterval: Duration = .seconds(3.5)

    var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(spacing: 0) {
                Spacer(minLength: 0)

                EchoLogo(size: 72)
                    .padding(.bottom, 14)

                Text("ECHO")
                    .font(.system(size: 22, weight: .semibold))
                    .kerning(-0.5)
                    .foregroundStyle(Color.Echo.primaryContainer)
                    .padding(.bottom, 48)

                carouselCopy
                    .frame(maxWidth: .infinity)
                    .padding(.horizontal, 28)

                Spacer(minLength: 0)

                pageDots
                    .padding(.bottom, 20)
                    .padding(.top, 24)

                continueButton
                    .padding(.horizontal, 24)
                    .padding(.bottom, 28)
            }
        }
        .navigationBarBackButtonHidden(true)
        .toolbar(.hidden, for: .navigationBar)
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .task { startAutoAdvance() }
        .onDisappear { autoAdvance?.cancel() }
    }

    // MARK: - Carousel

    private var carouselCopy: some View {
        TabView(selection: $currentIndex) {
            ForEach(Array(CarouselCopy.slides.enumerated()), id: \.offset) { index, slide in
                VStack(spacing: 14) {
                    Text(slide.title)
                        .font(.system(size: 22, weight: .semibold))
                        .kerning(-0.3)
                        .foregroundStyle(Color.Echo.onSurface)
                        .multilineTextAlignment(.center)
                        .lineSpacing(2)
                    Text(slide.body)
                        .font(.system(size: 14))
                        .foregroundStyle(Color.Echo.onSurfaceVariant)
                        .multilineTextAlignment(.center)
                        .lineSpacing(3)
                }
                .tag(index)
                .padding(.horizontal, 6)
            }
        }
        .tabViewStyle(.page(indexDisplayMode: .never))
        .frame(maxHeight: 160)
        .animation(.spring(response: 0.55, dampingFraction: 0.85), value: currentIndex)
        .onChange(of: currentIndex) { _, _ in restartAutoAdvance() }
    }

    private var pageDots: some View {
        HStack(spacing: 6) {
            ForEach(CarouselCopy.slides.indices, id: \.self) { index in
                Capsule()
                    .fill(
                        index == currentIndex
                            ? Color.Echo.primaryContainer
                            : Color.Echo.outlineVariant
                    )
                    .frame(width: index == currentIndex ? 24 : 6, height: 6)
                    .animation(.spring(response: 0.4, dampingFraction: 0.85), value: currentIndex)
                    .onTapGesture { currentIndex = index }
                    .accessibilityAddTraits(.isButton)
                    .accessibilityLabel("Page \(index + 1) of \(CarouselCopy.slides.count)")
            }
        }
        .accessibilityElement(children: .contain)
    }

    private var continueButton: some View {
        Button {
            autoAdvance?.cancel()
            coordinator.welcomeContinueTapped()
        } label: {
            HStack(spacing: 8) {
                Text("Continue")
                    .font(.system(size: 15, weight: .semibold))
                Image(systemName: "arrow.right")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.7))
            }
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                LinearGradient(
                    colors: [Color.Echo.deepNavy, Color.Echo.primaryContainer],
                    startPoint: .topLeading, endPoint: .bottomTrailing
                ),
                in: RoundedRectangle(cornerRadius: 22, style: .continuous)
            )
        }
        .buttonStyle(SpringPressStyle())
        .accessibilityHint("Go to display name entry")
    }

    // MARK: - Auto-advance

    private func startAutoAdvance() {
        autoAdvance?.cancel()
        autoAdvance = Task { @MainActor in
            while !Task.isCancelled {
                try? await Task.sleep(for: autoAdvanceInterval)
                guard !Task.isCancelled else { return }
                let next = (currentIndex + 1) % CarouselCopy.slides.count
                withAnimation(.spring(response: 0.55, dampingFraction: 0.85)) {
                    currentIndex = next
                }
            }
        }
    }

    private func restartAutoAdvance() { startAutoAdvance() }
}
#endif
