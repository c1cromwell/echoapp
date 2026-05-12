#if os(iOS)
// Features/Onboarding/FirstRun/DisplayNameEntryView.swift
// Page 2 of the first-run flow: single display-name field, privacy helper card,
// gradient "Start messaging" CTA, and a muted "Restore" link.

import SwiftUI

public struct DisplayNameEntryView: View {
    @Bindable var coordinator: FirstRunCoordinator
    @FocusState private var nameFieldFocused: Bool

    public init(coordinator: FirstRunCoordinator) {
        self.coordinator = coordinator
    }

    private var trimmed: String {
        coordinator.displayName.trimmingCharacters(in: .whitespacesAndNewlines)
    }
    private var isValid: Bool { DisplayNameValidator.isValid(coordinator.displayName) }

    public var body: some View {
        ZStack {
            Color.Echo.surface.ignoresSafeArea()
            AtmosphericBackground()

            VStack(alignment: .leading, spacing: 0) {
                titleBlock
                    .padding(.horizontal, 24)
                    .padding(.top, 24)

                nameField
                    .padding(.horizontal, 24)
                    .padding(.top, 28)

                helperCard
                    .padding(.horizontal, 24)
                    .padding(.top, 14)

                Spacer(minLength: 12)

                primaryCTA
                    .padding(.horizontal, 24)
                    .padding(.bottom, 10)

                restoreLink
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding(.bottom, 24)
            }
        }
        .navigationTitle("")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarLeading) {
                Button { coordinator.back() } label: {
                    Image(systemName: "chevron.backward")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(Color.Echo.primaryContainer)
                        .padding(8)
                        .background(Color.Echo.primaryContainer.opacity(0.09), in: Circle())
                }
                .accessibilityLabel("Back to welcome")
            }
            ToolbarItem(placement: .topBarTrailing) {
                Text("Step 1 of 3")
                    .font(.system(size: 11, weight: .semibold))
                    .tracking(1.4)
                    .foregroundStyle(Color.Echo.onSurfaceVariant)
            }
        }
        .safeAreaInset(edge: .top, spacing: 0) { SecureThreadIndicator() }
        .onAppear {
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                nameFieldFocused = true
            }
        }
    }

    // MARK: - Subviews

    private var titleBlock: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Choose your username")
                .font(.system(size: 24, weight: .semibold))
                .kerning(-0.3)
                .foregroundStyle(Color.Echo.onSurface)
                .lineSpacing(2)
            Text("This is your display name — change it anytime from your profile.")
                .font(.system(size: 13))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
                .lineSpacing(2)
        }
    }

    private var nameField: some View {
        HStack(spacing: 10) {
            TextField("", text: $coordinator.displayName)
                .textFieldStyle(.plain)
                .font(.system(size: 15, weight: .semibold))
                .kerning(-0.1)
                .foregroundStyle(Color.Echo.onSurface)
                .focused($nameFieldFocused)
                .textInputAutocapitalization(.words)
                .textContentType(.nickname)
                .autocorrectionDisabled(true)
                .submitLabel(.go)
                .onSubmit {
                    if isValid { coordinator.displayNameSubmitted(coordinator.displayName) }
                }
                .accessibilityLabel("Display name")
                .accessibilityHint("What others will see when you message them")

            Text("\(trimmed.count)/32")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(
                    trimmed.count > 32
                        ? Color.red
                        : Color.Echo.onSurfaceVariant.opacity(0.6)
                )
                .monospacedDigit()
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
        .background(
            Color.Echo.surfaceContainerLowest,
            in: RoundedRectangle(cornerRadius: 20, style: .continuous)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 20, style: .continuous)
                .strokeBorder(
                    nameFieldFocused
                        ? Color.Echo.primaryContainer.opacity(0.45)
                        : Color.Echo.primaryContainer.opacity(0.15),
                    lineWidth: 0.5
                )
        )
        .animation(.easeInOut(duration: 0.15), value: nameFieldFocused)
    }

    private var helperCard: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "lock.fill")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 16, height: 16)
                .padding(.top, 1)
            Text("You won't need a phone number, email, or password to use ECHO. Your identity lives only on this device.")
                .font(.system(size: 12))
                .foregroundStyle(Color.Echo.onSurface.opacity(0.85))
                .lineSpacing(2)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
        .background(
            Color.Echo.primaryContainer.opacity(0.05),
            in: RoundedRectangle(cornerRadius: 16, style: .continuous)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .strokeBorder(Color.Echo.primaryContainer.opacity(0.18), lineWidth: 0.5)
        )
    }

    private var primaryCTA: some View {
        Button {
            coordinator.displayNameSubmitted(coordinator.displayName)
        } label: {
            HStack(spacing: 8) {
                Text("Next")
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
            .opacity(isValid ? 1 : 0.55)
        }
        .buttonStyle(SpringPressStyle())
        .disabled(!isValid)
        .accessibilityHint(
            isValid
                ? "Provisions your identity and opens Messages"
                : "Enter a display name to continue"
        )
    }

    private var restoreLink: some View {
        HStack(spacing: 4) {
            Text("Already have an ECHO account?")
                .font(.system(size: 11.5))
                .foregroundStyle(Color.Echo.onSurfaceVariant)
            Button("Restore") { coordinator.restoreTapped() }
                .font(.system(size: 11.5, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
        }
    }
}
#endif
