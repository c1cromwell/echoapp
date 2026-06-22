#if os(iOS)
import SwiftUI

public struct DisplayNameEntryView: View {
    @Bindable var coordinator: FirstRunCoordinator
    @FocusState private var nameFieldFocused: Bool

    private let availabilityClient: any UsernameAvailabilityClient
    @State private var availability: AvailabilityState = .idle
    @State private var debounceTask: Task<Void, Never>?

    init(
        coordinator: FirstRunCoordinator,
        availabilityClient: (any UsernameAvailabilityClient)? = nil
    ) {
        self.coordinator = coordinator
        self.availabilityClient = availabilityClient ?? UsernameAvailabilityService()
    }

    private var trimmed: String {
        coordinator.displayName.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private var formatValid: Bool { UsernameValidator.isValid(coordinator.displayName) }

    private var canContinue: Bool {
        formatValid && availability == .available
    }

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

                availabilityHint
                    .padding(.horizontal, 24)
                    .padding(.top, 8)

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
                Text("Step 1 of 2")
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
        .onChange(of: coordinator.displayName) { _, newValue in
            scheduleAvailabilityCheck(for: newValue)
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
            Text("3–30 characters: letters, numbers, and underscores. Must be unique.")
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
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled(true)
                .keyboardType(.asciiCapable)
                .submitLabel(.go)
                .onSubmit {
                    if canContinue { coordinator.displayNameSubmitted(coordinator.displayName) }
                }
                .accessibilityLabel("Username")
                .accessibilityHint("Your public handle on Echo")

            availabilityIcon

            Text("\(trimmed.count)/30")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(
                    trimmed.count > 30 || !formatValid && !trimmed.isEmpty
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
                .strokeBorder(fieldBorderColor, lineWidth: 0.5)
        )
        .animation(.easeInOut(duration: 0.15), value: nameFieldFocused)
        .animation(.easeInOut(duration: 0.15), value: availability)
    }

    @ViewBuilder
    private var availabilityIcon: some View {
        switch availability {
        case .idle:
            EmptyView()
        case .checking:
            ProgressView()
                .controlSize(.small)
        case .available:
            Image(systemName: "checkmark.circle.fill")
                .foregroundStyle(Color.echoTrustGreen)
        case .taken, .invalid, .unavailable:
            Image(systemName: "xmark.circle.fill")
                .foregroundStyle(Color.red.opacity(0.85))
        case .error:
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(Color.orange)
        }
    }

    private var fieldBorderColor: Color {
        switch availability {
        case .available:
            return Color.echoTrustGreen.opacity(0.45)
        case .taken, .invalid, .unavailable:
            return Color.red.opacity(0.45)
        default:
            return nameFieldFocused
                ? Color.Echo.primaryContainer.opacity(0.45)
                : Color.Echo.primaryContainer.opacity(0.15)
        }
    }

    @ViewBuilder
    private var availabilityHint: some View {
        if let message = availability.message {
            Text(message)
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(availability.isNegative ? Color.red.opacity(0.9) : Color.echoTrustGreen)
        }
    }

    private var helperCard: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "lock.fill")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(Color.Echo.primaryContainer)
                .frame(width: 16, height: 16)
                .padding(.top, 1)
            Text("Your account is created right here on your phone — no email, password, or phone number needed.")
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
            .opacity(canContinue ? 1 : 0.55)
        }
        .buttonStyle(SpringPressStyle())
        .disabled(!canContinue)
        .accessibilityHint(
            canContinue
                ? "Continue to secure your account with Face ID"
                : "Enter an available username to continue"
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

    // MARK: - Availability

    private func scheduleAvailabilityCheck(for raw: String) {
        debounceTask?.cancel()
        let trimmed = raw.trimmingCharacters(in: .whitespacesAndNewlines)

        guard !trimmed.isEmpty else {
            availability = .idle
            return
        }
        guard UsernameValidator.isValid(trimmed) else {
            availability = .invalid
            return
        }

        availability = .checking
        debounceTask = Task {
            try? await Task.sleep(nanoseconds: 350_000_000)
            guard !Task.isCancelled else { return }
            do {
                let result = try await availabilityClient.checkAvailability(username: trimmed)
                guard !Task.isCancelled else { return }
                if result.available {
                    availability = .available
                } else if result.reason == "invalid_format" {
                    availability = .invalid
                } else if result.reason == "taken" {
                    availability = .taken
                } else {
                    availability = .unavailable
                }
            } catch {
                guard !Task.isCancelled else { return }
                availability = .error
            }
        }
    }

    private enum AvailabilityState: Equatable {
        case idle
        case checking
        case available
        case taken
        case invalid
        case unavailable
        case error

        var isNegative: Bool {
            switch self {
            case .available, .idle, .checking:
                return false
            default:
                return true
            }
        }

        var message: String? {
            switch self {
            case .idle, .checking:
                return nil
            case .available:
                return "Username is available"
            case .taken:
                return "That username is taken — try another"
            case .invalid:
                return "Use 3–30 letters, numbers, or underscores"
            case .unavailable:
                return "Username is not available"
            case .error:
                return "Could not check availability — try again"
            }
        }
    }
}
#endif
